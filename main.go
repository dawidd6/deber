package main

import "github.com/dawidd6/deber/pkg/app"

const (
	program = "deber"
	version = "0.1+git"
	desc    = `Debian packaging with Docker`
	example = `  basic:
    deber ubuntu xenial

  only with needed steps:
    deber ubuntu bionic --with-steps build
    deber debian buster --with-steps build,create

  without unneeded steps:
    deber debian unstable --without-steps remove,stop,build

  with gbp:
    gbp buildpackage --git-builder=deber ubuntu xenial

  with dpkg-buildpackage options:
    deber ubuntu xenial -- -nc -b`
)

func main() {
	app.Run(program, version, desc, example)
}
