package token

import "fmt"

type Token struct {
	Kind    Kind
	Lexeme  string
	Literal any
	Line    int
}

func New(kind Kind, lexeme string, literal any, line int) *Token {
	return &Token{kind, lexeme, literal, line}
}

func (t *Token) ToString() string {
	return fmt.Sprintf("{Kind(%v), Literal(%v), Lexeme(%s)}", t.Kind, t.Literal, t.Lexeme)
}
