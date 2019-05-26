// Package cli is the core of deber.
package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"os"
	"strings"
)

var (
	includeSteps []string
	excludeSteps []string
	listSteps    bool

	shell bool
	check bool

	dpkgFlags    string
	lintianFlags string
	archiveDir   string

	noColor bool

	genManpage bool
)

// Run function is the first that should be executed.
//
// It's the heart of cli.
func Run(program, version, description, examples string) {
	cmd := &cobra.Command{
		Use:     program,
		Version: version,
		Short:   description,
		Example: examples,
		RunE:    run,
	}

	cmd.Flags().StringArrayVarP(
		&includeSteps,
		"include-step",
		"i",
		nil,
		"which steps should be run exclusively",
	)
	cmd.Flags().StringArrayVarP(
		&excludeSteps,
		"exclude-step",
		"e",
		nil,
		"which steps should be omitted",
	)
	cmd.Flags().BoolVarP(
		&listSteps,
		"list-steps",
		"l",
		false,
		"list all available steps in order and exit",
	)

	cmd.Flags().BoolVarP(
		&shell,
		"shell",
		"s",
		false,
		"run only interactive bash session in container",
	)
	cmd.Flags().BoolVarP(
		&check,
		"check",
		"c",
		false,
		"check only if package is already in archive",
	)

	cmd.Flags().StringVar(
		&dpkgFlags,
		"dpkg-flags",
		"-tc",
		"additional dpkg-buildpackage flags to be passed",
	)
	cmd.Flags().StringVar(
		&lintianFlags,
		"lintian-flags",
		"-i -I",
		"additional lintian flags to be passed",
	)
	cmd.Flags().StringVar(
		&archiveDir,
		"archive-dir",
		os.Getenv("HOME"),
		"directory where built packages should be kept",
	)

	cmd.Flags().BoolVar(
		&noColor,
		"no-color",
		false,
		"do not colorize log output",
	)

	cmd.Flags().BoolVar(
		&genManpage,
		"gen-manpage",
		false,
		"generate manpage in current directory",
	)

	cmd.Flags().SortFlags = false
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
	steps := getSteps()

	if listSteps {
		for i, step := range steps.Keys() {
			fmt.Printf("%d\t%s\n", i+1, step)
		}

		return nil
	}

	if genManpage {
		header := &doc.GenManHeader{
			Title:   strings.Title(cmd.Use),
			Section: "1",
		}

		err := doc.GenManTree(cmd, header, "./")
		if err != nil {
			return err
		}

		return nil
	}

	if noColor {
		log.SetNoColor()
	}

	if shell {
		steps.InsertAfter(stepStart, stepShellOptional, runShellOptional)
		includeSteps = append(
			includeSteps,
			stepBuild,
			stepCreate,
			stepStart,
			stepShellOptional,
		)
	}

	if check {
		steps.Prepend(stepCheckOptional, runCheckOptional)
		includeSteps = append(
			includeSteps,
			stepCheckOptional,
		)
	}

	if includeSteps != nil {
		for _, step := range includeSteps {
			if !steps.Has(step) {
				return fmt.Errorf("step \"%s\" not recognized", step)
			}
		}

		steps.DeleteAllExcept(includeSteps...)
	}

	if excludeSteps != nil {
		for _, step := range excludeSteps {
			if !steps.Has(step) {
				return fmt.Errorf("step \"%s\" not recognized", step)
			}
		}

		steps.Delete(excludeSteps...)
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

	for _, step := range steps.Values() {
		f := step.(func(*debian.Debian, *docker.Docker, *naming.Naming) error)
		err = f(deb, dock, name)
		if err != nil {
			return err
		}
	}

	return nil
}
