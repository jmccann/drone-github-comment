package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var revision string // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "github pr comment plugin"
	app.Usage = "github pr comment plugin"
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
			Name:   "message",
			Usage:  "github token",
			EnvVar: "PLUGIN_MESSAGE",
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
	}).Info("Drone Github PR Plugin Version")

	plugin := Plugin{
		BaseURL:    c.String("base-url"),
		Message:    c.String("message"),
		IssueNum:   c.Int("pull-request"),
		RepoName:   c.String("repo-name"),
		RepoOwner:  c.String("repo-owner"),
		Token:      c.String("api-key"),
	}

	return plugin.Exec()
}
