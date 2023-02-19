package cmd

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"groundcover.com/pkg/api"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/segment"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
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

	ctx := cmd.Context()

	event := segment.NewEvent("authentication")
	event.Set("authType", "auth0")
	err = event.Start()
	defer func() {
		if err != nil {
			event.Failure(err)
		}

		event.UserId = auth0Token.GetEmail()
		event.Success()
	}()

	if auth0Token, err = attemptAuth0Login(ctx); err != nil {
		return errors.Wrap(err, "failed to login")
	}

	segment.NewUser(auth0Token.GetEmail(), auth0Token.GetOrg())
	sentry_utils.SetUserOnCurrentScope(sentry.User{Email: auth0Token.GetEmail()})
	sentry_utils.SetTagOnCurrentScope(sentry_utils.ORGANIZATION_TAG, auth0Token.GetOrg())

	if err = fetchAndSaveApiKey(auth0Token); err != nil {
		return errors.Wrap(err, "failed to fetch api key")
	}

	return nil
}

func attemptAuth0Login(ctx context.Context) (*auth.Auth0Token, error) {
	var err error

	var deviceCode *auth.DeviceCode
	if deviceCode, err = auth.NewDeviceCode(); err != nil {
		return nil, err
	}

	utils.TryOpenBrowser(ui.QuietWriter, "Browse to:", deviceCode.VerificationURIComplete)

	var auth0Token auth.Auth0Token
	if err = deviceCode.PollToken(ctx, &auth0Token); err != nil {
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

	var apiKey *auth.ApiKey
	if apiKey, err = apiClient.ApiKey(); err != nil {
		return err
	}

	if err = apiKey.Save(); err != nil {
		return err
	}

	return nil
}
