package main

import "github.com/dawidd6/deber/pkg/cli"

const (
	program = "deber"
	version = "0.4"
	desc    = `Debian packaging with Docker`
)

func main() {
	cli.Run(program, version, desc)
}
