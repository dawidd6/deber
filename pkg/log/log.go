package log

import (
	"fmt"
)

const (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var (
	Prefix string

	NoColor         = false
	ExtraInfoIndent = "  "
	newLine         = true
)

// TODO better function namings

func Drop() {
	newLine = true
	fmt.Println()
}

func None() error {
	return nil
}

func Failed(err error) error {
	if !newLine {
		fmt.Printf("failed\n")
	}

	return err
}

func Done() error {
	if !newLine {
		fmt.Printf("done\n")
	}

	return nil
}

func Skipped() error {
	if !newLine {
		fmt.Printf("skipped\n")
	}

	return nil
}

func Custom(s string) error {
	if !newLine {
		fmt.Printf("%s\n", s)
	}

	return nil
}

func ExtraInfo(v interface{}) {
	fmt.Printf("%s%s ...", ExtraInfoIndent, v)

	newLine = false
}

func Info(v interface{}) {
	if !NoColor {
		fmt.Printf("%s%s:info:%s %s ...", blue, Prefix, normal, v)
	} else {
		fmt.Printf("%s:info: %s ...", Prefix, v)
	}

	newLine = false
}

func Error(v interface{}) {
	if !NoColor {
		fmt.Printf("%s%s:error:%s %s\n", red, Prefix, normal, v)
	} else {
		fmt.Printf("%s:error: %s\n", Prefix, v)
	}
}
