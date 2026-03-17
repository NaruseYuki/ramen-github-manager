package github

import (
	"fmt"
	"sort"
	"sync"

	gh "github.com/google/go-github/v69/github"

	"github.com/NaruseYuki/ramen-github-manager/internal/model"
)

// ListIssuesOptions configures issue listing.
type ListIssuesOptions struct {
	Repos    []string
	Owner    string
	State    string
	Labels   []string
	Assignee string
	Sort     string
	Limit    int
}

// ListIssues fetches issues from multiple repositories concurrently.
func (c *Client) ListIssues(opts ListIssuesOptions) ([]model.IssueItem, error) {
	var (
		mu       sync.Mutex
		allItems []model.IssueItem
		wg       sync.WaitGroup
		errCh    = make(chan error, len(opts.Repos))
	)

	for _, repo := range opts.Repos {
		wg.Add(1)
		go func(repoName string) {
			defer wg.Done()

			ghOpts := &gh.IssueListByRepoOptions{
				State:       opts.State,
				Sort:        opts.Sort,
				Direction:   "desc",
				ListOptions: gh.ListOptions{PerPage: 100},
			}
			if opts.Assignee != "" {
				ghOpts.Assignee = opts.Assignee
			}
			if len(opts.Labels) > 0 {
				ghOpts.Labels = opts.Labels
			}

			var repoItems []model.IssueItem
			for {
				issues, resp, err := c.inner.Issues.ListByRepo(c.ctx, opts.Owner, repoName, ghOpts)
				if err != nil {
					errCh <- fmt.Errorf("[%s] %w", repoName, err)
					return
				}

				for _, issue := range issues {
					if issue.PullRequestLinks != nil {
						continue // skip PRs returned by issues API
					}

					item := model.IssueItem{
						Repo:      repoName,
						Number:    issue.GetNumber(),
						Title:     issue.GetTitle(),
						State:     issue.GetState(),
						Author:    issue.GetUser().GetLogin(),
						Comments:  issue.GetComments(),
						CreatedAt: issue.GetCreatedAt().Time,
						UpdatedAt: issue.GetUpdatedAt().Time,
						Body:      issue.GetBody(),
						URL:       issue.GetHTMLURL(),
					}

					for _, l := range issue.Labels {
						item.Labels = append(item.Labels, l.GetName())
					}
					if issue.Assignee != nil {
						item.Assignee = issue.Assignee.GetLogin()
					}

					repoItems = append(repoItems, item)
				}

				if resp.NextPage == 0 {
					break
				}
				ghOpts.Page = resp.NextPage

				if err := c.checkRateLimit(resp); err != nil {
					errCh <- err
					return
				}
			}

			mu.Lock()
			allItems = append(allItems, repoItems...)
			mu.Unlock()
		}(repo)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	// Sort by updated time (most recent first)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].UpdatedAt.After(allItems[j].UpdatedAt)
	})

	if opts.Limit > 0 && len(allItems) > opts.Limit {
		allItems = allItems[:opts.Limit]
	}

	return allItems, nil
}

// GetIssue fetches a single issue with its comments.
func (c *Client) GetIssue(owner, repo string, number int) (*model.IssueItem, []model.Comment, error) {
	issue, _, err := c.inner.Issues.Get(c.ctx, owner, repo, number)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get issue #%d: %w", number, err)
	}

	item := &model.IssueItem{
		Repo:      repo,
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		State:     issue.GetState(),
		Author:    issue.GetUser().GetLogin(),
		Comments:  issue.GetComments(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
		Body:      issue.GetBody(),
		URL:       issue.GetHTMLURL(),
	}
	for _, l := range issue.Labels {
		item.Labels = append(item.Labels, l.GetName())
	}
	if issue.Assignee != nil {
		item.Assignee = issue.Assignee.GetLogin()
	}

	// Fetch comments
	var comments []model.Comment
	opts := &gh.IssueListCommentsOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}
	for {
		ghComments, resp, err := c.inner.Issues.ListComments(c.ctx, owner, repo, number, opts)
		if err != nil {
			return item, nil, fmt.Errorf("failed to get comments: %w", err)
		}
		for _, gc := range ghComments {
			comments = append(comments, model.Comment{
				Author:    gc.GetUser().GetLogin(),
				Body:      gc.GetBody(),
				CreatedAt: gc.GetCreatedAt().Time,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return item, comments, nil
}

// CreateIssue creates a new issue.
func (c *Client) CreateIssue(owner, repo, title, body string, labels []string, assignees []string) (*model.IssueItem, error) {
	req := &gh.IssueRequest{
		Title: &title,
		Body:  &body,
	}
	if len(labels) > 0 {
		req.Labels = &labels
	}
	if len(assignees) > 0 {
		req.Assignees = &assignees
	}

	issue, _, err := c.inner.Issues.Create(c.ctx, owner, repo, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	item := &model.IssueItem{
		Repo:      repo,
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		State:     issue.GetState(),
		URL:       issue.GetHTMLURL(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
	}
	return item, nil
}

// CloseIssue closes an issue.
func (c *Client) CloseIssue(owner, repo string, number int) error {
	state := "closed"
	_, _, err := c.inner.Issues.Edit(c.ctx, owner, repo, number, &gh.IssueRequest{State: &state})
	if err != nil {
		return fmt.Errorf("failed to close issue #%d: %w", number, err)
	}
	return nil
}

// ReopenIssue reopens an issue.
func (c *Client) ReopenIssue(owner, repo string, number int) error {
	state := "open"
	_, _, err := c.inner.Issues.Edit(c.ctx, owner, repo, number, &gh.IssueRequest{State: &state})
	if err != nil {
		return fmt.Errorf("failed to reopen issue #%d: %w", number, err)
	}
	return nil
}

// AddLabels adds labels to an issue.
func (c *Client) AddLabels(owner, repo string, number int, labels []string) error {
	_, _, err := c.inner.Issues.AddLabelsToIssue(c.ctx, owner, repo, number, labels)
	if err != nil {
		return fmt.Errorf("failed to add labels: %w", err)
	}
	return nil
}

// RemoveLabel removes a label from an issue.
func (c *Client) RemoveLabel(owner, repo string, number int, label string) error {
	_, err := c.inner.Issues.RemoveLabelForIssue(c.ctx, owner, repo, number, label)
	if err != nil {
		return fmt.Errorf("failed to remove label %q: %w", label, err)
	}
	return nil
}

// AssignIssue assigns users to an issue.
func (c *Client) AssignIssue(owner, repo string, number int, assignees []string) error {
	_, _, err := c.inner.Issues.AddAssignees(c.ctx, owner, repo, number, assignees)
	if err != nil {
		return fmt.Errorf("failed to assign issue: %w", err)
	}
	return nil
}

// AddComment adds a comment to an issue or PR.
func (c *Client) AddComment(owner, repo string, number int, body string) error {
	_, _, err := c.inner.Issues.CreateComment(c.ctx, owner, repo, number, &gh.IssueComment{Body: &body})
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}
	return nil
}
