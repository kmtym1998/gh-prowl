package cmd

import (
	"context"
	"errors"
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

			targetRef, err := cmd.Flags().GetString("ref")
			if err != nil {
				return fmt.Errorf("failed to get flag: %w", err)
			}

			return rootRunE(&rootOption{
				ec:        ec,
				current:   current,
				targetRef: targetRef,
			})
		},
	}

	f := rootCmd.Flags()
	f.BoolP("current-branch", "c", false, "monitor the latest check status of the current branch's PR")
	f.StringP("ref", "r", "", "monitor the latest check status of the specified branch")

	return rootCmd
}

type rootOption struct {
	ec        *ExecutionContext
	current   bool
	targetRef string
}

func rootRunE(o *rootOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	ref, err := resolveRef(ctx, o)
	if err != nil {
		return err
	}

	indicator := spinner.New(spinner.CharSets[1], 100*time.Millisecond)
	indicator.Suffix = " Waiting for checks to complete..."
	indicator.Start()
	defer indicator.Stop()

	return monitorCheckRuns(ctx, o, ref, indicator)
}

func resolveRef(ctx context.Context, o *rootOption) (string, error) {
	if o.targetRef != "" {
		return o.targetRef, nil
	}

	prList, err := o.ec.ApiClient.ListPullRequests(ctx, o.ec.RepoOwner, o.ec.RepoName, 10)
	if err != nil {
		return "", fmt.Errorf("failed to list pull requests: %w", err)
	}

	selectedPR, err := selectPR(o, prList)
	if err != nil {
		return "", err
	}

	fmt.Printf("🦉 Selected PR: %s\n", selectedPR.Title)
	fmt.Printf("🦉 View this PR on GitHub: %s\n", selectedPR.URL)

	sha, err := o.ec.ApiClient.GetPRLatestCommitSHA(ctx, o.ec.RepoOwner, o.ec.RepoName, selectedPR.Number)
	if err != nil {
		return "", fmt.Errorf("failed to get latest commit SHA: %w", err)
	}
	return sha, nil
}

func selectPR(o *rootOption, prList *entity.SimplePRList) (*entity.SimplePR, error) {
	if o.current {
		if currentBranchPR, found := lo.Find(prList.Items, func(pr *entity.SimplePR) bool {
			return pr.HeadRef == o.ec.CurrentBranch
		}); found {
			return currentBranchPR, nil
		}
	}

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
			return nil, nil
		}
		return nil, fmt.Errorf("failed to select PR: %w", err)
	}

	return prList.Items[selected], nil
}

func monitorCheckRuns(ctx context.Context, o *rootOption, ref string, indicator *spinner.Spinner) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		checkRunList, err := o.ec.ApiClient.ListCheckRuns(ctx, o.ec.RepoOwner, o.ec.RepoName, ref)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
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

		indicator.Stop()
		printCheckRunResults(checkRunList.Items)
		return o.ec.SoundNotifier.Notify()
	}
}

func printCheckRunResults(checkRuns []*entity.SimpleCheckRun) {
	type output struct {
		Name       string
		Conclusion string
		URL        string
	}
	results := lo.Map(checkRuns, func(checkRun *entity.SimpleCheckRun, _ int) output {
		conclusion := "NULL"
		if checkRun.Conclusion != nil {
			conclusion = checkRun.Conclusion.String()
		}
		return output{
			Name:       checkRun.Name,
			Conclusion: conclusion,
			URL:        checkRun.URL,
		}
	})

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	printer := tableprinter.New(os.Stdout, true, 1000)
	printer.AddHeader([]string{"Check Name", "Conclusion"}, tableprinter.WithColor(grayWithBoldAndUnderline))
	for _, result := range results {
		printer.AddField(result.Name)
		printer.AddField(result.Conclusion, tableprinter.WithColor(colorForConclusion(result.Conclusion)))
		printer.EndRow()
	}
	printer.Render()
}

func colorForConclusion(conclusion string) func(string) string {
	switch conclusion {
	case entity.CheckRunConclusionSuccess.String():
		return green
	case entity.CheckRunConclusionFailure.String():
		return red
	default:
		return func(s string) string { return s }
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
