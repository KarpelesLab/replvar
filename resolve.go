package replvar

import (
	"context"

	"github.com/KarpelesLab/typutil"
)

// ResolveString returns the string value for a variable name based on the context
//
// Deprecated: use ParseVariable and resolve instead
func ResolveString(ctx context.Context, v string) string {
	res := Resolve(ctx, v)
	strres, _ := typutil.AsString(res)
	return strres
}

// Resolve returns the value for a variable name based on the context
//
// Deprecated: use ParseVariable and resolve instead
func Resolve(ctx context.Context, v string) any {
	obj, err := ParseVariable(v)
	if err != nil {
		return nil
	}
	res, _ := obj.Resolve(ctx)
	return res
}
