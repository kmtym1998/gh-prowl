package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kmtym1998/gh-prowl/entity"
)

type checkRunList struct {
	TotalCount int `json:"total_count"`
	CheckRuns  []struct {
		Name       string  `json:"name"`
		Status     string  `json:"status"`
		Conclusion *string `json:"conclusion"`
		HTMLURL    string  `json:"html_url"`
	} `json:"check_runs"`
}

// https://docs.github.com/ja/rest/checks/runs?apiVersion=2022-11-28#list-check-runs-for-a-git-reference
func (c *APIClient) ListCheckRuns(ctx context.Context, repoOwner, repoName string, commitSHA string) (*entity.SimpleCheckRunList, error) {
	var checkRuns checkRunList
	if err := c.rest.DoWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("repos/%s/%s/commits/%s/check-runs?per_page=100", repoOwner, repoName, commitSHA),
		nil,
		&checkRuns,
	); err != nil {
		return nil, err
	}

	var simpleCheckRuns []*entity.SimpleCheckRun
	for _, checkRun := range checkRuns.CheckRuns {
		simpleCheckRuns = append(simpleCheckRuns, &entity.SimpleCheckRun{
			Name:       checkRun.Name,
			Status:     entity.CheckRunStatus(checkRun.Status),
			Conclusion: (*entity.CheckRunConclusion)(checkRun.Conclusion),
			URL:        checkRun.HTMLURL,
		})
	}

	return &entity.SimpleCheckRunList{
		Total: checkRuns.TotalCount,
		Items: simpleCheckRuns,
	}, nil
}
