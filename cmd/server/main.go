package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/dnys1/unpub"
	"github.com/dnys1/unpub/server"
	"github.com/gorilla/mux"
)

var (
	launchUnpub   = flag.Bool("launch", false, "Seeds the unpub server using the provided environment variables")
	port          = flag.Int("port", 0, "The port to run the server on")
	uploaderEmail = flag.String("uploader-email", "test@example.com", "The default uploader email to use")
	inMemory      = flag.Bool("memory", false, "Runs the server in-memory, using no storage")
	path          = flag.String("path", "", "Directory to store DB files (defaults to temp dir, only valid if memory=false)")
	addr          = flag.String("addr", "http://localhost:${PORT}", "The hostname to serve unpub as")

	//go:embed build
	staticFS embed.FS
)

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	if !*inMemory && *path == "" {
		var err error
		*path, err = os.MkdirTemp("", "unpub")
		if err != nil {
			log.Fatalf("error creating temp dir: %v\n", err)
		}
	} else if *inMemory {
		*path = ""
	}
	db, err := unpub.NewUnpubLocalDb(*inMemory, *path)
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
	if *addr == "" {
		*addr = fmt.Sprintf("http://localhost:%d", *port)
	}
	svc := &server.UnpubServiceImpl{
		InMemory:      *inMemory,
		Path:          *path,
		DB:            db,
		UploaderEmail: *uploaderEmail,
		Addr:          *addr,
	}

	r := mux.NewRouter()
	server.SetupRoutes(r, svc)

	staticFS, err := fs.Sub(staticFS, "build")
	if err != nil {
		panic(err)
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticFS)))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: r,
	}

	go func() {
		log.Printf("Serving at %s\n", svc.Addr)
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
