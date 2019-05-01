package cli

import (
	"os"

	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	include string
	exclude string
	shell   bool
	remove  bool
	list    bool

	dpkgFlags    = os.Getenv("DEBER_DPKG_BUILDPACKAGE_FLAGS")
	lintianFlags = os.Getenv("DEBER_LINTIAN_FLAGS")
	archiveDir   = os.Getenv("DEBER_ARCHIVE")
)

func Run(program, version, description, examples string) {
	cmd := &cobra.Command{
		Use:     program,
		Version: version,
		Short:   description,
		Example: examples,
		RunE:    run,
	}
	cmd.Flags().StringVarP(
		&include,
		"include",
		"i",
		"",
		"which steps to run only",
	)
	cmd.Flags().StringVarP(
		&exclude,
		"exclude",
		"e",
		"",
		"which steps to exclude from complete set",
	)
	cmd.Flags().BoolVarP(
		&shell,
		"shell",
		"s",
		false,
		"run bash shell interactively in container",
	)
	cmd.Flags().BoolVarP(
		&remove,
		"remove",
		"r",
		false,
		"alias for '--include remove,stop'",
	)
	cmd.Flags().BoolVarP(
		&list,
		"list",
		"l",
		false,
		"list steps in order and exit",
	)

	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err != nil {
		logger.Error(program, err)
		os.Exit(1)
	}
}
