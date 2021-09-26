package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/dnys1/unpub"
	"github.com/gorilla/mux"
)

var (
	launchUnpub   = flag.Bool("launch", false, "Seeds the unpub server using the provided environment variables")
	port          = flag.Int("port", 0, "The port to run the server on")
	uploaderEmail = flag.String("uploader-email", "", "The default uploader email to use")
)

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	db, err := unpub.NewUnpubBadgerDb(true, "")
	if err != nil {
		log.Fatalf("error opening db: %v\n", err)
	}
	if *port == 0 {
		if envPort := os.Getenv("UNPUB_PORT"); envPort != "" {
			*port, err = strconv.Atoi(envPort)
			if err != nil {
				log.Fatalf("bad port: %v\n", err)
			}
		} else {
			*port = 4000
		}
	}
	svc := &UnpubServiceImpl{
		DB:            db,
		UploaderEmail: *uploaderEmail,
		Addr:          fmt.Sprintf("http://localhost:%d", *port),
	}

	r := mux.NewRouter()
	SetupRoutes(r, svc)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "error in server: %v\n", err)
		}
	}()

	if *launchUnpub {
		go func() {
			launcher := unpub.NewLaunchFromEnv(false)
			launcher.ServerHost = "localhost"
			launcher.ServerPort = fmt.Sprintf("%d", *port)
			if err := launcher.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "error seeding unpub: %v\n", err)
			}
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig

	err = server.Shutdown(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error shutting down server: %v\n", err)
	}

	err = db.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing db: %v\n", err)
	}
}
