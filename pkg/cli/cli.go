// Package cli is the core of deber.
package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
)

// Commands
var (
	cmdRoot = &cobra.Command{
		Use:     app.Name,
		Version: app.Version,
		Short:   app.Description,
		RunE:    run,
	}

	cmdList = &cobra.Command{
		Use:   "list",
		Short: "",
		RunE:  runList,
	}

	cmdShell = &cobra.Command{
		Use:   "shell",
		Short: "",
		RunE:  runShell,
	}
)

// Flags
var (
	// Root flags
	flagDpkgFlags     string
	flagLintianFlags  string
	flagWithNetwork   bool
	flagExtraPackages []string

	// List flags
	flagImages     bool
	flagContainers bool
	flagPackages   bool

	// Root-Global flags
	flagNoColor bool
)

func init() {
	cmdRoot.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmdRoot.Flags().SortFlags = false
	cmdRoot.SilenceErrors = true
	cmdRoot.SilenceUsage = true

	cmdRoot.AddCommand(
		cmdList,
		cmdShell,
	)

	// Root flags
	cmdRoot.Flags().StringVarP(&flagDpkgFlags, "dpkg-flags", "d", "-tc", "")
	cmdRoot.Flags().StringVarP(&flagLintianFlags, "lintian-flags", "l", "-i -I", "")
	cmdRoot.Flags().BoolVarP(&flagWithNetwork, "with-network", "n", false, "")
	cmdRoot.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "")
	cmdRoot.Flags().StringArrayVarP(&flagExtraPackages, "extra-package", "p", nil, "")

	// List flags
	cmdList.Flags().BoolVarP(&flagImages, "images", "i", false, "")
	cmdList.Flags().BoolVarP(&flagContainers, "containers", "c", false, "")
	cmdList.Flags().BoolVarP(&flagPackages, "packages", "p", false, "")
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

func run(cmd *cobra.Command, args []string) error {
	dock, err := docker.New()
	if err != nil {
		return err
	}

	deb, err := changelog.ParseFileOne("debian/changelog")
	if err != nil {
		return err
	}

	name := naming.New(deb)

	opts := &steps.Options{
		Naming:        name,
		DpkgFlags:     flagDpkgFlags,
		LintianFlags:  flagLintianFlags,
		Network:       flagWithNetwork,
		ExtraPackages: flagExtraPackages,
	}

	functions := []func(*docker.Docker, *steps.Options) error{
		steps.Build,
		steps.Create,
		steps.Package,
		steps.Archive,
		steps.Remove,
	}

	for _, fn := range functions {
		err := fn(dock, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func runShell(cmd *cobra.Command, args []string) error {
	dock, err := docker.New()
	if err != nil {
		return err
	}

	deb, err := changelog.ParseFileOne("debian/changelog")
	if err != nil {
		return err
	}

	name := naming.New(deb)

	containerArgs := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        name.Container.Name(),
	}

	err = dock.ContainerExec(containerArgs)
	if err != nil {
		return err
	}

	return err
}

func runList(cmd *cobra.Command, args []string) error {
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
		fmt.Println("Images:")

		images, err := dock.ImageList(app.Name)
		if err != nil {
			return err
		}

		for _, image := range images {
			fmt.Printf("  - %s\n", image)
		}
	}

	if flagContainers {
		fmt.Println("Containers:")

		containers, err := dock.ContainerList(app.Name)
		if err != nil {
			return err
		}

		for _, container := range containers {
			fmt.Printf("  - %s\n", container)
		}
	}

	if flagPackages {
		fmt.Println("Packages:")

		err := walk.Walk(naming.ArchiveBase, 3, func(node walk.Node) {
			indent := ""

			for i := 0; i < node.Depth; i++ {
				indent += "  "
			}

			fmt.Printf("%s- %s\n", indent, filepath.Base(node.Path))
		})

		if err != nil {
			return err
		}
	}

	return nil
}
