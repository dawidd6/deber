package naming

import (
	"fmt"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
)

type Naming struct {
	Program string

	*changelog.ChangelogEntry

	ArchiveBaseDir string
	CacheBaseDir   string
	BuildBaseDir   string
	SourceBaseDir  string
}

func (naming *Naming) ImageName() string {
	return fmt.Sprintf(
		"%s:%s",
		naming.Program,
		naming.standardizeImageTag(),
	)
}

func (naming *Naming) ImageRepo() string {
	return naming.Program
}

func (naming *Naming) ImageTag() string {
	return naming.standardizeImageTag()
}

func (naming *Naming) ContainerName() string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		naming.Program,
		naming.Target,
		naming.Source,
		naming.standardizePackageVersion(),
	)
}

func (naming *Naming) standardizePackageVersion() string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version := naming.Version.String()
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func (naming *Naming) standardizeImageTag() string {
	if strings.Contains(naming.Version.String(), "bpo") {
		if strings.Contains(naming.Target, "backports") {
			return naming.Target
		}

		if naming.Target == "UNRELEASED" {
			return "unstable"
		}
	}

	if strings.Contains(naming.Target, "-") {
		return strings.Split(naming.Target, "-")[0]
	}

	return naming.Target
}

func (naming *Naming) BuildDir() string {
	return filepath.Join(
		naming.BuildBaseDir,
		naming.ContainerName(),
	)
}

func (naming *Naming) CacheDir() string {
	return filepath.Join(
		naming.CacheBaseDir,
		naming.ImageName(),
	)
}

func (naming *Naming) ArchiveTargetDir() string {
	return filepath.Join(
		naming.ArchiveBaseDir,
		naming.Target,
	)
}

func (naming *Naming) ArchiveSourceDir() string {
	return filepath.Join(
		naming.ArchiveTargetDir(),
		naming.Source,
	)
}

func (naming *Naming) ArchiveVersionDir() string {
	return filepath.Join(
		naming.ArchiveSourceDir(),
		naming.Version.String(),
	)
}

func (naming *Naming) SourceDir() string {
	return naming.SourceBaseDir
}

func (naming *Naming) SourceParentDir() string {
	return filepath.Dir(
		naming.SourceDir(),
	)
}
