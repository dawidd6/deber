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
	keep bool
)

var cmdRoot = &cobra.Command{
	Use:     app.Name,
	Version: app.Version,
	Short:   app.Description,
	RunE:    runRoot,
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
	cmdRoot.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "n", steps.NoRebuild, "")
	cmdRoot.Flags().BoolVarP(&keep, "keep-container", "k", false, "")

	cmdRoot.Flags().SortFlags = false
	cmdRoot.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmdRoot.SilenceErrors = true
	cmdRoot.SilenceUsage = true
}

func runRoot(cmd *cobra.Command, args []string) error {
	dock, err := docker.New()
	if err != nil {
		return err
	}

	deb, err := debian.New()
	if err != nil {
		return err
	}

	n := naming.New(deb)
	s := steps.Steps()

	// Don't remove container at the end.
	//
	// Remove step should always be the last.
	if keep {
		s = s[:len(s)-1]
	}

	for _, step := range s {
		err := step(dock, deb, n)
		if err != nil {
			return err
		}
	}

	return nil
}
