// Package ast defines the Abstract Syntax Tree for FV 2.0
package ast

// Program represents the root AST node
type Program struct {
	Definitions []Definition
	MainBody    []Statement
}

// Definition represents top-level declarations
type Definition interface {
	definitionNode()
}

// FunctionDef represents a function definition
type FunctionDef struct {
	Name       string
	Parameters []Parameter
	ReturnType *Type
	Body       []Statement
}

func (*FunctionDef) definitionNode() {}

// TypeDef represents a type definition
type TypeDef struct {
	Name   string
	Fields []Field
}

func (*TypeDef) definitionNode() {}

// StructDef represents a struct definition
type StructDef struct {
	Name   string
	Fields []Field
}

func (*StructDef) definitionNode() {}

// InterfaceDef represents an interface definition
type InterfaceDef struct {
	Name    string
	Methods []MethodSig
}

func (*InterfaceDef) definitionNode() {}

// EnumDef represents an enum definition
type EnumDef struct {
	Name     string
	Variants []EnumVariant
}

func (*EnumDef) definitionNode() {}

// ExternDef represents an extern function declaration
type ExternDef struct {
	Name       string
	Parameters []Parameter
	ReturnType *Type
	Library    string // optional library name from @link annotation
}

func (*ExternDef) definitionNode() {}

// ImportStatement represents an import declaration
type ImportStatement struct {
	Module string // module name like "stdio", "math", etc.
}

func (*ImportStatement) definitionNode() {}

// Parameter represents a function parameter
type Parameter struct {
	Name string
	Type *Type
}

// Field represents a struct/type field
type Field struct {
	Name      string
	Type      *Type
	Mutable   bool
	Default   Expression
}

// MethodSig represents a method signature
type MethodSig struct {
	Name       string
	Parameters []Parameter
	ReturnType *Type
}

// EnumVariant represents an enum variant
type EnumVariant struct {
	Name  string
	Types []Type
}

// Statement represents a statement
type Statement interface {
	statementNode()
}

// LetStatement represents a let binding
type LetStatement struct {
	Name    string
	Type    *Type
	Init    Expression
	Mutable bool
}

func (*LetStatement) statementNode() {}

// ConstStatement represents a const binding
type ConstStatement struct {
	Name  string
	Type  *Type
	Value Expression
}

func (*ConstStatement) statementNode() {}

// IfStatement represents an if/else statement
type IfStatement struct {
	Condition Expression
	ThenBody  []Statement
	ElseBody  []Statement
}

func (*IfStatement) statementNode() {}

// ForStatement represents a for loop
type ForStatement struct {
	Variable string
	Iterator Expression
	Body     []Statement
}

func (*ForStatement) statementNode() {}

// ForRangeStatement represents a for range loop
type ForRangeStatement struct {
	Variable string
	Start    Expression
	End      Expression
	Body     []Statement
}

func (*ForRangeStatement) statementNode() {}

// MatchStatement represents a match expression
type MatchStatement struct {
	Expression Expression
	Arms       []MatchArm
}

func (*MatchStatement) statementNode() {}

// MatchArm represents a match arm
type MatchArm struct {
	Pattern Pattern
	Body    []Statement
}

// Pattern represents a match pattern
type Pattern interface {
	patternNode()
}

// LiteralPattern matches a literal
type LiteralPattern struct {
	Value Expression
}

func (*LiteralPattern) patternNode() {}

// IdentifierPattern matches an identifier
type IdentifierPattern struct {
	Name string
}

func (*IdentifierPattern) patternNode() {}

// WildcardPattern matches anything
type WildcardPattern struct{}

func (*WildcardPattern) patternNode() {}

// ExpressionStatement wraps an expression as a statement
type ExpressionStatement struct {
	Expression Expression
}

func (*ExpressionStatement) statementNode() {}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Value Expression
}

func (*ReturnStatement) statementNode() {}

// BlockStatement represents a block of statements
type BlockStatement struct {
	Statements []Statement
}

func (*BlockStatement) statementNode() {}

// Expression represents an expression
type Expression interface {
	expressionNode()
}

// IntegerLiteral represents an integer
type IntegerLiteral struct {
	Value int64
}

func (*IntegerLiteral) expressionNode() {}

// FloatLiteral represents a float
type FloatLiteral struct {
	Value float64
}

func (*FloatLiteral) expressionNode() {}

// StringLiteral represents a string
type StringLiteral struct {
	Value string
}

func (*StringLiteral) expressionNode() {}

// BoolLiteral represents a boolean
type BoolLiteral struct {
	Value bool
}

func (*BoolLiteral) expressionNode() {}

// NoneLiteral represents none
type NoneLiteral struct{}

func (*NoneLiteral) expressionNode() {}

// Identifier represents an identifier
type Identifier struct {
	Name string
}

func (*Identifier) expressionNode() {}

// BinaryExpression represents binary operations
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (*BinaryExpression) expressionNode() {}

// UnaryExpression represents unary operations
type UnaryExpression struct {
	Operator string
	Operand  Expression
}

func (*UnaryExpression) expressionNode() {}

// CallExpression represents a function call
type CallExpression struct {
	Function  Expression
	Arguments []Expression
}

func (*CallExpression) expressionNode() {}

// MethodCallExpression represents a method call
type MethodCallExpression struct {
	Object     Expression
	Method     string
	Arguments  []Expression
}

func (*MethodCallExpression) expressionNode() {}

// IndexExpression represents indexing
type IndexExpression struct {
	Object Expression
	Index  Expression
}

func (*IndexExpression) expressionNode() {}

// FieldExpression represents field access
type FieldExpression struct {
	Object Expression
	Field  string
}

func (*FieldExpression) expressionNode() {}

// IfExpression represents an if expression
type IfExpression struct {
	Condition Expression
	ThenExpr  Expression
	ElseExpr  Expression
}

func (*IfExpression) expressionNode() {}

// MatchExpression represents a match expression
type MatchExpression struct {
	Expression Expression
	Arms       []MatchExprArm
}

func (*MatchExpression) expressionNode() {}

// MatchExprArm represents a match arm in an expression
type MatchExprArm struct {
	Pattern Pattern
	Value   Expression
}

// ArrayExpression represents an array literal
type ArrayExpression struct {
	Elements []Expression
}

func (*ArrayExpression) expressionNode() {}

// StructExpression represents a struct literal
type StructExpression struct {
	Name   string
	Fields map[string]Expression
}

func (*StructExpression) expressionNode() {}

// Type represents a type
type Type struct {
	Name       string
	IsOption   bool
	IsResult   bool
	ErrorType  *Type
	ElementType *Type
	IsArray    bool
	IsPrimitive bool
	IsFunction bool
	ParamTypes []Type
	ReturnType *Type
}

// String returns string representation of type
func (t *Type) String() string {
	if t == nil {
		return "unknown"
	}

	if t.IsOption {
		return "?" + t.ElementType.String()
	}
	if t.IsResult {
		errStr := "String"
		if t.ErrorType != nil {
			errStr = t.ErrorType.String()
		}
		return "Result(" + t.ElementType.String() + ", " + errStr + ")"
	}
	if t.IsArray {
		return "[]" + t.ElementType.String()
	}
	if t.IsFunction {
		return "fn"
	}
	return t.Name
}

// ErrorPropagation represents the ? operator
type ErrorPropagation struct {
	Expression Expression
}

func (*ErrorPropagation) expressionNode() {}

// CastExpression represents type casting
type CastExpression struct {
	Expression Expression
	TargetType *Type
}

func (*CastExpression) expressionNode() {}
