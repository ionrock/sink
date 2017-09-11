package main

import (
	"context"
	"log"

	"github.com/ionrock/sink/command"
	"github.com/ionrock/sink/repo"
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

func main() {
	var (
		webhookSecret = envflag.String("SINK_WEBHOOK_SECRET", "", "The GitHub webhook secret")
		accessToken   = envflag.String("SINK_REPO_ACCESS_TOKEN", "", "An access token with access the PR repo")
		repoRemote    = envflag.String("SINK_REPO_URL", "", "The URL to the repo")
		commandMap    = envflag.String("SINK_COMMAND_MAP", "", "The command map file to use")
	)
	envflag.Parse()

	if *webhookSecret == "" {
		log.Fatal("The SINK_WEBHOOK_SECRET must be set")
	}

	if *accessToken == "" {
		log.Fatal("The SINK_REPO_ACCESS_TOKEN must be set")
	}

	if *repoRemote == "" {
		log.Fatal("The SINK_REPO_URL must be set")
	}

	r := repo.Git{Remote: *repoRemote}
	err := r.Clone()
	if err != nil {
		log.Fatalf("Error cloning repo: %q", err)
	}

	c := newClient(*accessToken)

	cmdMap, err := command.NewMap(*commandMap, c)
	if err != nil {
		log.Fatalf("Error loading command map: %q", err)
	}

	server := &server.Server{
		Cmds:   cmdMap,
		Client: c,
		Addr:   ":8888",
		Path:   "/",
		Secret: *webhookSecret,
	}
	server.ListenAndServe()
}
