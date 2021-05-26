// Copyright 2020-2021 The Datafuse Authors.
//
// SPDX-License-Identifier: Apache-2.0.
// Some codes from https://github.com/p1ass/mikku

package actions

import (
	"bots/common"
	"bots/config"
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	baseBranch  = "master"
	listPerPage = 10
)

type ReleaseAction struct {
	cfg    *config.Config
	cron   *cron.Cron
	client *github.Client
	yml    *config.ReleaseYml
}

func NewReleaseAction(cfg *config.Config) *ReleaseAction {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: cfg.GithubToken,
	})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	yml := config.NewReleaseYML(".github/release.yml")

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

	prs, err := s.getMergedPRsAfter(after)
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
		if _, err := s.createRelease(newTagName, releaseBody, preRelease); err != nil {
			return err
		}
		log.Infof("release: %v done", newTagName)
	}
	return nil
}

func (s *ReleaseAction) getLatestRelease() (*github.RepositoryRelease, error) {
	ctx := context.Background()
	releases, _, err := s.client.Repositories.ListReleases(ctx, s.cfg.RepoOwner, s.cfg.RepoName, &github.ListOptions{Page: 1, PerPage: 10})
	if err != nil {
		return nil, err
	}
	if len(releases) > 0 {
		return releases[0], nil
	}
	return nil, nil
}

func (s *ReleaseAction) getLastPublishedAndCurrentTag() (time.Time, string, error) {
	after := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	tag := ""
	release, err := s.getLatestRelease()
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

func (s *ReleaseAction) getMergedPRsAfter(after time.Time) ([]*github.PullRequest, error) {
	opt := &github.PullRequestListOptions{
		State:       "closed",
		Base:        baseBranch,
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: listPerPage},
	}

	ctx := context.Background()
	var prList []*github.PullRequest
	for {
		prs, resp, err := s.client.PullRequests.List(ctx, s.cfg.RepoOwner, s.cfg.RepoName, opt)
		if err != nil {
			return nil, fmt.Errorf("call listing pull requests API: %w", err)
		}

		extractedPR, done := extractMergedPRsAfter(prs, after)
		prList = append(prList, extractedPR...)
		if done {
			break
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return prList, nil
}

func extractMergedPRsAfter(prs []*github.PullRequest, after time.Time) ([]*github.PullRequest, bool) {
	var prList []*github.PullRequest
	done := false
	for _, pr := range prs {
		if pr.MergedAt != nil && pr.MergedAt.After(after) {
			prList = append(prList, pr)
		}
		if pr.UpdatedAt != nil && !pr.UpdatedAt.After(after) {
			done = true
			break
		}
	}
	return prList, done
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

func (s *ReleaseAction) createRelease(tagName, body string, preRelease bool) (*github.RepositoryRelease, error) {
	ctx := context.Background()
	release, _, err := s.client.Repositories.CreateRelease(ctx, s.cfg.RepoOwner, s.cfg.RepoName, &github.RepositoryRelease{
		TagName:    github.String(tagName),
		Name:       github.String(tagName),
		Body:       github.String(body),
		Prerelease: &preRelease,
	})
	if err != nil {
		return nil, fmt.Errorf("call creating release API: %w", err)
	}
	return release, nil
}
