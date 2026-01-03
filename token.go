package replvar

// Token represents a lexical token type identified during parsing.
// Tokens are the basic building blocks used to construct the AST.
type Token int

// Token type constants representing all recognized lexical elements.
const (
	TokenInvalid        Token = iota // Invalid or unrecognized token
	TokenVariable                    // Identifier/variable name (e.g., "foo", "myVar")
	TokenNumber                      // Numeric literal (integer or float)
	TokenStringConstant              // String literal delimiter (", ', or `)
	TokenVariableEnd                 // End of variable expression: }}

	// Operators
	TokenDot          // Member access: .
	TokenAdd          // Addition: +
	TokenSubtract     // Subtraction: -
	TokenMultiply     // Multiplication: *
	TokenDivide       // Division: /
	TokenModulo       // Modulo: %
	TokenEqual        // Equality comparison: ==
	TokenDifferent    // Inequality comparison: !=
	TokenLess         // Less than: <
	TokenLessEqual    // Less than or equal: <=
	TokenGreater      // Greater than: >
	TokenGreaterEqual // Greater than or equal: >=
	TokenNot          // Logical NOT: !
	TokenBitwiseNot   // Bitwise NOT: ~
	TokenOr           // Bitwise OR: |
	TokenLogicOr      // Logical OR: ||
	TokenAnd          // Bitwise AND: &
	TokenLogicAnd     // Logical AND: &&
	TokenXor          // Bitwise XOR: ^
	TokenShiftLeft    // Left shift: <<
	TokenShiftRight   // Right shift: >>
)

// operatorPrecedence defines the precedence of operators.
// Lower values bind tighter (higher precedence).
// Based on https://en.wikipedia.org/wiki/Order_of_operations
var operatorPrecedence = map[Token]int{
	TokenNot:          2,
	TokenBitwiseNot:   2,
	TokenMultiply:     3,
	TokenDivide:       3,
	TokenModulo:       3,
	TokenAdd:          4,
	TokenSubtract:     4,
	TokenShiftLeft:    5,
	TokenShiftRight:   5,
	TokenLess:         6,
	TokenLessEqual:    6,
	TokenGreater:      6,
	TokenGreaterEqual: 6,
	TokenEqual:        7,
	TokenDifferent:    7,
	TokenAnd:          8,
	TokenXor:          9,
	TokenOr:           10,
	TokenLogicAnd:     11,
	TokenLogicOr:      12,
}

// readToken reads the next token from the parser buffer.
// It returns the token type and any associated data (e.g., the characters
// that make up a number or variable name). Whitespace is automatically skipped.
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
			return TokenSubtract, nil
		case '*':
			p.forward()
			return TokenMultiply, nil
		case '/':
			p.forward()
			return TokenDivide, nil
		case '%':
			p.forward()
			return TokenModulo, nil
		case '^':
			p.forward()
			return TokenXor, nil
		case '~':
			p.forward()
			return TokenBitwiseNot, nil
		case '<':
			if p.next() == '<' {
				p.forward2()
				return TokenShiftLeft, nil
			}
			if p.next() == '=' {
				p.forward2()
				return TokenLessEqual, nil
			}
			p.forward()
			return TokenLess, nil
		case '>':
			if p.next() == '>' {
				p.forward2()
				return TokenShiftRight, nil
			}
			if p.next() == '=' {
				p.forward2()
				return TokenGreaterEqual, nil
			}
			p.forward()
			return TokenGreater, nil
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

// readNumberToken reads a numeric literal (integer or floating-point).
// It supports decimal numbers with an optional single decimal point.
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

// readVariableToken reads a variable/identifier name.
// Valid characters are alphanumeric (a-z, A-Z, 0-9) and underscore.
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

// MathOp returns the string representation of a math/logic operator token.
// Returns an empty string if the token is not a recognized operator.
func (t Token) MathOp() string {
	switch t {
	case TokenAdd:
		return "+"
	case TokenSubtract:
		return "-"
	case TokenMultiply:
		return "*"
	case TokenDivide:
		return "/"
	case TokenModulo:
		return "%"
	case TokenOr:
		return "|"
	case TokenAnd:
		return "&"
	case TokenXor:
		return "^"
	case TokenLogicOr:
		return "||"
	case TokenLogicAnd:
		return "&&"
	case TokenEqual:
		return "=="
	case TokenDifferent:
		return "!="
	case TokenLess:
		return "<"
	case TokenLessEqual:
		return "<="
	case TokenGreater:
		return ">"
	case TokenGreaterEqual:
		return ">="
	case TokenShiftLeft:
		return "<<"
	case TokenShiftRight:
		return ">>"
	default:
		return ""
	}
}

// Precedence returns the operator precedence for this token.
// Lower values bind tighter (higher precedence).
// Returns 0 for non-operator tokens.
func (t Token) Precedence() int {
	if p, ok := operatorPrecedence[t]; ok {
		return p
	}
	return 0
}

// IsUnary returns true if this token is a unary operator.
func (t Token) IsUnary() bool {
	return t == TokenNot || t == TokenBitwiseNot
}
