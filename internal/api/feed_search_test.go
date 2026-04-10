package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestFetchFeedSearch(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/proj-1/feed/search" {
			t.Fatalf("path = %s, want /api/projects/proj-1/feed/search", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "latent retrieval" {
			t.Fatalf("q = %q, want %q", got, "latent retrieval")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer my_token" {
			t.Fatalf("Authorization = %q", got)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"id":              "pp-1",
					"short_id":        "feedbeef",
					"paper_title":     "Latent Retrieval for Papers",
					"relevance_score": 0.94,
					"relevance_class": 2,
					"paper": map[string]any{
						"id":    "paper-1",
						"title": "Latent Retrieval for Papers",
					},
				},
			},
			"limit":    20,
			"offset":   0,
			"has_more": true,
			"query":    "latent retrieval",
		})
	})
	defer server.Close()

	resp, err := FetchFeedSearch("my_token", "proj-1", FeedSearchOptions{Query: "latent retrieval"})
	if err != nil {
		t.Fatalf("FetchFeedSearch: %v", err)
	}
	if resp.Query != "latent retrieval" {
		t.Fatalf("Query = %q", resp.Query)
	}
	if !resp.HasMore {
		t.Fatal("HasMore = false, want true")
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}
	if resp.Items[0].PaperTitle != "Latent Retrieval for Papers" {
		t.Fatalf("PaperTitle = %q", resp.Items[0].PaperTitle)
	}
}

func TestFetchFeedSearchQueryParams(t *testing.T) {
	mustRead := true

	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("q"); got != "latent retrieval" {
			t.Fatalf("q = %q", got)
		}
		if got := q.Get("feedback_filter"); got != "starred" {
			t.Fatalf("feedback_filter = %q", got)
		}
		if got := q.Get("must_read"); got != "true" {
			t.Fatalf("must_read = %q", got)
		}
		if got := q.Get("limit"); got != "5" {
			t.Fatalf("limit = %q", got)
		}
		if got := q.Get("offset"); got != "10" {
			t.Fatalf("offset = %q", got)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"items":    []any{},
			"limit":    5,
			"offset":   10,
			"has_more": false,
			"query":    "latent retrieval",
		})
	})
	defer server.Close()

	_, err := FetchFeedSearch("my_token", "proj-1", FeedSearchOptions{
		Query:          " latent retrieval ",
		FeedbackFilter: "starred",
		MustRead:       &mustRead,
		Limit:          5,
		Offset:         10,
	})
	if err != nil {
		t.Fatalf("FetchFeedSearch: %v", err)
	}
}

func TestFetchFeedSearchNoOptionalParams(t *testing.T) {
	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RawQuery; got != "q=latent+retrieval" {
			t.Fatalf("RawQuery = %q, want %q", got, "q=latent+retrieval")
		}

		json.NewEncoder(w).Encode(map[string]any{
			"items":    []any{},
			"limit":    20,
			"offset":   0,
			"has_more": false,
			"query":    "latent retrieval",
		})
	})
	defer server.Close()

	_, err := FetchFeedSearch("my_token", "proj-1", FeedSearchOptions{Query: "latent retrieval"})
	if err != nil {
		t.Fatalf("FetchFeedSearch: %v", err)
	}
}
