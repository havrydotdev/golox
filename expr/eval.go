package expr

import (
	"fmt"

	env "github.com/havrydotdev/golox/environment"
	"github.com/havrydotdev/golox/token"
)

// TODO: add special type for lox objects
type ExpEvaluator interface {
	Eval() (any, error)
}

type StmtEvaluator interface {
	Eval() error
}

type ExpEvalFunc func() (any, error)
type StmtEvalFunc func() error

func (fn ExpEvalFunc) Eval() (any, error) {
	return fn()
}

func (fn StmtEvalFunc) Eval() error {
	return fn()
}

type EvalExpr struct {
	environment *env.Env
}

func NewEval() ExprAlg[ExpEvaluator, StmtEvaluator] {
	return &EvalExpr{environment: env.New()}
}

func (e *EvalExpr) While(cond ExpEvaluator, body StmtEvaluator) StmtEvaluator {
	return StmtEvalFunc(func() error {
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

func (e *EvalExpr) Logical(op *token.Token, left, right ExpEvaluator) ExpEvaluator {
	return ExpEvalFunc(func() (any, error) {
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

func (e *EvalExpr) If(cond ExpEvaluator, then StmtEvaluator, _else StmtEvaluator) StmtEvaluator {
	return StmtEvalFunc(func() error {
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

func (e *EvalExpr) Block(stmts []StmtEvaluator) StmtEvaluator {
	return StmtEvalFunc(func() error {
		prev := e.environment

		e.environment = env.NewChild(prev)

		for _, stmt := range stmts {
			err := stmt.Eval()
			if err != nil {
				return err
			}
		}

		e.environment = prev
		return nil
	})
}

func (e *EvalExpr) Assign(name *token.Token, value ExpEvaluator) ExpEvaluator {
	return ExpEvalFunc(func() (any, error) {
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

func (e *EvalExpr) Variable(name *token.Token) ExpEvaluator {
	return ExpEvalFunc(func() (any, error) {
		val, ok := e.environment.Get(name.Lexeme)
		if !ok {
			return nil, fmt.Errorf("undefined variable %s", name.Lexeme)
		}

		return val, nil
	})
}

// this method is used as nil value in parser
// TODO: better initial value handling?
func (e *EvalExpr) Var(name *token.Token, init *ExpEvaluator) StmtEvaluator {
	return StmtEvalFunc(func() error {
		var err error
		var value any
		if init != nil {
			value, err = (*init).Eval()
		}

		if err != nil {
			return err
		}

		e.environment.Define(name.Lexeme, value)

		return nil
	})
}

func (*EvalExpr) ExprStatement(expr ExpEvaluator) StmtEvaluator {
	return StmtEvalFunc(func() error {
		_, err := expr.Eval()
		return err
	})
}

// TODO: move to stdlib (later)
func (*EvalExpr) Print(expr ExpEvaluator) StmtEvaluator {
	return StmtEvalFunc(func() error {
		value, err := expr.Eval()
		if err != nil {
			return err
		}

		switch value.(type) {
		case float32:
			fmt.Printf("%f\n", value)
		default:
			fmt.Println(value)
		}

		return nil
	})
}

func (*EvalExpr) Literal(value any) ExpEvaluator {
	return ExpEvalFunc(func() (any, error) {
		return value, nil
	})
}

func (*EvalExpr) Grouping(expr ExpEvaluator) ExpEvaluator {
	return ExpEvalFunc(func() (any, error) {
		return expr.Eval()
	})
}

func (*EvalExpr) Unary(op *token.Token, right ExpEvaluator) ExpEvaluator {
	return ExpEvalFunc(func() (any, error) {
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

func (*EvalExpr) Binary(op *token.Token, left, right ExpEvaluator) ExpEvaluator {
	return ExpEvalFunc(func() (any, error) {
		l, err := left.Eval()
		if err != nil {
			return nil, err
		}

		r, err := right.Eval()
		if err != nil {
			return nil, err
		}

		switch op.Kind {
		case token.Greater:
			lfloat, rfloat, err := checkNums(l, r)
			return lfloat > rfloat, err
		case token.GreaterEqual:
			lfloat, rfloat, err := checkNums(l, r)
			return lfloat >= rfloat, err
		case token.Less:
			lfloat, rfloat, err := checkNums(l, r)
			return lfloat < rfloat, err
		case token.LessEqual:
			lfloat, rfloat, err := checkNums(l, r)
			return lfloat <= rfloat, err

		case token.BangEqual:
			return !isEqual(l, r), nil
		case token.EqualEqual:
			return isEqual(l, r), nil

		case token.Minus:
			lfloat, rfloat, err := checkNums(l, r)
			return lfloat - rfloat, err
		case token.Slash:
			lfloat, rfloat, err := checkNums(l, r)
			return lfloat / rfloat, err
		case token.Star:
			lfloat, rfloat, err := checkNums(l, r)
			return lfloat * rfloat, err

		case token.Plus:
			ls, okl := l.(string)
			rs, okr := r.(string)
			if okl && okr {
				return ls + rs, nil
			}

			lfloat, rfloat, err := checkNums(l, r)
			return lfloat + rfloat, err
		}

		return nil, fmt.Errorf("Unexpected token %s", op.Lexeme)
	})
}

func isTruthy(value any) bool {
	if value == nil {
		return false
	}

	switch value := value.(type) {
	case bool:
		return value
	}

	return true
}

func isEqual(left, right any) bool {
	if left == nil && right == nil {
		return true
	}

	if left == nil {
		return false
	}

	return left == right
}

func checkNums(left, right any) (float32, float32, error) {
	l, okl := left.(float32)
	r, okr := right.(float32)
	if !okl {
		return 0, 0, fmt.Errorf("Expected number, got %v", left)
	}

	if !okr {
		return 0, 0, fmt.Errorf("Expected number, got %v", right)
	}

	return l, r, nil
}
