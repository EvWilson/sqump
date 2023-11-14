package cmder

import (
	"fmt"
	"io"
	"reflect"
)

type ErrNoop struct {
	name string
}

func NewErrNoop(name string) ErrNoop {
	return ErrNoop{
		name: name,
	}
}

func (en ErrNoop) Error() string {
	return fmt.Sprintf("'%s' is a no-op routing handler", en.name)
}

func (en ErrNoop) Is(other error) bool {
	return reflect.TypeOf(other) == reflect.TypeOf(ErrNoop{})
}

func NewNoopHandler(name string) func([]string) error {
	return func(_ []string) error {
		return NewErrNoop(name)
	}
}

type Root struct {
	description string
	writer      io.Writer
	ops         []Op
}

func NewRoot(
	description string,
	writer io.Writer,
) *Root {
	return &Root{
		description: description,
		writer:      writer,
	}
}

func (r *Root) Register(ops ...*Op) {
	for _, op := range ops {
		r.ops = append(r.ops, *op)
	}
}

func (r *Root) PrintUsage() {
	write := func(s string) {
		r.writer.Write([]byte(s))
	}
	write(fmt.Sprintf("%s\n\n", r.description))
	for _, op := range r.ops {
		write(fmt.Sprintf("%s - %s\n", op.cmdName, op.short))
	}
	write("\n")
}

func (r *Root) PrintExtendedUsage() {
	write := func(s string) {
		r.writer.Write([]byte(s))
	}
	write(fmt.Sprintf("%s\n\n", r.description))
	for _, subOp := range r.ops {
		printOps(r.writer, &subOp, "")
		write("\n")
	}
}

func printOps(writer io.Writer, op *Op, offset string) {
	write := func(s string) {
		writer.Write([]byte(s))
	}
	write(fmt.Sprintf("%s%s    %s\n%s---\n", offset, op.short, op.long, offset))
	for _, subOp := range op.subOps {
		printOps(writer, &subOp, offset+"    ")
	}
}

func (r *Root) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected at least one argument, got none")
	}

	r.Register(NewOp(
		"help",
		"help",
		"Prints an extended help menu for all commands",
		func(_ []string) error {
			r.PrintExtendedUsage()
			return nil
		},
	))

	for _, op := range r.ops {
		if op.cmdName == args[0] {
			return op.Handle(args[1:])
		}
	}

	return fmt.Errorf("did not find handler for subcommand '%s'", args[0])
}

type Op struct {
	cmdName string
	short   string
	long    string
	op      func(args []string) error
	subOps  []Op
}

func NewOp(
	cmdName,
	short,
	long string,
	op func(args []string) error,
	subOps ...*Op,
) *Op {
	operation := &Op{
		cmdName: cmdName,
		short:   short,
		long:    long,
		op:      op,
		subOps:  []Op{},
	}
	for _, subOp := range subOps {
		operation.subOps = append(operation.subOps, *subOp)
	}
	return operation
}

func (o *Op) Handle(args []string) error {
	if len(args) == 0 {
		return o.op(args)
	}
	for _, op := range o.subOps {
		if op.cmdName == args[0] {
			return op.Handle(args[1:])
		}
	}
	return o.op(args)
}
