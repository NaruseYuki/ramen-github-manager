package model

import "time"

// RepoConfig represents a repository configuration entry.
type RepoConfig struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
}

// Config represents the CLI configuration.
type Config struct {
	Owner        string       `yaml:"owner"`
	Repositories []RepoConfig `yaml:"repositories"`
	Defaults     Defaults     `yaml:"defaults"`
}

// Defaults holds default values for commands.
type Defaults struct {
	Sort  string `yaml:"sort"`
	Limit int    `yaml:"limit"`
	State string `yaml:"state"`
}

// IssueItem represents an issue for cross-repo display.
type IssueItem struct {
	Repo      string
	Number    int
	Title     string
	State     string
	Labels    []string
	Assignee  string
	Author    string
	Comments  int
	CreatedAt time.Time
	UpdatedAt time.Time
	Body      string
	URL       string
}

// PRItem represents a pull request for cross-repo display.
type PRItem struct {
	Repo         string
	Number       int
	Title        string
	State        string
	Labels       []string
	Author       string
	Assignee     string
	ReviewStatus string
	Additions    int
	Deletions    int
	ChangedFiles int
	Draft        bool
	Merged       bool
	MergedAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Body         string
	URL          string
	BaseBranch   string
	HeadBranch   string
}

// DashboardStats holds aggregated statistics for the dashboard.
type DashboardStats struct {
	RepoStats    []RepoStat
	RecentIssues []IssueItem
	RecentPRs    []PRItem
	StaleIssues  []IssueItem
	WaitingPRs   []PRItem
}

// RepoStat holds per-repo statistics.
type RepoStat struct {
	Repo            string
	OpenIssues      int
	ClosedIssues    int
	OpenPRs         int
	MergedPRs       int
	StaleIssueCount int
}

// Comment represents an issue or PR comment.
type Comment struct {
	Author    string
	Body      string
	CreatedAt time.Time
}
