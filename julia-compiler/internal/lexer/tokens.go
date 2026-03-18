package lexer

// Phase 1: Lexer - 모든 Julia 토큰 정의

// TokenType - 토큰 종류
type TokenType int

const (
	// 기본
	TokenEOF TokenType = iota
	TokenError
	TokenNewline

	// 키워드 (49개)
	TokenKeywordAbstract
	TokenKeywordAnd
	TokenKeywordAs
	TokenKeywordBegin
	TokenKeywordBreak
	TokenKeywordCatch
	TokenKeywordContinue
	TokenKeywordConst
	TokenKeywordDo
	TokenKeywordElse
	TokenKeywordElseif
	TokenKeywordEnd
	TokenKeywordExport
	TokenKeywordFalse
	TokenKeywordFinal
	TokenKeywordFinally
	TokenKeywordFor
	TokenKeywordFunction
	TokenKeywordGlobal
	TokenKeywordIf
	TokenKeywordImport
	TokenKeywordIn
	TokenKeywordIsA
	TokenKeywordLet
	TokenKeywordLocal
	TokenKeywordMacro
	TokenKeywordModule
	TokenKeywordMutable
	TokenKeywordNot
	TokenKeywordNothing
	TokenKeywordOr
	TokenKeywordQuote
	TokenKeywordReturn
	TokenKeywordStruct
	TokenKeywordTrue
	TokenKeywordTry
	TokenKeywordUsing
	TokenKeywordWhile

	// 리터럴
	TokenInteger
	TokenFloat
	TokenString
	TokenSymbol
	TokenIdentifier
	TokenMacroName

	// 연산자 (50+)
	TokenPlus           // +
	TokenMinus          // -
	TokenStar           // *
	TokenSlash          // /
	TokenPercent        // %
	TokenCaret          // ^
	TokenEqual          // =
	TokenEqualEqual     // ==
	TokenNotEqual       // !=, ≠
	TokenLess           // <
	TokenLessEqual      // <=, ≤
	TokenGreater        // >
	TokenGreaterEqual   // >=, ≥
	TokenAnd            // &&
	TokenOr             // ||
	TokenXor            // xor, ⊻
	TokenNot            // !
	TokenTilde          // ~
	TokenAmpersand      // &
	TokenPipe           // |
	TokenDot            // .
	TokenDoubleDot      // ..
	TokenTripleDot      // ...
	TokenDoubleColon    // ::
	TokenArrow          // ->
	TokenFatArrow       // =>
	TokenPipeArrow      // |>
	TokenCompArrow      // <|
	TokenLeftShift      // <<
	TokenRightShift     // >>
	TokenUnsignedRightShift // >>>
	TokenPlusAssign     // +=
	TokenMinusAssign    // -=
	TokenStarAssign     // *=
	TokenSlashAssign    // /=
	TokenPercentAssign  // %=
	TokenCaretAssign    // ^=
	TokenAndAssign      // &=
	TokenPipeAssign     // |=
	TokenLeftShiftAssign // <<=
	TokenRightShiftAssign // >>=
	TokenUnsignedRightShiftAssign // >>>=
	TokenLogicalAndAssign // &&=
	TokenLogicalOrAssign  // ||=
	TokenPlusPlus       // ++
	TokenMinusMinus     // --
	TokenDollar         // $
	TokenAt             // @

	// 괄호
	TokenLparen    // (
	TokenRparen    // )
	TokenLbrace    // {
	TokenRbrace    // }
	TokenLbracket  // [
	TokenRbracket  // ]

	// 구분자
	TokenComma
	TokenSemicolon
	TokenColon
	TokenQuestionMark
	TokenComma_Dot    // ,.

	// 특수
	TokenComplexNumber // 1im, 2.5e-10im
	TokenRationalNumber // 1//2
)

// TokenTypeNames - 토큰 이름 매핑
var tokenTypeNames = map[TokenType]string{
	TokenEOF:          "EOF",
	TokenError:        "ERROR",
	TokenNewline:      "NEWLINE",
	TokenInteger:      "INTEGER",
	TokenFloat:        "FLOAT",
	TokenString:       "STRING",
	TokenSymbol:       "SYMBOL",
	TokenIdentifier:   "IDENT",
	TokenMacroName:    "MACRO",
	TokenPlus:         "PLUS",
	TokenMinus:        "MINUS",
	TokenStar:         "STAR",
	TokenSlash:        "SLASH",
	TokenEqual:        "ASSIGN",
	TokenEqualEqual:   "EQ",
	TokenNotEqual:     "NE",
	TokenLparen:       "LPAREN",
	TokenRparen:       "RPAREN",
	TokenLbrace:       "LBRACE",
	TokenRbrace:       "RBRACE",
	TokenLbracket:     "LBRACKET",
	TokenRbracket:     "RBRACKET",
	TokenComma:        "COMMA",
	TokenDoubleColon:  "COLONCOLON",
	TokenArrow:        "ARROW",
	TokenPipeArrow:    "PIPEARROW",
}

func (t TokenType) String() string {
	if name, ok := tokenTypeNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Keywords - Julia 키워드 맵
var Keywords = map[string]TokenType{
	"abstract":   TokenKeywordAbstract,
	"and":        TokenKeywordAnd,
	"as":         TokenKeywordAs,
	"begin":      TokenKeywordBegin,
	"break":      TokenKeywordBreak,
	"catch":      TokenKeywordCatch,
	"continue":   TokenKeywordContinue,
	"const":      TokenKeywordConst,
	"do":         TokenKeywordDo,
	"else":       TokenKeywordElse,
	"elseif":     TokenKeywordElseif,
	"end":        TokenKeywordEnd,
	"export":     TokenKeywordExport,
	"false":      TokenKeywordFalse,
	"final":      TokenKeywordFinal,
	"finally":    TokenKeywordFinally,
	"for":        TokenKeywordFor,
	"function":   TokenKeywordFunction,
	"global":     TokenKeywordGlobal,
	"if":         TokenKeywordIf,
	"import":     TokenKeywordImport,
	"in":         TokenKeywordIn,
	"isa":        TokenKeywordIsA,
	"let":        TokenKeywordLet,
	"local":      TokenKeywordLocal,
	"macro":      TokenKeywordMacro,
	"module":     TokenKeywordModule,
	"mutable":    TokenKeywordMutable,
	"not":        TokenKeywordNot,
	"nothing":    TokenKeywordNothing,
	"or":         TokenKeywordOr,
	"quote":      TokenKeywordQuote,
	"return":     TokenKeywordReturn,
	"struct":     TokenKeywordStruct,
	"true":       TokenKeywordTrue,
	"try":        TokenKeywordTry,
	"using":      TokenKeywordUsing,
	"while":      TokenKeywordWhile,
}
