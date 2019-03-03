package app

import (
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
)

func runCreate(cmd *cobra.Command, args []string) error {
	logger.Info("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}
	if isContainerCreated {
		logger.Skip()
		return nil
	}

	err = dock.CreateContainer(names.Container(), names.Image(), names.BuildDir(), deb.Tarball)
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
