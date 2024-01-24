package core

import (
	"fmt"
)

type Printer interface {
	Printf(msg string, args ...any)
	Println(args ...any)
}

var printer Printer

func CurrentPrinter() Printer {
	return printer
}

func SetPrinter(p Printer) {
	printer = p
}

func Printf(msg string, args ...any) {
	printer.Printf(msg, args...)
}

func Println(args ...any) {
	printer.Println(args...)
}

type StandardPrinter struct{}

func (sp *StandardPrinter) Printf(msg string, args ...any) {
	_, _ = fmt.Printf(msg, args...)
}

func (sp *StandardPrinter) Println(args ...any) {
	_, _ = fmt.Println(args...)
}

type DualWriter struct {
	printfCB  func(string, ...any) (int, error)
	printlnCB func(...any) (int, error)
}

func NewDualWriter(
	printfCB func(string, ...any) (int, error),
	printlnCB func(...any) (int, error),
) *DualWriter {
	return &DualWriter{
		printfCB:  printfCB,
		printlnCB: printlnCB,
	}
}

func (dw *DualWriter) Printf(msg string, args ...any) {
	_, _ = fmt.Printf(msg, args...)
	_, _ = dw.printfCB(msg, args...)
}

func (dw *DualWriter) Println(args ...any) {
	_, _ = fmt.Println(args...)
	_, _ = dw.printlnCB(args...)
}
