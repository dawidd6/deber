package naming

import (
	"fmt"
	"os"
	"strings"
)

const (
	ContainerArchiveFromDir = "/archive"
	ContainerBuildOutputDir = "/build"
	ContainerSourceInputDir = "/build/source"
	ContainerBuildCacheDir  = "/var/cache/apt"
)

type Naming struct {
	program string
	pkg     string
	dist    string
	version string
	tarball string
}

func New(program, dist, pkg, version, tarball string) *Naming {
	return &Naming{
		program: program,
		pkg:     pkg,
		dist:    dist,
		version: version,
		tarball: tarball,
	}
}

func (n *Naming) Container() string {
	version := n.version

	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian versioning allows these characters
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return fmt.Sprintf(
		"%s_%s_%s_%s",
		n.program,
		n.dist,
		n.pkg,
		version,
	)
}

func (n *Naming) Image() string {
	return fmt.Sprintf(
		"%s:%s",
		n.program,
		n.dist,
	)
}

func (n *Naming) Tarball() string {
	return n.tarball
}

func (n *Naming) Dist() string {
	return n.dist
}

// SOURCE
func (n *Naming) HostSourceDir() string {
	return fmt.Sprintf(
		"%s/..",
		os.Getenv("PWD"),
	)
}

func (n *Naming) HostSourceInputDir() string {
	return os.Getenv("PWD")
}

func (n *Naming) HostSourceTarballDir() string {
	return n.HostSourceDir()
}

func (n *Naming) HostSourceSourceTarballFile() string {
	return fmt.Sprintf(
		"%s/%s",
		n.HostSourceTarballDir(),
		n.tarball,
	)
}

// ARCHIVE
func (n *Naming) HostArchiveDir() string {
	dir := os.Getenv("DEBER_ARCHIVE")
	if dir != "" {
		return dir
	}

	return fmt.Sprintf(
		"%s/%s",
		os.Getenv("HOME"),
		n.program,
	)
}

func (n *Naming) HostArchiveFromDir() string {
	return fmt.Sprintf(
		"%s/%s",
		n.HostArchiveDir(),
		n.dist,
	)
}

func (n *Naming) HostArchiveFromOutputDir() string {
	return fmt.Sprintf(
		"%s/%s_%s",
		n.HostArchiveFromDir(),
		n.pkg,
		n.version,
	)
}

// BUILD
func (n *Naming) HostBuildDir() string {
	return "/tmp"
}

func (n *Naming) HostBuildCacheDir() string {
	return fmt.Sprintf(
		"%s/%s",
		n.HostBuildDir(),
		n.Image(),
	)
}

func (n *Naming) HostBuildOutputDir() string {
	return fmt.Sprintf(
		"%s/%s",
		n.HostBuildDir(),
		n.Container(),
	)
}

func (n *Naming) HostBuildTargetTarballDir() string {
	return n.HostBuildOutputDir()
}

func (n *Naming) HostBuildTargetTarballFile() string {
	return fmt.Sprintf(
		"%s/%s",
		n.HostBuildTargetTarballDir(),
		n.tarball,
	)
}
