// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"bots/actions"
	"bots/config"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/webhooks/v6/github"
	log "github.com/sirupsen/logrus"
)

const (
	path = "/webhooks"
)

var (
	flagConfig string
)

func initFlags() {
	flag.StringVar(&flagConfig, "c", "", "config file")
}

func usage() {
	fmt.Println("Usage: " + os.Args[0] + " -c fusebots.ini")
	flag.PrintDefaults()
}

func main() {
	initFlags()
	flag.Usage = func() { usage() }
	flag.Parse()

	if flagConfig == "" {
		usage()
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(flagConfig)
	if err != nil {
		log.Fatal("Load config error: %v", err)
	}
	log.Infof("Repo: %v/%v webhooks starts... ", cfg.Github.RepoOwner, cfg.Github.RepoName)
	os.Setenv("GITHUB_TOKEN", cfg.Github.GithubToken)

	// Actions.
	labelAction := actions.NewLabelerAction(cfg)
	if !cfg.Disables.DisableLabel {
		labelAction.Start()
	}

	releaseAction := actions.NewReleaseAction(cfg)
	releaseAction.Start()

	autoMergeAction := actions.NewAutoMergeAction(cfg)
	if !cfg.Disables.DisableAutoMerge {
		autoMergeAction.Start()
	}

	issueAction := actions.NewIssueAction(cfg)
	issueAction.Start()

	hook, _ := github.New(github.Options.Secret(cfg.Github.GithubSecret))
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.ReleaseEvent, github.PullRequestEvent, github.IssueCommentEvent, github.IssuesEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				log.Errorf("Unhandle github event: %v", err)
			}
		}

		if labelAction.DoAction(payload) != nil {
			log.Errorf("Labeling error: %v", err)
		}

		if issueAction.DoAction(payload) != nil {
			log.Errorf("Issue error: %v", err)
		}
	})

	http.ListenAndServe(":3000", nil)
	if !cfg.Disables.DisableLabel {
		labelAction.Stop()
	}
	releaseAction.Stop()
	if !cfg.Disables.DisableAutoMerge {
		autoMergeAction.Stop()
	}
}
