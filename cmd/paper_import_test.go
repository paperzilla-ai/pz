package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/config"
)

func TestPaperImportProcess(t *testing.T) {
	if os.Getenv("PZ_TEST_PAPER_IMPORT_PROCESS") != "1" {
		return
	}
	rootCmd.SetArgs(os.Args[3:])
	Execute()
	os.Exit(0)
}

func runImportProcess(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	process := exec.Command(os.Args[0], append([]string{"-test.run=^TestPaperImportProcess$", "--"}, args...)...)
	process.Env = append(os.Environ(), "PZ_TEST_PAPER_IMPORT_PROCESS=1")
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	code := 0
	if err := process.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

func TestPaperImportHelpAndArgumentValidation(t *testing.T) {
	t.Setenv("PZ_TOKENS_PATH", filepath.Join(t.TempDir(), "missing.json"))
	stdout, stderr, code := runImportProcess(t, "paper", "--help")
	if code != 0 || stderr != "" || strings.Contains(stdout, "import") || !strings.Contains(stdout, "paper <paper-ref>") {
		t.Fatalf("parent help: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
	}

	for _, tt := range []struct {
		args    []string
		message string
	}{
		{[]string{"--markdown"}, "accepts 1 arg(s)"},
		{[]string{"ref", "other", "--markdown"}, "accepts 1 arg(s)"},
		{[]string{"ref"}, "specify exactly one of --markdown or --json"},
		{[]string{"ref", "--markdown=false"}, "specify exactly one of --markdown or --json"},
		{[]string{"ref", "--markdown", "--json"}, "specify exactly one of --markdown or --json"},
		{[]string{"", "--markdown"}, "a paper reference is required"},
		{[]string{"10.1234/example", "--markdown"}, "Not logged in. Run `pz login` first."},
		{[]string{"2025/1237", "--json"}, "Not logged in. Run `pz login` first."},
	} {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			args := append([]string{"paper", "import"}, tt.args...)
			stdout, stderr, code := runImportProcess(t, args...)
			if stdout != "" || code != 1 || !strings.Contains(stderr, tt.message) {
				t.Fatalf("exit=%d stdout=%q stderr=%q; want %q on stderr", code, stdout, stderr, tt.message)
			}
		})
	}
}

func TestPaperImportProcessResponses(t *testing.T) {
	const (
		ref = "https://arxiv.org/pdf/2401.12345v2.pdf"
		md  = "# A résumé\r\n\r\n**A proof.**\n\nNo trailing newline."
	)
	for _, tt := range []struct {
		name, mediaType, body string
		status, exit          int
		out, diagnostic       string
	}{
		{"ready", "text/markdown; charset=utf-8", md, 200, 0, md, ""},
		{"queued", "application/json", `{"status":"queued","paper_id":"p-1"}`, 202, 75, "", "Markdown is being prepared. Run this command again shortly.\n"},
		{"unavailable", "application/json", `{"code":"not_downloadable"}`, 404, 1, "", "not downloadable\n"},
		{"old server", "application/json", `{"detail":"Not Found"}`, 404, 1, "", "failed to import paper: HTTP 404: Not Found\n"},
		{"converter", "application/json", `{"code":"markdown_conversion_failed","detail":"Markdown conversion failed for this paper."}`, 409, 1, "", "failed to import paper: HTTP 409: Markdown conversion failed for this paper. (markdown_conversion_failed)\n"},
		{"terminal-safe diagnostic", "application/json", `{"code":"upstream_unavailable","detail":"Unavailable\u001b[2J\nRetry"}`, 503, 1, "", "failed to import paper: HTTP 503: Unavailable\\x1b[2J\\nRetry (upstream_unavailable)\n"},
		{"unexpected content", "text/html", "not markdown", 200, 1, "", "failed to import paper: expected a Markdown response\n"},
		{"invalid queued content", "application/json", `{"status":"queued"}`, 202, 1, "", "failed to import paper: invalid queued import response\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			savePaperTestTokens(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]string
				if r.Method != http.MethodPost || r.URL.Path != "/api/papers/import" ||
					r.Header.Get("X-Paperzilla-Client") != "cli" || r.Header.Get("Authorization") != "Bearer access-1" ||
					json.NewDecoder(r.Body).Decode(&body) != nil || body["reference"] != ref || len(body) != 1 {
					t.Errorf("unexpected import request: %s %s: %v", r.Method, r.URL.Path, body)
				}
				w.Header().Set("Content-Type", tt.mediaType)
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			t.Setenv("PZ_API_URL", server.URL)
			stdout, stderr, code := runImportProcess(t, "paper", "import", ref, "--markdown")
			if code != tt.exit || stdout != tt.out || stderr != tt.diagnostic {
				t.Fatalf("exit=%d stdout=%q stderr=%q; want %d, %q, %q", code, stdout, stderr, tt.exit, tt.out, tt.diagnostic)
			}
		})
	}
}

func TestPaperImportJSONMatchesCanonicalRecord(t *testing.T) {
	for _, tt := range []struct {
		reference, paperID, status, reason string
		ready                              bool
	}{
		{"2025/1237", "6eefbb22-7e50-46b8-bc56-2c8c959f64a3", "unavailable", "Full text not available.", false},
		{"https://arxiv.org/abs/2611.19926", "18e69253-6dc4-45ac-acdb-73dd5bc65e36", "ready", "", true},
	} {
		t.Run(tt.reference, func(t *testing.T) {
			savePaperTestTokens(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet && r.URL.Path == "/api/auth/cli-access" {
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprint(w, `{"allowed":true}`)
					return
				}
				paper := map[string]any{
					"id": tt.paperID, "title": "Regression fixture", "abstract": "Résumé",
					"authors": []map[string]string{{"name": "Fixture Author"}}, "markdown_ready": tt.ready,
				}
				if r.URL.Path == "/api/papers/import" {
					var body map[string]string
					if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer access-1" ||
						r.Header.Get("X-Paperzilla-Client") != "cli" || json.NewDecoder(r.Body).Decode(&body) != nil ||
						body["reference"] != tt.reference || body["format"] != "json" || len(body) != 2 {
						t.Errorf("unexpected import request: %s %s: %v", r.Method, r.URL.Path, body)
					}
					paper["full_text_status"] = tt.status
					paper["full_text_reason"] = tt.reason
				} else if r.Method != http.MethodGet || r.URL.Path != "/api/public/papers/"+tt.paperID {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(paper)
			}))
			defer server.Close()
			t.Setenv("PZ_API_URL", server.URL)

			stdout, stderr, code := runImportProcess(t, "paper", "import", tt.reference, "--json")
			var imported map[string]any
			if code != 0 || stderr != "" || json.Unmarshal([]byte(stdout), &imported) != nil {
				t.Fatalf("import exit=%d stdout=%q stderr=%q", code, stdout, stderr)
			}
			if imported["id"] != tt.paperID || imported["markdown_ready"] != tt.ready ||
				imported["full_text_status"] != tt.status || tt.reason != "" && imported["full_text_reason"] != tt.reason {
				t.Fatalf("unexpected import record: %v", imported)
			}
			delete(imported, "full_text_status")
			delete(imported, "full_text_reason")

			stdout, stderr, code = runImportProcess(t, "paper", tt.paperID, "--json")
			var canonical map[string]any
			if code != 0 || stderr != "" || json.Unmarshal([]byte(stdout), &canonical) != nil ||
				!reflect.DeepEqual(imported, canonical) {
				t.Fatalf("canonical exit=%d stdout=%q stderr=%q; imported record %v", code, stdout, stderr, imported)
			}
		})
	}
}

func TestPaperImportJSONErrorsDoNotPrintRecords(t *testing.T) {
	for _, tt := range []struct {
		body, diagnostic string
		status           int
	}{
		{`{"code":"reference_not_found","detail":"No paper found for this reference."}`, "No paper found for this reference.", 404},
		{`{"id":"p-1","full_text_status":"unavailable"}`, "invalid paper import record", 200},
		{strings.Repeat(" ", (16<<20)+1), "16777216-byte limit", 200},
	} {
		t.Run(tt.diagnostic, func(t *testing.T) {
			savePaperTestTokens(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			t.Setenv("PZ_API_URL", server.URL)
			stdout, stderr, code := runImportProcess(t, "paper", "import", "1900/001", "--json")
			if stdout != "" || code != 1 || !strings.Contains(stderr, tt.diagnostic) {
				t.Fatalf("exit=%d stdout=%q stderr=%q; want %q on stderr", code, stdout, stderr, tt.diagnostic)
			}
		})
	}
}

func TestPaperImportSilentlyRefreshesRejectedSession(t *testing.T) {
	savePaperTestTokens(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/papers/import":
			switch r.Header.Get("Authorization") {
			case "Bearer access-1":
				w.WriteHeader(http.StatusUnauthorized)
			case "Bearer fresh-access":
				w.Header().Set("Content-Type", "text/markdown")
				_, _ = w.Write([]byte("# Imported\n"))
			default:
				t.Error("import did not use a stored token")
			}
		case "/api/auth/refresh":
			var body map[string]string
			if json.NewDecoder(r.Body).Decode(&body) != nil || body["refresh_token"] != "refresh-1" {
				t.Errorf("unexpected refresh request: %v", body)
			}
			_, _ = w.Write([]byte(`{"access_token":"fresh-access","refresh_token":"fresh-refresh","expires_in":3600}`))
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	t.Setenv("PZ_API_URL", server.URL)
	stdout, stderr, code := runImportProcess(t, "paper", "import", "2026/001", "--markdown")
	if code != 0 || stdout != "# Imported\n" || stderr != "" {
		t.Fatalf("exit=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	tokens, err := config.LoadTokens()
	if err != nil || tokens.AccessToken != "fresh-access" || tokens.RefreshToken != "fresh-refresh" {
		t.Fatalf("the refreshed session was not saved: %v", err)
	}
}

func TestPaperImportDoesNotPrintPartialOrOversizeResponses(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusAccepted, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			savePaperTestTokens(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if status == http.StatusOK {
					w.Header().Set("Content-Type", "text/markdown")
					w.Header().Set("Content-Length", "100")
					_, _ = fmt.Fprint(w, "# Incomplete")
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = fmt.Fprint(w, strings.Repeat(" ", (16<<20)+1))
			}))
			defer server.Close()
			t.Setenv("PZ_API_URL", server.URL)
			stdout, stderr, code := runImportProcess(t, "paper", "import", "2026/001", "--markdown")
			diagnostic := "16777216-byte limit"
			if status == http.StatusOK {
				diagnostic = "unexpected EOF"
			}
			if code != 1 || stdout != "" || !strings.Contains(stderr, diagnostic) {
				t.Fatalf("exit=%d stdout=%q stderr=%q; want %q on stderr", code, stdout, stderr, diagnostic)
			}
		})
	}
}
