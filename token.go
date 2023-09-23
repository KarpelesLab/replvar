package replvar

type Token int

const (
	TokenInvalid Token = iota
	TokenVariable
	TokenNumber
	TokenStringConstant
	TokenVariableEnd // }}

	// operators
	TokenDot        // .
	TokenAdd        // +
	TokenSubstract  // -
	TokenMultiply   // *
	TokenDivide     // /
	TokenEqual      // ==
	TokenDifferent  // !=
	TokenNot        // !
	TokenOr         // |
	TokenBooleanOr  // ||
	TokenAnd        // &
	TokenBooleanAnd // &&
)

func (p *parser) readToken() (Token, []rune) {
	for {
		switch p.cur() {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return TokenNumber, p.readNumberToken()
		case '"', '\'', '`':
			return TokenStringConstant, []rune{p.take()}
		case '.':
			return TokenDot, nil
		case '+':
			return TokenAdd, nil
		case '-':
			return TokenSubstract, nil
		case '*':
			return TokenMultiply, nil
		case '/':
			return TokenDivide, nil
		case '=':
			if p.next() == '=' {
				p.forward2()
				return TokenEqual, nil
			}
			return TokenInvalid, nil
		case '!':
			if p.next() == '=' {
				p.forward2()
				return TokenDifferent, nil
			}
			p.forward()
			return TokenNot, nil
		case '|':
			if p.next() == '|' {
				p.forward2()
				return TokenBooleanOr, nil
			}
			p.forward()
			return TokenOr, nil
		case '&':
			if p.next() == '&' {
				p.forward2()
				return TokenBooleanAnd, nil
			}
			p.forward()
			return TokenAnd, nil
		case '}':
			if p.next() == '}' {
				return TokenVariableEnd, []rune{p.take(), p.take()}
			}
			return TokenInvalid, nil
		case ' ', '\t', '\r', '\n':
			// skip spaces
			p.forward()
		default:
			c := p.cur()
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
				return TokenVariable, p.readVariableToken()
			}
			return TokenInvalid, nil
		}
	}
}

func (p *parser) readNumberToken() []rune {
	var res []rune
	hasDot := false

	for {
		c := p.cur()
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			res = append(res, c)
			p.forward()
		case '.':
			if hasDot {
				return res
			}
			res = append(res, c)
			hasDot = true
			p.forward()
		default:
			return res
		}
	}
}

func (p *parser) readVariableToken() []rune {
	var res []rune

	for {
		c := p.cur()
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			res = append(res, c)
			p.forward()
		} else {
			return res
		}
	}
}
