package app

import (
	"fmt"
)

func logInfo(v interface{}) {
	blue := "\033[0;34m"
	normal := "\033[0m"
	fmt.Printf("%s%s:info:%s %s ...\n", blue, program, normal, v)
}

func logError(v interface{}) {
	red := "\033[0;31m"
	normal := "\033[0m"
	fmt.Printf("%s%s:error:%s %s\n", red, program, normal, v)
}
