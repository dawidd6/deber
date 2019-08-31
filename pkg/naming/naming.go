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

type Naming struct {
	Args

	Container string
	Image     string

	SourceDir         string
	SourceParentDir   string
	BuildDir          string
	CacheDir          string
	ArchiveDir        string
	ArchiveTargetDir  string
	ArchiveSourceDir  string
	ArchiveVersionDir string
}

type Args struct {
	Prefix string

	Source   string
	Version  string
	Upstream string
	Target   string

	SourceBaseDir  string
	BuildBaseDir   string
	CacheBaseDir   string
	ArchiveBaseDir string
}

func New(args Args) *Naming {
	stdVersion := standardizeVersion(args.Version)
	stdTarget := standardizeTarget(args.Version, args.Target)

	image := fmt.Sprintf("%s:%s", args.Prefix, stdTarget)
	container := fmt.Sprintf("%s_%s_%s_%s", args.Prefix, stdTarget, args.Source, stdVersion)

	return &Naming{
		Args: args,

		Container: container,
		Image:     image,

		SourceDir:         args.SourceBaseDir,
		SourceParentDir:   filepath.Dir(args.SourceBaseDir),
		BuildDir:          filepath.Join(args.BuildBaseDir, container),
		CacheDir:          filepath.Join(args.CacheBaseDir, image),
		ArchiveDir:        args.ArchiveBaseDir,
		ArchiveTargetDir:  filepath.Join(args.ArchiveBaseDir, stdTarget),
		ArchiveSourceDir:  filepath.Join(args.ArchiveBaseDir, stdTarget, args.Source),
		ArchiveVersionDir: filepath.Join(args.ArchiveBaseDir, stdTarget, args.Source, args.Version),
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
