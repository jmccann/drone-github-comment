package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type (
	Plugin struct {
		BaseURL        string
		BuildEvent     string
		Message        string
		PullRequestNum int
		RepoName       string
		RepoOwner      string
		Token          string
	}
)

func (p Plugin) Exec() error {
	err := validate(p)

	if err != nil {
		return err
	}

	if !strings.HasSuffix(p.BaseURL, "/") {
		p.BaseURL = p.BaseURL + "/"
	}

	baseURL, err := url.Parse(p.BaseURL)

	if err != nil {
		return fmt.Errorf("Failed to parse base URL. %s", err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: p.Token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	client.BaseURL = baseURL

	ic := &github.IssueComment{
		Body: &p.Message,
	}

	client.Issues.CreateComment(ctx, p.RepoOwner, p.RepoName, p.PullRequestNum, ic)

	return nil
}

func validate(p Plugin) error {
	if p.BuildEvent != "pull_request" {
		return fmt.Errorf("The GitHub PR plugin is only available for pull reuqests")
	}

	if p.Token == "" {
		return fmt.Errorf("You must provide an API key")
	}

	return nil
}
