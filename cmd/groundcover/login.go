package cmd

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/auth"
	sentry_utils "groundcover.com/pkg/sentry"
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

		ctx, cancel := context.WithTimeout(cmd.Context(), authTimeout)
		defer cancel()

		if err = auth.Login(ctx); err != nil {
			return err
		}

		if err = setAndValidateApiKey(); err != nil {
			return err
		}

		sentry.CaptureMessage("login executed successfully")
		return nil
	},
}

func setAndValidateApiKey() error {
	var err error

	var customClaims *auth.CustomClaims
	if customClaims, err = auth.FetchAndSaveApiKey(); err != nil {
		return err
	}

	sentry_utils.SetUserOnCurrentScope(sentry.User{Email: customClaims.Email})
	sentry_utils.SetTagOnCurrentScope(sentry_utils.ORGANIZATION_TAG, customClaims.Org)
	return nil
}
