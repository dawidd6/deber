// Package log includes logging utilities
package log

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	cyan   = "\033[0;36m"
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var (
	// NoColor controls if log will be colored or not
	NoColor bool
	// Prefix is the program name, will be outputted before info messages
	Prefix  string
	dropped bool
)

func init() {
	Prefix = filepath.Base(os.Args[0])
}

// Drop function prints new line
func Drop() {
	if dropped {
		return
	}

	dropped = true
	fmt.Println()
}

// Info function prints given string
func Info(info string) {
	dropped = false

	if NoColor {
		fmt.Printf("%s:info: %s ...", Prefix, info)
	} else {
		fmt.Printf("%s%s:info:%s %s ...", blue, Prefix, normal, info)
	}
}

// Error function prints given error
func Error(err error) {
	if NoColor {
		fmt.Printf("%s:error: %s\n", Prefix, err)
	} else {
		fmt.Printf("%s%s:error:%s %s\n", red, Prefix, normal, err)
	}
}

// ExtraInfo prints given info with indent and without colors or prefix
func ExtraInfo(info string) {
	dropped = false
	fmt.Printf("  %s ...", info)
}

// Skipped function prints 'skipped' and new line
func Skipped() error {
	if !dropped {
		fmt.Printf("%s", "skipped")
		Drop()
	}

	return nil
}

// Done function prints 'done' and new line
func Done() error {
	if !dropped {
		fmt.Printf("%s", "done")
		Drop()
	}

	return nil
}

// Failed function prints 'failed' and new line
func Failed(err error) error {
	if !dropped {
		fmt.Printf("%s", "failed")
		Drop()
	}

	return err
}
