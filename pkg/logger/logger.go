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
	program string

	drop bool
}

func New(program string) *Logger {
	return &Logger{
		program: program,
		drop:    false,
	}
}

func (l *Logger) Info(v interface{}) {
	l.drop = false

	fmt.Printf("%s%s:info:%s %s ...", blue, l.program, normal, v)
}

// Error function is effectively used only once
// so there is for it to be struct method.
func Error(program string, v interface{}) {
	fmt.Printf("%s%s:error:%s %s\n", red, program, normal, v)
}

// Call this function before operation that you know will output to Stdout.
func (l *Logger) Drop() {
	l.drop = true

	fmt.Println()
}

func (l *Logger) Skip() {
	if !l.drop {
		fmt.Printf("skipped\n")
	}
}

func (l *Logger) Done() {
	if !l.drop {
		fmt.Printf("done\n")
	}
}

func (l *Logger) Fail() {
	if !l.drop {
		fmt.Printf("failed\n")
	}
}

func (l *Logger) SkipE() error {
	l.Skip()

	return nil
}

func (l *Logger) DoneE() error {
	l.Done()

	return nil
}

func (l *Logger) FailE(err error) error {
	l.Fail()

	return err
}
