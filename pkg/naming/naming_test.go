package naming_test

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"pault.ag/go/debian/version"
	"testing"
)

func TestNew(t *testing.T) {
	debian := &changelog.ChangelogEntry{
		Source: "blah",
		Target: "buster",
		Version: version.Version{
			Version: "1.0.0",
		},
	}

	pkg := &naming.Package{
		ChangelogEntry: debian,
	}

	container := &naming.Container{
		Package: pkg,
	}

	image := &naming.Image{
		Package: pkg,
	}

	dirs := &naming.Directories{
		Source: &naming.Source{
			Base: os.Getenv("PWD"),
		}, Build: &naming.Build{
			Base:      "/tmp",
			Container: container,
		}, Cache: &naming.Cache{
			Base:  "/tmp",
			Image: image,
		}, Archive: &naming.Archive{
			Base:    filepath.Join(os.Getenv("HOME"), app.Name),
			Package: pkg,
		},
	}

	name := naming.New(debian)

	assert.Equal(
		t,
		pkg,
		name.Package,
	)

	assert.Equal(
		t,
		container,
		name.Container,
	)

	assert.Equal(
		t,
		image,
		name.Image,
	)

	assert.Equal(
		t,
		dirs,
		name.Directories,
	)
}

func TestArchive(t *testing.T) {
	base := "/home/user212/deb"

	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "buster",
			Version: version.Version{
				Version: "1.0.0",
			},
		},
	}

	archive := &naming.Archive{
		Base:    base,
		Package: pkg,
	}

	assert.Equal(
		t,
		filepath.Join(base, pkg.Target),
		archive.PackageTargetPath(),
	)

	assert.Equal(
		t,
		filepath.Join(base, pkg.Target, pkg.Source),
		archive.PackageSourcePath(),
	)

	assert.Equal(
		t,
		filepath.Join(base, pkg.Target, pkg.Source, pkg.Version.String()),
		archive.PackageVersionPath(),
	)
}

func TestCache(t *testing.T) {
	base := "/tmp"

	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "buster",
			Version: version.Version{
				Version: "1.0.0",
			},
		},
	}

	image := &naming.Image{
		Package: pkg,
	}

	cache := &naming.Cache{
		Base:  base,
		Image: image,
	}

	assert.Equal(
		t,
		filepath.Join(base, image.Name()),
		cache.ImagePath(),
	)
}

func TestBuild(t *testing.T) {
	base := "/tmp"

	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "buster",
			Version: version.Version{
				Version: "1.0.0",
			},
		},
	}

	container := &naming.Container{
		Package: pkg,
	}

	build := &naming.Build{
		Base:      base,
		Container: container,
	}

	assert.Equal(
		t,
		filepath.Join(base, container.Name()),
		build.ContainerPath(),
	)
}

func TestSource(t *testing.T) {
	parent := "/home/user231/debian"
	base := filepath.Join(parent, "source-name")

	source := &naming.Source{
		Base: base,
	}

	assert.Equal(
		t,
		base,
		source.SourcePath(),
	)

	assert.Equal(
		t,
		parent,
		source.ParentPath(),
	)
}

func TestImage(t *testing.T) {
	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "buster",
			Version: version.Version{
				Version: "1.0.0",
			},
		},
	}

	image := &naming.Image{
		Package: pkg,
	}

	assert.Equal(
		t,
		fmt.Sprintf("%s:%s", app.Name, "buster"),
		image.Name(),
	)
}

func TestImageUbuntuBackport(t *testing.T) {
	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "bionic-backports",
			Version: version.Version{
				Epoch:    2,
				Version:  "1.0.0",
				Revision: "1~ubuntu18.04.1",
			},
		},
	}

	image := &naming.Image{
		Package: pkg,
	}

	assert.Equal(
		t,
		fmt.Sprintf("%s:%s", app.Name, "bionic"),
		image.Name(),
	)
}

func TestImageDebianBackport(t *testing.T) {
	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "buster-backports",
			Version: version.Version{
				Version:  "1.0.0",
				Revision: "1~bpo10+1",
			},
		},
	}

	image := &naming.Image{
		Package: pkg,
	}

	assert.Equal(
		t,
		fmt.Sprintf("%s:%s", app.Name, "buster-backports"),
		image.Name(),
	)
}

func TestImageUnreleased(t *testing.T) {
	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "UNRELEASED",
			Version: version.Version{
				Version:  "1.0.0",
				Revision: "1~bpo10+1",
			},
		},
	}

	image := &naming.Image{
		Package: pkg,
	}

	assert.Equal(
		t,
		fmt.Sprintf("%s:%s", app.Name, "unstable"),
		image.Name(),
	)
}

func TestContainer(t *testing.T) {
	pkg := &naming.Package{
		ChangelogEntry: &changelog.ChangelogEntry{
			Source: "blah",
			Target: "buster-backports",
			Version: version.Version{
				Epoch:    2,
				Version:  "1.0.0",
				Revision: "1~bpo10+1",
			},
		},
	}

	container := &naming.Container{
		Package: pkg,
	}

	assert.Equal(
		t,
		fmt.Sprintf(
			"%s_%s_%s_%s",
			app.Name,
			pkg.Target,
			pkg.Source,
			"2-1.0.0-1-bpo10-1",
		),
		container.Name(),
	)
}
