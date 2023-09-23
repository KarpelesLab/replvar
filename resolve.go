package replvar

import (
	"context"
	"log"
	"strings"

	"github.com/KarpelesLab/typutil"
)

func resolveStringVariable(ctx context.Context, v string) string {
	res := resolveVariable(ctx, v)
	strres, _ := typutil.AsString(res)
	return strres
}

// resolveVariable returns the value for a variable name based on the current context
func resolveVariable(ctx context.Context, v string) any {
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
