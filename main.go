package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
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

	indicator := spinner.New(spinner.CharSets[1], 100*time.Millisecond)
	indicator.Suffix = " Waiting for checks to complete..."
	indicator.Start()
	defer indicator.Stop()

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
			Conclusion string
			URL        string
		}
		resultOutputs := []output{}
		for _, checkRun := range checkRunList.Items {
			if checkRun.Conclusion == nil {
				resultOutputs = append(resultOutputs, output{
					Name:       checkRun.Name,
					Conclusion: "NULL",
					URL:        checkRun.URL,
				})
				continue
			}

			if !checkRun.Conclusion.IsSuccess() {
				resultOutputs = append(resultOutputs, output{
					Name:       checkRun.Name,
					Conclusion: checkRun.Conclusion.String(),
					URL:        checkRun.URL,
				})
				continue
			}

			resultOutputs = append(resultOutputs, output{
				Name:       checkRun.Name,
				Conclusion: checkRun.Conclusion.String(),
				URL:        checkRun.URL,
			})
		}
		sort.Slice(resultOutputs, func(i, j int) bool {
			return resultOutputs[i].Name < resultOutputs[j].Name
		})

		indicator.Stop()

		printer := tableprinter.New(os.Stdout, true, 1000)

		printer.AddHeader([]string{"Check Name", "Conclusion"}, tableprinter.WithColor(grayWithBoldAndUnderline))
		for _, output := range resultOutputs {
			printer.AddField(output.Name)
			printer.AddField(output.Conclusion, tableprinter.WithColor(func(s string) string {
				switch s {
				case entity.CheckRunConclusionSuccess.String():
					return green(s)
				case entity.CheckRunConclusionFailure.String():
					return red(s)
				default:
					return s
				}
			}))
			printer.EndRow()
		}

		printer.Render()

		return nil
	}
}

func red(s string) string {
	return "\033[31m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func grayWithBoldAndUnderline(s string) string {
	return "\033[1;4;90m" + s + "\033[0m"
}
