package docker

import (
	"github.com/dawidd6/deber/pkg/naming"
	"bytes"
	"text/template"
)

type DockerfileTemplate struct {
	From    string
	User    string
	Archive string
	Source  string
}

const dockerfileTemplate = `
# From which Docker image do we start?
FROM {{ .From }}

# Install required packages and remove not needed apt configs.
RUN apt-get update && \
	apt-get install -y build-essential devscripts dpkg-dev debhelper equivs sudo && \
	rm /etc/apt/apt.conf.d/*

# Add normal user and with su access.
RUN useradd {{ .User }} && \
	echo "{{ .User }} ALL=NOPASSWD: ALL" > /etc/sudoers

# Create apt-get wrapper script.
RUN printf '#!/bin/bash\napt-get -o Debug::pkgProblemResolver=yes -y $@\n' > /bin/apty && \
	chmod +x /bin/apty

# Add local apt repository.
RUN mkdir -p {{ .Archive }} && \
    touch {{ .Archive }}/Packages && \
    echo "deb [trusted=yes] file://{{ .Archive }} ./" > /etc/apt/sources.list.d/a.list

# Define default user.
USER {{ .User }}:{{ .User }}

# Set working directory.
WORKDIR {{ .Source }}

# Sleep all the time and just wait for commands.
CMD ["sleep", "inf"]
`

func dockerfileParse(from string) (string, error) {
	t := DockerfileTemplate{
		From:    from,
		User:    "builder",
		Archive: naming.ContainerArchiveDir,
		Source: naming.ContainerSourceDir,
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
