package replvar

import (
	"errors"
	"fmt"
	"io"
	"unicode"
)

type parser struct {
	buf []rune
}

var escapedChars = map[rune]rune{
	'r':  '\r',
	'n':  '\n',
	't':  '\t',
	'v':  '\v',
	'\\': '\\',
}

// ParseString parses a constant string
func ParseString(s string) (Var, error) {
	p := newParser(s)
	return p.parseString(-1)
}

// ParseVariable parses a variable string, such as what is typically found inside {{}}
func ParseVariable(s string) (Var, error) {
	p := newParser(s)
	return p.parse(false)
}

func newParser(s string) *parser {
	p := &parser{
		buf: []rune(s),
	}
	return p
}

// parse will parse content of a variable. if varStart is false, TokenVariableEnd will raise an error
// instead of returning
func (p *parser) parse(varStart bool) (Var, error) {
	var res Var

	for {
		tok, dat := p.readToken()

		switch tok {
		case TokenVariableEnd:
			if !varStart {
				return nil, fmt.Errorf("unexpected token }}")
			}
			return res, nil
		case TokenStringConstant:
			if res != nil {
				// TODO
				return nil, errors.New("res not nil")
			}
			sub, err := p.parseString(dat[0])
			if err != nil {
				return nil, err
			}
			res = sub
		case TokenVariable:
			if res != nil {
				// TODO
				return nil, errors.New("res not nil")
			}
			res = varFetchFromCtx(string(dat))
		default:
			return nil, fmt.Errorf("unexpected token %v cur=%c", tok, p.cur())
		}
	}
}

func (p *parser) parseString(cut rune) (Var, error) {
	var str []rune
	var res []Var

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

func (p *parser) cur() rune {
	if len(p.buf) == 0 {
		return -1
	}
	return p.buf[0]
}

func (p *parser) forward() {
	if len(p.buf) > 0 {
		p.buf = p.buf[1:]
	}
}

func (p *parser) forward2() {
	if len(p.buf) > 1 {
		p.buf = p.buf[2:]
	} else {
		p.buf = nil
	}
}

func (p *parser) take() rune {
	if len(p.buf) == 0 {
		return -1
	}
	r := p.buf[0]
	p.buf = p.buf[1:]
	return r
}

func (p *parser) next() rune {
	if len(p.buf) < 2 {
		return -1
	}
	return p.buf[1]
}

func (p *parser) skipSpaces() {
	for unicode.IsSpace(p.cur()) {
		p.forward()
	}
}
