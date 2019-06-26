package main

import (
	"github.com/dawidd6/deber/pkg/cli"
	"github.com/dawidd6/deber/pkg/env"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
)

const (
	Name        = "deber"
	Version     = "0.5"
	Description = "Debian packaging with Docker."
)

func init() {
	// TODO this is bullshit, must admit
	log.Prefix = Name
	cli.Prefix = Name
	env.Prefix = Name
	naming.Prefix = Name
}

func main() {
	cli.Run(Name, Version, Description)
}
