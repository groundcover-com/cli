package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/getsentry/sentry-go"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/selfupdate"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/ui"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/strings/slices"
)

const (
	GITHUB_REPO          = "cli"
	GITHUB_OWNER         = "groundcover-com"
	NAMESPACE_FLAG       = "namespace"
	KUBECONFIG_FLAG      = "kubeconfig"
	KUBECONTEXT_FLAG     = "kube-context"
	HELM_RELEASE_FLAG    = "release-name"
	CLUSTER_NAME_FLAG    = "cluster-name"
	SKIP_CLI_UPDATE_FLAG = "skip-cli-update"
)

var (
	JOIN_SLACK_LINK       = ui.UrlLink("https://groundcover.com/join-slack")
	SUPPORT_SLACK_MESSAGE = fmt.Sprintf("questions? issues? ping us anytime %s", JOIN_SLACK_LINK)
	JOIN_SLACK_MESSAGE    = fmt.Sprintf("join us on slack, we promise to keep things interesting %s", JOIN_SLACK_LINK)
)

func init() {
	home := homedir.HomeDir()

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

var skipAuthCommandNames = []string{
	"help",
	LoginCmd.Name(),
	VersionCmd.Name(),
}

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

		ctx := cmd.Context()

		sentry_utils.SetTransactionOnCurrentScope(cmd.Name())

		if err = validateAuthentication(cmd, args); err != nil {
			return err
		}

		if !viper.GetBool(SKIP_CLI_UPDATE_FLAG) {
			if shouldUpdate, selfUpdater := checkLatestVersionUpdate(cmd.Context()); shouldUpdate {
				if err = selfUpdater.Apply(ctx); err != nil {
					return err
				}
				sentry.CaptureMessage("cli-update executed successfully")
				os.Exit(0)
			}
		}

		return nil
	},
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
	shouldUpdate := ui.YesNoPrompt(fmt.Sprintf(promptFormat, currentVersion, selfUpdater.Version), true)

	if shouldUpdate {
		sentry_utils.SetTransactionOnCurrentScope(sentry_utils.SELF_UPDATE_CONTEXT_NAME)
		sentryContext := sentry_utils.NewSelfUpdateContext(currentVersion, selfUpdater.Version)
		sentryContext.SetOnCurrentScope()
	}

	return shouldUpdate, selfUpdater
}

func validateAuthentication(cmd *cobra.Command, args []string) error {
	var err error

	if slices.Contains(skipAuthCommandNames, cmd.Name()) {
		return nil
	}

	fmt.Println("Validating groundcover authentication:")

	err = validateAuth0Token()

	if err == nil {
		ui.PrintSuccessMessage("Device authentication is valid")
		return nil
	}

	if !ui.YesNoPrompt("authentication is required, do you want to login?", true) {
		os.Exit(0)
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTransaction(LoginCmd.Name())
		err = runLoginCmd(cmd, args)
	})

	return err
}

func ExecuteContext(ctx context.Context) error {
	err := RootCmd.ExecuteContext(ctx)

	if err == nil {
		return nil
	}

	if strings.HasPrefix(err.Error(), "unknown command") {
		ui.PrintErrorMessageln(err.Error())
		return nil
	}

	sentry.CaptureException(err)
	ui.PrintErrorMessageln(err.Error())
	fmt.Printf("\n%s\n", SUPPORT_SLACK_MESSAGE)
	return err
}

func validateAuth0Token() error {
	var err error

	var auth0Token auth.Auth0Token
	err = auth0Token.Load()

	if errors.Is(err, jwt.ErrTokenExpired) {
		err = auth0Token.RefreshAndSave()
	}

	if err != nil {
		return err
	}

	sentry_utils.SetUserOnCurrentScope(sentry.User{Email: auth0Token.Claims.Email})
	sentry_utils.SetTagOnCurrentScope(sentry_utils.ORGANIZATION_TAG, auth0Token.Claims.Org)
	return nil
}
