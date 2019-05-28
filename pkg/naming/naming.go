// Package naming standardizes some names used internally by deber.
package naming

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const program = "deber"

// Naming struct represents a collection of directory names
// used on host system.
type Naming struct {
	// Docker container name.
	//
	// Example: deber_unstable_wget_1.0.0-1
	Container string

	// Docker image name.
	//
	// Example: deber:unstable
	Image string

	// Debian distribution
	//
	// Example: unstable
	Distribution string

	// Debian package name
	//
	// Example: wget
	PackageName string

	// Debian package packageVersion
	//
	// Example: 1.0.0-1
	PackageVersion string

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

// New function returns a fresh Naming struct with defined fields.
func New(dist, packageName, packageVersion string) *Naming {
	return &Naming{
		Container: Container(dist, packageName, packageVersion),
		Image:     Image(dist, packageName, packageVersion),

		Distribution:   Distribution(dist, packageName, packageVersion),
		PackageName:    PackageName(dist, packageName, packageVersion),
		PackageVersion: PackageVersion(dist, packageName, packageVersion),

		SourceDir:       SourceDir(),
		SourceParentDir: SourceParentDir(),

		ArchiveDir:        ArchiveDir(dist, packageName, packageVersion),
		ArchivePackageDir: ArchivePackageDir(dist, packageName, packageVersion),

		CacheDir: CacheDir(dist, packageName, packageVersion),

		BuildDir: BuildDir(dist, packageName, packageVersion),
	}
}

// Container function returns a standardized container name.
func Container(dist, packageName, packageVersion string) string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian packageVersioning allows these characters
	packageVersion = strings.Replace(packageVersion, "~", "-", -1)
	packageVersion = strings.Replace(packageVersion, ":", "-", -1)
	packageVersion = strings.Replace(packageVersion, "+", "-", -1)

	return fmt.Sprintf(
		"%s_%s_%s_%s",
		program,
		Distribution(dist, packageName, packageVersion),
		PackageName(dist, packageName, packageVersion),
		PackageVersion(dist, packageName, packageVersion),
	)
}

// Image function returns a standardized image name.
func Image(dist, packageName, packageVersion string) string {
	return fmt.Sprintf(
		"%s:%s",
		program,
		Distribution(dist, packageName, packageVersion),
	)
}

// Distribution function returns standardized distribution name.
func Distribution(dist, packageName, packageVersion string) string {
	if strings.Contains(dist, "-") {
		dist = strings.Split(dist, "-")[0]
	}

	// Debian backport
	if strings.Contains(packageVersion, "bpo") {
		dist += "-backports"
	}

	if dist == "UNRELEASED" {
		dist = "unstable"
	}

	return dist
}

// PackageName function just returns passed package name.
func PackageName(dist, packageName, packageVersion string) string {
	return packageName
}

// PackageVersion function just returns passed package version.
func PackageVersion(dist, packageName, packageVersion string) string {
	return packageVersion
}

// SourceDir function returns simply current directory.
func SourceDir() string {
	return os.Getenv("PWD")
}

// SourceParentDir function returns parent of current directory.
func SourceParentDir() string {
	return filepath.Dir(SourceDir())
}

// ArchiveDir function returns archive directory, but already with distribution,
// so it's not $HOME/deber, but for example $HOME/deber/unstable.
func ArchiveDir(dist, packageName, packageVersion string) string {
	return filepath.Join(
		os.Getenv("HOME"),
		program,
		Distribution(dist, packageName, packageVersion),
	)
}

// ArchivePackageDir function returns package directory in archive.
func ArchivePackageDir(dist, packageName, packageVersion string) string {
	return filepath.Join(
		ArchiveDir(dist, packageName, packageVersion),
		PackageName(dist, packageName, packageVersion),
		PackageVersion(dist, packageName, packageVersion),
	)
}

// CacheDir function returns image's cache directory on host system.
func CacheDir(dist, packageName, packageVersion string) string {
	return filepath.Join(
		"/tmp",
		Image(dist, packageName, packageVersion),
	)
}

// BuildDir function returns container's build directory on host system.
func BuildDir(dist, packageName, packageVersion string) string {
	return filepath.Join(
		"/tmp",
		Container(dist, packageName, packageVersion),
	)
}
