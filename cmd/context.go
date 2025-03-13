package cmd

import (
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/kmtym1998/gh-prowl/entity"
	"github.com/kmtym1998/gh-prowl/notify"
	"github.com/spf13/pflag"
)

type ExecutionContext struct {
	Version         string
	Repo            repository.Repository
	CurrentBranch   string
	PollingInterval time.Duration
	SoundNotifier   entity.Notifier
}

func NewExecutionContext(soundFile io.ReadCloser) (*ExecutionContext, error) {
	soundNotifier, err := notify.NewNotifier(soundFile)
	if err != nil {
		return nil, err
	}

	branchName, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return nil, err
	}

	return &ExecutionContext{
		Version:         "v1.1.1",
		CurrentBranch:   strings.TrimSpace(string(branchName)),
		PollingInterval: 5 * time.Second,
		SoundNotifier:   soundNotifier,
	}, nil
}

func (ec *ExecutionContext) SetRepository(flagSet *pflag.FlagSet) error {
	repo, err := flagSet.GetString("repo")
	if err != nil {
		return err
	}

	if repo != "" {
		ec.Repo, err = repository.Parse(repo)
		if err != nil {
			return err
		}
	} else {
		ec.Repo, err = repository.Current()
		if err != nil {
			return err
		}
	}

	return nil
}
