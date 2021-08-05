package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dnys1/launch_unpub/cmd"
)

const (
	envUnpubHost = "UNPUB_HOST"
	envPort      = "UNPUB_PORT"
	envGitUrl    = "UNPUB_GIT_URL"
	envBranch    = "UNPUB_GIT_REF"
)

func warnDefaultEnv(env string, defaultVal interface{}) {
	log.Printf("%s not provided, defaulting to %v\n", env, defaultVal)
}

func main() {
	gitUrl := os.Getenv(envGitUrl)
	if gitUrl == "" {
		log.Fatalf("must set %s\n", envGitUrl)
	}
	gitRef := os.Getenv(envBranch)
	if gitRef == "" {
		gitRef = "main"
		warnDefaultEnv(envBranch, gitRef)
	}
	host := os.Getenv(envUnpubHost)
	if host == "" {
		host = "unpub"
		warnDefaultEnv(envUnpubHost, host)
	}
	port := os.Getenv(envPort)
	if port == "" {
		port = "8000"
		warnDefaultEnv(envPort, port)
	}

	if err := cmd.Run(gitUrl, gitRef, fmt.Sprintf("http://%s:%s", host, port)); err != nil {
		log.Fatalln(err)
	}
}
