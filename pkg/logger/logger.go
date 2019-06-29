package logger

import (
	"fmt"
)

const (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

type Logger struct {
	prefix string
	color  bool
}

func New(prefix string, color bool) *Logger {
	return &Logger{
		prefix: prefix,
		color:  color,
	}
}

func (log *Logger) Info(v ...interface{}) {
	if log.color {
		fmt.Printf("%s%s:info:%s %s\n", blue, log.prefix, normal, v)
	} else {
		fmt.Printf("%s:info: %s\n", log.prefix, v)
	}
}

func (log *Logger) Error(v ...interface{}) {
	if log.color {
		fmt.Printf("%s%s:error:%s %s\n", red, log.prefix, normal, v)
	} else {
		fmt.Printf("%s:error: %s\n", log.prefix, v)
	}
}

func (log *Logger) InfoExtra(v interface{}) {
	fmt.Printf("  %s\n", v)
}
