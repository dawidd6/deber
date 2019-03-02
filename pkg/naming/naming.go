package naming

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/constants"
	"github.com/dawidd6/deber/pkg/debian"
)

type Naming struct {
	program string
	os      string
	dist    string
	deb     *debian.Debian
}

func New(os, dist string, deb *debian.Debian) *Naming {
	return &Naming{
		program: constants.Program,
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
