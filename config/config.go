// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package config

import (
	"log"

	ini "gopkg.in/ini.v1"
)

type GithubConfig struct {
	GithubToken  string `ini:"token"`
	GithubSecret string `ini:"secret"`
	RepoOwner    string `ini:"owner"`
	RepoName     string `ini:"name"`
}

type PRDescriptionActionConfig struct {
	Title       string   `ini:"title"`
	PendingDesc string   `ini:"pending_desc"`
	ErrorDesc   string   `ini:"error_desc"`
	SuccessDesc string   `ini:"success_desc"`
	TargetUrl   string   `ini:"target_url"`
	Checks      []string `ini:"checks"`
	AllowList   []string `ini:"allowlist"`
}

type Config struct {
	Github                       *GithubConfig
	PRDescriptionAction          *PRDescriptionActionConfig
	NightReleaseCron             string
	MergeCheckCron               string
	ApprovedRule                 string
	PullRequestNeedReviewComment string
	ReviewerHints                string
}

func LoadConfig(file string) (*Config, error) {
	cfg := &Config{}
	load, err := ini.Load(file)
	if err != nil {
		return nil, err
	}

	// Github.
	cfg.Github = new(GithubConfig)
	if err := load.Section("github").MapTo(cfg.Github); err != nil {
		log.Fatalf("Can not load gihutb section:%+v", err)
	}

	// PR desc action.
	cfg.PRDescriptionAction = new(PRDescriptionActionConfig)
	if err := load.Section("pr_description_action").MapTo(cfg.PRDescriptionAction); err != nil {
		log.Fatalf("Can not load pr description action section:%+v", err)
	}
	log.Printf("Pr desc action action:%+v", cfg.PRDescriptionAction)

	// Schedule.
	cfg.NightReleaseCron = load.Section("schedule").Key("nightly_release_cron").String()
	if cfg.NightReleaseCron == "" {
		cfg.NightReleaseCron = "@daliy"
	}
	cfg.MergeCheckCron = load.Section("schedule").Key("merge_check_cron").String()
	if cfg.MergeCheckCron == "" {
		cfg.MergeCheckCron = "@every 30s"
	}

	// Rule.
	cfg.ApprovedRule = load.Section("rule").Key("approved_rule").String()
	if cfg.ApprovedRule == "" {
		cfg.ApprovedRule = "most"
	}

	// Comments.
	cfg.PullRequestNeedReviewComment = load.Section("comment").Key("pr_need_review_comment").String()

	// Review.
	cfg.ReviewerHints = load.Section("reviewer").Key("hints").String()
	if cfg.ReviewerHints == "" {
		cfg.ReviewerHints = "@BohuTANG"
	}

	return cfg, nil
}
