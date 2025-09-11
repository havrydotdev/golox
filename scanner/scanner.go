package scanner

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/havrydotdev/golox/token"
)

type Scanner struct {
	source string
	tokens []*token.Token

	start   int
	current int
	line    int
}

func New(source string) *Scanner {
	return &Scanner{source: source, tokens: make([]*token.Token, 0), start: 0, current: 0, line: 1}
}

func (s *Scanner) Scan() ([]*token.Token, error) {
	for !s.isAtEnd() {
		s.start = s.current

		if err := s.scanToken(); err != nil {
			return nil, err
		}
	}

	eof := token.New(token.Eof, "", nil, s.line)
	s.tokens = append(s.tokens, eof)

	return s.tokens, nil
}

func (s *Scanner) scanToken() error {
	c := s.advance()

	switch c {
	// one-character tokens
	case '(':
		s.addToken(token.LeftParen)
	case ')':
		s.addToken(token.RightParen)
	case '{':
		s.addToken(token.LeftBrace)
	case '}':
		s.addToken(token.RightBrace)
	case ',':
		s.addToken(token.Comma)
	case '.':
		s.addToken(token.Dot)
	case '-':
		s.addToken(token.Minus)
	case '+':
		s.addToken(token.Plus)
	case ';':
		s.addToken(token.Semicolon)
	case '*':
		s.addToken(token.Star)

	// two or one character tokens
	case '!':
		kind := token.Bang
		if s.match('=') {
			kind = token.BangEqual
		}

		s.addToken(kind)
	case '=':
		kind := token.Equal
		if s.match('=') {
			kind = token.EqualEqual
		}

		s.addToken(kind)
	case '<':
		kind := token.Less
		if s.match('=') {
			kind = token.LessEqual
		}

		s.addToken(kind)
	case '>':
		kind := token.Greater
		if s.match('=') {
			kind = token.GreaterEqual
		}

		s.addToken(kind)

	// multiple character tokens
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else if s.match('*') {
			// multi-line comment
			for !s.isAtEnd() {
				if s.peek() == '\n' {
					s.line++
				}

				if s.peek() == '*' && s.peekNext() == '/' {
					// eat comment
					s.advance()
					s.advance()
					break
				}

				s.advance()
			}
		} else {
			s.addToken(token.Slash)
		}

	// special characters
	case ' ', '\t', '\r':
		break

	case '\n':
		s.line++

	// literals
	case '"':
		if err := s.string(); err != nil {
			return err
		}

	// keywords

	default:
		if isDigit(c) {
			s.number()
			break
		} else if isAlpha(c) {
			s.identifier()
		} else {
			return fmt.Errorf("Unknown token %s at %d, %d", string(c), s.current, s.line)
		}
	}

	return nil
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func (s *Scanner) identifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]
	kind, ok := keywords[text]
	if !ok {
		kind = token.Identifier
	}

	s.addToken(kind)
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	num, _ := strconv.ParseFloat(s.source[s.start:s.current], 32)

	s.addToken(token.Number, float32(num))
}

func (s *Scanner) string() error {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}

		s.advance()
	}

	if s.isAtEnd() {
		return errors.New("Unterminated string.")
	}

	s.advance()
	s.addToken(token.String, s.source[s.start+1:s.current-1])

	return nil
}

func (s *Scanner) addToken(kind token.Kind, literal ...any) {
	var l any
	if len(literal) != 0 {
		l = literal[0]
	}

	lexeme := s.source[s.start:s.current]

	s.tokens = append(s.tokens, token.New(kind, lexeme, l, s.line))
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() || s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) advance() byte {
	curr := s.current
	s.current++
	return s.source[curr]
}

func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return '\000'
	}

	return s.source[s.current]
}

func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return '\000'
	}

	return s.source[s.current+1]
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}
