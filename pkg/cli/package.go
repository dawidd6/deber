package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

// TODO options instead of subcommands (info, depends, ...)

var (
	cmdPackage = &cobra.Command{
		Use:   "package",
		Short: "",
		RunE:  runPackage,
	}

	cmdPackageDepends = &cobra.Command{
		Use:   "depends",
		Short: "",
		RunE:  runPackageDepends,
	}

	cmdPackageBuild = &cobra.Command{
		Use:   "build",
		Short: "",
		RunE:  runPackageBuild,
	}

	cmdPackageInfo = &cobra.Command{
		Use:   "info",
		Short: "",
		RunE:  runPackageInfo,
	}
)

func init() {
	cmdRoot.AddCommand(
		cmdPackage,
	)
	cmdPackage.AddCommand(
		cmdPackageDepends,
		cmdPackageBuild,
		cmdPackageInfo,
	)

	cmdPackageBuild.Flags().BoolVarP(&flagPackageNoTest, "no-test", "n", flagPackageNoTest, "")
}

func runPackage(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func runPackageDepends(cmd *cobra.Command, args []string) error {
	return steps.Depends(dock, deb, n)
}

func runPackageBuild(cmd *cobra.Command, args []string) error {
	err := steps.Package(dock, deb, n)
	if err != nil {
		return err
	}

	if flagPackageNoTest {
		return nil
	}

	return steps.Test(dock, deb, n)
}

func runPackageInfo(cmd *cobra.Command, args []string) error {
	fmt.Println(*deb)

	return nil
}
