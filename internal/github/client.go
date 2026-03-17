package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gh "github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Client wraps the GitHub API client.
type Client struct {
	inner *gh.Client
	ctx   context.Context
}

type ghHosts struct {
	Hosts map[string]ghHostEntry `yaml:",inline"`
}

type ghHostEntry struct {
	OAuthToken string `yaml:"oauth_token"`
	User       string `yaml:"user"`
}

// NewClient creates a new GitHub API client using the gh CLI token.
func NewClient() (*Client, error) {
	token, err := loadGHToken()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := gh.NewClient(tc)

	return &Client{inner: client, ctx: ctx}, nil
}

// loadGHToken reads the GitHub token from gh CLI or environment.
func loadGHToken() (string, error) {
	// 1. Check environment variable first
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token, nil
	}

	// 2. Try `gh auth token` command (works with keychain-based storage)
	if token, err := execGHAuthToken(); err == nil && token != "" {
		return token, nil
	}

	// 3. Read from gh CLI config file (legacy oauth_token field)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	hostsPath := filepath.Join(home, ".config", "gh", "hosts.yml")
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return "", fmt.Errorf("gh CLI not configured. Run 'gh auth login' first: %w", err)
	}

	var raw map[string]map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("failed to parse gh hosts.yml: %w", err)
	}

	ghHost, ok := raw["github.com"]
	if !ok {
		return "", fmt.Errorf("github.com not found in gh hosts.yml")
	}

	if token, ok := ghHost["oauth_token"].(string); ok && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no GitHub token found. Run 'gh auth login' or set GITHUB_TOKEN")
}

// execGHAuthToken runs `gh auth token` to get the token from gh CLI.
func execGHAuthToken() (string, error) {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RateLimit returns the current rate limit status.
func (c *Client) RateLimit() (*gh.RateLimits, error) {
	limits, _, err := c.inner.RateLimit.Get(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}
	return limits, nil
}

// checkRateLimit verifies we have API calls remaining.
func (c *Client) checkRateLimit(resp *gh.Response) error {
	if resp != nil && resp.Rate.Remaining < 10 {
		return fmt.Errorf("GitHub API rate limit nearly exhausted (%d remaining, resets at %s)",
			resp.Rate.Remaining, resp.Rate.Reset.Time.Format("15:04:05"))
	}
	return nil
}

// isNotFound checks if an error is a 404 response.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if ghErr, ok := err.(*gh.ErrorResponse); ok {
		return ghErr.Response.StatusCode == http.StatusNotFound
	}
	return false
}
