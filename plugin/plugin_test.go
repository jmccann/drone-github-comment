package plugin

import (
	"fmt"
	"testing"

	"github.com/franela/goblin"
	"gopkg.in/h2non/gock.v1"
)

func TestPlugin(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("add comment", func() {
		p := Plugin{
			BaseURL:   "http://server.com",
			Message:   "test message",
			IssueNum:  12,
			RepoName:  "test-repo",
			RepoOwner: "test-org",
			Token:     "fake",
		}
		err := p.InitGitClient()
		if err != nil {
			t.Fail()
		}

		g.It("creates a new comment", func() {
			gock.New("http://server.com").
			Post("/repos/test-org/test-repo/issues/12/comments").
			Reply(201).
			JSON(map[string]string{})

			err := p.Exec()
			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
		})
	})

	g.Describe("updates existing comment", func() {
		p := Plugin{
			BaseURL:        "http://server.com",
			Message:        "test message",
			IssueNum:       12,
			Key:            "123",
			RepoName:       "test-repo",
			RepoOwner:      "test-org",
			Update:         true,
			Token:          "fake",
		}
		err := p.InitGitClient()
		if err != nil {
			t.Fail()
		}

		g.It("creates a new comment if one does not exist", func() {
			gock.New("http://server.com").
			Get("repos/test-org/test-repo/issues/12/comments").
			Reply(200).
			File("../testdata/response/non-existing-comment.json")

			gock.New("http://server.com").
			Post("repos/test-org/test-repo/issues/12/comments").
			Reply(201).
			JSON(map[string]string{})

			err := p.Exec()

			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
		})

		g.It("updates the correct comment", func() {
			gock.New("http://server.com").
			Get("repos/test-org/test-repo/issues/12/comments").
			Reply(200).
			File("../testdata/response/existing-comment.json")

			gock.New("http://server.com").
			Patch("repos/test-org/test-repo/issues/comments/7").
			Reply(200).
			JSON(map[string]string{})

			err := p.Exec()

			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
			g.Assert(gock.IsDone()).IsTrue()
		})

		g.It("does not create a new comment if one exists", func() {
			gock.New("http://server.com").
			Get("repos/test-org/test-repo/issues/12/comments").
			Reply(200).
			File("../testdata/response/existing-comment.json")

			// We do not expect this endpoint to get called
			gock.New("http://server.com").
			Post("repos/test-org/test-repo/issues/12/comments").
			Reply(201).
			JSON(map[string]string{})

			gock.New("http://server.com").
			Patch("repos/test-org/test-repo/issues/comments/7").
			Reply(200).
			JSON(map[string]string{})

			err := p.Exec()

			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))

			// Make sure we didn't process all API calls mocked
			g.Assert(gock.IsDone()).IsFalse()
		})

		g.It("generate comment key", func() {
			id := defaultKey(p)

			g.Assert(id).Equal("e056c5655126a83191821948eef7db35762dd9bde43441524aacf3fbfba0ef17")
		})

		g.It("injects comment key", func() {
			gock.New("http://server.com").
			Get("repos/test-org/test-repo/issues/12/comments").
			Reply(200).
			File("../testdata/response/existing-comment.json")

			gock.New("http://server.com").
			Patch("repos/test-org/test-repo/issues/comments/7").
			MatchType("json").
			// Make sure we are sending expected generated message
			File("../testdata/request/patch-comment.json").
			Reply(201).
			JSON(map[string]string{})

			err := p.Exec()

			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
		})
	})
}
