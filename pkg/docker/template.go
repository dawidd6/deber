package docker

import (
	"bytes"
	"text/template"
)

type DockerfileTemplate struct {
	From               string
	ContainerSourceDir string
	ContainerRepoDir   string
}

const dockerfileTemplate = `
FROM {{ .From }}

ARG pkgs="build-essential devscripts dpkg-dev debhelper equivs sudo"
ARG user="builder"
ARG apty="/usr/local/bin/apty"
ARG sources="/etc/apt/sources.list.d/repo.list"

RUN apt-get update && \
    apt-get install -y ${pkgs} && \
    rm /etc/apt/apt.conf.d/*
RUN useradd ${user} && \
    echo "${user} ALL=NOPASSWD: ALL" > /etc/sudoers
RUN echo "apt-get -y \$@" > ${apty} && \
    chmod +x ${apty}
RUN mkdir -p {{ .ContainerRepoDir }} && \
    touch {{ .ContainerRepoDir }}/Packages && \
    echo "deb [trusted=yes] file://{{ .ContainerRepoDir }} ./" > ${sources}

USER ${user}:${user}

WORKDIR {{ .ContainerSourceDir }}

CMD ["sleep", "inf"]
`

func dockerfileParse(from string) (string, error) {
	t := DockerfileTemplate{
		From:               from,
		ContainerSourceDir: ContainerSourceDir,
		ContainerRepoDir:   ContainerRepoDir,
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
