package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/pkg/errors"
)

// DeploymentStates contains valid deployment states
var DeploymentStates []string = []string{
	"error",
	"failure",
	"inactive",
	"in_progress",
	"queued",
	"pending",
	"success",
}

// CreateDeployment creates new deployment in a provided environment.
func CreateDeployment(cli Client, env, ref string) (int64, error) {
	var prod bool
	if env == "production" {
		prod = true
	}

	if err := validateGitHubRef(ref); err != nil {
		return 0, err
	}

	if err := validateGitHubRepository(); err != nil {
		return 0, err
	}

	r := strings.Split(ref, "/")[2]

	dep, _, err := cli.CreateDeployment(
		context.Background(),
		strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[0],
		strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[1],
		&github.DeploymentRequest{
			Environment:           &env,
			ProductionEnvironment: &prod,
			Ref:                   &r,
		})
	if err != nil {
		return 0, err
	}

	return dep.GetID(), nil
}

// UpdateStatus set a status of a desired deployment to a specified value.
func UpdateStatus(cli Client, id int64, state string) error {
	if err := validateGitHubRepository(); err != nil {
		return err
	}

	if !IsStateValid(state) {
		return errors.New(fmt.Sprintf("not supported deployment state %v, use one of %+v", state, DeploymentStates))
	}

	_, _, err := cli.CreateDeploymentStatus(
		context.Background(),
		strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[0],
		strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[1],
		id,
		&github.DeploymentStatusRequest{
			State: &state,
		})

	return err
}

func validateGitHubRef(ref string) error {
	if len(strings.Split(ref, "/")) != 3 {
		return errors.New(fmt.Sprintf("unexpected GITHUB_REF format: %v", ref))
	}

	return nil
}

func validateGitHubRepository() error {
	if len(strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")) != 2 {
		return errors.New(fmt.Sprintf("unexpected GITHUB_REPOSITORY format: %v", os.Getenv("GITHUB_REPOSITORY")))
	}

	return nil
}

// IsStateValid is an input validator for CreateDeploymentStatus api call
func IsStateValid(state string) bool {
	var valid bool
	for _, v := range DeploymentStates {
		if state == v {
			valid = true
		}
	}

	return valid
}
