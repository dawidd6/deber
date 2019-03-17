package main

import "github.com/dawidd6/deber/pkg/app"

const (
	program = "deber"
	version = "0.2+git"
	desc    = `Debian packaging with Docker`
)

func main() {
	app.Run(program, version, desc)
}
