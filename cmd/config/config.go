package configcmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gomicro/align/cmd"
	"github.com/gomicro/align/config"
	"github.com/spf13/cobra"
)

func init() {
	cmd.RootCmd.AddCommand(configCmd)
}

var configValidArgs = []string{
	"github.limits.burst\tburstable rate for the github client",
	"github.limits.requests_per_second\tmaximum requests per second for the github client",
}

var configCmd = &cobra.Command{
	Use:       "config [config_field] [value]",
	Short:     "config align",
	Long:      `configure align`,
	Args:      cobra.ExactArgs(2),
	Run:       configFunc,
	ValidArgs: configValidArgs,
}

func configFunc(cmd *cobra.Command, args []string) {
	field := args[0]
	value := args[1]

	confFile, err := config.ParseFromFile()
	if err != nil {
		fmt.Printf("error: %v", err.Error())
	}

	switch strings.ToLower(field) {
	case "github.limits.burst":
		burst, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("invalid value provided for burst")
			os.Exit(1)
		}

		confFile.Github.Limits.Burst = burst
	case "github.limits.requests_per_second":
		rps, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("invalid value provided for requests per second")
			os.Exit(1)
		}

		confFile.Github.Limits.RequestsPerSecond = rps
	default:
		fmt.Printf("unreconized config field: %v\n", field)
		os.Exit(1)
	}

	err = confFile.WriteFile()
	if err != nil {
		fmt.Printf("error: %v", err.Error())
		os.Exit(1)
	}

	fmt.Println("Config file updated")
}
