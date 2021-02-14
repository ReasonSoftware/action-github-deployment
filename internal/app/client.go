package app

import (
	"context"

	"github.com/google/go-github/v33/github"
)

// Client is Repositories service of a GitHub Client
type Client interface {
	CreateDeployment(context.Context, string, string, *github.DeploymentRequest) (*github.Deployment, *github.Response, error)
	CreateDeploymentStatus(context.Context, string, string, int64, *github.DeploymentStatusRequest) (*github.DeploymentStatus, *github.Response, error)
}
