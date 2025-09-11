package expr

import "github.com/havrydotdev/golox/token"

// Visitor pattern doesn't really work in golang
// so we have to use object algebras
// https://www.cs.utexas.edu/%7Ewcook/Drafts/2012/ecoop2012.pdf
//
// E is for expression, S is for statement
// TODO: this interface handles both expressions
// and statements, fix naming
type ExprAlg[E any, S any] interface {
	Literal(value any) E
	Grouping(expr E) E
	Variable(name *token.Token) E
	Unary(op *token.Token, right E) E
	Assign(name *token.Token, value E) E
	Binary(op *token.Token, left, right E) E
	Logical(op *token.Token, left, right E) E

	Print(expr E) S
	Block(stmts []S) S
	While(cond E, body S) S
	ExprStatement(expr E) S
	If(cond E, then S, _else S) S
	Var(name *token.Token, init *E) S
}
