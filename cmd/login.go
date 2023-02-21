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

const (
	AUTHENTICATION_EVENT_NAME            = "authentication"
	AUTHENTICATION_VALIDATION_EVENT_NAME = "authentication_validation"
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

	event := segment.NewEvent(AUTHENTICATION_EVENT_NAME)
	event.Set("authType", "auth0")
	event.Start()
	defer func() {
		event.StatusByError(err)
	}()

	if auth0Token, err = attemptAuth0Login(ctx); err != nil {
		return errors.Wrap(err, "failed to login")
	}

	email := auth0Token.GetEmail()
	org := auth0Token.GetOrg()

	event.UserId = email
	segment.NewUser(email, org)

	sentry_utils.SetUserOnCurrentScope(sentry.User{Email: email})
	sentry_utils.SetTagOnCurrentScope(sentry_utils.ORGANIZATION_TAG, org)

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
