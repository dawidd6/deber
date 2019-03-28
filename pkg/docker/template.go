package docker

import (
	"bytes"
	"text/template"
)

type DockerfileTemplate struct {
	From string
}

const dockerfileTemplate = `
FROM {{ .From }}

RUN apt-get update && \
    apt-get install -y build-essential devscripts dpkg-dev debhelper equivs sudo && \
    rm /etc/apt/apt.conf.d/*
RUN useradd builder && \
    echo "builder ALL=NOPASSWD: ALL" > /etc/sudoers
RUN printf '#!/bin/bash\napt-get -y $@\n' > /bin/apty && \
	chmod +x /bin/apty
RUN printf '#!/bin/bash\ncd /archive\ndpkg-scanpackages . > Packages\n' > /bin/scan && \
	chmod +x /bin/scan
RUN mkdir -p /archive && \
    touch /archive/Packages && \
    echo "deb [trusted=yes] file:///archive ./" > /etc/apt/sources.list.d/a.list

USER builder:builder

WORKDIR /build/source

CMD ["sleep", "inf"]
`

func dockerfileParse(from string) (string, error) {
	t := DockerfileTemplate{
		From: from,
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
