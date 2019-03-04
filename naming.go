package main

import (
	"fmt"
)

type Naming struct {
	program string
	os      string
	dist    string
	deb     *Debian
}

func NewNaming(os, dist string, deb *Debian) *Naming {
	return &Naming{
		program: Program,
		os:      os,
		dist:    dist,
		deb:     deb,
	}
}

func (n *Naming) Container() string {
	return fmt.Sprintf(
		"%s_%s-%s_%s-%s",
		n.program,
		n.os,
		n.dist,
		n.deb.Source,
		n.deb.Version,
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

func (n *Naming) BuildDir() string {
	return n.Container()
}
