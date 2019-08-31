package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/dawidd6/deber/pkg/tree"
	"github.com/spf13/cobra"
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
	dist          string
	extraPackages []string
	withNetwork   = false
	maxImageAge   = time.Hour * 24 * 14
	launchShell   = false
	keepContainer = false

	dpkgFlags    = "-tc"
	lintianFlags = "-i -I"

	archiveBaseDir = filepath.Join(os.Getenv("HOME"), Name)
	cacheBaseDir   = "/tmp"
	buildBaseDir   = "/tmp"

	listPackages   = false
	listContainers = false
	listImages     = false
	listAll        = false

	sourceBaseDir string
)

func main() {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [FLAGS ...] [DIR]", Name),
		Short:   Description,
		Version: Version,
		RunE:    run,
	}
	cmd.Flags().StringVarP(&dist, "distribution", "d", dist, "override target distribution")
	cmd.Flags().StringArrayVarP(&extraPackages, "extra-package", "p", extraPackages, "additional packages to be installed in container (either single .deb or a directory)")
	cmd.Flags().BoolVarP(&withNetwork, "with-network", "n", withNetwork, "allow network access during package build")
	cmd.Flags().DurationVarP(&maxImageAge, "max-image-age", "a", maxImageAge, "time after which image will be refreshed")
	cmd.Flags().BoolVarP(&launchShell, "launch-shell", "s", launchShell, "launch interactive shell in container")
	cmd.Flags().BoolVarP(&keepContainer, "keep-container", "k", keepContainer, "do not remove container at the end of the process")
	cmd.Flags().BoolVar(&log.NoColor, "no-log-color", log.NoColor, "do not colorize log output")
	cmd.Flags().StringVar(&dpkgFlags, "dpkg-flags", dpkgFlags, "additional flags to be passed to dpkg-buildpackage in container")
	cmd.Flags().StringVar(&lintianFlags, "lintian-flags", lintianFlags, "additional flags to be passed to lintian in container")
	cmd.Flags().StringVar(&archiveBaseDir, "archive-dir", archiveBaseDir, "where to store build artifacts")
	cmd.Flags().StringVar(&cacheBaseDir, "cache-dir", cacheBaseDir, "where to store images' apt cache")
	cmd.Flags().StringVar(&buildBaseDir, "build-dir", buildBaseDir, "where to place temporary build directory")
	cmd.Flags().BoolVar(&listPackages, "list-packages", listPackages, "print all packages available in archive")
	cmd.Flags().BoolVar(&listContainers, "list-containers", listContainers, "print all currently created containers")
	cmd.Flags().BoolVar(&listImages, "list-images", listImages, "print all built images")
	cmd.Flags().BoolVar(&listAll, "list-all", listAll, "print packages, images and containers")

	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.DisableFlagsInUseLine = true
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.Flags().SortFlags = false

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	dock, err := docker.New()
	if err != nil {
		return err
	}

	if listAll {
		listContainers = true
		listImages = true
		listPackages = true
	}

	if listContainers || listImages || listPackages {
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

	if dist == "" {
		dist = ch.Target
	}

	namingArgs := naming.Args{
		Prefix:         Name,
		Source:         ch.Source,
		Version:        ch.Version.String(),
		Upstream:       ch.Version.Version,
		Target:         dist,
		SourceBaseDir:  sourceBaseDir,
		BuildBaseDir:   buildBaseDir,
		CacheBaseDir:   cacheBaseDir,
		ArchiveBaseDir: archiveBaseDir,
	}
	n := naming.New(namingArgs)

	err = steps.Build(dock, n, maxImageAge)
	if err != nil {
		return err
	}

	err = steps.Create(dock, n, extraPackages)
	if err != nil {
		return err
	}

	err = steps.Start(dock, n)
	if err != nil {
		return err
	}

	if launchShell {
		return steps.ShellOptional(dock, n)
	}

	err = steps.Tarball(n)
	if err != nil {
		return err
	}

	err = steps.Depends(dock, n, extraPackages)
	if err != nil {
		return err
	}

	err = steps.Package(dock, n, dpkgFlags, withNetwork)
	if err != nil {
		return err
	}

	err = steps.Test(dock, n, lintianFlags)
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

	if !keepContainer {
		err = steps.Remove(dock, n)
		if err != nil {
			return err
		}
	}

	return nil
}

func list(dock *docker.Docker) error {
	indent := "    "

	if listPackages {
		_, err := os.Stat(archiveBaseDir)
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

			t, err := tree.New(archiveBaseDir, 3)
			if err != nil {
				return err
			}

			err = t.Walk(walker)
			if err != nil {
				return err
			}
		}
	}

	if listContainers {
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

	if listImages {
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
