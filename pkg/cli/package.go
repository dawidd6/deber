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
	if flagPackageInfo {

	}

	if flagPackageDepends {
		err := steps.Depends()
		if err != nil {
			return err
		}
	}

	if flagPackageBuild {
		err := steps.Package()
		if err != nil {
			return err
		}
	}

	if flagPackageTest {
		err := steps.Test()
		if err != nil {
			return err
		}
	}

	if cmd.Flags().NFlag() > 0 {
		return nil
	}

	return cmd.Help()
}
