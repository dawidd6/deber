package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"syscall"
)

const (
	Program = "deber"
	Version = "0.0+git"
)

var (
	debian       *Debian
	docker       *Docker
	names        *Naming
	verbose      bool
	withSteps    string
	withoutSteps string
	cmd          = &cobra.Command{
		Use:               fmt.Sprintf("%s OS DIST", Program),
		Version:           Version,
		Short:             "Debian packaging with Docker",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: pre,
		RunE:              run,
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
}

func main() {
	if err := cmd.Execute(); err != nil {
		LogError(err)
		syscall.Exit(1)
	}
}
