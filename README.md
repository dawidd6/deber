<p align="center">
    <img src="art/logo.svg" />
</p>

<h1 align="center">deber</h1>

<p align="center">
    <a href="https://cirrus-ci.com/github/dawidd6/deber">
        <img alt="Build Status" src="https://api.cirrus-ci.com/github/dawidd6/deber.svg">
    </a>
    <a href="https://goreportcard.com/report/github.com/dawidd6/deber">
        <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/dawidd6/deber">
    </a>
    <a href="https://github.com/dawidd6/deber/releases/latest">
        <img alt="Latest Release" src="https://img.shields.io/github/tag-date/dawidd6/deber.svg">
    </a>
</p>

`Debian` **+** `Docker` **=** `deber`

Utility made with simplicity in mind to provide
an easy way for building Debian packages in
Docker containers.

## Screencast

[![asciicast](https://asciinema.org/a/237780.svg)](https://asciinema.org/a/237780)

## Features

- Build packages for Debian and Ubuntu
- Use official Debian and Ubuntu images from DockerHub
- Automatically determine if target distribution is Ubuntu or Debian
  by querying DockerHub API
- Skip already ran steps (not every one)
- Include or exclude steps per your likings
- Plays nice with `gbp-buildpackage`
- Easy local package dependency resolve
- Don't clutter your parent directories with `.deb`, `.dsc` and company
- Every successfully built package goes to local repo automatically
  so you can easily build another package that depends on previous one
- Ability to provide custom `dpkg-buildpackage` and `lintian`
  options by exporting a couple of environment variables

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

Only one option passed to deber is respected,
it is pointless to specify multiple options at once.

## FAQ

**Okay everything went well, but... where the hell is my `.deb`?!**

If you haven't specified `DEBER_ARCHIVE` environment variable, then
it's located in `~/deber`.
I made it this way, because it was just hard to look at my parent directory,
cluttered with `.orig.tar.gz`, `.deb`, `.changes` and God knows what else.

**Where is build directory located?**

`/tmp/$container`

**Where is apt's cache directory located?**

`/tmp/deber:$dist`

**How images built by deber are named?**

Repository is `deber` and tag is `$dist`

**I have already built image but it is building again?!**

Probably because it is 14 days old and deber decided to
update it.

## More

Options, environment variables and others are listed and explained in [manpage](doc/deber.md).
