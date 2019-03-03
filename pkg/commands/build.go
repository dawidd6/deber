package commands

import (
	"errors"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
)

var cmdBuild = &cobra.Command{
	Use:               "build OS DIST",
	Short:             "build Docker image",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {
	logger.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(names.Image())
	if err != nil {
		logger.Fail()
		return err
	}
	if isImageBuilt {
		logger.Skip()
		return nil
	}

	dockerfile, err := docker.GetDockerfile(args[0], args[1])
	if err != nil {
		logger.Fail()
		return err
	}

	err = dock.BuildImage(names.Image(), dockerfile)
	if err != nil {
		logger.Fail()
		return err
	}

	isImageBuilt, err = dock.IsImageBuilt(names.Image())
	if err != nil {
		logger.Fail()
		return err
	}

	if !isImageBuilt {
		logger.Fail()
		return errors.New("image didn't built successfully")
	}

	logger.Done()
	return nil
}
