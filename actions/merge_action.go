// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.
// Some codes from https://github.com/p1ass/mikku

package actions

import (
	"bots/config"
	"context"

	"github.com/google/go-github/v35/github"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type AutoMergeAction struct {
	cfg    *config.Config
	cron   *cron.Cron
	client *github.Client
}

func NewAutoMergeAction(cfg *config.Config) *AutoMergeAction {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: cfg.GithubToken,
	})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &AutoMergeAction{
		cfg:    cfg,
		cron:   cron.New(),
		client: client,
	}
}

func (s *AutoMergeAction) autoMergeCron() {
	prs, err := s.listOpenPullRequests()
	if err != nil {
		log.Errorf("List open pull requests error:%v", err)
	}
	for _, pr := range prs {
		shouldMerge, err := s.shouldMergePR(pr)
		if err != nil {
			log.Errorf("Check should merge pr error:%v", err)
		}

		if shouldMerge {
			opts := github.PullRequestOptions{
				MergeMethod: "merge",
			}

			ctx := context.Background()
			result, _, err := s.client.PullRequests.Merge(ctx, s.cfg.RepoOwner, s.cfg.RepoName, pr.GetNumber(), "", &opts)
			if err != nil {
				log.Errorf("Do merge error:%+v", err)
			}
			log.Infof("Merge %v, sha:%v succuess", pr.GetNumber(), result.GetSHA())
		}
	}
}

func (s *AutoMergeAction) Start() {
	s.cron.AddFunc(s.cfg.MergeCheckCron, s.autoMergeCron)
	s.cron.Start()
	log.Infof("AutoMerge action start:%v...", s.cfg.MergeCheckCron)
}

func (s *AutoMergeAction) Stop() {
	s.cron.Stop()
}

func (s *AutoMergeAction) listOpenPullRequests() ([]*github.PullRequest, error) {
	var results []*github.PullRequest

	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	ctx := context.Background()
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

func (s *AutoMergeAction) shouldMergePR(pr *github.PullRequest) (bool, error) {
	allowedCheckConclusions := map[string]bool{
		"success": true,
		"neutral": true,
		"skipped": true,
	}

	// Draft.
	if pr.GetDraft() {
		return false, nil
	}

	ctx := context.Background()
	opts := &github.ListCheckRunsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	checkRuns, _, err := s.client.Checks.ListCheckRunsForRef(ctx, s.cfg.RepoOwner, s.cfg.RepoName, pr.GetHead().GetSHA(), opts)
	if err != nil {
		return false, err
	}
	for _, s := range checkRuns.CheckRuns {
		if !allowedCheckConclusions[s.GetConclusion()] {
			log.Infof("Check run:%v, status:%s, %v", pr.GetTitle(), s.GetName(), s.GetConclusion())
			return false, nil
		}
	}

	listOpts := &github.ListOptions{PerPage: 100}
	reviewers, _, err := s.client.PullRequests.ListReviewers(ctx, s.cfg.RepoOwner, s.cfg.RepoName, pr.GetNumber(), listOpts)
	if err != nil {
		return false, err
	}

	reviews, _, err := s.client.PullRequests.ListReviews(ctx, s.cfg.RepoOwner, s.cfg.RepoName, pr.GetNumber(), listOpts)
	if err != nil {
		return false, err
	}

	for _, review := range reviews {
		log.Infof("Review name:%v, status:%v", review.GetUser().GetLogin(), review.GetState())
		if review.GetState() == "APPROVED" {
			for _, user := range reviewers.Users {
				if review.GetUser().GetID() == user.GetID() {
					log.Infof("PR apprevoed by :%v", review.GetUser().GetName())
					return true, nil
				}
			}
		}
	}
	return false, nil
}
