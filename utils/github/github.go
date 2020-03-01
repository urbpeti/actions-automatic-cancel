package github

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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

// MakeAPI creates the api
func MakeAPI(token, organization, repository string) *API {
	return &API{
		Organization: organization,
		Repository:   repository,
		Token:        token,
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
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return err
}

// ListWorkflows returns list of workflows
func (api *API) ListWorkflows() ([]WorkflowRun, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+api.Organization+"/"+api.Repository+"/actions/runs", nil)
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

	workflowRunRes, err := getWorkflowsFrom([]byte(body))
	if err != nil {
		return nil, err
	}

	return workflowRunRes.WorkflowRuns, nil
}

func getWorkflowsFrom(body []byte) (*WorkflowRunAPIResponse, error) {
	var res = new(WorkflowRunAPIResponse)
	err := json.Unmarshal(body, &res)
	return res, err
}
