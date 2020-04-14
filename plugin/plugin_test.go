package plugin

import (
	"fmt"
	"testing"

	"github.com/franela/goblin"
	"gopkg.in/h2non/gock.v1"
)

func TestPlugin(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("NewFromPlugin", func() {
		pl := Plugin{
			BaseURL:   "http://server.com",
			Message:   "test message",
			IssueNum:  12,
			RepoName:  "test-repo",
			RepoOwner: "test-org",
			Token:     "fake",
		}

		g.It("creates a new/initialized Plugin from a Plugin", func() {
			_, err := NewFromPlugin(pl)
			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
		})

		g.It("errors when defined directly", func() {
			defer gock.Off()

			gock.New("http://server.com").
				Post("/repos/test-org/test-repo/issues/12/comments").
				Reply(201).
				JSON(map[string]string{})

			defer func() {
				r := recover()
				if r != nil {
					g.Fail("The code should not panic")
				}
			}()

			err := pl.Exec()
			g.Assert(err != nil).IsTrue("should have received error that git client not initialized")
		})
	})

	g.Describe("add comment", func() {
		pl := Plugin{
			BaseURL:   "http://server.com",
			Message:   "test message",
			IssueNum:  12,
			RepoName:  "test-repo",
			RepoOwner: "test-org",
			Token:     "fake",
		}
		p, err := NewFromPlugin(pl)
		if err != nil {
			g.Fail("Failed to create plugin for testing")
		}

		g.It("creates a new comment", func() {
			defer gock.Off()

			gock.New("http://server.com").
				Post("/repos/test-org/test-repo/issues/12/comments").
				Reply(201).
				JSON(map[string]string{})

			err := p.Exec()
			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
		})
	})

	g.Describe("updates existing comment", func() {
		pl := Plugin{
			BaseURL:   "http://server.com",
			Message:   "test message",
			IssueNum:  12,
			Key:       "123",
			RepoName:  "test-repo",
			RepoOwner: "test-org",
			Update:    true,
			Token:     "fake",
		}
		p, err := NewFromPlugin(pl)
		if err != nil {
			g.Fail("Failed to create plugin for testing")
		}

		g.It("creates a new comment if one does not exist", func() {
			defer gock.Off()

			// Get Comments
			gock.New("http://server.com").
				Get("repos/test-org/test-repo/issues/12/comments").
				Reply(200).
				File("../testdata/response/non-existing-comment.json")

			// Create new comment
			gock.New("http://server.com").
				Post("repos/test-org/test-repo/issues/12/comments").
				Reply(201).
				JSON(map[string]string{})

			err := p.Exec()

			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
			g.Assert(gock.HasUnmatchedRequest()).IsFalse(fmt.Sprintf("Received unmatched requests: %v\n", gock.GetUnmatchedRequests()))

			if !gock.IsDone() {
				for _, m := range gock.Pending() {
					g.Fail(fmt.Sprintf("Did not make expected request: %s(%s)", m.Request().Method, m.Request().URLStruct))
				}
			}
		})

		g.It("updates the correct comment", func() {
			defer gock.Off()

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
			g.Assert(gock.HasUnmatchedRequest()).IsFalse(fmt.Sprintf("Received unmatched requests: %v\n", gock.GetUnmatchedRequests()))

			if !gock.IsDone() {
				for _, m := range gock.Pending() {
					g.Fail(fmt.Sprintf("Did not make expected request: %s(%s)", m.Request().Method, m.Request().URLStruct))
				}
			}
		})

		g.It("does not create a new comment if one exists", func() {
			defer gock.Off()

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

			// Make sure we DID NOT process all API calls mocked
			g.Assert(gock.IsDone()).IsFalse()
		})

		g.It("generate comment key", func() {
			id := defaultKey(*p)

			g.Assert(id).Equal("e056c5655126a83191821948eef7db35762dd9bde43441524aacf3fbfba0ef17")
		})

		g.It("injects comment key", func() {
			defer gock.Off()

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

	g.Describe("add comment with no issue num", func() {
		pl := Plugin{
			BaseURL:   "http://server.com",
			Message:   "test message",
			CommitSha: "deadbeef12",
			RepoName:  "test-repo",
			RepoOwner: "test-org",
			Token:     "fake",
		}
		p, err := NewFromPlugin(pl)
		if err != nil {
			g.Fail("Failed to create plugin for testing")
		}

		g.It("creates a new comment", func() {
			defer gock.Off()

			gock.New("http://server.com").
				Get("/search/issues").
				MatchParam("q", "deadbeef12 repo:test-org/test-repo").
				Reply(200).
				File("../testdata/response/search-issues.json")

			gock.New("http://server.com").
				Post("/repos/test-org/test-repo/issues/12/comments").
				Reply(201).
				JSON(map[string]string{})

			err := p.Exec()
			g.Assert(err == nil).IsTrue(fmt.Sprintf("Received err: %s", err))
		})
	})
}
