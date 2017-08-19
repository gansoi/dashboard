package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/github"
)

var (
	org          = "gansoi"
	repo         = "gansoi"
	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")

	epoch = monday()

	kpis = make(map[string]int)
)

func monday() time.Time {
	t := time.Now()
	for {
		if t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, -1)
		} else {
			return t.Truncate(24 * time.Hour)
		}
	}
}

func init() {
	kpis["pullRequests"] = 0
	kpis["pullRequestsCreated"] = 0
	kpis["pullRequestsClosed"] = 0
	kpis["issues"] = 0
	kpis["issuesCreated"] = 0
	kpis["issuesClosed"] = 0
}

func main() {
	var client *github.Client

	if clientID != "" && clientSecret != "" {
		t := &github.UnauthenticatedRateLimitedTransport{
			// https://github.com/settings/applications
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}

		client = github.NewClient(t.Client())
	} else {
		client = github.NewClient(nil)
	}

	pullRequestOptions := &github.PullRequestListOptions{
		State: "all",
	}
	pullRequests, _, _ := client.PullRequests.List(context.Background(), org, repo, pullRequestOptions)
	for _, pr := range pullRequests {
		add := 0
		if pr.CreatedAt.After(epoch) {
			add = 1
			kpis["pullRequestsCreated"]++
		}

		if pr.ClosedAt != nil && pr.ClosedAt.After(epoch) {
			add = 1
			kpis["pullRequestsClosed"]++
		}

		kpis["pullRequests"] += add
	}

	issuesOptions := &github.IssueListByRepoOptions{
		State: "all",
	}
	issues, _, _ := client.Issues.ListByRepo(context.Background(), org, repo, issuesOptions)
	for _, i := range issues {
		add := 0
		if i.CreatedAt.After(epoch) {
			add = 1
			kpis["issuesCreated"]++
		}
		if i.ClosedAt != nil && i.ClosedAt.After(epoch) {
			add = 1
			kpis["issuesClosed"]++
		}

		kpis["issues"] += add
		kpis["score"] += add
	}

	commitsOptions := &github.CommitsListOptions{
		Since: epoch,
	}

	commits, _, _ := client.Repositories.ListCommits(context.Background(), org, repo, commitsOptions)

	kpis["commits"] = len(commits)
	kpis["score"] += len(commits)

	// pull requests is included in the issue-count. Subtract.
	kpis["issues"] -= kpis["pullRequests"]
	kpis["issuesCreated"] -= kpis["pullRequestsCreated"]
	kpis["issuesClosed"] -= kpis["pullRequestsClosed"]

	kpis["elapsed"] = int(time.Now().Sub(epoch) / time.Second)

	j, _ := json.MarshalIndent(kpis, "", "  ")
	fmt.Printf("%s\n", j)
}
