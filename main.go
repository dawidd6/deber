package main

import "github.com/dawidd6/deber/pkg/cli"

const (
	program = "deber"
	version = "0.4"
	desc    = `Debian packaging with Docker`
)

const examples = `
Basic usage of deber with gbp:

    $ gbp buildpackage --git-builder deber

Excluding some steps:

    $ deber --exclude remove,stop,archive

Removing container after unsuccessful build (if needed):

    $ deber --include remove,stop

Only building image:

    $ deber --include build

Only moving tarball and creating container:

Note: this example assumes that you specified **builder = deber** in **gbp.conf**.

    $ gbp buildpackage --include tarball,create

Build package regardless it's existence in archive:

    $ deber --exclude check

Build package without checking archive, updating apt's cache and scanning packages:

    $ deber --exclude check,update,scan

Launch interactive bash shell session in container:

Note: specifying other options after or before this, takes no effect.

    $ deber --shell`

func main() {
	cli.Run(program, version, desc, examples)
}
