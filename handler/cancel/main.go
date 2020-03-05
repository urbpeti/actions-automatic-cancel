package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/urbpeti/actions-automatic-cancel/utils/github"
)

// AutomaticCancel struct
type AutomaticCancel struct {
	GithubAPI     github.IGithubAPI
	WebHookSecret string
}

func sortRunsByCreatedAtDesc(runs []github.WorkflowRun) {
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].CreatedAt.After(runs[j].CreatedAt)
	})
}

// AutomaticCancel function
func (canceler *AutomaticCancel) AutomaticCancel(runs []github.WorkflowRun) error {
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
	err := verifyRequest(req, canceler.WebHookSecret)
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

func verifyRequest(req events.APIGatewayProxyRequest, secret string) error {
	signature, err := hex.DecodeString(strings.Split(req.Headers["X-Hub-Signature"], "=")[1])
	if err != nil {
		return err
	}
	if !verifyPayload(secret, []byte(req.Body), signature) {
		return fmt.Errorf("Signature missmatch")
	}

	return nil
}

func verifyPayload(secret string, payload, signature []byte) bool {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(signature, expectedMAC)
}

func main() {
	canceler := AutomaticCancel{
		GithubAPI:     github.MakeAPI(),
		WebHookSecret: os.Getenv("WEBHOOK_SECRET"),
	}
	lambda.Start(canceler.HandleRequest)
}
