package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	flagImageBuild bool
	flagImageList  bool
)

var (
	cmdImage = &cobra.Command{
		Use:   "image",
		Short: "",
		RunE:  runImage,
	}
)

func init() {
	cmdImage.Flags().BoolVar(&flagImageBuild, "build", flagImageBuild, "")
	cmdImage.Flags().BoolVar(&flagImageList, "list", flagImageList, "")
}

func runImage(cmd *cobra.Command, args []string) error {
	flag := false

	if flagImageList {
		flag = true

		images, err := docker.ImageList(app.Name)
		if err != nil {
			return err
		}

		for i := range images {
			fmt.Println(images[i])
		}
	}

	if flagImageBuild {
		flag = true

		err := steps.Build()
		if err != nil {
			return err
		}
	}

	if flag {
		return nil
	}

	return cmd.Help()
}
