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
	envLocalPath = "UNPUB_LOCAL_PATH"
)

func warnDefaultEnv(env string, defaultVal interface{}) {
	log.Printf("%s not provided, defaulting to %v\n", env, defaultVal)
}

func main() {
	localPath := os.Getenv(envLocalPath)
	gitUrl := os.Getenv(envGitUrl)
	if localPath == "" && gitUrl == "" {
		log.Fatalf("must set either %s or %s\n", envGitUrl, envLocalPath)
	}
	gitRef := os.Getenv(envBranch)
	if localPath == "" && gitRef == "" {
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

	if err := cmd.Run(localPath, gitUrl, gitRef, fmt.Sprintf("http://%s:%s", host, port)); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("\nUnpub has been successfully seeded!")
	fmt.Printf(`
Add the following to your pubspec.yaml for each package you want to use:

  package:
    hosted:
      name: package
      url: http://localhost:%s
    version: ^x.y.z
	`, port)
	fmt.Print("\nRemember to set an environment variable when publishing:\n\n")
	fmt.Printf("  $ PUB_HOSTED_URL=http://localhost:%s pub publish\n\n", port)
}
