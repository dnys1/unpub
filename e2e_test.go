// +build e2e

package unpub_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func addFakeCredentials(t *testing.T) {
	file, err := os.Create(filepath.Join(os.TempDir(), "pub-credentials.json"))
	require.NoError(t, err)

	err = json.NewEncoder(file).Encode(struct {
		AccessToken   string   `json:"accessToken"`
		RefreshToken  string   `json:"refreshToken"`
		IDToken       string   `json:"idToken"`
		TokenEndpoint string   `json:"tokenEndpoint"`
		Scopes        []string `json:"scopes"`
		Expiration    int64    `json:"expiration"`
	}{
		AccessToken:   "12345",
		RefreshToken:  "12345",
		IDToken:       "12345",
		TokenEndpoint: "https://accounts.google.com/o/oauth2/token",
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Expiration: time.Now().Add(time.Hour).Unix(),
	})
	require.NoError(t, err)
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
		fmt.Sprintf("_PUB_TEST_CONFIG_DIR=%s", os.TempDir()),
		fmt.Sprintf("PUB_HOSTED_URL=http://localhost:%s", unpubPort),
	)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("y")
	cmd.Stdout = os.Stdout
	return nil, cmd.Run()
}

func pubGet(dir string) ([]byte, error) {
	cmd := exec.Command("dart", "pub", "get")
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("_PUB_TEST_CONFIG_DIR=%s", os.TempDir()),
	)
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

	addFakeCredentials(t)

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
