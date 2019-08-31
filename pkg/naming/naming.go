// Package naming includes various naming nuances
package naming

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// ContainerArchiveDir constant represents where on container will
	// archive directory be mounted
	ContainerArchiveDir = "/archive"
	// ContainerBuildDir constant represents where on container will
	// build directory be mounted
	ContainerBuildDir = "/build"
	// ContainerSourceDir constant represents where on container will
	// source directory be mounted
	ContainerSourceDir = "/build/source"
	// ContainerCacheDir constant represents where on container will
	// cache directory be mounted
	ContainerCacheDir = "/var/cache/apt"
)

// Naming struct holds various information naming information
// about package, container, image, directories
type Naming struct {
	// Args embedded here for quick reference
	Args

	// Container name
	Container string
	// Image name
	Image string

	// SourceDir is an absolute path where source lives
	SourceDir string
	// SourceParentDir is an absolute path where orig upstream tarball lives
	SourceParentDir string
	// BuildDir is an absolute path where build artifacts are stored
	BuildDir string
	// CacheDir is an absolute path where apt cache is stored
	CacheDir string
	// ArchiveDir is an absolute path where
	// all built packages are stored
	ArchiveDir string
	// ArchiveTargetDir is an absolute path where
	// all built packages for given target are stored
	ArchiveTargetDir string
	// ArchiveSourceDir is an absolute path where
	// all built packages for given source are stored
	ArchiveSourceDir string
	// ArchiveVersionDir is an absolute path where
	// all built packages for given source version are stored
	ArchiveVersionDir string
}

// Args struct holds information about package base directories and prefix
type Args struct {
	// Prefix is the program name
	Prefix string

	// Source is the name of source package
	Source string
	// Version is the version of source package
	Version string
	// Upstream is the upstream version of source package
	Upstream string
	// Target is the target distribution the package is building for
	Target string

	// SourceBaseDir is a directory where source lives
	SourceBaseDir string
	// BuildBaseDir is a directory where all build dirs are stored
	BuildBaseDir string
	// CacheBaseDir is a directory where all cache dirs are stored
	CacheBaseDir string
	// ArchiveBaseDir is a directory where all build packages are stored
	ArchiveBaseDir string
}

// New creates new instance of Naming struct
func New(args Args) *Naming {
	args.Target = standardizeTarget(args.Version, args.Target)

	version := standardizeVersion(args.Version)
	image := fmt.Sprintf("%s:%s", args.Prefix, args.Target)
	container := fmt.Sprintf("%s_%s_%s_%s", args.Prefix, args.Target, args.Source, version)

	return &Naming{
		Args: args,

		Container: container,
		Image:     image,

		SourceDir:         args.SourceBaseDir,
		SourceParentDir:   filepath.Dir(args.SourceBaseDir),
		BuildDir:          filepath.Join(args.BuildBaseDir, container),
		CacheDir:          filepath.Join(args.CacheBaseDir, image),
		ArchiveDir:        args.ArchiveBaseDir,
		ArchiveTargetDir:  filepath.Join(args.ArchiveBaseDir, args.Target),
		ArchiveSourceDir:  filepath.Join(args.ArchiveBaseDir, args.Target, args.Source),
		ArchiveVersionDir: filepath.Join(args.ArchiveBaseDir, args.Target, args.Source, args.Version),
	}
}

func standardizeVersion(version string) string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func standardizeTarget(version, target string) string {
	// UNRELEASED == unstable
	target = strings.Replace(target, "UNRELEASED", "unstable", -1)
	target = strings.Split(target, "-")[0]

	// Debian backport
	if strings.Contains(version, "bpo") {
		target = target + "-backports"
	}

	return target
}
