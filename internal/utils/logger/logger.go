package logger

import "fmt"

type Logger struct{}

// Simulate logger
func (log *Logger) Error(err error, msg string) {
	fmt.Printf("[Error] message=%s error=%s\n", msg, err.Error())
}

func (log *Logger) Info(msg string) {
	fmt.Printf("[Info] message=%s\n", msg)
}
