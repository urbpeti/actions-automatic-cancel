package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/urbpeti/actions-automatic-cancel/lib"
)

type MockGithubAPI struct {
	MockListWorkflows func() ([]lib.WorkflowRun, error)
	MockCancelRun     func(lib.WorkflowRun) error
}

func (api *MockGithubAPI) ListWorkflows() ([]lib.WorkflowRun, error) {
	return api.MockListWorkflows()
}
func (api *MockGithubAPI) CancelRun(run lib.WorkflowRun) error {
	return api.MockCancelRun(run)
}

func TestHandleRequest(t *testing.T) {
	canceler := AutomaticCancel{
		GithubAPI:     &MockGithubAPI{},
		WebHookSecret: "secret",
	}

	t.Run("Bad signature", func(t *testing.T) {
		reqBody, err := json.Marshal(&lib.WorkflowRunAPIResponse{})
		res, err := canceler.HandleRequest(events.APIGatewayProxyRequest{
			Body:    string(reqBody),
			Headers: map[string]string{"X-Hub-Signature": "sha1=829c3804401b0727f70f73d4415e162400cbe57b"},
		})

		if err != nil {
			t.Errorf(err.Error())
		}
		if res.Body != "Signature missmatch" {
			t.Errorf("Bad body %s", res.Body)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status: %d, actual: %d", http.StatusBadRequest, res.StatusCode)
		}
	})

	t.Run("Bad signature decode error", func(t *testing.T) {
		reqBody, err := json.Marshal(&lib.WorkflowRunAPIResponse{})
		res, err := canceler.HandleRequest(events.APIGatewayProxyRequest{
			Body:    string(reqBody),
			Headers: map[string]string{"X-Hub-Signature": "sha1=fff"},
		})

		if err != nil {
			t.Errorf(err.Error())
		}
		if res.Body != "encoding/hex: odd length hex string" {
			t.Errorf("Bad body %s", res.Body)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status: %d, actual: %d", http.StatusBadRequest, res.StatusCode)
		}
	})

	t.Run("List workflows err should return internal server error", func(t *testing.T) {
		canceler.GithubAPI = &MockGithubAPI{MockListWorkflows: func() ([]lib.WorkflowRun, error) { return []lib.WorkflowRun{}, fmt.Errorf("Dummy Error") }}
		reqBody, err := json.Marshal(&lib.WorkflowRunAPIResponse{})
		_, err = canceler.HandleRequest(events.APIGatewayProxyRequest{
			Body:    string(reqBody),
			Headers: map[string]string{"X-Hub-Signature": "sha1=80f5bd58cfb34a5316382d28977159e854f3aa9d"},
		})

		if err == nil || err.Error() != "Dummy Error" {
			t.Errorf("Bad error")
		}
	})

	t.Run("Cancel run err should skip cancel and return 200", func(t *testing.T) {
		canceler.GithubAPI = &MockGithubAPI{
			MockListWorkflows: func() ([]lib.WorkflowRun, error) {
				return []lib.WorkflowRun{
					lib.WorkflowRun{
						ID:         1,
						CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 0, time.UTC),
						HeadBranch: "master",
						Status:     "running",
						CancelURL:  "cancel.url",
					},
					lib.WorkflowRun{
						ID:         2,
						CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 1, time.UTC),
						HeadBranch: "master",
						Status:     "running",
						CancelURL:  "cancel.url",
					},
				}, nil
			},
			MockCancelRun: func(lib.WorkflowRun) error { return fmt.Errorf("Dummy Error") }}
		reqBody, err := json.Marshal(&lib.WorkflowRunAPIResponse{})
		res, err := canceler.HandleRequest(events.APIGatewayProxyRequest{
			Body:    string(reqBody),
			Headers: map[string]string{"X-Hub-Signature": "sha1=80f5bd58cfb34a5316382d28977159e854f3aa9d"},
		})

		if err != nil {
			t.Errorf(err.Error())
		}
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status: %d, actual: %d", http.StatusOK, res.StatusCode)
		}
	})

	t.Run("Cancel run err should return internal server error", func(t *testing.T) {
		canceler.GithubAPI = &MockGithubAPI{
			MockListWorkflows: func() ([]lib.WorkflowRun, error) { return []lib.WorkflowRun{}, nil },
			MockCancelRun:     func(lib.WorkflowRun) error { return nil },
		}

		reqBody, err := json.Marshal(&lib.WorkflowRunAPIResponse{})
		res, err := canceler.HandleRequest(events.APIGatewayProxyRequest{
			Body:    string(reqBody),
			Headers: map[string]string{"X-Hub-Signature": "sha1=80f5bd58cfb34a5316382d28977159e854f3aa9d"},
		})

		if err != nil {
			t.Errorf("Bad error: %s", err.Error())
		}
		expectedStatus := http.StatusOK
		if res.StatusCode != expectedStatus {
			t.Errorf("Expected status: %d, actual: %d", expectedStatus, res.StatusCode)
		}
	})
}

func TestAutomaticCancel(t *testing.T) {
	canceler := AutomaticCancel{
		GithubAPI:     &MockGithubAPI{},
		WebHookSecret: "secret",
	}

	t.Run("Should not cancel completed runs", func(t *testing.T) {
		cancelCount := 0
		var cancelCalls []lib.WorkflowRun
		canceler.GithubAPI = &MockGithubAPI{MockCancelRun: func(run lib.WorkflowRun) error {
			cancelCount++
			cancelCalls = append(cancelCalls, run)
			return nil
		}}
		canceler.AutomaticCancel([]lib.WorkflowRun{
			lib.WorkflowRun{
				ID:         1,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 0, time.UTC),
				HeadBranch: "master",
				Status:     "completed",
				CancelURL:  "cancel.url",
			},
			lib.WorkflowRun{
				ID:         2,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 1, time.UTC),
				HeadBranch: "master",
				Status:     "completed",
				CancelURL:  "cancel.url",
			},
		})

		expectedCancelCount := 0
		if cancelCount != expectedCancelCount {
			t.Errorf("Excepted cancel count %d actual %d", expectedCancelCount, cancelCount)
		}
	})

	t.Run("Should not cancel different branches", func(t *testing.T) {
		cancelCount := 0
		var cancelCalls []lib.WorkflowRun
		canceler.GithubAPI = &MockGithubAPI{MockCancelRun: func(run lib.WorkflowRun) error {
			cancelCount++
			cancelCalls = append(cancelCalls, run)
			return nil
		}}
		canceler.AutomaticCancel([]lib.WorkflowRun{
			lib.WorkflowRun{
				ID:         1,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 0, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel.url",
			},
			lib.WorkflowRun{
				ID:         2,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 1, time.UTC),
				HeadBranch: "featurebrach",
				Status:     "running",
				CancelURL:  "cancel.url",
			},
		})

		expectedCancelCount := 0
		if cancelCount != expectedCancelCount {
			t.Errorf("Excepted cancel count %d actual %d", expectedCancelCount, cancelCount)
		}
	})

	t.Run("Should cancel older runs", func(t *testing.T) {
		cancelCount := 0
		var cancelCalls []lib.WorkflowRun
		canceler.GithubAPI = &MockGithubAPI{MockCancelRun: func(run lib.WorkflowRun) error {
			cancelCount++
			cancelCalls = append(cancelCalls, run)
			return nil
		}}
		canceler.AutomaticCancel([]lib.WorkflowRun{
			lib.WorkflowRun{
				ID:         1,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 2, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel1",
			},
			lib.WorkflowRun{
				ID:         2,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 3, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel2",
			},
			lib.WorkflowRun{
				ID:         3,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 1, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel3",
			},
		})

		expectedCancelCount := 2
		if cancelCount != expectedCancelCount {
			t.Errorf("Excepted cancel count %d actual %d", expectedCancelCount, cancelCount)
		}
		var expectedCancelID int64 = 1
		if cancelCalls[0].ID != expectedCancelID {
			t.Errorf("Expected cancel ID: %d actual: %d", expectedCancelID, cancelCalls[0].ID)
		}
		expectedCancelID = 3
		if cancelCalls[1].ID != expectedCancelID {
			t.Errorf("Expected cancel ID: %d actual: %d", expectedCancelID, cancelCalls[1].ID)
		}
	})

	t.Run("Should cancel older runs on multiple branch", func(t *testing.T) {
		cancelCount := 0
		var cancelCalls []lib.WorkflowRun
		canceler.GithubAPI = &MockGithubAPI{MockCancelRun: func(run lib.WorkflowRun) error {
			cancelCount++
			cancelCalls = append(cancelCalls, run)
			return nil
		}}
		canceler.AutomaticCancel([]lib.WorkflowRun{
			lib.WorkflowRun{
				ID:         1,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 1, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel1",
			},
			lib.WorkflowRun{
				ID:         2,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 2, time.UTC),
				HeadBranch: "master",
				Status:     "running",
				CancelURL:  "cancel2",
			},
			lib.WorkflowRun{
				ID:         3,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 4, time.UTC),
				HeadBranch: "featureBaranch",
				Status:     "running",
				CancelURL:  "cancel3",
			},
			lib.WorkflowRun{
				ID:         4,
				CreatedAt:  time.Date(2020, 02, 29, 0, 0, 0, 3, time.UTC),
				HeadBranch: "featureBaranch",
				Status:     "running",
				CancelURL:  "cancel4",
			},
		})

		expectedCancelCount := 2
		if cancelCount != expectedCancelCount {
			t.Errorf("Excepted cancel count %d actual %d", expectedCancelCount, cancelCount)
		}
		var expectedCancelID int64 = 4
		if cancelCalls[0].ID != expectedCancelID {
			t.Errorf("Expected cancel ID: %d actual: %d", expectedCancelID, cancelCalls[0].ID)
		}
		expectedCancelID = 1
		if cancelCalls[1].ID != expectedCancelID {
			t.Errorf("Expected cancel ID: %d actual: %d", expectedCancelID, cancelCalls[1].ID)
		}
	})
}
