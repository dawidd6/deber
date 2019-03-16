package app

import (
	"github.com/spf13/cobra"
	"syscall"
)

var (
	program string

	verbose      bool
	network      bool
	showSteps    bool
	withSteps    string
	withoutSteps string
	repo         string
	os           string
	dist         string
	dpkgFlags    string
	lintianFlags string
)

func Run(p, version, description string) {
	program = p

	cmd := &cobra.Command{
		Use:     program,
		Version: version,
		Short:   description,
		RunE:    run,
	}
	cmd.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"show more output",
	)
	cmd.Flags().BoolVarP(
		&network,
		"network",
		"n",
		false,
		"enable network in container during packaging step",
	)
	cmd.Flags().BoolVar(
		&showSteps,
		"show-steps",
		false,
		"show available steps in order")
	cmd.Flags().StringVar(
		&withSteps,
		"with-steps",
		"",
		"specify which of the steps should execute",
	)
	cmd.Flags().StringVar(
		&withoutSteps,
		"without-steps",
		"",
		"specify which of the steps should not execute",
	)
	cmd.Flags().StringVar(
		&repo,
		"repo",
		"",
		"specify a local repository to be mounted in container")
	cmd.Flags().StringVar(
		&dpkgFlags,
		"dpkg-buildpackage-flags",
		"-tc",
		"specify flags passed to dpkg-buildpackage")
	cmd.Flags().StringVar(
		&lintianFlags,
		"lintian-flags",
		"-i",
		"specify flags passed to lintian")
	cmd.Flags().StringVarP(
		&os,
		"os",
		"o",
		"debian",
		"specify which OS to use",
	)
	cmd.Flags().StringVarP(
		&dist,
		"dist",
		"d",
		"unstable",
		"specify which Distribution to use",
	)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		logError(err)
		syscall.Exit(1)
	}
}
