deber(1) -- Debian packaging with Docker
=============================================

## SYNOPSIS

`deber` [flags]

## DESCRIPTION

Utility made with simplicity in mind to provide
an easy way for building Debian packages in
Docker containers.

Only one option is respected,
it is pointless to specify multiple options at once.

## OPTIONS

 * `-i`, `--include` *string* :
  steps which will be executed as the only ones

 * `-e`, `--exclude` *string* :
  steps which should be omitted

 * `-s`, `--shell` :
  run bash shell interactively in container
  
 * `-r`, `--remove` :
  alias for '--include remove,stop'
    
 * `-l`, `--list` :
  list steps in order and exit

 * `--version` :
  version for deber

 * `-h`, `--help` :
  help for deber

## STEPS

The following steps are executed (in that exact order):

`1. check`

        Checks if to-be-built package is already built and in archive.
        If package is in archive, then deber will simply exit.
        To build package anyway, simply exclude this step.

`2. build`

        Builds image for deber's use.
        This step is skipped if an image is already built.
        Image's parent name is derived from Debian's changelog, for example
        if in `debian/changelog` target distribution is `bionic`, then
        deber will use `ubuntu:bionic` image as a parent from Docker Hub.
        Image's repository name is determined by querying Docker Hub API.
        So, if one wants to build for other distribution than specified in
        `debian/changelog`, just change target distribution to whatever
        one desires and deber will follow.
        Also if image is older than 14 days, deber will try to rebuild it.

`3. create`

        Creates container and makes needed directories on host system.
        Will fail if image is nonexistent.

`4. start`

        Starts previously created container.
        The entry command is `sleep inf`, which means that container
        will just sit there, doing nothing and waiting for commands.

`5. tarball`

        Moves orig upstream tarball from parent directory to build directory.
        Will fail if tarball is nonexistent and skip if package is native.

`6. update`

        Updates apt's cache.
        Also creates empty `Packages` file in archive if nonexistent

`7. deps`

        Installs package's build dependencies in container.
        Runs `mk-build-deps` with appropriate options.

`8. package`

        Runs `dpkg-buildpackage` in container.
        Options passed to `dpkg-buildpackage` are taken from environment variable

`9. test`

        Runs series of commands in container:
          - debc
          - debi
          - lintian
        Options passed to `lintian` are taken from environment variable

`10. archive`

        Moves built package artifacts (like .deb, .dsc and others) to archive.
        Package directory in archive is overwritten every time.

`11. scan`

        Scans available packages in archive and writes result to `Packages` file.
        This `Packages` file is then used by apt in container.

`12. stop`

        Stops container.
        With 10ms timeout.

`13. remove`

        Removes container.
        Nothing more.

## EXAMPLES

Basic usage of deber with gbp:

    $ gbp buildpackage --git-builder deber

Excluding some steps:

    $ deber --exclude remove,stop,archive

Removing container after unsuccessful build (if needed):

    $ deber --include remove,stop

Only building image:

    $ deber --include build

Only moving tarball and creating container:

Note: this example assumes that you specified `builder = deber` in `gbp.conf`.

    $ gbp buildpackage --include tarball,create

Build package regardless it's existence in archive:

    $ deber --exclude check

Build package without checking archive, updating apt's cache and scanning packages:

    $ deber --exclude check,update,scan

Launch interactive bash shell session in container:

Note: specifying other options after or before this, takes no effect.

    $ deber --shell

## ENVIRONMENT VARIABLES

**DEBER_ARCHIVE**

    Directory where deber will put built packages.
    Defaults to "$HOME/deber".

**DEBER_DPKG_BUILDPACKAGE_FLAGS**

    Space separated flags to be passed to dpkg-buildpackage in container.

**DEBER_LINTIAN_FLAGS**

    Space separated flags to be passed to lintian in container.
    
**DEBER_LOG_COLOR**

    Set to "no", "false" or "off" to disable log coloring.

## SEE ALSO

gbp(1), gbp.conf(5), gbp-buildpackage(1), dpkg-buildpackage(1), lintian(1), debc(1), debi(1), mk-build-deps(1)