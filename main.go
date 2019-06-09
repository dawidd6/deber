package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/util"
	"github.com/spf13/cobra"
	"os"
	"pault.ag/go/debian/changelog"
)

var a = &app.App{
	Name:        "deber",
	Version:     "0.5",
	Description: "Debian packaging with Docker.",
}

var (
	err error

	cmdRoot = &cobra.Command{
		Use:     a.Name,
		Version: a.Version,
		Short:   a.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Docker, err = docker.New()
			if err != nil {
				return err
			}

			a.Debian, err = changelog.ParseFileOne(a.Config.Changelog)
			if err != nil {
				return err
			}

			for _, step := range a.Steps() {
				err := step()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmdBuild = &cobra.Command{
		Use:   "build",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Docker, err = docker.New()
			if err != nil {
				return err
			}

			if a.Config.Dist == "" {
				a.Debian, err = changelog.ParseFileOne(a.Config.Changelog)
				if err != nil {
					return err
				}
			} else {
				a.Debian = new(changelog.ChangelogEntry)
				a.Debian.Target = a.Config.Dist
			}

			return a.RunBuild()
		},
	}

	cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Docker, err = docker.New()
			if err != nil {
				return err
			}

			a.Debian, err = changelog.ParseFileOne(a.Config.Changelog)
			if err != nil {
				return err
			}

			err = a.RunCreate()
			if err != nil {
				return err
			}

			if a.Config.Start {
				err = a.RunStart()
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmdRemove = &cobra.Command{
		Use:   "remove",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Docker, err = docker.New()
			if err != nil {
				return err
			}

			a.Debian, err = changelog.ParseFileOne(a.Config.Changelog)
			if err != nil {
				return err
			}

			if a.Config.Stop {
				err = a.RunStop()
				if err != nil {
					return err
				}
			}

			err = a.RunRemove()
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmdShell = &cobra.Command{
		Use:   "shell",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Docker, err = docker.New()
			if err != nil {
				return err
			}

			a.Debian, err = changelog.ParseFileOne(a.Config.Changelog)
			if err != nil {
				return err
			}

			return a.RunShellOptional()
		},
	}

	cmdList = &cobra.Command{
		Use:   "list",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Docker, err = docker.New()
			if err != nil {
				return err
			}

			images, err := a.Docker.ImageList(a.Name)
			if err != nil {
				return err
			}

			containers, err := a.Docker.ContainerList(a.Name)
			if err != nil {
				return err
			}

			fmt.Println("Images:")
			for i := range images {
				fmt.Printf("  - %s\n", images[i])
			}

			fmt.Println("Containers:")
			for i := range containers {
				fmt.Printf("  - %s\n", containers[i])
			}

			fmt.Println("Packages:")
			return util.Walk(a.Config.ArchiveBaseDir, 3, func(node util.Node) bool {
				indent := ""
				for i := 0; i < node.Depth; i++ {
					indent += "  "
				}

				fmt.Printf("%s- %s\n", indent, node.Base())

				return false
			})
		},
	}
)

func main() {
	a.Configure()

	cmdRoot.Flags().StringVar(&a.Config.DpkgFlags, "dpkg-flags", a.Config.DpkgFlags, "")
	cmdRoot.Flags().StringVar(&a.Config.LintianFlags, "lintian-flags", a.Config.LintianFlags, "")
	cmdRoot.Flags().StringArrayVar(&a.Config.ExtraPackages, "extra-package", a.Config.ExtraPackages, "")
	cmdRoot.Flags().StringVar(&a.Config.ArchiveBaseDir, "archive-base-dir", a.Config.ArchiveBaseDir, "")
	cmdRoot.Flags().StringVar(&a.Config.CacheBaseDir, "cache-base-dir", a.Config.CacheBaseDir, "")
	cmdRoot.Flags().StringVar(&a.Config.BuildBaseDir, "build-base-dir", a.Config.BuildBaseDir, "")
	cmdRoot.Flags().BoolVar(&a.Config.LogNoColor, "log-no-color", a.Config.LogNoColor, "")

	cmdBuild.Flags().BoolVarP(&a.Config.NoRebuild, "no-rebuild", "n", a.Config.NoRebuild, "")
	cmdBuild.Flags().StringVarP(&a.Config.Dist, "distribution", "d", a.Config.Dist, "")

	cmdCreate.Flags().BoolVarP(&a.Config.Start, "start", "s", a.Config.Start, "")
	cmdCreate.Flags().StringArrayVar(&a.Config.ExtraPackages, "extra-package", a.Config.ExtraPackages, "")

	cmdRemove.Flags().BoolVarP(&a.Config.Stop, "stop", "s", a.Config.Stop, "")
	cmdList.Flags().StringVar(&a.Config.ArchiveBaseDir, "archive-base-dir", a.Config.ArchiveBaseDir, "")

	cmdRoot.Flags().SortFlags = false
	cmdRoot.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmdRoot.SilenceErrors = true
	cmdRoot.SilenceUsage = true
	cmdRoot.AddCommand(
		cmdBuild,
		cmdCreate,
		cmdRemove,
		cmdShell,
		cmdList,
	)

	err := cmdRoot.Execute()
	if err != nil {
		a.LogError(err)
		os.Exit(1)
	}
}
