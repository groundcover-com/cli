package cmd

import (
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage groundcover auth credentials",
}

func init() {
	RootCmd.AddCommand(authCmd)
}
