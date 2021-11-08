// Copyright 2020-2021 The Datafuselabs Authors.
//
// SPDX-License-Identifier: Apache-2.0.
// Some codes from https://github.com/p1ass/mikku

package actions

import (
	"bots/common"
	"bots/config"
	"fmt"

	"github.com/google/go-github/v35/github"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type AutoMergeAction struct {
	cfg    *config.Config
	cron   *cron.Cron
	client *common.Client
}

func NewAutoMergeAction(cfg *config.Config) *AutoMergeAction {
	client := common.NewClient(cfg)
	return &AutoMergeAction{
		cfg:    cfg,
		cron:   cron.New(),
		client: client,
	}
}

func (s *AutoMergeAction) autoMergeCron() {
	prs, err := s.client.PullRequestList()
	if err != nil {
		log.Errorf("List open pull requests error:%v", err)
	}

	for _, pr := range prs {
		approveCount, err := s.shouldMergePR(pr)
		if err != nil {
			log.Errorf("Check should merge pr error:%v", err)
			continue
		}

		approve_comments := "Wait for another reviewer approval"
		ci_passed_comments := fmt.Sprintf("CI Passed\nReviewers Approved\nLet's Merge\nThank you for the PR @%s", *pr.User.Login)
		switch approveCount {
		case -1, 0:

		case 1:
			// Check has approved comment.
			lastComment, err := s.client.GetLastComment(pr.GetNumber())
			if err != nil {
				log.Errorf("Get last comments error:%+v", err)
				continue
			}
			if lastComment != nil && (*lastComment.Body == approve_comments) {
				log.Warn("PR:%+v has 1 proved", pr.GetNumber())
			} else {
				s.client.CreateComment(pr.GetNumber(), &approve_comments)
			}
		default:
			// Check is approved.
			lastComment, err := s.client.GetLastComment(pr.GetNumber())
			if err != nil {
				log.Errorf("Get last comments error:%+v", err)
				continue
			}
			log.Infof("%v last comments: %v", pr.GetNumber(), lastComment)

			if lastComment != nil && (*lastComment.Body == ci_passed_comments) {
				log.Warn("PR:%+v has proved", pr.GetNumber())
			} else {
				s.client.CreateComment(pr.GetNumber(), &ci_passed_comments)
			}

			log.Warn("PR:%+v try to merge", pr.GetNumber())
			if err := s.client.PullRequestMerge(pr.GetNumber(), ""); err != nil {
				log.Errorf("Do merge error:%+v", err)
			}
			log.Warn("PR:%+v merge send", pr.GetNumber())
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

func (s *AutoMergeAction) shouldMergePR(pr *github.PullRequest) (int, error) {
	allowedCheckConclusions := map[string]bool{
		"success": true,
		"neutral": true,
		"skipped": true,
	}

	if pr.GetMerged() {
		log.Infof("%v merged...", pr.GetNumber())
		return -1, nil
	}

	// Draft.
	if pr.GetDraft() {
		log.Infof("%v in draft...", pr.GetNumber())
		return -1, nil
	}

	checkRuns, err := s.client.ListCheckRunsForRef(pr.GetHead().GetSHA())
	if err != nil {
		return -1, err
	}

	for _, s := range checkRuns.CheckRuns {
		if !allowedCheckConclusions[s.GetConclusion()] {
			log.Infof("Check run:%v, status:%s, %v", pr.GetTitle(), s.GetName(), s.GetConclusion())
			return -1, nil
		}
	}

	reviews, err := s.client.PullRequestListReviews(pr.GetNumber())
	if err != nil {
		return -1, err
	}

	labels, err := s.client.ListLabelsForIssue(pr.GetNumber())
	if err != nil {
		return -1, err
	}

	approveCount := 0
	for _, review := range reviews {
		if review.GetState() == "APPROVED" {
			approveCount++
		}
	}

	if approveCount == 2 {
		for _, l := range labels {
			if *l.Name == "need-review" {
				s.client.RemoveLabelFromIssue(pr.GetNumber(), *l.Name)
			}
		}
	}

	// if bot approve twice, check label
	if approveCount == 1 {
		for _, l := range labels {
			if *l.Name == "lgtm2" {
				approveCount = 2
			}
		}
	}

	return approveCount, nil
}
