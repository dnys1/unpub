package unpub

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	uploader    = "test@example.com"
	packageName = "my_pkg"
)

func TestDB(t *testing.T) {
	assert := assert.New(t)
	db, err := NewUnpubBadgerDb(true, "")
	assert.NoError(err)
	assert.NotNil(db)

	pkg := NewPackage(
		packageName,
		false,
		[]string{uploader},
	)
	pkg.CreateVersion(
		"0.0.1",
		fmt.Sprintf(`
name: %s
version: 0.0.1
description: My package`, packageName),
		nil, nil, nil,
	)
	err = db.SavePackage(pkg)
	assert.NoError(err)

	getPkg, err := db.QueryPackage(packageName)
	assert.NoError(err)
	assert.Truef(reflect.DeepEqual(pkg, getPkg), "Want: %+v\nGot: %+v", pkg, getPkg)
}
