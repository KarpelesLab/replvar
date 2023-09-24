package replvar

import (
	"context"

	"github.com/KarpelesLab/typutil"
)

// Replace will replace any variable found in s with their value from the context
func Replace(ctx context.Context, s string, mode string) (string, error) {
	obj, err := ParseString(s, mode)
	if err != nil {
		return "", err
	}

	res, err := obj.Resolve(ctx)
	if err != nil {
		return "", err
	}
	strres, _ := typutil.AsString(res)
	return strres, nil
}
