package unpub

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

type ListApi struct {
	Count    int              `json:"count"`
	Packages []ListApiPackage `json:"packages"`
}

type ListApiPackage struct {
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Tags        []string  `json:"tags"`
	Latest      string    `json:"latest"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type DetailViewVersion struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
}

type WebAPIDetailView struct {
	Name         string              `json:"name"`
	Version      string              `json:"version"`
	Description  string              `json:"description"`
	Homepage     string              `json:"homepage"`
	Uploaders    []string            `json:"uploaders"`
	CreatedAt    time.Time           `json:"createdAt"`
	Readme       *string             `json:"readme"`
	Changelog    *string             `json:"changelog"`
	Versions     []DetailViewVersion `json:"versions"`
	Authors      []string            `json:"authors"`
	Dependencies []string            `json:"dependencies"`
	Tags         []string            `json:"tags"`
}

type UnpubVersion struct {
	Version     string    `json:"version"`
	PubspecYAML string    `json:"pubspecYaml"`
	Uploader    *string   `json:"uploader,omitempty"`
	Readme      *string   `json:"readme,omitempty"`
	Changelog   *string   `json:"changelog,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (v UnpubVersion) Pubspec() (*Pubspec, error) {
	var pubspec Pubspec
	return &pubspec, yaml.Unmarshal([]byte(v.PubspecYAML), &pubspec)
}

func UnpubVersions(versionMap map[string]UnpubVersion) []UnpubVersion {
	var unpubVersions []UnpubVersion
	for _, v := range versionMap {
		unpubVersions = append(unpubVersions, v)
	}
	return unpubVersions
}

type UnpubPackage struct {
	Name      string                  `json:"name"`
	Versions  map[string]UnpubVersion `json:"versions"`
	Latest    string                  `json:"latest"`
	Private   bool                    `json:"private"`
	Uploaders []string                `json:"uploaders"`
	Downloads int                     `json:"download"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

func (pkg *UnpubPackage) AddVersion(version UnpubVersion) error {
	if _, ok := pkg.Versions[version.Version]; ok {
		return errors.New("version already exists")
	}
	if pkg.Latest != "" && semver.Compare(pkg.Latest, version.Version) != 1 {
		return fmt.Errorf("version must be > %s", pkg.Latest)
	}
	pkg.Versions[version.Version] = version
	pkg.Latest = version.Version
	return nil
}

func (pkg *UnpubPackage) CreateVersion(
	version,
	pubspec string,
	uploader,
	readme,
	changelog *string,
) (UnpubVersion, error) {
	v := UnpubVersion{
		Version:     version,
		PubspecYAML: pubspec,
		Readme:      readme,
		Changelog:   changelog,
		Uploader:    uploader,
		CreatedAt:   time.Now().Truncate(time.Millisecond),
		UpdatedAt:   time.Now().Truncate(time.Millisecond),
	}
	return v, pkg.AddVersion(v)
}

func NewPackage(
	name string,
	private bool,
	uploaders []string,
) UnpubPackage {
	return UnpubPackage{
		Name:      name,
		Private:   private,
		Uploaders: uploaders,
		Downloads: 0,
		Versions:  make(map[string]UnpubVersion),
		CreatedAt: time.Now().Truncate(time.Millisecond),
		UpdatedAt: time.Now().Truncate(time.Millisecond),
	}
}

func (pkg *UnpubPackage) LatestVersion() UnpubVersion {
	return pkg.Versions[pkg.Latest]
}

func (pkg *UnpubPackage) ToListApiPackage() ListApiPackage {
	latest := pkg.LatestVersion()
	pubspec, err := latest.Pubspec()
	if err != nil {
		panic(err)
	}
	return ListApiPackage{
		Name:        pkg.Name,
		Description: &pubspec.Description,
		Tags:        []string{"flutter", "web", "other"},
		Latest:      latest.Version,
		UpdatedAt:   latest.UpdatedAt,
	}
}

type UnpubQueryResult struct {
	Count    int             `json:"count"`
	Packages []*UnpubPackage `json:"packages"`
}
