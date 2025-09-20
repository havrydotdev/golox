package parser

import (
	"errors"
	"fmt"
	"slices"

	interp "github.com/havrydotdev/golox/interpreter"
	"github.com/havrydotdev/golox/token"
)

type ErrorSet[E any] struct {
	name   token.Token
	object E
}

func (e ErrorSet[E]) Error() string {
	return "set error"
}

type Parser[E any, S any] struct {
	current uint
	errors  []error
	tokens  []token.Token
	alg     interp.Alg[E, S]
}

func New[E any, S any](tokens []token.Token, alg interp.Alg[E, S]) *Parser[E, S] {
	return &Parser[E, S]{tokens: tokens, alg: alg, current: 0}
}

func (p *Parser[E, S]) Parse() ([]S, []error) {
	var stmts []S
	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			p.errors = append(p.errors, err)
			p.synchronize()
			continue
		}

		stmts = append(stmts, stmt)
	}

	return stmts, p.errors
}

func (p *Parser[E, S]) declaration() (S, error) {
	switch {
	case p.match(token.Class):
		return p.classDeclaration()
	case p.match(token.Fun):
		return p.function("function")
	case p.match(token.Var):
		return p.varDeclaration()
	default:
		return p.statement()
	}
}

func (p *Parser[E, S]) classDeclaration() (S, error) {
	name, err := p.consume(token.Identifier, "expected class name.")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	_, err = p.consume(token.LeftBrace, "expected '}' before class body.")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	var methods []S
	for !p.check(token.RightBrace) && !p.isAtEnd() {
		fun, err := p.function("method")
		if err != nil {
			return p.alg.NilStmt(), err
		}

		methods = append(methods, fun)
	}

	_, err = p.consume(token.RightBrace, "expected '}' after class body.")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	return p.alg.Class(name, methods), nil
}

func (p *Parser[E, S]) function(kind string) (S, error) {
	name, err := p.consume(token.Identifier, fmt.Sprintf("expected %s name.", kind))
	if err != nil {
		return p.alg.NilStmt(), err
	}

	_, err = p.consume(token.LeftParen, fmt.Sprintf("expected '(' after %s name.", kind))
	if err != nil {
		return p.alg.NilStmt(), err
	}

	var params []token.Token
	if !p.check(token.RightParen) {
		for {
			if len(params) >= 255 {
				return p.alg.NilStmt(), errors.New("cant have more than 255 parameters.")
			}

			paramName, err := p.consume(token.Identifier, "expected identifier name.")
			if err != nil {
				return p.alg.NilStmt(), err
			}

			params = append(params, paramName)

			if !p.match(token.Comma) {
				break
			}
		}
	}

	_, err = p.consume(token.RightParen, "expected ')' after parameters")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	_, err = p.consume(token.LeftBrace, "expected '{' before function body")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	body, err := p.blockStmts()
	if err != nil {
		return p.alg.NilStmt(), err
	}

	_, err = p.consume(token.RightBrace, "expected '}' after function body.")

	return p.alg.Function(name, params, body), nil
}

func (p *Parser[E, S]) varDeclaration() (S, error) {
	name, err := p.consume(token.Identifier, "Expected variable name.")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	init := p.alg.Literal(nil)
	if p.match(token.Equal) {
		init, err = p.expression()
		if err != nil {
			return p.alg.NilStmt(), err
		}
	}

	_, err = p.consume(token.Semicolon, "Expected ';' after variable declaration.")

	return p.alg.Var(name, init), err
}

// how tf am i supposed to figure this out
// we cant get kind of function (possible with kinds in interp package)
// we cant access p.alg.Get fields when its returned (because its a function)
// the only thing i see here is to return error for now
func (p *Parser[E, S]) assignment() (E, error) {
	expr, err := p.or()
	if err != nil {
		if err, ok := err.(ErrorSet[E]); ok {
			if p.match(token.Equal) {
				value, errNew := p.assignment()
				if errNew != nil {
					return p.alg.NilExpr(), err
				}

				return p.alg.Set(err.object, err.name, value), nil
			}
		}

		return p.alg.Literal(nil), err
	}

	name := p.previous()
	if p.match(token.Equal) {
		value, err := p.assignment()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		return p.alg.Assign(name, value), nil
	}

	return expr, nil
}

func (p *Parser[E, S]) or() (E, error) {
	expr, err := p.and()
	if err != nil {
		return p.alg.Literal(nil), err
	}

	for p.match(token.Or) {
		op := p.previous()
		right, err := p.and()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		expr = p.alg.Logical(op, expr, right)
	}

	return expr, nil
}

func (p *Parser[E, S]) and() (E, error) {
	expr, err := p.equality()
	if err != nil {
		return p.alg.Literal(nil), err
	}

	for p.match(token.Or) {
		op := p.previous()
		right, err := p.and()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		expr = p.alg.Logical(op, expr, right)
	}

	return expr, nil
}

func (p *Parser[E, S]) statement() (S, error) {
	switch {
	case p.match(token.For):
		return p.forStatement()
	case p.match(token.Return):
		return p.returnStatement()
	case p.match(token.If):
		return p.ifStatement()
	case p.match(token.While):
		return p.whileStatement()
	case p.match(token.LeftBrace):
		return p.block()
	default:
		return p.expressionStatement()
	}
}

func (p *Parser[E, S]) returnStatement() (S, error) {
	keyword := p.previous()

	var value E
	var err error
	if !p.check(token.Semicolon) {
		value, err = p.expression()
		if err != nil {
			return p.alg.NilStmt(), err
		}
	}

	_, err = p.consume(token.Semicolon, "expected ';' after return")

	return p.alg.Return(keyword, value), err
}

func (p *Parser[E, S]) forStatement() (S, error) {
	_, err := p.consume(token.LeftParen, "expected '(' after 'for'.")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	init := p.alg.ExprStatement(p.alg.Literal(nil))
	if p.match(token.Var) {
		init, err = p.varDeclaration()
	} else if !p.match(token.Semicolon) {
		init, err = p.expressionStatement()
	}

	if err != nil {
		return p.alg.NilStmt(), err
	}

	cond := p.alg.Literal(true)
	if !p.check(token.Semicolon) {
		cond, err = p.expression()
		if err != nil {
			return p.alg.NilStmt(), err
		}
	}

	_, err = p.consume(token.Semicolon, "expected ';' after loop condition")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	incr := p.alg.Literal(nil)
	if !p.check(token.RightParen) {
		incr, err = p.expression()
		if err != nil {
			return p.alg.NilStmt(), err
		}
	}

	_, err = p.consume(token.RightParen, "expected ')' after clause")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	body, err := p.statement()
	if err != nil {
		return p.alg.NilStmt(), err
	}

	body = p.alg.Block([]S{body, p.alg.ExprStatement(incr)})
	body = p.alg.While(cond, body)
	body = p.alg.Block([]S{init, body})

	return body, nil
}

func (p *Parser[E, S]) whileStatement() (S, error) {
	_, err := p.consume(token.LeftParen, "expected '(' after 'while'.")
	if err != nil {
		return p.alg.NilStmt(), nil
	}

	cond, err := p.expression()
	if err != nil {
		return p.alg.NilStmt(), nil
	}

	_, err = p.consume(token.RightParen, "expected ')' after condition.")
	if err != nil {
		return p.alg.NilStmt(), nil
	}

	body, err := p.statement()

	return p.alg.While(cond, body), err
}

func (p *Parser[E, S]) ifStatement() (S, error) {
	_, err := p.consume(token.LeftParen, "expected '(' after 'if'.")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	cond, err := p.expression()
	if err != nil {
		return p.alg.NilStmt(), err
	}

	_, err = p.consume(token.RightParen, "expected ')' after if condition")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	then, err := p.statement()
	if err != nil {
		return p.alg.NilStmt(), err
	}

	var _else S
	if p.match(token.Else) {
		_else, err = p.statement()
	}

	return p.alg.If(cond, then, _else), err
}

func (p *Parser[E, S]) blockStmts() ([]S, error) {
	var stmts []S
	for !p.check(token.RightBrace) && !p.isAtEnd() {
		exp, err := p.declaration()
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, exp)
	}

	return stmts, nil
}

func (p *Parser[E, S]) block() (S, error) {
	stmts, err := p.blockStmts()
	if err != nil {
		return p.alg.NilStmt(), err
	}

	_, err = p.consume(token.RightBrace, "expected '}' after block.")
	return p.alg.Block(stmts), err
}

func (p *Parser[E, S]) expressionStatement() (S, error) {
	expr, err := p.expression()
	if err != nil {
		return p.alg.NilStmt(), err
	}

	_, err = p.consume(token.Semicolon, "Expected ';' after expression")
	if err != nil {
		return p.alg.NilStmt(), err
	}

	return p.alg.ExprStatement(expr), nil
}

func (p *Parser[E, S]) expression() (E, error) {
	return p.assignment()
}

func (p *Parser[E, S]) equality() (E, error) {
	expr, err := p.comparison()
	if err != nil {
		return p.alg.Literal(nil), err
	}

	for p.match(token.BangEqual, token.EqualEqual) {
		op := p.previous()
		right, err := p.comparison()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		expr = p.alg.Binary(op, expr, right)
	}

	return expr, nil
}

func (p *Parser[E, S]) comparison() (E, error) {
	expr, err := p.term()
	if err != nil {
		return p.alg.Literal(nil), err
	}

	for p.match(token.Greater, token.GreaterEqual, token.Less, token.LessEqual) {
		op := p.previous()
		right, err := p.term()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		expr = p.alg.Binary(op, expr, right)
	}

	return expr, nil
}

func (p *Parser[E, S]) term() (E, error) {
	expr, err := p.factor()
	if err != nil {
		return p.alg.Literal(nil), err
	}

	for p.match(token.Minus, token.Plus) {
		op := p.previous()
		right, err := p.factor()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		expr = p.alg.Binary(op, expr, right)
	}

	return expr, err
}

func (p *Parser[E, S]) factor() (E, error) {
	expr, err := p.unary()
	if err != nil {
		return p.alg.Literal(nil), err
	}

	for p.match(token.Slash, token.Star) {
		op := p.previous()
		right, err := p.unary()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		expr = p.alg.Binary(op, expr, right)
	}

	return expr, nil
}

func (p *Parser[E, S]) unary() (E, error) {
	if p.match(token.Bang, token.Minus) {
		right, err := p.unary()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		return p.alg.Unary(p.previous(), right), nil
	}

	return p.call()
}

func (p *Parser[E, S]) call() (E, error) {
	expr, err := p.primary()
	if err != nil {
		return p.alg.NilExpr(), err
	}

	for {
		if p.match(token.LeftParen) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return p.alg.NilExpr(), err
			}
		} else if p.match(token.Dot) {
			name, err := p.consume(token.Identifier, "expected property name after '.'.")
			if err != nil {
				return p.alg.NilExpr(), err
			}

			if p.peek().Kind == token.Equal {
				return expr, ErrorSet[E]{name, expr}
			}

			expr = p.alg.Get(name, expr)
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser[E, S]) finishCall(callee E) (E, error) {
	var args []E

	if !p.check(token.RightParen) {
		for {
			if len(args) >= 255 {
				return p.alg.Literal(nil), errors.New("can't have more than 255 arguments")
			}

			expr, err := p.expression()
			if err != nil {
				return p.alg.Literal(nil), err
			}

			args = append(args, expr)

			if !p.match(token.Comma) {
				break
			}
		}
	}

	paren, err := p.consume(token.RightParen, "expected ')' after arguments.")
	if err != nil {
		return p.alg.Literal(nil), err
	}

	return p.alg.Call(callee, paren, args), nil
}

func (p *Parser[E, S]) primary() (E, error) {
	switch {
	case p.match(token.Identifier):
		return p.alg.Variable(p.previous()), nil
	case p.match(token.False):
		return p.alg.Literal(false), nil
	case p.match(token.True):
		return p.alg.Literal(true), nil
	case p.match(token.Nil):
		return p.alg.Literal(nil), nil
	case p.match(token.Number, token.String):
		return p.alg.Literal(p.previous().Literal), nil
	case p.match(token.LeftParen):
		expr, err := p.expression()
		if err != nil {
			return p.alg.Literal(nil), err
		}

		_, err = p.consume(token.RightParen, "Expect ')' after expression.")
		if err != nil {
			return p.alg.Literal(nil), err
		}

		return p.alg.Grouping(expr), nil
	}

	return p.alg.Literal(nil), fmt.Errorf("Unexpected token %s at %d", p.peek().Lexeme, p.peek().Line)
}

// synchronize method moves cursor
// to the next statement
func (p *Parser[E, S]) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Kind == token.Semicolon {
			return
		}

		switch p.peek().Kind {
		case token.Class, token.Fun, token.Var, token.For, token.If, token.While, token.Return:
			return
		}

		p.advance()
	}
}

func (p *Parser[E, S]) consume(kind token.Kind, message string) (token.Token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}

	return token.NilV, errors.New(message)
}

func (p *Parser[E, S]) match(kinds ...token.Kind) bool {
	if slices.ContainsFunc(kinds, p.check) {
		p.advance()
		return true
	}

	return false
}

func (p *Parser[E, S]) check(kind token.Kind) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Kind == kind
}

func (p *Parser[E, S]) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser[E, S]) isAtEnd() bool {
	return p.peek().Kind == token.Eof
}

func (p *Parser[E, S]) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser[E, S]) previous() token.Token {
	return p.tokens[p.current-1]
}
