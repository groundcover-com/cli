package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/auth"
	"groundcover.com/pkg/selfupdate"
)

const (
	USER_CUSTOM_CLAIMS_KEY = "user_custom_claims"
)

func init() {
	RootCmd.PersistentFlags().Bool(selfupdate.SKIP_SELFUPDATE_FLAG, false, "disable automatic selfupdate check")
	viper.BindPFlag(selfupdate.SKIP_SELFUPDATE_FLAG, RootCmd.PersistentFlags().Lookup(selfupdate.SKIP_SELFUPDATE_FLAG))

	RootCmd.PersistentFlags().String(KUBECONFIG_PATH_FLAG, "", "kubeconfig path")
	viper.BindPFlag(KUBECONFIG_PATH_FLAG, RootCmd.PersistentFlags().Lookup(KUBECONFIG_PATH_FLAG))

	RootCmd.PersistentFlags().String(GROUNDCOVER_NAMESPACE_FLAG, DEFAULT_GROUNDCOVER_NAMESPACE, "groundcover deployment namespace")
	viper.BindPFlag(GROUNDCOVER_NAMESPACE_FLAG, RootCmd.PersistentFlags().Lookup(GROUNDCOVER_NAMESPACE_FLAG))
}

var RootCmd = &cobra.Command{
	Use:   "groundcover",
	Short: "groundcover cli",
	Long: `
                                   _                         
    __ _ _ __ ___  _   _ _ __   __| | ___ _____   _____ _ __ 
   / _` + "`" + ` | '__/ _ \| | | | '_ \ / _` + "`" + ` |/ __/ _ \ \ / / _ \ '__|
  | (_| | | | (_) | |_| | | | | (_| | (_| (_) \ V /  __/ |   
   \__, |_|  \___/ \__,_|_| |_|\__,_|\___\___/ \_/ \___|_|   
   |___/                                                     

groundcover, more data at: https://groundcover.com/docs`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		currentVersion, err := GetVersion()
		if !viper.GetBool(selfupdate.SKIP_SELFUPDATE_FLAG) && err == nil {
			if err = selfupdate.TrySelfUpdate(context.Background(), currentVersion); err != nil {
				return err
			}
		}
		customClaims, err := checkAuthForCmd(cmd)
		if err != nil {
			return fmt.Errorf("failed to authenticate. Please retry `groundcover login`")
		}

		ctx := context.WithValue(cmd.Context(), USER_CUSTOM_CLAIMS_KEY, customClaims)
		cmd.SetContext(ctx)
		return nil
	},
	// this mutes usage printing on command errors
	SilenceUsage: true,
	// this mutes error printing on command errors
	SilenceErrors: true,
}

func checkAuthForCmd(c *cobra.Command) (*auth.CustomClaims, error) {
	// here we need to check if the command requires auth, currently we only check for the login command
	switch c {
	case LoginCmd:
		// skip IsAuthenticated
		return nil, nil
	default:
		return auth.FetchAndSaveApiKey()
	}
}

func Execute() error {
	return RootCmd.Execute()
}
