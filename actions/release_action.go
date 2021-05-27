// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.
// Some codes from https://github.com/p1ass/mikku

package actions

import (
	"bots/common"
	"bots/config"
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

const (
	baseBranch  = "master"
	listPerPage = 10
)

type ReleaseAction struct {
	cfg    *config.Config
	cron   *cron.Cron
	client *common.Client
	yml    *config.ReleaseConfig
}

func NewReleaseAction(cfg *config.Config) *ReleaseAction {
	client := common.NewClient(cfg)
	yml := config.NewReleaseConfig(".github/release.yml")

	return &ReleaseAction{
		cfg:    cfg,
		cron:   cron.New(),
		client: client,
		yml:    yml,
	}
}

func (s *ReleaseAction) nightReleaseCron() {
	if err := s.releaseHandle("patch", true); err != nil {
		log.Errorf("build release log error:%+v", err)
	}
}

func (s *ReleaseAction) Start() {
	if err := s.yml.Load(); err != nil {
		log.Panicf("Can not load release yml:%+v", err)
	}

	s.cron.AddFunc(s.cfg.NightReleaseCron, s.nightReleaseCron)
	s.cron.Start()
	log.Infof("Release action start...")
}

func (s *ReleaseAction) Stop() {
	s.cron.Stop()
}

func (s *ReleaseAction) releaseHandle(typ string, preRelease bool) error {
	after, currentTag, err := s.getLastPublishedAndCurrentTag()
	if err != nil {
		return err
	}
	newTagName, err := common.DetermineNewTag(currentTag, typ)
	if err != nil {
		return err
	}
	log.Infof("Latest tag:%v, new tag:%v, type:%v", currentTag, newTagName, typ)

	prs, err := s.client.GetMergedPullRequestsAfter(baseBranch, after)
	if err != nil {
		return err
	}
	if len(prs) > 0 {
		var labelPr = make(map[string][]*github.PullRequest)
		for _, pr := range prs {
			for _, label := range pr.Labels {
				title := s.yml.GetCategoryByLabel(label.GetName())
				prs = labelPr[title]
				if prs == nil {
					prs = make([]*github.PullRequest, 0)
				}
				exists := false
				for _, check := range prs {
					if *check.Number == *pr.Number {
						exists = true
					}
				}
				if !exists {
					prs = append(prs, pr)
				}
				labelPr[title] = prs
			}
		}
		releaseBody, err := generateReleaseBody(labelPr)
		if err != nil {
			return err
		}

		log.Infof("prepare release: %v, %v", newTagName, releaseBody)
		if _, err := s.client.CreateRelease(newTagName, releaseBody, preRelease); err != nil {
			return err
		}
		log.Infof("release: %v done", newTagName)
	}
	return nil
}

func (s *ReleaseAction) getLastPublishedAndCurrentTag() (time.Time, string, error) {
	tag := ""
	after := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	release, err := s.client.GetLatestRelease()
	if err != nil {
		return after, "", fmt.Errorf("get latest release: %w", err)
	}
	if release == nil {
		return after, "", fmt.Errorf("release not found")
	}

	after = release.GetPublishedAt().Time
	tag = release.GetTagName()
	return after, tag, nil
}

func generateReleaseBody(prs map[string][]*github.PullRequest) (string, error) {
	const (
		releaseBodyTemplate = `
# What's Changed
{{ range $key, $value := .prs}}
## {{$key}}
{{ range $i, $pr := $value}}
- {{ $pr.Title }} (#{{ $pr.Number }}) by @{{ $pr.User.Login }}{{ end }}
{{ end }}
`
	)

	tmpl, err := template.New("body").Parse(releaseBodyTemplate)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	buff := bytes.NewBuffer([]byte{})
	body := map[string]interface{}{"prs": prs}

	if err := tmpl.Execute(buff, body); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}
	return buff.String(), nil
}
