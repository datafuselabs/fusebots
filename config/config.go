// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.

package config

import (
	ini "gopkg.in/ini.v1"
)

type Config struct {
	GithubToken  string
	GithubSecret string
	RepoOwner    string
	RepoName     string
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
	return cfg, nil
}
