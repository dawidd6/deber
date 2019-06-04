// Package naming standardizes some names used internally by deber.
package naming

import (
	"pault.ag/go/debian/changelog"
)

type Naming struct {
	*changelog.ChangelogEntry
}

/*
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
		app.Name,
		Distribution(dist, packageName, packageVersion),
		PackageName(dist, packageName, packageVersion),
		PackageVersion(dist, packageName, packageVersion),
	)
}

// Image function returns a standardized image name.
func Image(dist, packageName, packageVersion string) string {
	return fmt.Sprintf(
		"%s:%s",
		app.Name,
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
		app.Name,
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
*/
