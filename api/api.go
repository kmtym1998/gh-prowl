package api

import (
	"github.com/cli/go-gh/v2/pkg/api"
)

type APIClient struct {
	gql  *api.GraphQLClient
	rest *api.RESTClient
}

type APIClientOption struct {
	GitHubHost      string
	GitHubAuthToken string
}

func NewAPIClient(o APIClientOption) (*APIClient, error) {
	clientOptions := api.ClientOptions{
		Host:      o.GitHubHost,
		AuthToken: o.GitHubAuthToken,
	}
	gql, err := api.NewGraphQLClient(clientOptions)
	if err != nil {
		return nil, err
	}

	rest, err := api.NewRESTClient(clientOptions)
	if err != nil {
		return nil, err
	}

	return &APIClient{gql, rest}, nil
}
