package naming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Directories in container where their host counterpart should be mounted.
const (
	ContainerArchiveDir = "/archive"
	ContainerBuildDir   = "/build"
	ContainerSourceDir  = "/build/source"
	ContainerCacheDir   = "/var/cache/apt"
)

type Naming struct {
	// Docker container name.
	//
	// Example: deber_unstable_wget_1.0.0-1
	Container string
	// Docker image name.
	//
	// Example: deber:unstable
	Image string

	// Current working directory, where debianized source lives.
	SourceDir string
	// Parent of current working directory.
	//
	// Used for locating orig upstream tarball mostly (if only).
	SourceParentDir string

	// Directory where built packages for a specific distribution live.
	//
	// Example: /home/user/deber/unstable
	ArchiveDir string
	// Specific directory of package for a distribution in archive.
	//
	// Example: /home/user/deber/unstable/wget_1.0.0-1
	ArchivePackageDir string

	// Directory where image's apt cache is stored
	//
	// Example: /tmp/deber:unstable
	CacheDir string

	// Directory where package building output is gathered.
	//
	// Example: /tmp/deber_unstable_wget_1.0.0-1
	BuildDir string
}

func New(program, dist, pkg, version string) *Naming {
	return &Naming{
		Container: Container(program, dist, pkg, version),
		Image:     Image(program, dist),

		SourceDir:       SourceDir(),
		SourceParentDir: SourceParentDir(),

		ArchiveDir:        ArchiveDir(program, dist),
		ArchivePackageDir: ArchivePackageDir(program, dist, pkg, version),

		CacheDir: CacheDir(program, dist),

		BuildDir: BuildDir(program, dist, pkg, version),
	}
}

func Container(program, dist, pkg, version string) string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian versioning allows these characters
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return fmt.Sprintf(
		"%s_%s_%s_%s",
		program,
		dist,
		pkg,
		version,
	)
}

func Image(program, dist string) string {
	return fmt.Sprintf(
		"%s:%s",
		program,
		dist,
	)
}

// SOURCE
func SourceDir() string {
	return os.Getenv("PWD")
}

func SourceParentDir() string {
	return filepath.Dir(SourceDir())
}

// ARCHIVE
func ArchiveDir(program, dist string) string {
	dir := os.Getenv("DEBER_ARCHIVE")
	if dir == "" {
		dir = os.Getenv("HOME")
	}

	return fmt.Sprintf(
		"%s/%s/%s",
		dir,
		program,
		dist,
	)
}

func ArchivePackageDir(program, dist, pkg, version string) string {
	return fmt.Sprintf(
		"%s/%s_%s",
		ArchiveDir(program, dist),
		pkg,
		version,
	)
}

// CACHE
func CacheDir(program, dist string) string {
	return filepath.Join("/tmp", Image(program, dist))
}

// BUILD
func BuildDir(program, dist, pkg, version string) string {
	return filepath.Join("/tmp", Container(program, dist, pkg, version))
}
