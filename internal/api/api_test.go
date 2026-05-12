package api

import (
	"encoding/json"
	"errors"
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
			"positive_keywords": []string{"retrieval"},
			"negative_keywords": []string{"survey"},
			"sources": []map[string]any{
				{"id": 1, "name": "arXiv", "base_url": "https://arxiv.org"},
			},
			"categories": []map[string]any{
				{
					"id": 10, "code": "cs.CL", "name": "Computation and Language",
					"source_id": 1, "source_name": "arXiv", "weight": 1.0,
				},
			},
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
	if len(project.PositiveKeywords) != 1 || project.PositiveKeywords[0] != "retrieval" {
		t.Errorf("PositiveKeywords = %#v", project.PositiveKeywords)
	}
	if len(project.NegativeKeywords) != 1 || project.NegativeKeywords[0] != "survey" {
		t.Errorf("NegativeKeywords = %#v", project.NegativeKeywords)
	}
	if len(project.Sources) != 1 || project.Sources[0].Name != "arXiv" {
		t.Errorf("Sources = %#v", project.Sources)
	}
	if len(project.Categories) != 1 || project.Categories[0].Code != "cs.CL" {
		t.Errorf("Categories = %#v", project.Categories)
	}
}

func TestFetchProjectNormalizesMissingMetadata(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
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
	if project.PositiveKeywords == nil || project.NegativeKeywords == nil || project.Sources == nil || project.Categories == nil {
		t.Fatalf("metadata arrays were not normalized: %#v", project)
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

func TestFetchPaper(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/papers/paper-1" {
			t.Errorf("path = %s, want /api/papers/paper-1", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer my_token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}

		json.NewEncoder(w).Encode(map[string]any{
			"id": "paper-1", "title": "Test Paper", "short_id": "abc12345",
			"venue_name": "arXiv", "reference_label": "arXiv 2401.12345",
			"source": map[string]any{"name": "arxiv"},
		})
	})
	defer server.Close()

	paper, err := FetchPaper("my_token", "paper-1")
	if err != nil {
		t.Fatalf("FetchPaper: %v", err)
	}
	if paper.ID != "paper-1" {
		t.Errorf("ID = %q", paper.ID)
	}
	if paper.Source == nil || paper.Source.Name != "arxiv" {
		t.Errorf("Source.Name = %#v, want %q", paper.Source, "arxiv")
	}
	if paper.VenueName != "arXiv" {
		t.Errorf("VenueName = %q, want %q", paper.VenueName, "arXiv")
	}
	if paper.ReferenceLabel != "arXiv 2401.12345" {
		t.Errorf("ReferenceLabel = %q", paper.ReferenceLabel)
	}
}

func TestFetchPaperNotFound(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"detail":"Paper not found"}`))
	})
	defer server.Close()

	_, err := FetchPaper("my_token", "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchPublicPaper(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/papers/paper-1" {
			t.Errorf("path = %s, want /api/public/papers/paper-1", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization = %q, want empty", got)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"id": "paper-1", "title": "Public Paper", "short_id": "abc12345",
			"venue_name": "arXiv", "reference_label": "arXiv 2401.12345",
			"markdown_ready": true,
		})
	})
	defer server.Close()

	paper, err := FetchPublicPaper("paper-1")
	if err != nil {
		t.Fatalf("FetchPublicPaper: %v", err)
	}
	if paper.ID != "paper-1" {
		t.Errorf("ID = %q", paper.ID)
	}
	if !paper.MarkdownReady {
		t.Errorf("MarkdownReady = false, want true")
	}
	if paper.VenueName != "arXiv" {
		t.Errorf("VenueName = %q, want %q", paper.VenueName, "arXiv")
	}
	if paper.ReferenceLabel != "arXiv 2401.12345" {
		t.Errorf("ReferenceLabel = %q", paper.ReferenceLabel)
	}
}

func TestFetchPaperMarkdown(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/papers/paper-1/markdown" {
			t.Errorf("path = %s, want /api/papers/paper-1/markdown", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer my_token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}

		w.Write([]byte("# Test Paper\n\nHello markdown.\n"))
	})
	defer server.Close()

	markdown, err := FetchPaperMarkdown("my_token", "paper-1")
	if err != nil {
		t.Fatalf("FetchPaperMarkdown: %v", err)
	}
	if markdown != "# Test Paper\n\nHello markdown.\n" {
		t.Fatalf("markdown = %q", markdown)
	}
}

func TestFetchPublicPaperMarkdown(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/papers/paper-1/markdown" {
			t.Errorf("path = %s, want /api/public/papers/paper-1/markdown", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization = %q, want empty", got)
		}

		w.Write([]byte("# Public Markdown\n"))
	})
	defer server.Close()

	markdown, err := FetchPublicPaperMarkdown("paper-1")
	if err != nil {
		t.Fatalf("FetchPublicPaperMarkdown: %v", err)
	}
	if markdown != "# Public Markdown\n" {
		t.Fatalf("markdown = %q", markdown)
	}
}

func TestFetchPublicPaperMarkdownNotReady(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		w.Write([]byte(`{"detail":"Markdown is not ready for this paper","code":"markdown_not_ready"}`))
	})
	defer server.Close()

	_, err := FetchPublicPaperMarkdown("paper-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want APIError", err)
	}
	if apiErr.StatusCode != 409 || apiErr.Code != "markdown_not_ready" {
		t.Fatalf("APIError = %#v", apiErr)
	}
}

func TestFetchPaperMarkdownNotFound(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"detail":"Paper markdown not found"}`))
	})
	defer server.Close()

	_, err := FetchPaperMarkdown("my_token", "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchPaperMarkdownQueued(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"detail":"Markdown queued","code":"markdown_queued","job_id":"job-1","created":true}`))
	})
	defer server.Close()

	_, err := FetchPaperMarkdown("my_token", "paper-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var pending *PaperMarkdownPendingError
	if !errors.As(err, &pending) {
		t.Fatalf("err = %v, want PaperMarkdownPendingError", err)
	}
	if pending.Code != "markdown_queued" {
		t.Fatalf("Code = %q", pending.Code)
	}
}

func TestFetchPaperMarkdownAlreadyQueued(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"detail":"Markdown already queued","code":"markdown_already_queued","job_id":"job-1","created":false}`))
	})
	defer server.Close()

	_, err := FetchPaperMarkdown("my_token", "paper-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var pending *PaperMarkdownPendingError
	if !errors.As(err, &pending) {
		t.Fatalf("err = %v, want PaperMarkdownPendingError", err)
	}
	if pending.Code != "markdown_already_queued" {
		t.Fatalf("Code = %q", pending.Code)
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
					"paper": map[string]any{
						"id":              "p-1",
						"title":           "Test Paper",
						"authors":         []map[string]string{{"name": "Smith"}},
						"venue_name":      "arXiv",
						"reference_label": "arXiv 2401.12345",
						"source_id":       1,
						"source":          map[string]any{"name": "arxiv"},
					},
					"feedback": map[string]any{"vote": "star"},
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
	if feed.Items[0].Paper.Source == nil || feed.Items[0].Paper.Source.Name != "arxiv" {
		t.Errorf("Source.Name = %#v, want %q", feed.Items[0].Paper.Source, "arxiv")
	}
	if feed.Items[0].Paper.VenueName != "arXiv" {
		t.Errorf("VenueName = %q, want %q", feed.Items[0].Paper.VenueName, "arXiv")
	}
	if feed.Items[0].Paper.ReferenceLabel != "arXiv 2401.12345" {
		t.Errorf("ReferenceLabel = %q", feed.Items[0].Paper.ReferenceLabel)
	}
	if feed.Items[0].Feedback == nil || feed.Items[0].Feedback.Vote != "star" {
		t.Errorf("Feedback = %#v, want star", feed.Items[0].Feedback)
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

func TestFetchFeedToken(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/auth/feed-token" {
			t.Errorf("path = %s, want /api/auth/feed-token", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer my_token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}

		json.NewEncoder(w).Encode(map[string]any{
			"token":      "pzft_abc123",
			"created_at": "2025-01-01T00:00:00Z",
		})
	})
	defer server.Close()

	resp, err := FetchFeedToken("my_token")
	if err != nil {
		t.Fatalf("FetchFeedToken: %v", err)
	}
	if resp.Token != "pzft_abc123" {
		t.Errorf("Token = %q", resp.Token)
	}
	if resp.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q", resp.CreatedAt)
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

func TestFetchProjectPaper(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project-papers/feedbeef" {
			t.Errorf("path = %s, want /api/project-papers/feedbeef", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":          "pp-1",
			"short_id":    "feedbeef",
			"paper_title": "Recommendation",
			"feedback":    map[string]any{"vote": "upvote"},
			"paper":       map[string]any{"id": "paper-1", "title": "Recommendation"},
		})
	})
	defer server.Close()

	projectPaper, err := FetchProjectPaper("my_token", "feedbeef")
	if err != nil {
		t.Fatalf("FetchProjectPaper: %v", err)
	}
	if projectPaper.ID != "pp-1" {
		t.Fatalf("ID = %q", projectPaper.ID)
	}
	if projectPaper.Feedback == nil || projectPaper.Feedback.Vote != "upvote" {
		t.Fatalf("Feedback = %#v", projectPaper.Feedback)
	}
}

func TestFetchProjectPaperForProject(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/proj-1/papers/paper-1" {
			t.Errorf("path = %s, want /api/projects/proj-1/papers/paper-1", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":          "pp-1",
			"paper_title": "Recommendation",
			"paper":       map[string]any{"id": "paper-1", "title": "Recommendation"},
		})
	})
	defer server.Close()

	projectPaper, err := FetchProjectPaperForProject("my_token", "proj-1", "paper-1")
	if err != nil {
		t.Fatalf("FetchProjectPaperForProject: %v", err)
	}
	if projectPaper.ID != "pp-1" {
		t.Fatalf("ID = %q", projectPaper.ID)
	}
}

func TestSetProjectPaperFeedback(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/api/project-papers/feedbeef/feedback" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"vote": "downvote", "downvote_reason": "not_relevant", "updated_at": "2026-04-02T10:00:00Z",
		})
	})
	defer server.Close()

	feedback, err := SetProjectPaperFeedback("my_token", "feedbeef", "downvote", "not_relevant")
	if err != nil {
		t.Fatalf("SetProjectPaperFeedback: %v", err)
	}
	if feedback.Vote != "downvote" || feedback.DownvoteReason != "not_relevant" {
		t.Fatalf("Feedback = %#v", feedback)
	}
}

func TestClearProjectPaperFeedback(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/project-papers/feedbeef/feedback" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	if err := ClearProjectPaperFeedback("my_token", "feedbeef"); err != nil {
		t.Fatalf("ClearProjectPaperFeedback: %v", err)
	}
}
