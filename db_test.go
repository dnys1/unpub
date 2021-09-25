package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	uploader    = "test@example.com"
	packageName = "my_pkg"
)

func TestDB(t *testing.T) {
	assert := assert.New(t)
	db, err := NewUnpubLocalDb(true, "")
	assert.NoError(err)
	assert.NotNil(db)

	pkg := &UnpubPackage{
		Name:      packageName,
		Uploaders: uploader,
		Versions: []UnpubVersion{
			{
				Version: "0.0.1",
				PubspecYAML: fmt.Sprintf(`
				name: %s
				version: 0.0.1
				description: My package
				`, packageName),
				UnpubPackageName: packageName,
			},
		},
		Private:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.AddVersion(packageName, pkg.Versions[0])
	assert.NoError(err)

	getPkg, err := db.QueryPackage(packageName)
	assert.NoError(err)
	assert.Truef(reflect.DeepEqual(pkg, getPkg), "Want: %+v\nGot: %+v", pkg, getPkg)
}
