package cmd

import (
	"context"
	"fmt"

	ctxhelper "github.com/gomicro/align/client/context"
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

	cloneCmd.Flags().StringSliceVarP(&topics, "topics", "t", []string{}, "clone only repos with matching topics")
}

var cloneCmd = &cobra.Command{
	Use:               "clone [user|org] [directory]",
	Short:             "Clone all active repos from an org or user",
	Long: `Clone all active (non-archived) repositories from a GitHub org or user into the target directory.
The optional directory argument specifies where to clone the repos (defaults to the current directory).

When --topics is provided multiple times, only repos that have ALL specified topics are cloned (AND logic).
To clone repos matching any one topic, run separate clone invocations per topic.`,
	Args:              cobra.MaximumNArgs(2),
	ValidArgsFunction: createCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              cloneFunc,
}

func cloneFunc(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	ctx := ctxhelper.WithVerbose(context.Background(), verbose)

	if !verbose {
		uiprogress.Start()
		defer uiprogress.Stop()
	}

	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	destDir := "."
	if len(args) > 1 {
		destDir = args[1]
	}

	repos, err := clt.GetRepos(ctx, name)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("get repos: %w", err)
	}

	repos = filterByTopics(repos, topics)

	ctx = ctxhelper.WithRepos(ctx, repos)

	_, err = clt.CloneRepos(ctx, destDir)
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
