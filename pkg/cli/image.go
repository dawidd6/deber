package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	cmdImage = &cobra.Command{
		Use:   "image",
		Short: "",
		RunE:  runImage,
	}

	cmdImageBuild = &cobra.Command{
		Use:   "build",
		Short: "",
		RunE:  runImageBuild,
	}

	cmdImageList = &cobra.Command{
		Use:   "list",
		Short: "",
		RunE:  runImageList,
	}
)

func init() {
	cmdRoot.AddCommand(
		cmdImage,
	)
	cmdImage.AddCommand(
		cmdImageBuild,
		cmdImageList,
	)

	cmdImageBuild.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "n", steps.NoRebuild, "")
	cmdImageBuild.Flags().StringVarP(&flagDistribution, "distribution", "d", flagDistribution, "")
}

func runImage(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func runImageBuild(cmd *cobra.Command, args []string) error {
	return steps.Build(dock, deb, n)
}

func runImageList(cmd *cobra.Command, args []string) error {
	images, err := dock.ImageList(app.Name)
	if err != nil {
		return err
	}

	for i := range images {
		fmt.Println(images[i])
	}

	return nil
}
