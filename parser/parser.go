package parser

import (
	"errors"
	"fmt"
	"slices"

	"github.com/havrydotdev/golox/expr"
	"github.com/havrydotdev/golox/token"
)

type Parser[E any, S any] struct {
	tokens  []*token.Token
	current uint
	errors  []error

	alg expr.ExprAlg[E, S]
}

func New[E any, S any](tokens []*token.Token, alg expr.ExprAlg[E, S]) *Parser[E, S] {
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
	if p.match(token.Var) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *Parser[E, S]) varDeclaration() (S, error) {
	name, err := p.consume(token.Identifier, "Expected variable name.")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	var init E
	if p.match(token.Equal) {
		init, err = p.expression()
		if err != nil {
			return p.alg.Var(nil, nil), err
		}
	}

	_, err = p.consume(token.Semicolon, "Expected ';' after variable declaration.")

	return p.alg.Var(name, &init), err
}

func (p *Parser[E, S]) assignment() (E, error) {
	expr, err := p.or()
	if err != nil {
		return p.alg.Literal(nil), err
	}

	// modified this part because
	// we cant check type of expression
	// during parsing (for now)
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
	case p.match(token.If):
		return p.ifStatement()
	case p.match(token.While):
		return p.whileStatement()
	case p.match(token.LeftBrace):
		return p.block()
	case p.match(token.Print):
		return p.printStatement()
	default:
		return p.expressionStatement()
	}
}

// it doesn't work
// TODO: refactor these methods to return pointer to eval functions?
func (p *Parser[E, S]) forStatement() (S, error) {
	_, err := p.consume(token.LeftParen, "expected '(' after 'for'.")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	var ini S
	var init *S
	if p.match(token.Var) {
		ini, err = p.varDeclaration()
		init = &ini
	} else if !p.match(token.Semicolon) {
		ini, err = p.expressionStatement()
		init = &ini
	}

	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	var con E
	var cond *E
	if !p.check(token.Semicolon) {
		con, err = p.expression()
		cond = &con
	}

	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	_, err = p.consume(token.Semicolon, "expected ';' after loop condition")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	var incr *E
	if !p.check(token.RightParen) {
		var inc E
		inc, err = p.expression()
		incr = &inc
	}

	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	_, err = p.consume(token.RightParen, "expected ')' after clause")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	body, err := p.statement()
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	if incr != nil {
		body = p.alg.Block([]S{body, p.alg.ExprStatement(*incr)})
	}

	if cond == nil {
		con = p.alg.Literal(true)
	}

	body = p.alg.While(con, body)

	if init != nil {
		body = p.alg.Block([]S{*init, body})
	}

	return body, nil
}

func (p *Parser[E, S]) whileStatement() (S, error) {
	_, err := p.consume(token.LeftParen, "expected '(' after 'while'.")
	if err != nil {
		return p.alg.Var(nil, nil), nil
	}

	cond, err := p.expression()
	if err != nil {
		return p.alg.Var(nil, nil), nil
	}

	_, err = p.consume(token.RightParen, "expected ')' after condition.")
	if err != nil {
		return p.alg.Var(nil, nil), nil
	}

	body, err := p.statement()

	return p.alg.While(cond, body), err
}

func (p *Parser[E, S]) ifStatement() (S, error) {
	_, err := p.consume(token.LeftParen, "expected '(' after 'if'.")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	cond, err := p.expression()
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	_, err = p.consume(token.RightParen, "expected ')' after if condition")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	then, err := p.statement()
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	var _else S
	if p.match(token.Else) {
		_else, err = p.statement()
	}

	return p.alg.If(cond, then, _else), err
}

func (p *Parser[E, S]) block() (S, error) {
	var stmts []S
	for !p.check(token.RightBrace) && !p.isAtEnd() {
		exp, err := p.declaration()
		if err != nil {
			return p.alg.Var(nil, nil), nil
		}

		stmts = append(stmts, exp)
	}

	_, err := p.consume(token.RightBrace, "expected '}' after block.")

	return p.alg.Block(stmts), err
}

func (p *Parser[E, S]) expressionStatement() (S, error) {
	expr, err := p.expression()
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	_, err = p.consume(token.Semicolon, "Expected ';' after expression")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	return p.alg.ExprStatement(expr), nil
}

func (p *Parser[E, S]) printStatement() (S, error) {
	value, err := p.expression()
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	_, err = p.consume(token.Semicolon, "expected ';' after value.")
	if err != nil {
		return p.alg.Var(nil, nil), err
	}

	return p.alg.Print(value), nil
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

	return p.primary()
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
		case token.Class, token.Fun, token.Var, token.For, token.If, token.While, token.Print, token.Return:
			return
		}

		p.advance()
	}
}

func (p *Parser[E, S]) consume(kind token.Kind, message string) (*token.Token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}

	return nil, errors.New(message)
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

func (p *Parser[E, S]) advance() *token.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser[E, S]) isAtEnd() bool {
	return p.peek().Kind == token.Eof
}

func (p *Parser[E, S]) peek() *token.Token {
	return p.tokens[p.current]
}

func (p *Parser[E, S]) previous() *token.Token {
	return p.tokens[p.current-1]
}
