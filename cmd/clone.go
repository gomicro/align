package cmd

import (
	"context"
	"fmt"

	"github.com/gomicro/align/client"
	"github.com/google/go-github/github"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	topics []string
)

func init() {
	RootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVar(&dir, "dir", ".", "directory to clone repos into")
	cloneCmd.Flags().StringSliceVarP(&topics, "topics", "t", []string{}, "clone only repos with matching topics")
}

var cloneCmd = &cobra.Command{
	Use:               "clone [user|org]",
	Short:             "Clone all active repos from an org or user.",
	Long:              `Clone all active repos from an org or user.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: createCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              cloneFunc,
}

func cloneFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := client.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	repos, err := clt.GetRepos(ctx, name)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get repos: %w", err)
	}

	repos = filterByTopics(repos, topics)

	ctx = client.WithRepos(ctx, repos)

	_, err = clt.CloneRepos(ctx, dir)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("clone repos: %w", err)
	}

	return nil
}

func createCmdValidArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	setupClient(cmd, args)

	valid, err := clt.GetLogins(context.Background())
	if err != nil {
		valid = []string{"error fetching"}
	}

	return valid, cobra.ShellCompDirectiveNoFileComp
}

func filterByTopics(repos []*github.Repository, topics []string) []*github.Repository {
	if len(topics) == 0 {
		return repos
	}

	var filtered []*github.Repository

	for _, r := range repos {
		searchMap := make(map[string]struct{})
		for i := 0; i < len(r.Topics); i++ {
			searchMap[r.Topics[i]] = struct{}{}
		}

		allFound := true
		for _, t := range topics {
			if _, ok := searchMap[t]; !ok {
				allFound = false
				break
			}
		}

		if allFound {
			filtered = append(filtered, r)
		}
	}

	return filtered
}
