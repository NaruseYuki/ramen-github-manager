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

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage issues across repositories",
	Aliases: []string{"i", "issues"},
}

var issueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues across all repositories",
	Aliases: []string{"ls", "l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		repoFlag, _ := cmd.Flags().GetString("repo")
		state, _ := cmd.Flags().GetString("state")
		labelFlag, _ := cmd.Flags().GetStringSlice("label")
		assignee, _ := cmd.Flags().GetString("assignee")
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

		items, err := client.ListIssues(ghclient.ListIssuesOptions{
			Repos:    repos,
			Owner:    cfg.Owner,
			State:    state,
			Labels:   labelFlag,
			Assignee: assignee,
			Sort:     sortBy,
			Limit:    limit,
		})
		if err != nil {
			return err
		}

		display.IssueTable(items, cfg)
		return nil
	},
}

var issueViewCmd = &cobra.Command{
	Use:   "view <repo> <number>",
	Short: "View issue details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[1])
		}

		item, comments, err := client.GetIssue(cfg.Owner, repo, number)
		if err != nil {
			return err
		}

		display.IssueDetail(item, comments, cfg)
		return nil
	},
}

var issueCreateCmd = &cobra.Command{
	Use:   "create <repo>",
	Short: "Create a new issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		labels, _ := cmd.Flags().GetStringSlice("label")
		assignees, _ := cmd.Flags().GetStringSlice("assignee")

		reader := bufio.NewReader(os.Stdin)

		if title == "" {
			fmt.Print("Title: ")
			title, _ = reader.ReadString('\n')
			title = strings.TrimSpace(title)
		}
		if body == "" {
			fmt.Print("Body (enter empty line to finish):\n")
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

		item, err := client.CreateIssue(cfg.Owner, repo, title, body, labels, assignees)
		if err != nil {
			return err
		}

		fmt.Printf("✅ Created issue #%d: %s\n   %s\n", item.Number, item.Title, item.URL)
		return nil
	},
}

var issueCloseCmd = &cobra.Command{
	Use:   "close <repo> <number>",
	Short: "Close an issue",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[1])
		}

		if err := client.CloseIssue(cfg.Owner, repo, number); err != nil {
			return err
		}

		fmt.Printf("✅ Closed issue %s#%d\n", repo, number)
		return nil
	},
}

var issueReopenCmd = &cobra.Command{
	Use:   "reopen <repo> <number>",
	Short: "Reopen an issue",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[1])
		}

		if err := client.ReopenIssue(cfg.Owner, repo, number); err != nil {
			return err
		}

		fmt.Printf("✅ Reopened issue %s#%d\n", repo, number)
		return nil
	},
}

var issueLabelCmd = &cobra.Command{
	Use:   "label <repo> <number>",
	Short: "Manage labels on an issue",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[1])
		}

		addLabels, _ := cmd.Flags().GetStringSlice("add")
		removeLabels, _ := cmd.Flags().GetStringSlice("remove")

		if len(addLabels) > 0 {
			if err := client.AddLabels(cfg.Owner, repo, number, addLabels); err != nil {
				return err
			}
			fmt.Printf("✅ Added labels: %s\n", strings.Join(addLabels, ", "))
		}

		for _, label := range removeLabels {
			if err := client.RemoveLabel(cfg.Owner, repo, number, label); err != nil {
				return err
			}
			fmt.Printf("✅ Removed label: %s\n", label)
		}

		return nil
	},
}

var issueAssignCmd = &cobra.Command{
	Use:   "assign <repo> <number> <user>",
	Short: "Assign users to an issue",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[1])
		}

		assignees := args[2:]
		if err := client.AssignIssue(cfg.Owner, repo, number, assignees); err != nil {
			return err
		}

		fmt.Printf("✅ Assigned %s to %s#%d\n", strings.Join(assignees, ", "), repo, number)
		return nil
	},
}

var issueCommentCmd = &cobra.Command{
	Use:   "comment <repo> <number>",
	Short: "Add a comment to an issue",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := config.ResolveRepo(cfg, args[0])
		if err != nil {
			return err
		}
		number, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[1])
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

		fmt.Printf("✅ Added comment to %s#%d\n", repo, number)
		return nil
	},
}

func init() {
	// issue list flags
	issueListCmd.Flags().StringP("repo", "r", "", "Filter by repository (name or alias)")
	issueListCmd.Flags().StringP("state", "s", "", "Filter by state: open, closed, all")
	issueListCmd.Flags().StringSliceP("label", "l", nil, "Filter by labels")
	issueListCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	issueListCmd.Flags().String("sort", "", "Sort by: created, updated, comments")
	issueListCmd.Flags().IntP("limit", "n", 0, "Maximum number of results")

	// issue create flags
	issueCreateCmd.Flags().StringP("title", "t", "", "Issue title")
	issueCreateCmd.Flags().StringP("body", "b", "", "Issue body")
	issueCreateCmd.Flags().StringSliceP("label", "l", nil, "Labels to add")
	issueCreateCmd.Flags().StringSliceP("assignee", "a", nil, "Assignees")

	// issue label flags
	issueLabelCmd.Flags().StringSlice("add", nil, "Labels to add")
	issueLabelCmd.Flags().StringSlice("remove", nil, "Labels to remove")

	// issue comment flags
	issueCommentCmd.Flags().StringP("body", "b", "", "Comment body")

	issueCmd.AddCommand(issueListCmd)
	issueCmd.AddCommand(issueViewCmd)
	issueCmd.AddCommand(issueCreateCmd)
	issueCmd.AddCommand(issueCloseCmd)
	issueCmd.AddCommand(issueReopenCmd)
	issueCmd.AddCommand(issueLabelCmd)
	issueCmd.AddCommand(issueAssignCmd)
	issueCmd.AddCommand(issueCommentCmd)
}
