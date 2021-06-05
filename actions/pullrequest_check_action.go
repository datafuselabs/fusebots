// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.
// Some codes from https://github.com/p1ass/mikku

package actions

import (
	"bots/common"
	"bots/config"
	"fmt"

	"github.com/go-playground/webhooks/v6/github"
	log "github.com/sirupsen/logrus"
)

const (
	state_pending = "pending"
	state_error   = "error"
	state_success = "success"
)

type PullRequestCheckAction struct {
	cfg    *config.Config
	client *common.Client
}

func NewPullRequestCheckAction(cfg *config.Config) *PullRequestCheckAction {
	client := common.NewClient(cfg)
	return &PullRequestCheckAction{
		cfg:    cfg,
		client: client,
	}
}

func (s *PullRequestCheckAction) Start() {
}

func (s *PullRequestCheckAction) Stop() {
}

func (s *PullRequestCheckAction) DoAction(event interface{}) error {
	switch event := event.(type) {
	case github.PullRequestPayload:
		log.Infof("Pull request check: %+v coming", event.Number)

		if err := s.descriptionCheck(event); err != nil {
			log.Errorf("Desciption check error: %+v ", err)
		}

		if err := s.reviewerCheck(event); err != nil {
			log.Errorf("Reviewer check error: %+v ", err)
		}

	}
	return nil
}

func (s *PullRequestCheckAction) descriptionCheck(payload github.PullRequestPayload) error {
	pr := payload.PullRequest
	sha := pr.Head.Sha
	title := "Description check"
	desc := "Checking"
	url := "https://datafuse.rs"

	go func() {
		if err := s.client.CreateStatus(sha, title, desc, state_pending, url); err != nil {
			log.Errorf("Desciption check status create error: %+v ", err)
			return
		}

		if err := s.client.CreateStatus(sha, title, desc, state_success, url); err != nil {
			log.Errorf("Desciption check status create error: %+v ", err)
			return
		}

	}()
	return nil
}

func (s *PullRequestCheckAction) reviewerCheck(payload github.PullRequestPayload) error {
	pr := payload.PullRequest
	// Pr need reviewer.
	if !pr.Draft && *pr.Mergeable {
		reviewers, err := s.client.PullRequestListReviewers(int(pr.Number))
		if err != nil {
			return err
		}
		if len(reviewers.Users) == 0 {
			if s.cfg.PullRequestNeedReviewComment != "" {
				comments := fmt.Sprintf(s.cfg.PullRequestNeedReviewComment, pr.User.Login)
				comments += s.cfg.ReviewerHints
				s.client.CreateComment(int(pr.Number), &comments)
			}
		}
	}

	return nil
}
