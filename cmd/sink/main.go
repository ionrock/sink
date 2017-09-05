package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ionrock/sink/server"

	"github.com/google/go-github/github"
	"github.com/ianschenck/envflag"
	"golang.org/x/oauth2"
)

func newClient(token string) *github.Client {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

type EchoCommandMap struct {
	c *github.Client
}

func clean(s string) string {
	return strings.TrimSpace(s)
}

func (cm *EchoCommandMap) ExecuteIssueCommentEvent(event *github.IssueCommentEvent) (string, error) {
	message := strings.TrimSpace(event.Comment.GetBody())
	if !strings.HasPrefix(message, "sink: ") {
		return "", nil
	}

	org := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	prNum := *event.Issue.Number

	comment := fmt.Sprintf("Hey I heard you say: %q", message)

	entry := &github.IssueComment{Body: &comment}
	ctx := context.Background()

	_, _, err := cm.c.Issues.CreateComment(ctx, org, repo, prNum, entry)
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("%s %s %d %q", org, repo, prNum, comment)
	fmt.Println(msg)
	return msg, nil
}

func main() {
	var (
		webhookSecret = envflag.String("SINK_WEBHOOK_SECRET", "", "The GitHub webhook secret")
		accessToken   = envflag.String("SINK_REPO_ACCESS_TOKEN", "", "An access token with access the PR repo")
	)
	envflag.Parse()

	if *webhookSecret == "" {
		log.Fatal("The SINK_WEBHOOK_SECRET must be set")
	}

	if *accessToken == "" {
		log.Fatal("The SINK_REPO_ACCESS_TOKEN must be set")
	}

	c := newClient(*accessToken)
	server := &server.Server{
		Cmds:   &EchoCommandMap{c: c},
		Client: c,
		Addr:   ":8888",
		Path:   "/",
		Secret: *webhookSecret,
	}
	server.ListenAndServe()
}
