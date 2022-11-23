package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"os/exec"

	"github.com/google/go-github/v33/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/ReasonSoftware/action-github-deployment/internal/app"
)

func init() {
	log.SetReportCaller(false)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	log.Info("action version ", app.Version)

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

		cmd := exec.Command("/bin/sh","-c",fmt.Sprintf("echo \"ID=%v\" >> $GITHUB_OUTPUT",id))
        cmd.Stdout = os.Stdout
        if err := cmd.Run(); err != nil {
           fmt.Printf("error executing shell command: %v", err.Error())
           os.Exit(1)
        }
	} else {
		n, err := strconv.Atoi(os.Getenv("DEPLOYMENT"))
		if err != nil {
			log.Fatal("invalid deployment id")
		}

		id = int64(n)
	}

	if os.Getenv("STATUS") != "" {
		log.Infof("updating status of a deployment %v in %v environment", id, env)

		failed := false
		if os.Getenv("FAIL") != "" {
			var err error
			failed, err = strconv.ParseBool(os.Getenv("FAIL"))
			if err != nil {
				log.Fatal(errors.Wrap(err, "error parsing env.var 'FAIL'"))
			}
		}

		if failed {
			log.Warnf("FAIL=%v, updating deployment status to 'failure'", failed)

			if err := app.UpdateStatus(cli.Repositories, id, "failure"); err != nil {
				log.Fatal(errors.Wrap(err, "error updating deployment status"))
			}
		} else {
			if err := app.UpdateStatus(cli.Repositories, id, os.Getenv("STATUS")); err != nil {
				log.Fatal(errors.Wrap(err, "error updating deployment status"))
			}
		}

		log.Info("deployment status set to ", os.Getenv("STATUS"))
	}
}
