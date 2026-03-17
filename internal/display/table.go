package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/NaruseYuki/ramen-github-manager/internal/config"
	"github.com/NaruseYuki/ramen-github-manager/internal/model"
)

var (
	bold      = color.New(color.Bold).SprintFunc()
	green     = color.New(color.FgGreen).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
	yellow    = color.New(color.FgYellow).SprintFunc()
	cyan      = color.New(color.FgCyan).SprintFunc()
	magenta   = color.New(color.FgMagenta).SprintFunc()
	dim       = color.New(color.Faint).SprintFunc()
	boldGreen = color.New(color.Bold, color.FgGreen).SprintFunc()
	boldRed   = color.New(color.Bold, color.FgRed).SprintFunc()
)

func stateIcon(state string) string {
	switch state {
	case "open":
		return green("●")
	case "closed":
		return red("●")
	case "merged":
		return magenta("●")
	default:
		return "○"
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func formatLabels(labels []string) string {
	if len(labels) == 0 {
		return dim("—")
	}
	colored := make([]string, len(labels))
	for i, l := range labels {
		colored[i] = yellow(l)
	}
	return strings.Join(colored, ", ")
}

// IssueTable prints issues in a table format.
func IssueTable(items []model.IssueItem, cfg *model.Config) {
	if len(items) == 0 {
		fmt.Println(dim("  No issues found."))
		return
	}

	fmt.Printf("\n%s\n\n", bold(fmt.Sprintf("📋 Issues (%d total)", len(items))))

	// Header
	fmt.Printf("  %-10s %-5s %-40s %-20s %-12s %s\n",
		dim("REPO"), dim("#"), dim("TITLE"), dim("LABELS"), dim("ASSIGNEE"), dim("UPDATED"))
	fmt.Printf("  %s\n", dim(strings.Repeat("─", 100)))

	for _, item := range items {
		alias := config.AliasFor(cfg, item.Repo)
		assignee := item.Assignee
		if assignee == "" {
			assignee = dim("—")
		}

		fmt.Printf("  %-10s %s%-4d %-40s %-20s %-12s %s\n",
			cyan(alias),
			stateIcon(item.State),
			item.Number,
			truncate(item.Title, 38),
			truncate(formatLabels(item.Labels), 18),
			truncate(assignee, 10),
			dim(timeAgo(item.UpdatedAt)),
		)
	}
	fmt.Println()
}

// IssueDetail prints a single issue in detail.
func IssueDetail(item *model.IssueItem, comments []model.Comment, cfg *model.Config) {
	alias := config.AliasFor(cfg, item.Repo)

	fmt.Printf("\n%s %s#%d\n", stateIcon(item.State), cyan(alias), item.Number)
	fmt.Printf("%s\n\n", bold(item.Title))

	fmt.Printf("  State:    %s\n", item.State)
	fmt.Printf("  Author:   %s\n", item.Author)
	if item.Assignee != "" {
		fmt.Printf("  Assignee: %s\n", item.Assignee)
	}
	if len(item.Labels) > 0 {
		fmt.Printf("  Labels:   %s\n", formatLabels(item.Labels))
	}
	fmt.Printf("  Created:  %s\n", item.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("  Updated:  %s\n", item.UpdatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("  URL:      %s\n", dim(item.URL))

	if item.Body != "" {
		fmt.Printf("\n%s\n%s\n", dim("───────────────────────────────────────"), item.Body)
	}

	if len(comments) > 0 {
		fmt.Printf("\n%s\n", bold(fmt.Sprintf("💬 Comments (%d)", len(comments))))
		for _, c := range comments {
			fmt.Printf("\n  %s %s\n", bold(c.Author), dim(timeAgo(c.CreatedAt)))
			for _, line := range strings.Split(c.Body, "\n") {
				fmt.Printf("  %s\n", line)
			}
		}
	}
	fmt.Println()
}

// PRTable prints pull requests in a table format.
func PRTable(items []model.PRItem, cfg *model.Config) {
	if len(items) == 0 {
		fmt.Println(dim("  No pull requests found."))
		return
	}

	fmt.Printf("\n%s\n\n", bold(fmt.Sprintf("🔀 Pull Requests (%d total)", len(items))))

	fmt.Printf("  %-10s %-5s %-38s %-12s %-10s %-10s %s\n",
		dim("REPO"), dim("#"), dim("TITLE"), dim("AUTHOR"), dim("STATUS"), dim("+/-"), dim("UPDATED"))
	fmt.Printf("  %s\n", dim(strings.Repeat("─", 100)))

	for _, item := range items {
		alias := config.AliasFor(cfg, item.Repo)

		state := item.State
		icon := stateIcon(state)
		if item.Merged {
			state = "merged"
			icon = magenta("●")
		}
		if item.Draft {
			icon = dim("◌")
		}

		review := ""
		switch item.ReviewStatus {
		case "APPROVED":
			review = green("✓")
		case "CHANGES_REQUESTED":
			review = red("✗")
		default:
			review = yellow("○")
		}

		diffStr := fmt.Sprintf("%s/%s", boldGreen(fmt.Sprintf("+%d", item.Additions)), boldRed(fmt.Sprintf("-%d", item.Deletions)))

		fmt.Printf("  %-10s %s%-4d %-38s %-12s %s %-8s %s\n",
			cyan(alias),
			icon,
			item.Number,
			truncate(item.Title, 36),
			truncate(item.Author, 10),
			review,
			diffStr,
			dim(timeAgo(item.UpdatedAt)),
		)
	}
	fmt.Println()
}

// PRDetail prints a single PR in detail.
func PRDetail(item *model.PRItem, comments []model.Comment, reviewStatus string, cfg *model.Config) {
	alias := config.AliasFor(cfg, item.Repo)

	state := item.State
	if item.Merged {
		state = "merged"
	}

	fmt.Printf("\n%s %s#%d\n", stateIcon(state), cyan(alias), item.Number)
	fmt.Printf("%s\n\n", bold(item.Title))

	fmt.Printf("  State:    %s\n", state)
	fmt.Printf("  Author:   %s\n", item.Author)
	fmt.Printf("  Branch:   %s → %s\n", cyan(item.HeadBranch), item.BaseBranch)
	fmt.Printf("  Review:   %s\n", reviewStatus)
	fmt.Printf("  Changes:  %s, %s (%d files)\n",
		boldGreen(fmt.Sprintf("+%d", item.Additions)),
		boldRed(fmt.Sprintf("-%d", item.Deletions)),
		item.ChangedFiles)
	if item.Draft {
		fmt.Printf("  Draft:    %s\n", yellow("yes"))
	}
	if len(item.Labels) > 0 {
		fmt.Printf("  Labels:   %s\n", formatLabels(item.Labels))
	}
	fmt.Printf("  Created:  %s\n", item.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("  Updated:  %s\n", item.UpdatedAt.Format("2006-01-02 15:04"))
	if item.MergedAt != nil {
		fmt.Printf("  Merged:   %s\n", item.MergedAt.Format("2006-01-02 15:04"))
	}
	fmt.Printf("  URL:      %s\n", dim(item.URL))

	if item.Body != "" {
		fmt.Printf("\n%s\n%s\n", dim("───────────────────────────────────────"), item.Body)
	}

	if len(comments) > 0 {
		fmt.Printf("\n%s\n", bold(fmt.Sprintf("💬 Comments (%d)", len(comments))))
		for _, c := range comments {
			fmt.Printf("\n  %s %s\n", bold(c.Author), dim(timeAgo(c.CreatedAt)))
			for _, line := range strings.Split(c.Body, "\n") {
				fmt.Printf("  %s\n", line)
			}
		}
	}
	fmt.Println()
}
