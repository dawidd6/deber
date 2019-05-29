// Package naming standardizes some names used internally by deber.
package naming

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
)

var (
	BuildBase   = "/tmp"
	CacheBase   = "/tmp"
	ArchiveBase = filepath.Join(os.Getenv("HOME"), app.Name)
)

type Naming struct {
	Container *Container
	Image     *Image
	Dirs      *Dirs
	Package   *Package
}

type Dirs struct {
	Build   *Build
	Cache   *Cache
	Archive *Archive
	Source  *Source
}

type Package struct {
	*changelog.ChangelogEntry
}

type Image struct {
	Package *Package
}

type Container struct {
	Package *Package
}

type Build struct {
	Base      string
	Container *Container
}

type Cache struct {
	Base  string
	Image *Image
}

type Archive struct {
	Base    string
	Package *Package
}

type Source struct {
	Base string
}

func New(debian *changelog.ChangelogEntry) *Naming {
	pkg := &Package{
		debian,
	}

	container := &Container{
		Package: pkg,
	}

	image := &Image{
		Package: pkg,
	}

	dirs := &Dirs{
		Source: &Source{
			Base: os.Getenv("PWD"),
		}, Build: &Build{
			Base:      BuildBase,
			Container: container,
		}, Cache: &Cache{
			Base:  CacheBase,
			Image: image,
		}, Archive: &Archive{
			Base:    ArchiveBase,
			Package: pkg,
		},
	}

	return &Naming{
		Container: container,
		Image:     image,
		Package:   pkg,
		Dirs:      dirs,
	}
}

func (image *Image) Name() string {
	return fmt.Sprintf(
		"%s:%s",
		app.Name,
		standardizeImageTag(image.Package),
	)
}

func (image *Image) Repo() string {
	return app.Name
}

func (image *Image) Tag() string {
	return standardizeImageTag(image.Package)
}

func (container *Container) Name() string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		app.Name,
		container.Package.Target,
		container.Package.Source,
		standardizePackageVersion(container.Package.Version.String()),
	)
}

func standardizePackageVersion(version string) string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func standardizeImageTag(pkg *Package) string {
	if strings.Contains(pkg.Version.String(), "bpo") {
		if strings.Contains(pkg.Target, "backports") {
			return pkg.Target
		}

		if pkg.Target == "UNRELEASED" {
			return "unstable"
		}
	}

	if strings.Contains(pkg.Target, "-") {
		return strings.Split(pkg.Target, "-")[0]
	}

	return pkg.Target
}

func (build *Build) ContainerPath() string {
	return filepath.Join(
		build.Base,
		build.Container.Name(),
	)
}

func (cache *Cache) ImagePath() string {
	return filepath.Join(
		cache.Base,
		cache.Image.Name(),
	)
}

func (archive *Archive) PackageTargetPath() string {
	return filepath.Join(
		archive.Base,
		archive.Package.Target,
	)
}

func (archive *Archive) PackageSourcePath() string {
	return filepath.Join(
		archive.Base,
		archive.Package.Target,
		archive.Package.Source,
	)
}

func (archive *Archive) PackageVersionPath() string {
	return filepath.Join(
		archive.Base,
		archive.Package.Target,
		archive.Package.Source,
		archive.Package.Version.String(),
	)
}

func (source *Source) SourcePath() string {
	return source.Base
}

func (source *Source) ParentPath() string {
	return filepath.Dir(
		source.Base,
	)
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
