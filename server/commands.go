package server

import (
	"fmt"
	"strings"
)

type CommandMap struct{}

func clean(s string) string {
	return strings.TrimSpace(s)
}

func (cm *CommandMap) Execute(comment string) {
	comment = clean(comment)
	if strings.HasPrefix("sink: ", comment) {
		fmt.Printf("thanks for commenting!")
	}
}
