package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/spf13/cobra"
)

var (
	cmdArchive = &cobra.Command{
		Use:   "archive",
		Short: "",
		RunE:  runArchive,
	}

	cmdArchiveCheck = &cobra.Command{
		Use:   "check",
		Short: "",
		RunE:  runArchiveCheck,
	}

	cmdArchiveList = &cobra.Command{
		Use:   "list",
		Short: "",
		RunE:  runArchiveList,
	}

	cmdArchiveCopy = &cobra.Command{
		Use:   "copy",
		Short: "",
		RunE:  runArchiveCopy,
	}
)

func init() {
	cmdRoot.AddCommand(
		cmdArchive,
	)
	cmdArchive.AddCommand(
		cmdArchiveCheck,
		cmdArchiveList,
		cmdArchiveCopy,
	)
}

func runArchive(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func runArchiveCheck(cmd *cobra.Command, args []string) error {
	return steps.CheckOptional(dock, deb, n)
}

func runArchiveList(cmd *cobra.Command, args []string) error {
	return walk.Walk(naming.ArchiveBaseDir, 3, func(node *walk.Node) bool {
		indent := ""
		for i := 1; i < node.Depth(); i++ {
			indent += "    "
		}

		fmt.Printf("%s%s\n", indent, node.Name())

		return false
	})
}

func runArchiveCopy(cmd *cobra.Command, args []string) error {
	return steps.Archive(dock, deb, n)
}
