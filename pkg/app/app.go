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
	allStepsRuns map[string]func(command *cobra.Command, args []string) error
	allSteps     []string
	withSteps    string
	withoutSteps string
)

func init() {
	allStepsRuns = map[string]func(command *cobra.Command, args []string) error{
		"build":   runBuild,
		"create":  runCreate,
		"start":   runStart,
		"package": runPackage,
		"test":    runTest,
		"stop":    runStop,
		"remove":  runRemove,
	}
	allSteps = []string{
		"build",
		"create",
		"start",
		"package",
		"test",
		"stop",
		"remove",
	}

	cmd.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"show more output",
	)
	cmd.Flags().StringVar(
		&withSteps,
		"with-step",
		"",
		"specify which of the steps should execute",
	)
	cmd.Flags().StringVar(
		&withoutSteps,
		"without-step",
		"",
		"specify which of the steps should not execute",
	)
	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true

	/*cmd.AddCommand(
		cmdBuild,
		cmdCreate,
		cmdStart,
		cmdStop,
		cmdRemove,
		cmdPackage,
		cmdTest,
	)*/
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
		for a := range allSteps {
			if !strings.Contains(withSteps, allSteps[a]) {
				delete(allStepsRuns, allSteps[a])
				allSteps[a] = ""
			}
		}
	}

	if withoutSteps != "" {
		for a := range allSteps {
			if strings.Contains(withoutSteps, allSteps[a]) {
				delete(allStepsRuns, allSteps[a])
				allSteps[a] = ""
			}
		}
	}

	for _, a := range allSteps {
		if a == "" {
			continue
		}

		err := allStepsRuns[a](cmd, args)
		if err != nil {
			return err
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
