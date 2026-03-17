package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/NaruseYuki/ramen-github-manager/internal/config"
	"github.com/NaruseYuki/ramen-github-manager/internal/model"
)

// Dashboard renders the project dashboard.
func Dashboard(stats *model.DashboardStats, cfg *model.Config) {
	now := time.Now()
	fmt.Printf("\n%s %s\n", bold("📊 Project Dashboard"), dim(now.Format("2006-01-02")))
	fmt.Printf("%s\n\n", dim(strings.Repeat("═", 65)))

	// Overview table
	fmt.Printf("%s\n\n", bold("📊 Overview"))

	const (
		wRepo  = 20
		wIssue = 12
		wPR    = 10
		wStale = 12
	)

	fmt.Printf("  %s %s %s %s %s %s %s\n",
		pad(dim("Repository"), wRepo), sep,
		pad(dim("Open Issues"), wIssue), sep,
		pad(dim("Open PRs"), wPR), sep,
		dim("Stale(>30d)"))
	fmt.Printf("  %s\n", dim(strings.Repeat("─", wRepo+wIssue+wPR+wStale+10)))

	totalIssues := 0
	totalPRs := 0
	totalStale := 0
	for _, rs := range stats.RepoStats {
		alias := config.AliasFor(cfg, rs.Repo)
		staleStr := fmt.Sprintf("%d", rs.StaleIssueCount)
		if rs.StaleIssueCount > 0 {
			staleStr = yellow(staleStr)
		}

		fmt.Printf("  %s %s %s %s %s %s %s\n",
			pad(alias, wRepo), sep,
			padNum(fmt.Sprintf("%d", rs.OpenIssues), wIssue), sep,
			padNum(fmt.Sprintf("%d", rs.OpenPRs), wPR), sep,
			padNum(staleStr, wStale))
		totalIssues += rs.OpenIssues
		totalPRs += rs.OpenPRs
		totalStale += rs.StaleIssueCount
	}
	fmt.Printf("  %s\n", dim(strings.Repeat("─", wRepo+wIssue+wPR+wStale+10)))

	totalStaleStr := fmt.Sprintf("%d", totalStale)
	if totalStale > 0 {
		totalStaleStr = yellow(totalStaleStr)
	}
	fmt.Printf("  %s %s %s %s %s %s %s\n\n",
		pad(bold("Total"), wRepo), sep,
		padNum(bold(fmt.Sprintf("%d", totalIssues)), wIssue), sep,
		padNum(bold(fmt.Sprintf("%d", totalPRs)), wPR), sep,
		padNum(totalStaleStr, wStale))

	// Recent activity
	if len(stats.RecentIssues) > 0 || len(stats.RecentPRs) > 0 {
		createdIssues := 0
		closedIssues := 0
		mergedPRs := 0
		for _, item := range stats.RecentIssues {
			if item.State == "open" {
				createdIssues++
			} else {
				closedIssues++
			}
		}
		for _, item := range stats.RecentPRs {
			if item.Merged {
				mergedPRs++
			}
		}

		fmt.Printf("%s\n", bold("🔄 Last 7 Days Activity"))
		fmt.Printf("  Issues Created: %s    Issues Closed: %s    PRs Merged: %s\n\n",
			bold(fmt.Sprintf("%d", createdIssues)),
			bold(fmt.Sprintf("%d", closedIssues)),
			bold(fmt.Sprintf("%d", mergedPRs)))
	}

	// Stale issues
	if len(stats.StaleIssues) > 0 {
		fmt.Printf("%s\n", bold("⚠️  Attention Needed"))
		for _, item := range stats.StaleIssues {
			alias := config.AliasFor(cfg, item.Repo)
			days := int(time.Since(item.UpdatedAt).Hours() / 24)
			fmt.Printf("  • [%s#%d] %s — %s\n",
				alias, item.Number, trunc(item.Title, 40),
				yellow(fmt.Sprintf("stale for %d days", days)))
		}
		fmt.Println()
	}

	// PRs awaiting review
	if len(stats.WaitingPRs) > 0 {
		fmt.Printf("%s\n", bold("👀 PRs Awaiting Review"))
		for _, item := range stats.WaitingPRs {
			alias := config.AliasFor(cfg, item.Repo)
			waiting := timeAgo(item.CreatedAt)
			fmt.Printf("  • [%s#%d] %s — waiting %s\n",
				alias, item.Number, trunc(item.Title, 40),
				cyan(waiting))
		}
		fmt.Println()
	}

	if len(stats.StaleIssues) == 0 && len(stats.WaitingPRs) == 0 {
		fmt.Printf("  %s\n\n", green("✓ All clear — no stale issues or waiting PRs!"))
	}
}

// WeeklyReport renders a weekly progress report.
func WeeklyReport(
	createdIssues, closedIssues []model.IssueItem,
	mergedPRs []model.PRItem,
	cfg *model.Config,
) {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)

	fmt.Printf("\n%s\n", bold("📈 Weekly Progress Report"))
	fmt.Printf("  %s — %s\n\n", weekAgo.Format("2006-01-02"), now.Format("2006-01-02"))

	fmt.Printf("  %s  Issues Created: %s\n", green("▲"), bold(fmt.Sprintf("%d", len(createdIssues))))
	fmt.Printf("  %s  Issues Closed:  %s\n", red("▼"), bold(fmt.Sprintf("%d", len(closedIssues))))
	fmt.Printf("  %s  PRs Merged:     %s\n\n", magenta("●"), bold(fmt.Sprintf("%d", len(mergedPRs))))

	if len(closedIssues) > 0 {
		fmt.Printf("  %s\n", bold("Closed Issues:"))
		for _, item := range closedIssues {
			alias := config.AliasFor(cfg, item.Repo)
			fmt.Printf("    ✓ [%s#%d] %s\n", alias, item.Number, trunc(item.Title, 50))
		}
		fmt.Println()
	}

	if len(mergedPRs) > 0 {
		fmt.Printf("  %s\n", bold("Merged PRs:"))
		for _, item := range mergedPRs {
			alias := config.AliasFor(cfg, item.Repo)
			fmt.Printf("    ● [%s#%d] %s\n", alias, item.Number, trunc(item.Title, 50))
		}
		fmt.Println()
	}
}
