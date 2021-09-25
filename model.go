package main

import (
	"time"
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
	Authors      []*string           `json:"authors"`
	Dependencies *[]string           `json:"dependencies"`
	Tags         []string            `json:"tags"`
}

type UnpubVersion struct {
	Version          string    `json:"version" gorm:"primarykey"`
	PubspecYAML      string    `json:"pubspecYaml"`
	Uploader         *string   `json:"uploader"`
	Readme           *string   `json:"readme"`
	Changelog        *string   `json:"changelog"`
	UnpubPackageName string    `json:"-"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type UnpubPackage struct {
	Name      string         `json:"name" gorm:"primarykey"`
	Versions  []UnpubVersion `json:"versions"`
	Private   bool           `json:"private"`
	Uploaders string         `json:"uploaders" gorm:"index"`
	Downloads int            `json:"download"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type UnpubQueryResult struct {
	Count    int             `json:"count"`
	Packages []*UnpubPackage `json:"packages"`
}
