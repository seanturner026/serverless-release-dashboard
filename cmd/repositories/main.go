package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/google/go-github/github"
	util "github.com/seanturner026/moot/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type application struct {
	AWS    awsController
	GH     githubController
	GL     gitlabController
	Config configuration
}

type awsController struct {
	TableName string
	DB        dynamodbiface.DynamoDBAPI
	SSM       ssmiface.SSMAPI
}

type githubController struct {
	Client    *github.Client
	GithubCtx context.Context
}

type gitlabController struct {
	MergeRequestSquash bool
	RemoveSourceBranch bool
	Client             *gitlab.Client
}

type configuration struct {
	DashboardName string
}

type repository struct {
	RepoName        string `json:"repo_name,omitempty"`
	RepoProvider    string `json:"repo_provider,omitempty"   dynamodbav:"SK"`
	RepoOwner       string `json:"repo_owner,omitempty"      dynamodbav:"RepoOwner"`
	BranchBase      string `json:"branch_base,omitempty"     dynamodbav:"BranchBase"`
	BranchHead      string `json:"branch_head,omitempty"     dynamodbav:"BranchHead"`
	CurrentVersion  string `json:"current_version,omitempty" dynamodbav:"CurrentVersion"`
	GitlabProjectID string `json:"gitlab_repo_id,omitempty"  dynamodbav:"GitlabProjectID,omitempty"`
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/repositories/create" {
		log.Info(fmt.Sprintf("handling request on %s", event.RawPath))
		message, statusCode := app.repositoriesCreateHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/repositories/delete" {
		log.Info(fmt.Sprintf("handling request on %s", event.RawPath))
		message, statusCode := app.repositoriesDeleteHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/repositories/list" {
		log.Info(fmt.Sprintf("handling request on %s", event.RawPath))
		message, statusCode := app.repositoriesListHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil
	}

	log.Error(fmt.Sprintf("path %v does not exist", event.RawPath))
	resp := util.GenerateResponseBody(fmt.Sprintf("Path does not exist %s", event.RawPath), 404, nil, headers, []string{})
	return resp, nil
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	app := application{
		AWS: awsController{
			TableName: os.Getenv("TABLE_NAME"),
			DB:        dynamodb.New(session.Must(session.NewSession())),
			SSM:       ssm.New(session.Must(session.NewSession())),
		},
		Config: configuration{
			DashboardName: os.Getenv("DASHBOARD_NAME"),
		},
	}

	lambda.Start(app.handler)
}
