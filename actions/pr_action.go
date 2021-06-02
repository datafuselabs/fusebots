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

type PrAction struct {
	cfg    *config.Config
	client *common.Client
}

func NewPrAction(cfg *config.Config) *PrAction {
	client := common.NewClient(cfg)

	return &PrAction{
		cfg:    cfg,
		client: client,
	}
}

func (s *PrAction) Start() {
	log.Infof("Pr action start...")
}

func (s *PrAction) Stop() {
}

func (s *PrAction) DoAction(event interface{}) error {
	switch event := event.(type) {
	case github.PullRequestPayload:
		body := event.PullRequest.Body
		log.Infof("Pr comments: %+v , %+v coming", event.Sender.Login, body)
		switch body {
		case "/runperf":
			s.client.RepositoriesDispatch("runperf")
		}

	}
	return nil
}
