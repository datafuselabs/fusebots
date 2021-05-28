// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.
// Some codes from https://github.com/p1ass/mikku

package common

import (
	"context"
	"fmt"
	"time"

	"bots/config"
	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

type Client struct {
	cfg    *config.Config
	ctx    *context.Context
	client *github.Client
}

func NewClient(cfg *config.Config) *Client {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: cfg.GithubToken,
	})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		cfg:    cfg,
		ctx:    &ctx,
		client: client,
	}
}

func (s *Client) CreateComment(number int, comment *string) error {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	issueComment := &github.IssueComment{
		Body: comment,
	}
	_, _, err := s.client.Issues.CreateComment(ctx, s.cfg.RepoOwner, s.cfg.RepoName, number, issueComment)
	return err
}

func (s *Client) PullRequestMerge(number int, comment string) error {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	opts := github.PullRequestOptions{
		MergeMethod: "merge",
	}
	_, _, err := s.client.PullRequests.Merge(ctx, s.cfg.RepoOwner, s.cfg.RepoName, number, comment, &opts)
	return err
}

func (s *Client) PullRequestList() ([]*github.PullRequest, error) {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	var results []*github.PullRequest
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		prs, resp, err := s.client.PullRequests.List(ctx, s.cfg.RepoOwner, s.cfg.RepoName, opts)
		if err != nil {
			return results, err
		}
		for _, pr := range prs {
			results = append(results, pr)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return results, nil
}

func (s *Client) ListCheckRunsForRef(ref string) (*github.ListCheckRunsResults, error) {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	opts := &github.ListCheckRunsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	checkRuns, _, err := s.client.Checks.ListCheckRunsForRef(ctx, s.cfg.RepoOwner, s.cfg.RepoName, ref, opts)

	return checkRuns, err
}

func (s *Client) PullRequestListReviewers(number int) (*github.Reviewers, error) {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	opts := &github.ListOptions{PerPage: 100}
	reviewers, _, err := s.client.PullRequests.ListReviewers(ctx, s.cfg.RepoOwner, s.cfg.RepoName, number, opts)
	return reviewers, err
}

func (s *Client) PullRequestListReviews(number int) ([]*github.PullRequestReview, error) {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	opts := &github.ListOptions{PerPage: 100}
	reviews, _, err := s.client.PullRequests.ListReviews(ctx, s.cfg.RepoOwner, s.cfg.RepoName, number, opts)
	return reviews, err

}

func (s *Client) GetMergedPullRequestsAfter(branch string, after time.Time) ([]*github.PullRequest, error) {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	opts := &github.PullRequestListOptions{
		State:       "closed",
		Base:        branch,
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var prList []*github.PullRequest
	for {
		prs, resp, err := s.client.PullRequests.List(ctx, s.cfg.RepoOwner, s.cfg.RepoName, opts)
		if err != nil {
			return nil, fmt.Errorf("call listing pull requests API: %w", err)
		}

		extractedPR, done := extractMergedPRsAfter(prs, after)
		prList = append(prList, extractedPR...)
		if done {
			break
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return prList, nil
}

func extractMergedPRsAfter(prs []*github.PullRequest, after time.Time) ([]*github.PullRequest, bool) {
	var prList []*github.PullRequest
	done := false
	for _, pr := range prs {
		if pr.MergedAt != nil && pr.MergedAt.After(after) {
			prList = append(prList, pr)
		}
		if pr.UpdatedAt != nil && !pr.UpdatedAt.After(after) {
			done = true
			break
		}
	}
	return prList, done
}

func (s *Client) CreateRelease(tagName, body string, preRelease bool) (*github.RepositoryRelease, error) {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	release, _, err := s.client.Repositories.CreateRelease(ctx, s.cfg.RepoOwner, s.cfg.RepoName, &github.RepositoryRelease{
		TagName:    github.String(tagName),
		Name:       github.String(tagName),
		Body:       github.String(body),
		Prerelease: &preRelease,
	})
	if err != nil {
		return nil, fmt.Errorf("call creating release API: %w", err)
	}
	return release, nil
}

func (s *Client) GetLatestRelease() (*github.RepositoryRelease, error) {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	releases, _, err := s.client.Repositories.ListReleases(ctx, s.cfg.RepoOwner, s.cfg.RepoName, &github.ListOptions{Page: 1, PerPage: 10})
	if err != nil {
		return nil, err
	}
	if len(releases) > 0 {
		return releases[0], nil
	}
	return nil, nil
}

func (s *Client) IssueAssignTo(number int, assignee string) error {
	ctx, timeout := context.WithTimeout(*s.ctx, 10*time.Second)
	defer timeout()

	_, _, err := s.client.Issues.AddAssignees(ctx, s.cfg.RepoOwner, s.cfg.RepoName, number, []string{assignee})
	return err
}
