package cmd

import (
	"bytes"
	"net/url"
	"strings"
	"testing"
)

func TestTerminalSafeInlinePreservesPrintableUnicode(t *testing.T) {
	input := "A safe title — Ελληνικά 日本語"
	if got := terminalSafeInline(input); got != input {
		t.Fatalf("terminalSafeInline() = %q, want %q", got, input)
	}
}

func TestTerminalSafeInlineEscapesNonPrintingRunes(t *testing.T) {
	input := "a\n\r\t\x1b\x7f\u0085\u202e"
	want := `a\n\r\t\x1b\x7f\u0085\u202e`
	if got := terminalSafeInline(input); got != want {
		t.Fatalf("terminalSafeInline() = %q, want %q", got, want)
	}
}

func TestTerminalSafeBlockPreservesLineBreaksWithIndentation(t *testing.T) {
	input := "first\r\nsecond\nthird\rfourth"
	want := "first\n  second\n  third\\rfourth"
	if got := terminalSafeBlock(input, "  "); got != want {
		t.Fatalf("terminalSafeBlock() = %q, want %q", got, want)
	}
}

func TestWriteJSONRetainsStructuredEscaping(t *testing.T) {
	var output bytes.Buffer
	if err := writeJSON(&output, map[string]string{"value": "safe\x1b[2J"}); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	want := "{\n  \"value\": \"safe\\u001b[2J\"\n}\n"
	if got := output.String(); got != want {
		t.Fatalf("JSON output = %q, want %q", got, want)
	}
}

func TestAtomFeedURLPreservesNormalOutput(t *testing.T) {
	got, err := atomFeedURL("https://paperzilla.ai", "proj-1", "feed-token")
	if err != nil {
		t.Fatalf("atomFeedURL: %v", err)
	}
	if want := "https://paperzilla.ai/api/feed/atom/proj-1?token=feed-token"; got != want {
		t.Fatalf("atomFeedURL() = %q, want %q", got, want)
	}
}

func TestAtomFeedURLEncodesProjectAndToken(t *testing.T) {
	projectID := "proj /1\x1b"
	token := "feed token&\x1b[2J"
	got, err := atomFeedURL("https://paperzilla.ai", projectID, token)
	if err != nil {
		t.Fatalf("atomFeedURL: %v", err)
	}
	if strings.ContainsAny(got, "\x1b\r\n") {
		t.Fatalf("URL contains active controls: %q", got)
	}

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}
	if gotPath := strings.TrimPrefix(parsed.Path, "/api/feed/atom/"); gotPath != projectID {
		t.Fatalf("decoded project ID = %q, want %q", gotPath, projectID)
	}
	if gotToken := parsed.Query().Get("token"); gotToken != token {
		t.Fatalf("decoded token = %q, want %q", gotToken, token)
	}
}
