package commands

import (
	"deber/pkg/logger"
	"github.com/spf13/cobra"
)

var cmdStop = &cobra.Command{
	Use:               "stop OS DIST",
	Short:             "stop Docker container",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runStop,
}

func runStop(cmd *cobra.Command, args []string) error {
	logger.Info("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}
	if isContainerStopped {
		logger.Skip()
		return nil
	}

	err = dock.StopContainer(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
