package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/segment"
	"groundcover.com/pkg/selfupdate"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/strings/slices"
)

const (
	GITHUB_REPO           = "cli"
	GITHUB_OWNER          = "groundcover-com"
	TOKEN_FLAG            = "token"
	API_KEY_FLAG          = "api-key"
	TENANT_UUID_FLAG      = "tenant-uuid"
	NAMESPACE_FLAG        = "namespace"
	KUBECONFIG_FLAG       = "kubeconfig"
	KUBECONTEXT_FLAG      = "kube-context"
	HELM_RELEASE_FLAG     = "release-name"
	CLUSTER_NAME_FLAG     = "cluster-name"
	SKIP_CLI_UPDATE_FLAG  = "skip-cli-update"
	INSTALLATION_ID_FLAG  = "installation-id"
	INVALID_TOKEN_MESSAGE = "Issue with authentication - try again to copy command line and rerun"
)

var (
	JOIN_SLACK_LINK       = ui.GlobalWriter.UrlLink("https://groundcover.com/join-slack")
	SUPPORT_SLACK_MESSAGE = fmt.Sprintf("questions? issues? ping us anytime %s", JOIN_SLACK_LINK)
	JOIN_SLACK_MESSAGE    = fmt.Sprintf("join us on slack, we promise to keep things interesting %s", JOIN_SLACK_LINK)
)

func init() {
	home := homedir.HomeDir()

	RootCmd.PersistentFlags().String(API_KEY_FLAG, "", "optional api-key")
	viper.BindPFlag(API_KEY_FLAG, RootCmd.PersistentFlags().Lookup(API_KEY_FLAG))

	RootCmd.PersistentFlags().String(TENANT_UUID_FLAG, "", "optional tenant-uuid")
	viper.BindPFlag(TENANT_UUID_FLAG, RootCmd.PersistentFlags().Lookup(TENANT_UUID_FLAG))

	RootCmd.PersistentFlags().String(TOKEN_FLAG, "", "optional login token")
	viper.BindPFlag(TOKEN_FLAG, RootCmd.PersistentFlags().Lookup(TOKEN_FLAG))

	RootCmd.PersistentFlags().Bool(ui.ASSUME_YES_FLAG, false, "assume yes on interactive prompts")
	viper.BindPFlag(ui.ASSUME_YES_FLAG, RootCmd.PersistentFlags().Lookup(ui.ASSUME_YES_FLAG))

	RootCmd.PersistentFlags().Bool(SKIP_CLI_UPDATE_FLAG, false, "disable automatic cli update check")
	viper.BindPFlag(SKIP_CLI_UPDATE_FLAG, RootCmd.PersistentFlags().Lookup(SKIP_CLI_UPDATE_FLAG))

	RootCmd.PersistentFlags().String(CLUSTER_NAME_FLAG, "", "cluster name")
	viper.BindPFlag(CLUSTER_NAME_FLAG, RootCmd.PersistentFlags().Lookup(CLUSTER_NAME_FLAG))

	RootCmd.PersistentFlags().String(KUBECONTEXT_FLAG, "", "name of the kubeconfig context to use")
	viper.BindPFlag(KUBECONTEXT_FLAG, RootCmd.PersistentFlags().Lookup(KUBECONTEXT_FLAG))

	RootCmd.PersistentFlags().String(KUBECONFIG_FLAG, filepath.Join(home, ".kube", "config"), "path to the kubeconfig file")
	viper.BindPFlag(KUBECONFIG_FLAG, RootCmd.PersistentFlags().Lookup(KUBECONFIG_FLAG))
	viper.BindEnv(KUBECONFIG_FLAG)

	RootCmd.PersistentFlags().String(NAMESPACE_FLAG, DEFAULT_GROUNDCOVER_NAMESPACE, "groundcover deployment namespace")
	viper.BindPFlag(NAMESPACE_FLAG, RootCmd.PersistentFlags().Lookup(NAMESPACE_FLAG))

	RootCmd.PersistentFlags().String(HELM_RELEASE_FLAG, DEFAULT_GROUNDCOVER_RELEASE, "groundcover chart release name")
	viper.BindPFlag(HELM_RELEASE_FLAG, RootCmd.PersistentFlags().Lookup(HELM_RELEASE_FLAG))
}

var (
	skipAuthCommandNames = []string{
		"help",
		LoginCmd.Name(),
		VersionCmd.Name(),
	}

	ErrExecutionAborted        = errors.New("execution aborted")
	ErrSilentExecutionAbort    = errors.New("silent execution abort")
	ErrExecutionPartialSuccess = errors.New("execution partial success")
)

var RootCmd = &cobra.Command{
	SilenceUsage:      true,
	SilenceErrors:     true,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	Use:               "groundcover",
	Short:             "groundcover cli",
	Long: `
                                   _                         
    __ _ _ __ ___  _   _ _ __   __| | ___ _____   _____ _ __ 
   / _` + "`" + ` | '__/ _ \| | | | '_ \ / _` + "`" + ` |/ __/ _ \ \ / / _ \ '__|
  | (_| | | | (_) | |_| | | | | (_| | (_| (_) \ V /  __/ |   
   \__, |_|  \___/ \__,_|_| |_|\__,_|\___\___/ \_/ \___|_|   
   |___/                                                     

groundcover, more data at: https://docs.groundcover.com/docs`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		segment.SetScope(cmd.Name())
		sentry_utils.SetTransactionOnCurrentScope(cmd.Name())

		event := segment.NewEvent(cmd.Name())
		defer event.Start()

		if err = validateAuthentication(cmd, args); err != nil {
			return err
		}

		if !viper.GetBool(SKIP_CLI_UPDATE_FLAG) {
			return checkAndUpgradeVersion(cmd.Context())
		}

		return nil
	},
}

func checkAndUpgradeVersion(ctx context.Context) error {
	if shouldUpdate, selfUpdater := checkLatestVersionUpdate(ctx); shouldUpdate {
		if err := selfUpdater.Apply(ctx); err != nil {
			return err
		}
		command := strings.Join(os.Args, " ")
		ui.GlobalWriter.PrintWarningMessage(fmt.Sprintf("Please re-run %s\n", command))
		sentry.CaptureMessage("cli-update executed successfully")
		return ErrSilentExecutionAbort
	}

	return nil
}

func checkLatestVersionUpdate(ctx context.Context) (bool, *selfupdate.SelfUpdater) {
	var err error
	var currentVersion semver.Version
	var selfUpdater *selfupdate.SelfUpdater

	if currentVersion, err = GetVersion(); err != nil {
		sentry.CaptureException(err)
		return false, nil
	}

	if selfUpdater, err = selfupdate.NewSelfUpdater(ctx, GITHUB_OWNER, GITHUB_REPO); err != nil {
		sentry.CaptureException(err)
		return false, nil
	}

	if !selfUpdater.IsLatestNewer(currentVersion) || selfUpdater.IsDevVersion(currentVersion) {
		return false, nil
	}

	promptFormat := "Your groundcover cli version %s is out of date! The latest cli version is %s. Do you want to update your cli?"
	shouldUpdate := ui.GlobalWriter.YesNoPrompt(fmt.Sprintf(promptFormat, currentVersion, selfUpdater.Version), true)

	if shouldUpdate {
		sentry_utils.SetTransactionOnCurrentScope(sentry_utils.SELF_UPDATE_CONTEXT_NAME)
		sentryContext := sentry_utils.NewSelfUpdateContext(currentVersion, selfUpdater.Version)
		sentryContext.SetOnCurrentScope()
	}

	return shouldUpdate, selfUpdater
}

func validateAuthentication(cmd *cobra.Command, args []string) error {
	var err error

	isAuthenicationRequired := !viper.IsSet(TOKEN_FLAG)

	if slices.Contains(skipAuthCommandNames, cmd.Name()) {
		return nil
	}

	event := segment.NewEvent(AUTHENTICATION_VALIDATION_EVENT_NAME)
	defer func() {
		event.StatusByError(err)
	}()

	ui.GlobalWriter.Println("Validating groundcover authentication:")

	var token auth.Token
	if isAuthenicationRequired {
		event.Set("authType", "auth0")
		if token, err = auth.LoadAuth0Token(); err != nil {
			if ui.GlobalWriter.YesNoPrompt("authentication is required, do you want to login?", true) {
				return runLoginCmd(cmd, args)
			}
			os.Exit(0)
		}

		segment.SetUserId(token.GetEmail())
		ui.GlobalWriter.PrintSuccessMessageln("Device authentication is valid")
	} else {
		event.Set("authType", "installationToken")
		if token, err = validateInstallationToken(); err != nil {
			ui.GlobalWriter.PrintErrorMessageln(INVALID_TOKEN_MESSAGE)
			return err
		}

		segment.SetSessionId(token.GetSessionId())
		segment.NewUser(token.GetEmail(), token.GetOrg())
		ui.GlobalWriter.PrintSuccessMessageln("Token authentication success")
	}

	event.Set("installationId", token.GetId())
	viper.Set(INSTALLATION_ID_FLAG, token.GetId())
	sentry_utils.SetUserOnCurrentScope(sentry.User{Email: token.GetEmail()})
	sentry_utils.SetTagOnCurrentScope(sentry_utils.TOKEN_ID_TAG, token.GetId())
	sentry_utils.SetTagOnCurrentScope(sentry_utils.ORGANIZATION_TAG, token.GetOrg())

	return nil
}

func ExecuteContext(ctx context.Context) error {
	start := time.Now()
	err := RootCmd.ExecuteContext(ctx)

	event := segment.NewEvent(segment.GetScope())
	sentryCommandContext := sentry_utils.NewCommandContext(start)
	sentryCommandContext.SetOnCurrentScope()

	if err == nil {
		event.Success()
		sentry.CaptureMessage(fmt.Sprintf("%s executed successfully", sentryCommandContext.Name))
		return nil
	}

	if errors.Is(err, ErrExecutionPartialSuccess) {
		event.PartialSuccess()
		sentry_utils.SetLevelOnCurrentScope(sentry.LevelWarning)
		sentry.CaptureMessage(fmt.Sprintf("%s execution partial success", sentryCommandContext.Name))
		return nil
	}

	if errors.Is(err, ErrSilentExecutionAbort) {
		event.Abort()
		sentry.CaptureMessage(fmt.Sprintf("%s execution aborted silently", sentryCommandContext.Name))
		return nil
	}

	if errors.Is(err, ErrExecutionAborted) {
		event.Abort()
		sentry.CaptureMessage(fmt.Sprintf("%s execution aborted", sentryCommandContext.Name))
		return nil
	}

	if strings.HasPrefix(err.Error(), "unknown") {
		ui.GlobalWriter.PrintErrorMessageln(err.Error())
		// in case the unknown flag / command is due to an old version of the cli
		checkAndUpgradeVersion(ctx)
		return nil
	}

	ui.GlobalWriter.PrintErrorMessageln(err.Error())
	ui.GlobalWriter.PrintlnWithPrefixln(SUPPORT_SLACK_MESSAGE)

	sentry.CaptureMessage(fmt.Sprintf("%s execution failed - %s", sentryCommandContext.Name, err.Error()))
	event.Failure(err)
	return err
}

func validateInstallationToken() (*auth.InstallationToken, error) {
	var err error

	encodedToken := viper.GetString(TOKEN_FLAG)

	var installationToken *auth.InstallationToken
	if installationToken, err = auth.NewInstallationToken(encodedToken); err != nil {
		return nil, err
	}

	viper.Set(API_KEY_FLAG, installationToken.ApiKey.ApiKey)
	viper.Set(TENANT_UUID_FLAG, installationToken.TenantUUID)
	return installationToken, nil
}
