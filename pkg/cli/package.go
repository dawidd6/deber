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
	cmdPackage.Flags().BoolVar(&flagPackageBuild, "build", flagPackageBuild, "")
	cmdPackage.Flags().BoolVar(&flagPackageTest, "test", flagPackageTest, "")
	cmdPackage.Flags().BoolVar(&flagPackageDepends, "depends", flagPackageDepends, "")
	cmdPackage.Flags().BoolVar(&flagPackageInfo, "info", flagPackageInfo, "")
}

func runPackage(cmd *cobra.Command, args []string) error {
	flag := false

	if flagPackageInfo {
		flag = true

	}

	if flagPackageDepends {
		flag = true

		err := steps.Depends()
		if err != nil {
			return err
		}
	}

	if flagPackageBuild {
		flag = true

		err := steps.Package()
		if err != nil {
			return err
		}
	}

	if flagPackageTest {
		flag = true

		err := steps.Test()
		if err != nil {
			return err
		}
	}

	if flag {
		return nil
	}

	return cmd.Help()
}
