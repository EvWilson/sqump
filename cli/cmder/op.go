package cmder

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

type ContextKey int

const (
	OverrideContextKey ContextKey = iota
	ReadonlyContextKey
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

func NewNoopHandler(name string) func(context.Context, []string) error {
	return func(_ context.Context, _ []string) error {
		return NewErrNoop(name)
	}
}

type Root struct {
	description  string
	writer       io.Writer
	ops          []Op
	envOverrides map[string]string
}

func NewRoot(
	description string,
	writer io.Writer,
) *Root {
	r := &Root{
		description:  description,
		writer:       writer,
		ops:          make([]Op, 0),
		envOverrides: make(map[string]string),
	}

	r.Register(NewOp(
		"help",
		"help <optional: top level command>",
		"Prints a help summary of top level commands, or options tree of command if one given",
		func(ctx context.Context, args []string) error {
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

	args, overrides, err := ExtractOverrideMappings(args, "-e")
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), OverrideContextKey, overrides)
	ctx = context.WithValue(ctx, ReadonlyContextKey, isReadonlyMode(args))
	for _, op := range r.ops {
		if op.cmdName == args[0] {
			return op.Handle(ctx, args[1:])
		}
	}

	return fmt.Errorf("did not find handler for subcommand '%s'", args[0])
}

func ExtractOverrideMappings(startArgs []string, delimiter string) ([]string, map[string]string, error) {
	args := make([]string, len(startArgs))
	copy(args, startArgs)
	mappings := make(map[string]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg != delimiter {
			continue
		}
		if len(args) <= i+1 {
			args = args[:len(args)-1]
			break
		}
		mapping := args[i+1]
		pair := strings.Split(mapping, "=")
		if len(pair) != 2 {
			return nil, nil, fmt.Errorf("error: '%s' produced invalid map split length '%d'", mapping, len(pair))
		}
		k, v := pair[0], pair[1]
		mappings[k] = v
		args = append(args[:i], args[i+2:]...)
		i--
	}
	return args, mappings, nil
}

func isReadonlyMode(args []string) bool {
	readonly := false
	for _, arg := range args {
		if arg == "--readonly" {
			readonly = true
			break
		}
	}
	return readonly
}

type Op struct {
	cmdName string
	short   string
	long    string
	op      func(ctx context.Context, args []string) error
	subOps  []Op
}

func NewOp(
	cmdName,
	short,
	long string,
	op func(ctx context.Context, args []string) error,
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

func (o *Op) Handle(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return o.op(ctx, args)
	}
	for _, op := range o.subOps {
		if op.cmdName == args[0] {
			return op.Handle(ctx, args[1:])
		}
	}
	return o.op(ctx, args)
}
