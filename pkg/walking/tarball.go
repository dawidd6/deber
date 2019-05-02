package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
	"path/filepath"
)

var StepTarball = &stepping.Step{
	Name: "tarball",
	Run:  Tarball,
	Description: []string{
		"Moves orig upstream tarball from parent directory to build directory.",
		"Will fail if tarball is nonexistent and skip if package is native.",
	},
}

func Tarball(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Moving tarball")

	tarball, err := deb.LocateTarball()
	if err != nil {
		return log.FailE(err)
	}

	if tarball == "" {
		return log.SkipE()
	}

	source := filepath.Join(name.SourceParentDir, tarball)
	target := filepath.Join(name.BuildDir, tarball)

	source, err = filepath.EvalSymlinks(source)
	if err != nil {
		return log.FailE(err)
	}

	err = os.Rename(source, target)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
