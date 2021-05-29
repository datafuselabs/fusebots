// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package actions

import (
	"bots/common"
	"bots/config"

	"github.com/go-playground/webhooks/v6/github"
	log "github.com/sirupsen/logrus"
)

type IssueAction struct {
	cfg    *config.Config
	client *common.Client
}

func NewIssueAction(cfg *config.Config) *IssueAction {
	client := common.NewClient(cfg)

	return &IssueAction{
		cfg:    cfg,
		client: client,
	}
}

func (s *IssueAction) Start() {
	log.Infof("Issue action start...")
}

func (s *IssueAction) Stop() {
}

func (s *IssueAction) DoAction(event interface{}) error {
	switch event := event.(type) {
	case github.IssueCommentPayload:
		body := event.Comment.Body
		log.Infof("Issue comments: %+v , %+v coming", event.Sender.Login, body)
		switch body {
		case "/assignme":
			s.client.IssueAssignTo(int(event.Issue.Number), event.Sender.Login)
			s.client.AddLabelToIssue(int(event.Issue.Number), "community-take")
		case "/help":
			help := common.HelpMessage()
			s.client.CreateComment(int(event.Issue.Number), &help)
		}

	}
	return nil
}
