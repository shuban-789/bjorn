package util

import "fmt"

func Success(msg string, args ...any) string {
	return fmt.Sprintf("\033[32m[SUCCESS]\033[0m "+msg, args...)
}

func Fail(msg string, args ...any) string {
	return fmt.Sprintf("\033[31m[FAIL]\033[0m "+msg, args...)
}

func Info(msg string, args ...any) string {
	return fmt.Sprintf("\033[33m[INFO]\033[0m "+msg, args...)
}
