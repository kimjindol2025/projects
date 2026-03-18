package lexer

import (
	"fmt"
	"unicode"
)

// Lexer - Julia 렉서
type Lexer struct {
	input  string
	pos    int  // 현재 위치
	line   int  // 현재 줄
	column int  // 현재 열
	offset int  // 바이트 오프셋
	ch     rune // 현재 문자
	file   string
}

// NewLexer - 새로운 Lexer 생성
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 0,
		offset: 0,
		ch:     rune(0),
		file:   "<stdin>",
	}
	l.readChar()
	return l
}

// NewLexerWithFile - 파일명을 포함한 Lexer 생성
func NewLexerWithFile(input, filename string) *Lexer {
	l := NewLexer(input)
	l.file = filename
	return l
}

// readChar - 다음 문자 읽기
func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = rune(0)
	} else {
		l.ch = rune(l.input[l.pos])
	}
	l.pos++
	l.offset++
	l.column++
}

// peekChar - 다음 문자 미리보기 (읽지 않음)
func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return rune(0)
	}
	return rune(l.input[l.pos])
}

// peekCharN - n번째 앞의 문자 보기
func (l *Lexer) peekCharN(n int) rune {
	pos := l.pos - 1 + n
	if pos >= len(l.input) || pos < 0 {
		return rune(0)
	}
	return rune(l.input[pos])
}

// skipWhitespace - 공백과 탭 건너뛰기 (줄바꿈 제외)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// skipLineComment - 줄 주석 건너뛰기 (#...)
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
}

// skipBlockComment - 블록 주석 건너뛰기 (#=...=#)
func (l *Lexer) skipBlockComment() bool {
	if l.ch != '#' {
		return false
	}
	if l.peekChar() != '=' {
		return false
	}
	
	l.readChar() // '#'
	l.readChar() // '='
	
	depth := 1
	for depth > 0 && l.ch != rune(0) {
		if l.ch == '#' && l.peekChar() == '=' {
			l.readChar() // '#'
			l.readChar() // '='
			depth++
		} else if l.ch == '=' && l.peekChar() == '#' {
			l.readChar() // '='
			l.readChar() // '#'
			depth--
		} else {
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			l.readChar()
		}
	}
	return true
}

// NextToken - 다음 토큰 반환
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	
	// 주석 처리
	for {
		if l.ch == '#' {
			if l.peekChar() == '=' {
				if l.skipBlockComment() {
					l.skipWhitespace()
					continue
				}
			} else {
				l.skipLineComment()
				l.skipWhitespace()
				continue
			}
		}
		break
	}

	line := l.line
	column := l.column

	// EOF
	if l.ch == rune(0) {
		return l.makeToken(TokenEOF, "", line, column)
	}

	// 개행
	if l.ch == '\n' {
		l.readChar()
		l.line++
		l.column = 0
		return l.makeToken(TokenNewline, "\n", line, column)
	}

	// 문자열
	if l.ch == '"' || l.ch == '\'' {
		return l.readString(line, column)
	}

	// 숫자
	if unicode.IsDigit(l.ch) {
		return l.readNumber(line, column)
	}

	// 심볼 (:name)
	if l.ch == ':' && unicode.IsLetter(l.peekChar()) {
		return l.readSymbol(line, column)
	}

	// 식별자 또는 키워드 또는 매크로
	if unicode.IsLetter(l.ch) || l.ch == '_' {
		return l.readIdentifierOrKeyword(line, column)
	}

	// 연산자 및 기타
	return l.readOperatorOrDelimiter(line, column)
}

// readString - 문자열 읽기
func (l *Lexer) readString(line, column int) Token {
	quote := l.ch
	l.readChar()
	start := l.pos - 1

	for l.ch != quote && l.ch != rune(0) {
		if l.ch == '\\' {
			l.readChar() // escape 문자
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	end := l.pos
	if end > len(l.input) {
		end = len(l.input)
	}
	
	lexeme := l.input[start:end]
	if l.ch == quote {
		l.readChar()
	}

	return l.makeToken(TokenString, lexeme, line, column)
}

// readNumber - 숫자 읽기 (정수, 부동소수, 지수, 복소수, 유리수)
func (l *Lexer) readNumber(line, column int) Token {
	start := l.pos - 1
	if start < 0 {
		start = 0
	}

	isFloat := false
	isComplex := false
	isRational := false

	// 정수 부분
	for unicode.IsDigit(l.ch) {
		l.readChar()
	}

	// 부동소수점
	if l.ch == '.' && unicode.IsDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for unicode.IsDigit(l.ch) {
			l.readChar()
		}
	}

	// 지수 표기법
	if (l.ch == 'e' || l.ch == 'E') && (unicode.IsDigit(l.peekChar()) || l.peekChar() == '+' || l.peekChar() == '-') {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for unicode.IsDigit(l.ch) {
			l.readChar()
		}
	}

	// 복소수 (e.g., 1im, 2.5e-10im)
	if l.ch == 'i' && l.peekChar() == 'm' {
		isComplex = true
		l.readChar()
		l.readChar()
	}

	// 유리수 (e.g., 1//2)
	if l.ch == '/' && l.peekChar() == '/' {
		isRational = true
		l.readChar()
		l.readChar()
		for unicode.IsDigit(l.ch) {
			l.readChar()
		}
	}

	end := l.pos
	if end > len(l.input) {
		end = len(l.input)
	}
	lexeme := l.input[start:end]

	tokenType := TokenInteger
	if isFloat {
		tokenType = TokenFloat
	}
	if isComplex {
		tokenType = TokenComplexNumber
	}
	if isRational {
		tokenType = TokenRationalNumber
	}

	return l.makeToken(tokenType, lexeme, line, column)
}

// readSymbol - 심볼 읽기 (:name)
func (l *Lexer) readSymbol(line, column int) Token {
	l.readChar() // ':'
	start := l.pos

	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	end := l.pos
	if end > len(l.input) {
		end = len(l.input)
	}
	lexeme := ":" + l.input[start:end]

	return l.makeToken(TokenSymbol, lexeme, line, column)
}

// readIdentifierOrKeyword - 식별자 또는 키워드 읽기
func (l *Lexer) readIdentifierOrKeyword(line, column int) Token {
	start := l.pos - 1
	if start < 0 {
		start = 0
	}

	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' || l.ch == '!' || l.ch == '?' {
		l.readChar()
	}

	// 현재 위치 - 1이 끝 위치 (readChar()로 한 칸 넘어갔으므로)
	end := l.pos - 1
	if end > len(l.input) {
		end = len(l.input)
	}
	if end < start {
		end = start
	}
	lexeme := l.input[start:end]

	// 매크로 확인
	if l.ch == '(' || l.ch == '[' || l.ch == '{' {
		// 잠시 보류
	}

	// 키워드 확인
	if tokenType, ok := Keywords[lexeme]; ok {
		return l.makeToken(tokenType, lexeme, line, column)
	}

	return l.makeToken(TokenIdentifier, lexeme, line, column)
}

// readOperatorOrDelimiter - 연산자 또는 구분자 읽기
func (l *Lexer) readOperatorOrDelimiter(line, column int) Token {
	ch := l.ch
	next := l.peekChar()

	switch ch {
	case '+':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenPlusAssign, "+=", line, column)
		} else if next == '+' {
			l.readChar()
			return l.makeToken(TokenPlusPlus, "++", line, column)
		}
		return l.makeToken(TokenPlus, "+", line, column)

	case '-':
		l.readChar()
		if next == '>' {
			l.readChar()
			return l.makeToken(TokenArrow, "->", line, column)
		} else if next == '=' {
			l.readChar()
			return l.makeToken(TokenMinusAssign, "-=", line, column)
		} else if next == '-' {
			l.readChar()
			return l.makeToken(TokenMinusMinus, "--", line, column)
		}
		return l.makeToken(TokenMinus, "-", line, column)

	case '*':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenStarAssign, "*=", line, column)
		}
		return l.makeToken(TokenStar, "*", line, column)

	case '/':
		l.readChar()
		if next == '/' {
			l.readChar()
			return l.makeToken(TokenTripleDot, "//", line, column)
		} else if next == '=' {
			l.readChar()
			return l.makeToken(TokenSlashAssign, "/=", line, column)
		}
		return l.makeToken(TokenSlash, "/", line, column)

	case '%':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenPercentAssign, "%=", line, column)
		}
		return l.makeToken(TokenPercent, "%", line, column)

	case '^':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenCaretAssign, "^=", line, column)
		}
		return l.makeToken(TokenCaret, "^", line, column)

	case '=':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenEqualEqual, "==", line, column)
		} else if next == '>' {
			l.readChar()
			return l.makeToken(TokenFatArrow, "=>", line, column)
		}
		return l.makeToken(TokenEqual, "=", line, column)

	case '!':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenNotEqual, "!=", line, column)
		}
		return l.makeToken(TokenNot, "!", line, column)

	case '<':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenLessEqual, "<=", line, column)
		} else if next == '<' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return l.makeToken(TokenLeftShiftAssign, "<<=", line, column)
			}
			return l.makeToken(TokenLeftShift, "<<", line, column)
		} else if next == '|' {
			l.readChar()
			return l.makeToken(TokenCompArrow, "<|", line, column)
		}
		return l.makeToken(TokenLess, "<", line, column)

	case '>':
		l.readChar()
		if next == '=' {
			l.readChar()
			return l.makeToken(TokenGreaterEqual, ">=", line, column)
		} else if next == '>' {
			l.readChar()
			if l.ch == '>' {
				l.readChar()
				if l.ch == '=' {
					l.readChar()
					return l.makeToken(TokenUnsignedRightShiftAssign, ">>>=", line, column)
				}
				return l.makeToken(TokenUnsignedRightShift, ">>>", line, column)
			} else if l.ch == '=' {
				l.readChar()
				return l.makeToken(TokenRightShiftAssign, ">>=", line, column)
			}
			return l.makeToken(TokenRightShift, ">>", line, column)
		}
		return l.makeToken(TokenGreater, ">", line, column)

	case '&':
		l.readChar()
		if next == '&' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return l.makeToken(TokenLogicalAndAssign, "&&=", line, column)
			}
			return l.makeToken(TokenAnd, "&&", line, column)
		} else if next == '=' {
			l.readChar()
			return l.makeToken(TokenAndAssign, "&=", line, column)
		}
		return l.makeToken(TokenAmpersand, "&", line, column)

	case '|':
		l.readChar()
		if next == '|' {
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return l.makeToken(TokenLogicalOrAssign, "||=", line, column)
			}
			return l.makeToken(TokenOr, "||", line, column)
		} else if next == '>' {
			l.readChar()
			return l.makeToken(TokenPipeArrow, "|>", line, column)
		} else if next == '=' {
			l.readChar()
			return l.makeToken(TokenPipeAssign, "|=", line, column)
		}
		return l.makeToken(TokenPipe, "|", line, column)

	case '.':
		l.readChar()
		if next == '.' {
			l.readChar()
			if l.ch == '.' {
				l.readChar()
				return l.makeToken(TokenTripleDot, "...", line, column)
			}
			return l.makeToken(TokenDoubleDot, "..", line, column)
		} else if next == ',' {
			l.readChar()
			return l.makeToken(TokenComma_Dot, ",.", line, column)
		}
		return l.makeToken(TokenDot, ".", line, column)

	case ':':
		l.readChar()
		if next == ':' {
			l.readChar()
			return l.makeToken(TokenDoubleColon, "::", line, column)
		}
		return l.makeToken(TokenColon, ":", line, column)

	case '(':
		l.readChar()
		return l.makeToken(TokenLparen, "(", line, column)
	case ')':
		l.readChar()
		return l.makeToken(TokenRparen, ")", line, column)
	case '{':
		l.readChar()
		return l.makeToken(TokenLbrace, "{", line, column)
	case '}':
		l.readChar()
		return l.makeToken(TokenRbrace, "}", line, column)
	case '[':
		l.readChar()
		return l.makeToken(TokenLbracket, "[", line, column)
	case ']':
		l.readChar()
		return l.makeToken(TokenRbracket, "]", line, column)
	case ',':
		l.readChar()
		return l.makeToken(TokenComma, ",", line, column)
	case ';':
		l.readChar()
		return l.makeToken(TokenSemicolon, ";", line, column)
	case '~':
		l.readChar()
		return l.makeToken(TokenTilde, "~", line, column)
	case '$':
		l.readChar()
		return l.makeToken(TokenDollar, "$", line, column)
	case '@':
		l.readChar()
		return l.makeToken(TokenAt, "@", line, column)
	case '?':
		l.readChar()
		return l.makeToken(TokenQuestionMark, "?", line, column)

	default:
		l.readChar()
		errMsg := fmt.Sprintf("예상치 못한 문자: %c (U+%04X)", ch, ch)
		return l.makeErrorToken(errMsg, line, column)
	}
}

// makeToken - 토큰 생성 헬퍼
func (l *Lexer) makeToken(tokenType TokenType, lexeme string, line, column int) Token {
	return Token{
		Type:   tokenType,
		Lexeme: lexeme,
		Pos: Position{
			File:   l.file,
			Line:   line,
			Column: column,
			Offset: l.offset - len(lexeme),
		},
	}
}

// makeErrorToken - 에러 토큰 생성
func (l *Lexer) makeErrorToken(msg string, line, column int) Token {
	return Token{
		Type:   TokenError,
		Lexeme: msg,
		Pos: Position{
			File:   l.file,
			Line:   line,
			Column: column,
			Offset: l.offset,
		},
	}
}

// ScanAll - 모든 토큰을 스캔하여 반환
func (l *Lexer) ScanAll() []Token {
	var tokens []Token
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.Type == TokenEOF {
			break
		}
	}
	return tokens
}
