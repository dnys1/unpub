package unpub

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func makePkg() *UnpubPackage {
	return &UnpubPackage{
		Name: "example",
		Versions: map[string]UnpubVersion{
			"0.1.0": {
				Version: "0.1.0",
			},
		},
		Latest: "0.1.0",
	}
}

func TestUnpubPackageAddVersion(t *testing.T) {
	currVersion := UnpubVersion{
		Version: "0.1.0",
	}
	prevVersion := UnpubVersion{
		Version: "0.0.9",
	}
	patchVersion := UnpubVersion{
		Version: "0.1.1",
	}
	minorVersion := UnpubVersion{
		Version: "0.2.0",
	}
	majorVersion := UnpubVersion{
		Version: "1.0.0",
	}

	err := makePkg().AddVersion(currVersion)
	require.Error(t, err)

	err = makePkg().AddVersion(prevVersion)
	require.Error(t, err)

	err = makePkg().AddVersion(patchVersion)
	require.NoError(t, err)

	err = makePkg().AddVersion(minorVersion)
	require.NoError(t, err)

	err = makePkg().AddVersion(majorVersion)
	require.NoError(t, err)
}
