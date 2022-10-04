package cmd

import (
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/utils"
)

func init() {
	AuthCmd.AddCommand(LoginCmd)
	RootCmd.AddCommand(LoginCmd)
}

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to groundcover",
	RunE:  runLoginCmd,
}

func runLoginCmd(cmd *cobra.Command, args []string) error {
	var err error

	var auth0Token *auth.Auth0Token
	if auth0Token, err = attemptAuth0Login(); err != nil {
		return errors.Wrap(err, "failed to login")
	}

	sentry_utils.SetUserOnCurrentScope(sentry.User{Email: auth0Token.Claims.Email})
	sentry_utils.SetTagOnCurrentScope(sentry_utils.ORGANIZATION_TAG, auth0Token.Claims.Org)

	if err = fetchAndSaveApiKey(auth0Token); err != nil {
		return errors.Wrap(err, "failed to fetch api key")
	}

	return nil
}

func attemptAuth0Login() (*auth.Auth0Token, error) {
	var err error

	var deviceCode auth.DeviceCode
	if err = deviceCode.Fetch(); err != nil {
		return nil, err
	}

	utils.TryOpenBrowser("Browse to:", deviceCode.VerificationURIComplete)

	var auth0Token auth.Auth0Token
	if err = deviceCode.PollToken(&auth0Token); err != nil {
		return nil, err
	}

	if err = auth0Token.Save(); err != nil {
		return nil, err
	}

	return &auth0Token, err
}

func fetchAndSaveApiKey(auth0Token *auth.Auth0Token) error {
	var err error

	apiClient := api.NewClient(auth0Token)

	var apiKey *api.ApiKey
	if apiKey, err = apiClient.ApiKey(); err != nil {
		return err
	}

	if err = apiKey.Save(); err != nil {
		return err
	}

	return nil
}
