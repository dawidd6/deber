package app

import (
	"fmt"
	"github.com/fatih/color"
)

func logInfo(v interface{}) {
	s := color.BlueString("%s:info:", program)
	fmt.Printf("%s %s ...", s, v)
}

func logDone() {
	fmt.Printf("done\n")
}

func logSkip() {
	fmt.Printf("skipped\n")
}

func logFail() {
	fmt.Printf("failed\n")
}

func logError(v interface{}) {
	s := color.RedString("%s:error:", program)
	fmt.Println(s, v)
}
