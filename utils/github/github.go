package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// WorkflowRun struct
type WorkflowRun struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	HeadBranch string    `json:"head_branch"`
	Status     string    `json:"status"`
	CancelURL  string    `json:"cancel_url"`
}

// WorkflowRunAPIResponse struct
type WorkflowRunAPIResponse struct {
	TotalCount   int64         `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

// IGithubAPI interface
type IGithubAPI interface {
	ListWorkflows() ([]WorkflowRun, error)
	CancelRun(run WorkflowRun) error
}

// API struct
type API struct {
	Organization string
	Repository   string
	Token        string
}

const listRunsEndpointFormat = "https://api.github.com/repos/%s/%s/actions/runs"

// MakeAPI creates the api
func MakeAPI() *API {
	return &API{
		Organization: os.Getenv("GITHUB_ORG"),
		Repository:   os.Getenv("GITHUB_REPO"),
		Token:        os.Getenv("GITHUB_TOKEN"),
	}
}

// CancelRun cancels a running workflow
func (api *API) CancelRun(run WorkflowRun) error {
	client := &http.Client{}
	req, err := http.NewRequest("POST", run.CancelURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "token "+api.Token)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("Bad status code: %d \nBody: %s", res.StatusCode, body)
	}

	return nil
}

// ListWorkflows returns list of workflows
func (api *API) ListWorkflows() ([]WorkflowRun, error) {
	client := &http.Client{}
	endpoint := fmt.Sprintf(listRunsEndpointFormat, api.Organization, api.Repository)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token "+api.Token)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	workflowRunRes, err := parseWorkflowsFrom([]byte(body))
	if err != nil {
		return nil, err
	}

	return workflowRunRes.WorkflowRuns, nil
}

func parseWorkflowsFrom(body []byte) (WorkflowRunAPIResponse, error) {
	res := WorkflowRunAPIResponse{}
	err := json.Unmarshal(body, &res)
	return res, err
}
