package dockerfile

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TODO this package needs a rewrite

// Template struct defines parameters passed to
// dockerfile template.
type Template struct {
	From      string
	Packages  string
	Backports string
}

const dockerfileTemplate = `
# From which Docker image do we start?
FROM {{ .From }}

# Remove not needed apt configs.
RUN rm /etc/apt/apt.conf.d/*

# Run apt without confirmations.
RUN echo "APT::Get::Assume-Yes "true";" > /etc/apt/apt.conf.d/00noconfirm

# Set debconf to be non interactive
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# Backports pinning
{{ .Backports }}

# Install required packages
RUN apt-get update && \
	apt-get install --no-install-recommends -y {{ .Packages }}

# Set working directory.
WORKDIR /build/source

# Sleep all the time and just wait for commands.
CMD ["sleep", "inf"]
`

func Parse(from string) (string, error) {
	t := Template{
		From:     from,
		Packages: "build-essential devscripts debhelper lintian fakeroot",
	}

	dist := strings.Split(from, ":")[1]

	if strings.Contains(dist, "-backports") {
		t.Backports = fmt.Sprintf(
			"RUN printf \"%s\" > %s",
			"Package: *\\nPin: release a="+dist+"\\nPin-Priority: 800",
			"/etc/apt/preferences.d/backports",
		)
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
