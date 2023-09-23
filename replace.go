package replvar

import (
	"bytes"
	"context"
	"strings"

	"github.com/KarpelesLab/pjson"
)

func Replace(ctx context.Context, s string, mode string) (string, error) {
	// attempt to locate {{ in s, if found locate }} and change string. If nothing is found, return empty string, false
	//r := &strings.Builder{}
	r := &bytes.Buffer{}

	prevC := rune(0)
	var vdata *strings.Builder

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

					switch mode {
					case "script":
						v := resolveVariable(ctx, varName)
						buf, err := pjson.Marshal(v)
						if err != nil {
							return "", err
							break
						}
						r.Write(buf)
					default:
						r.WriteString(resolveStringVariable(ctx, varName))
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

	if vdata != nil {
		r.WriteString("{{")
		r.WriteString(vdata.String())
	}

	return r.String(), nil
}
