package replvar

import (
	"context"
	"html"
	"net/url"
	"strings"

	"github.com/KarpelesLab/pjson"
	"github.com/KarpelesLab/typutil"
)

// FilterFunc is a function that transforms a value. It receives the resolved
// input, any additional arguments, and returns the transformed value.
type FilterFunc func(ctx context.Context, input any, args []any) (any, error)

var filters = map[string]FilterFunc{}

// RegisterFilter registers a named filter function that can be used with
// the pipe syntax (e.g. {{var|name}}).
func RegisterFilter(name string, fn FilterFunc) {
	filters[name] = fn
}

// LookupFilter returns the FilterFunc for the given name, or nil if not found.
func LookupFilter(name string) FilterFunc {
	return filters[name]
}

func init() {
	RegisterFilter("json", filterJSON)
	RegisterFilter("html", filterHTML)
	RegisterFilter("url", filterURL)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("lower", filterLower)
}

func filterJSON(ctx context.Context, input any, args []any) (any, error) {
	enc, err := pjson.MarshalContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return string(enc), nil
}

func filterHTML(_ context.Context, input any, _ []any) (any, error) {
	s, _ := typutil.AsString(input)
	return html.EscapeString(s), nil
}

func filterURL(_ context.Context, input any, _ []any) (any, error) {
	s, _ := typutil.AsString(input)
	return url.QueryEscape(s), nil
}

func filterUpper(_ context.Context, input any, _ []any) (any, error) {
	s, _ := typutil.AsString(input)
	return strings.ToUpper(s), nil
}

func filterLower(_ context.Context, input any, _ []any) (any, error) {
	s, _ := typutil.AsString(input)
	return strings.ToLower(s), nil
}
