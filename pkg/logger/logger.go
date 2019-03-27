package logger

import (
	"fmt"
)

type Logger struct {
	program string
}

func New(program string) *Logger {
	return &Logger{
		program: program,
	}
}

func (l *Logger) Info(v interface{}) {
	blue := "\033[0;34m"
	normal := "\033[0m"
	fmt.Printf("%s%s:info:%s %s ...\n", blue, l.program, normal, v)
}

func (l *Logger) Error(v interface{}) {
	red := "\033[0;31m"
	normal := "\033[0m"
	fmt.Printf("%s%s:error:%s %s\n", red, l.program, normal, v)
}
