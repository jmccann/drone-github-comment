package plugin

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-github/github"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

type (
	Plugin struct {
		BaseURL      string
		IssueNum     int
		Key          string
		Message      string
		Password     string
		RepoName     string
		RepoOwner    string
		Update       bool
		DeleteCreate bool
		Username     string
		Token        string

		gitClient  *github.Client
		gitContext context.Context
	}
)

func NewFromCLI(c *cli.Context) (*Plugin, error) {
	p := Plugin{
		BaseURL:      c.String("base-url"),
		Key:          c.String("key"),
		Message:      c.String("message"),
		IssueNum:     c.Int("issue-num"),
		Password:     c.String("password"),
		RepoName:     c.String("repo-name"),
		RepoOwner:    c.String("repo-owner"),
		Token:        c.String("api-key"),
		Update:       c.Bool("update"),
		DeleteCreate: c.Bool("delete_create"),
		Username:     c.String("username"),
	}

	err := p.init()

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func NewFromPlugin(p Plugin) (*Plugin, error) {
	err := p.init()

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// Exec executes the plugin
func (p Plugin) Exec() error {
	fmt.Println("exec1")
	if p.gitClient == nil {
		return fmt.Errorf("Exec(): git client not initialized")
	}

	ic := &github.IssueComment{
		Body: &p.Message,
	}

	var err error
	if p.Update {
		fmt.Println("update")
		// Append plugin comment ID to comment message so we can search for it later
		message := fmt.Sprintf("%s\n<!-- id: %s -->\n", p.Message, p.Key)
		ic.Body = &message

		comment, err := p.Comment()

		if err != nil {
			return err
		}

		if comment != nil {
			_, _, err = p.gitClient.Issues.EditComment(p.gitContext, p.RepoOwner, p.RepoName, int(*comment.ID), ic)
			return err
		}
	}

	if p.DeleteCreate {
		fmt.Println("delete")
		// Append plugin comment ID to comment message so we can search for it later
		message := fmt.Sprintf("%s\n<!-- id: %s -->\n", p.Message, p.Key)
		ic.Body = &message

		comment, err := p.Comment()

		if err != nil {
			return err
		}

		if comment != nil {
			_, err = p.gitClient.Issues.DeleteComment(p.gitContext, p.RepoOwner, p.RepoName, int(*comment.ID))
			if err != nil {
				return err
			}
		}
		_, _, err = p.gitClient.Issues.CreateComment(p.gitContext, p.RepoOwner, p.RepoName, p.IssueNum, ic)
		return err
	}

	fmt.Println("create")
	_, _, err = p.gitClient.Issues.CreateComment(p.gitContext, p.RepoOwner, p.RepoName, p.IssueNum, ic)
	return err
}

func (p *Plugin) init() error {
	err := p.validate()

	if err != nil {
		return err
	}

	err = p.initGitClient()

	if err != nil {
		return err
	}

	// Generate default plugin key if not specified
	if p.Key == "" {
		p.Key = defaultKey(*p)
	}

	return nil
}

func (p *Plugin) initGitClient() error {
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
	if p.gitClient == nil {
		return nil, fmt.Errorf("allIssueComments(): git client not initialized")
	}

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

func (p Plugin) validate() error {
	if p.Token == "" && (p.Username == "" || p.Password == "") {
		return fmt.Errorf("You must provide an API key or Username and Password")
	}

	return nil
}
