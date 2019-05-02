package log

import (
	"fmt"
)

const (
	program = "deber"

	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var drop bool

func Info(v interface{}) {
	drop = false

	fmt.Printf("%s%s:info:%s %s ...", blue, program, normal, v)
}

// Error function is effectively used only once
// so there is for it to be struct method.
func Error(v interface{}) {
	fmt.Printf("%s%s:error:%s %s\n", red, program, normal, v)
}

// Call this function before operation that you know will output to Stdout.
func Drop() {
	drop = true

	fmt.Println()
}

func Skip() {
	if !drop {
		fmt.Printf("skipped\n")
	}
}

func Done() {
	if !drop {
		fmt.Printf("done\n")
	}
}

func Fail() {
	if !drop {
		fmt.Printf("failed\n")
	}
}

func SkipE() error {
	Skip()

	return nil
}

func DoneE() error {
	Done()

	return nil
}

func FailE(err error) error {
	Fail()

	return err
}
