package main

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/env"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/spf13/cobra"
	"os"
)

var (
	start bool
	stop  bool
	dist  string
)

var (
	cmdRoot = &cobra.Command{
		Use:     app.Name,
		Version: app.Version,
		Short:   app.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := debian.New()
			if err != nil {
				return err
			}

			n := naming.New(deb)

			for _, step := range steps.Steps() {
				err := step(dock, deb, n)
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
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb := new(debian.Debian)

			if dist == "" {
				deb, err = debian.New()
				if err != nil {
					return err
				}
			} else {
				deb.Target = dist
			}

			n := naming.New(deb)

			return steps.RunBuild(dock, deb, n)
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

			deb, err := debian.New()
			if err != nil {
				return err
			}

			n := naming.New(deb)

			err = steps.RunCreate(dock, deb, n)
			if err != nil {
				return err
			}

			if start {
				err = steps.RunStart(dock, deb, n)
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
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := debian.New()
			if err != nil {
				return err
			}

			n := naming.New(deb)

			if stop {
				err = steps.RunStop(dock, deb, n)
				if err != nil {
					return err
				}
			}

			err = steps.RunRemove(dock, deb, n)
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
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := debian.New()
			if err != nil {
				return err
			}

			n := naming.New(deb)

			return steps.RunShellOptional(dock, deb, n)
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

			images, err := dock.ImageList(app.Name)
			if err != nil {
				return err
			}

			containers, err := dock.ContainerList(app.Name)
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
			return walk.Walk(naming.ArchiveBaseDir, 3, func(node walk.Node) bool {
				indent := ""
				for i := 0; i < node.Depth; i++ {
					indent += "  "
				}

				fmt.Printf("%s- %s\n", indent, node.Base())

				return false
			})
		},
	}

	cmdInfo = &cobra.Command{
		Use:   "info",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			dock, err := docker.New()
			if err != nil {
				return err
			}

			deb, err := debian.New()
			if err != nil {
				return err
			}

			n := naming.New(deb)

			return steps.RunInfoOptional(dock, deb, n)
		},
	}
)

func main() {
	steps.DpkgFlags = env.Get("DPKG_FLAGS", steps.DpkgFlags)
	steps.LintianFlags = env.Get("LINTIAN_FLAGS", steps.LintianFlags)

	cmdRoot.Flags().StringVar(&steps.DpkgFlags, "dpkg-flags", steps.DpkgFlags, "")
	cmdRoot.Flags().StringVar(&steps.LintianFlags, "lintian-flags", steps.LintianFlags, "")
	cmdRoot.Flags().StringArrayVar(&steps.ExtraPackages, "extra-package", steps.ExtraPackages, "")
	cmdRoot.Flags().StringVar(&naming.ArchiveBaseDir, "archive-base-dir", naming.ArchiveBaseDir, "")
	cmdRoot.Flags().StringVar(&naming.CacheBaseDir, "cache-base-dir", naming.CacheBaseDir, "")
	cmdRoot.Flags().StringVar(&naming.BuildBaseDir, "build-base-dir", naming.BuildBaseDir, "")
	cmdRoot.Flags().BoolVar(&log.NoColor, "log-no-color", log.NoColor, "")
	cmdRoot.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "n", steps.NoRebuild, "")

	cmdBuild.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "n", steps.NoRebuild, "")
	cmdBuild.Flags().StringVarP(&dist, "distribution", "d", dist, "")

	cmdCreate.Flags().BoolVarP(&start, "start", "s", start, "")
	cmdCreate.Flags().StringArrayVar(&steps.ExtraPackages, "extra-package", steps.ExtraPackages, "")

	cmdRemove.Flags().BoolVarP(&stop, "stop", "s", stop, "")

	cmdList.Flags().StringVar(&naming.ArchiveBaseDir, "archive-base-dir", naming.ArchiveBaseDir, "")

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
		cmdInfo,
	)

	err := cmdRoot.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
