package main

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/dnys1/unpub"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/mod/semver"
)

//go:embed build
var staticFS embed.FS

var (
	memFS = InMemFS{}
)

type PkgVersion struct {
	Package string
	Version string
}

type InMemFS map[PkgVersion]fs.File

func SetupRoutes(r *mux.Router, s UnpubService) {
	r.Path("/api/packages/{name}").Methods(http.MethodOptions, http.MethodGet).HandlerFunc(s.GetVersions)
	r.Path("/api/packages/{name}/versions/{version}").Methods(http.MethodOptions, http.MethodGet).HandlerFunc(s.GetVersion)
	r.Path("/packages{name}/versions/{version}.tar.gz").Methods(http.MethodOptions, http.MethodGet).HandlerFunc(s.Download)
	r.Path("/api/packages/versions/new").Methods(http.MethodOptions, http.MethodGet).HandlerFunc(s.GetUploadUrl)
	r.Path("/api/packages/versions/newUpload").Methods(http.MethodOptions, http.MethodPost).HandlerFunc(s.Upload)
	r.Path("/api/packages/versions/newUploadFinish").Methods(http.MethodOptions, http.MethodGet).HandlerFunc(s.UploadFinish)
	r.Path("/api/packages/{name}/uploaders").Methods(http.MethodOptions, http.MethodPost).HandlerFunc(s.AddUploader)
	r.Path("/api/packages/{name}/uploaders/{email}").Methods(http.MethodOptions, http.MethodDelete).HandlerFunc(s.RemoveUploader)
	r.Path("/webapi/packages").Methods(http.MethodOptions, http.MethodGet).HandlerFunc(s.GetPackages)
	r.Path("/webapi/package/{name}/{version}").Methods(http.MethodOptions, http.MethodGet).HandlerFunc(s.GetPackageDetails)

	staticFS, err := fs.Sub(staticFS, "build")
	if err != nil {
		panic(err)
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticFS)))

	r.Use(func(next http.Handler) http.Handler {
		return handlers.LoggingHandler(os.Stdout, next)
	})
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	})
	r.Use(mux.CORSMethodMiddleware(r))
}

type UnpubService interface {
	GetVersions(w http.ResponseWriter, r *http.Request)
	GetVersion(w http.ResponseWriter, r *http.Request)
	Download(w http.ResponseWriter, r *http.Request)
	GetUploadUrl(w http.ResponseWriter, r *http.Request)
	Upload(w http.ResponseWriter, r *http.Request)
	UploadFinish(w http.ResponseWriter, r *http.Request)
	AddUploader(w http.ResponseWriter, r *http.Request)
	RemoveUploader(w http.ResponseWriter, r *http.Request)
	GetPackages(w http.ResponseWriter, r *http.Request)
	GetPackageDetails(w http.ResponseWriter, r *http.Request)
}

type UnpubServiceImpl struct {
	DB            unpub.UnpubDb
	UploaderEmail string
	Addr          string
}

func (s *UnpubServiceImpl) GetVersions(w http.ResponseWriter, r *http.Request) {
	pkgName, ok := mux.Vars(r)["name"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}
	pkg, err := s.DB.QueryPackage(pkgName)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		writeInternalErr(w, err)
		return
	}

	sort.Slice(pkg.Versions, func(i, j int) bool {
		return semver.Compare(pkg.Versions[i].Version, pkg.Versions[j].Version) == -1
	})

	resp := struct {
		Name     string               `json:"name"`
		Latest   unpub.UnpubVersion   `json:"latest"`
		Versions []unpub.UnpubVersion `json:"versions"`
	}{
		Name:     pkg.Name,
		Latest:   pkg.Versions[len(pkg.Versions)-1],
		Versions: pkg.Versions,
	}
	writeJSON(w, resp)
}

func (s *UnpubServiceImpl) GetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pkgName, ok := vars["name"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}
	version, ok := vars["version"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}
	version = semver.Canonical(version)

	pkg, err := s.DB.QueryPackage(pkgName)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		writeInternalErr(w, err)
		return
	}
	var foundVersion *unpub.UnpubVersion
	for _, v := range pkg.Versions {
		if semver.Compare(version, v.Version) == 0 {
			foundVersion = &v
		}
	}
	if foundVersion == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, foundVersion)
}

func (s *UnpubServiceImpl) Download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pkgName, ok := vars["name"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}
	version, ok := vars["version"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}
	version = semver.Canonical(version)

	file, ok := memFS[PkgVersion{Package: pkgName, Version: version}]
	if !ok {
		http.Redirect(w, r, fmt.Sprintf("https://pub.dev%s", r.URL.Path), http.StatusFound)
		return
	}
	if isPubClient(r) {
		err := s.DB.IncreaseDownloads(pkgName, version)
		if err != nil {
			writeInternalErr(w, err)
			return
		}
	}
	st, err := file.Stat()
	if err != nil {
		writeInternalErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, filepath.Base(r.URL.Path), st.ModTime(), file.(io.ReadSeeker))
}

func (s *UnpubServiceImpl) GetUploadUrl(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		URL    string                 `json:"url"`
		Fields map[string]interface{} `json:"fields"`
	}{
		URL:    fmt.Sprintf("%s/api/packages/versions/newUpload", s.Addr),
		Fields: map[string]interface{}{},
	}
	writeJSON(w, resp)
}

func (s *UnpubServiceImpl) Upload(w http.ResponseWriter, r *http.Request) {
	email, err := s.getUploaderEmail(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	reader, err := r.MultipartReader()
	if err != nil {
		writeInternalErr(w, err)
		return
	}
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	var file multipart.File
	var found bool
outer:
	for _, headers := range form.File {
		for _, header := range headers {
			if strings.Contains(header.Filename, ".tar.gz") {
				found = true
				file, err = header.Open()
				if err != nil {
					writeInternalErr(w, err)
					return
				}
				break outer
			}
		}
	}
	if !found {
		writeBadRequest(w, errors.New("no file upload"))
		return
	}
	gr, err := gzip.NewReader(file)
	if err != nil {
		writeInternalErr(w, err)
		return
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	readFile := func(header *tar.Header) (string, error) {
		var sb strings.Builder
		_, err := io.CopyN(&sb, tr, header.Size)
		return sb.String(), err
	}

	var version unpub.UnpubVersion
	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			writeInternalErr(w, err)
			return
		}

		switch strings.ToLower(filepath.Base(header.Name)) {
		case "pubspec.yaml":
			str, err := readFile(header)
			if err != nil {
				writeInternalErr(w, err)
				return
			}
			version.PubspecYAML = str
		case "readme.md":
			str, err := readFile(header)
			if err != nil {
				writeInternalErr(w, err)
				return
			}
			version.Readme = &str
		case "changelog.md":
			str, err := readFile(header)
			if err != nil {
				writeInternalErr(w, err)
				return
			}
			version.Readme = &str
		}
	}

	if version.PubspecYAML == "" {
		writeBadRequest(w, errors.New("no pubspec found"))
		return
	}

	pubspec, err := version.Pubspec()
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	pkg, err := s.DB.QueryPackage(pubspec.Name)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			pkg = unpub.NewPackage(
				pubspec.Name,
				pubspec.PublishTo == "none",
				[]string{email},
			)
		} else {
			writeInternalErr(w, err)
			return
		}
	}
	pkg.AddVersion(version)
	err = s.DB.SavePackage(pkg)
	if err != nil {
		writeInternalErr(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/api/packages/versions/newUploadFinish", s.Addr), http.StatusFound)
}

func (s *UnpubServiceImpl) UploadFinish(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, struct {
		Success interface{} `json:"success"`
	}{
		Success: struct {
			Message string `json:"message"`
		}{
			Message: "Successfully uploaded package",
		},
	})
}

func (s *UnpubServiceImpl) AddUploader(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pkgName, ok := vars["name"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		writeBadRequest(w, nil)
		return
	}

	uploaderEmail, err := s.getUploaderEmail(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	pkg, err := s.DB.QueryPackage(pkgName)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	var foundEmail bool
	for _, uploader := range pkg.Uploaders {
		if uploader == email {
			writeBadRequest(w, errors.New("uploader already exists"))
			return
		}
		if uploader == uploaderEmail {
			foundEmail = true
		}
	}
	if !foundEmail {
		writeBadRequest(w, errors.New("no permission"))
		return
	}

	err = s.DB.AddUploader(pkgName, email)
	if err != nil {
		writeInternalErr(w, err)
		return
	}

	w.Write([]byte("uploader added"))
}

func (s *UnpubServiceImpl) RemoveUploader(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pkgName, ok := vars["name"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		writeBadRequest(w, nil)
		return
	}

	uploaderEmail, err := s.getUploaderEmail(r)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	pkg, err := s.DB.QueryPackage(pkgName)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	var foundEmail bool
	for _, uploader := range pkg.Uploaders {
		if uploader == email {
			writeBadRequest(w, errors.New("uploader already exists"))
			return
		}
		if uploader == uploaderEmail {
			foundEmail = true
		}
	}
	if !foundEmail {
		writeBadRequest(w, errors.New("no permission"))
		return
	}

	err = s.DB.RemoveUploader(pkgName, email)
	if err != nil {
		writeInternalErr(w, err)
		return
	}

	w.Write([]byte("uploader removed"))
}

func (s *UnpubServiceImpl) GetPackages(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	size, err := strconv.Atoi(params.Get("size"))
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	page, err := strconv.Atoi(params.Get("page"))
	if err != nil {
		writeBadRequest(w, err)
		return
	}
	sort := params.Get("sort")
	if sort == "" {
		sort = "download"
	}
	q := params.Get("q")

	queryReq := unpub.UnpubDbQuery{
		Size: size,
		Page: page,
		Sort: sort,
	}
	if strings.HasPrefix(q, "email:") {
		queryReq.Uploader = strings.TrimPrefix(q, "email:")
	} else if strings.HasPrefix(q, "dependency:") {
		queryReq.Dependency = strings.TrimPrefix(q, "dependency:")
	} else {
		queryReq.Keyword = q
	}

	packages, err := s.DB.QueryPackages(queryReq)
	if err != nil {
		writeInternalErr(w, err)
		return
	}
	var listApiPackages []unpub.ListApiPackage
	for _, pkg := range packages.Packages {
		listApiPackages = append(listApiPackages, pkg.ToListApiPackage())
	}
	writeJSON(w, struct {
		Data unpub.ListApi `json:"data"`
	}{
		Data: unpub.ListApi{
			Count:    packages.Count,
			Packages: listApiPackages,
		},
	})
}

func (s *UnpubServiceImpl) GetPackageDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pkgName, ok := vars["name"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}
	version, ok := vars["version"]
	if !ok {
		writeBadRequest(w, nil)
		return
	}
	version = semver.Canonical(version)

	pkg, err := s.DB.QueryPackage(pkgName)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		writeInternalErr(w, err)
		return
	}
	var v *unpub.UnpubVersion
	if version == "latest" {
		latest := pkg.LatestVersion()
		v = &latest
	} else {
		for _, _v := range pkg.Versions {
			if semver.Compare(_v.Version, version) == 0 {
				v = &_v
			}
		}
	}
	if v == nil {
		http.NotFound(w, r)
		return
	}
	sort.Slice(pkg.Versions, func(i, j int) bool {
		return semver.Compare(pkg.Versions[i].Version, pkg.Versions[j].Version) == -1
	})

	var detailViewVersions []unpub.DetailViewVersion

	for _, _v := range pkg.Versions {
		detailViewVersions = append(detailViewVersions, unpub.DetailViewVersion{
			Version:   _v.Version,
			CreatedAt: _v.CreatedAt,
		})
	}

	pubspec, err := v.Pubspec()
	if err != nil {
		writeInternalErr(w, err)
		return
	}

	authors := []string{}
	if pubspec.Author != "" {
		authors = append(authors, pubspec.Author)
	}

	dependencies := []string{}
	for _, dep := range pubspec.Dependencies {
		dependencies = append(dependencies, dep.Name)
	}
	data := unpub.WebAPIDetailView{
		Name:         pkg.Name,
		Version:      v.Version,
		Description:  pubspec.Description,
		Homepage:     pubspec.Homepage,
		Uploaders:    pkg.Uploaders,
		CreatedAt:    v.CreatedAt,
		Readme:       v.Readme,
		Changelog:    v.Changelog,
		Versions:     detailViewVersions,
		Authors:      authors,
		Dependencies: dependencies,
		Tags:         []string{"flutter", "web", "other"},
	}

	writeJSON(w, struct {
		Data unpub.WebAPIDetailView `json:"data"`
	}{
		Data: data,
	})
}

var _ = (UnpubService)(&UnpubServiceImpl{})

func writeJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.Write(b)
}

func writeInternalErr(w http.ResponseWriter, err error) {
	log.Printf("internal server error: %v\n", err)
	v := fmt.Sprintf("%v", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(v))
}

func writeBadRequest(w http.ResponseWriter, err error) {
	log.Printf("bad request: %v\n", err)
	v := fmt.Sprintf("%v", err)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(v))
}

func isPubClient(r *http.Request) bool {
	userAgent := r.Header.Get("User-Agent")
	return strings.Contains(strings.ToLower(userAgent), "dart pub")
}

func (s *UnpubServiceImpl) getUploaderEmail(r *http.Request) (string, error) {
	var (
		unauthorized = errors.New("missing authorization")
	)

	if s.UploaderEmail != "" {
		return s.UploaderEmail, nil
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", unauthorized
	}

	fields := strings.Fields(auth)
	if len(fields) != 2 {
		return "", unauthorized
	}
	authToken := fields[1]

	uri, err := url.Parse("https://oauth2.googleapis.com/tokeninfo")
	if err != nil {
		return "", err
	}
	q := uri.Query()
	q.Add("access_token", authToken)
	uri.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, uri.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", errors.New(string(body))
	}
	var tokenInfo struct {
		Email string `json:"string"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenInfo)
	return tokenInfo.Email, err
}
