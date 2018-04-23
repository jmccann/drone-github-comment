package main

import (
	"crypto/sha256"
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
		ID             string
		IssueNum       int
		Message        string
		RepoName       string
		RepoOwner      string
		UpdateExisting bool
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

	// Generate plugin comment ID if not specified
	if p.ID == "" {
		key := fmt.Sprintf("%s/%s/%s", p.RepoOwner, p.RepoName, p.IssueNum)
		hash := sha256.Sum256([]byte(key))
		p.ID = string(hash[:sha256.Size])
	}

	// Append plugin comment ID to comment message so we can search for it later
	message := fmt.Sprintf("%s\n<!-- id: %s -->\n", p.Message, p.ID)

	ic := &github.IssueComment{
		Body: &message,
	}

	if p.UpdateExisting {
		comments, err := allIssueComments(ctx, client, p)

		if err != nil {
			return err
		}

		commentID := filterComment(comments, p.ID)

		if commentID != 0 {
			_, _, err = client.Issues.EditComment(ctx, p.RepoOwner, p.RepoName, commentID, ic)
			return err
		}
	}

	_, _, err = client.Issues.CreateComment(ctx, p.RepoOwner, p.RepoName, p.IssueNum, ic)
	return err
}

func allIssueComments(ctx context.Context, client *github.Client, p Plugin) ([]*github.IssueComment, error) {
	opts := &github.IssueListCommentsOptions{}

	// get all pages of results
	var allComments []*github.IssueComment
	for {
		comments, resp, err := client.Issues.ListComments(ctx, p.RepoOwner, p.RepoName, p.IssueNum, opts)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

func filterComment(comments []*github.IssueComment, id string) int {
	for _, comment := range comments {
		if strings.Contains(*comment.Body, fmt.Sprintf("<!-- id: %s -->", id)) {
			return *comment.ID
		}
	}

	return 0
}

func validate(p Plugin) error {
	if p.Token == "" {
		return fmt.Errorf("You must provide an API key")
	}

	return nil
}
