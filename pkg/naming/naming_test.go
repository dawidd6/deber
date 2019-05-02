package naming_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dawidd6/deber/pkg/naming"
)

const (
	program = "deber"
	dist    = "buster"
	pkg     = "some-package"
	version = "2:1.0.0-1~bpo10+2"

	image     = "deber:buster"
	container = "deber_buster_some-package_2-1.0.0-1-bpo10-2"

	home = "/home/user23001"
	cwd  = "/home/user23001/packaging/wget"
	env  = "/var/lib/deber"
)

func test(t *testing.T, got, expected string) {
	if got != expected {
		t.Log("got:", got)
		t.Log("expected:", expected)
		t.Fatal("got != expected")
	}
}

func setEnv(t *testing.T, key, newValue string) (string, string) {
	oldValue := os.Getenv(key)

	err := os.Setenv(key, newValue)
	if err != nil {
		t.Fatal(err)
	}

	return key, oldValue
}

func restoreEnv(t *testing.T, key, oldValue string) {
	err := os.Setenv(key, oldValue)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSourceDir(t *testing.T) {
	key, val := setEnv(t, "PWD", cwd)
	defer restoreEnv(t, key, val)

	got := naming.SourceDir()
	expected := cwd

	test(t, got, expected)
}

func TestSourceParentDir(t *testing.T) {
	key, val := setEnv(t, "PWD", cwd)
	defer restoreEnv(t, key, val)

	got := naming.SourceParentDir()
	expected := filepath.Dir(cwd)

	test(t, got, expected)
}

func TestBuildDir(t *testing.T) {
	got := naming.BuildDir(program, dist, pkg, version)
	expected := fmt.Sprintf("/tmp/%s", container)

	test(t, got, expected)
}

func TestCacheDir(t *testing.T) {
	got := naming.CacheDir(program, dist)
	expected := fmt.Sprintf("/tmp/%s", image)

	test(t, got, expected)
}

func TestArchiveDir(t *testing.T) {
	key, val := setEnv(t, "HOME", home)
	defer restoreEnv(t, key, val)

	got := naming.ArchiveDir(program, dist, "")
	expected := fmt.Sprintf("%s/%s/%s", home, program, dist)

	test(t, got, expected)
}

func TestArchiveDirWithEnv(t *testing.T) {
	key, val := setEnv(t, "DEBER_ARCHIVE", env)
	defer restoreEnv(t, key, val)
	dir := os.Getenv("DEBER_ARCHIVE")

	got := naming.ArchiveDir(program, dist, dir)
	expected := fmt.Sprintf("%s/%s/%s", env, program, dist)

	test(t, got, expected)
}

func TestArchivePackageDir(t *testing.T) {
	key, val := setEnv(t, "HOME", home)
	defer restoreEnv(t, key, val)

	got := naming.ArchivePackageDir(program, dist, pkg, version, "")
	expected := fmt.Sprintf("%s/%s/%s/%s_%s", home, program, dist, pkg, version)

	test(t, got, expected)
}

func TestArchivePackageDirWithEnv(t *testing.T) {
	key, val := setEnv(t, "DEBER_ARCHIVE", env)
	defer restoreEnv(t, key, val)
	dir := os.Getenv("DEBER_ARCHIVE")

	got := naming.ArchivePackageDir(program, dist, pkg, version, dir)
	expected := fmt.Sprintf("%s/%s/%s/%s_%s", env, program, dist, pkg, version)

	test(t, got, expected)
}

func TestImage(t *testing.T) {
	got := naming.Image(program, dist)
	expected := image

	test(t, got, expected)
}

func TestContainer(t *testing.T) {
	got := naming.Container(program, dist, pkg, version)
	expected := container

	test(t, got, expected)
}

func TestNew(t *testing.T) {
	name := naming.New(program, dist, pkg, version, "")

	if name.Container != container {
		t.Fatal()
	}
	if name.Image != image {
		t.Fatal()
	}
}
