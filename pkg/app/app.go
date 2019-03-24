package app

import (
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	program string

	update       time.Duration
	repo         string
	from         string
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
	cmd.Flags().StringVarP(
		&repo,
		"repo",
		"r",
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
		&from,
		"from",
		"f",
		"debian:unstable",
		"specify which Docker image to use",
	)
	cmd.Flags().DurationVarP(
		&update,
		"update-after",
		"u",
		time.Minute*30,
		"perform apt cache update after specified interval",
	)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		logError(err)
		os.Exit(1)
	}
}
