// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package config

import (
	"log"

	ini "gopkg.in/ini.v1"
)

type DisablesConfig struct {
	DisableAutoMerge bool `ini:"disable_auto_merge"`
}

type HintConfig struct {
	IssueFirstTimeComment string `ini:"issue_first_time_comment"`
	PRNeedReviewComment   string `ini:"pr_need_review_comment"`
}

type GithubConfig struct {
	GithubToken  string `ini:"token"`
	GithubSecret string `ini:"secret"`
	RepoOwner    string `ini:"owner"`
	RepoName     string `ini:"name"`
	BaseBranch   string `ini:"base_branch"`
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
	Github              *GithubConfig
	PRDescriptionAction *PRDescriptionActionConfig
	Hints               *HintConfig
	Disables            *DisablesConfig
	NightReleaseCron    string
	MergeCheckCron      string
	ApprovedRule        string
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
	if cfg.Github.BaseBranch == "" {
		cfg.Github.BaseBranch = "main"
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

	// Hints.
	cfg.Hints = new(HintConfig)
	if err := load.Section("hint").MapTo(cfg.Hints); err != nil {
		log.Fatalf("Can not load hint section:%+v", err)
	}
	log.Printf("Hint action:%+v", cfg.Hints)

	// Disables.
	cfg.Disables = new(DisablesConfig)
	if err := load.Section("disables").MapTo(cfg.Disables); err != nil {
		log.Fatalf("Can not load disables section:%+v", err)
	}
	log.Printf("Disables conf:%+v", cfg.Disables)

	return cfg, nil
}
