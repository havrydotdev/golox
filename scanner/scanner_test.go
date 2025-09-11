package scanner

import (
	"testing"
)

const (
	TestBasicInput = "123 * 123"
)

func TestBasic(t *testing.T) {
	tokens, err := New(TestBasicInput).Scan()
	if err != nil {
		t.Errorf("Scanning failed: %s\n", err.Error())
	}

	for _, token := range tokens {
		t.Log(token.ToString())
	}
}
