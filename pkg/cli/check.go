package cli

import (
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
)

var stepCheck = &stepping.Step{
	Name: "check",
	Run:  runCheck,
	Description: []string{
		"Checks if to-be-built package is already built and in archive.",
		"If package is in archive, then deber will simply exit.",
		"To build package anyway, simply exclude this step.",
	},
}

func runCheck() error {
	log.Info("Checking archive")

	info, _ := os.Stat(name.ArchivePackageDir)
	if info != nil {
		log.Skip()
		os.Exit(0)
	}

	return log.DoneE()
}
