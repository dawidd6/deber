package commands

import (
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
)

var cmdTest = &cobra.Command{
	Use:               "test OS DIST",
	Short:             "test package in Docker container",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	logger.Info("Testing package")

	err := dock.ExecContainer(names.Container(), "sudo", "debi")
	if err != nil {
		logger.Fail()
		return err
	}

	err = dock.ExecContainer(names.Container(), "debc")
	if err != nil {
		logger.Fail()
		return err
	}

	err = dock.ExecContainer(names.Container(), "lintian")
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
