package app

import (
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
	"os"
)

var (
	log *logger.Logger

	include      string
	exclude      string
	dpkgFlags    string
	lintianFlags string
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
	cmd.Flags().StringVar(
		&dpkgFlags,
		"dpkg-buildpackage-flags",
		"-tc",
		"specify flags passed to dpkg-buildpackage",
	)
	cmd.Flags().StringVar(
		&lintianFlags,
		"lintian-flags",
		"-i",
		"specify flags passed to lintian",
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
