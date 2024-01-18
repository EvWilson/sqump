package core

import (
	"fmt"
	"io"
)

var writer io.Writer

func SetWriter(w io.Writer) {
	writer = w
}

func Printf(msg string, args ...any) {
	_, _ = fmt.Fprintf(writer, msg, args...)
}

func Println(args ...any) {
	_, _ = fmt.Fprintln(writer, args...)
}
