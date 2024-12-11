package entity

import "context"

type GitHubAPIClient interface {
	ListPullRequests(ctx context.Context, repoOwner, repoName string, limit int) (*SimplePRList, error)
	GetPRLatestCommitSHA(ctx context.Context, repoOwner, repoName string, prNumber int) (string, error)
	ListCheckRuns(ctx context.Context, repoOwner, repoName string, commitSHA string) (*SimpleCheckRunList, error)
}

type Notifier interface {
	Notify() error
}