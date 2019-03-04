package app

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/constants"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"strings"
	"syscall"
)

var cmd = &cobra.Command{
	Use:               fmt.Sprintf("%s OS DIST", constants.Program),
	Version:           constants.Version,
	Short:             "Debian packaging with Docker",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              run,
}

var (
	deb          *debian.Debian
	dock         *docker.Docker
	names        *naming.Naming
	verbose      bool
	withSteps    string
	withoutSteps string
	steps        = []Step{
		{
			label: "build",
			run:   runBuild,
		}, {
			label: "create",
			run:   runCreate,
		}, {
			label: "start",
			run:   runStart,
		}, {
			label: "package",
			run:   runPackage,
		}, {
			label: "test",
			run:   runTest,
		}, {
			label: "stop",
			run:   runStop,
		}, {
			label: "remove",
			run:   runRemove,
		},
	}
)

type Step struct {
	label    string
	run      func(string, string) error
	disabled bool
}

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

func pre(cmd *cobra.Command, args []string) error {
	var err error

	logger.Info("Parsing Debian changelog")
	deb, err = debian.New()
	if err != nil {
		logger.Fail()
		return err
	}
	logger.Done()

	logger.Info("Connecting with Docker")
	dock, err = docker.New(verbose)
	if err != nil {
		logger.Fail()
		return err
	}
	logger.Done()

	names = naming.New(args[0], args[1], deb)

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	if withSteps != "" && withoutSteps != "" {
		return errors.New("can't specify with and without steps together")
	}

	if withSteps != "" {
		for i := range steps {
			if !strings.Contains(withSteps, steps[i].label) {
				steps[i].disabled = true
			}
		}
	}

	if withoutSteps != "" {
		for i := range steps {
			if strings.Contains(withoutSteps, steps[i].label) {
				steps[i].disabled = true
			}
		}
	}

	for i := range steps {
		if !steps[i].disabled {
			err := steps[i].run(args[0], args[1])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Execute() {
	err := cmd.Execute()
	if err != nil {
		logger.Error(err)
		syscall.Exit(1)
	}
}
