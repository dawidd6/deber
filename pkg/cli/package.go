package cli

import (
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	flagPackageBuild   bool
	flagPackageTest    bool
	flagPackageInfo    bool
	flagPackageDepends bool
)

var (
	cmdPackage = &cobra.Command{
		Use:   "package",
		Short: "",
		RunE:  runPackage,
	}
)

func init() {
	cmdRoot.AddCommand(
		cmdPackage,
	)

	cmdPackage.Flags().BoolVarP(&flagPackageBuild, "build", "b", flagPackageBuild, "")
	cmdPackage.Flags().BoolVarP(&flagPackageTest, "test", "t", flagPackageTest, "")
	cmdPackage.Flags().BoolVarP(&flagPackageDepends, "depends", "d", flagPackageDepends, "")
	cmdPackage.Flags().BoolVarP(&flagPackageInfo, "info", "i", flagPackageInfo, "")
}

func runPackage(cmd *cobra.Command, args []string) error {
	flag := false

	if flagPackageInfo {
		flag = true

	}

	if flagPackageDepends {
		flag = true

		err = steps.Depends(dock, deb, n)
		if err != nil {
			return err
		}
	}

	if flagPackageBuild {
		flag = true

		err = steps.Package(dock, deb, n)
		if err != nil {
			return err
		}
	}

	if flagPackageTest {
		flag = true

		err = steps.Test(dock, deb, n)
		if err != nil {
			return err
		}
	}

	if flag {
		return nil
	}

	return cmd.Help()
}
