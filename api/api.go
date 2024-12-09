package api

import (
	"github.com/cli/go-gh/v2/pkg/api"
)

type APIClient struct {
	gql  *api.GraphQLClient
	rest *api.RESTClient
}

func NewAPIClient() (*APIClient, error) {
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}

	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}

	return &APIClient{gql, rest}, nil
}
