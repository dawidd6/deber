package app

import (
	"github.com/dawidd6/deber/pkg/cli"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
)

type App struct {
	*logger.Logger
	*docker.Docker
	*naming.Naming
	*cli.Options
}
