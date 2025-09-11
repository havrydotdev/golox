package expr

// import (
// 	"testing"

// 	"github.com/havrydotdev/golox/token"
// )

// const (
// 	TestBasicExpected = "(* (- 123) (group 45.67))"
// )

// func BasicExpr[E any, S any](alg ExprAlg[E, S]) E {
// 	return alg.Binary(
// 		token.New(token.Star, "*", nil, 1),
// 		alg.Unary(
// 			token.New(token.Minus, "-", nil, 1),
// 			alg.Literal(123),
// 		),
// 		alg.Grouping(alg.Literal(45.67)),
// 	)
// }

// func TestBasic(t *testing.T) {
// 	output := BasicExpr(&PrintExpr{}).Print()
// 	t.Logf("TestBasic: %s", output)
// 	if output != TestBasicExpected {
// 		t.Errorf("Expected: %s, got: %s", TestBasicExpected, output)
// 	}
// }
