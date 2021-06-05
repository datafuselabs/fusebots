// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package config

import (
	"log"

	ini "gopkg.in/ini.v1"
)

type GithubConfig struct {
	GithubToken  string `ini:token`
	GithubSecret string `ini:"secret"`
	RepoOwner    string `ini:"owner"`
	RepoName     string `ini:"name"`
}

type Config struct {
	Github                       *GithubConfig
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

	cfg.Github = new(GithubConfig)
	if err := load.Section("github").MapTo(cfg.Github); err != nil {
		log.Fatalf("Can not load gihutb section:%+v", err)
	}

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
