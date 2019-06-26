package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	flagImageBuild  bool
	flagImageList   bool
	flagImageRemove bool
	flagImagePrune  bool
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
	cmdImage.Flags().BoolVar(&flagImageRemove, "remove", flagImageRemove, "")
	cmdImage.Flags().BoolVar(&flagImagePrune, "prune", flagImagePrune, "")
}

func runImage(cmd *cobra.Command, args []string) error {
	if flagImagePrune {
		images, err := docker.ImageList(Prefix)
		if err != nil {
			return err
		}

		for i := range images {
			err = docker.ImageRemove(images[i])
			if err != nil {
				return err
			}
		}
	}

	if flagImageRemove {
		err := docker.ImageRemove(naming.Image())
		if err != nil {
			return err
		}
	}

	if flagImageList {
		images, err := docker.ImageList(Prefix)
		if err != nil {
			return err
		}

		for i := range images {
			fmt.Println(images[i])
		}
	}

	if flagImageBuild {
		err := steps.Build()
		if err != nil {
			return err
		}
	}

	if cmd.Flags().NFlag() > 0 {
		return nil
	}

	return cmd.Help()
}
