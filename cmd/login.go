package cmd

import (
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with your email via magic link OTP",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := runLogin()
		return err
	},
}
