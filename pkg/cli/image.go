package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	flagImageBuild bool
	flagImageList  bool
	flagImageDist  string
)

var (
	cmdImage = &cobra.Command{
		Use:   "image",
		Short: "",
		RunE:  runImage,
	}
)

func init() {
	cmdRoot.AddCommand(
		cmdImage,
	)

	cmdImage.Flags().StringVarP(&flagImageDist, "distribution", "d", flagImageDist, "")
	cmdImage.Flags().BoolVarP(&flagImageBuild, "build", "b", flagImageBuild, "")
	cmdImage.Flags().BoolVarP(&flagImageList, "list", "l", flagImageList, "")
}

func runImage(cmd *cobra.Command, args []string) error {
	flag := false

	if flagImageList {
		flag = true

		images, err := dock.ImageList(app.Name)
		if err != nil {
			return err
		}

		for i := range images {
			fmt.Println(images[i])
		}
	}

	if flagImageBuild {
		flag = true

		err = steps.Build(dock, deb, n)
		if err != nil {
			return err
		}
	}

	if flag {
		return nil
	}

	return cmd.Help()
}
