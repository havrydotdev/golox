package main

import (
	_ "embed"
	"testing"

	eval "github.com/havrydotdev/golox/evaluator"
	"github.com/havrydotdev/golox/parser"
	"github.com/havrydotdev/golox/scanner"
)

//go:embed _examples/fib_recur.lox
var fibRecur []byte

func TestFibRecursion(t *testing.T) {
	tokens, err := scanner.New(string(fibRecur)).Scan()
	if err != nil {
		t.Error(err)
	}

	exprs, errs := parser.New(tokens, eval.New()).Parse()
	for _, err := range errs {
		t.Error(err)
	}

	for _, expr := range exprs {
		err := expr.Eval()
		if err != nil {
			t.Error(err)
		}
	}
}
