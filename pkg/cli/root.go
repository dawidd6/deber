package cli

import (
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/env"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
	"os"
)

var (
	dock *docker.Docker
	deb  *debian.Debian
	n    *naming.Naming
	err  error
)

var (
	keep bool
	// TODO check should be a separate command
	check bool
)

var cmdRoot = &cobra.Command{
	Use:               app.Name,
	Version:           app.Version,
	Short:             app.Description,
	PersistentPreRunE: runPersistentPre,
	RunE:              runRoot,
}

func Run() {
	err := cmdRoot.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	steps.DpkgFlags = env.Get("DPKG_FLAGS", steps.DpkgFlags)
	steps.LintianFlags = env.Get("LINTIAN_FLAGS", steps.LintianFlags)

	cmdRoot.Flags().StringVar(&steps.DpkgFlags, "dpkg-flags", steps.DpkgFlags, "")
	cmdRoot.Flags().StringVar(&steps.LintianFlags, "lintian-flags", steps.LintianFlags, "")
	cmdRoot.Flags().StringArrayVar(&steps.ExtraPackages, "extra-package", steps.ExtraPackages, "")
	cmdRoot.Flags().StringVar(&naming.ArchiveBaseDir, "archive-base-dir", naming.ArchiveBaseDir, "")
	cmdRoot.Flags().StringVar(&naming.CacheBaseDir, "cache-base-dir", naming.CacheBaseDir, "")
	cmdRoot.Flags().StringVar(&naming.BuildBaseDir, "build-base-dir", naming.BuildBaseDir, "")
	cmdRoot.Flags().BoolVar(&log.NoColor, "log-no-color", log.NoColor, "")
	cmdRoot.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "r", steps.NoRebuild, "")
	cmdRoot.Flags().BoolVarP(&steps.NoUpdate, "no-update", "u", steps.NoUpdate, "")
	cmdRoot.Flags().BoolVarP(&steps.WithNetwork, "with-network", "n", steps.WithNetwork, "")
	cmdRoot.Flags().BoolVarP(&keep, "keep-container", "k", false, "")
	cmdRoot.Flags().BoolVarP(&check, "check-before", "c", check, "")

	cmdRoot.Flags().SortFlags = false
	cmdRoot.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmdRoot.SilenceErrors = true
	cmdRoot.SilenceUsage = true
}

func runPersistentPre(cmd *cobra.Command, args []string) error {
	dock, err = docker.New()
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

	n = naming.New(deb)

	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {

	runBuild(cmd, args)
	err = steps.Build(dock, deb, n)
	if err != nil {
		return err
	}

	start = true
	runCreate(cmd, args)
	err = steps.Create(dock, deb, n)
	if err != nil {
		return err
	}

	err = steps.Start(dock, deb, n)
	if err != nil {
		return err
	}

	err = steps.Tarball(dock, deb, n)
	if err != nil {
		return err
	}

	err = steps.Depends(dock, deb, n)
	if err != nil {
		return err
	}

	err = steps.Package(dock, deb, n)
	if err != nil {
		return err
	}

	err = steps.Test(dock, deb, n)
	if err != nil {
		return err
	}

	err = steps.Archive(dock, deb, n)
	if err != nil {
		return err
	}
	err = steps.Stop(dock, deb, n)
	if err != nil {
		return err
	}

	if !keep {
		err = steps.Remove(dock, deb, n)
		if err != nil {
			return err
		}
	}

	return nil
}
