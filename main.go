package main

import (
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/config"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/steps"
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

	conf := &config.Config{
		DpkgFlags:     *dpkgFlags,
		LintianFlags:  *lintianFlags,
		ExtraPackages: *extraPackages,

		ArchiveBaseDir: *archiveDir,
		CacheBaseDir:   *cacheDir,
		BuildBaseDir:   *buildDir,
		SourceBaseDir:  os.Getenv("PWD"),
	}

	a := &app.App{
		Name: Name,

		Logger:         log,
		Config:         conf,
		Docker:         dock,
		ChangelogEntry: debian,
	}

	err = steps.Run(a)
	check(log, err)
}

func check(log *logger.Logger, err error) {
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
