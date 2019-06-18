package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/spf13/cobra"
)

var cmdList = &cobra.Command{
	Use:   "list",
	Short: "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dock, err = docker.New()
		return err
	},
	RunE: runList,
}

func init() {
	cmdRoot.AddCommand(cmdList)

	cmdList.Flags().StringVar(&naming.ArchiveBaseDir, "archive-base-dir", naming.ArchiveBaseDir, "")
}

func runList(cmd *cobra.Command, args []string) error {
	images, err := dock.ImageList(app.Name)
	if err != nil {
		return err
	}

	containers, err := dock.ContainerList(app.Name)
	if err != nil {
		return err
	}

	fmt.Println("Images:")
	for i := range images {
		fmt.Printf("  - %s\n", images[i])
	}

	fmt.Println("Containers:")
	for i := range containers {
		fmt.Printf("  - %s\n", containers[i])
	}

	fmt.Println("Packages:")
	return walk.Walk(naming.ArchiveBaseDir, 3, func(node *walk.Node) bool {
		indent := ""
		for i := 0; i < node.Depth(); i++ {
			indent += "  "
		}

		fmt.Printf("%s- %s\n", indent, node.Name())

		return false
	})
}
