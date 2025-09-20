package eval

import (
	"errors"
	"fmt"

	env "github.com/havrydotdev/golox/environment"
	interp "github.com/havrydotdev/golox/interpreter"
	"github.com/havrydotdev/golox/token"
)

var (
	ErrNilValue = errors.New("internal error: exp/stmt is nil, cannot invoke")
)

type Return struct {
	value any
}

func (r Return) Error() string {
	return "return statement"
}

// TODO: add special type for lox objects
type ExpEvaluator interface {
	Eval() (any, error)
}

type StmtEvaluator interface {
	Eval() error
}

type expEvalFunc func() (any, error)
type stmtEvalFunc func() error

func (fn expEvalFunc) Eval() (any, error) {
	return fn()
}

func (fn stmtEvalFunc) Eval() error {
	return fn()
}

type Evaluator struct {
	globals     *env.Env
	environment *env.Env
}

func New() interp.Alg[ExpEvaluator, StmtEvaluator] {
	globals := newGlobals()

	return &Evaluator{environment: globals, globals: globals}
}

func (e *Evaluator) Set(object ExpEvaluator, name token.Token, value ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		obj, err := object.Eval()
		if err != nil {
			return nil, err
		}

		inst, ok := obj.(Instance)
		if !ok {
			return nil, errors.New("only instances have fields")
		}

		val, err := value.Eval()
		if err != nil {
			return nil, err
		}

		inst.Set(name.Lexeme, val)
		return val, nil
	})
}

func (e *Evaluator) Get(name token.Token, expr ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		rawInst, err := expr.Eval()
		if err != nil {
			return nil, err
		}

		inst, ok := rawInst.(Instance)
		if !ok {
			return nil, errors.New("only instances have properties.")
		}

		val, ok := inst.Get(name.Lexeme)
		if !ok {
			return nil, errors.New("unknown key")
		}

		return val, nil
	})
}

func (e *Evaluator) Class(name token.Token, methods []StmtEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		e.environment.Define(name.Lexeme, Class{Name: name.Lexeme})
		return nil
	})
}

func (e *Evaluator) Return(keyword token.Token, value ExpEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		var val any
		var err error
		if value != nil {
			val, err = value.Eval()
		}

		if err != nil {
			return err
		}

		return Return{value: val}
	})
}

func (e *Evaluator) Function(name token.Token, params []token.Token, body []StmtEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		e.environment.Define(name.Lexeme, Function{params, body, e.environment})
		return nil
	})
}

func (e *Evaluator) Call(callee ExpEvaluator, paren token.Token, args []ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		callee, err := callee.Eval()
		if err != nil {
			return nil, err
		}

		var arguments []any
		for _, arg := range args {
			argValue, err := arg.Eval()
			if err != nil {
				return nil, err
			}

			arguments = append(arguments, argValue)
		}

		fun, ok := callee.(Callable)
		if !ok {
			return nil, errors.New("callee is not callable")
		}

		if len(arguments) != int(fun.Arity()) {
			return nil, fmt.Errorf("expected %d arguments, got %d", fun.Arity(), len(arguments))
		}

		return fun.Call(e, arguments)
	})
}

func (e *Evaluator) While(cond ExpEvaluator, body StmtEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		for {
			c, err := cond.Eval()
			if err != nil {
				return err
			}

			if !isTruthy(c) {
				break
			}

			err = body.Eval()
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (e *Evaluator) Logical(op token.Token, left, right ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		l, err := left.Eval()
		if err != nil {
			return nil, err
		}

		if op.Kind == token.Or {
			if isTruthy(l) {
				return l, nil
			}
		} else {
			if !isTruthy(l) {
				return l, nil
			}
		}

		return right.Eval()
	})
}

func (e *Evaluator) If(cond ExpEvaluator, then StmtEvaluator, _else StmtEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		condRes, err := cond.Eval()
		if err != nil {
			return err
		}

		if isTruthy(condRes) {
			err = then.Eval()
		} else if _else != nil {
			err = _else.Eval()
		}

		return err
	})
}

func (e *Evaluator) Block(stmts []StmtEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		_, err := e.executeBlock(stmts, env.NewChild(e.environment))
		return err
	})
}

func (e *Evaluator) executeBlock(stmts []StmtEvaluator, environment *env.Env) (any, error) {
	prev := e.environment
	e.environment = environment
	defer func() { e.environment = prev }()

	for _, stmt := range stmts {
		err := stmt.Eval()

		ret, ok := err.(Return)
		if ok {
			return ret.value, nil
		}

		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (e *Evaluator) Assign(name token.Token, value ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		val, err := value.Eval()
		if err != nil {
			return nil, err
		}

		ok := e.environment.Assign(name.Lexeme, val)
		if !ok {
			return nil, fmt.Errorf("undefined variable %s", name.Lexeme)
		}

		return val, nil
	})
}

func (e *Evaluator) Variable(name token.Token) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		val, ok := e.environment.Get(name.Lexeme)
		if !ok {
			return nil, fmt.Errorf("undefined variable %s", name.Lexeme)
		}

		return val, nil
	})
}

// this method is used as nil value in parser
// TODO: better initial value handling?
func (e *Evaluator) Var(name token.Token, init ExpEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		var err error
		var value any
		if init != nil {
			value, err = init.Eval()
		}

		if err != nil {
			return err
		}

		e.environment.Define(name.Lexeme, value)

		return nil
	})
}

func (*Evaluator) ExprStatement(expr ExpEvaluator) StmtEvaluator {
	return stmtEvalFunc(func() error {
		_, err := expr.Eval()
		return err
	})
}

func (*Evaluator) Literal(value any) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		return value, nil
	})
}

func (*Evaluator) Grouping(expr ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		return expr.Eval()
	})
}

func (*Evaluator) Unary(op token.Token, right ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		right, err := right.Eval()
		if err != nil {
			return nil, err
		}

		switch op.Kind {
		case token.Minus:
			rfloat, ok := right.(float32)
			if !ok {
				return nil, fmt.Errorf("Expected number, got %v", right)
			}

			return -rfloat, nil
		case token.Bang:
			return !isTruthy(right), nil
		}

		return nil, fmt.Errorf("Unexpected operator %s", op.Lexeme)
	})
}

func (*Evaluator) Binary(op token.Token, left, right ExpEvaluator) ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		l, err := left.Eval()
		if err != nil {
			return nil, err
		}

		r, err := right.Eval()
		if err != nil {
			return nil, err
		}

		switch lparsed := l.(type) {
		case string:
			rparsed, ok := r.(string)
			if !ok {
				return nil, fmt.Errorf("expected string, got %v", r)
			}

			if op.Kind == token.Plus {
				return lparsed + rparsed, nil
			}
		case float32:
			rparsed, ok := r.(float32)
			if !ok {
				return nil, fmt.Errorf("expected number, got %v", r)
			}

			switch op.Kind {
			case token.Greater:
				return lparsed > rparsed, err
			case token.GreaterEqual:
				return lparsed >= rparsed, err
			case token.Less:
				return lparsed < rparsed, err
			case token.LessEqual:
				return lparsed <= rparsed, err

			case token.BangEqual:
				return !isEqual(l, r), nil
			case token.EqualEqual:
				return isEqual(l, r), nil

			case token.Minus:
				return lparsed - rparsed, err
			case token.Slash:
				return lparsed / rparsed, err
			case token.Star:
				return lparsed * rparsed, err

			case token.Plus:
				return lparsed + rparsed, err
			}
		}

		return nil, fmt.Errorf("Unexpected token %s", op.Lexeme)
	})
}

func (*Evaluator) NilExpr() ExpEvaluator {
	return expEvalFunc(func() (any, error) {
		return nil, ErrNilValue
	})
}

func (*Evaluator) NilStmt() StmtEvaluator {
	return stmtEvalFunc(func() error {
		return ErrNilValue
	})
}
