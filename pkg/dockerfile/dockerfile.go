// Package dockerfile includes template Dockerfile
package dockerfile

import (
	"bytes"
	"github.com/dawidd6/deber/pkg/naming"
	"text/template"
)

// Template struct defines parameters passed to
// dockerfile template.
type Template struct {
	// Repo is the image repository
	Repo string
	// Tag is the image tag
	Tag string
	// SourceDir = /build/source
	SourceDir string
}

const dockerfileTemplate = `
# From which Docker image do we start?
FROM {{ .Repo }}:{{ .Tag }}

# Remove not needed apt configs.
RUN rm /etc/apt/apt.conf.d/*

# Run apt without confirmations.
RUN echo "APT::Get::Assume-Yes "true";" > /etc/apt/apt.conf.d/00noconfirm

# Set debconf to be non interactive
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# Install required packages
RUN apt-get update && \
	apt-get install --no-install-recommends -y build-essential devscripts debhelper lintian fakeroot dpkg-dev

# Set working directory.
WORKDIR {{ .SourceDir }}

# Sleep all the time and just wait for commands.
CMD ["sleep", "inf"]
`

func Parse(repo, tag string) ([]byte, error) {
	t := Template{
		Repo:      repo,
		Tag:       tag,
		SourceDir: naming.ContainerSourceDir,
	}

	temp, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, t)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
