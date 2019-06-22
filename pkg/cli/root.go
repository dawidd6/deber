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
	flagPackageNoTest  bool
	flagKeepContainer  bool
	flagStopContainer  bool
	flagStartContainer bool
	flagDistribution   string
)

var cmdRoot = &cobra.Command{
	Use:               app.Name,
	Version:           app.Version,
	Short:             app.Description,
	PersistentPreRunE: preRoot,
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
	cmdRoot.Flags().BoolVarP(&flagKeepContainer, "keep-container", "k", flagKeepContainer, "")

	cmdRoot.Flags().SortFlags = false
	cmdRoot.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmdRoot.SilenceErrors = true
	cmdRoot.SilenceUsage = true
}

func preRoot(cmd *cobra.Command, args []string) error {
	dock, err = docker.New()
	if err != nil {
		return err
	}

	if flagDistribution == "" {
		deb, err = debian.New()
		if err != nil {
			return err
		}
	} else {
		deb = &debian.Debian{
			Target: flagDistribution,
		}
	}

	n = naming.New(deb)

	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	err = runImageBuild(cmd, args)
	if err != nil {
		return err
	}

	flagStartContainer = true
	err = runContainerCreate(cmd, args)
	if err != nil {
		return err
	}

	err = runPackageDepends(cmd, args)
	if err != nil {
		return err
	}

	err = runPackageBuild(cmd, args)
	if err != nil {
		return err
	}

	err = runArchiveCopy(cmd, args)
	if err != nil {
		return err
	}

	if !flagKeepContainer {
		flagStopContainer = true
		err = runContainerRemove(cmd, args)
		if err != nil {
			return err
		}
	}

	return nil
}
