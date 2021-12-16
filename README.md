# deber

![](https://github.com/dawidd6/deber/workflows/Tests/badge.svg)
[![GoDoc](https://godoc.org/github.com/dawidd6/deber?status.svg)](https://godoc.org/github.com/dawidd6/deber)
[![go report card](https://goreportcard.com/badge/github.com/dawidd6/deber)](https://goreportcard.com/report/github.com/dawidd6/deber)
[![latest tag](https://img.shields.io/github/tag-date/dawidd6/deber.svg)](https://github.com/dawidd6/deber/releases/latest)

### `Debian` **+** `Docker` **=** `deber`

Utility made with simplicity in mind to provide
an easy way for building Debian packages in
Docker containers.

## Screencast

[![asciicast](https://asciinema.org/a/H2bjgbvzYnFNZLvEZruztIdnZ.svg)](https://asciinema.org/a/H2bjgbvzYnFNZLvEZruztIdnZ)

## Features

- Build packages for Debian and Ubuntu
- Use official Debian and Ubuntu images from DockerHub
- Automatically determine if target distribution is Ubuntu or Debian
  by querying DockerHub API
- Skip already ran steps (not every one)
- Install extra local packages in container
- Plays nice with `gbp-buildpackage`
- Easy local package dependency resolve
- Don't clutter your parent directories with `.deb`, `.dsc` and company
- Every successfully built package goes to local repo automatically
  so you can easily build another package that depends on previous one
- Ability to provide custom `dpkg-buildpackage` and `lintian` options
- Packages downloaded by apt are stored in temporary directory,
  to avoid repetitive unnecessary network load
- Option to drop into interactive bash shell session in container,
  for debugging or other purposes
- Use network in build process or not
- Automatically rebuilds image if old enough

## Installation

**Homebrew**

```bash
brew install dawidd6/tap/deber
```

**Source**

```bash
go install github.com/dawidd6/deber@latest
```

## Usage

I recommend to use deber with gbp if possible, but it will work just fine
as a standalone builder, like sbuild or pbuilder.

Let's assume that you are in directory with already debianized source, have
orig upstream tarball in parent directory and you want to build a package.
Just run:

```bash
deber
```

or if you use gbp and have `builder = deber` in `gbp.conf`

```bash
gbp buildpackage
```

If you run it first time, it will build Docker image and then proceed to build
your package.

To make use of packages from archive to build another package, specify desired directories with built artifacts and `deber` will take them to consideration when installing dependencies:

```bash
deber -p ~/deber/unstable/pkg1/1.0.0-1 -p ~/deber/unstable/pkg2/2.0.0-2
```

## FAQ

**Okay everything went well, but... where the hell is my `.deb`?!**

The location for all build outputs defaults to `$HOME/deber`.
I made it this way, because it was just hard to look at my parent directory,
cluttered with `.orig.tar.gz`, `.deb`, `.changes` and God knows what else.

**Where is build directory located?**

`/tmp/$CONTAINER`

**Where is apt's cache directory located?**

`/tmp/$IMAGE`

**How images built by deber are named?**

`deber:$DIST`

**I have already built image but it is building again?!**

Probably because it is 14 days old and deber decided to
update it.

**How to build a package for different distributions?**

Make a new entry with desired target distribution in `debian/changelog`
and run `deber`.

Or specify the desired distribution with `--distribution` option.

**How to cross-build package for different architecture?**

This is not implemented yet. But I'm planning to make use of `qemu` or something else.

## CONTRIBUTING

I appreciate any contributions, so feel free to do so!
