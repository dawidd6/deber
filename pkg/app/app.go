package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"syscall"
)

var (
	program string

	verboseFlag      bool
	showStepsFlag    bool
	withStepsFlag    string
	withoutStepsFlag string
	dpkgOptions      []string
)

func Run(p, version, description, example string) {
	program = p

	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s OS DIST [flags] [-- dpkg-buildpackage options]", program),
		Version: version,
		Short:   description,
		Example: example,
		RunE:    run,
	}
	cmd.Flags().BoolVarP(
		&verboseFlag,
		"verbose",
		"v",
		false,
		"show more output",
	)
	cmd.Flags().BoolVar(
		&showStepsFlag,
		"show-steps",
		false,
		"show available steps in order")
	cmd.Flags().StringVar(
		&withStepsFlag,
		"with-steps",
		"",
		"specify which of the steps should execute",
	)
	cmd.Flags().StringVar(
		&withoutStepsFlag,
		"without-steps",
		"",
		"specify which of the steps should not execute",
	)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.DisableFlagsInUseLine = true

	if err := cmd.Execute(); err != nil {
		logError(err)
		syscall.Exit(1)
	}
}
