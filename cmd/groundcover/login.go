package cmd

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/utils"
)

func init() {
	RootCmd.AddCommand(LoginCmd)
	authCmd.AddCommand(LoginCmd)
}

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to groundcover",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var deviceCode auth.DeviceCode
		if err = deviceCode.Fetch(); err != nil {
			return err
		}

		utils.TryOpenBrowser(deviceCode.VerificationURIComplete)

		var auth0Token auth.Auth0Token
		if err = deviceCode.PollToken(&auth0Token); err != nil {
			return err
		}

		if err = auth0Token.Save(); err != nil {
			return err
		}

		if err = validateAuth0Token(); err != nil {
			return err
		}

		apiClient := api.NewClient(&auth0Token)

		var apiKey *api.ApiKey
		if apiKey, err = apiClient.ApiKey(); err != nil {
			return err
		}

		if err = apiKey.Save(); err != nil {
			return err
		}

		fmt.Print("You are successfully logged in!\n")
		sentry.CaptureMessage("login executed successfully")
		return nil
	},
}

func validateAuth0Token() error {
	var err error

	var auth0Token auth.Auth0Token
	if err = auth0Token.Load(); err != nil {
		return err
	}

	if err = auth0Token.RefreshAndSave(); err != nil {
		return err
	}

	sentry_utils.SetUserOnCurrentScope(sentry.User{Email: auth0Token.Claims.Email})
	sentry_utils.SetTagOnCurrentScope(sentry_utils.ORGANIZATION_TAG, auth0Token.Claims.Org)
	return nil
}
