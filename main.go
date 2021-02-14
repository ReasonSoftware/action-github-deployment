package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/go-github/v33/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/ReasonSoftware/action-github-deployment/internal/app"
)

// Version of an application
const Version string = "1.0.0"

func init() {
	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	log.Info("action version ", Version)

	vars := []string{
		"GITHUB_TOKEN",
		"GITHUB_REPOSITORY",
		"GITHUB_REF",
	}

	for i, v := range vars {
		if os.Getenv(v) == "" {
			log.Fatal("missing environmental variable ", vars[i])
		}
	}

	if os.Getenv("DEPLOYMENT") != "" && os.Getenv("STATUS") == "" {
		log.Fatal("missing environmental variable STATUS")
	}

	if os.Getenv("DEPLOYMENT") != "" && os.Getenv("STATUS") != "" && os.Getenv("ENVIRONMENT") != "" {
		log.Warn("ENVIRONMENT environmental variable is redundant and can be omitted")
	}
}

func main() {
	cli := github.NewClient(
		oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
		)))

	var env string
	if os.Getenv("ENVIRONMENT") != "" {
		env = os.Getenv("ENVIRONMENT")
	} else {
		env = "production"
	}

	var id int64
	if os.Getenv("DEPLOYMENT") == "" {
		log.Infof("creating new deployment for %v environment", env)

		var err error
		id, err = app.CreateDeployment(cli.Repositories, env, os.Getenv("GITHUB_REF"))
		if err != nil {
			log.Fatal(errors.Wrap(err, "error creating deployment"))
		}

		log.Info("successfully created deployment ", id)

		fmt.Printf("::set-output name=ID::%v\n", id)
	} else {
		n, err := strconv.Atoi(os.Getenv("DEPLOYMENT"))
		if err != nil {
			log.Fatal("invalid deployment id")
		}

		id = int64(n)
	}

	if os.Getenv("STATUS") != "" {
		log.Infof("updating status of a deployment %v in %v environment", id, env)

		if err := app.UpdateStatus(cli.Repositories, id, os.Getenv("STATUS")); err != nil {
			log.Fatal(errors.Wrap(err, "error updating deployment status"))
		}

		log.Info("deployment status set to ", os.Getenv("STATUS"))
	}
}
