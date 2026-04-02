package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/paperzilla/pz/internal/api"
)

func writeCanonicalPaper(w io.Writer, paper api.Paper) {
	fmt.Fprintf(w, "Title:           %s\n", displayValue(paper.Title))
	fmt.Fprintf(w, "ID:              %s\n", displayValue(paper.ID))
	fmt.Fprintf(w, "Short ID:        %s\n", displayValue(paper.ShortID))
	fmt.Fprintf(w, "Slug:            %s\n", displayValue(paper.Slug))
	fmt.Fprintf(w, "Source:          %s\n", sourceLabel(paper))
	fmt.Fprintf(w, "Published:       %s\n", formatTime(paper.PublishedDate))
	fmt.Fprintf(w, "Authors:         %s\n", displayValue(authorNames(paper.Authors)))
	fmt.Fprintf(w, "URL:             %s\n", displayValue(paper.URL))
	fmt.Fprintf(w, "PDF URL:         %s\n", displayValue(paper.PdfURL))
	fmt.Fprintf(w, "DOI:             %s\n", displayValue(paper.DOI))
	fmt.Fprintf(w, "Source Paper ID: %s\n", displayValue(paper.SourcePaperID))

	if strings.TrimSpace(paper.Abstract) != "" {
		fmt.Fprintf(w, "\nAbstract:\n  %s\n", paper.Abstract)
	}

	metadata := formatMetadata(paper.Metadata)
	if metadata != "" {
		fmt.Fprintf(w, "\nMetadata:\n%s\n", metadata)
	}
}

func writeProjectPaper(w io.Writer, projectPaper api.ProjectPaper) {
	fmt.Fprintf(w, "Title:             %s\n", displayValue(projectPaper.PaperTitle))
	fmt.Fprintf(w, "Recommendation ID: %s\n", displayValue(projectPaper.ID))
	fmt.Fprintf(w, "Short ID:          %s\n", displayValue(projectPaper.ShortID))
	fmt.Fprintf(w, "Relevance:         %s\n", formatRelevance(projectPaper.RelevanceClass, projectPaper.RelevanceScore))
	fmt.Fprintf(w, "Feedback:          %s\n", formatFeedback(projectPaper.Feedback))
	fmt.Fprintf(w, "Ready At:          %s\n", formatTime(projectPaper.ReadyAt))

	if strings.TrimSpace(projectPaper.PersonalizedNote) != "" {
		fmt.Fprintf(w, "\nNote:\n  %s\n", projectPaper.PersonalizedNote)
	}
	if strings.TrimSpace(projectPaper.Summary) != "" {
		fmt.Fprintf(w, "\nSummary:\n  %s\n", projectPaper.Summary)
	}

	paper := projectPaper.Paper
	fmt.Fprintf(w, "\nPaper:\n")
	fmt.Fprintf(w, "  ID:              %s\n", displayValue(paper.ID))
	fmt.Fprintf(w, "  Short ID:        %s\n", displayValue(paper.ShortID))
	fmt.Fprintf(w, "  Source:          %s\n", sourceLabel(paper))
	fmt.Fprintf(w, "  Published:       %s\n", formatTime(paper.PublishedDate))
	fmt.Fprintf(w, "  Authors:         %s\n", displayValue(authorNames(paper.Authors)))
	fmt.Fprintf(w, "  URL:             %s\n", displayValue(paper.URL))
	fmt.Fprintf(w, "  PDF URL:         %s\n", displayValue(paper.PdfURL))
	fmt.Fprintf(w, "  DOI:             %s\n", displayValue(paper.DOI))

	if strings.TrimSpace(paper.Abstract) != "" {
		fmt.Fprintf(w, "\nAbstract:\n  %s\n", paper.Abstract)
	}
}

func authorNames(authors []api.Author) string {
	if len(authors) == 0 {
		return ""
	}

	names := make([]string, 0, len(authors))
	for _, author := range authors {
		name := strings.TrimSpace(author.Name)
		if name != "" {
			names = append(names, name)
		}
	}

	return strings.Join(names, ", ")
}

func displayValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func formatMetadata(metadata any) string {
	if metadata == nil {
		return ""
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil || string(data) == "null" || string(data) == "{}" {
		return ""
	}

	return string(data)
}

func sourceLabel(paper api.Paper) string {
	if paper.Source != nil {
		name := strings.TrimSpace(paper.Source.Name)
		if name != "" {
			return name
		}
	}
	if paper.SourceID > 0 {
		return fmt.Sprintf("source:%d", paper.SourceID)
	}
	return "unknown"
}

func formatFeedback(feedback *api.Feedback) string {
	if feedback == nil || strings.TrimSpace(feedback.Vote) == "" {
		return "—"
	}
	if feedback.Vote == "downvote" && strings.TrimSpace(feedback.DownvoteReason) != "" {
		return fmt.Sprintf("%s (%s)", feedback.Vote, feedback.DownvoteReason)
	}
	return feedback.Vote
}

func feedbackMarker(feedback *api.Feedback) string {
	if feedback == nil {
		return ""
	}
	switch feedback.Vote {
	case "upvote":
		return "[↑]"
	case "downvote":
		return "[↓]"
	case "star":
		return "[★]"
	default:
		return ""
	}
}

func formatRelevance(class int, score float64) string {
	label := "Related"
	if class == 2 {
		label = "Must Read"
	}
	if score <= 0 {
		return label
	}
	return fmt.Sprintf("%s (%d%%)", label, int(score*100))
}

func firstAuthorSurname(authors []api.Author) string {
	if len(authors) == 0 {
		return "Unknown"
	}

	name := strings.TrimSpace(authors[0].Name)
	if name == "" {
		return "Unknown"
	}

	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "Unknown"
	}

	surname := parts[len(parts)-1]
	if len(authors) > 1 {
		return surname + " et al."
	}
	return surname
}
