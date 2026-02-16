package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// startTestServer creates a test HTTP server and sets PZ_API_URL to point to it.
// Returns the server (caller must defer server.Close()).
func startTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Setenv("PZ_API_URL", server.URL)
	return server
}

// --- Auth tests ---

func TestSendOTP(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/auth/otp" {
			t.Errorf("path = %s, want /api/auth/otp", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		json.Unmarshal(body, &req)
		if req["email"] != "test@example.com" {
			t.Errorf("email = %q, want %q", req["email"], "test@example.com")
		}

		w.WriteHeader(200)
		w.Write([]byte("{}"))
	})
	defer server.Close()

	err := SendOTP("test@example.com")
	if err != nil {
		t.Fatalf("SendOTP: %v", err)
	}
}

func TestVerifyOTP(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/verify" {
			t.Errorf("path = %s, want /api/auth/verify", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		json.Unmarshal(body, &req)
		if req["email"] != "test@example.com" {
			t.Errorf("email = %q", req["email"])
		}
		if req["code"] != "123456" {
			t.Errorf("code = %q", req["code"])
		}

		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access_abc",
			"refresh_token": "refresh_xyz",
			"expires_in":    3600,
		})
	})
	defer server.Close()

	tokens, err := VerifyOTP("test@example.com", "123456")
	if err != nil {
		t.Fatalf("VerifyOTP: %v", err)
	}
	if tokens.AccessToken != "access_abc" {
		t.Errorf("AccessToken = %q", tokens.AccessToken)
	}
	if tokens.RefreshToken != "refresh_xyz" {
		t.Errorf("RefreshToken = %q", tokens.RefreshToken)
	}
	if tokens.ExpiresAt == 0 {
		t.Error("ExpiresAt should be set")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/refresh" {
			t.Errorf("path = %s, want /api/auth/refresh", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("refresh should not send Authorization header")
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]string
		json.Unmarshal(body, &req)
		if req["refresh_token"] != "old_token" {
			t.Errorf("refresh_token = %q", req["refresh_token"])
		}

		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new_access",
			"refresh_token": "new_refresh",
			"expires_in":    3600,
		})
	})
	defer server.Close()

	tokens, err := RefreshAccessToken("old_token")
	if err != nil {
		t.Fatalf("RefreshAccessToken: %v", err)
	}
	if tokens.AccessToken != "new_access" {
		t.Errorf("AccessToken = %q", tokens.AccessToken)
	}
	if tokens.RefreshToken != "new_refresh" {
		t.Errorf("RefreshToken = %q", tokens.RefreshToken)
	}
}

func TestRefreshAccessTokenError(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"detail":"Invalid refresh token"}`))
	})
	defer server.Close()

	_, err := RefreshAccessToken("bad_token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Projects tests ---

func TestFetchProjects(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects" {
			t.Errorf("path = %s, want /api/projects", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer my_token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}

		json.NewEncoder(w).Encode([]map[string]any{
			{"id": "proj-1", "name": "Test Project", "mode": "auto", "visibility": "private", "created_at": "2025-01-01T00:00:00Z"},
		})
	})
	defer server.Close()

	projects, err := FetchProjects("my_token")
	if err != nil {
		t.Fatalf("FetchProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("len = %d, want 1", len(projects))
	}
	if projects[0].Name != "Test Project" {
		t.Errorf("Name = %q", projects[0].Name)
	}
}

func TestFetchProject(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/proj-1" {
			t.Errorf("path = %s, want /api/projects/proj-1", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"id": "proj-1", "name": "Test Project", "mode": "auto",
			"visibility": "private", "created_at": "2025-01-01T00:00:00Z",
		})
	})
	defer server.Close()

	project, err := FetchProject("my_token", "proj-1")
	if err != nil {
		t.Fatalf("FetchProject: %v", err)
	}
	if project.ID != "proj-1" {
		t.Errorf("ID = %q", project.ID)
	}
}

func TestFetchProjectNotFound(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"detail":"Project not found"}`))
	})
	defer server.Close()

	_, err := FetchProject("my_token", "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Feed tests ---

func TestFetchFeed(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/proj-1/feed" {
			t.Errorf("path = %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"id": "fp-1", "paper_title": "Test Paper",
					"relevance_score": 0.95, "relevance_class": 2,
					"paper": map[string]any{"id": "p-1", "title": "Test Paper", "authors": []map[string]string{{"name": "Smith"}}},
				},
			},
			"total": 1, "limit": 20, "offset": 0,
		})
	})
	defer server.Close()

	feed, err := FetchFeed("my_token", "proj-1", FeedOptions{})
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
	if feed.Total != 1 {
		t.Errorf("Total = %d", feed.Total)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("len(Items) = %d", len(feed.Items))
	}
	if feed.Items[0].PaperTitle != "Test Paper" {
		t.Errorf("PaperTitle = %q", feed.Items[0].PaperTitle)
	}
}

func TestFetchFeedQueryParams(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("must_read") != "true" {
			t.Errorf("must_read = %q", q.Get("must_read"))
		}
		if q.Get("since") != "2025-08-01" {
			t.Errorf("since = %q", q.Get("since"))
		}
		if q.Get("limit") != "5" {
			t.Errorf("limit = %q", q.Get("limit"))
		}
		if q.Get("offset") != "10" {
			t.Errorf("offset = %q", q.Get("offset"))
		}

		json.NewEncoder(w).Encode(map[string]any{
			"items": []any{}, "total": 0, "limit": 5, "offset": 10,
		})
	})
	defer server.Close()

	_, err := FetchFeed("my_token", "proj-1", FeedOptions{
		MustReadOnly: true,
		Since:        "2025-08-01",
		Limit:        5,
		Offset:       10,
	})
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
}

func TestFetchFeedNoOptionalParams(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %q", r.URL.RawQuery)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"items": []any{}, "total": 0, "limit": 20, "offset": 0,
		})
	})
	defer server.Close()

	_, err := FetchFeed("my_token", "proj-1", FeedOptions{})
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
}
