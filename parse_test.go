package replvar_test

import (
	"context"
	"testing"

	"github.com/KarpelesLab/replvar"
)

type testVector struct {
	in  string
	out string
}

func TestParser(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "var", "world")
	ctx = context.WithValue(ctx, "var2", map[string]any{"foo": "bar", "num": 40})

	testV := []*testVector{
		// Basic variable substitution
		&testVector{"hello {{var}}", "hello world"},
		&testVector{"hello {{  var   \r\n}}", "hello world"},
		&testVector{"hello {{'world'}}", "hello world"},
		&testVector{"hello {{`world 2\\t`}}", "hello world 2\\t"},
		&testVector{"hello {{\"world \\t\"}}", "hello world \t"},
		// Field access
		&testVector{"hello {{var2.foo}}", "hello bar"},
		&testVector{"hello {{  var2  .   foo   }}", "hello bar"},
		// Arithmetic
		&testVector{"hello {{var2.num + 2}}", "hello 42"},
		// Equality
		&testVector{"hello {{var2.foo == 'bar'}}", "hello 1"},
		&testVector{"hello {{var2.foo == 0}}", "hello 0"},
		&testVector{"hello {{var2.num == '40.0'}}", "hello 1"},
		&testVector{"hello {{var2.num != '40.1'}}", "hello 0"},
		// Operator precedence: * binds tighter than +
		&testVector{"{{2 + 3 * 4}}", "14"},
		&testVector{"{{10 - 2 * 3}}", "4"},
		&testVector{"{{2 * 3 + 4 * 5}}", "26"},
		// Comparison operators
		&testVector{"{{5 > 3}}", "1"},
		&testVector{"{{5 < 3}}", "0"},
		&testVector{"{{5 >= 5}}", "1"},
		&testVector{"{{5 <= 4}}", "0"},
		// Modulo
		&testVector{"{{10 % 3}}", "1"},
		&testVector{"{{15 % 4}}", "3"},
		// Shift operators
		&testVector{"{{1 << 4}}", "16"},
		&testVector{"{{16 >> 2}}", "4"},
		// Logical operators with precedence
		&testVector{"{{1 || 0 && 0}}", "1"},  // && binds tighter than ||
		&testVector{"{{0 || 1 && 1}}", "1"},
		// Bitwise operators
		&testVector{"{{5 | 3}}", "7"},
		&testVector{"{{5 & 3}}", "1"},
		&testVector{"{{5 ^ 3}}", "6"},
		// filter tests
		&testVector{"hello {{var|upper}}", "hello WORLD"},
		&testVector{"hello {{var|upper|lower}}", "hello world"},
		&testVector{"hello {{var2.foo|upper}}", "hello BAR"},
		&testVector{"{{var2|json}}", `{"foo":"bar","num":40}`},
	}

	for _, vect := range testV {
		res, err := replvar.Replace(ctx, vect.in, "text")
		if err != nil {
			t.Errorf("failed to run %s: %s", vect.in, err)
			continue
		}
		if res != vect.out {
			t.Errorf("invalid result for %s: got %s but expected %s", vect.in, res, vect.out)
		}
	}
}
