// +build e2e

package unpub_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/dnys1/unpub"
	"github.com/dnys1/unpub/server"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

const (
	unpubPort     = "4000"
	uploaderEmail = "test@example.com"
)

func newServer(t *testing.T) *server.UnpubServiceImpl {
	const (
		inMemory = true
		path     = ""
	)

	db, err := unpub.NewUnpubLocalDb(inMemory, path)
	require.NoError(t, err)

	return &server.UnpubServiceImpl{
		InMemory:      inMemory,
		Path:          path,
		DB:            db,
		UploaderEmail: uploaderEmail,
		Addr:          "http://localhost:4000",
	}
}

func pubClean() ([]byte, error) {
	cmd := exec.Command("dart", "pub", "cache", "clean")
	cmd.Stdin = strings.NewReader("y")
	return cmd.CombinedOutput()
}

func pubPublish(dir string) ([]byte, error) {
	cmd := exec.Command("dart", "pub", "publish")
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("PUB_HOSTED_URL=http://localhost:%s", unpubPort),
	)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("y")
	cmd.Stdout = os.Stdout
	return nil, cmd.Run()
}

func pubGet(dir string) ([]byte, error) {
	cmd := exec.Command("dart", "pub", "get")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	return nil, cmd.Run()
}

func TestE2E(t *testing.T) {
	const (
		pkgA = "test/pkg_a"
		pkgB = "test/pkg_b"
	)

	require := require.New(t)

	dartPath, err := exec.LookPath("dart")
	require.NoError(err)
	t.Logf("Found dart executable: %s\n", dartPath)

	r := mux.NewRouter()
	svc := newServer(t)
	server.SetupRoutes(r, svc)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", unpubPort),
		Handler: r,
	}
	defer httpServer.Close()

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			require.NotErrorIs(err, http.ErrServerClosed)
		}
	}()

	out, err := pubClean()
	require.NoErrorf(err, "pub clean: %v\n%s\n", err, out)

	out, err = pubPublish(pkgA)
	require.NoErrorf(err, "pub publish: %v\n%s\n", err, out)

	out, err = pubGet(pkgB)
	require.NoErrorf(err, "pub get: %v\n%s\n", err, out)
}
