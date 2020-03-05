package main

import (
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/urbpeti/actions-automatic-cancel/lib"
	"github.com/urbpeti/actions-automatic-cancel/utils"
)

// AutomaticCancel struct
type AutomaticCancel struct {
	GithubAPI     lib.IGithubAPI
	WebHookSecret string
}

func sortRunsByCreatedAtDesc(runs []lib.WorkflowRun) {
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].CreatedAt.After(runs[j].CreatedAt)
	})
}

// AutomaticCancel function
func (canceler *AutomaticCancel) AutomaticCancel(runs []lib.WorkflowRun) error {
	sortRunsByCreatedAtDesc(runs)

	seenBranch := make(map[string]bool)
	for _, run := range runs {
		if run.Status == "completed" {
			continue
		}

		branch := run.HeadBranch

		if _, ok := seenBranch[branch]; ok {
			err := canceler.GithubAPI.CancelRun(run)
			if err != nil {
				log.Println(err.Error())
				continue
			}
		} else {
			seenBranch[branch] = true
		}
	}

	return nil
}

// HandleRequest cancels running workflows
func (canceler *AutomaticCancel) HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := utils.VerifyGithubWebhookRequest(req, canceler.WebHookSecret)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest, Body: err.Error()}, nil
	}

	workflows, err := canceler.GithubAPI.ListWorkflows()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	err = canceler.AutomaticCancel(workflows)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func main() {
	canceler := AutomaticCancel{
		GithubAPI:     lib.MakeGithubAPI(),
		WebHookSecret: os.Getenv("WEBHOOK_SECRET"),
	}
	lambda.Start(canceler.HandleRequest)
}
