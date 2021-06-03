// Copyright 2020-2021 The Datafuse Authors.
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
		shouldMerge, err := s.shouldMergePR(pr)
		if err != nil {
			log.Errorf("Check should merge pr error:%v", err)
		}

		if shouldMerge {
			comments := fmt.Sprintf("CI Passed\nReviewer Approved\nLet's Merge")
			s.client.CreateComment(pr.GetNumber(), &comments)

			if err := s.client.PullRequestMerge(pr.GetNumber(), ""); err != nil {
				log.Errorf("Do merge error:%+v", err)
				continue
			}
			log.Infof("Merge %v succuess", pr.GetNumber())
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

	checkRuns, err := s.client.ListCheckRunsForRef(pr.GetHead().GetSHA())
	if err != nil {
		return false, err
	}

	for _, s := range checkRuns.CheckRuns {
		if !allowedCheckConclusions[s.GetConclusion()] {
			log.Infof("Check run:%v, status:%s, %v", pr.GetTitle(), s.GetName(), s.GetConclusion())
			return false, nil
		}
	}

	reviews, err := s.client.PullRequestListReviews(pr.GetNumber())
	if err != nil {
		return false, err
	}

	reviewers, err := s.client.PullRequestListReviewers(pr.GetNumber())
	if err != nil {
		return false, err
	}

	approves := 0
	for _, review := range reviews {
		name := review.GetUser().GetLogin()
		state := review.GetState()
		log.Infof("Review name:%v, status:%v", name, state)
		if review.GetState() == "APPROVED" {
			for _, user := range reviewers.Users {
				if user.GetLogin() == name {
					approves++
					log.Infof("Review name:%v approved:%v", user.GetLogin(), approves)
					break
				}
			}
		}
	}

	major := (len(reviewers.Users) + 1) / 2
	log.Infof("PR approved(rule:%v) with approved:%v of major %v", s.cfg.ApprovedRule, approves, major)
	switch s.cfg.ApprovedRule {
	case "most":
		if approves != 0 {
			if approves >= major {
				return true, nil
			}
		}

	case "less":
		if approves > 0 {
			return true, nil
		}
	}

	return false, nil
}
