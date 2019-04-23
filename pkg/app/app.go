package app

import (
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
	"os"
)

var (
	log *logger.Logger

	include string
	exclude string

	dpkgFlags    = os.Getenv("DEBER_DPKG_BUILDPACKAGE_FLAGS")
	lintianFlags = os.Getenv("DEBER_LINTIAN_FLAGS")
	archiveDir   = os.Getenv("DEBER_ARCHIVE")
)

func Run(program, version, description string) {
	log = logger.New(program)

	cmd := &cobra.Command{
		Use:     program,
		Version: version,
		Short:   description,
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

	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
