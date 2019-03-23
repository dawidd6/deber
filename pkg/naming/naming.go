package naming

import (
	"fmt"
	"strings"
)

func Container(program, image, source, version string) string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian versioning allows below characters
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)
	image = strings.Replace(image, ":", "-", -1)
	image = strings.Replace(image, "/", "-", -1)

	return fmt.Sprintf(
		"%s_%s_%s-%s",
		program,
		image,
		source,
		version,
	)
}

func Image(program, image string) string {
	return fmt.Sprintf(
		"%s-%s",
		program,
		image,
	)
}
