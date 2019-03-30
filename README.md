# deber
[![Build Status](https://api.cirrus-ci.com/github/dawidd6/deber.svg)](https://cirrus-ci.com/github/dawidd6/deber)

**deb**_(ian)_ + _(dock)_**er** = **deber**

Create Debian packages in Docker containers easily.

## Screencast

[![asciicast](https://asciinema.org/a/237780.svg)](https://asciinema.org/a/237780)
## Dependencies

Name | Min Version | Notes
---|:---:|:---:
**Go** | `1.8` | compiler version
**Docker** | `1.30` | only daemon is required, version of API

## Installation

```bash
go get -u github.com/dawidd6/deber
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

## FAQ

**Okay everything went well, but... where the hell is my `.deb`?!**

If you haven't specified `DEBER_ARCHIVE` environment variable, then
it's located in `~/deber`.
I made it this way, because it was just hard looking at my parent directory,
cluttered with `.orig.tar.gz`, `.deb`, `.changes` and God knows what else.

**Where is build directory?**

`/tmp/$container`

**Where is apt cache directory?**

`/tmp/deber:$dist`

**How are images tagged?**

Repository is `deber` and tag is `$dist`

**I have already built image but it is building again?!**

Probably because it is 14 days old and deber decided to
update it.

## Info

Options, environment variables and others are listed and explained in [manpage](deber.md).
