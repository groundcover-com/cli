package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/auth"
	cs "groundcover.com/pkg/custom_sentry"
)

const (
	MANUAL_FLAG = "manual"
)

var (
	authTimeout = time.Second * 30
	ManualLogin bool
)

func init() {
	RootCmd.AddCommand(LoginCmd)

	LoginCmd.PersistentFlags().Bool(MANUAL_FLAG, false, "Open your browser manualy using the auth url")
	viper.BindPFlag(MANUAL_FLAG, LoginCmd.PersistentFlags().Lookup(MANUAL_FLAG))
}

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to groundcover",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), authTimeout)
		defer cancel()

		err := auth.Login(ctx, viper.GetBool(MANUAL_FLAG))
		if err != nil {
			return err
		}

		customClaims, err := auth.FetchAndSaveApiKey()
		if err != nil {
			return err
		}

		cs.CaptureLoginEvent(customClaims)
		return nil
	},
}
