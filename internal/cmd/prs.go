package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NaruseYuki/ramen-github-manager/internal/config"
	"github.com/NaruseYuki/ramen-github-manager/internal/display"
	ghclient "github.com/NaruseYuki/ramen-github-manager/internal/github"
)

var prCmd = &cobra.Command{
	Use:     "pr",
	Short:   "Manage pull requests across repositories",
	Aliases: []string{"prs", "pull-request"},
}

var prListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List pull requests across all repositories",
	Aliases: []string{"ls", "l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		repoFlag, _ := cmd.Flags().GetString("repo")
		state, _ := cmd.Flags().GetString("state")
		author, _ := cmd.Flags().GetString("author")
		sortBy, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")

		repos, err := resolveRepos(repoFlag)
		if err != nil {
			return err
		}

		if state == "" {
			state = cfg.Defaults.State
		}
		if sortBy == "" {
			sortBy = cfg.Defaults.Sort
		}
		if limit == 0 {
			limit = cfg.Defaults.Limit
		}

		items, err := client.ListPRs(ghclient.ListPRsOptions{
			Repos:  repos,
			Owner:  cfg.Owner,
			State:  state,
			Author: author,
			Sort:   sortBy,
			Limit:  limit,
		})
		if err != nil {
			return err
		}

		display.PRTable(items, cfg)
		return nil
	},
}

var prViewCmd = &cobra.Command{
	Use:   "view <repo> <number>",
	Short: "View pull request details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid PR number: %s", args[1])
		}

		item, comments, reviewStatus, err := client.GetPR(cfg.Owner, repo, number)
		if err != nil {
			return err
		}

		display.PRDetail(item, comments, reviewStatus, cfg)
		return nil
	},
}

var prApproveCmd = &cobra.Command{
	Use:   "approve <repo> <number>",
	Short: "Approve a pull request",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid PR number: %s", args[1])
		}

		body, _ := cmd.Flags().GetString("body")
		if err := client.ApprovePR(cfg.Owner, repo, number, body); err != nil {
			return err
		}

		fmt.Printf("✅ Approved PR %s#%d\n", repo, number)
		return nil
	},
}

var prMergeCmd = &cobra.Command{
	Use:   "merge <repo> <number>",
	Short: "Merge a pull request",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid PR number: %s", args[1])
		}

		method, _ := cmd.Flags().GetString("method")
		if err := client.MergePR(cfg.Owner, repo, number, method); err != nil {
			return err
		}

		fmt.Printf("✅ Merged PR %s#%d (method: %s)\n", repo, number, method)
		return nil
	},
}

var prCommentCmd = &cobra.Command{
	Use:   "comment <repo> <number>",
	Short: "Add a comment to a pull request",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid PR number: %s", args[1])
		}

		body, _ := cmd.Flags().GetString("body")
		if body == "" {
			fmt.Print("Comment:\n")
			reader := bufio.NewReader(os.Stdin)
			var lines []string
			for {
				line, _ := reader.ReadString('\n')
				line = strings.TrimRight(line, "\n")
				if line == "" {
					break
				}
				lines = append(lines, line)
			}
			body = strings.Join(lines, "\n")
		}

		if err := client.AddComment(cfg.Owner, repo, number, body); err != nil {
			return err
		}

		fmt.Printf("✅ Added comment to PR %s#%d\n", repo, number)
		return nil
	},
}

func init() {
	// pr list flags
	prListCmd.Flags().StringP("repo", "r", "", "Filter by repository (name or alias)")
	prListCmd.Flags().StringP("state", "s", "", "Filter by state: open, closed, merged, all")
	prListCmd.Flags().StringP("author", "a", "", "Filter by author")
	prListCmd.Flags().String("sort", "", "Sort by: created, updated")
	prListCmd.Flags().IntP("limit", "n", 0, "Maximum number of results")

	// pr approve flags
	prApproveCmd.Flags().StringP("body", "b", "", "Approval message")

	// pr merge flags
	prMergeCmd.Flags().StringP("method", "m", "merge", "Merge method: merge, squash, rebase")

	// pr comment flags
	prCommentCmd.Flags().StringP("body", "b", "", "Comment body")

	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prViewCmd)
	prCmd.AddCommand(prApproveCmd)
	prCmd.AddCommand(prMergeCmd)
	prCmd.AddCommand(prCommentCmd)
}
