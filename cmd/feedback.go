package cmd

import (
	"fmt"
	"strings"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

func init() {
	feedbackCmd.Flags().String("reason", "", "Optional downvote reason (not_relevant or low_quality)")
	feedbackCmd.AddCommand(feedbackClearCmd)
}

var feedbackCmd = &cobra.Command{
	Use:   "feedback <project-paper-ref> <upvote|downvote|star>",
	Short: "Set recommendation feedback for one project paper",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		if len(args) != 2 {
			return fmt.Errorf("accepts 2 arg(s), received %d", len(args))
		}

		vote := strings.TrimSpace(args[1])
		switch vote {
		case "upvote", "downvote", "star":
		default:
			return fmt.Errorf("invalid feedback %q (expected upvote, downvote, or star)", vote)
		}

		reason, _ := cmd.Flags().GetString("reason")
		if reason != "" && vote != "downvote" {
			return fmt.Errorf("--reason is only allowed with downvote")
		}

		tokens, err := loadRequiredAuth()
		if err != nil {
			return err
		}

		feedback, err := withAuth(&tokens, func(at string) (api.Feedback, error) {
			return api.SetProjectPaperFeedback(at, args[0], vote, reason)
		})
		if err != nil {
			return fmt.Errorf("failed to set feedback: %w", err)
		}

		message := fmt.Sprintf("Feedback set: %s", feedback.Vote)
		if feedback.Vote == "downvote" && strings.TrimSpace(feedback.DownvoteReason) != "" {
			message += fmt.Sprintf(" (%s)", feedback.DownvoteReason)
		}
		fmt.Fprintln(cmd.OutOrStdout(), message)
		return nil
	},
}

var feedbackClearCmd = &cobra.Command{
	Use:   "clear <project-paper-ref>",
	Short: "Clear recommendation feedback for one project paper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens, err := loadRequiredAuth()
		if err != nil {
			return err
		}

		if _, err := withAuth(&tokens, func(at string) (struct{}, error) {
			return struct{}{}, api.ClearProjectPaperFeedback(at, args[0])
		}); err != nil {
			return fmt.Errorf("failed to clear feedback: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Feedback cleared.")
		return nil
	},
}
