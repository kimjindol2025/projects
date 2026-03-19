// Package lexer provides tokenization for FV 2.0 (V-compatible syntax)
package lexer

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Keywords
	TknFn TokenType = iota
	TknLet
	TknMut
	TknConst
	TknIf
	TknElse
	TknFor
	TknIn
	TknMatch
	TknType
	TknStruct
	TknInterface
	TknEnum
	TknTrait
	TknImpl
	TknReturn
	TknModule
	TknImport

	// Literals
	TknIdentifier
	TknInteger
	TknFloat
	TknString
	TknRawString
	TknTrue
	TknFalse
	TknNone

	// Operators
	TknPlus
	TknMinus
	TknStar
	TknSlash
	TknPercent
	TknCaret
	TknAmpersand
	TknPipe
	TknTilde
	TknLeftShift
	TknRightShift

	// Assignment & Comparison
	TknAssign
	TknPlusAssign
	TknMinusAssign
	TknStarAssign
	TknSlashAssign
	TknColonAssign // :=
	TknEq          // ==
	TknNe          // !=
	TknLt          // <
	TknLe          // <=
	TknGt          // >
	TknGe          // >=
	TknLogicalAnd  // &&
	TknLogicalOr   // ||
	TknNot         // !
	TknQuestion    // ?
	TknDot         // .
	TknDoubleDot   // ..
	TknDotDotEq    // ..=

	// Arrows & Compound
	TknArrow      // ->
	TknFatArrow   // =>
	TknColon      // :
	TknDoubleColon // ::

	// Delimiters
	TknLParen
	TknRParen
	TknLBrace
	TknRBrace
	TknLBracket
	TknRBracket
	TknComma
	TknSemicolon
	TknAt // @

	// Special
	TknEof
	TknError
)

var tokenTypeNames = map[TokenType]string{
	TknFn:          "fn",
	TknLet:         "let",
	TknMut:         "mut",
	TknConst:       "const",
	TknIf:          "if",
	TknElse:        "else",
	TknFor:         "for",
	TknIn:          "in",
	TknMatch:       "match",
	TknType:        "type",
	TknStruct:      "struct",
	TknInterface:   "interface",
	TknEnum:        "enum",
	TknTrait:       "trait",
	TknImpl:         "impl",
	TknReturn:      "return",
	TknModule:      "module",
	TknImport:      "import",
	TknTrue:        "true",
	TknFalse:       "false",
	TknNone:        "none",
	TknPlus:        "+",
	TknMinus:       "-",
	TknStar:        "*",
	TknSlash:       "/",
	TknPercent:     "%",
	TknCaret:       "^",
	TknAmpersand:   "&",
	TknPipe:        "|",
	TknTilde:       "~",
	TknLeftShift:   "<<",
	TknRightShift:  ">>",
	TknAssign:      "=",
	TknPlusAssign:  "+=",
	TknMinusAssign: "-=",
	TknStarAssign:  "*=",
	TknSlashAssign: "/=",
	TknColonAssign: ":=",
	TknEq:          "==",
	TknNe:          "!=",
	TknLt:          "<",
	TknLe:          "<=",
	TknGt:          ">",
	TknGe:          ">=",
	TknLogicalAnd:  "&&",
	TknLogicalOr:   "||",
	TknNot:         "!",
	TknQuestion:    "?",
	TknDot:         ".",
	TknDoubleDot:   "..",
	TknDotDotEq:    "..=",
	TknArrow:       "->",
	TknFatArrow:    "=>",
	TknColon:       ":",
	TknDoubleColon: "::",
	TknLParen:      "(",
	TknRParen:      ")",
	TknLBrace:      "{",
	TknRBrace:      "}",
	TknLBracket:    "[",
	TknRBracket:    "]",
	TknComma:       ",",
	TknSemicolon:   ";",
	TknAt:          "@",
	TknEof:         "EOF",
	TknError:       "ERROR",
	TknIdentifier:  "IDENT",
	TknInteger:     "INT",
	TknFloat:       "FLOAT",
	TknString:      "STRING",
	TknRawString:   "RAW_STRING",
}

// Token represents a single token with metadata
type Token struct {
	Type   TokenType
	Text   string
	Line   int
	Column int
}

// String returns a string representation of the token
func (t Token) String() string {
	name := tokenTypeNames[t.Type]
	if name == "" {
		name = fmt.Sprintf("UNKNOWN(%d)", t.Type)
	}

	if t.Type == TknIdentifier || t.Type == TknInteger || t.Type == TknFloat || t.Type == TknString {
		return fmt.Sprintf("%s(%s)@%d:%d", name, t.Text, t.Line, t.Column)
	}

	return fmt.Sprintf("%s@%d:%d", name, t.Line, t.Column)
}
