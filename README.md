# deber
[![Build Status](https://api.cirrus-ci.com/github/dawidd6/deber.svg)](https://cirrus-ci.com/github/dawidd6/deber)

**deb**_(ian)_ + _(dock)_**er** = **deber**

Create Debian packages in Docker containers easily.

## Dependencies

- **Go** 1.8 or newer
- **Docker** (daemon)

## Installation

```bash
go get -u github.com/dawidd6/deber

```

## Help

```
Debian packaging with Docker

Usage:
  deber OS DIST [flags] [-- dpkg-buildpackage options]

Examples:
  basic:
    deber ubuntu xenial

  only with needed steps:
    deber ubuntu bionic --with-steps build
    deber debian buster --with-steps build,create

  without unneeded steps:
    deber debian unstable --without-steps remove,stop,build

  with gbp:
    gbp buildpackage --git-builder=deber ubuntu xenial

  with dpkg-buildpackage options
    deber ubuntu xenial -- -nc -b

Flags:
  -h, --help                   help for deber
      --show-steps             show available steps in order
  -v, --verbose                show more output
      --version                version for deber
      --with-steps string      specify which of the steps should execute
      --without-steps string   specify which of the steps should not execute
```