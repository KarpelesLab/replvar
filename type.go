package replvar

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/KarpelesLab/pjson"
	"github.com/KarpelesLab/typutil"
)

// Var is the interface for all resolvable variable expressions.
// Implementations represent different node types in the expression AST,
// from simple static values to complex operations.
type Var interface {
	// Resolve evaluates this variable in the given context and returns its value.
	Resolve(context.Context) (any, error)
	// IsStatic returns true if the value is constant (does not depend on context).
	IsStatic() bool
}

// staticVar wraps a constant value that never changes.
type staticVar struct {
	v any
}

func (s *staticVar) Resolve(context.Context) (any, error) {
	return s.v, nil
}

func (s *staticVar) IsStatic() bool {
	return true
}

// varConcat concatenates multiple Var values into a single string.
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

// varFetchFromCtx retrieves a value from the context by key name.
// This is used for variable references like {{myvar}}.
type varFetchFromCtx string

func (a varFetchFromCtx) Resolve(ctx context.Context) (any, error) {
	return ctx.Value(string(a)), nil
}

func (a varFetchFromCtx) IsStatic() bool {
	return false
}

// varPendingToken is a placeholder for an operator during Stage 1 parsing.
// It should never exist in the final AST after Stage 2 processing.
type varPendingToken Token

func (v varPendingToken) Resolve(ctx context.Context) (any, error) {
	return nil, errors.New("this value should never happen (pending token)")
}

func (v varPendingToken) IsStatic() bool {
	return true
}

// varNull represents an empty or nil value.
type varNull struct{}

func (varNull) Resolve(ctx context.Context) (any, error) {
	return nil, nil
}

func (varNull) IsStatic() bool {
	return true
}

// varNot performs logical negation on its sub-expression.
// Implements the ! operator.
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

// varAccessOffset accesses a field/key of a map or object.
// Implements the . (dot) operator for member access like obj.field.
type varAccessOffset struct {
	sub    Var    // the object to access
	offset string // the field/key name
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

// varMath performs binary operations (arithmetic, logical, comparison).
// Supports: +, -, *, /, |, &, ^, ||, &&, ==, !=
type varMath struct {
	a, b Var    // left and right operands
	op   string // the operator ("+", "-", "==", etc.)
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

// varJsonMarshal wraps a variable and JSON-encodes its resolved value.
// Used when parsing in "json" mode to ensure embedded values are valid JSON.
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
