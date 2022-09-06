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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/selfupdate"
	sentry_utils "groundcover.com/pkg/sentry"
	"groundcover.com/pkg/utils"
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

func init() {
	home := homedir.HomeDir()

	RootCmd.PersistentFlags().Bool(utils.ASSUME_YES_FLAG, false, "assume yes on interactive prompts")
	viper.BindPFlag(utils.ASSUME_YES_FLAG, RootCmd.PersistentFlags().Lookup(utils.ASSUME_YES_FLAG))

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
	
	groundcover, more data at: https://groundcover.com/docs`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		sentry_utils.SetTransactionOnCurrentScope(cmd.Name())

		if err = checkAuthForCmd(cmd); err != nil {
			return err
		}

		if !viper.GetBool(SKIP_CLI_UPDATE_FLAG) {
			if shouldUpdate, selfUpdater := checkLatestVersionUpdate(cmd.Context()); shouldUpdate {
				if err = selfUpdater.Apply(); err != nil {
					fmt.Println("cli update has failed")
					return err
				}
				fmt.Println("cli update was successfully")
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

	promptFormat := "Your version %s is out of date! The latest version is %s.\nDo you want to update?"
	shouldUpdate := utils.YesNoPrompt(fmt.Sprintf(promptFormat, currentVersion, selfUpdater.Version), true)

	if shouldUpdate {
		sentry_utils.SetTransactionOnCurrentScope(sentry_utils.SELF_UPDATE_CONTEXT_NAME)
		sentryContext := sentry_utils.NewSelfUpdateContext(currentVersion, selfUpdater.Version)
		sentryContext.SetOnCurrentScope()
	}

	return shouldUpdate, selfUpdater
}

func checkAuthForCmd(cmd *cobra.Command) error {
	if slices.Contains(skipAuthCommandNames, cmd.Name()) {
		return nil
	}

	if err := validateAuth0Token(); err != nil {
		return errors.Wrap(err, "failed to authenticate. Please retry `groundcover login`")
	}

	return nil
}

func Execute() error {
	err := RootCmd.Execute()

	if err == nil {
		return nil
	}

	if strings.HasPrefix(err.Error(), "unknown command") {
		logrus.Error(err)
		return nil
	}

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
