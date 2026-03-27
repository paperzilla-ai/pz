package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

func init() {
	paperCmd.Flags().BoolP("json", "j", false, "Output as JSON")
}

var paperCmd = &cobra.Command{
	Use:   "paper <id>",
	Short: "Show details for a paper or feed item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens, err := loadAuth()
		if err != nil {
			return err
		}

		paper, err := withAuth(&tokens, func(at string) (api.Paper, error) {
			return api.FetchPaper(at, args[0])
		})
		if err != nil {
			return fmt.Errorf("failed to fetch paper: %w", err)
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			data, err := json.MarshalIndent(paper, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Title:           %s\n", displayValue(paper.Title))
		fmt.Printf("ID:              %s\n", displayValue(paper.ID))
		fmt.Printf("Short ID:        %s\n", displayValue(paper.ShortID))
		fmt.Printf("Slug:            %s\n", displayValue(paper.Slug))
		fmt.Printf("Source:          %s\n", sourceLabel(paper))
		fmt.Printf("Published:       %s\n", formatTime(paper.PublishedDate))
		fmt.Printf("Authors:         %s\n", displayValue(authorNames(paper.Authors)))
		fmt.Printf("URL:             %s\n", displayValue(paper.URL))
		fmt.Printf("PDF URL:         %s\n", displayValue(paper.PdfURL))
		fmt.Printf("DOI:             %s\n", displayValue(paper.DOI))
		fmt.Printf("Source Paper ID: %s\n", displayValue(paper.SourcePaperID))

		if strings.TrimSpace(paper.Abstract) != "" {
			fmt.Printf("\nAbstract:\n  %s\n", paper.Abstract)
		}

		metadata := formatMetadata(paper.Metadata)
		if metadata != "" {
			fmt.Printf("\nMetadata:\n%s\n", metadata)
		}

		return nil
	},
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
