package interp

import "github.com/havrydotdev/golox/token"

// Visitor pattern doesn't really work in golang
// so we have to use object algebras
// https://www.cs.utexas.edu/%7Ewcook/Drafts/2012/ecoop2012.pdf
//
// E is for expression, S is for statement
type Alg[E any, S any] interface {
	Grouping(expr E) E
	Literal(value any) E
	Variable(name token.Token) E
	Get(name token.Token, expr E) E
	Unary(op token.Token, right E) E
	Assign(name token.Token, value E) E
	Binary(op token.Token, left, right E) E
	Logical(op token.Token, left, right E) E
	Call(callee E, paren token.Token, args []E) E
	Set(object E, name token.Token, value E) E

	Block(stmts []S) S
	While(cond E, body S) S
	ExprStatement(expr E) S
	If(cond E, then S, _else S) S
	Var(name token.Token, init E) S
	Return(keyword token.Token, value E) S
	Class(name token.Token, methods []S) S
	Function(name token.Token, params []token.Token, body []S) S

	NilExpr() E
	NilStmt() S
}
