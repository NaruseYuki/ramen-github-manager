package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	ghclient "github.com/NaruseYuki/ramen-github-manager/internal/github"
	"github.com/NaruseYuki/ramen-github-manager/internal/config"
	"github.com/NaruseYuki/ramen-github-manager/internal/display"
	"github.com/NaruseYuki/ramen-github-manager/internal/model"
)

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Short:   "Show project dashboard with summary statistics",
	Aliases: []string{"dash", "d"},
	RunE: func(cmd *cobra.Command, args []string) error {
		weekly, _ := cmd.Flags().GetBool("weekly")
		monthly, _ := cmd.Flags().GetBool("monthly")

		repos := config.RepoNames(cfg)

		if weekly || monthly {
			return runReport(repos, weekly)
		}

		return runDashboard(repos)
	},
}

func runDashboard(repos []string) error {
	stats := &model.DashboardStats{}

	// Fetch open issues and PRs for each repo
	openIssues, err := client.ListIssues(ghclient.ListIssuesOptions{
		Repos: repos, Owner: cfg.Owner, State: "open", Sort: "updated", Limit: 200,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	openPRs, err := client.ListPRs(ghclient.ListPRsOptions{
		Repos: repos, Owner: cfg.Owner, State: "open", Sort: "updated", Limit: 200,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch PRs: %w", err)
	}

	// Build per-repo stats
	staleThreshold := time.Now().AddDate(0, 0, -30)

	repoStatsMap := make(map[string]*model.RepoStat)
	for _, r := range repos {
		repoStatsMap[r] = &model.RepoStat{Repo: r}
	}

	for _, item := range openIssues {
		if rs, ok := repoStatsMap[item.Repo]; ok {
			rs.OpenIssues++
			if item.UpdatedAt.Before(staleThreshold) {
				rs.StaleIssueCount++
				stats.StaleIssues = append(stats.StaleIssues, item)
			}
		}
	}

	for _, item := range openPRs {
		if rs, ok := repoStatsMap[item.Repo]; ok {
			rs.OpenPRs++
			if item.ReviewStatus == "pending" || item.ReviewStatus == "" {
				stats.WaitingPRs = append(stats.WaitingPRs, item)
			}
		}
	}

	for _, r := range repos {
		stats.RepoStats = append(stats.RepoStats, *repoStatsMap[r])
	}

	// Sort stale issues by staleness (oldest first)
	sort.Slice(stats.StaleIssues, func(i, j int) bool {
		return stats.StaleIssues[i].UpdatedAt.Before(stats.StaleIssues[j].UpdatedAt)
	})

	// Limit stale issues display
	if len(stats.StaleIssues) > 10 {
		stats.StaleIssues = stats.StaleIssues[:10]
	}
	if len(stats.WaitingPRs) > 10 {
		stats.WaitingPRs = stats.WaitingPRs[:10]
	}

	display.Dashboard(stats, cfg)
	return nil
}

func runReport(repos []string, weekly bool) error {
	days := 7
	if !weekly {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	// Fetch recently closed issues
	closedIssues, err := client.ListIssues(ghclient.ListIssuesOptions{
		Repos: repos, Owner: cfg.Owner, State: "closed", Sort: "updated", Limit: 200,
	})
	if err != nil {
		return err
	}

	// Filter to period
	var recentClosed []model.IssueItem
	for _, item := range closedIssues {
		if item.UpdatedAt.After(since) {
			recentClosed = append(recentClosed, item)
		}
	}

	// Fetch recently created issues
	openIssues, err := client.ListIssues(ghclient.ListIssuesOptions{
		Repos: repos, Owner: cfg.Owner, State: "all", Sort: "created", Limit: 200,
	})
	if err != nil {
		return err
	}

	var recentCreated []model.IssueItem
	for _, item := range openIssues {
		if item.CreatedAt.After(since) {
			recentCreated = append(recentCreated, item)
		}
	}

	// Fetch recently merged PRs
	closedPRs, err := client.ListPRs(ghclient.ListPRsOptions{
		Repos: repos, Owner: cfg.Owner, State: "closed", Sort: "updated", Limit: 200,
	})
	if err != nil {
		return err
	}

	var recentMerged []model.PRItem
	for _, item := range closedPRs {
		if item.Merged && item.MergedAt != nil && item.MergedAt.After(since) {
			recentMerged = append(recentMerged, item)
		}
	}

	display.WeeklyReport(recentCreated, recentClosed, recentMerged, cfg)
	return nil
}

func init() {
	dashboardCmd.Flags().Bool("weekly", false, "Show weekly progress report")
	dashboardCmd.Flags().Bool("monthly", false, "Show monthly progress report")
}
