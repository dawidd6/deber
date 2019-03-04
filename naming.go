package main

import (
	"fmt"
)

type Naming struct {
	program string
	os      string
	dist    string
	debian  *Debian
}

func NewNaming(os, dist string, debian *Debian) *Naming {
	return &Naming{
		program: program,
		os:      os,
		dist:    dist,
		debian:  debian,
	}
}

func (n *Naming) Container() string {
	return fmt.Sprintf(
		"%s_%s-%s_%s-%s",
		n.program,
		n.os,
		n.dist,
		n.debian.Source,
		n.debian.Version,
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
