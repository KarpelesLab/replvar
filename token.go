package replvar

type Token int

const (
	TokenInvalid Token = iota
	TokenVariable
	TokenNumber
	TokenStringConstant
	TokenVariableEnd // }}

	// operators
	TokenDot       // .
	TokenAdd       // +
	TokenSubstract // -
	TokenMultiply  // *
	TokenDivide    // /
	TokenEqual     // ==
	TokenDifferent // !=
	TokenNot       // !
	TokenOr        // |
	TokenLogicOr   // ||
	TokenAnd       // &
	TokenLogicAnd  // &&
	TokenXor       // ^
)

func (p *parser) readToken() (Token, []rune) {
	for {
		switch p.cur() {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return TokenNumber, p.readNumberToken()
		case '"', '\'', '`':
			return TokenStringConstant, []rune{p.take()}
		case '.':
			p.forward()
			return TokenDot, nil
		case '+':
			p.forward()
			return TokenAdd, nil
		case '-':
			p.forward()
			return TokenSubstract, nil
		case '*':
			p.forward()
			return TokenMultiply, nil
		case '/':
			p.forward()
			return TokenDivide, nil
		case '^':
			p.forward()
			return TokenXor, nil
		case '=':
			if p.next() == '=' {
				p.forward2()
				return TokenEqual, nil
			}
			return TokenInvalid, []rune{p.cur()}
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
				return TokenLogicOr, nil
			}
			p.forward()
			return TokenOr, nil
		case '&':
			if p.next() == '&' {
				p.forward2()
				return TokenLogicAnd, nil
			}
			p.forward()
			return TokenAnd, nil
		case '}':
			if p.next() == '}' {
				return TokenVariableEnd, []rune{p.take(), p.take()}
			}
			return TokenInvalid, []rune{p.cur()}
		case ' ', '\t', '\r', '\n':
			// skip spaces
			p.forward()
		default:
			c := p.cur()
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
				return TokenVariable, p.readVariableToken()
			}
			return TokenInvalid, []rune{p.cur()}
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
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			res = append(res, c)
			p.forward()
		} else {
			return res
		}
	}
}

func (t Token) MathOp() string {
	switch t {
	case TokenAdd:
		return "+"
	case TokenSubstract:
		return "-"
	case TokenMultiply:
		return "*"
	case TokenDivide:
		return "/"
	case TokenOr:
		return "|"
	case TokenAnd:
		return "&"
	case TokenLogicOr:
		return "||"
	case TokenLogicAnd:
		return "&&"
	case TokenEqual:
		return "=="
	case TokenDifferent:
		return "!="
	default:
		return ""
	}
}
