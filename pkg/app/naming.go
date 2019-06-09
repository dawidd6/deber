package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (a *App) ImageName() string {
	return fmt.Sprintf(
		"%s:%s",
		a.Name,
		a.standardizeImageTag(),
	)
}

func (a *App) ImageRepo() string {
	return a.Name
}

func (a *App) ImageTag() string {
	return a.standardizeImageTag()
}

func (a *App) ContainerName() string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		a.Name,
		a.Debian.Target,
		a.Debian.Source,
		a.standardizePackageVersion(),
	)
}

func (a *App) standardizePackageVersion() string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version := a.Debian.Version.String()
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func (a *App) standardizeImageTag() string {
	if strings.Contains(a.Debian.Version.String(), "bpo") {
		if strings.Contains(a.Debian.Target, "backports") {
			return a.Debian.Target
		}

		if a.Debian.Target == "UNRELEASED" {
			return "unstable"
		}
	}

	if strings.Contains(a.Debian.Target, "-") {
		return strings.Split(a.Debian.Target, "-")[0]
	}

	return a.Debian.Target
}

func (a *App) BuildDir() string {
	return filepath.Join(
		a.Config.BuildBaseDir,
		a.ContainerName(),
	)
}

func (a *App) CacheDir() string {
	return filepath.Join(
		a.Config.CacheBaseDir,
		a.ImageName(),
	)
}

func (a *App) ArchiveTargetDir() string {
	return filepath.Join(
		a.Config.ArchiveBaseDir,
		a.Debian.Target,
	)
}

func (a *App) ArchiveSourceDir() string {
	return filepath.Join(
		a.ArchiveTargetDir(),
		a.Debian.Source,
	)
}

func (a *App) ArchiveVersionDir() string {
	return filepath.Join(
		a.ArchiveSourceDir(),
		a.Debian.Version.String(),
	)
}

func (a *App) SourceDir() string {
	return a.Config.SourceBaseDir
}

func (a *App) SourceParentDir() string {
	return filepath.Dir(
		a.SourceDir(),
	)
}
