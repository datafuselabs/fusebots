// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package actions

import (
	"bots/config"
	"encoding/json"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/jimschubert/labeler"
	log "github.com/sirupsen/logrus"
)

type LabelerAction struct {
	cfg *config.Config
}

func NewLabelerAction(cfg *config.Config) *LabelerAction {
	return &LabelerAction{
		cfg: cfg,
	}
}

func (s *LabelerAction) Start() {
	log.Infof("Labeler action start...")
}

func (s *LabelerAction) Stop() {
}

func (s *LabelerAction) DoAction(event interface{}) error {
	switch event.(type) {
	case github.PullRequestPayload:
		pr := event.(github.PullRequestPayload)
		log.Infof("Pull reqeust: %+v coming", pr.Number)
		body, _ := json.Marshal(pr)
		data := string(body)
		l, err := labeler.New(s.cfg.Github.RepoOwner, s.cfg.Github.RepoName, "pull_request", int(pr.Number), &data)
		if err != nil {
			return err
		}

		log.Infof("Labeling prepare...")
		err = l.Execute()
		if err != nil {
			return err
		}
		log.Infof("Labeling done...")
	}
	return nil
}
