package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
)

// StepArchive defines archive step.
var StepArchive = &stepping.Step{
	Name: "archive",
	Run:  Archive,
	Description: []string{
		"Moves built package artifacts (like .deb, .dsc and others) to archive.",
		"Package directory in archive is overwritten every time.",
	},
}

// Archive function moves successful build to archive by overwriting.
func Archive(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Archiving build")

	info, _ := os.Stat(name.ArchivePackageDir)
	if info != nil {
		err := os.RemoveAll(name.ArchivePackageDir)
		if err != nil {
			return log.FailE(err)
		}
	}

	err := os.Rename(name.BuildDir, name.ArchivePackageDir)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
