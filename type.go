package replvar

import (
	"bytes"
	"context"

	"github.com/KarpelesLab/typutil"
)

// Var is a resolvable variable
type Var interface {
	Resolve(context.Context) (any, error) // resolve this variable and return its value
	IsStatic() bool                       // if true, it means this var will not change no matter what
}

type staticVar struct {
	v any
}

func (s *staticVar) Resolve(context.Context) (any, error) {
	return s.v, nil
}

func (s *staticVar) IsStatic() bool {
	return true
}

type varConcat []Var

func (a varConcat) Resolve(ctx context.Context) (any, error) {
	res := &bytes.Buffer{}

	for _, sub := range a {
		v, err := sub.Resolve(ctx)
		if err != nil {
			return nil, err
		}
		str, _ := typutil.AsString(v)
		res.WriteString(str)
	}
	return res.String(), nil
}

func (a varConcat) IsStatic() bool {
	for _, sub := range a {
		if !sub.IsStatic() {
			return false
		}
	}
	return true
}

type varFetchFromCtx string

func (a varFetchFromCtx) Resolve(ctx context.Context) (any, error) {
	return ctx.Value(string(a)), nil
}

func (a varFetchFromCtx) IsStatic() bool {
	return false
}
