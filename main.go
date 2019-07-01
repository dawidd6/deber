package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/filewalk"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"time"
)

const (
	Name        = "deber"
	Version     = "0.5"
	Description = "Debian packaging with Docker."
)

var (
	dist          = pflag.StringP("distribution", "d", "", "")
	extraPackages = pflag.StringArrayP("extra-package", "p", nil, "")
	withNetwork   = pflag.BoolP("with-network", "n", false, "")
	maxImageAge   = pflag.DurationP("max-image-age", "a", time.Hour*24*14, "")

	includeSteps = pflag.StringArrayP("include-step", "i", nil, "")
	excludeSteps = pflag.StringArrayP("exclude-step", "e", nil, "")

	dpkgFlags    = pflag.String("dpkg-flags", "-tc", "")
	lintianFlags = pflag.String("lintian-flags", "-i -I", "")

	archiveBaseDir = pflag.String("archive-base-dir", filepath.Join(os.Getenv("HOME"), Name), "")
	cacheBaseDir   = pflag.String("cache-base-dir", "/tmp", "")
	buildBaseDir   = pflag.String("build-base-dir", "/tmp", "")
	sourceBaseDir  = pflag.String("source-base-dir", os.Getenv("PWD"), "")

	listPackages   = pflag.Bool("list-packages", false, "")
	listContainers = pflag.Bool("list-containers", false, "")
	listImages     = pflag.Bool("list-images", false, "")
	listSteps      = pflag.Bool("list-steps", false, "")
	noLogColor     = pflag.Bool("no-log-color", false, "")

	version = pflag.Bool("version", false, "")
)

func init() {
	pflag.Parse()

	if *version {
		fmt.Println(Name, Version)
		os.Exit(0)
	}
}

func main() {
	log := logger.New(Name, !*noLogColor)

	dock, err := docker.New()
	check(log, err)

	dch := filepath.Join(*sourceBaseDir, "debian/changelog")
	ch, err := changelog.ParseFileOne(dch)
	check(log, err)

	n := &naming.Naming{
		BaseArchiveDir: *archiveBaseDir,
		BaseSourceDir:  *sourceBaseDir,
		BaseCacheDir:   *cacheBaseDir,
		BaseBuildDir:   *buildBaseDir,

		Prefix:    Name,
		Changelog: ch,
	}

	s := &stepping.Stepping{
		Docker: dock,
		Log:    log,
		Naming: n,
		Debian: ch,

		DpkgFlags:    *dpkgFlags,
		LintianFlags: *lintianFlags,

		PackageWithNetwork: *withNetwork,
		RebuildImageIfOld:  true,
		MaxImageAge:        *maxImageAge,
		ExtraPackages:      *extraPackages,
	}

	if *dist != "" {
		ch.Target = *dist
	}

	if *listPackages {
		err := filewalk.Walk(n.BaseArchiveDir, 3, func(file *filewalk.File) bool {
			indent := ""
			for i := 1; i < file.Depth(); i++ {
				indent += "    "
			}

			fmt.Printf("%s%s\n", indent, file.Name())

			return false
		})
		check(log, err)
	}

	if *listContainers {
		list, err := dock.ContainerList(Name)
		check(log, err)

		for _, container := range list {
			fmt.Println(container)
		}
	}

	if *listImages {
		list, err := dock.ImageList(Name)
		check(log, err)

		for _, image := range list {
			fmt.Println(image)
		}
	}

	if *listSteps {
		s.Steps().Walk(func(step *stepping.Step) {
			if step.Optional {
				fmt.Println(step.Name, "(optional)")
			} else {
				fmt.Println(step.Name)
			}
			fmt.Println("  ", step.Description)
		})
	}

	err = s.Steps().Include(*includeSteps...).Exclude(*excludeSteps...).Run()
	check(log, err)
}

func check(l *logger.Logger, err error) {
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
}
