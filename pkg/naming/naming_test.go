package naming_test

import (
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

const (
	program = "deber"
	dist    = "buster"
	pkg     = "some-package"
	version = "2:1.0.0-1~bpo10+2"

	image     = "deber:buster"
	container = "deber_buster_some-package_2-1.0.0-1-bpo10-2"

	home = "/home/user23001"
)

var test = &naming.Naming{
	Container: container,
	Image:     image,

	SourceDir:         os.Getenv("PWD"),
	SourceParentDir:   filepath.Dir(os.Getenv("PWD")),
	CacheDir:          filepath.Join("/tmp", image),
	BuildDir:          filepath.Join("/tmp", container),
	ArchiveDir:        filepath.Join(home, program, dist),
	ArchivePackageDir: filepath.Join(home, program, dist, pkg+"_"+version),
}

func TestNew(t *testing.T) {
	name := naming.New(program, dist, pkg, version, home)

	assert.Equal(t, test, name)
}
