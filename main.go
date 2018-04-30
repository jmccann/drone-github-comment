package main

import (
	"io/ioutil"
	"os"

	p "github.com/jmccann/drone-github-comment/plugin"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var revision string // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "github comment plugin"
	app.Usage = "github comment plugin"
	app.Action = run
	app.Version = revision
	app.Flags = []cli.Flag{

		//
		// plugin args
		//

		cli.StringFlag{
			Name:   "api-key",
			Usage:  "api key to access github api",
			EnvVar: "PLUGIN_API_KEY,GITHUB_RELEASE_API_KEY,GITHUB_TOKEN",
		},
		cli.StringFlag{
			Name:   "username",
			Usage:  "basic auth username",
			EnvVar: "PLUGIN_USERNAME,GITHUB_GITHUB_USERNAME,DRONE_NETRC_USERNAME",
		},
		cli.StringFlag{
			Name:   "password",
			Usage:  "basic auth password",
			EnvVar: "PLUGIN_PASSWORD,GITHUB_PASSWORD,DRONE_NETRC_PASSWORD",
		},
		cli.StringFlag{
			Name:   "base-url",
			Value:  "https://api.github.com/",
			Usage:  "api url, needs to be changed for ghe",
			EnvVar: "PLUGIN_BASE_URL,GITHUB_BASE_URL",
		},
		cli.IntFlag{
			Name:   "issue-num",
			Usage:  "Issue #",
			EnvVar: "PLUGIN_ISSUE_NUM,DRONE_PULL_REQUEST",
		},
		cli.StringFlag{
			Name: "key",
			Usage: "key to assign comment",
			EnvVar: "PLUGIN_KEY",
		},
		cli.StringFlag{
			Name:   "message",
			Usage:  "comment message",
			EnvVar: "PLUGIN_MESSAGE",
		},
		cli.StringFlag{
			Name:   "message-file",
			Usage:  "comment message read from file",
			EnvVar: "PLUGIN_MESSAGE_FILE",
		},
		cli.BoolFlag{
			Name: "update",
			Usage: "update an existing comment that matches the key",
			EnvVar: "PLUGIN_UPDATE",
		},

		//
		// drone env
		//

		cli.StringFlag{
			Name:   "repo-name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "repo-owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	logrus.WithFields(logrus.Fields{
		"Revision": revision,
	}).Info("Drone Github Comment Plugin Version")

	// Read message from file
	message := c.String("message")
	if message == "" {
		if _, err := os.Stat(c.String("message-file")); err == nil {
			dat, err := ioutil.ReadFile(c.String("message-file"))

			if err != nil {
				return err
			}

			message = string(dat)
		}
	}

	plugin := p.Plugin{
		BaseURL:   c.String("base-url"),
		Key:       c.String("key"),
		Message:   message,
		IssueNum:  c.Int("issue-num"),
		Password:  c.String("password"),
		RepoName:  c.String("repo-name"),
		RepoOwner: c.String("repo-owner"),
		Token:     c.String("api-key"),
		Update:    c.Bool("update"),
		Username:  c.String("username"),
	}

	return plugin.Exec()
}
