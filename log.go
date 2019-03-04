package main

import (
	"fmt"
	"github.com/fatih/color"
)

func LogInfo(v interface{}) {
	s := color.BlueString("%s:info:", program)
	fmt.Printf("%s %s ...", s, v)
}

func LogDone() {
	fmt.Printf("done\n")
}

func LogSkip() {
	fmt.Printf("skipped\n")
}

func LogFail() {
	fmt.Printf("failed\n")
}

func LogError(v interface{}) {
	s := color.RedString("%s:error:", program)
	fmt.Println(s, v)
}
