package api

import (
	"context"
	"fmt"
	"net/http"
)

// https://docs.github.com/ja/rest/pulls/pulls?apiVersion=2022-11-28#get-a-pull-request
type pullRequest struct {
	Title string `json:"title"`
	Head  struct {
		SHA string `json:"sha"`
	} `json:"head"`
}

func (c *APIClient) GetPRLatestCommitSHA(ctx context.Context, repoOwner, repoName string, prNumber int) (string, error) {
	var pr pullRequest
	if err := c.rest.DoWithContext(
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
