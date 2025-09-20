package eval

type Class struct {
	Name string
}

func (c Class) Arity() uint8 {
	return 0
}

func (c Class) Call(e *Evaluator, args []any) (any, error) {
	return Instance{Class: c, fields: map[string]any{"hello": ""}}, nil
}

func (c Class) String() string {
	return c.Name
}
