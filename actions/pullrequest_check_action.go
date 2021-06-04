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
	switch event.(type) {
	case github.PullRequestPayload:
		pr := event.(github.PullRequestPayload).PullRequest
		log.Infof("Pull request check: %+v coming", pr.Number)

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

	}
	return nil
}
