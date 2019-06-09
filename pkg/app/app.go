package app

import (
	"github.com/dawidd6/deber/pkg/docker"
	"pault.ag/go/debian/changelog"
)

type App struct {
	Name        string
	Version     string
	Description string
	Config      *Config

	Docker *docker.Docker
	Debian *changelog.ChangelogEntry
}
