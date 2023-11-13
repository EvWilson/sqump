package cmder

type Op struct {
	cmdName string
	short   string
	long    string
	op      func()
	subOps  []Op
}

func NewOp(
	cmdName,
	short,
	long string,
	op func(),
) *Op {
	return &Op{
		cmdName: cmdName,
		short:   short,
		long:    long,
		op:      op,
		subOps:  make([]Op, 0),
	}
}

func (o *Op) Register(op *Op) {
	o.subOps = append(o.subOps, *op)
}
