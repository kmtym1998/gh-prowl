package api

import (
	"context"

	"github.com/kmtym1998/gh-prowl/entity"
)

type listPullRequestsResult struct {
	Repository struct {
		PullRequests struct {
			TotalCount int `json:"totalCount"`
			Nodes      []struct {
				Number int    `json:"number"`
				Title  string `json:"title"`
				Author struct {
					Login string `json:"login"`
				} `json:"author"`
				BaseRef struct {
					Name       string `json:"name"`
					Repository struct {
						Name  string `json:"name"`
						Owner struct {
							Login string `json:"login"`
						} `json:"owner"`
					} `json:"repository"`
				} `json:"baseRef"`
				HeadRef struct {
					Name       string `json:"name"`
					Repository struct {
						Name  string `json:"name"`
						Owner struct {
							Login string `json:"login"`
						} `json:"owner"`
					} `json:"repository"`
				} `json:"headRef"`
			} `json:"nodes"`
		} `json:"pullRequests"`
	} `json:"repository"`
}

func (c *APIClient) ListPullRequests(ctx context.Context, repoOwner, repoName string, limit int) (*entity.SimplePRList, error) {
	query := `
		query PullRequestList(
			$owner: String!,
			$repo: String!,
			$limit: Int!,
			$state: [PullRequestState!] = OPEN
		) {
			repository(owner: $owner, name: $repo) {
				pullRequests(
					states: $state,
					first: $limit,
					orderBy: {field: CREATED_AT, direction: DESC}
				) {
					totalCount
					nodes {
						number
						title
						author {
							login
						}
						baseRef {
							repository {
								name
								owner {
									login
								}
							}
							name
						}
						headRef {
							repository {
								name
								owner {
									login
								}
							}
							name
						}
					}
				}
			}
		}`

	validLimit := min(limit, 100)
	variables := map[string]interface{}{
		"owner": repoOwner,
		"repo":  repoName,
		"limit": validLimit,
	}

	response := listPullRequestsResult{}
	if err := c.gql.DoWithContext(ctx, query, variables, &response); err != nil {
		return nil, err
	}

	var result []*entity.SimplePR
	for _, pr := range response.Repository.PullRequests.Nodes {
		result = append(result, &entity.SimplePR{
			RepoOwner: repoOwner,
			RepoName:  repoName,
			Number:    pr.Number,
			Title:     pr.Title,
			Author:    pr.Author.Login,
			BaseRef:   pr.BaseRef.Name,
			HeadRef:   pr.HeadRef.Name,
		})
	}

	return &entity.SimplePRList{
		Total: response.Repository.PullRequests.TotalCount,
		Items: result,
	}, nil
}
