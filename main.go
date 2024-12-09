package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/v2/pkg/prompter"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/fatih/color"
	"github.com/kmtym1998/gh-prowl/api"
	"github.com/kmtym1998/gh-prowl/entity"
)

type ghAPIClient interface {
	ListPullRequests(ctx context.Context, repoOwner, repoName string, limit int) (*entity.SimplePRList, error)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			color.Red(fmt.Sprint(r))
		}
	}()

	client, err := api.NewAPIClient()
	if err != nil {
		panic(err)
	}

	if err := run(client); err != nil {
		color.Red(err.Error())
		panic("failed to execute command")
	}
}

func run(client ghAPIClient) error {
	ctx := context.Background()
	repo, err := repository.Current()
	if err != nil {
		return fmt.Errorf("failed to get current repository: %w", err)
	}

	prList, err := client.ListPullRequests(ctx, repo.Owner, repo.Name, 10)
	if err != nil {
		return fmt.Errorf("failed to list pull requests: %w", err)
	}

	io := iostreams.System()
	fmt.Printf("Total PRs: %d\n", prList.Total)
	p := prompter.New(io.In, io.Out, io.ErrOut)
	selected, err := p.Select(
		"Select a PR to prowl",
		"",
		func() (result []string) {
			for _, pr := range prList.Items {
				result = append(result, fmt.Sprintf("#%d %s", pr.Number, pr.Title))
			}
			return
		}(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "interrupt") {
			return nil
		}
		return fmt.Errorf("failed to prompt user: %w", err)
	}

	fmt.Printf("Selected PR: %s\n", prList.Items[selected].Title)

	return nil
}
