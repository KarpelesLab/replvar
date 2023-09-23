package replvar

import (
	"bytes"
	"context"
	"strings"

	"github.com/KarpelesLab/pjson"
)

// Replace will replace any variable found in s with their value from the context
func Replace(ctx context.Context, s string, mode string) (string, error) {
	// attempt to locate {{ in s, if found locate }} and change string. If nothing is found, return empty string, false
	//r := &strings.Builder{}
	r := &bytes.Buffer{}

	prevC := rune(0)
	var vdata *strings.Builder
	changed := false

	for _, C := range s {
		if vdata != nil {
			// currently reading a variable
			switch C {
			case ' ', '\n', '\r', '\t':
				// do nothing
			case '}':
				if prevC == '}' {
					// we need to remove the last char (})
					varName := vdata.String()
					varName = varName[:len(varName)-1]
					vdata = nil
					changed = true

					switch mode {
					case "script":
						v := Resolve(ctx, varName)
						buf, err := pjson.Marshal(v)
						if err != nil {
							return s, err
							break
						}
						r.Write(buf)
					default:
						r.WriteString(ResolveString(ctx, varName))
					}

					break
				}
				fallthrough
			default:
				vdata.WriteRune(C)
				prevC = C
			}
			continue
		}
		if prevC == '{' && C == '{' {
			// start of variable
			vdata = &strings.Builder{}
			// remove previous char
			r.Truncate(r.Len() - 1)
			continue
		}
		// nothing
		r.WriteRune(C)
		prevC = C
	}

	if !changed {
		// no change → return initial s value
		return s, nil
	}

	if vdata != nil {
		r.WriteString("{{")
		r.WriteString(vdata.String())
	}

	return r.String(), nil
}
