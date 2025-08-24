package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var (
	// Version is the current version of align, made available for use through
	// out the application.
	Version string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version",
	Long:  `Display the version of the CLI.`,
	Run:   versionFunc,
}

func versionFunc(cmd *cobra.Command, args []string) {
	if Version == "" {
		fmt.Printf("Align version dev-local\n")
	} else {
		fmt.Printf("Align version %v\n", Version)
	}
}
