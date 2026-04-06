package cmd

import (
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
	if label := paperDetailLabel(paper); label != "" {
		fmt.Fprintf(w, "Source:          %s\n", label)
	}
	fmt.Fprintf(w, "Published:       %s\n", formatTime(paper.PublishedDate))
	fmt.Fprintf(w, "Authors:         %s\n", displayValue(authorNames(paper.Authors)))
	fmt.Fprintf(w, "URL:             %s\n", displayValue(paper.URL))
	fmt.Fprintf(w, "PDF URL:         %s\n", displayValue(paper.PdfURL))
	fmt.Fprintf(w, "DOI:             %s\n", displayValue(paper.DOI))

	if strings.TrimSpace(paper.Abstract) != "" {
		fmt.Fprintf(w, "\nAbstract:\n  %s\n", paper.Abstract)
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
	if label := paperDetailLabel(paper); label != "" {
		fmt.Fprintf(w, "  Source:          %s\n", label)
	}
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

func paperListLabel(paper api.Paper) string {
	return strings.TrimSpace(paper.VenueName)
}

func paperDetailLabel(paper api.Paper) string {
	if label := strings.TrimSpace(paper.ReferenceLabel); label != "" {
		return label
	}
	return paperListLabel(paper)
}

func joinDisplayParts(parts ...string) string {
	display := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		display = append(display, part)
	}
	return strings.Join(display, " · ")
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
