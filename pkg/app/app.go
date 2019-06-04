package app

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/config"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
)

type App struct {
	Name string

	*logger.Logger
	*config.Config
	*docker.Docker
	*changelog.ChangelogEntry
}

func (app *App) ImageName() string {
	return fmt.Sprintf(
		"%s:%s",
		app.Name,
		app.standardizeImageTag(),
	)
}

func (app *App) ImageRepo() string {
	return app.Name
}

func (app *App) ImageTag() string {
	return app.standardizeImageTag()
}

func (app *App) ContainerName() string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		app.Name,
		app.Target,
		app.Source,
		app.standardizePackageVersion(),
	)
}

func (app *App) standardizePackageVersion() string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version := app.Version.String()
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func (app *App) standardizeImageTag() string {
	if strings.Contains(app.Version.String(), "bpo") {
		if strings.Contains(app.Target, "backports") {
			return app.Target
		}

		if app.Target == "UNRELEASED" {
			return "unstable"
		}
	}

	if strings.Contains(app.Target, "-") {
		return strings.Split(app.Target, "-")[0]
	}

	return app.Target
}

func (app *App) BuildDir() string {
	return filepath.Join(
		app.BuildBaseDir,
		app.ContainerName(),
	)
}

func (app *App) CacheDir() string {
	return filepath.Join(
		app.CacheBaseDir,
		app.ImageName(),
	)
}

func (app *App) ArchiveTargetDir() string {
	return filepath.Join(
		app.ArchiveBaseDir,
		app.Target,
	)
}

func (app *App) ArchiveSourceDir() string {
	return filepath.Join(
		app.ArchiveTargetDir(),
		app.Source,
	)
}

func (app *App) ArchiveVersionDir() string {
	return filepath.Join(
		app.ArchiveSourceDir(),
		app.Version.String(),
	)
}

func (app *App) SourceDir() string {
	return app.SourceBaseDir
}

func (app *App) SourceParentDir() string {
	return filepath.Dir(
		app.SourceDir(),
	)
}
