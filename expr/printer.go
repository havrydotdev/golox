package expr

import (
	"fmt"

	"github.com/havrydotdev/golox/token"
)

type Printer interface {
	Print() string
}

// implements Printer
type PrintFunc func() string

func (fn PrintFunc) Print() string {
	return fn()
}

// implements ExprAlg[Printer]
type PrintExpr struct{}

// func NewPrinter() ExprAlg[Printer, Printer] {
// 	return &PrintExpr{}
// }

func (*PrintExpr) If(cond Printer, then Printer, _else Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize("if", cond, then, _else)
	})
}

func (*PrintExpr) Block(stmts []Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize("block", stmts...)
	})
}

func (*PrintExpr) Variable(name *token.Token) Printer {
	return PrintFunc(func() string {
		return name.Lexeme
	})
}

func (*PrintExpr) Assign(name *token.Token, value Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize(fmt.Sprintf("assign %s", name.Lexeme), value)
	})
}

func (*PrintExpr) Var(name *token.Token, init *Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize(fmt.Sprintf("var %s", name.Lexeme), *init)
	})
}

func (*PrintExpr) Literal(value any) Printer {
	return PrintFunc(func() string {
		return fmt.Sprintf("%v", value)
	})
}

func (*PrintExpr) Grouping(expr Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize("group", expr)
	})
}

func (*PrintExpr) Unary(op *token.Token, right Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize(op.Lexeme, right)
	})
}

func (*PrintExpr) Binary(op *token.Token, left, right Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize(op.Lexeme, left, right)
	})
}

func (*PrintExpr) Print(expr Printer) Printer {
	return PrintFunc(func() string {
		return parenthesize("print", expr)
	})
}

func (*PrintExpr) ExprStatement(expr Printer) Printer {
	return PrintFunc(func() string {
		return expr.Print()
	})
}
