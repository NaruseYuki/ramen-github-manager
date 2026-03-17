package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NaruseYuki/ramen-github-manager/internal/config"
	ghclient "github.com/NaruseYuki/ramen-github-manager/internal/github"
	"github.com/NaruseYuki/ramen-github-manager/internal/model"
)

var (
	cfg      *model.Config
	client   *ghclient.Client
	noColor  bool
	jsonOut  bool
)

var rootCmd = &cobra.Command{
	Use:   "rgm",
	Short: "rgm — manage GitHub issues and PRs across multiple repositories",
	Long: `rgm is a CLI tool for centrally managing GitHub Issues and Pull Requests
across multiple repositories in a project.

Run 'rgm config init' to set up your project, then use 'rgm issue list'
or 'rgm dashboard' to get started.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config/client init for all config subcommands (they don't need API)
		for c := cmd; c != nil; c = c.Parent() {
			if c.Name() == "config" {
				return nil
			}
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}

		client, err = ghclient.NewClient()
		if err != nil {
			return err
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output in JSON format")

	rootCmd.AddCommand(issueCmd)
	rootCmd.AddCommand(prCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(configCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func exitWithError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

// resolveRepos resolves the --repo flag or returns all configured repos.
func resolveRepos(repoFlag string) ([]string, error) {
	if repoFlag != "" {
		resolved, err := config.ResolveRepo(cfg, repoFlag)
		if err != nil {
			return nil, err
		}
		return []string{resolved}, nil
	}
	return config.RepoNames(cfg), nil
}
