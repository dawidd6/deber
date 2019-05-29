// Package log provides convenient way of logging stuff.
package log

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
)

var (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

// SetNoColor function empties color string constants.
func SetNoColor() {
	blue = ""
	red = ""
	normal = ""
}

// Info function prints informational log messages.
func Info(v interface{}) {
	fmt.Printf("%s%s:info:%s %s ...\n", blue, app.Name, normal, v)
}

// Error function prints error log messages.
//
// It is effectively used only once
// so there is for it to be struct method.
func Error(v interface{}) {
	fmt.Printf("%s%s:error:%s %s\n", red, app.Name, normal, v)
}
