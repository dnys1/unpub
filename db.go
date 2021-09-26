package unpub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v3"
)

type UnpubDbQuery struct {
	Size       int
	Page       int
	Sort       string
	Keyword    string
	Uploader   string
	Dependency string
}

type UnpubDb interface {
	QueryPackage(name string) (UnpubPackage, error)
	QueryPackages(query UnpubDbQuery) (*UnpubQueryResult, error)
	SavePackage(pkg UnpubPackage) error
	AddUploader(name, email string) error
	RemoveUploader(name, email string) error
	IncreaseDownloads(name, version string) error
	SaveFile(pkgName, version string, data []byte) error
	GetFile(pkgName, version string) (io.Reader, error)
}

type UnpubLocalDb struct {
	InMemory bool
	Path     string
	db       *badger.DB
}

func NewUnpubLocalDb(inMem bool, path string) (*UnpubLocalDb, error) {
	var dbPath string
	if !inMem {
		dbPath = filepath.Join(path, "db")
	}
	badgerDb, err := badger.Open(
		badger.
			DefaultOptions(dbPath).
			WithInMemory(inMem),
	)
	if err != nil {
		return nil, err
	}
	dbLoc := dbPath
	if inMem {
		dbLoc = "memory"
	}
	log.Printf("Opened DB at: %s", dbLoc)
	return &UnpubLocalDb{
		InMemory: inMem,
		Path:     dbPath,
		db:       badgerDb,
	}, nil
}

const (
	packagePrefix = "package_"
	filePrefix    = "file_"
)

func makePackageKey(packageName string) []byte {
	return []byte(fmt.Sprintf("%s%s", packagePrefix, packageName))
}

func makeFileKey(packageName, version string) []byte {
	return []byte(fmt.Sprintf("%s%s_%s", filePrefix, packageName, version))
}

func (db *UnpubLocalDb) Close() error {
	return db.db.Close()
}

func (db *UnpubLocalDb) QueryPackage(name string) (pkg UnpubPackage, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		key := makePackageKey(name)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &pkg)
		})
	})
	return
}

func (db *UnpubLocalDb) QueryPackages(query UnpubDbQuery) (*UnpubQueryResult, error) {
	var packages []*UnpubPackage
	err := db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(packagePrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := strings.TrimPrefix(string(item.Key()), packagePrefix)
			if query.Keyword == "" || strings.Contains(key, query.Keyword) {
				var pkg UnpubPackage
				err := item.Value(func(val []byte) error {
					return json.Unmarshal(val, &pkg)
				})
				if err != nil {
					return err
				}
				packages = append(packages, &pkg)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &UnpubQueryResult{
		Count:    len(packages),
		Packages: packages,
	}, nil
}

func (db *UnpubLocalDb) SavePackage(pkg UnpubPackage) (err error) {
	return db.db.Update(func(txn *badger.Txn) error {
		b, err := json.Marshal(pkg)
		if err != nil {
			return err
		}
		return txn.Set(makePackageKey(pkg.Name), b)
	})
}

func (db *UnpubLocalDb) AddUploader(name, email string) error {
	pkg, err := db.QueryPackage(name)
	if err != nil {
		return err
	}
	var contains bool
	for _, uploader := range pkg.Uploaders {
		if uploader == email {
			contains = true
		}
	}
	if !contains {
		pkg.Uploaders = append(pkg.Uploaders, email)
		return db.SavePackage(pkg)
	} else {
		return errors.New("uploader already exists")
	}
}

func (db *UnpubLocalDb) RemoveUploader(name, email string) error {
	pkg, err := db.QueryPackage(name)
	if err != nil {
		return err
	}
	var newUploaders []string
	for _, uploader := range pkg.Uploaders {
		if uploader != email {
			newUploaders = append(newUploaders, email)
		}
	}
	if len(newUploaders) == len(pkg.Uploaders) {
		return errors.New("uploader does not exist")
	}
	return db.SavePackage(pkg)
}

func (db *UnpubLocalDb) IncreaseDownloads(name, version string) error {
	pkg, err := db.QueryPackage(name)
	if err != nil {
		return err
	}
	pkg.Downloads++
	return db.SavePackage(pkg)
}

func (db *UnpubLocalDb) SaveFile(pkgName, version string, data []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(makeFileKey(pkgName, version), data)
	})
}

func (db *UnpubLocalDb) GetFile(pkgName, version string) (io.Reader, error) {
	var data []byte
	err := db.db.View(func(txn *badger.Txn) error {
		key := makeFileKey(pkgName, version)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		data, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), err
}

// Interface guard
var _ = (UnpubDb)(&UnpubLocalDb{})
