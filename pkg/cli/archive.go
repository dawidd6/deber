package cli

import (
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
)

var stepArchive = &stepping.Step{
	Name: "archive",
	Run:  runArchive,
	Description: []string{
		"Moves built package artifacts (like .deb, .dsc and others) to archive.",
		"Package directory in archive is overwritten every time.",
	},
}

func runArchive() error {
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
