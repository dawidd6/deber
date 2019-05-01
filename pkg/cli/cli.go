package cli

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
	"strings"

	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	include string
	exclude string
	shell   bool
	remove  bool
	list    bool

	dpkgFlags    = os.Getenv("DEBER_DPKG_BUILDPACKAGE_FLAGS")
	lintianFlags = os.Getenv("DEBER_LINTIAN_FLAGS")
	archiveDir   = os.Getenv("DEBER_ARCHIVE")

	deb  *debian.Debian
	dock *docker.Docker
	name *naming.Naming
	log  *logger.Logger
)

func Run(program, version, description, examples string) {
	cmd := &cobra.Command{
		Use:     program,
		Version: version,
		Short:   description,
		Example: examples,
		RunE:    run,
	}
	cmd.Flags().StringVarP(
		&include,
		"include",
		"i",
		"",
		"which steps to run only",
	)
	cmd.Flags().StringVarP(
		&exclude,
		"exclude",
		"e",
		"",
		"which steps to exclude from complete set",
	)
	cmd.Flags().BoolVarP(
		&shell,
		"shell",
		"s",
		false,
		"run bash shell interactively in container",
	)
	cmd.Flags().BoolVarP(
		&remove,
		"remove",
		"r",
		false,
		"alias for '--include remove,stop'",
	)
	cmd.Flags().BoolVarP(
		&list,
		"list",
		"l",
		false,
		"list steps in order and exit",
	)

	cmd.SetHelpCommand(&cobra.Command{Hidden: true, Use: "no"})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err != nil {
		logger.Error(program, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	steps := stepping.Steps{
		stepCheck,
		stepBuild,
		stepCreate,
		stepStart,
		stepTarball,
		stepUpdate,
		stepDeps,
		stepPackage,
		stepTest,
		stepArchive,
		stepScan,
		stepStop,
		stepRemove,
	}

	err := initOptions(steps)
	if err != nil {
		return err
	}

	err = initStuff(cmd.Use)
	if err != nil {
		return err
	}

	err = steps.Run()
	if err != nil {
		return err
	}

	return nil
}

func initOptions(steps stepping.Steps) error {
	switch {
	case shell:
		steps.Reset()
		steps.ExtraFunctionAfterRun(runShellOptional)

		include = "build,create,start"
	case remove:
		steps.Reset()

		include = "remove,stop"
	case list:
		for i, step := range steps {
			fmt.Printf("%d. %s\n\n", i+1, step.Name)
			for _, desc := range step.Description {
				fmt.Printf("\t%s\n", desc)
			}

			if i < len(steps)-1 {
				fmt.Println()
			}
		}
		return nil
	}

	if include != "" && exclude != "" {
		return errors.New("can't specify --include and --exclude together")
	} else if include != "" {
		err := steps.Include(strings.Split(include, ",")...)
		if err != nil {
			return err
		}
	} else if exclude != "" {
		err := steps.Exclude(strings.Split(exclude, ",")...)
		if err != nil {
			return err
		}
	}

	return nil
}

func initStuff(program string) error {
	var err error

	log = logger.New(program)

	deb, err = debian.ParseChangelog()
	if err != nil {
		return err
	}

	dock, err = docker.New()
	if err != nil {
		return err
	}

	name = naming.New(
		program,
		deb.TargetDist,
		deb.SourceName,
		deb.PackageVersion,
		archiveDir,
	)

	return nil
}
