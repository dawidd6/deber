package main

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

RUN apt-get update
RUN apt-get install -y build-essential devscripts dpkg-dev debhelper equivs sudo
RUN rm /etc/apt/apt.conf.d/*
RUN adduser --gecos '' --disabled-password --uid 1000 builder
RUN echo "builder ALL=NOPASSWD: ALL" > /etc/sudoers

USER builder:builder

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
