package replvar

import (
	"fmt"
	"io"
	"unicode"

	"github.com/KarpelesLab/typutil"
)

// parser holds the state for parsing variable expressions and strings.
// It operates on a buffer of runes and provides methods for tokenization
// and AST construction.
type parser struct {
	buf []rune // input buffer of runes to be parsed
}

// escapedChars maps escape sequence characters to their actual values.
// These are recognized within double-quoted strings (e.g., "\n" becomes newline).
var escapedChars = map[rune]rune{
	'r':  '\r',
	'n':  '\n',
	't':  '\t',
	'v':  '\v',
	'\\': '\\',
}

// ParseString parses a string that may contain embedded variable expressions.
// Variable expressions are delimited by {{ and }}. The mode parameter controls
// how nested variables are handled:
//   - "text": variables are resolved to their string representation
//   - "json": variables are automatically JSON-encoded when embedded
func ParseString(s string, mode string) (Var, error) {
	p := newParser(s)
	return p.parseString(-1, mode)
}

// ParseVariable parses a variable expression (the content typically found inside {{}}).
// This handles variable names, operators, and nested expressions directly.
func ParseVariable(s string) (Var, error) {
	p := newParser(s)
	return p.parse(false)
}

// newParser creates a new parser initialized with the given string.
func newParser(s string) *parser {
	p := &parser{
		buf: []rune(s),
	}
	return p
}

// parse parses a variable expression using a two-stage approach:
//
// Stage 1: Tokenization - reads tokens and converts them to Var objects.
// Operators are stored as varPendingToken placeholders.
//
// Stage 2: Operator association - processes pending tokens to build the
// final AST by associating operators with their operands.
//
// If varStart is true, parsing expects to end with }} (TokenVariableEnd).
// If varStart is false, }} will raise an error.
func (p *parser) parse(varStart bool) (Var, error) {
	var res []Var

	// Stage 1: Tokenization loop
mainloop:
	for {
		if p.empty() {
			if varStart {
				// unexpected
				return nil, io.ErrUnexpectedEOF
			}
			// reached end of buffer
			break
		}
		tok, dat := p.readToken()
		switch tok {
		case TokenVariableEnd:
			if !varStart {
				return nil, fmt.Errorf("unexpected token }}")
			}
			break mainloop
		case TokenStringConstant:
			sub, err := p.parseString(dat[0], "text")
			if err != nil {
				return nil, err
			}
			res = append(res, sub)
		case TokenNumber:
			v, ok := typutil.AsNumber(string(dat))
			if !ok {
				return nil, fmt.Errorf("invalid number: %s", string(dat))
			}
			res = append(res, &staticVar{v})
		case TokenVariable:
			res = append(res, varFetchFromCtx(string(dat)))
		case TokenInvalid:
			return nil, fmt.Errorf("invalid token found, value=%v", dat)
		default:
			// unknown token, defer to step 2
			res = append(res, varPendingToken(tok))
		}
	}

	if len(res) == 0 {
		return varNull{}, nil
	}

	// Stage 2: Operator association
	// Process pending tokens (operators) and build the final AST.
	// This handles unary operators (!) and binary operators (+, -, *, /, etc.)
	for {
		if len(res) == 1 {
			return res[0], nil
		}

		if tok, ok := res[0].(varPendingToken); ok {
			// only ! (TokenNot) or ^ (binary not) operators can be here
			switch Token(tok) {
			case TokenNot:
				not := &varNot{res[1]}
				res = append([]Var{not}, res[2:]...)
			default:
				return nil, fmt.Errorf("step 2: unexpected token %v", tok)
			}
			continue
		}

		if tok, ok := res[1].(varPendingToken); ok {
			if len(res) < 2 {
				return nil, fmt.Errorf("invalid syntax: expected something after token %v", tok)
			}
			switch Token(tok) {
			case TokenDot:
				// access a sub element of array, we expect res[2] to be a varFetchFromCtx
				if v2, ok := res[2].(varFetchFromCtx); ok {
					access := &varAccessOffset{res[0], string(v2)}
					res = append([]Var{access}, res[3:]...)
				} else {
					return nil, fmt.Errorf("invalid syntax: dot not followed by var")
				}
			default:
				if math := Token(tok).MathOp(); math != "" {
					res = append([]Var{&varMath{res[0], res[2], math}}, res[3:]...)
					break
				}
				return nil, fmt.Errorf("step 2: unexpected token %v", tok)
			}
			continue
		}

		return nil, fmt.Errorf("invalid syntax: expected token in 1st or 2nd position of res")
	}
}

// parseString parses a string literal or template string.
//
// The cut parameter specifies the closing delimiter:
//   - -1: parse until end of input (for top-level strings)
//   - '"', '\'', '`': parse until matching quote (for quoted strings)
//
// The mode parameter controls variable handling ("text" or "json").
// Supports escape sequences (in double-quoted strings) and nested {{}} expressions.
func (p *parser) parseString(cut rune, mode string) (Var, error) {
	var str []rune // accumulator for literal characters
	var res []Var  // result Var objects (static strings and variables)

mainloop:
	for {
		c := p.take()
		if c == cut {
			// reached the end of the string
			break
		}
		if c == -1 {
			// unexpected end of string
			return nil, io.ErrUnexpectedEOF
		}

		switch c {
		case '\\':
			// escape char
			nc := p.cur()
			if nc == cut && cut != -1 {
				str = append(str, nc)
				p.forward()
				continue mainloop
			}
			if cut == '"' {
				if n, ok := escapedChars[nc]; ok {
					str = append(str, n)
					p.forward()
					continue mainloop
				}
			}
			if nc == '\\' {
				str = append(str, nc)
				p.forward()
				continue mainloop
			}
			// not a matching thing, just include the \ character to the output
		case '{':
			if p.cur() == '{' {
				// we have a string, flush it
				if len(str) > 0 {
					res = append(res, &staticVar{string(str)})
					str = nil
				}
				p.forward()
				// parse subvar
				sub, err := p.parse(true)
				if err != nil {
					return nil, err
				}
				if mode == "json" {
					// if json mode, encode any subvar as json
					sub = &varJsonMarshal{sub}
				}
				res = append(res, sub)
				continue mainloop
			}
		}

		// nothing happened
		str = append(str, c)
	}

	if len(str) > 0 {
		res = append(res, &staticVar{string(str)})
		str = nil
	}
	if len(res) == 1 {
		return res[0], nil
	}

	return varConcat(res), nil
}

// cur returns the current rune without consuming it, or -1 if at end.
func (p *parser) cur() rune {
	if len(p.buf) == 0 {
		return -1
	}
	return p.buf[0]
}

// empty returns true if the parser buffer is exhausted.
func (p *parser) empty() bool {
	return len(p.buf) == 0
}

// forward advances the parser by one rune.
func (p *parser) forward() {
	if len(p.buf) > 0 {
		p.buf = p.buf[1:]
	}
}

// forward2 advances the parser by two runes (used for two-character tokens like == or &&).
func (p *parser) forward2() {
	if len(p.buf) > 1 {
		p.buf = p.buf[2:]
	} else {
		p.buf = nil
	}
}

// take returns the current rune and advances the parser, or -1 if at end.
func (p *parser) take() rune {
	if len(p.buf) == 0 {
		return -1
	}
	r := p.buf[0]
	p.buf = p.buf[1:]
	return r
}

// next returns the rune after the current one (lookahead), or -1 if unavailable.
func (p *parser) next() rune {
	if len(p.buf) < 2 {
		return -1
	}
	return p.buf[1]
}

// skipSpaces advances past any whitespace characters.
func (p *parser) skipSpaces() {
	for unicode.IsSpace(p.cur()) {
		p.forward()
	}
}
