package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var ErrInvalidHookFormat = errors.New("Unable to parse event string. Invalid Format.")

type CmdMap interface {
	ExecuteIssueCommentEvent(string) (string, error)
}

type Server struct {
	Addr       string // Addr to listen on. Defaults to ":8888"
	Path       string // Path to receive on. Defaults to "/postreceive"
	Secret     string // Option secret key for authenticating via HMAC
	IgnoreTags bool   // If set to false, also execute command if tag is pushed
	Client     *github.Client
	Cmds       CmdMap
}

// Spin up the server and listen for github webhook push events
func (s *Server) ListenAndServe() error {
	r := mux.NewRouter()
	r.HandleFunc("/postreceive", s.GHEventHandler)

	// add any handlers
	loggedHandlers := handlers.LoggingHandler(os.Stdout, r)

	srv := &http.Server{
		Handler:      loggedHandlers,
		Addr:         s.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println(fmt.Sprintf("Listening on %s; Path for event is %s", s.Addr, "/postreceive"))

	return srv.ListenAndServe()
}

// Checks if the given ref should be ignored
func (s *Server) ignoreRef(rawRef string) bool {
	if rawRef[:10] == "refs/tags/" && !s.IgnoreTags {
		return false
	}
	return rawRef[:11] != "refs/heads/"
}

// Satisfies the http.Handler interface.
func (s *Server) GHEventHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "POST" {
		http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventType := req.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "400 Bad Request - Missing X-GitHub-Event Header", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If we have a Secret set, we should check the MAC
	if s.Secret != "" {
		sig := req.Header.Get("X-Hub-Signature")

		if sig == "" {
			forbiddenErr := "403 Forbidden - Missing X-Hub-Signature required for HMAC verification"
			http.Error(w, forbiddenErr, http.StatusForbidden)
			return
		}

		mac := hmac.New(sha1.New, []byte(s.Secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSig := "sha1=" + hex.EncodeToString(expectedMAC)
		if !hmac.Equal([]byte(expectedSig), []byte(sig)) {
			http.Error(w, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
			return
		}
	}

	event, err := github.ParseWebHook(github.WebHookType(req), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		s.processIssueCommentEvent(event, w, req)
		return
	default:
		fmt.Println(fmt.Sprintf("unknown event type: %s %#v", eventType, event))
	}
}

func (s Server) processIssueCommentEvent(event *github.IssueCommentEvent, w http.ResponseWriter, req *http.Request) {
	if event.GetAction() != "created" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Ignoring comment since action was not created")
		return
	}

	message := strings.TrimSpace(event.Comment.GetBody())
	result, commandErr := s.Cmds.ExecuteIssueCommentEvent(message)

	result, commentErr := s.postComment(event, result)
	if commentErr != nil {
		http.Error(w, commentErr.Error(), http.StatusInternalServerError)
		return
	}

	if commandErr != nil {
		http.Error(w, commandErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, result)
}

func (s Server) postComment(event *github.IssueCommentEvent, msg string) (string, error) {
	org := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	prNum := *event.Issue.Number

	entry := &github.IssueComment{Body: &msg}
	ctx := context.Background()

	_, _, err := s.Client.Issues.CreateComment(ctx, org, repo, prNum, entry)
	if err != nil {
		return "", err
	}
	output := fmt.Sprintf("%s %s %d %q", org, repo, prNum, msg)
	fmt.Printf(output)
	return output, nil
}
