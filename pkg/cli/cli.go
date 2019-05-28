// Package cli is the core of deber.
package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
	"os"
	"pault.ag/go/debian/changelog"
)

// Commands
var (
	cmdRoot = &cobra.Command{
		Use:     "deber",
		Version: "0.5",
		Short:   "Debian packaging with Docker.",
		RunE: func(cmd *cobra.Command, args []string) error {
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := changelog.ParseFileOne("debian/changelog")
			if err != nil {
				return err
			}

			name := naming.New(
				deb.Target,
				deb.Source,
				deb.Version.String(),
			)

			buildArgs := steps.BuildArgs{
				ImageName:    name.Image,
				Distribution: name.Distribution,
			}

			createArgs := steps.CreateArgs{
				ImageName:     name.Image,
				ContainerName: name.Container,
				SourceDir:     name.SourceDir,
				BuildDir:      name.BuildDir,
				CacheDir:      name.CacheDir,
				ExtraPackages: flagExtraPackages,
			}

			dependsArgs := steps.DependsArgs{
				ContainerName: name.Container,
				ExtraPackages: flagExtraPackages,
			}

			packageArgs := steps.PackageArgs{
				ContainerName:    name.Container,
				DpkgFlags:        flagDpkgFlags,
				LintianFlags:     flagLintianFlags,
				IsTestNeeded:     true,
				IsNetworkNeeded:  flagWithNetwork,
				PackageName:      name.PackageName,
				PackageVersion:   name.PackageVersion,
				TarballSourceDir: name.SourceParentDir,
				TarballTargetDir: name.BuildDir,
			}

			archiveArgs := steps.ArchiveArgs{
				ArchivePackageDir: name.ArchivePackageDir,
				BuildDir:          name.BuildDir,
			}

			removeArgs := steps.RemoveArgs{
				ContainerName: name.Container,
			}

			err = steps.Build(dock, buildArgs)
			if err != nil {
				return err
			}

			err = steps.Create(dock, createArgs)
			if err != nil {
				return err
			}

			err = steps.Depends(dock, dependsArgs)
			if err != nil {
				return err
			}

			err = steps.Package(dock, packageArgs)
			if err != nil {
				return err
			}

			err = steps.Archive(dock, archiveArgs)
			if err != nil {
				return err
			}

			err = steps.Remove(dock, removeArgs)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmdBuild = &cobra.Command{
		Use:   "build",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			buildArgs := steps.BuildArgs{
				ImageName:    "deber:" + flagDistribution,
				Distribution: flagDistribution,
				Rebuild:      flagRebuild,
			}

			if flagDistribution == "" {
				deb, err := changelog.ParseFileOne("debian/changelog")
				if err != nil {
					return err
				}

				name := naming.New(
					deb.Target,
					deb.Source,
					deb.Version.String(),
				)

				buildArgs.ImageName = name.Image
				buildArgs.Distribution = name.Distribution
			}

			dock, err := docker.New()
			if err != nil {
				return err
			}

			return steps.Build(dock, buildArgs)
		},
	}

	cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := changelog.ParseFileOne("debian/changelog")
			if err != nil {
				return err
			}

			name := naming.New(
				deb.Target,
				deb.Source,
				deb.Version.String(),
			)

			createArgs := steps.CreateArgs{
				ImageName:     name.Image,
				ContainerName: name.Container,
				SourceDir:     name.SourceDir,
				BuildDir:      name.BuildDir,
				CacheDir:      name.CacheDir,
				ExtraPackages: flagExtraPackages,
			}

			return steps.Create(dock, createArgs)
		},
	}

	cmdList = &cobra.Command{
		Use:   "list",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			dock, err := docker.New()
			if err != nil {
				return err
			}

			if !flagContainers && !flagImages && !flagPackages {
				flagContainers = true
				flagImages = true
				flagPackages = true
			}

			if flagImages {
				listArgs := docker.ImageListArgs{
					Prefix: "deber",
				}

				images, err := dock.ImageList(listArgs)
				if err != nil {
					return err
				}

				fmt.Println("Images:")
				for _, image := range images {
					fmt.Printf("  - %s\n", image)
				}
			}

			if flagContainers {
				listArgs := docker.ContainerListArgs{
					Prefix: "deber",
				}

				containers, err := dock.ContainerList(listArgs)
				if err != nil {
					return err
				}

				fmt.Println("Containers:")
				for _, container := range containers {
					fmt.Printf("  - %s\n", container)
				}
			}

			if flagPackages {

			}

			return nil
		},
	}

	cmdShell = &cobra.Command{
		Use:   "shell",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := changelog.ParseFileOne("debian/changelog")
			if err != nil {
				return err
			}

			name := naming.New(
				deb.Target,
				deb.Source,
				deb.Version.String(),
			)

			shellArgs := steps.ShellArgs{
				ContainerName: name.Container,
			}

			return steps.Shell(dock, shellArgs)
		},
	}

	cmdRemove = &cobra.Command{
		Use:   "remove",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := changelog.ParseFileOne("debian/changelog")
			if err != nil {
				return err
			}

			name := naming.New(
				deb.Target,
				deb.Source,
				deb.Version.String(),
			)

			removeArgs := steps.RemoveArgs{
				ContainerName: name.Container,
			}

			return steps.Remove(dock, removeArgs)
		},
	}
)

// Flags
var (
	// Root flags
	flagDpkgFlags    string
	flagLintianFlags string
	flagWithNetwork  bool

	// Root|Create flags
	flagExtraPackages []string

	// List flags
	flagImages     bool
	flagContainers bool
	flagPackages   bool

	// Build flags
	flagDistribution string
	flagRebuild      bool

	// Remove flags
	flagAll bool

	// Root-Global flags
	flagNoColor bool
)

func init() {
	cmdRoot.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmdRoot.Flags().SortFlags = false
	cmdRoot.SilenceErrors = true
	cmdRoot.SilenceUsage = true

	cmdRoot.AddCommand(
		cmdBuild,
		cmdCreate,
		cmdList,
		cmdShell,
		cmdRemove,
	)

	// Root flags
	cmdRoot.Flags().StringVarP(&flagDpkgFlags, "dpkg-flags", "d", "-tc", "")
	cmdRoot.Flags().StringVarP(&flagLintianFlags, "lintian-flags", "l", "-i -I", "")
	cmdRoot.Flags().BoolVarP(&flagWithNetwork, "with-network", "n", false, "")
	cmdRoot.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "")
	cmdRoot.Flags().StringArrayVarP(&flagExtraPackages, "extra-package", "p", nil, "")

	// Build flags
	cmdBuild.Flags().StringVarP(&flagDistribution, "distribution", "d", "", "")
	cmdBuild.Flags().BoolVarP(&flagRebuild, "rebuild", "r", false, "")

	// Create flags
	cmdCreate.Flags().StringArrayVarP(&flagExtraPackages, "extra-package", "p", nil, "")

	// List flags
	cmdList.Flags().BoolVarP(&flagImages, "images", "i", false, "")
	cmdList.Flags().BoolVarP(&flagContainers, "containers", "c", false, "")
	cmdList.Flags().BoolVarP(&flagPackages, "packages", "p", false, "")

	// Remove flags
	cmdRemove.Flags().BoolVarP(&flagAll, "all", "a", false, "")
}

// Run function is the first that should be executed.
//
// It's the heart of cli.
func Run() {
	err := cmdRoot.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
