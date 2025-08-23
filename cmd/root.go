package cmd

import (
	"fmt"
	"os"

	"github.com/gomicro/align/client"
	configcmd "github.com/gomicro/align/cmd/config"
	"github.com/gomicro/align/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clt *client.Client
)

func init() {
	cobra.OnInitialize(initEnvs)
	rootCmd.AddCommand(
		authCmd,
		completionCmd,
		cloneCmd,
		checkoutCmd,
		pullCmd,
		pushCmd,
		versionCmd,

		configcmd.ConfigCmd,
	)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more verbose output")

	err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		fmt.Printf("Error setting up: %s\n", err)
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
