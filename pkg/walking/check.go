package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
)

var StepCheck = &stepping.Step{
	Name: "check",
	Run:  Check,
	Description: []string{
		"Checks if to-be-built package is already built and in archive.",
		"If package is in archive, then deber will simply exit.",
		"To build package anyway, simply exclude this step.",
	},
}

func Check(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Checking archive")

	info, _ := os.Stat(name.ArchivePackageDir)
	if info != nil {
		log.Skip()
		os.Exit(0)
	}

	return log.DoneE()
}
