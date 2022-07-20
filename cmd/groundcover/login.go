package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"groundcover.com/pkg/auth"
	sentry "groundcover.com/pkg/custom_sentry"
)

var (
	authTimeout = time.Second * 30
)

func init() {
	RootCmd.AddCommand(LoginCmd)
}

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to groundcover",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var customClaims *auth.CustomClaims

		ctx, cancel := context.WithTimeout(cmd.Context(), authTimeout)
		defer cancel()

		if err = auth.Login(ctx); err != nil {
			return err
		}

		if customClaims, err = auth.FetchAndSaveApiKey(); err != nil {
			return err
		}

		sentry.CaptureLoginEvent(customClaims)
		return nil
	},
}
