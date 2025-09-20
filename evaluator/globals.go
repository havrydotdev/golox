package eval

import (
	"fmt"
	"time"

	env "github.com/havrydotdev/golox/environment"
)

func newClock() Callable {
	return NewNativeFun(0, func(e *Evaluator, args []any) (any, error) {
		return time.Now().Unix() / 1000, nil
	})
}

func newPrint() Callable {
	return NewNativeFun(1, func(e *Evaluator, args []any) (any, error) {
		arg := args[0]
		switch arg.(type) {
		case float32:
			fmt.Printf("%f\n", arg)
		case string:
			fmt.Printf("%s\n", arg)
		default:
			fmt.Printf("%v\n", arg)
		}

		return nil, nil
	})
}

func newGlobals() *env.Env {
	global := env.New()
	global.Define("clock", newClock())
	global.Define("print", newPrint())

	return global
}
