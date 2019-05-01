package cli

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"strings"
)

func initOptions(steps stepping.Steps) error {
	switch {
	case shell:
		steps.Reset()
		steps.ExtraFunctionAfterRun(runShellOptional)

		err := steps.Include("build", "create", "start")
		if err != nil {
			return err
		}
	case remove:
		steps.Reset()

		err := steps.Include("remove", "stop")
		if err != nil {
			return err
		}
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
	default:
		if include != "" && exclude != "" {
			return errors.New("can't specify --include and --exclude together")
		}
		if include != "" {
			err := steps.Include(strings.Split(include, ",")...)
			if err != nil {
				return err
			}
		}
		if exclude != "" {
			err := steps.Exclude(strings.Split(exclude, ",")...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func initToys(program string) error {
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
