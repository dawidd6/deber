package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"github.com/dawidd6/deber/pkg/walk"
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
	dist = pflag.StringP("distribution", "d", "", "")

	extraPackages = pflag.StringArrayP("extra-package", "p", nil, "")

	noTarball      = pflag.BoolP("no-tarball", "t", false, "")
	noImageRebuild = pflag.BoolP("no-image-rebuild", "r", false, "")
	noAptUpdate    = pflag.BoolP("no-apt-update", "u", false, "")
	withNetwork    = pflag.BoolP("with-network", "n", false, "")
	launchShell    = pflag.BoolP("launch-shell", "s", false, "")
	keepContainer  = pflag.BoolP("keep-container", "k", false, "")
	checkBefore    = pflag.BoolP("check-before", "c", false, "")
	maxImageAge    = pflag.DurationP("max-image-age", "a", time.Hour*24*14, "")

	dpkgFlags    = pflag.String("dpkg-flags", "-tc", "")
	lintianFlags = pflag.String("lintian-flags", "-i -I", "")

	archiveBaseDir = pflag.String("archive-base-dir", filepath.Join(os.Getenv("HOME"), Name), "")
	cacheBaseDir   = pflag.String("cache-base-dir", "/tmp", "")
	buildBaseDir   = pflag.String("build-base-dir", "/tmp", "")
	sourceBaseDir  = pflag.String("source-base-dir", os.Getenv("PWD"), "")

	listPackages   = pflag.Bool("list-packages", false, "")
	listContainers = pflag.Bool("list-containers", false, "")
	listImages     = pflag.Bool("list-images", false, "")
	noLogColor     = pflag.Bool("no-log-color", false, "")
)

func init() {
	pflag.Parse()
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

	if *dist != "" {
		ch.Target = *dist
	}

	if *listPackages {
		err := walk.Walk(n.BaseArchiveDir, 3, func(node *walk.Node) bool {
			indent := ""
			for i := 1; i < node.Depth(); i++ {
				indent += "    "
			}

			fmt.Printf("%s%s\n", indent, node.Name())

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

	steps := stepping.Stepping{
		Docker: dock,
		Log:    log,
		Naming: n,
		Debian: ch,
	}

	if *checkBefore {
		err = steps.CheckOptional()
		check(log, err)
	}

	err = steps.Build(*noImageRebuild, *maxImageAge)
	check(log, err)

	err = steps.Create(*extraPackages)
	check(log, err)

	err = steps.Start()
	check(log, err)

	if *launchShell {
		err = steps.ShellOptional()
		check(log, err)
		return
	}

	if !*noTarball {
		err = steps.Tarball()
		check(log, err)
	}

	err = steps.Depends(*extraPackages, *noAptUpdate)
	check(log, err)

	err = steps.Package(*dpkgFlags, *withNetwork)
	check(log, err)

	err = steps.Test(*lintianFlags)
	check(log, err)

	err = steps.Archive()
	check(log, err)

	err = steps.Stop()
	check(log, err)

	if !*keepContainer {
		err = steps.Remove()
		check(log, err)
	}
}

func check(l *logger.Logger, err error) {
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
}
