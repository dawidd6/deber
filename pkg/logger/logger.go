package logger

import (
	"deber/pkg/constants"
	"fmt"
	"github.com/fatih/color"
)

func Info(v interface{}) {
	s := color.BlueString("%s:info:", constants.Program)
	fmt.Printf("%s %s ...", s, v)
}

func Done() {
	fmt.Printf("done\n")
}

func Skip() {
	fmt.Printf("skipped\n")
}

func Fail() {
	fmt.Printf("failed\n")
}

func Error(v interface{}) {
	s := color.RedString("%s:error:", constants.Program)
	fmt.Println(s, v)
}
