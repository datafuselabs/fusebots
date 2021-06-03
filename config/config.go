// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package config

import (
	ini "gopkg.in/ini.v1"
)

type Config struct {
	GithubToken      string
	GithubSecret     string
	RepoOwner        string
	RepoName         string
	NightReleaseCron string
	MergeCheckCron   string
	ApprovedRule     string
}

func LoadConfig(file string) (*Config, error) {
	cfg := &Config{}
	load, err := ini.Load(file)
	if err != nil {
		return nil, err
	}

	cfg.GithubToken = load.Section("github").Key("token").String()
	cfg.GithubSecret = load.Section("github").Key("secret").String()
	cfg.RepoOwner = load.Section("repo").Key("owner").String()
	cfg.RepoName = load.Section("repo").Key("name").String()

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
	return cfg, nil
}
