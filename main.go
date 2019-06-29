package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
)

const (
	Name        = "deber"
	Version     = "0.5"
	Description = "Debian packaging with Docker."
)

var (
	dist           = pflag.StringP("distribution", "d", "", "")
	extraPackages  = pflag.StringArrayP("extra-package", "p", nil, "")
	noImageRebuild = pflag.BoolP("no-image-rebuild", "r", false, "")
	noAptUpdate    = pflag.BoolP("no-apt-update", "u", false, "")
	noLogColor     = pflag.Bool("no-log-color", false, "")

	listPackages   = pflag.Bool("list-packages", false, "")
	listContainers = pflag.Bool("list-containers", false, "")
	listImages     = pflag.Bool("list-images", false, "")
	listSteps      = pflag.Bool("list-steps", false, "")

	includeSteps = pflag.StringArrayP("include-step", "i", nil, "")
	excludeSteps = pflag.StringArrayP("exclude-step", "e", nil, "")
)

func init() {
	pflag.Parse()
}

func main() {
	log := logger.New(Name, !*noLogColor)

	ch, err := changelog.ParseFileOne("debian/changelog")
	check(log, err)

	dock, err := docker.New()
	check(log, err)

	cwd, err := os.Getwd()
	check(log, err)

	n := &naming.Naming{
		BaseArchiveDir: filepath.Join(os.Getenv("HOME"), Name),
		BaseSourceDir:  cwd,
		BaseCacheDir:   "/tmp",
		BaseBuildDir:   "/tmp",

		Prefix:    Name,
		Changelog: ch,
	}

	step

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

	}

	if *listImages {

	}

	if *listSteps {

	}

}

func check(l *logger.Logger, err error) {
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
}
