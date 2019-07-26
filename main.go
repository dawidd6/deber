package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/dawidd6/deber/pkg/utils"
	"github.com/spf13/cobra"
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

	archiveBaseDir = pflag.String("archive-dir", filepath.Join(os.Getenv("HOME"), Name), "")
	cacheBaseDir   = pflag.String("cache-dir", "/tmp", "")
	buildBaseDir   = pflag.String("build-dir", "/tmp", "")

	listPackages   = pflag.Bool("list-packages", false, "")
	listContainers = pflag.Bool("list-containers", false, "")
	listImages     = pflag.Bool("list-images", false, "")
	noLogColor     = pflag.Bool("no-log-color", false, "")

	sourceBaseDir string
)

func main() {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [DIR]", Name),
		Short:   Description,
		Version: Version,
		PreRunE: pre,
		RunE:    run,
	}
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func pre(cmd *cobra.Command, args []string) error {
	log.Prefix = Name
	log.Color = !*noLogColor

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	dock, err := docker.New()
	if err != nil {
		return err
	}

	listed, err := list(dock)
	if err != nil {
		return err
	}
	if listed {
		return nil
	}

	if len(args) > 0 {
		sourceBaseDir = args[0]
	} else {
		sourceBaseDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	dch := filepath.Join(sourceBaseDir, "debian/changelog")
	ch, err := changelog.ParseFileOne(dch)
	if err != nil {
		return err
	}

	n := &naming.Naming{
		BaseArchiveDir: *archiveBaseDir,
		BaseSourceDir:  sourceBaseDir,
		BaseCacheDir:   *cacheBaseDir,
		BaseBuildDir:   *buildBaseDir,

		Prefix:    Name,
		Changelog: ch,
	}

	if *dist != "" {
		ch.Target = *dist
	}

	if *checkBefore {
		err = steps.CheckOptional(n)
		if err != nil {
			return err
		}
	}

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
		err = steps.ShellOptional(dock, n)
		if err != nil {
			return err
		}

		return nil
	}

	err = steps.Tarball(n, ch)
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

func list(dock *docker.Docker) (bool, error) {
	listed := false

	if *listPackages {
		listed = true

		fmt.Println("Packages:")
		err := utils.Walk(*archiveBaseDir, 3, func(file *utils.File) bool {
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
