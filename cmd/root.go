package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/go-gh/v2/pkg/prompter"
	"github.com/cli/go-gh/v2/pkg/tableprinter"

	"github.com/briandowns/spinner"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/kmtym1998/gh-prowl/entity"
)

func NewRootCmd(ec *ExecutionContext) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gh-prowl",
		Short: "Notify GitHub Actions status to your device",
		RunE: func(cmd *cobra.Command, args []string) error {
			current, err := cmd.Flags().GetBool("current-branch")
			if err != nil {
				return fmt.Errorf("failed to get flag: %w", err)
			}

			return run(&rootOption{
				ec:      ec,
				current: current,
			})
		},
	}

	f := rootCmd.Flags()
	f.BoolP("current-branch", "c", false, "monitor the latest check status of the current branch's PR")

	return rootCmd
}

type rootOption struct {
	ec      *ExecutionContext
	current bool
}

func run(o *rootOption) error {
	_ctx := context.Background()
	ctx, cancel := context.WithTimeout(_ctx, 30*time.Minute)
	defer cancel()

	prList, err := o.ec.ApiClient.ListPullRequests(ctx, o.ec.RepoOwner, o.ec.RepoName, 10)
	if err != nil {
		return fmt.Errorf("failed to list pull requests: %w", err)
	}

	var selectedPR *entity.SimplePR
	if o.current {
		currentBranchPR, found := lo.Find(prList.Items, func(pr *entity.SimplePR) bool {
			return pr.HeadRef == o.ec.CurrentBranch
		})
		if found {
			selectedPR = currentBranchPR
		}
	}

	if selectedPR == nil {
		io := iostreams.System()
		fmt.Printf("Total PRs: %d\n", prList.Total)
		p := prompter.New(io.In, io.Out, io.ErrOut)
		selected, err := p.Select(
			"Select a PR to prowl",
			"",
			lo.Map(prList.Items, func(pr *entity.SimplePR, _ int) string {
				return fmt.Sprintf("#%d %s", pr.Number, pr.Title)
			}),
		)
		if err != nil {
			if strings.Contains(err.Error(), "interrupt") {
				return nil
			}
			return fmt.Errorf("failed to prompt user: %w", err)
		}

		selectedPR = prList.Items[selected]
	}

	fmt.Printf("🦉 Selected PR: %s\n", selectedPR.Title)
	fmt.Printf("🦉 View this PR on GitHub: %s\n", selectedPR.URL)

	sha, err := o.ec.ApiClient.GetPRLatestCommitSHA(ctx, o.ec.RepoOwner, o.ec.RepoName, selectedPR.Number)
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

		checkRunList, err := o.ec.ApiClient.ListCheckRuns(ctx, o.ec.RepoOwner, o.ec.RepoName, sha)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				return err
			}

			return fmt.Errorf("failed to list check runs: %w", err)
		}

		if lo.SomeBy(checkRunList.Items, func(checkRun *entity.SimpleCheckRun) bool {
			return checkRun.Status != entity.CheckRunStatusCompleted
		}) {
			time.Sleep(o.ec.PollingInterval)
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

		return o.ec.SoundNotifier.Notify()
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
