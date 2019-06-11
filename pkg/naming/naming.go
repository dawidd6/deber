package naming

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/debian"
	"os"
	"path/filepath"
	"strings"
)

var (
	ArchiveBaseDir = filepath.Join(os.Getenv("HOME"), app.Name)
	BuildBaseDir   = "/tmp"
	CacheBaseDir   = "/tmp"
	SourceBaseDir  = os.Getenv("PWD")
)

type Naming struct {
	deb *debian.Debian
}

func New(deb *debian.Debian) *Naming {
	return &Naming{
		deb: deb,
	}
}

func (n *Naming) ImageName() string {
	return fmt.Sprintf(
		"%s:%s",
		app.Name,
		n.standardizeImageTag(),
	)
}

func (n *Naming) ImageRepo() string {
	return app.Name
}

func (n *Naming) ImageTag() string {
	return n.standardizeImageTag()
}

func (n *Naming) ContainerName() string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		app.Name,
		n.deb.Target,
		n.deb.Source,
		n.standardizePackageVersion(),
	)
}

func (n *Naming) standardizePackageVersion() string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version := n.deb.Version.Package
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func (n *Naming) standardizeImageTag() string {
	if strings.Contains(n.deb.Version.Package, "bpo") {
		if strings.Contains(n.deb.Target, "backports") {
			return n.deb.Target
		}

		if n.deb.Target == "UNRELEASED" {
			return "unstable"
		}
	}

	if strings.Contains(n.deb.Target, "-") {
		return strings.Split(n.deb.Target, "-")[0]
	}

	return n.deb.Target
}

func (n *Naming) BuildDir() string {
	return filepath.Join(
		BuildBaseDir,
		n.ContainerName(),
	)
}

func (n *Naming) CacheDir() string {
	return filepath.Join(
		CacheBaseDir,
		n.ImageName(),
	)
}

func (n *Naming) ArchiveTargetDir() string {
	return filepath.Join(
		ArchiveBaseDir,
		n.deb.Target,
	)
}

func (n *Naming) ArchiveSourceDir() string {
	return filepath.Join(
		n.ArchiveTargetDir(),
		n.deb.Source,
	)
}

func (n *Naming) ArchiveVersionDir() string {
	return filepath.Join(
		n.ArchiveSourceDir(),
		n.deb.Version.Package,
	)
}

func (n *Naming) SourceDir() string {
	return SourceBaseDir
}

func (n *Naming) SourceParentDir() string {
	return filepath.Dir(
		n.SourceDir(),
	)
}
