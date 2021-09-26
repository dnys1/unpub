package unpub

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

var (
	uploader    = "test@example.com"
	packageName = "my_pkg"
)

func TestDB(t *testing.T) {
	require := require.New(t)
	db, err := NewUnpubLocalDb(true, "")
	require.NoError(err)
	require.NotNil(db)

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
	require.NoError(err)

	getPkg, err := db.QueryPackage(packageName)
	require.NoError(err)
	require.Truef(cmp.Equal(pkg, getPkg), "Want: %+v\nGot: %+v", pkg, getPkg)
}
