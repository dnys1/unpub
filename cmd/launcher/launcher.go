package main

import (
	"log"

	"github.com/dnys1/unpub"
)

func main() {
	launcher := unpub.NewLaunchFromEnv(true)

	if err := launcher.Run(); err != nil {
		log.Fatalln(err)
	}
}
