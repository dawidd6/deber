package cli

import (
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	stop bool
)

var cmdRemove = &cobra.Command{
	Use:   "remove",
	Short: "",
	RunE:  runRemove,
}

func init() {
	cmdRoot.AddCommand(cmdRemove)

	cmdRemove.Flags().BoolVarP(&stop, "stop", "s", stop, "")
}

func runRemove(cmd *cobra.Command, args []string) error {
	if stop {
		err := steps.Stop(dock, deb, n)
		if err != nil {
			return err
		}
	}

	err := steps.Remove(dock, deb, n)
	if err != nil {
		return err
	}

	return nil
}
