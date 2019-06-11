package env

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"os"
	"strings"
)

var Prefix = strings.ToUpper(app.Name)

func Get(envName, defaultValue string) string {
	envValue := os.Getenv(fmt.Sprintf("%s_%s", Prefix, envName))
	if envValue != "" {
		return envValue
	}

	return defaultValue
}
