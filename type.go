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

// varBitwiseNot performs bitwise negation on its sub-expression.
// Implements the ~ operator.
type varBitwiseNot struct {
	sub Var
}

func (n *varBitwiseNot) Resolve(ctx context.Context) (any, error) {
	sub, err := n.sub.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	// Convert to int64 and negate
	if num, ok := typutil.AsNumber(sub); ok {
		switch v := num.(type) {
		case int64:
			return ^v, nil
		case uint64:
			return ^v, nil
		case float64:
			return ^int64(v), nil
		}
	}
	return nil, fmt.Errorf("bitwise NOT requires numeric operand, got %T", sub)
}

func (n *varBitwiseNot) IsStatic() bool {
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
// Supports: +, -, *, /, %, |, &, ^, ||, &&, ==, !=, <, <=, >, >=, <<, >>
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
		return typutil.AsBool(a) && typutil.AsBool(b), nil
	case "||":
		return typutil.AsBool(a) || typutil.AsBool(b), nil
	case "==":
		return typutil.Equal(a, b), nil
	case "!=":
		return !typutil.Equal(a, b), nil
	case "<", "<=", ">", ">=":
		return m.resolveComparison(a, b)
	case "<<", ">>":
		return m.resolveShift(a, b)
	default:
		res, _ := typutil.Math(m.op, a, b)
		return res, nil
	}
}

// resolveComparison handles <, <=, >, >= operators.
func (m *varMath) resolveComparison(a, b any) (any, error) {
	// Try numeric comparison first
	numA, okA := typutil.AsNumber(a)
	numB, okB := typutil.AsNumber(b)
	if okA && okB {
		var cmp int
		switch va := numA.(type) {
		case int64:
			switch vb := numB.(type) {
			case int64:
				cmp = compareInt64(va, vb)
			case float64:
				cmp = compareFloat64(float64(va), vb)
			}
		case float64:
			switch vb := numB.(type) {
			case int64:
				cmp = compareFloat64(va, float64(vb))
			case float64:
				cmp = compareFloat64(va, vb)
			}
		}
		switch m.op {
		case "<":
			return cmp < 0, nil
		case "<=":
			return cmp <= 0, nil
		case ">":
			return cmp > 0, nil
		case ">=":
			return cmp >= 0, nil
		}
	}
	// Fall back to string comparison
	strA, _ := typutil.AsString(a)
	strB, _ := typutil.AsString(b)
	switch m.op {
	case "<":
		return strA < strB, nil
	case "<=":
		return strA <= strB, nil
	case ">":
		return strA > strB, nil
	case ">=":
		return strA >= strB, nil
	}
	return false, nil
}

// resolveShift handles << and >> operators.
func (m *varMath) resolveShift(a, b any) (any, error) {
	numA, okA := typutil.AsNumber(a)
	numB, okB := typutil.AsNumber(b)
	if !okA || !okB {
		return nil, fmt.Errorf("shift operators require numeric operands")
	}
	var va int64
	switch v := numA.(type) {
	case int64:
		va = v
	case float64:
		va = int64(v)
	}
	var vb uint64
	switch v := numB.(type) {
	case int64:
		vb = uint64(v)
	case uint64:
		vb = v
	case float64:
		vb = uint64(v)
	}
	switch m.op {
	case "<<":
		return va << vb, nil
	case ">>":
		return va >> vb, nil
	}
	return nil, nil
}

func compareInt64(a, b int64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareFloat64(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
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
