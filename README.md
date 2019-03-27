# deber
[![Build Status](https://api.cirrus-ci.com/github/dawidd6/deber.svg)](https://cirrus-ci.com/github/dawidd6/deber)

**deb**_(ian)_ + _(dock)_**er** = **deber**

Create Debian packages in Docker containers easily.

## Screencast

[![asciicast](https://asciinema.org/a/236225.svg)](https://asciinema.org/a/236225)

## Dependencies

Name | Min Version | Notes
---|:---:|:---:
**Go** | `1.8` | compiler version
**Docker** | `1.30` | only daemon is required, version of API

## Installation

```bash
go get -u github.com/dawidd6/deber
```

## Directology

**ARCHIVE**

```
HostArchiveDir = /home/dawidd6/deber
HostArchiveFromDir = /home/dawidd6/deber/debian:unstable << **MOUNT**
HostArchiveFromOutputDir = /home/dawidd6/deber/debian:unstable/golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88-1
```

```
/home/dawidd6/deber
└── debian:unstable
    └── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88-1
        ├── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88-1_amd64.buildinfo
        ├── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88-1_amd64.changes
        ├── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88-1.debian.tar.xz
        ├── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88-1.dsc
        ├── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88.orig.tar.gz
        ├── golang-github-alcortesm-tgz-dev_0.0~git20161220.9c5fe88-1_all.deb
        └── source
```

**SOURCE**

```
HostSourceDir, HostSourceSourceTarballDir = /home/dawidd6/TEST
HostSourceInputDir = /home/dawidd6/TEST/golang-github-alcortesm-tgz << **MOUNT**
HostSourceSourceTarballFile = /home/dawidd6/TEST/golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88.orig.tar.gz
```

```
/home/dawidd6/TEST
├── golang-github-alcortesm-tgz
│   ├── debian
│   │   ├── changelog
│   │   ├── control
│   │   ├── copyright
│   │   ├── gbp.conf
│   │   ├── rules
│   │   ├── source
│   │   │   └── format
│   │   └── watch
│   ├── fixtures
│   │   ├── invalid-gzip.tgz
│   │   ├── not-a-tar.tgz
│   │   ├── test-01.tgz
│   │   ├── test-02.tgz
│   │   └── test-03.tgz
│   ├── LICENSE
│   ├── README.md
│   ├── tgz.go
│   └── tgz_test.go
└── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88.orig.tar.gz
```
    
**BUILD**

```
HostBuildDir = /tmp
HostBuildCacheDir = /tmp/deber-debian:unstable << **MOUNT**
HostBuildOutputDir, HostBuildTargetTarballDir = /tmp/deber_debian-unstable_golang-github-alcortesm-tgz_0.0-git20161220.9c5fe88-1 << **MOUNT**
HostBuildTargetTarballFile = /tmp/deber_debian-unstable_golang-github-alcortesm-tgz_0.0-git20161220.9c5fe88-1/golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88.orig.tar.gz
```

```
/tmp
├── deber_debian-unstable_golang-github-alcortesm-tgz_0.0-git20161220.9c5fe88-1
│   ├── golang-github-alcortesm-tgz_0.0~git20161220.9c5fe88.orig.tar.gz
│   └── source
├── deber-debian:unstable
│   ├── archives
│   │   ├── dh-golang_1.39_all.deb
│   │   ├── golang-1.11-go_1.11.5-1_amd64.deb
│   │   ├── golang-1.11-go_1.11.6-1_amd64.deb
│   │   ├── golang-1.11-src_1.11.5-1_amd64.deb
│   │   ├── golang-1.11-src_1.11.6-1_amd64.deb
│   │   ├── golang-any_2%3a1.11~1_amd64.deb
│   │   ├── golang-go_2%3a1.11~1_amd64.deb
│   │   ├── golang-src_2%3a1.11~1_amd64.deb
│   │   ├── lock
│   │   └── pkg-config_0.29-6_amd64.deb
│   ├── last-updated
│   ├── pkgcache.bin
│   └── srcpkgcache.bin

```
