package cmder

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"text/tabwriter"
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
	r := &Root{
		description: description,
		writer:      writer,
	}

	r.Register(NewOp(
		"help",
		"help <optional: top level command>",
		"Prints a help summary of top level commands, or options tree of command if one given",
		func(args []string) error {
			if len(args) == 0 {
				r.PrintUsage()
			} else {
				r.PrintExtendedUsage(args)
			}
			return nil
		},
	))

	return r
}

func (r *Root) Register(ops ...*Op) {
	for _, op := range ops {
		r.ops = append(r.ops, *op)
	}
}

func (r *Root) PrintUsage() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	for _, subOp := range r.ops {
		_, _ = w.Write([]byte(fmt.Sprintf("%s\t%s\n", subOp.short, subOp.long)))
	}
	_ = w.Flush()
	fmt.Println()
}

func (r *Root) PrintExtendedUsage(args []string) {
	if len(args) != 1 {
		fmt.Printf("expected 1 arg to `help`, got: %d\n", len(args))
		return
	}
	cmdName := args[0]
	write := func(s string) {
		_, _ = r.writer.Write([]byte(s))
	}
	write(fmt.Sprintf("%s\n\n", r.description))
	for _, subOp := range r.ops {
		if subOp.cmdName == cmdName {
			printOps(r.writer, &subOp, "")
			write("\n")
			return
		}
	}
	fmt.Printf("did not find listing for command '%s'\n", cmdName)
}

func printOps(writer io.Writer, op *Op, offset string) {
	write := func(s string) {
		_, _ = writer.Write([]byte(s))
	}
	write(fmt.Sprintf("%s%s    %s\n", offset, op.short, op.long))
	for _, subOp := range op.subOps {
		printOps(writer, &subOp, offset+"    ")
	}
}

func (r *Root) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected at least one argument, got none")
	}

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
