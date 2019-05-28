package naming_test

import (
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

const program = "deber"
const pkg = "some-package"

func TestNew1(t *testing.T) {
	actual := naming.New("buster", pkg, "2:1.0.0-1")
	expected := &naming.Naming{
		Container:      "deber_buster_some-package_2-1.0.0-1",
		Image:          "deber:buster",
		Distribution:   "buster",
		PackageName:    pkg,
		PackageVersion: "2:1.0.0-1",

		SourceDir:         os.Getenv("PWD"),
		SourceParentDir:   filepath.Dir(os.Getenv("PWD")),
		CacheDir:          "/tmp/deber:buster",
		BuildDir:          "/tmp/deber_buster_some-package_2-1.0.0-1",
		ArchiveDir:        filepath.Join(os.Getenv("HOME"), program, "buster"),
		ArchivePackageDir: filepath.Join(os.Getenv("HOME"), program, "buster", pkg, "2:1.0.0-1"),
	}

	assert.Equal(t, expected, actual)
}

func TestNew2(t *testing.T) {
	actual := naming.New("bionic-backports", pkg, "2:1.0.0-1")
	expected := &naming.Naming{
		Container:      "deber_bionic_some-package_2-1.0.0-1",
		Image:          "deber:bionic",
		Distribution:   "bionic",
		PackageName:    pkg,
		PackageVersion: "2:1.0.0-1",

		SourceDir:         os.Getenv("PWD"),
		SourceParentDir:   filepath.Dir(os.Getenv("PWD")),
		CacheDir:          "/tmp/deber:bionic",
		BuildDir:          "/tmp/deber_bionic_some-package_2-1.0.0-1",
		ArchiveDir:        filepath.Join(os.Getenv("HOME"), program, "bionic"),
		ArchivePackageDir: filepath.Join(os.Getenv("HOME"), program, "bionic", pkg, "2:1.0.0-1"),
	}

	assert.Equal(t, expected, actual)
}

func TestNew3(t *testing.T) {
	actual := naming.New("UNRELEASED", pkg, "2:1.0.0-1")
	expected := &naming.Naming{
		Container:      "deber_unstable_some-package_2-1.0.0-1",
		Image:          "deber:unstable",
		Distribution:   "unstable",
		PackageName:    pkg,
		PackageVersion: "2:1.0.0-1",

		SourceDir:         os.Getenv("PWD"),
		SourceParentDir:   filepath.Dir(os.Getenv("PWD")),
		CacheDir:          "/tmp/deber:unstable",
		BuildDir:          "/tmp/deber_unstable_some-package_2-1.0.0-1",
		ArchiveDir:        filepath.Join(os.Getenv("HOME"), program, "unstable"),
		ArchivePackageDir: filepath.Join(os.Getenv("HOME"), program, "unstable", pkg, "2:1.0.0-1"),
	}

	assert.Equal(t, expected, actual)
}

func TestNew4(t *testing.T) {
	actual := naming.New("buster-backports", pkg, "2:1.0.0-1~bpo10+1")
	expected := &naming.Naming{
		Container:      "deber_buster-backports_some-package_2-1.0.0-1-bpo10-1",
		Image:          "deber:buster-backports",
		Distribution:   "buster-backports",
		PackageName:    pkg,
		PackageVersion: "2:1.0.0-1~bpo10+1",

		SourceDir:         os.Getenv("PWD"),
		SourceParentDir:   filepath.Dir(os.Getenv("PWD")),
		CacheDir:          "/tmp/deber:buster-backports",
		BuildDir:          "/tmp/deber_buster-backports_some-package_2-1.0.0-1-bpo10-1",
		ArchiveDir:        filepath.Join(os.Getenv("HOME"), program, "buster-backports"),
		ArchivePackageDir: filepath.Join(os.Getenv("HOME"), program, "buster-backports", pkg, "2:1.0.0-1~bpo10+1"),
	}

	assert.Equal(t, expected, actual)
}
