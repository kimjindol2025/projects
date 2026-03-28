// Package ast defines AST node types for the mini FreeLang subset
package ast

type NodeKind int

const (
	NodeProgram NodeKind = iota
	NodeLetDecl        // let x = expr
	NodeFnDecl         // fn name(params) body
	NodeIfStmt         // if cond { body }
	NodeForStmt        // for i in range { body }
	NodeBinaryExpr     // a + b, a * b
	NodeCallExpr       // fn(args)
	NodeIdent          // 변수 참조
	NodeIntLit         // 정수 리터럴
	NodeReturn         // return expr
	NodeBlockStmt      // { statements }
	NodeRangeExpr      // 0..10
	NodeStructDecl     // struct Name { fields }
	NodeFieldDecl      // fieldName: Type
	NodeFieldAccess    // expr.field
)

// Node represents a single AST node
type Node struct {
	Kind     NodeKind
	Value    string   // ident/literal 값
	Children []*Node  // 자식 노드
	Line     int
	Col      int
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
	TokenEq  // ==
	TokenNe  // !=
	TokenLt  // <
	TokenGt  // >
	TokenLe  // <=
	TokenGe  // >=
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
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}
