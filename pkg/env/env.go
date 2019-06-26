package env

import (
	"fmt"
	"os"
	"strings"
)

var Prefix string

func Get(envName, defaultValue string) string {
	prefix := strings.ToUpper(Prefix)
	envName = fmt.Sprintf("%s_%s", prefix, envName)
	envValue := os.Getenv(envName)
	if envValue != "" {
		return envValue
	}

	return defaultValue
}
