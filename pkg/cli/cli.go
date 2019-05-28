// Package cli is the core of deber.
package cli

import (
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
		RunE: func(cmd *cobra.Command, a []string) error {
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

			packageArgs := steps.PackageArgs{}

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
		RunE: func(cmd *cobra.Command, a []string) error {
			args := steps.BuildArgs{
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

				args.ImageName = name.Image
				args.Distribution = name.Distribution
			}

			dock, err := docker.New()
			if err != nil {
				return err
			}

			return steps.Build(dock, args)
		},
	}

	cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "",
		RunE: func(cmd *cobra.Command, a []string) error {

		},
	}

	cmdList = &cobra.Command{
		Use:   "list",
		Short: "",
		RunE: func(cmd *cobra.Command, a []string) error {

		},
	}

	cmdShell = &cobra.Command{
		Use:   "shell",
		Short: "",
		RunE: func(cmd *cobra.Command, a []string) error {

		},
	}

	cmdRemove = &cobra.Command{
		Use:   "remove",
		Short: "",
		RunE: func(cmd *cobra.Command, a []string) error {

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
