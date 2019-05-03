package naming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Naming struct represents a collection of directory names
// used on host system
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

// New function returns a fresh Naming struct with defined fields
func New(program, dist, pkg, version, archiveDir string) *Naming {
	return &Naming{
		Container: Container(program, dist, pkg, version),
		Image:     Image(program, dist),

		SourceDir:       SourceDir(),
		SourceParentDir: SourceParentDir(),

		ArchiveDir:        ArchiveDir(program, dist, archiveDir),
		ArchivePackageDir: ArchivePackageDir(program, dist, pkg, version, archiveDir),

		CacheDir: CacheDir(program, dist),

		BuildDir: BuildDir(program, dist, pkg, version),
	}
}

// Container function returns a standardized container name
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

// Image function returns a standardized image name
func Image(program, dist string) string {
	return fmt.Sprintf(
		"%s:%s",
		program,
		dist,
	)
}

// SourceDir function returns simply current directory
func SourceDir() string {
	return os.Getenv("PWD")
}

// SourceParentDir function returns parent of current directory
func SourceParentDir() string {
	return filepath.Dir(SourceDir())
}

// ArchiveDir function returns archive directory, but already with distribution,
// so it's not $HOME/deber, but for example $HOME/deber/unstable
func ArchiveDir(program, dist, dir string) string {
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

// ArchivePackageDir function returns package directory in archive
func ArchivePackageDir(program, dist, pkg, version, dir string) string {
	return fmt.Sprintf(
		"%s/%s_%s",
		ArchiveDir(program, dist, dir),
		pkg,
		version,
	)
}

// CacheDir function returns image's cache directory on host system
func CacheDir(program, dist string) string {
	return filepath.Join("/tmp", Image(program, dist))
}

// BuildDir function returns container's build directory on host system
func BuildDir(program, dist, pkg, version string) string {
	return filepath.Join("/tmp", Container(program, dist, pkg, version))
}
