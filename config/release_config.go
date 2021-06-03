// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.
// Some codes from https://github.com/p1ass/mikku

package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Category struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`
}

type Categories struct {
	Categories []Category `yaml:"categories"`
	Excludes   []string   `yaml:"excludes"`
}

type ReleaseConfig struct {
	file       string
	Categories *Categories
}

func NewReleaseConfig(file string) *ReleaseConfig {
	return &ReleaseConfig{
		file: file,
	}
}

func (s *ReleaseConfig) Load() error {
	file, err := ioutil.ReadFile(s.file)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(file, &s.Categories); err != nil {
		return err
	}
	return nil
}

func (s *ReleaseConfig) GetCategoryByLabel(label string) string {
	for _, category := range s.Categories.Categories {
		for _, l := range category.Labels {
			if l == label {
				return category.Title
			}
		}
	}
	return "Others"
}

func (s *ReleaseConfig) ExcludeCheck(label string) bool {
	for _, exclude := range s.Categories.Excludes {
		if label == exclude {
			return true
		}
	}
	return false
}
