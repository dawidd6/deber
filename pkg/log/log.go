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

	drop bool
)

// SetNoColor function empties color string constants.
func SetNoColor() {
	blue = ""
	red = ""
	normal = ""
}

// Info function prints informational log messages.
func Info(v interface{}) {
	drop = false

	fmt.Printf("%s%s:info:%s %s ...", blue, app.Name, normal, v)
}

// Error function prints error log messages.
//
// It is effectively used only once
// so there is for it to be struct method.
func Error(v interface{}) {
	fmt.Printf("%s%s:error:%s %s\n", red, app.Name, normal, v)
}

// Drop function prints just a new line
// and informs to not print anything after dots.
//
// Call this before operation that you know will output to Stdout.
func Drop() {
	drop = true

	fmt.Println()
}

// Skip function prints "skipped" after dots.
func Skip() {
	if !drop {
		fmt.Printf("skipped\n")
	}
}

// Done function prints "done" after dots.
func Done() {
	if !drop {
		fmt.Printf("done\n")
	}
}

// Fail function prints "failed" after dots.
func Fail() {
	if !drop {
		fmt.Printf("failed\n")
	}
}

// SkipE function wraps Skip() and returns nil error.
func SkipE() error {
	Skip()

	return nil
}

// DoneE function wraps Done() and returns nil error.
func DoneE() error {
	Done()

	return nil
}

// FailE function wraps Fail() and returns passed error.
func FailE(err error) error {
	Fail()

	return err
}
