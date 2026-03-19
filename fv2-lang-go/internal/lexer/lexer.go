package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	maxInputSize      = 10_000_000
	maxIdentifierLen  = 10_000
	maxStringLen      = 1_000_000
	maxNumberLen      = 100
)

// Lexer tokenizes V-compatible source code
type Lexer struct {
	input     string
	pos       int
	line      int
	column    int
	keywords  map[string]TokenType
}

// New creates a new lexer with input validation
func New(input string) (*Lexer, error) {
	// 보안: 입력 크기 제한
	if len(input) > maxInputSize {
		return nil, fmt.Errorf("input too large: %d bytes (max %d)", len(input), maxInputSize)
	}

	// 보안: NULL 바이트 확인 (단일 바이트로 검사)
	for _, r := range input {
		if r == 0 {
			return nil, fmt.Errorf("input contains null bytes")
		}
	}

	keywords := map[string]TokenType{
		"fn":        TknFn,
		"let":       TknLet,
		"mut":       TknMut,
		"const":     TknConst,
		"if":        TknIf,
		"else":      TknElse,
		"for":       TknFor,
		"in":        TknIn,
		"match":     TknMatch,
		"type":      TknType,
		"struct":    TknStruct,
		"interface": TknInterface,
		"enum":      TknEnum,
		"trait":     TknTrait,
		"impl":      TknImpl,
		"return":    TknReturn,
		"module":    TknModule,
		"import":    TknImport,
		"true":      TknTrue,
		"false":     TknFalse,
		"none":      TknNone,
	}

	return &Lexer{
		input:    input,
		pos:      0,
		line:     1,
		column:   1,
		keywords: keywords,
	}, nil
}

// Tokenize returns all tokens from the input
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for {
		l.skipWhitespaceAndComments()

		if l.isAtEnd() {
			tokens = append(tokens, Token{
				Type:   TknEof,
				Text:   "",
				Line:   l.line,
				Column: l.column,
			})
			break
		}

		token, err := l.nextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (l *Lexer) nextToken() (Token, error) {
	startLine := l.line
	startColumn := l.column
	startPos := l.pos

	ch := l.current()
	var token Token

	switch ch {
	case '(':
		l.advance()
		token = Token{TknLParen, "(", startLine, startColumn}
	case ')':
		l.advance()
		token = Token{TknRParen, ")", startLine, startColumn}
	case '{':
		l.advance()
		token = Token{TknLBrace, "{", startLine, startColumn}
	case '}':
		l.advance()
		token = Token{TknRBrace, "}", startLine, startColumn}
	case '[':
		l.advance()
		token = Token{TknLBracket, "[", startLine, startColumn}
	case ']':
		l.advance()
		token = Token{TknRBracket, "]", startLine, startColumn}
	case ',':
		l.advance()
		token = Token{TknComma, ",", startLine, startColumn}
	case ';':
		l.advance()
		token = Token{TknSemicolon, ";", startLine, startColumn}
	case '@':
		l.advance()
		token = Token{TknAt, "@", startLine, startColumn}
	case '~':
		l.advance()
		token = Token{TknTilde, "~", startLine, startColumn}
	case '+':
		l.advance()
		if l.current() == '=' {
			l.advance()
			token = Token{TknPlusAssign, "+=", startLine, startColumn}
		} else {
			token = Token{TknPlus, "+", startLine, startColumn}
		}
	case '-':
		l.advance()
		switch l.current() {
		case '=':
			l.advance()
			token = Token{TknMinusAssign, "-=", startLine, startColumn}
		case '>':
			l.advance()
			token = Token{TknArrow, "->", startLine, startColumn}
		default:
			token = Token{TknMinus, "-", startLine, startColumn}
		}
	case '*':
		l.advance()
		if l.current() == '=' {
			l.advance()
			token = Token{TknStarAssign, "*=", startLine, startColumn}
		} else {
			token = Token{TknStar, "*", startLine, startColumn}
		}
	case '/':
		l.advance()
		if l.current() == '=' {
			l.advance()
			token = Token{TknSlashAssign, "/=", startLine, startColumn}
		} else {
			token = Token{TknSlash, "/", startLine, startColumn}
		}
	case '%':
		l.advance()
		token = Token{TknPercent, "%", startLine, startColumn}
	case '^':
		l.advance()
		token = Token{TknCaret, "^", startLine, startColumn}
	case '&':
		l.advance()
		if l.current() == '&' {
			l.advance()
			token = Token{TknLogicalAnd, "&&", startLine, startColumn}
		} else {
			token = Token{TknAmpersand, "&", startLine, startColumn}
		}
	case '|':
		l.advance()
		if l.current() == '|' {
			l.advance()
			token = Token{TknLogicalOr, "||", startLine, startColumn}
		} else {
			token = Token{TknPipe, "|", startLine, startColumn}
		}
	case '!':
		l.advance()
		if l.current() == '=' {
			l.advance()
			token = Token{TknNe, "!=", startLine, startColumn}
		} else {
			token = Token{TknNot, "!", startLine, startColumn}
		}
	case '?':
		l.advance()
		token = Token{TknQuestion, "?", startLine, startColumn}
	case '=':
		l.advance()
		switch l.current() {
		case '=':
			l.advance()
			token = Token{TknEq, "==", startLine, startColumn}
		case '>':
			l.advance()
			token = Token{TknFatArrow, "=>", startLine, startColumn}
		default:
			token = Token{TknAssign, "=", startLine, startColumn}
		}
	case '<':
		l.advance()
		switch l.current() {
		case '=':
			l.advance()
			token = Token{TknLe, "<=", startLine, startColumn}
		case '<':
			l.advance()
			token = Token{TknLeftShift, "<<", startLine, startColumn}
		default:
			token = Token{TknLt, "<", startLine, startColumn}
		}
	case '>':
		l.advance()
		switch l.current() {
		case '=':
			l.advance()
			token = Token{TknGe, ">=", startLine, startColumn}
		case '>':
			l.advance()
			token = Token{TknRightShift, ">>", startLine, startColumn}
		default:
			token = Token{TknGt, ">", startLine, startColumn}
		}
	case ':':
		l.advance()
		switch l.current() {
		case ':':
			l.advance()
			token = Token{TknDoubleColon, "::", startLine, startColumn}
		case '=':
			l.advance()
			token = Token{TknColonAssign, ":=", startLine, startColumn}
		default:
			token = Token{TknColon, ":", startLine, startColumn}
		}
	case '.':
		l.advance()
		switch l.current() {
		case '.':
			l.advance()
			if l.current() == '=' {
				l.advance()
				token = Token{TknDotDotEq, "..=", startLine, startColumn}
			} else {
				token = Token{TknDoubleDot, "..", startLine, startColumn}
			}
		default:
			token = Token{TknDot, ".", startLine, startColumn}
		}
	case '"':
		l.advance()
		text, err := l.readString(byte('"'))
		if err != nil {
			return Token{}, err
		}
		token = Token{TknString, text, startLine, startColumn}
	case '\'':
		l.advance()
		text, err := l.readString(byte('\''))
		if err != nil {
			return Token{}, err
		}
		token = Token{TknString, text, startLine, startColumn}
	case '`':
		l.advance()
		text, err := l.readRawString()
		if err != nil {
			return Token{}, err
		}
		token = Token{TknRawString, text, startLine, startColumn}
	default:
		if unicode.IsDigit(rune(ch)) {
			return l.readNumber(startLine, startColumn)
		} else if unicode.IsLetter(rune(ch)) || ch == '_' {
			return l.readIdentifier(startLine, startColumn)
		} else {
			return Token{}, fmt.Errorf("unexpected character at %d:%d: '%c'", l.line, l.column, ch)
		}
	}

	token.Line = startLine
	token.Column = startColumn
	if token.Text == "" {
		token.Text = string(l.input[startPos:l.pos])
	}

	return token, nil
}

func (l *Lexer) readIdentifier(line, column int) (Token, error) {
	start := l.pos

	for !l.isAtEnd() && (unicode.IsLetter(rune(l.current())) || unicode.IsDigit(rune(l.current())) || l.current() == '_') {
		l.advance()
	}

	ident := l.input[start:l.pos]

	// Check if it's a keyword
	if tknType, ok := l.keywords[ident]; ok {
		return Token{tknType, ident, line, column}, nil
	}

	return Token{TknIdentifier, ident, line, column}, nil
}

func (l *Lexer) readNumber(line, column int) (Token, error) {
	start := l.pos

	// 정수 부분
	for !l.isAtEnd() && unicode.IsDigit(rune(l.current())) {
		l.advance()
	}

	// 부동소수점 확인
	if l.current() == '.' && l.isDigitAt(l.pos+1) {
		l.advance() // skip '.'

		for !l.isAtEnd() && unicode.IsDigit(rune(l.current())) {
			l.advance()
		}

		return Token{TknFloat, l.input[start:l.pos], line, column}, nil
	}

	return Token{TknInteger, l.input[start:l.pos], line, column}, nil
}

func (l *Lexer) readString(quote byte) (string, error) {
	start := l.pos

	for !l.isAtEnd() && l.current() != quote {
		if l.current() == '\\' {
			l.advance()
			if !l.isAtEnd() {
				l.advance()
			}
		} else {
			if l.current() == '\n' {
				l.line++
				l.column = 0
			}
			l.advance()
		}
	}

	if l.isAtEnd() {
		return "", fmt.Errorf("unterminated string at line %d", l.line)
	}

	content := l.input[start:l.pos]
	l.advance() // skip closing quote

	return l.unescapeString(content), nil
}

func (l *Lexer) readRawString() (string, error) {
	start := l.pos

	for !l.isAtEnd() && l.current() != '`' {
		if l.current() == '\n' {
			l.line++
			l.column = 0
		}
		l.advance()
	}

	if l.isAtEnd() {
		return "", fmt.Errorf("unterminated raw string at line %d", l.line)
	}

	content := l.input[start:l.pos]
	l.advance() // skip closing backtick

	return content, nil
}

func (l *Lexer) unescapeString(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteByte('\n')
				i += 2
			case 't':
				result.WriteByte('\t')
				i += 2
			case 'r':
				result.WriteByte('\r')
				i += 2
			case '\\':
				result.WriteByte('\\')
				i += 2
			case '"':
				result.WriteByte('"')
				i += 2
			case '\'':
				result.WriteByte('\'')
				i += 2
			default:
				result.WriteByte(s[i+1])
				i += 2
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

func (l *Lexer) skipWhitespaceAndComments() {
	for !l.isAtEnd() {
		switch l.current() {
		case ' ', '\t', '\r':
			l.advance()
		case '\n':
			l.line++
			l.column = 0
			l.advance()
		case '/' :
			if l.peekNext() == '/' {
				// 한 줄 주석
				for !l.isAtEnd() && l.current() != '\n' {
					l.advance()
				}
			} else if l.peekNext() == '*' {
				// 블록 주석
				l.advance()
				l.advance()
				for !l.isAtEnd() {
					if l.current() == '*' && l.peekNext() == '/' {
						l.advance()
						l.advance()
						break
					}
					if l.current() == '\n' {
						l.line++
						l.column = 0
					}
					l.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

func (l *Lexer) current() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) isDigitAt(pos int) bool {
	if pos >= len(l.input) {
		return false
	}
	return unicode.IsDigit(rune(l.input[pos]))
}

func (l *Lexer) advance() {
	if !l.isAtEnd() {
		l.pos++
		l.column++
	}
}

func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.input)
}
