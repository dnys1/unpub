package unpub

import (
	"database/sql/driver"
	"errors"

	"gopkg.in/yaml.v3"
)

// DependencyType identifies the type of dependency, i.e. regular, dev or override.
type DependencyType string

// Depdendency types
const (
	DependencyTypeRegular  DependencyType = "dependencies"
	DependencyTypeDev      DependencyType = "dev_dependencies"
	DependencyTypeOverride DependencyType = "dependency_overrides"
)

// DependencySource identifies the location of the dependency source, e.g. pub.dev, git, local
type DependencySource uint8

// Dependency sources
const (
	DependencySourceVersion DependencySource = iota
	DependencySourcePath
	DependencySourceGit
	DependencySourceHosted
	DependencySourceSDK
)

// Pubspec holds the contents of pubspec.yaml
type Pubspec struct {
	Name                string                 `yaml:"name"`
	Description         string                 `yaml:"description,omitempty"`
	Homepage            string                 `yaml:"homepage,omitempty"`
	Author              string                 `yaml:"author,omitempty"`
	Repository          string                 `yaml:"repository,omitempty"`
	PublishTo           string                 `yaml:"publish_to,omitempty"`
	Version             string                 `yaml:"version,omitempty"`
	Environment         *Environment           `yaml:"environment,omitempty"`
	Dependencies        map[string]*Dependency `yaml:"dependencies,omitempty"`
	DevDependencies     map[string]*Dependency `yaml:"dev_dependencies,omitempty"`
	DependencyOverrides map[string]*Dependency `yaml:"dependency_overrides,omitempty"`
}

// Scanner

func (pubspec *Pubspec) Scan(src interface{}) error {
	if pubspecYaml, ok := src.([]byte); ok {
		return yaml.Unmarshal(pubspecYaml, pubspec)
	}
	return errors.New("not a string")
}

// Valuer

func (pubspec *Pubspec) Value() (driver.Value, error) {
	if pubspec == nil {
		return nil, errors.New("nil pubspec")
	}
	return yaml.Marshal(pubspec)
}

// Marshaller
func (p *Pubspec) MarshalJSON() ([]byte, error) {
	return yaml.Marshal(p)
}

func (p *Pubspec) UnmarshalJSON(b []byte) error {
	return yaml.Unmarshal(b, p)
}

// Environment identifies the Dart environment.
type Environment struct {
	SDK string `yaml:"sdk,omitempty"`
}

// Dependency holds dependency information.
type Dependency struct {
	Name    string
	Version string
	Source  DependencySource
	Git     DependencyGit
	Path    DependencyPath
	Hosted  DependencyHosted
	SDK     DependencySDK
}

// DependencySDK is a dependency from an SDK.
type DependencySDK struct {
	SDK     string `yaml:"sdk"`
	Version string `yaml:"version,omitempty"`
}

// DependencyGit is a dependency from git.
type DependencyGit struct {
	Git *GitInfo `yaml:"git,omitempty"`
}

// DependencyPath is a local dependency.
type DependencyPath struct {
	Path string `yaml:"path,omitempty"`
}

// DependencyHosted is an externally hosted dependency.
type DependencyHosted struct {
	Version string      `yaml:"version,omitempty"`
	Hosted  *HostedInfo `yaml:"hosted,omitempty"`
}

// UnmarshalYAML unmarshals dependency info into the correct concrete type.
func (dep *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var version string
	err := unmarshal(&version)
	if err == nil && version != "" {
		dep.Source = DependencySourceVersion
		dep.Version = version
		return nil
	}

	var gitDep DependencyGit
	err = unmarshal(&gitDep)
	if err == nil && gitDep.Git != nil {
		dep.Source = DependencySourceGit
		dep.Git = gitDep
		return nil
	}

	var pathDep DependencyPath
	err = unmarshal(&pathDep)
	if err == nil && pathDep.Path != "" {
		dep.Source = DependencySourcePath
		dep.Path = pathDep
		return nil
	}

	var hostedDep DependencyHosted
	err = unmarshal(&hostedDep)
	if err == nil && hostedDep.Hosted != nil {
		dep.Source = DependencySourceHosted
		dep.Version = hostedDep.Version
		dep.Hosted = hostedDep
		return nil
	}

	var sdkDep DependencySDK
	err = unmarshal(&sdkDep)
	if err == nil && sdkDep.SDK != "" {
		dep.Source = DependencySourceSDK
		dep.Version = sdkDep.Version
		dep.SDK = sdkDep
		return nil
	}

	return nil
}

// GitInfo is the git-specific dependency information.
type GitInfo struct {
	URL     string `yaml:"git,omitempty"`
	SubInfo GitSubInfo
}

// GitSubInfo is the subset of associated git info.
type GitSubInfo struct {
	URL  string `yaml:"git,omitempty"`
	Ref  string `yaml:"ref,omitempty"`
	Path string `yaml:"path,omitempty"`
}

// UnmarshalYAML unmarshals git-related info when unmarshalling dependencies.
func (info *GitInfo) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var url string
	if err := unmarshal(&url); err == nil {
		info.URL = url
		return nil
	}

	var subInfo GitSubInfo
	err := unmarshal(&subInfo)
	info.URL = subInfo.URL
	info.SubInfo = subInfo

	return err
}

// HostedInfo is the externally hosted specific info.
type HostedInfo struct {
	Name string `yaml:"name,omitempty"`
	URL  string `yaml:"url,omitempty"`
	Path string `yaml:"path,omitempty"`
}

// AddDependency adds the given dependency to the Pubspec file.
func (pubspec *Pubspec) AddDependency(typ DependencyType, dependency *Dependency) {
	switch typ {
	case DependencyTypeRegular:
		pubspec.Dependencies[dependency.Name] = dependency
	case DependencyTypeDev:
		pubspec.DevDependencies[dependency.Name] = dependency
	case DependencyTypeOverride:
		pubspec.DependencyOverrides[dependency.Name] = dependency
	}
}

// HasDependency checks whether the pubspec contains a given dependency.
func (pubspec *Pubspec) HasDependency(dependency string) bool {
	_, ok := pubspec.Dependencies[dependency]
	return ok
}
