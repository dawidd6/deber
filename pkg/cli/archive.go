package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/spf13/cobra"
)

var (
	flagArchiveCheck bool
	flagArchiveList  bool
	flagArchiveCopy  bool
)

var (
	cmdArchive = &cobra.Command{
		Use:   "archive",
		Short: "",
		RunE:  runArchive,
	}
)

func init() {
	cmdArchive.Flags().BoolVar(&flagArchiveCheck, "check", flagArchiveCheck, "")
	cmdArchive.Flags().BoolVar(&flagArchiveList, "list", flagArchiveList, "")
	cmdArchive.Flags().BoolVar(&flagArchiveCopy, "copy", flagArchiveCopy, "")
}

func runArchive(cmd *cobra.Command, args []string) error {
	if flagArchiveList {
		err := walk.Walk(naming.ArchiveBaseDir, 3, func(node *walk.Node) bool {
			indent := ""
			for i := 1; i < node.Depth(); i++ {
				indent += "    "
			}

			fmt.Printf("%s%s\n", indent, node.Name())

			return false
		})
		if err != nil {
			return err
		}
	}

	if flagArchiveCheck {
		err := steps.CheckOptional()
		if err != nil {
			return err
		}
	}

	if flagArchiveCopy {
		err := steps.Archive()
		if err != nil {
			return err
		}
	}

	if cmd.Flags().NFlag() > 0 {
		return nil
	}

	return cmd.Help()
}
