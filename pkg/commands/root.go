package commands

import (
	"deber/pkg/constants"
	"deber/pkg/logger"
	"fmt"
	"github.com/spf13/cobra"
	"syscall"
)

var cmd = &cobra.Command{
	Use:               fmt.Sprintf("%s OS DIST", constants.Program),
	Version:           constants.Version,
	Short:             "Debian packaging with Docker",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runRoot,
}

var verbose bool

func init() {
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show more output")
	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.AddCommand(
		cmdBuild,
		cmdCreate,
		cmdStart,
		cmdStop,
		cmdRemove,
		cmdPackage,
		cmdTest,
	)
}

func runRoot(cmd *cobra.Command, args []string) error {
	var err error

	err = runBuild(cmd, args)
	if err != nil {
		return err
	}

	err = runCreate(cmd, args)
	if err != nil {
		return err
	}

	err = runStart(cmd, args)
	if err != nil {
		return err
	}

	err = runPackage(cmd, args)
	if err != nil {
		return err
	}

	err = runTest(cmd, args)
	if err != nil {
		return err
	}

	err = runStop(cmd, args)
	if err != nil {
		return err
	}

	err = runRemove(cmd, args)
	if err != nil {
		return err
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
