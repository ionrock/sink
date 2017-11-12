package command

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"strings"
)

// CommandDefinition is named command for running commands in a CommandMap.
type CommandDefinition struct {
	Name string
	Run  string
}

// Execute runs the command from the CommandDefinition and returns the
// output of the command along with any errors.
func (c *CommandDefinition) Execute(msg string) (string, error) {
	args := strings.Fields(clean(msg))

	cmd := exec.Command(c.Run, args[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// CommandMapDefinition defines a prefix and list of commands to run
// when the prefix matched in a PR.
type CommandMapDefinition struct {
	Prefix   string
	Commands []*CommandDefinition
}

// NewMap returns a CommandMap based on a JSON file.
func NewMap(p string) (*CommandMap, error) {
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	m := &CommandMapDefinition{}

	err = json.Unmarshal(f, m)
	if err != nil {
		return nil, err
	}

	return &CommandMap{Def: m}, nil
}

// CommandMap takes a prefix and matches it to an action that is
// implemented by a command.
type CommandMap struct {
	Def *CommandMapDefinition
}

func clean(s string) string {
	return strings.TrimSpace(s)
}

// ExecuteIssueCommentEvent handles new comment message and tries to
// match a prefix in the CommandMap.
func (cm *CommandMap) ExecuteIssueCommentEvent(message string) (string, error) {
	if !strings.HasPrefix(message, cm.Def.Prefix) {
		return "", nil
	}

	// TODO: This is probably not necessary. There is no need to
	// enforce a `sink $prefix $args` pattern
	//
	// Trim the globa prefix
	message = clean(strings.TrimPrefix(message, cm.Def.Prefix))
	for _, cd := range cm.Def.Commands {
		if strings.HasPrefix(message, cd.Name) {
			return cd.Execute(message)
		}
	}

	// Not a valid command, ignoring b/c it is a normal comment.
	return "", nil
}
