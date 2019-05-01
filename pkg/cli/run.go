package cli

import (
	"github.com/dawidd6/deber/pkg/stepping"

	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string) error {
	steps := stepping.Steps{
		stepCheck,
		stepBuild,
		stepCreate,
		stepStart,
		stepTarball,
		stepUpdate,
		stepDeps,
		stepPackage,
		stepTest,
		stepArchive,
		stepScan,
		stepStop,
		stepRemove,
	}

	err := initOptions(steps)
	if err != nil {
		return err
	}

	err = initToys(cmd.Use)
	if err != nil {
		return err
	}

	err = steps.Run()
	if err != nil {
		return err
	}

	return nil
}
