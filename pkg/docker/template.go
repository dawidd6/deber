package docker

import (
	"bytes"
	"text/template"
)

type DockerfileTemplate struct {
	From               string
	ContainerSourceDir string
}

const dockerfileTemplate = `
FROM {{ .From }}

ARG pkgs="build-essential devscripts dpkg-dev debhelper equivs sudo"
ARG user="builder"
ARG apty="/usr/local/bin/apty"

RUN apt-get update && apt-get install -y ${pkgs}
RUN rm /etc/apt/apt.conf.d/*
RUN useradd ${user} && echo "${user} ALL=NOPASSWD: ALL" > /etc/sudoers
RUN echo "apt-get -y \$@" > ${apty} && chmod +x ${apty}

USER ${user}:${user}

WORKDIR {{ .ContainerSourceDir }}

CMD ["sleep", "inf"]
`

func dockerfileParse(from string) (string, error) {
	t := DockerfileTemplate{
		From:               from,
		ContainerSourceDir: ContainerSourceDir,
	}

	temp, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)
	err = temp.Execute(buffer, t)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
