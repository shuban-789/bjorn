package bot

import "fmt"

func success(msg string, args ...any) string {
	return fmt.Sprintf("\033[32m[SUCCESS]\033[0m "+msg, args...)
}

func fail(msg string, args ...any) string {
	return fmt.Sprintf("\033[31m[FAIL]\033[0m "+msg, args...)
}

func info(msg string, args ...any) string {
	return fmt.Sprintf("\033[33m[INFO]\033[0m "+msg, args...)
}
