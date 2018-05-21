package plugin

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
		IssueNum       int
		Key            string
		Message        string
		Password       string
		RepoName       string
		RepoOwner      string
		Update         bool
		Username       string
		Token          string

		gitClient      *github.Client
		gitContext     context.Context
	}
)

// Exec executes the plugin
func (p Plugin) Exec() error {
	err := validate(p)

	if err != nil {
		return err
	}

	err = p.newGitClient()

	if err != nil {
		return err
	}

	// Generate default plugin key if not specified
	if p.Key == "" {
		p.Key = defaultKey(p)
	}

	ic := &github.IssueComment{
		Body: &p.Message,
	}

	if p.Update {
		// Append plugin comment ID to comment message so we can search for it later
		message := fmt.Sprintf("%s\n<!-- id: %s -->\n", p.Message, p.Key)
		ic.Body = &message

		comment, err := p.Comment()

		if err != nil {
			return err
		}

		if *comment.ID != 0 {
			_, _, err = p.gitClient.Issues.EditComment(p.gitContext, p.RepoOwner, p.RepoName, *comment.ID, ic)
			return err
		}
	}

	_, _, err = p.gitClient.Issues.CreateComment(p.gitContext, p.RepoOwner, p.RepoName, p.IssueNum, ic)
	return err
}

func (p Plugin) newGitClient() error {
	if !strings.HasSuffix(p.BaseURL, "/") {
		p.BaseURL = p.BaseURL + "/"
	}

	baseURL, err := url.Parse(p.BaseURL)

	if err != nil {
		return fmt.Errorf("Failed to parse base URL. %s", err)
	}

	p.gitContext = context.Background()

	if p.Token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: p.Token})
		tc := oauth2.NewClient(p.gitContext, ts)
		p.gitClient = github.NewClient(tc)
	} else {
		tp := github.BasicAuthTransport{
			Username: strings.TrimSpace(p.Username),
			Password: strings.TrimSpace(p.Password),
		}
		p.gitClient = github.NewClient(tp.Client())
	}
	p.gitClient.BaseURL = baseURL

	return nil
}

// Comment returns existing comment, nil if none exist
func (p Plugin) Comment() (*github.IssueComment, error) {
	comments, err := p.allIssueComments(p.gitContext)

	if err != nil {
		return nil, err
	}

	return filterComment(comments, p.Key), nil
}

func (p Plugin) allIssueComments(ctx context.Context) ([]*github.IssueComment, error) {
	opts := &github.IssueListCommentsOptions{}

	// get all pages of results
	var allComments []*github.IssueComment
	for {
		comments, resp, err := p.gitClient.Issues.ListComments(ctx, p.RepoOwner, p.RepoName, p.IssueNum, opts)
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

func defaultKey(p Plugin) string {
	key := fmt.Sprintf("%s/%s/%d", p.RepoOwner, p.RepoName, p.IssueNum)
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

func filterComment(comments []*github.IssueComment, key string) *github.IssueComment {
	for _, comment := range comments {
		if strings.Contains(*comment.Body, fmt.Sprintf("<!-- id: %s -->", key)) {
			return comment
		}
	}

	return nil
}

func validate(p Plugin) error {
	if p.Token == "" && (p.Username == "" || p.Password == "") {
		return fmt.Errorf("You must provide an API key or Username and Password")
	}

	return nil
}
