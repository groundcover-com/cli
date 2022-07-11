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
	TSDB_SERVICE_CONFIG_NAME = "service/groundcover-tsdb-config"
)

var (
	labelsToDelete = []string{"release=groundcover", "app.kubernetes.io/instance=groundcover"}
)

func init() {
	RootCmd.AddCommand(UninstallCmd)
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

		shouldUninstallPvcs := utils.YesNoPrompt("Do you want to delete groundcover's Persistent Volume Claims? This will remove all of groundcover data", false)
		if !shouldUninstallPvcs {
			fmt.Println("Not removing groundcover pvcs")
			return nil
		}
		return kubectl.DeletePvcByLabels(cmd.Context(), groundcoverNamespace, labelsToDelete)
	},
}
