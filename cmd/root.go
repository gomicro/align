package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gomicro/align/client"
	cfgCmd "github.com/gomicro/align/cmd/config"
	"github.com/gomicro/align/cmd/remote"
	"github.com/gomicro/align/cmd/stash"
	"github.com/gomicro/align/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clt client.Clienter
)

func init() {
	cobra.OnInitialize(initEnvs)

	RootCmd.AddCommand(cfgCmd.ConfigCmd)
	RootCmd.AddCommand(remote.RemoteCmd)
	RootCmd.AddCommand(stash.StashCmd)

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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := exec.LookPath("git"); err != nil {
			return fmt.Errorf("git is not installed or not on PATH — it is required for align to function")
		}
		return nil
	},
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
