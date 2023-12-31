package replvar

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/KarpelesLab/pjson"
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

type varPendingToken Token

func (v varPendingToken) Resolve(ctx context.Context) (any, error) {
	return nil, errors.New("this value should never happen (pending token)")
}

func (v varPendingToken) IsStatic() bool {
	return true
}

type varNull struct{}

func (varNull) Resolve(ctx context.Context) (any, error) {
	return nil, nil
}

func (varNull) IsStatic() bool {
	return true
}

type varNot struct {
	sub Var
}

func (n *varNot) Resolve(ctx context.Context) (any, error) {
	sub, err := n.sub.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	return !typutil.AsBool(sub), nil
}

func (n *varNot) IsStatic() bool {
	return n.sub.IsStatic()
}

type varAccessOffset struct {
	sub    Var
	offset string
}

func (a *varAccessOffset) Resolve(ctx context.Context) (any, error) {
	sub, err := a.sub.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	switch elem := sub.(type) {
	case map[string]any:
		return elem[a.offset], nil
	case map[string]string:
		return elem[a.offset], nil
	default:
		// ??
		return nil, fmt.Errorf("lookup failed, offset=%s cur type=%T", a.offset, elem)
	}
}

func (a *varAccessOffset) IsStatic() bool {
	return a.sub.IsStatic()
}

type varMath struct {
	a, b Var
	op   string
}

func (m *varMath) Resolve(ctx context.Context) (any, error) {
	a, err := m.a.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	b, err := m.b.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	switch m.op {
	case "&&":
		// logic and
		return typutil.AsBool(a) && typutil.AsBool(b), nil
	case "||":
		// logic or
		return typutil.AsBool(a) || typutil.AsBool(b), nil
	case "==":
		// equal
		return typutil.Equal(a, b), nil
	case "!=":
		// not equal
		return !typutil.Equal(a, b), nil
	default:
		res, _ := typutil.Math(m.op, a, b)
		return res, nil
	}
}

func (m *varMath) IsStatic() bool {
	return m.a.IsStatic() && m.b.IsStatic()
}

type varJsonMarshal struct {
	obj Var
}

func (j *varJsonMarshal) Resolve(ctx context.Context) (any, error) {
	res, err := j.obj.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	enc, err := pjson.MarshalContext(ctx, res)
	return string(enc), err
}

func (j *varJsonMarshal) IsStatic() bool {
	return j.obj.IsStatic()
}
