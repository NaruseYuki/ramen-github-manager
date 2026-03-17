package github

import (
	"fmt"
	"sort"
	"sync"

	gh "github.com/google/go-github/v69/github"

	"github.com/NaruseYuki/ramen-github-manager/internal/model"
)

// ListPRsOptions configures PR listing.
type ListPRsOptions struct {
	Repos        []string
	Owner        string
	State        string
	Author       string
	ReviewStatus string
	Sort         string
	Limit        int
}

// ListPRs fetches pull requests from multiple repositories concurrently.
func (c *Client) ListPRs(opts ListPRsOptions) ([]model.PRItem, error) {
	var (
		mu       sync.Mutex
		allItems []model.PRItem
		wg       sync.WaitGroup
		errCh    = make(chan error, len(opts.Repos))
	)

	ghState := opts.State
	if ghState == "merged" {
		ghState = "closed"
	}

	for _, repo := range opts.Repos {
		wg.Add(1)
		go func(repoName string) {
			defer wg.Done()

			ghOpts := &gh.PullRequestListOptions{
				State:       ghState,
				Sort:        opts.Sort,
				Direction:   "desc",
				ListOptions: gh.ListOptions{PerPage: 100},
			}

			var repoItems []model.PRItem
			for {
				prs, resp, err := c.inner.PullRequests.List(c.ctx, opts.Owner, repoName, ghOpts)
				if err != nil {
					errCh <- fmt.Errorf("[%s] %w", repoName, err)
					return
				}

				for _, pr := range prs {
					// Filter by author
					if opts.Author != "" && pr.GetUser().GetLogin() != opts.Author {
						continue
					}

					// Filter by merged status
					if opts.State == "merged" && !pr.GetMerged() {
						continue
					}

					item := model.PRItem{
						Repo:         repoName,
						Number:       pr.GetNumber(),
						Title:        pr.GetTitle(),
						State:        pr.GetState(),
						Author:       pr.GetUser().GetLogin(),
						Additions:    pr.GetAdditions(),
						Deletions:    pr.GetDeletions(),
						ChangedFiles: pr.GetChangedFiles(),
						Draft:        pr.GetDraft(),
						Merged:       pr.GetMerged(),
						CreatedAt:    pr.GetCreatedAt().Time,
						UpdatedAt:    pr.GetUpdatedAt().Time,
						Body:         pr.GetBody(),
						URL:          pr.GetHTMLURL(),
						BaseBranch:   pr.GetBase().GetRef(),
						HeadBranch:   pr.GetHead().GetRef(),
					}

					if pr.MergedAt != nil {
						t := pr.MergedAt.Time
						item.MergedAt = &t
					}

					for _, l := range pr.Labels {
						item.Labels = append(item.Labels, l.GetName())
					}
					if pr.Assignee != nil {
						item.Assignee = pr.Assignee.GetLogin()
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

	// Sort by updated time
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].UpdatedAt.After(allItems[j].UpdatedAt)
	})

	if opts.Limit > 0 && len(allItems) > opts.Limit {
		allItems = allItems[:opts.Limit]
	}

	return allItems, nil
}

// GetPR fetches a single pull request with reviews.
func (c *Client) GetPR(owner, repo string, number int) (*model.PRItem, []model.Comment, string, error) {
	pr, _, err := c.inner.PullRequests.Get(c.ctx, owner, repo, number)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to get PR #%d: %w", number, err)
	}

	item := &model.PRItem{
		Repo:         repo,
		Number:       pr.GetNumber(),
		Title:        pr.GetTitle(),
		State:        pr.GetState(),
		Author:       pr.GetUser().GetLogin(),
		Additions:    pr.GetAdditions(),
		Deletions:    pr.GetDeletions(),
		ChangedFiles: pr.GetChangedFiles(),
		Draft:        pr.GetDraft(),
		Merged:       pr.GetMerged(),
		CreatedAt:    pr.GetCreatedAt().Time,
		UpdatedAt:    pr.GetUpdatedAt().Time,
		Body:         pr.GetBody(),
		URL:          pr.GetHTMLURL(),
		BaseBranch:   pr.GetBase().GetRef(),
		HeadBranch:   pr.GetHead().GetRef(),
	}

	if pr.MergedAt != nil {
		t := pr.MergedAt.Time
		item.MergedAt = &t
	}

	for _, l := range pr.Labels {
		item.Labels = append(item.Labels, l.GetName())
	}
	if pr.Assignee != nil {
		item.Assignee = pr.Assignee.GetLogin()
	}

	// Fetch reviews to determine review status
	reviews, _, err := c.inner.PullRequests.ListReviews(c.ctx, owner, repo, number, nil)
	reviewStatus := "pending"
	if err == nil && len(reviews) > 0 {
		for i := len(reviews) - 1; i >= 0; i-- {
			state := reviews[i].GetState()
			if state == "APPROVED" || state == "CHANGES_REQUESTED" {
				reviewStatus = state
				break
			}
		}
	}
	item.ReviewStatus = reviewStatus

	// Fetch comments
	var comments []model.Comment
	ghComments, _, err := c.inner.Issues.ListComments(c.ctx, owner, repo, number, nil)
	if err == nil {
		for _, gc := range ghComments {
			comments = append(comments, model.Comment{
				Author:    gc.GetUser().GetLogin(),
				Body:      gc.GetBody(),
				CreatedAt: gc.GetCreatedAt().Time,
			})
		}
	}

	return item, comments, reviewStatus, nil
}

// MergePR merges a pull request.
func (c *Client) MergePR(owner, repo string, number int, method string) error {
	opts := &gh.PullRequestOptions{}
	switch method {
	case "squash":
		opts.MergeMethod = "squash"
	case "rebase":
		opts.MergeMethod = "rebase"
	default:
		opts.MergeMethod = "merge"
	}

	_, _, err := c.inner.PullRequests.Merge(c.ctx, owner, repo, number, "", opts)
	if err != nil {
		return fmt.Errorf("failed to merge PR #%d: %w", number, err)
	}
	return nil
}

// ApprovePR creates an approval review on a PR.
func (c *Client) ApprovePR(owner, repo string, number int, body string) error {
	event := "APPROVE"
	_, _, err := c.inner.PullRequests.CreateReview(c.ctx, owner, repo, number, &gh.PullRequestReviewRequest{
		Event: &event,
		Body:  &body,
	})
	if err != nil {
		return fmt.Errorf("failed to approve PR #%d: %w", number, err)
	}
	return nil
}
