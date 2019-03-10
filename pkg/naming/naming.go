package naming

import (
	"fmt"
	"strings"
)

type Naming struct {
	program string
	os      string
	dist    string
	source  string
	version string
}

func New(program, os, dist, source, version string) *Naming {
	return &Naming{
		program: program,
		os:      os,
		dist:    dist,
		source:  source,
		version: version,
	}
}

func (n *Naming) Container() string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian versioning allows below characters
	version := strings.Replace(n.version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return fmt.Sprintf(
		"%s_%s-%s_%s-%s",
		n.program,
		n.os,
		n.dist,
		n.source,
		version,
	)
}

func (n *Naming) Image() string {
	return fmt.Sprintf(
		"%s-%s:%s",
		n.program,
		n.os,
		n.dist,
	)
}

func (n *Naming) From() string {
	return fmt.Sprintf(
		"%s:%s",
		n.os,
		n.dist,
	)
}

func (n *Naming) BuildDir() string {
	return n.Container()
}
