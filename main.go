package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/dawidd6/deber/pkg/tree"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"time"
)

const (
	// Name is the name of program
	Name = "deber"
	// Version of program
	Version = "0.5"
	// Description of program
	Description = "Debian packaging with Docker."
)

var (
	dist           = pflag.StringP("distribution", "d", "", "override target distribution")
	extraPackages  = pflag.StringArrayP("extra-package", "p", nil, "additional packages to be installed in container (either single .deb or a directory)")
	maxImageAge    = pflag.DurationP("max-image-age", "a", time.Hour*24*14, "time after which image will be refreshed")
	withNetwork    = pflag.BoolP("with-network", "n", false, "allow network access during package build")
	launchShell    = pflag.BoolP("launch-shell", "s", false, "launch interactive shell in container")
	keepContainer  = pflag.BoolP("keep-container", "k", false, "do not remove container at the end of the process")
	dpkgFlags      = pflag.String("dpkg-flags", "-tc", "additional flags to be passed to dpkg-buildpackage in container")
	lintianFlags   = pflag.String("lintian-flags", "-i -I", "additional flags to be passed to lintian in container")
	archiveBaseDir = pflag.String("archive-dir", filepath.Join(os.Getenv("HOME"), Name), "where to store build artifacts")
	cacheBaseDir   = pflag.String("cache-dir", "/tmp", "where to store images' apt cache")
	buildBaseDir   = pflag.String("build-dir", "/tmp", "where to place temporary build directory")
	noLogColor     = pflag.Bool("no-log-color", false, "do not colorize log output")
	listPackages   = pflag.Bool("list-packages", false, "print all packages available in archive")
	listContainers = pflag.Bool("list-containers", false, "print all currently created containers")
	listImages     = pflag.Bool("list-images", false, "print all built images")
	listAll        = pflag.Bool("list-all", false, "print packages, images and containers")

	sourceBaseDir string
)

func main() {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [FLAGS ...] [DIR]", Name),
		Short:   Description,
		Version: Version,
		RunE:    run,
	}

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	log.NoColor = *noLogColor

	dock, err := docker.New()
	if err != nil {
		return err
	}

	if *listAll {
		*listContainers = true
		*listImages = true
		*listPackages = true
	}

	if *listContainers || *listImages || *listPackages {
		return list(dock)
	}

	if len(args) > 0 {
		sourceBaseDir = args[0]
	} else {
		sourceBaseDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	path := filepath.Join(sourceBaseDir, "debian/changelog")
	ch, err := changelog.ParseFileOne(path)
	if err != nil {
		return err
	}

	if *dist == "" {
		*dist = ch.Target
	}

	namingArgs := naming.Args{
		Prefix:         Name,
		Source:         ch.Source,
		Version:        ch.Version.String(),
		Upstream:       ch.Version.Version,
		Target:         *dist,
		SourceBaseDir:  sourceBaseDir,
		BuildBaseDir:   *buildBaseDir,
		CacheBaseDir:   *cacheBaseDir,
		ArchiveBaseDir: *archiveBaseDir,
	}
	n := naming.New(namingArgs)

	err = steps.Build(dock, n, *maxImageAge)
	if err != nil {
		return err
	}

	err = steps.Create(dock, n, *extraPackages)
	if err != nil {
		return err
	}

	err = steps.Start(dock, n)
	if err != nil {
		return err
	}

	if *launchShell {
		return steps.ShellOptional(dock, n)
	}

	err = steps.Tarball(n)
	if err != nil {
		return err
	}

	err = steps.Depends(dock, n, *extraPackages)
	if err != nil {
		return err
	}

	err = steps.Package(dock, n, *dpkgFlags, *withNetwork)
	if err != nil {
		return err
	}

	err = steps.Test(dock, n, *lintianFlags)
	if err != nil {
		return err
	}

	err = steps.Archive(n)
	if err != nil {
		return err
	}

	err = steps.Stop(dock, n)
	if err != nil {
		return err
	}

	if !*keepContainer {
		err = steps.Remove(dock, n)
		if err != nil {
			return err
		}
	}

	return nil
}

func list(dock *docker.Docker) error {
	indent := "    "

	if *listPackages {
		_, err := os.Stat(*archiveBaseDir)
		if err == nil {
			fmt.Println("Packages:")

			walker := func(file *tree.File) error {
				addIndent := ""

				for i := 0; i < file.Depth; i++ {
					addIndent += indent
				}

				fmt.Printf("%s%s\n", addIndent, filepath.Base(file.Path))

				return nil
			}

			t, err := tree.New(*archiveBaseDir, 3)
			if err != nil {
				return err
			}

			err = t.Walk(walker)
			if err != nil {
				return err
			}
		}
	}

	if *listContainers {
		list, err := dock.ContainerList(Name)
		if err != nil {
			return err
		}

		if len(list) > 0 {
			fmt.Println("Containers:")
		}

		for _, container := range list {
			fmt.Printf("%s%s\n", indent, container)
		}
	}

	if *listImages {
		list, err := dock.ImageList(Name)
		if err != nil {
			return err
		}

		if len(list) > 0 {
			fmt.Println("Images:")
		}

		for _, image := range list {
			fmt.Printf("%s%s\n", indent, image)
		}
	}

	return nil
}
