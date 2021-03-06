package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type githubController struct {
	Client    *github.Client
	GithubCtx context.Context
}

// CreatePullRequest generates a pull request on Github according to the ReleaseEvent
func (app githubController) CreatePullRequest(e releaseEvent) (github.PullRequest, error) {
	input := &github.NewPullRequest{
		Title: github.String(e.ReleaseVersion),
		Base:  github.String(e.BranchBase),
		Head:  github.String(e.BranchHead),
		Body:  github.String(e.ReleaseBody),
	}

	log.Info(fmt.Sprintf("creating %v pull request...", e.RepoName))
	resp, _, err := app.Client.PullRequests.Create(
		app.GithubCtx,
		e.RepoOwner,
		e.RepoName,
		input,
	)

	if err != nil {
		log.Error(fmt.Sprintf("unable to create %v pull request, %v", e.RepoName, err))
		return *resp, err
	}

	return *resp, nil
}

// MergePullRequest merges the pull request created by ghCreatePullRequest
func (app githubController) MergePullRequest(prNumber int, e releaseEvent) (github.PullRequestMergeResult, error) {
	log.Info(fmt.Sprintf("merging pull request %v...", prNumber))
	mergeResult, _, err := app.Client.PullRequests.Merge(
		app.GithubCtx,
		e.RepoOwner,
		e.RepoName,
		prNumber,
		fmt.Sprintf("Merging pull request number %v", prNumber),
		&github.PullRequestOptions{},
	)

	if err != nil {
		log.Error(fmt.Sprintf("unable to merge %v pull request %v, %v", e.RepoName, prNumber, err))
		return *mergeResult, err
	}
	return *mergeResult, nil
}

// CreateRelease creates a release on Github according to the ReleaseEvent
func (app githubController) CreateRelease(e releaseEvent) error {
	input := &github.RepositoryRelease{
		TargetCommitish: github.String(e.BranchBase),
		TagName:         github.String(e.ReleaseVersion),
		Name:            github.String(e.ReleaseVersion),
		Body:            github.String(e.ReleaseBody),
		Prerelease:      github.Bool(false),
	}

	log.Info(fmt.Sprintf("creating %v release version %v...", e.RepoName, e.ReleaseVersion))
	_, _, err := app.Client.Repositories.CreateRelease(
		app.GithubCtx,
		e.RepoOwner,
		e.RepoName,
		input,
	)

	if err != nil {
		log.Error(fmt.Sprintf("unable to create %v release version %v, %v", e.RepoName, e.ReleaseVersion, err))
		return err
	}
	return nil
}

func (app application) releasesGithubHandler(e releaseEvent, token string) (string, int) {
	githubCtx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(githubCtx, ts)

	app.GH = githubController{
		Client:    github.NewClient(tc),
		GithubCtx: githubCtx,
	}

	var err error
	if !e.Hotfix {
		prResp, err := app.GH.CreatePullRequest(e)
		if err != nil {
			message := fmt.Sprintf("Could not create Github pull request for %v version %v, please check github for further details.",
				e.RepoName,
				e.ReleaseVersion)
			statusCode := 400
			return message, statusCode
		}

		mergeResp, err := app.GH.MergePullRequest(*prResp.Number, e)
		if err != nil {
			message := fmt.Sprintf("API request to merge github pull request %v for %v version %v failed, please check the pull request on github for further details.",
				*prResp.Number,
				e.RepoName,
				e.ReleaseVersion)
			statusCode := 400
			return message, statusCode
		}

		if !*mergeResp.Merged {
			log.Error(fmt.Sprintf("%v pull request %v not merged", e.RepoName, *prResp.Number))
			message := fmt.Sprintf("API request to merge github pull request %v for %v version %v failed, please check the pull request on github for further details.",
				*prResp.Number,
				e.RepoName,
				e.ReleaseVersion)
			statusCode := 400
			return message, statusCode
		}
	}

	err = app.GH.CreateRelease(e)
	if err != nil {
		message := fmt.Sprintf("Unable to create %v release version %v on Github.",
			e.RepoName,
			e.ReleaseVersion)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Created %v release version %v on Github.",
		e.RepoName,
		e.ReleaseVersion)
	statusCode := 200
	return message, statusCode
}
