package naming

import (
	"fmt"
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
	return fmt.Sprintf(
		"%s_%s-%s_%s-%s",
		n.program,
		n.os,
		n.dist,
		n.source,
		n.version,
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
