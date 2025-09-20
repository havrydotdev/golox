package eval

import "fmt"

func isTruthy(value any) bool {
	if value == nil {
		return false
	}

	val, ok := value.(bool)
	if ok {
		return val
	}

	return true
}

func isEqual(left, right any) bool {
	if left == nil && right == nil {
		return true
	}

	if left == nil {
		return false
	}

	return left == right
}

func checkNums(left, right any) (float32, float32, error) {
	l, okl := left.(float32)
	r, okr := right.(float32)
	if !okl {
		return 0, 0, fmt.Errorf("Expected number, got %v", left)
	}

	if !okr {
		return 0, 0, fmt.Errorf("Expected number, got %v", right)
	}

	return l, r, nil
}
