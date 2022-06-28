package cmd

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"
)

var (
	// this is a placeholder value which will be overriden by the build process
	BinaryVersion = "unknown"
)

func init() {
	RootCmd.AddCommand(VersionCmd)
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get groundcover cli version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("groundcover cli version: %s\n", BinaryVersion)
		return nil
	},
}

func GetVersion() (semver.Version, error) {
	return semver.Parse(BinaryVersion)
}

func IsDevVersion() bool {
	_, err := semver.Parse(BinaryVersion)
	if err != nil {
		return false
	}

	return true
}
