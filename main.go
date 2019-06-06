package main

import (
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/cli"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/pflag"
	"path/filepath"

	"os"
	"pault.ag/go/debian/changelog"
)

const (
	Name        = "deber"
	Version     = "0.5"
	Description = "Debian packaging with Docker."
)

var (
	dpkgFlags     = pflag.String("dpkg-flags", "-tc", "")
	lintianFlags  = pflag.String("lintian-flags", "-i -I", "")
	extraPackages = pflag.StringArray("extra-package", nil, "")

	archiveDir = pflag.String("archive-base-dir", filepath.Join(os.Getenv("HOME"), Name), "")
	cacheDir   = pflag.String("cache-base-dir", "/tmp", "")
	buildDir   = pflag.String("build-base-dir", "/tmp", "")
)

func main() {
	pflag.Parse()

	log := logger.New(Name)

	dock, err := docker.New()
	check(log, err)

	debian, err := changelog.ParseFileOne("debian/changelog")
	check(log, err)

	options := &cli.Options{
		DpkgFlags:     *dpkgFlags,
		LintianFlags:  *lintianFlags,
		ExtraPackages: *extraPackages,
	}

	name := &naming.Naming{
		Program: Name,

		ChangelogEntry: debian,

		ArchiveBaseDir: *archiveDir,
		CacheBaseDir:   *cacheDir,
		BuildBaseDir:   *buildDir,
		SourceBaseDir:  os.Getenv("PWD"),
	}

	a := &app.App{
		Logger:  log,
		Docker:  dock,
		Naming:  name,
		Options: options,
	}

	err = a.Run()
	check(log, err)
}

func check(log *logger.Logger, err error) {
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
