package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"
)

func TestListWorkflows(t *testing.T) {
	githubAPI := GithubAPI{
		Organization: "org",
		Repository:   "repo",
		Token:        "dummytoken",
	}

	t.Run("List workflows return error on server error", func(t *testing.T) {
		gock.New("https://api.github.com").
			Get("/repos/org/repo/actions/runs").
			MatchHeader("Authorization", "token dummytoken").
			ReplyError(fmt.Errorf("Server error"))

		_, err := githubAPI.ListWorkflows()
		if err == nil {
			t.Errorf("Missing error")
		}
		if !strings.Contains(err.Error(), "Server error") {
			t.Errorf("Bad error: %s", err.Error())
		}

		if !gock.IsDone() {
			t.Errorf("Endpoinds was not called")
		}
	})

	t.Run("List and parse workflows", func(t *testing.T) {
		expectedRuns := []WorkflowRun{
			WorkflowRun{
				ID:         1,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 0, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel.url",
			},
			WorkflowRun{
				ID:         2,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 1, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel.url",
			},
		}
		apiReply, err := json.Marshal(WorkflowRunAPIResponse{
			TotalCount:   2,
			WorkflowRuns: expectedRuns,
		})

		if err != nil {
			t.Errorf("Json marshal failed with %s", err.Error())
		}

		gock.New("https://api.github.com").
			Get("/repos/org/repo/actions/runs").
			MatchHeader("Authorization", "token dummytoken").
			Reply(200).
			JSON(apiReply)

		runs, err := githubAPI.ListWorkflows()
		if err != nil {
			t.Errorf(err.Error())
		}

		for i, run := range runs {
			if run != expectedRuns[i] {
				t.Errorf("Expected run: %d, Actual run: %d", run.ID, expectedRuns[i].ID)
			}
		}

		if !gock.IsDone() {
			t.Errorf("Endpoinds was not called")
		}
	})
}

func TestCancelRun(t *testing.T) {
	githubAPI := GithubAPI{
		Organization: "org",
		Repository:   "repo",
		Token:        "dummytoken",
	}

	t.Run("Cancel Run error", func(t *testing.T) {
		defer gock.Off()
		gock.New("https://api.github.com").
			Post("/org/repo/cancel").
			MatchHeader("Authorization", "token dummytoken").
			ReplyError(fmt.Errorf("Server error"))

		err := githubAPI.CancelRun(WorkflowRun{
			CancelURL: "https://api.github.com/org/repo/cancel",
		})

		if err == nil {
			t.Errorf("Missing error")
		}
		if !strings.Contains(err.Error(), "Server error") {
			t.Errorf("Bad error: %s", err.Error())
		}

		if !gock.IsDone() {
			t.Errorf("Endpoinds was not called")
		}
	})

	t.Run("Cancel not return 200", func(t *testing.T) {
		defer gock.Off()
		gock.New("https://api.github.com").
			Post("/org/repo/cancel").
			MatchHeader("Authorization", "token dummytoken").
			Reply(http.StatusInternalServerError)

		err := githubAPI.CancelRun(WorkflowRun{
			CancelURL: "https://api.github.com/org/repo/cancel",
		})

		if err == nil {
			t.Errorf("Missing error")
		}
		if !strings.Contains(err.Error(), "Bad status code: 500") {
			t.Errorf("Bad error: %s", err.Error())
		}

		if !gock.IsDone() {
			t.Errorf("Endpoinds was not called")
		}
	})

	t.Run("Cancels run", func(t *testing.T) {
		defer gock.Off()
		gock.New("https://api.github.com").
			Post("/org/repo/cancel").
			MatchHeader("Authorization", "token dummytoken").
			Reply(http.StatusAccepted)

		err := githubAPI.CancelRun(WorkflowRun{
			CancelURL: "https://api.github.com/org/repo/cancel",
		})

		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}

		if !gock.IsDone() {
			t.Errorf("Endpoinds was not called")
		}
	})
}
