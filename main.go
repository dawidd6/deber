package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"syscall"
)

const (
	program = "deber"
	version = "0.0+git"
)

const desc = `Debian packaging with Docker`

const example = `  basic:
    deber ubuntu xenial

  only with needed steps:
    deber ubuntu bionic --with-steps build
    deber debian buster --with-steps build,create

  without unneeded steps:
    deber debian unstable --without-steps remove,stop,build

  with gbp:
    gbp buildpackage --git-builder=deber ubuntu xenial

  with dpkg-buildpackage options
    deber ubuntu xenial -- -nc -b`

var (
	debian       *Debian
	docker       *Docker
	names        *Naming
	verbose      bool
	stepping     bool
	withSteps    string
	withoutSteps string
	dpkgOptions  []string
	cmd          = &cobra.Command{
		Use:     fmt.Sprintf("%s OS DIST [flags] [-- dpkg-buildpackage options]", program),
		Version: version,
		Short:   desc,
		Example: example,
		RunE:    run,
	}
)

func init() {
	cmd.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"show more output",
	)
	cmd.Flags().BoolVar(
		&stepping,
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
	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.DisableFlagsInUseLine = true
}

func main() {
	if err := cmd.Execute(); err != nil {
		LogError(err)
		syscall.Exit(1)
	}
}
