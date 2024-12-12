package cmd

import (
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/kmtym1998/gh-prowl/api"
	"github.com/kmtym1998/gh-prowl/entity"
	"github.com/kmtym1998/gh-prowl/notify"
)

type ExecutionContext struct {
	RepoOwner       string
	RepoName        string
	CurrentBranch   string
	PollingInterval time.Duration
	ApiClient       entity.GitHubAPIClient
	SoundNotifier   entity.Notifier
}

func NewExecutionContext(soundFile io.ReadCloser) (*ExecutionContext, error) {
	soundNotifier, err := notify.NewNotifier(soundFile)
	if err != nil {
		return nil, err
	}

	client, err := api.NewAPIClient()
	if err != nil {
		return nil, err
	}

	repo, err := repository.Current()
	if err != nil {
		return nil, err
	}

	branchName, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return nil, err
	}

	return &ExecutionContext{
		RepoOwner:       repo.Owner,
		RepoName:        repo.Name,
		CurrentBranch:   strings.TrimSpace(string(branchName)),
		PollingInterval: 5 * time.Second,
		ApiClient:       client,
		SoundNotifier:   soundNotifier,
	}, nil
}
