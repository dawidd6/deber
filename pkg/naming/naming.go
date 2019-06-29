package naming

import (
	"fmt"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
)

type Naming struct {
	BaseSourceDir  string
	BaseBuildDir   string
	BaseCacheDir   string
	BaseArchiveDir string

	Prefix    string
	Changelog *changelog.ChangelogEntry
}

func (naming *Naming) Image() string {
	return fmt.Sprintf(
		"%s:%s",
		naming.Prefix,
		naming.standardizePackageTarget(),
	)
}

func (naming *Naming) Container() string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		naming.Prefix,
		naming.Changelog.Target,
		naming.Changelog.Source,
		naming.standardizePackageVersion(),
	)
}

func (naming *Naming) standardizePackageVersion() string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version := strings.Replace(naming.Changelog.Version.String(), "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func (naming *Naming) standardizePackageTarget() string {
	// TODO figure out what to do with ubuntu backports, cause there are no docker ubuntu backports images
	if strings.Contains(naming.Changelog.Target, "backports") {
		return naming.Changelog.Target
	}

	if naming.Changelog.Target == "UNRELEASED" {
		return "unstable"
	}

	if strings.Contains(naming.Changelog.Target, "-") {
		return strings.Split(naming.Changelog.Target, "-")[0]
	}

	return naming.Changelog.Target
}

func (naming *Naming) BuildDir() string {
	return filepath.Join(
		naming.BaseBuildDir,
		naming.Container(),
	)
}

func (naming *Naming) CacheDir() string {
	return filepath.Join(
		naming.BaseCacheDir,
		naming.Image(),
	)
}

func (naming *Naming) SourceDir() string {
	return naming.BaseSourceDir
}

func (naming *Naming) SourceParentDir() string {
	return filepath.Dir(naming.SourceDir())
}

func (naming *Naming) ArchiveTargetDir() string {
	return filepath.Join(
		naming.BaseArchiveDir,
		naming.standardizePackageTarget(),
	)
}
func (naming *Naming) ArchivePackageDir() string {
	return filepath.Join(
		naming.ArchiveTargetDir(),
		naming.Changelog.Source,
	)
}
func (naming *Naming) ArchiveVersionDir() string {
	return filepath.Join(
		naming.ArchivePackageDir(),
		naming.Changelog.Version.String(),
	)
}
