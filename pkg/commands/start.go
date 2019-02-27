package commands

import (
	"deber/pkg/logger"
	"github.com/spf13/cobra"
)

var cmdStart = &cobra.Command{
	Use:               "start OS DIST",
	Short:             "start Docker container",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	logger.Info("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}
	if isContainerStarted {
		logger.Skip()
		return nil
	}

	err = dock.StartContainer(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
