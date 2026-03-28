// Package lexer implements a tokenizer for mini FreeLang
package lexer

import (
	"unicode"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

type Lexer struct {
	input        string
	pos          int    // current position
	readPos      int    // next reading position
	ch           byte   // current character
	line         int
	col          int
	startCol     int
	prevLineLen  int
}

// New creates a new lexer for the input string
func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPos]
	}

	if l.ch == '\n' {
		l.prevLineLen = l.col
		l.line++
		l.col = 0
	} else {
		l.col++
	}

	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdent() string {
	startPos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[startPos:l.pos]
}

func (l *Lexer) readNumber() string {
	startPos := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPos:l.pos]
}

func (l *Lexer) NextToken() ast.Token {
	l.skipWhitespace()

	startCol := l.col
	startLine := l.line

	var tok ast.Token
	tok.Line = startLine
	tok.Col = startCol

	switch l.ch {
	case 0:
		tok.Type = ast.TokenEOF
		tok.Value = ""
	case '+':
		tok.Type = ast.TokenPlus
		tok.Value = "+"
		l.readChar()
	case '-':
		tok.Type = ast.TokenMinus
		tok.Value = "-"
		l.readChar()
	case '*':
		tok.Type = ast.TokenStar
		tok.Value = "*"
		l.readChar()
	case '/':
		tok.Type = ast.TokenSlash
		tok.Value = "/"
		l.readChar()
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = ast.TokenEq
			tok.Value = "=="
			l.readChar()
		} else {
			tok.Type = ast.TokenAssign
			tok.Value = "="
			l.readChar()
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = ast.TokenNe
			tok.Value = "!="
			l.readChar()
		} else {
			// Single ! not supported for now
			l.readChar()
			return l.NextToken()
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = ast.TokenLe
			tok.Value = "<="
			l.readChar()
		} else {
			tok.Type = ast.TokenLt
			tok.Value = "<"
			l.readChar()
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = ast.TokenGe
			tok.Value = ">="
			l.readChar()
		} else {
			tok.Type = ast.TokenGt
			tok.Value = ">"
			l.readChar()
		}
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			tok.Type = ast.TokenDotDot
			tok.Value = ".."
			l.readChar()
		} else {
			tok.Type = ast.TokenDot
			tok.Value = "."
			l.readChar()
		}
	case '(':
		tok.Type = ast.TokenLParen
		tok.Value = "("
		l.readChar()
	case ')':
		tok.Type = ast.TokenRParen
		tok.Value = ")"
		l.readChar()
	case '{':
		tok.Type = ast.TokenLBrace
		tok.Value = "{"
		l.readChar()
	case '}':
		tok.Type = ast.TokenRBrace
		tok.Value = "}"
		l.readChar()
	case ',':
		tok.Type = ast.TokenComma
		tok.Value = ","
		l.readChar()
	case ':':
		tok.Type = ast.TokenColon
		tok.Value = ":"
		l.readChar()
	case ';':
		tok.Type = ast.TokenSemicolon
		tok.Value = ";"
		l.readChar()
	default:
		if isLetter(l.ch) {
			ident := l.readIdent()
			tok.Type = lookupKeyword(ident)
			tok.Value = ident
		} else if isDigit(l.ch) {
			num := l.readNumber()
			tok.Type = ast.TokenInt
			tok.Value = num
		} else {
			l.readChar()
			return l.NextToken()
		}
	}

	return tok
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

func lookupKeyword(ident string) ast.TokenType {
	switch ident {
	case "let":
		return ast.TokenLet
	case "fn":
		return ast.TokenFn
	case "if":
		return ast.TokenIf
	case "for":
		return ast.TokenFor
	case "in":
		return ast.TokenIn
	case "return":
		return ast.TokenReturn
	default:
		return ast.TokenIdent
	}
}
