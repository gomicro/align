package cmd

import (
	"context"
	"strings"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

func init() {
}

var downloadCmd = &cobra.Command{
	Use:               "download",
	Short:             "Download all active repos from an org or user.",
	Long:              `Download all active repos from an org or user.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: createCmdValidArgsFunc,
	PersistentPreRun:  setupClient,
	RunE:              downloadFunc,
}

func downloadFunc(cmd *cobra.Command, args []string) error {
	name := args[0]
	ctx := context.Background()

	repos, err := clt.GetRepos(ctx, name)
	if err != nil {
		return err
	}

	dirRepo := parseDirRepoMap(repos)

	r := dirRepo["dan9186"][0]

	err = clt.CloneRepo(ctx, r.name, "dan9186", r.url)
	if err != nil {
		return err
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

type repository struct {
	name string
	url  string
}

func parseDirRepoMap(repos []*github.Repository) map[string][]*repository {
	var dirRepo = map[string][]*repository{}
	for _, repo := range repos {
		parts := strings.Split(*repo.SSHURL, "/")

		dir := strings.Split(parts[0], ":")[1]
		name := strings.TrimSuffix(parts[1], ".git")

		r := &repository{
			name: name,
			url:  *repo.SSHURL,
		}

		dirRepo[dir] = append(dirRepo[dir], r)
	}

	return dirRepo
}
