package replvar

import (
	"context"
	"log"
	"strings"

	"github.com/KarpelesLab/typutil"
)

// ResolveStringVariable returns the string value for a variable name based on the context
func ResolveStringVariable(ctx context.Context, v string) string {
	res := ResolveVariable(ctx, v)
	strres, _ := typutil.AsString(res)
	return strres
}

// ResolveVariable returns the value for a variable name based on the context
func ResolveVariable(ctx context.Context, v string) any {
	// we expect . to be the separator
	vA := strings.Split(v, ".")
	cur := ctx.Value(vA[0])
	if cur == nil {
		return nil
	}
	vA = vA[1:]
	for _, sub := range vA {
		switch elem := cur.(type) {
		case map[string]any:
			cur = elem[sub]
		case map[string]string:
			cur = elem[sub]
		default:
			// ??
			log.Printf("lookup failed, sub=%s cur type=%T", sub, cur)
			return nil
		}
	}
	return cur
}
