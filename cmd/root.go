package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dryRun bool
)

func init() {
	cobra.OnInitialize(initEnvs)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more verbose output")
	rootCmd.PersistentFlags().BoolP("dryRun", "d", false, "attempt the specified command without actually making live changes")

	err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		fmt.Printf("Error setting up: %v\n", err.Error())
		os.Exit(1)
	}

	err = viper.BindPFlag("dryRun", rootCmd.PersistentFlags().Lookup("dryRun"))
	if err != nil {
		fmt.Printf("Error setting up: %v\n", err.Error())
		os.Exit(1)
	}
}

func initEnvs() {
}

var rootCmd = &cobra.Command{
	Use:   "align [flags]",
	Short: "Tool for managing repos",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Failed to execute: %v\n", err.Error())
		os.Exit(1)
	}
}