package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/api"
)

// https://docs.github.com/ja/rest/pulls/pulls?apiVersion=2022-11-28#get-a-pull-request
type pullRequest struct {
	Title string `json:"title"`
	Head  struct {
		SHA string `json:"sha"`
	} `json:"head"`
}

func (c *APIClient) GetPRLatestCommitSHA(ctx context.Context, repoOwner, repoName string, prNumber int) (string, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return "", err
	}

	var pr pullRequest
	if err := restClient.DoWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("repos/%s/%s/pulls/%d", repoOwner, repoName, prNumber),
		nil,
		&pr,
	); err != nil {
		return "", err
	}

	return pr.Head.SHA, nil
}
