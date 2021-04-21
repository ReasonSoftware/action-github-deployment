package app_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-github/v33/github"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/ReasonSoftware/action-github-deployment/internal/app"
	"github.com/ReasonSoftware/action-github-deployment/mocks"
)

func TestCreateDeployment(t *testing.T) {
	assert := assert.New(t)

	reqContexts := make([]string, 0)

	type test struct {
		GitHubRepository string
		Environment      string
		Ref              string
		MockIn           *github.DeploymentRequest
		MockOut          *github.Deployment
		MockErr          error
		ExpectedOut      int64
		ExpectedErr      string
	}

	suite := map[string]test{
		"Self": {
			Environment:      "",
			Ref:              "refs/heads/master",
			GitHubRepository: "org/repo",
			MockIn: &github.DeploymentRequest{
				Environment:           github.String(""),
				ProductionEnvironment: github.Bool(false),
				Ref:                   github.String("master"),
				RequiredContexts:      &reqContexts,
			},
			MockOut: &github.Deployment{
				ID: github.Int64(123456789),
			},
			MockErr:     nil,
			ExpectedOut: 123456789,
			ExpectedErr: "",
		},
		"Branch": {
			Environment:      "production",
			Ref:              "refs/heads/master",
			GitHubRepository: "org/repo",
			MockIn: &github.DeploymentRequest{
				Environment:           github.String("production"),
				ProductionEnvironment: github.Bool(true),
				Ref:                   github.String("master"),
				RequiredContexts:      &reqContexts,
			},
			MockOut: &github.Deployment{
				ID: github.Int64(123456789),
			},
			MockErr:     nil,
			ExpectedOut: 123456789,
			ExpectedErr: "",
		},
		"Tag": {
			Environment:      "production",
			Ref:              "refs/tags/v1.0.0",
			GitHubRepository: "org/repo",
			MockIn: &github.DeploymentRequest{
				Environment:           github.String("production"),
				ProductionEnvironment: github.Bool(true),
				Ref:                   github.String("v1.0.0"),
				RequiredContexts:      &reqContexts,
			},
			MockOut: &github.Deployment{
				ID: github.Int64(123456789),
			},
			MockErr:     nil,
			ExpectedOut: 123456789,
			ExpectedErr: "",
		},
		"Custom Environment": {
			Environment:      "staging",
			Ref:              "refs/heads/dev",
			GitHubRepository: "org/repo",
			MockIn: &github.DeploymentRequest{
				Environment:           github.String("staging"),
				ProductionEnvironment: github.Bool(false),
				Ref:                   github.String("dev"),
				RequiredContexts:      &reqContexts,
			},
			MockOut: &github.Deployment{
				ID: github.Int64(123456789),
			},
			MockErr:     nil,
			ExpectedOut: 123456789,
			ExpectedErr: "",
		},
		"Incorrect GITHUB_REF": {
			Environment:      "production",
			Ref:              "refs/x",
			GitHubRepository: "org/repo",
			MockIn:           nil,
			MockOut:          nil,
			MockErr:          nil,
			ExpectedOut:      0,
			ExpectedErr:      "unexpected GITHUB_REF format: refs/x",
		},
		"Incorrect GITHUB_REPOSITORY": {
			Environment:      "production",
			Ref:              "refs/heads/master",
			GitHubRepository: "org/repo/x",
			MockIn:           nil,
			MockOut:          nil,
			MockErr:          nil,
			ExpectedOut:      0,
			ExpectedErr:      "unexpected GITHUB_REPOSITORY format: org/repo/x",
		},
		"API Error": {
			Environment:      "production",
			Ref:              "refs/heads/master",
			GitHubRepository: "org/repo",
			MockIn: &github.DeploymentRequest{
				Environment:           github.String("production"),
				ProductionEnvironment: github.Bool(true),
				Ref:                   github.String("master"),
				RequiredContexts:      &reqContexts,
			},
			MockOut:     nil,
			MockErr:     errors.New("reason"),
			ExpectedOut: 0,
			ExpectedErr: "reason",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		os.Setenv("GITHUB_REPOSITORY", test.GitHubRepository)
		defer os.Unsetenv("GITHUB_REPOSITORY")

		m := new(mocks.Client)
		m.On("CreateDeployment",
			context.Background(),
			strings.Split(test.GitHubRepository, "/")[0],
			strings.Split(test.GitHubRepository, "/")[1],
			test.MockIn).Return(test.MockOut, nil, test.MockErr).Once()

		result, err := app.CreateDeployment(m, test.Environment, test.Ref)

		if test.ExpectedErr != "" {
			assert.EqualError(err, test.ExpectedErr)
		} else {
			assert.Equal(nil, err)
		}

		assert.Equal(test.ExpectedOut, result)
	}
}

func TestUpdateStatus(t *testing.T) {
	assert := assert.New(t)

	type test struct {
		GitHubRepository string
		ID               int64
		State            string
		MockIn           *github.DeploymentStatusRequest
		MockErr          error
		ExpectedErr      string
	}

	suite := map[string]test{
		"Success": {
			GitHubRepository: "org/repo",
			ID:               123456789,
			State:            "pending",
			MockIn: &github.DeploymentStatusRequest{
				State: github.String("pending"),
			},
			MockErr:     nil,
			ExpectedErr: "",
		},
		"Not Supported State": {
			GitHubRepository: "org/repo",
			ID:               123456789,
			State:            "x",
			MockIn:           nil,
			MockErr:          nil,
			ExpectedErr:      fmt.Sprintf("not supported deployment state x, use one of %+v", app.DeploymentStates),
		},
		"Incorrect GITHUB_REPOSITORY": {
			GitHubRepository: "org/repo/x",
			ID:               123456789,
			State:            "pending",
			MockIn: &github.DeploymentStatusRequest{
				State: github.String("pending"),
			},
			MockErr:     nil,
			ExpectedErr: "unexpected GITHUB_REPOSITORY format: org/repo/x",
		},
		"API Error": {
			GitHubRepository: "org/repo",
			ID:               123456789,
			State:            "pending",
			MockIn: &github.DeploymentStatusRequest{
				State: github.String("pending"),
			},
			MockErr:     errors.New("reason"),
			ExpectedErr: "reason",
		},
	}

	var counter int
	for name, test := range suite {
		counter++
		t.Logf("Test Case %v/%v - %s", counter, len(suite), name)

		os.Setenv("GITHUB_REPOSITORY", test.GitHubRepository)
		defer os.Unsetenv("GITHUB_REPOSITORY")

		m := new(mocks.Client)
		m.On("CreateDeploymentStatus",
			context.Background(),
			strings.Split(test.GitHubRepository, "/")[0],
			strings.Split(test.GitHubRepository, "/")[1],
			test.ID,
			test.MockIn).Return(nil, nil, test.MockErr).Once()

		err := app.UpdateStatus(m, test.ID, test.State)

		if test.ExpectedErr != "" {
			assert.EqualError(err, test.ExpectedErr)
		} else {
			assert.Equal(nil, err)
		}
	}
}

func TestIsStateValid(t *testing.T) {
	var counter int
	for i, state := range app.DeploymentStates {
		counter++
		t.Logf("Test Case %v/%v - State %s", counter, len(app.DeploymentStates), app.DeploymentStates[i])

		res := app.IsStateValid(state)
		assert.EqualValues(t, true, res)
	}

	counter++
	t.Logf("Test Case %v/%v - Not Supported State", counter, counter)
	res := app.IsStateValid("not_supported_state")
	assert.EqualValues(t, false, res)
}
