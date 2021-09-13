package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func Run(gitUrl, branch, url string) error {
	dir, err := cloneGit(gitUrl, branch)
	if err != nil {
		return err
	}

	packageDirs, err := gatherPackages(gitUrl, dir)
	if err != nil {
		return err
	}
	if len(packageDirs) == 0 {
		return fmt.Errorf("no packages found in git repo")
	}

	err = uploadPackages(packageDirs, dir, url)
	if err != nil {
		return err
	}

	return nil
}

// cloneGit clones gitUrl@branch into a new temp directory and returns the directory.
func cloneGit(gitUrl, branch string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "unpub")
	if err != nil {
		return "", errors.Wrap(err, "create temp dir failed")
	}

	log.Printf("Created temporary dir: %s\n", tmpDir)
	log.Printf("Cloning %s branch %q to %s\n", gitUrl, branch, tmpDir)

	repo, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:        gitUrl,
		Depth:      1,
		NoCheckout: true,
	})
	if err != nil {
		return "", errors.Wrap(err, "git clone failed")
	}

	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/pull/*/head:refs/remotes/origin/pull/*"),
		},
		Depth: 1,
	})
	if err != nil {
		return "", errors.Wrap(err, "git fetch failed")
	}

	remoteRef := plumbing.NewRemoteReferenceName("origin", branch)
	commit, err := repo.ResolveRevision(plumbing.Revision(remoteRef))
	if err != nil {
		return "", errors.Wrapf(err, "git could not resolve hash for %s", remoteRef)
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", errors.Wrap(err, "git worktree failed")
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Hash:   *commit,
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	})
	if err != nil {
		return "", errors.Wrap(err, "git checkout failed")
	}

	return tmpDir, nil
}

// gatherPackages returns all the directories in the given directory tree
// which are Dart packages (i.e. have a pubspec.yaml).
func gatherPackages(gitUrl, dir string) ([]string, error) {
	packageDirs := []string{}
	const filename = ".gitignore"
	ignoreFile, err := os.ReadFile(filepath.Join(dir, filename))
	var matcher *gitignore.Matcher
	if err == nil {
		var patterns []gitignore.Pattern
		for _, line := range strings.Split(string(ignoreFile), "\n") {
			pattern := gitignore.ParsePattern(line, []string{})
			patterns = append(patterns, pattern)
		}
		m := gitignore.NewMatcher(patterns)
		matcher = &m
	}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.Name() == ".git" && d.IsDir() {
			return fs.SkipDir
		}
		if matcher != nil && (*matcher).Match(filepath.SplitList(path), d.IsDir()) {
			if d.IsDir() {
				return fs.SkipDir
			} else {
				return nil
			}
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "pubspec.yaml" {
			file, err := os.ReadFile(path)
			if err != nil {
				return errors.Wrapf(err, "error reading file %s", path)
			}
			var yamlMap map[string]interface{}
			err = yaml.Unmarshal(file, &yamlMap)
			if err != nil {
				return errors.Wrapf(err, "error decoding yaml %s", path)
			}
			if publishTo, ok := yamlMap["publish_to"]; ok && publishTo == "none" {
				return nil
			}
			packageDir := filepath.Dir(path)
			log.Printf("Found package: %s\n", filepath.Base(packageDir))
			packageDirs = append(packageDirs, packageDir)
		}
		return nil
	})
	return packageDirs, nil
}

// uploadPackages compresses and uploads packages to running unpub server.
func uploadPackages(packageDirs []string, tempDir, url string) error {
	for _, packageDir := range packageDirs {
		tarball, err := createTarball(tempDir, packageDir)
		if err != nil {
			return errors.Wrapf(err, "error creating tarball for dir %s", packageDir)
		}
		defer tarball.Close()

		err = uploadTarball(tarball, url)
		if err != nil {
			return errors.Wrapf(err, "error uploading %s", filepath.Base(packageDir))
		}
	}
	return nil
}

// createTarball compresses a package directory into a .tar.gz file.
func createTarball(tempDir, packageDir string) (*os.File, error) {
	dirname := filepath.Base(packageDir)
	filename := dirname + ".tar"
	changedir, err := filepath.Rel(tempDir, packageDir)
	if err != nil {
		return nil, err
	}
	tarfile := filepath.Join(tempDir, filename)
	gzipfile := filepath.Join(tempDir, filename+".gz")

	var xform string
	currentOS := runtime.GOOS
	if currentOS == "darwin" {
		xform = `-s:^\.::`
	} else if currentOS == "linux" {
		xform = `--xform=s:^\./::`
	} else {
		log.Fatalf("OS not supported: %s\n", currentOS)
	}
	cmd := exec.Command(
		"tar",
		"-cf",
		tarfile, // store archives in temp dir
		"-C",
		changedir, // change into the dir (may be nested)
		xform,     // rename files from ./filename.txt to filename.txt
		".",
	)
	cmd.Dir = tempDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("tar create: %s\n", out)
		return nil, err
	}

	// Delete top-level directory on linux (handled on mac).
	if currentOS == "linux" {
		delCmd := exec.Command(
			"tar",
			"--delete",
			"-f",
			tarfile,
			".",
		)
		delCmd.Dir = tempDir

		out, err = delCmd.CombinedOutput()
		if err != nil {
			log.Printf("tar delete: %s\n", out)
			return nil, err
		}

		// Compress tar with gzip
		compressCmd := exec.Command(
			"gzip",
			"-f",
			tarfile,
		)
		compressCmd.Dir = tempDir

		out, err = compressCmd.CombinedOutput()
		if err != nil {
			log.Printf("tar gzip: %s\n", out)
			return nil, err
		}
	} else {
		// Compress tar with gzip
		compressCmd := exec.Command(
			"tar",
			"-czf",
			gzipfile,
			"@"+filename,
		)
		compressCmd.Dir = tempDir

		out, err = compressCmd.CombinedOutput()
		if err != nil {
			log.Printf("tar gzip: %s\n", out)
			return nil, err
		}
	}

	return os.Open(gzipfile)
}

// uploadTarball pushes a tarball to a running unpub server.
func uploadTarball(tarball *os.File, url string) error {
	endpoint := fmt.Sprintf("%s/api/packages/versions/newUpload", url)

	var bb bytes.Buffer
	mw := multipart.NewWriter(&bb)

	field, err := mw.CreateFormFile("file", filepath.Base(tarball.Name()))
	if err != nil {
		return errors.Wrap(err, "could not create field")
	}
	if _, err := io.Copy(field, tarball); err != nil {
		return errors.Wrap(err, "could not read tarball")
	}
	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, &bb)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http status %d: %s", resp.StatusCode, bb)
	}
	return nil
}
