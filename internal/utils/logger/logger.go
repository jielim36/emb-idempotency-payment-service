package logger

import "fmt"

type Logger struct{}

// Simulate logger
func (log *Logger) Error(err error) {
	fmt.Printf("[Error] error=%s\n", err.Error())
}

func (log *Logger) Info(msg string) {
	fmt.Printf("[Info] message=%s\n", msg)
}
