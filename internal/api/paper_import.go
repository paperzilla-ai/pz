package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"time"
)

type PaperImportResult struct {
	Markdown []byte
	Queued   bool
}

type PaperImportRecord struct {
	Paper
	FullTextStatus string `json:"full_text_status"`
	FullTextReason string `json:"full_text_reason,omitempty"`
}

func requestPaperImport(ctx context.Context, accessToken, reference, format string, limit func(int) int64) (responseDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	request := struct {
		Reference string `json:"reference"`
		Format    string `json:"format,omitempty"`
	}{Reference: reference, Format: format}
	return doRequestWithContext(ctx, http.MethodPost, "/api/papers/import", request, accessToken, limit)
}

func ImportPaperRecord(ctx context.Context, accessToken, reference string) (PaperImportRecord, error) {
	response, err := requestPaperImport(ctx, accessToken, reference, "json", func(int) int64 {
		return maxJSONResponseBytes
	})
	if err != nil {
		return PaperImportRecord{}, err
	}
	contentType, _, _ := mime.ParseMediaType(response.contentType)
	var decoded struct {
		PaperImportRecord
		MarkdownReady *bool `json:"markdown_ready"`
	}
	if response.status != http.StatusOK || contentType != "application/json" ||
		json.Unmarshal(response.body, &decoded) != nil || decoded.ID == "" ||
		decoded.FullTextStatus == "" || decoded.MarkdownReady == nil ||
		*decoded.MarkdownReady != (decoded.FullTextStatus == "ready") {
		return PaperImportRecord{}, errors.New("invalid paper import record")
	}
	decoded.PaperImportRecord.MarkdownReady = *decoded.MarkdownReady
	return decoded.PaperImportRecord, nil
}

func ImportPaperMarkdown(ctx context.Context, accessToken, reference string) (PaperImportResult, error) {
	response, err := requestPaperImport(ctx, accessToken, reference, "", func(status int) int64 {
		if status == http.StatusOK {
			return maxMarkdownResponseBytes
		}
		return maxJSONResponseBytes
	})
	if err != nil {
		return PaperImportResult{}, err
	}

	contentType, _, _ := mime.ParseMediaType(response.contentType)
	switch response.status {
	case http.StatusOK:
		if contentType != "text/markdown" {
			return PaperImportResult{}, errors.New("expected a Markdown response")
		}
		return PaperImportResult{Markdown: response.body}, nil
	case http.StatusAccepted:
		var pending struct {
			Status  string `json:"status"`
			PaperID string `json:"paper_id"`
		}
		if contentType != "application/json" || json.Unmarshal(response.body, &pending) != nil ||
			pending.Status != "queued" || pending.PaperID == "" {
			return PaperImportResult{}, errors.New("invalid queued import response")
		}
		return PaperImportResult{Queued: true}, nil
	default:
		return PaperImportResult{}, fmt.Errorf("unexpected import response (HTTP %d)", response.status)
	}
}
