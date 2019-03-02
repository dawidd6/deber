package commands

import (
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
)

var cmdRemove = &cobra.Command{
	Use:               "remove OS DIST",
	Short:             "remove Docker container",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	logger.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}
	if !isContainerCreated {
		logger.Skip()
		return nil
	}

	err = dock.RemoveContainer(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
