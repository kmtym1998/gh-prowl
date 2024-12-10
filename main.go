package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/v2/pkg/prompter"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/cli/go-gh/v2/pkg/tableprinter"

	"github.com/fatih/color"
	"github.com/samber/lo"

	"github.com/kmtym1998/gh-prowl/api"
	"github.com/kmtym1998/gh-prowl/entity"
)

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

	repo, err := repository.Current()
	if err != nil {
		panic(err)
	}

	if err := run(client, option{
		repoOwner:       repo.Owner,
		repoName:        repo.Name,
		pollingInterval: 5 * time.Second,
	}); err != nil {
		color.Red(err.Error())
		panic("failed to execute command")
	}
}

type ghAPIClient interface {
	ListPullRequests(ctx context.Context, repoOwner, repoName string, limit int) (*entity.SimplePRList, error)
	GetPRLatestCommitSHA(ctx context.Context, repoOwner, repoName string, prNumber int) (string, error)
	ListCheckRuns(ctx context.Context, repoOwner, repoName string, commitSHA string) (*entity.SimpleCheckRunList, error)
}

type option struct {
	repoOwner       string
	repoName        string
	pollingInterval time.Duration
}

func run(client ghAPIClient, o option) error {
	_ctx := context.Background()
	ctx, cancel := context.WithTimeout(_ctx, 30*time.Minute)
	defer cancel()

	prList, err := client.ListPullRequests(ctx, o.repoOwner, o.repoName, 10)
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

	fmt.Printf("🦉 Selected PR: %s\n", prList.Items[selected].Title)
	fmt.Printf("🦉 View this PR on GitHub: %s\n", prList.Items[selected].URL)

	sha, err := client.GetPRLatestCommitSHA(ctx, o.repoOwner, o.repoName, prList.Items[selected].Number)
	if err != nil {
		return fmt.Errorf("failed to get latest commit SHA: %w", err)
	}

	fmt.Printf("🦉 Watching %s/%s@%s checks\n", o.repoOwner, o.repoName, sha)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		checkRunList, err := client.ListCheckRuns(ctx, o.repoOwner, o.repoName, sha)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				return err
			}

			return fmt.Errorf("failed to list check runs: %w", err)
		}

		if lo.SomeBy(checkRunList.Items, func(checkRun *entity.SimpleCheckRun) bool {
			return checkRun.Status != entity.CheckRunStatusCompleted
		}) {
			time.Sleep(o.pollingInterval)
			continue
		}

		type output struct {
			Name       string
			Status     string
			Conclusion string
		}
		resultOutputs := []output{}
		for _, checkRun := range checkRunList.Items {
			if checkRun.Conclusion == nil {
				resultOutputs = append(resultOutputs, output{
					Name:       checkRun.Name,
					Status:     checkRun.Status.String(),
					Conclusion: "NULL",
				})
				continue
			}

			if !checkRun.Conclusion.IsSuccess() {
				resultOutputs = append(resultOutputs, output{
					Name:       checkRun.Name,
					Status:     checkRun.Status.String(),
					Conclusion: checkRun.Conclusion.String(),
				})
				continue
			}

			resultOutputs = append(resultOutputs, output{
				Name:       checkRun.Name,
				Status:     checkRun.Status.String(),
				Conclusion: checkRun.Conclusion.String(),
			})
		}
		sort.Slice(resultOutputs, func(i, j int) bool {
			return resultOutputs[i].Name < resultOutputs[j].Name
		})

		// FIXME: more beautiful output
		printer := tableprinter.New(os.Stdout, false, 300)
		printer.AddHeader([]string{"Name", "Status", "Conclusion"})
		for _, output := range resultOutputs {
			printer.AddField(output.Name)
			printer.AddField(output.Status)
			printer.AddField(output.Conclusion)
			printer.EndRow()
		}

		printer.Render()

		return nil
	}
}
