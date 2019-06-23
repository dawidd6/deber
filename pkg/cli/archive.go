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
	cmdRoot.AddCommand(
		cmdArchive,
	)

	cmdArchive.Flags().BoolVarP(&flagArchiveCheck, "check", "c", flagArchiveCheck, "")
	cmdArchive.Flags().BoolVarP(&flagArchiveList, "list", "l", flagArchiveList, "")
	cmdArchive.Flags().BoolVarP(&flagArchiveCopy, "copy", "k", flagArchiveCopy, "")
}

func runArchive(cmd *cobra.Command, args []string) error {
	flag := false

	if flagArchiveList {
		flag = true

		err = walk.Walk(naming.ArchiveBaseDir, 3, func(node *walk.Node) bool {
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
		flag = true

		err = steps.CheckOptional(dock, deb, n)
		if err != nil {
			return err
		}
	}

	if flagArchiveCopy {
		flag = true

		err = steps.Archive(dock, deb, n)
		if err != nil {
			return err
		}
	}

	if flag {
		return nil
	}

	return cmd.Help()
}
