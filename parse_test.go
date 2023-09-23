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

	testV := []*testVector{
		&testVector{"hello {{var}}", "hello world"},
		&testVector{"hello {{  var   \r\n}}", "hello world"},
		&testVector{"hello {{'world'}}", "hello world"},
		&testVector{"hello {{`world 2\\t`}}", "hello world 2\\t"},
		&testVector{"hello {{\"world \\t\"}}", "hello world \t"},
	}

	for _, vect := range testV {
		v, err := replvar.ParseString(vect.in)
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
