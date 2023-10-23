package plugin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

type (
	Plugin struct {
		BaseURL   string
		IssueNum  int
		Key       string
		Section   string
		Message   string
		Password  string
		RepoName  string
		RepoOwner string
		Update    bool
		Username  string
		Token     string

		gitClient  *github.Client
		gitContext context.Context
	}
)

func NewFromCLI(c *cli.Context) (*Plugin, error) {
	p := Plugin{
		BaseURL:   c.String("base-url"),
		Key:       c.String("key"),
		Section:   c.String("section"),
		Message:   c.String("message"),
		IssueNum:  c.Int("issue-num"),
		Password:  c.String("password"),
		RepoName:  c.String("repo-name"),
		RepoOwner: c.String("repo-owner"),
		Token:     c.String("api-key"),
		Update:    c.Bool("update"),
		Username:  c.String("username"),
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
	if p.gitClient == nil {
		return fmt.Errorf("Exec(): git client not initialized")
	}

	ic := &github.IssueComment{
		Body: &p.Message,
	}

	if p.Update {
		message, err := formatTmpl(p.Section, "\n", p.Message)
		if err != nil {
			return err
		}

		// Prepend plugin comment ID to comment message so we can search for it later
		body := fmt.Sprintf("<!-- id: %s -->\n%s\n", p.Key, message)
		ic.Body = &body

		comment, err := p.Comment()
		if err != nil {
			return err
		}

		if comment != nil {
			// If comment exists, filter comment body to look for section key
			if hasSection(comment.GetBody(), p.Section) {
				// If section exists, update it
				body, err := updateSection(comment.GetBody(), p.Section, p.Message)
				if err != nil {
					return err
				}
				ic.Body = &body
			} else {
				// Otherwise add section
				body := fmt.Sprintf("%s\n\n%s\n", comment.GetBody(), message)
				ic.Body = &body
			}

			_, _, err := p.gitClient.Issues.EditComment(p.gitContext, p.RepoOwner, p.RepoName, int(*comment.ID), ic)
			return err
		}
	}

	_, _, err := p.gitClient.Issues.CreateComment(p.gitContext, p.RepoOwner, p.RepoName, p.IssueNum, ic)
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

	// Generate default plugin section if not specified
	if p.Section == "" {
		p.Section = "main"
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

const tmpl = `<!-- start: {{.Section}} -->{{.NL}}{{.Message}}{{.NL}}<!-- end: {{.Section}} -->`

func formatTmpl(section, newline, message string) (string, error) {
	t := template.Must(template.New("tmpl").Parse(tmpl))
	data := map[string]string{
		"Section": section,
		"NL":      newline,
		"Message": message,
	}
	buf := &bytes.Buffer{}
	err := t.Execute(buf, data)
	return buf.String(), err
}

func hasSection(body, section string) bool {
	return strings.Contains(body, fmt.Sprintf("<!-- start: %s -->", section))
}

func updateSection(body, section, message string) (string, error) {
	expression, err := formatTmpl(section, "\\n", ".*")
	if err != nil {
		return "", err
	}

	msg, err := formatTmpl(section, "\n", message)
	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(expression)
	if err != nil {
		return "", err
	}

	return re.ReplaceAllString(body, msg), nil
}
