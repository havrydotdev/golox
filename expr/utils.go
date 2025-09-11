package expr

import "strings"

func parenthesize(name string, exprs ...Printer) string {
	b := strings.Builder{}

	b.WriteByte('(')
	b.WriteString(name)
	for _, expr := range exprs {
		b.WriteByte(' ')
		b.WriteString(expr.Print())
	}

	b.WriteByte(')')

	return b.String()
}
