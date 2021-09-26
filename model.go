package unpub

import (
	"sort"
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
	Uploader    *string   `json:"uploader"`
	Readme      *string   `json:"readme"`
	Changelog   *string   `json:"changelog"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (v UnpubVersion) Pubspec() (*Pubspec, error) {
	var pubspec Pubspec
	return &pubspec, yaml.Unmarshal([]byte(v.PubspecYAML), &pubspec)
}

type UnpubPackage struct {
	Name      string         `json:"name"`
	Versions  []UnpubVersion `json:"versions"`
	Private   bool           `json:"private"`
	Uploaders []string       `json:"uploaders"`
	Downloads int            `json:"download"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func (pkg *UnpubPackage) AddVersion(version UnpubVersion) {
	pkg.Versions = append(pkg.Versions, version)
}

func (pkg *UnpubPackage) CreateVersion(
	version,
	pubspec string,
	uploader,
	readme,
	changelog *string,
) UnpubVersion {
	v := UnpubVersion{
		Version:     version,
		PubspecYAML: pubspec,
		Readme:      readme,
		Changelog:   changelog,
		Uploader:    uploader,
		CreatedAt:   time.Now().Truncate(time.Millisecond),
		UpdatedAt:   time.Now().Truncate(time.Millisecond),
	}
	pkg.Versions = append(pkg.Versions, v)
	return v
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
		CreatedAt: time.Now().Truncate(time.Millisecond),
		UpdatedAt: time.Now().Truncate(time.Millisecond),
	}
}

func (pkg *UnpubPackage) LatestVersion() UnpubVersion {
	sort.Slice(pkg.Versions, func(i, j int) bool {
		return semver.Compare(pkg.Versions[i].Version, pkg.Versions[j].Version) == -1
	})

	return pkg.Versions[len(pkg.Versions)-1]
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
