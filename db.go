package main

import (
	"errors"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UnpubDbQuery struct {
	Size    int
	Page    int
	Sort    string
	Keyword string
	// Uploader   string
	// Dependency string
}

type UnpubDb interface {
	QueryPackage(name string) (*UnpubPackage, error)
	QueryPackages(query UnpubDbQuery) (*UnpubQueryResult, error)
	AddVersion(name string, version UnpubVersion) error
	AddUploader(name, email string) error
	RemoveUploader(name, email string) error
	IncreaseDownloads(name, version string) error
}

type UnpubLocalDb struct {
	InMemory bool
	Path     string
	db       *gorm.DB
}

func NewUnpubLocalDb(inMem bool, path string) (*UnpubLocalDb, error) {
	if !inMem && path == "" {
		path = filepath.Join(os.TempDir(), "unpub.db")
	} else if inMem {
		path = "file::memory:?cache=shared"
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&UnpubVersion{},
		&UnpubPackage{},
	)
	if err != nil {
		return nil, err
	}

	return &UnpubLocalDb{Path: path, InMemory: inMem, db: db}, nil
}

func (db *UnpubLocalDb) QueryPackage(name string) (pkg *UnpubPackage, err error) {
	tx := db.db.Where("name = ?", name).First(&pkg)
	err = tx.Error
	return
}

func (db *UnpubLocalDb) QueryPackages(query UnpubDbQuery) (*UnpubQueryResult, error) {
	res := &UnpubQueryResult{}
	tx := db.db.
		Limit(query.Size).
		Offset(query.Page*query.Size).
		Where("name LIKE ?", "%"+query.Keyword+"%").
		Find(&res.Packages)
	if err := tx.Error; err != nil {
		return nil, err
	}
	res.Count = len(res.Packages)
	return res, nil
}

func (db *UnpubLocalDb) AddVersion(name string, version UnpubVersion) error {
	pkg, err := db.QueryPackage(name)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		var uploader string
		if version.Uploader != nil {
			uploader = *version.Uploader
		}
		res := db.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&UnpubPackage{
			Name:      name,
			Uploaders: uploader,
			Versions: []UnpubVersion{
				version,
			},
			CreatedAt: version.CreatedAt,
			UpdatedAt: version.UpdatedAt,
			Private:   true,
		})
		return res.Error
	} else if err != nil {
		return err
	}
	pkg.Versions = append(pkg.Versions, version)
	return db.db.Save(pkg).Error
}

func (db *UnpubLocalDb) AddUploader(name, email string) error {
	pkg, err := db.QueryPackage(name)
	if err != nil {
		return err
	}
	// pkg.Uploaders = append(pkg.Uploaders, email)
	pkg.Uploaders = email
	return db.db.Save(pkg).Error
}

func (db *UnpubLocalDb) RemoveUploader(name, email string) error {
	pkg, err := db.QueryPackage(name)
	if err != nil {
		return err
	}
	// var newUploaders []string
	// for _, uploader := range pkg.Uploaders {
	// 	if uploader != email {
	// 		newUploaders = append(newUploaders, uploader)
	// 	}
	// }
	// pkg.Uploaders = newUploaders
	return db.db.Save(pkg).Error
}

func (db *UnpubLocalDb) IncreaseDownloads(name, version string) error {
	pkg, err := db.QueryPackage(name)
	if err != nil {
		return err
	}
	pkg.Downloads++
	return db.db.Save(pkg).Error
}

// Interface guard
var _ = (UnpubDb)(&UnpubLocalDb{})
