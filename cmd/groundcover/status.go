package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"groundcover.com/pkg/helm"
	"groundcover.com/pkg/k8s"
)

const (
	WAITING_ALL_NODES_MONITORED_MSG = " Waiting until all nodes are monitored. %d/%d"
	WAIT_FOR_ALLIGATORS_TO_RUN      = time.Minute * 2
	ALLIGATORS_POLLING_INTERVAL     = time.Second * 10
	SPINNER_TYPE                    = 8 // .oO@*
)

func init() {
	RootCmd.AddCommand(StatusCmd)

	StatusCmd.PersistentFlags().String(GROUNDCOVER_NAMESPACE_FLAG, DEFAULT_GROUNDCOVER_NAMESPACE, "groundcover deployment namespace")
	viper.BindPFlag(GROUNDCOVER_NAMESPACE_FLAG, StatusCmd.PersistentFlags().Lookup(GROUNDCOVER_NAMESPACE_FLAG))
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get groundcover current status",
	RunE: func(cmd *cobra.Command, args []string) error {
		helmCmd, err := helm.NewHelmCmd()
		if err != nil {
			return err
		}

		version, err := helmCmd.GetLatestChartVersion(cmd.Context())
		if err != nil {
			return err
		}

		metadataFetcher, err := k8s.NewMetadataFetcher(viper.GetString(KUBECONFIG_PATH_FLAG))
		if err != nil {
			return err
		}

		err = waitForAlligators(cmd.Context(), metadataFetcher, viper.GetString(GROUNDCOVER_NAMESPACE_FLAG), version)
		if err != nil {
			return fmt.Errorf("failed while waiting for all nodes to be monitored: %s", err.Error())
		}

		return nil
	},
}

func waitForAlligators(ctx context.Context, metadataFetcher *k8s.MetadataFetcher, groundcoverNamespace string, version string) error {
	ctx, cancel := context.WithTimeout(ctx, WAIT_FOR_ALLIGATORS_TO_RUN)
	defer cancel()

	numberOfNodes, err := metadataFetcher.GetNumberOfNodes(ctx)
	if err != nil {
		return err
	}

	var numberOfAlligators int
	s := spinner.New(spinner.CharSets[SPINNER_TYPE], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(WAITING_ALL_NODES_MONITORED_MSG, numberOfAlligators, numberOfNodes)
	s.Color("red")
	s.Start()
	defer s.Stop()

	ticker := time.NewTicker(ALLIGATORS_POLLING_INTERVAL)
	for {
		select {
		case <-ticker.C:
			numberOfAlligators, err = metadataFetcher.GetNumberOfAlligators(ctx, groundcoverNamespace, version)
			if err != nil {
				return err
			}
			if numberOfAlligators == numberOfNodes {
				s.Stop()
				fmt.Printf("All nodes are monitored %d/%d !\n", numberOfAlligators, numberOfNodes)
				return nil
			}

			s.Suffix = fmt.Sprintf(WAITING_ALL_NODES_MONITORED_MSG, numberOfAlligators, numberOfNodes)
		case <-ctx.Done():
			fmt.Printf("timed out while waiting for all nodes to be monitored, got only: %d/%d", numberOfAlligators, numberOfNodes)
			return nil
		}
	}
}
