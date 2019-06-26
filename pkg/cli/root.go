package cli

import (
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/env"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
	"os"
	"pault.ag/go/debian/changelog"
)

var Prefix string

var (
	check         bool
	info          bool
	keepContainer bool
	distribution  string
)

var cmdRoot = &cobra.Command{
	PersistentPreRunE: preRoot,
	RunE:              runRoot,
}

func Run(program, version, desc string) {
	cmdRoot.Use = program
	cmdRoot.Version = version
	cmdRoot.Short = desc

	err := cmdRoot.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	steps.DpkgFlags = env.Get("DPKG_FLAGS", steps.DpkgFlags)
	steps.LintianFlags = env.Get("LINTIAN_FLAGS", steps.LintianFlags)

	cmdRoot.PersistentFlags().BoolVar(&log.NoColor, "log-no-color", log.NoColor, "")

	cmdRoot.Flags().StringVar(&steps.DpkgFlags, "dpkg-flags", steps.DpkgFlags, "")
	cmdRoot.Flags().StringVar(&steps.LintianFlags, "lintian-flags", steps.LintianFlags, "")
	cmdPackage.Flags().StringVar(&steps.DpkgFlags, "dpkg-flags", steps.DpkgFlags, "")
	cmdPackage.Flags().StringVar(&steps.LintianFlags, "lintian-flags", steps.LintianFlags, "")

	cmdRoot.Flags().StringArrayVarP(&steps.ExtraPackages, "extra-package", "e", steps.ExtraPackages, "")
	cmdContainer.Flags().StringArrayVarP(&steps.ExtraPackages, "extra-package", "e", steps.ExtraPackages, "")

	cmdRoot.Flags().StringVar(&naming.ArchiveBaseDir, "archive-base-dir", naming.ArchiveBaseDir, "")
	cmdRoot.Flags().StringVar(&naming.CacheBaseDir, "cache-base-dir", naming.CacheBaseDir, "")
	cmdRoot.Flags().StringVar(&naming.BuildBaseDir, "build-base-dir", naming.BuildBaseDir, "")
	cmdContainer.Flags().StringVar(&naming.ArchiveBaseDir, "archive-base-dir", naming.ArchiveBaseDir, "")
	cmdContainer.Flags().StringVar(&naming.CacheBaseDir, "cache-base-dir", naming.CacheBaseDir, "")
	cmdContainer.Flags().StringVar(&naming.BuildBaseDir, "build-base-dir", naming.BuildBaseDir, "")
	cmdArchive.Flags().StringVar(&naming.ArchiveBaseDir, "archive-base-dir", naming.ArchiveBaseDir, "")

	cmdRoot.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "r", steps.NoRebuild, "")
	cmdImage.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "r", steps.NoRebuild, "")

	cmdRoot.Flags().BoolVarP(&steps.NoUpdate, "no-update", "u", steps.NoUpdate, "")
	cmdPackage.Flags().BoolVarP(&steps.NoUpdate, "no-update", "u", steps.NoUpdate, "")

	cmdRoot.Flags().BoolVarP(&steps.WithNetwork, "with-network", "n", steps.WithNetwork, "")
	cmdPackage.Flags().BoolVarP(&steps.WithNetwork, "with-network", "n", steps.WithNetwork, "")

	cmdRoot.Flags().BoolVarP(&keepContainer, "keep-container", "k", keepContainer, "")
	cmdRoot.Flags().BoolVarP(&check, "check", "c", check, "")
	cmdRoot.Flags().BoolVarP(&info, "info", "i", info, "")

	cmdRoot.Flags().StringVarP(&distribution, "distribution", "d", distribution, "")
	cmdImage.Flags().StringVarP(&distribution, "distribution", "d", distribution, "")
	cmdContainer.Flags().StringVarP(&distribution, "distribution", "d", distribution, "")
	cmdPackage.Flags().StringVarP(&distribution, "distribution", "d", distribution, "")

	cmdRoot.AddCommand(
		cmdArchive,
		cmdContainer,
		cmdImage,
		cmdPackage,
	)

	cmdRoot.Flags().SortFlags = false
	cmdRoot.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmdRoot.SilenceErrors = true
	cmdRoot.SilenceUsage = true
}

func preRoot(cmd *cobra.Command, args []string) error {
	err := docker.New()
	if err != nil {
		return err
	}

	debian, err := changelog.ParseFileOne("debian/changelog")
	if err != nil {
		return err
	}

	naming.PackageName = debian.Source
	naming.PackageVersion = debian.Version.String()
	naming.PackageUpstream = debian.Version.Version
	naming.PackageTarget = debian.Target

	if distribution != "" {
		naming.PackageTarget = distribution
	}

	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	err := steps.Build()
	if err != nil {
		return err
	}

	err = steps.Create()
	if err != nil {
		return err
	}

	err = steps.Start()
	if err != nil {
		return err
	}

	err = steps.Depends()
	if err != nil {
		return err
	}

	err = steps.Package()
	if err != nil {
		return err
	}

	err = steps.Test()
	if err != nil {
		return err
	}

	err = steps.Archive()
	if err != nil {
		return err
	}

	err = steps.Stop()
	if err != nil {
		return err
	}

	if !keepContainer {
		err = steps.Remove()
		if err != nil {
			return err
		}
	}

	return nil
}
