package utils

import (
	"fmt"
	"io"
	"os"
)

var (
	InfoWriter  io.Writer = os.Stdout // For startup messages
	ErrorWriter io.Writer = os.Stderr // For errors/warnings
	// logger      io.WriteCloser
)

// Info logging (startup, success messages)
func Infof(format string, args ...interface{}) {
	fmt.Fprintf(InfoWriter, format+"\n", args...)
}

// Error/Warning logging
func Errorf(format string, args ...interface{}) {
	fmt.Fprintf(ErrorWriter, format+"\n", args...)
}

func Warningf(format string, args ...interface{}) {
	fmt.Fprintf(ErrorWriter, "Warning: "+format+"\n", args...)
}

func WarningIf(err error) {
	if err != nil {
		Warningf("%v\n", err)
	}
}
