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
		&testVector{"hello {{var}}", "hello world"},
		&testVector{"hello {{  var   \r\n}}", "hello world"},
		&testVector{"hello {{'world'}}", "hello world"},
		&testVector{"hello {{`world 2\\t`}}", "hello world 2\\t"},
		&testVector{"hello {{\"world \\t\"}}", "hello world \t"},
		&testVector{"hello {{var2.foo}}", "hello bar"},
		&testVector{"hello {{  var2  .   foo   }}", "hello bar"},
		&testVector{"hello {{var2.num + 2}}", "hello 42"},
		&testVector{"hello {{var2.foo == 'bar'}}", "hello 1"},
		&testVector{"hello {{var2.foo == 0}}", "hello 0"},
		&testVector{"hello {{var2.num == '40.0'}}", "hello 1"},
		&testVector{"hello {{var2.num != '40.1'}}", "hello 0"},
	}

	for _, vect := range testV {
		v, err := replvar.ParseString(vect.in, "text")
		if err != nil {
			t.Errorf("failed to parse %s: %s", vect.in, err)
			continue
		}
		res, err := v.Resolve(ctx)
		if err != nil {
			t.Errorf("failed to run %s: %s", vect.in, err)
			continue
		}
		if res != vect.out {
			t.Errorf("invalid result for %s: got %s but expected %s", vect.in, res, vect.out)
		}
	}
}
