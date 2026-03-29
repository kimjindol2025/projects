// Package ast defines AST node types for the mini FreeLang subset
package ast

type NodeKind int

const (
	NodeProgram        NodeKind = iota
	NodeLetDecl                 // let x = expr
	NodeFnDecl                  // fn name(params) body
	NodeIfStmt                  // if cond { body }
	NodeForStmt                 // for i in range { body }
	NodeBinaryExpr              // a + b, a * b
	NodeCallExpr                // fn(args)
	NodeIdent                   // 변수 참조
	NodeIntLit                  // 정수 리터럴
	NodeReturn                  // return expr
	NodeBlockStmt               // { statements }
	NodeRangeExpr               // 0..10
	NodeStructDecl              // struct Name { fields }
	NodeFieldDecl               // fieldName: Type
	NodeFieldAccess             // expr.field
	NodeTypeAnnotation          // 타입 표현식 (int, bool, Point)
	NodeBoolLit                 // true / false 리터럴
	NodeStringLit               // "hello" 문자열 리터럴
	NodeStructLit               // Point{x: 1, y: 2} 구조체 초기화
	NodeArrayLit                // [1, 2, 3] 배열 리터럴
	NodeIndexExpr               // arr[i] 인덱싱
)

// Node represents a single AST node
type Node struct {
	Kind           NodeKind
	Value          string  // ident/literal 값
	Children       []*Node // 자식 노드
	Line           int
	Col            int
	TypeAnnotation string // 명시적 타입 어노테이션 ("int", "bool", "Point", "")
}

// TokenType defines token types
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenInt
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenAssign
	TokenEq // ==
	TokenNe // !=
	TokenLt // <
	TokenGt // >
	TokenLe // <=
	TokenGe // >=
	TokenDot
	TokenDotDot // ..
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenComma
	TokenColon
	TokenSemicolon
	TokenLet
	TokenFn
	TokenIf
	TokenFor
	TokenIn
	TokenReturn
	TokenStruct
	TokenTrue   // true 키워드
	TokenFalse  // false 키워드
	TokenString   // "hello" 문자열 리터럴
	TokenElse     // else 키워드
	TokenLBracket // [
	TokenRBracket // ]
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}
