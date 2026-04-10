package cmd

import (
	"fmt"
	"io"

	"github.com/paperzilla/pz/internal/api"
)

func writeProjectPaperFeedList(w io.Writer, items []api.ProjectPaper) {
	for _, p := range items {
		prefix := "○ Related"
		if p.RelevanceClass == 2 {
			prefix = "★ Must Read"
		}
		if marker := feedbackMarker(p.Feedback); marker != "" {
			prefix += " " + marker
		}

		title := p.PaperTitle
		if len(title) > 80 {
			title = title[:77] + "..."
		}

		fmt.Fprintf(w, "%s  %s\n", prefix, title)

		surname := firstAuthorSurname(p.Paper.Authors)
		date := formatTime(p.Paper.PublishedDate)
		score := int(p.RelevanceScore * 100)
		meta := joinDisplayParts(
			surname,
			paperListLabel(p.Paper),
			date,
			fmt.Sprintf("relevance: %d%%", score),
		)

		fmt.Fprintf(w, "  %s\n\n", meta)
	}
}
