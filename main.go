package main

import "github.com/dawidd6/deber/pkg/cli"

const (
	program = "deber"
	version = "0.5"
	desc    = `Debian packaging with Docker.`
)

const examples = `
Basic usage of deber with gbp:

    $ gbp buildpackage --git-builder deber

Listing all available steps in order:

	$ deber -l

Excluding some steps:

    $ deber -e remove -e stop -e archive

Removing container after unsuccessful build (if needed):

    $ deber -i remove -i stop

Only building image:

    $ deber -i build

Only moving tarball and creating container:
Note: this example assumes that you specified 'builder = deber' in 'gbp.conf'.

    $ gbp buildpackage -i tarball -i create

Check archive before starting the machinery:

    $ deber --check

Run without updating apt's cache:

    $ deber -e update

Launch interactive bash shell session in container:
Note: this option should be used standalone,
      when shell exits, deber exits too.

    $ deber --shell`

func main() {
	cli.Run(program, version, desc, examples)
}
