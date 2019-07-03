package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"github.com/dawidd6/deber/pkg/utils"
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
	launchShell   = pflag.BoolP("launch-shell", "s", false, "")
	keepContainer = pflag.BoolP("keep-container", "k", false, "")
	checkBefore   = pflag.BoolP("check-before", "c", false, "")

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

	err := run(log)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func list(dock *docker.Docker, n *naming.Naming) (bool, error) {
	listed := false

	if *listPackages {
		listed = true

		fmt.Println("Packages:")
		err := utils.Walk(n.BaseArchiveDir, 3, func(file *utils.File) bool {
			indent := ""
			for i := 1; i < file.Depth(); i++ {
				indent += "    "
			}

			fmt.Printf("%s%s\n", indent, file.Name())

			return false
		})
		if err != nil {
			return listed, err
		}
	}

	if *listContainers {
		listed = true

		list, err := dock.ContainerList(Name)
		if err != nil {
			return listed, err
		}

		fmt.Println("Containers:")
		for _, container := range list {
			fmt.Println(container)
		}
	}

	if *listImages {
		listed = true

		list, err := dock.ImageList(Name)
		if err != nil {
			return listed, err
		}

		fmt.Println("Images:")
		for _, image := range list {
			fmt.Println(image)
		}
	}

	return listed, nil
}

func run(log *logger.Logger) error {
	dock, err := docker.New()
	if err != nil {
		return err
	}

	dch := filepath.Join(*sourceBaseDir, "debian/changelog")
	ch, err := changelog.ParseFileOne(dch)
	if err != nil {
		return err
	}

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

	listed, err := list(dock, n)
	if err != nil {
		return err
	}
	if listed {
		return nil
	}

	if *checkBefore {
		err = s.CheckOptional()
		if err != nil {
			return err
		}
	}

	err = s.Build()
	if err != nil {
		return err
	}

	err = s.Create()
	if err != nil {
		return err
	}

	err = s.Start()
	if err != nil {
		return err
	}

	if *launchShell {
		err = s.ShellOptional()
		if err != nil {
			return err
		}

		return nil
	}

	err = s.Tarball()
	if err != nil {
		return err
	}

	err = s.Depends()
	if err != nil {
		return err
	}

	err = s.Package()
	if err != nil {
		return err
	}

	err = s.Test()
	if err != nil {
		return err
	}

	err = s.Archive()
	if err != nil {
		return err
	}

	err = s.Stop()
	if err != nil {
		return err
	}

	if !*keepContainer {
		err = s.Remove()
		if err != nil {
			return err
		}
	}

	return nil
}
