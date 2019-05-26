package debian_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dawidd6/deber/pkg/debian"
)

func TestNew1(t *testing.T) {
	line := "golang-github-alcortesm-tgz (0.0~git20161220.9c5fe88-1) unstable; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "golang-github-alcortesm-tgz",
		PackageVersion:  "0.0~git20161220.9c5fe88-1",
		UpstreamVersion: "0.0~git20161220.9c5fe88",
		TargetDist:      "unstable",
		IsNative:        false,
	}

	assert.Equal(t, expected, got)
}

func TestNew2(t *testing.T) {
	line := "ansible (2.7.5+dfsg-2) experimental; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "ansible",
		PackageVersion:  "2.7.5+dfsg-2",
		UpstreamVersion: "2.7.5+dfsg",
		TargetDist:      "experimental",
		IsNative:        false,
	}

	assert.Equal(t, expected, got)
}

func TestNew3(t *testing.T) {
	line := "ansible (2.2.1.0-2+deb9u1) stretch-security; urgency=high"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "ansible",
		PackageVersion:  "2.2.1.0-2+deb9u1",
		UpstreamVersion: "2.2.1.0",
		TargetDist:      "stretch",
		IsNative:        false,
	}

	assert.Equal(t, expected, got)
}

func TestNew4(t *testing.T) {
	line := "procps (2:3.3.10-4ubuntu2.4) xenial-security; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "procps",
		PackageVersion:  "2:3.3.10-4ubuntu2.4",
		UpstreamVersion: "3.3.10",
		TargetDist:      "xenial",
		IsNative:        false,
	}

	assert.Equal(t, expected, got)
}

func TestNew5(t *testing.T) {
	line := "git-buildpackage (0.9.14) unstable; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "git-buildpackage",
		PackageVersion:  "0.9.14",
		UpstreamVersion: "0.9.14",
		TargetDist:      "unstable",
		IsNative:        true,
	}

	assert.Equal(t, expected, got)
}

func TestNew6(t *testing.T) {
	line := "ansible (2.7.5+dfsg-1~bpo9+1) stretch-backports; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "ansible",
		PackageVersion:  "2.7.5+dfsg-1~bpo9+1",
		UpstreamVersion: "2.7.5+dfsg",
		TargetDist:      "stretch-backports",
		IsNative:        false,
	}

	assert.Equal(t, expected, got)
}

func TestNew7(t *testing.T) {
	line := "procps (2:3.3.10-4ubuntu2.4) xenial-backports; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "procps",
		PackageVersion:  "2:3.3.10-4ubuntu2.4",
		UpstreamVersion: "3.3.10",
		TargetDist:      "xenial",
		IsNative:        false,
	}

	assert.Equal(t, expected, got)
}

func TestNew8(t *testing.T) {
	line := "git-buildpackage (1.0.0) UNRELEASED; urgency=medium"

	got := debian.New(line)
	expected := &debian.Debian{
		SourceName:      "git-buildpackage",
		PackageVersion:  "1.0.0",
		UpstreamVersion: "1.0.0",
		TargetDist:      "unstable",
		IsNative:        true,
	}

	assert.Equal(t, expected, got)
}

func TestParseChangelog(t *testing.T) {
	line := "git-buildpackage (1.0.0) UNRELEASED; urgency=medium\n"

	err := os.Mkdir("debian", os.ModePerm)
	assert.NoError(t, err)

	err = ioutil.WriteFile("debian/changelog", []byte(line), os.ModePerm)
	assert.NoError(t, err)

	got, err := debian.ParseChangelog()
	expected := &debian.Debian{
		SourceName:      "git-buildpackage",
		PackageVersion:  "1.0.0",
		UpstreamVersion: "1.0.0",
		TargetDist:      "unstable",
		IsNative:        true,
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	assert.NoError(t, os.RemoveAll("debian"))
}

func TestLocateTarball1(t *testing.T) {
	line := "git-buildpackage (1.0.0) UNRELEASED; urgency=medium"
	got := debian.New(line)

	tarball, err := got.LocateTarball("")

	assert.Equal(t, "", tarball)
	assert.NoError(t, err)
}

func TestLocateTarball2(t *testing.T) {
	line := "ansible (2.7.5+dfsg-1~bpo9+1) stretch-backports; urgency=medium"
	got := debian.New(line)

	tarball, err := got.LocateTarball("")

	assert.Equal(t, "", tarball)
	assert.Error(t, err)
}

func TestLocateTarball3(t *testing.T) {
	line := "ansible (2.7.5+dfsg-1~bpo9+1) stretch-backports; urgency=medium"
	got := debian.New(line)

	dir := filepath.Dir(os.Getenv("PWD"))
	fileName := fmt.Sprintf("%s_%s.orig.tar.xz", got.SourceName, got.UpstreamVersion)
	filePath := filepath.Join(dir, fileName)

	file, err := os.Create(filePath)

	assert.NoError(t, err)

	tarball, err := got.LocateTarball(dir)

	assert.Equal(t, fileName, tarball)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())
	assert.NoError(t, os.Remove(filePath))
}
