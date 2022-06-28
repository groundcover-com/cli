package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/kubectl"
	"groundcover.com/pkg/utils"
)

const (
	GROUNDCOVER_HELM_RELEASE_FLAG = "groundcover-release"
	GROUNDCOVER_HELM_RELEASE_NAME = "groundcover"
	TSDB_SERVICE_CONFIG_NAME      = "service/groundcover-tsdb-config"
)

func init() {
	RootCmd.AddCommand(UninstallCmd)

	UninstallCmd.PersistentFlags().String(GROUNDCOVER_NAMESPACE_FLAG, DEFAULT_GROUNDCOVER_NAMESPACE, "groundcover deployment namespace")
	viper.BindPFlag(GROUNDCOVER_NAMESPACE_FLAG, UninstallCmd.PersistentFlags().Lookup(GROUNDCOVER_NAMESPACE_FLAG))

	UninstallCmd.PersistentFlags().String(GROUNDCOVER_HELM_RELEASE_FLAG, GROUNDCOVER_HELM_RELEASE_NAME, "groundcover release name")
	viper.BindPFlag(GROUNDCOVER_HELM_RELEASE_FLAG, UninstallCmd.PersistentFlags().Lookup(GROUNDCOVER_HELM_RELEASE_FLAG))
}

var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall groundcover",
	RunE: func(cmd *cobra.Command, args []string) error {
		groundcoverNamespace := viper.GetString(GROUNDCOVER_NAMESPACE_FLAG)
		fmt.Printf("Uninstalling groundcover with namespace: '%s'\n", groundcoverNamespace)

		uninstall := utils.YesNoPrompt("Are you sure you want to uninstall groundcover?", false)
		if !uninstall {
			fmt.Println("Not uninstalling groundcover :)")
			return nil
		}

		helmCmd, err := helm.NewHelmCmd()
		if err != nil {
			return err
		}

		err = helmCmd.Uninstall(cmd.Context(), groundcoverNamespace, viper.GetString(GROUNDCOVER_HELM_RELEASE_FLAG))
		if err != nil {
			return err
		}

		err = kubectl.Delete(cmd.Context(), groundcoverNamespace, TSDB_SERVICE_CONFIG_NAME)
		if err != nil {
			return err
		}

		fmt.Println("Uninstalled groundcover :(")
		return nil
	},
}
