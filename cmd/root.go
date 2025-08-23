package cmd

import (
	"fmt"
	"os"

	"github.com/gomicro/align/client"
	"github.com/gomicro/align/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clt *client.Client
)

func init() {
	cobra.OnInitialize(initEnvs)

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more verbose output")

	err := viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		fmt.Printf("Error setting up: %s\n", err)
		os.Exit(1)
	}
}

func initEnvs() {
}

var RootCmd = &cobra.Command{
	Use:   "align [flags]",
	Short: "Tool for managing repos",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupClient(cmd *cobra.Command, args []string) {
	c, err := config.ParseFromFile()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	clt, err = client.New(c)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
