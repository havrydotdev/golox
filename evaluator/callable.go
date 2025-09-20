package eval

import (
	env "github.com/havrydotdev/golox/environment"
	"github.com/havrydotdev/golox/token"
)

type Callable interface {
	Arity() uint8
	Call(e *Evaluator, args []any) (any, error)
}

type NativeFun struct {
	arity uint8
	call  func(e *Evaluator, args []any) (any, error)
}

type Function struct {
	params  []token.Token
	body    []StmtEvaluator
	closure *env.Env
}

func NewNativeFun(arity uint8, call func(e *Evaluator, args []any) (any, error)) Callable {
	return NativeFun{arity, call}
}

func (c NativeFun) Arity() uint8 {
	return c.arity
}

func (c NativeFun) Call(e *Evaluator, args []any) (any, error) {
	return c.call(e, args)
}

func (f Function) Arity() uint8 {
	return uint8(len(f.params))
}

func (f Function) Call(e *Evaluator, args []any) (any, error) {
	env := env.NewChild(f.closure)
	for i, param := range f.params {
		env.Define(param.Lexeme, args[i])
	}

	return e.executeBlock(f.body, env)
}
