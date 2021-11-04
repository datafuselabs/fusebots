// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package actions

import (
	"bots/common"
	"bots/config"
	"fmt"
	"strings"

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
		switch body := strings.ToLower(body); {
		case strings.HasPrefix(body, "/assignme"):
			{
				s.client.IssueAssignTo(int(event.Issue.Number), event.Sender.Login)
				s.client.AddLabelToIssue(int(event.Issue.Number), "community-take")
			}
		case strings.HasPrefix(body, "/review "):
			{
				user := strings.TrimSpace(strings.TrimPrefix(body, "/review @"))
				if err := s.client.PullRequestRequestReviewer(int(event.Issue.Number), user); err != nil {
					return err
				}
				msg := "Take the reviewer to " + user
				s.client.CreateComment(int(event.Issue.Number), &msg)
			}

		case strings.HasPrefix(body, "/approve"), strings.HasPrefix(body, "/lgtm"):
			{
				if err := s.client.PullRequestReview(int(event.Issue.Number), "APPROVE"); err != nil {
					return err
				}
				if err := s.prMergeStateChange(int(event.Issue.Number)); err != nil {
					return err
				}

				msg := fmt.Sprintf("Approved by %s!", event.Comment.User.Login)
				s.client.CreateComment(int(event.Issue.Number), &msg)

			}
		case strings.HasPrefix(body, "/help"):
			{
				help := common.HelpMessage()
				s.client.CreateComment(int(event.Issue.Number), &help)
			}

		}

	case github.IssuesPayload:
		if event.Issue.State == "open" {
			first, err := s.client.IssuesForFirstTime(event.Issue.User.Login)
			if err != nil {
				return err
			}
			if first {
				comments := fmt.Sprintf(s.cfg.Hints.IssueFirstTimeComment, event.Issue.User.Login)
				s.client.CreateComment(int(event.Issue.Number), &comments)
			}
		}

	}
	return nil
}

func (s *IssueAction) prMergeStateChange(number int) error {
	labels, err := s.client.ListLabelsForIssue(number)
	if err != nil {
		return err
	}
	newLabels := make([]string, len(labels))
	for _, l := range labels {
		switch *l.Name {
		case "need-review":
			{
				newLabels = append(newLabels, "lgtm1")
			}
		case "lgtm1":
			{
				newLabels = append(newLabels, "lgtm2")
			}
		case "lgtm2": // no need change, save one network request:)
			return nil
		default:
			newLabels = append(newLabels, *l.Name)
		}
	}

	return s.client.ReplaceLabelsForIssue(number, newLabels...)
}
