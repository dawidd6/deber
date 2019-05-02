package cli

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"github.com/dawidd6/deber/pkg/walking"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	include string
	exclude string
	shell   bool
	remove  bool
	list    bool

	archiveDir = os.Getenv("DEBER_ARCHIVE")
	logColor   = os.Getenv("DEBER_LOG_COLOR")
)

func init() {
	if logColor == "no" || logColor == "false" || logColor == "off" {
		log.SetNoColor()
	}
}

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
		log.Error(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	steps := stepping.Steps{
		walking.StepCheck,
		walking.StepBuild,
		walking.StepCreate,
		walking.StepStart,
		walking.StepTarball,
		walking.StepUpdate,
		walking.StepDeps,
		walking.StepPackage,
		walking.StepTest,
		walking.StepArchive,
		walking.StepScan,
		walking.StepStop,
		walking.StepRemove,
	}

	switch {
	case list:
		return printSteps(steps)
	case remove:
		include = "remove,stop"
	case shell:
		// TODO figure out how to put optional steps to stepping
		steps.ExtraFunctionAfterRun(walking.ShellOptional)
		include = "build,create,start"
	}

	err := handleIncludeExclude(steps)
	if err != nil {
		return err
	}

	deb, err := debian.ParseChangelog()
	if err != nil {
		return err
	}

	dock, err := docker.New()
	if err != nil {
		return err
	}

	name := naming.New(
		cmd.Use,
		deb.TargetDist,
		deb.SourceName,
		deb.PackageVersion,
		archiveDir,
	)

	err = steps.Run(deb, dock, name)
	if err != nil {
		return err
	}

	return nil
}

func printSteps(steps stepping.Steps) error {
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

func handleIncludeExclude(steps stepping.Steps) error {
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
