package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/google/go-github/github"
)

type CommandDefinition struct {
	Name string
	Run  string
}

func (c *CommandDefinition) Execute(msg string) (string, error) {
	args := strings.Fields(clean(msg))

	cmd := exec.Command(c.Run, args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return "error", err
	}

	return string(out), nil
}

type CommandMapDefinition struct {
	Prefix   string
	Commands []*CommandDefinition
}

func NewMap(p string, c *github.Client) (*CommandMap, error) {
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	m := &CommandMapDefinition{}

	err = json.Unmarshal(f, m)
	if err != nil {
		return nil, err
	}

	return &CommandMap{Def: m, c: c}, nil
}

type CommandMap struct {
	Def *CommandMapDefinition
	c   *github.Client
}

func clean(s string) string {
	return strings.TrimSpace(s)
}

func (cm *CommandMap) ExecuteIssueCommentEvent(event *github.IssueCommentEvent) (string, error) {
	message := strings.TrimSpace(event.Comment.GetBody())
	if !strings.HasPrefix(message, cm.Def.Prefix) {
		return "", nil
	}

	message = clean(strings.TrimPrefix(message, cm.Def.Prefix))

	for _, cd := range cm.Def.Commands {
		if strings.HasPrefix(message, cd.Name) {
			result, err := cd.Execute(message)
			if err != nil {
				log.Printf("error running command: %q", err)
				return "", err
			}

			org := *event.Repo.Owner.Login
			repo := *event.Repo.Name
			prNum := *event.Issue.Number

			entry := &github.IssueComment{Body: &result}
			ctx := context.Background()

			_, _, err = cm.c.Issues.CreateComment(ctx, org, repo, prNum, entry)
			if err != nil {
				return "", err
			}
			msg := fmt.Sprintf("%s %s %d %q", org, repo, prNum, result)
			fmt.Println(msg)
			return msg, nil
		}
	}

	// Not a valid command, ignoring b/c it is a normal comment.
	return "", nil
}
