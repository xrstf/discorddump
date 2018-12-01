package main

import (
	"fmt"
	"os"
	"time"
)

func logPrint(format string, args ...interface{}) {
	format = fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC822), format)

	fmt.Printf(format, args...)
}

func logEndLine(format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%s\n", format), args...)
}

func logFatal(format string, args ...interface{}) {
	logPrint(format, args...)
	os.Exit(1)
}
