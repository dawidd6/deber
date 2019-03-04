package app

import (
	"errors"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
)

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

	err = dock.BuildImage(names.Image(), args[0]+":"+args[1])
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
