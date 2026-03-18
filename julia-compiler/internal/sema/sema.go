package sema

import (
	"fmt"
	"juliacc/internal/ast"
	"juliacc/internal/lexer"
	"juliacc/internal/types"
)

// Error represents a semantic analysis error
type Error struct {
	Message string
	Line    int
	Column  int
}

// Analyzer performs semantic analysis on the AST
type Analyzer struct {
	scopes      *ScopeStack
	hierarchy   *Hierarchy
	errors      []Error
	typeHier    *types.Hierarchy
	dispatch    *types.Dispatch
	currentFunc *FunctionContext
}

// FunctionContext tracks info about the current function
type FunctionContext struct {
	name       string
	returnType types.Type
	params     map[string]types.Type
}

// Hierarchy is a type hierarchy wrapper
type Hierarchy struct {
	hierarchy *types.Hierarchy
}

// NewAnalyzer creates a new semantic analyzer
func NewAnalyzer(typeReg *types.Registry, typeHier *types.Hierarchy) *Analyzer {
	dispatch := types.NewDispatch(typeHier)

	return &Analyzer{
		scopes:    NewScopeStack(typeReg),
		hierarchy: &Hierarchy{hierarchy: typeHier},
		errors:    make([]Error, 0),
		typeHier:  typeHier,
		dispatch:  dispatch,
	}
}

// Analyze performs semantic analysis on a program
func (a *Analyzer) Analyze(program *ast.Program) error {
	// Initialize builtin functions and types
	a.initializeBuiltins()

	// Analyze each statement
	for _, stmt := range program.Statements {
		a.analyzeStatement(stmt)
	}

	if len(a.errors) > 0 {
		return fmt.Errorf("semantic analysis failed with %d errors", len(a.errors))
	}

	return nil
}

// GetErrors returns all collected errors
func (a *Analyzer) GetErrors() []Error {
	return a.errors
}

// addError adds a new error
func (a *Analyzer) addError(message string, token lexer.Token) {
	a.errors = append(a.errors, Error{
		Message: message,
		Line:    token.Pos.Line,
		Column:  token.Pos.Column,
	})
}

// initializeBuiltins registers built-in functions and types
func (a *Analyzer) initializeBuiltins() {
	// Register built-in functions
	printMethods := a.dispatch.RegisterFunction("print")
	anyType := a.scopes.TypeRegistry().Get("Any")
	nothingType := a.scopes.TypeRegistry().Get("Nothing")

	printMethods.AddMethod([]types.Type{anyType}, nothingType, true)

	// Register length function
	lengthMethods := a.dispatch.RegisterFunction("length")
	arrayType := a.scopes.TypeRegistry().Get("Array")
	int64Type := a.scopes.TypeRegistry().Get("Int64")
	lengthMethods.AddMethod([]types.Type{arrayType}, int64Type, false)

	// Register arithmetic operators
	a.registerArithmetic()

	// Register type constructors
	a.registerTypeConstructors()
}

// registerArithmetic registers arithmetic operations
func (a *Analyzer) registerArithmetic() {
	numberType := a.scopes.TypeRegistry().Get("Number")
	int64Type := a.scopes.TypeRegistry().Get("Int64")
	float64Type := a.scopes.TypeRegistry().Get("Float64")

	// +
	plusMethods := a.dispatch.RegisterFunction("+")
	plusMethods.AddMethod([]types.Type{int64Type, int64Type}, int64Type, false)
	plusMethods.AddMethod([]types.Type{float64Type, float64Type}, float64Type, false)
	plusMethods.AddMethod([]types.Type{numberType, numberType}, numberType, false)

	// -
	minusMethods := a.dispatch.RegisterFunction("-")
	minusMethods.AddMethod([]types.Type{int64Type, int64Type}, int64Type, false)
	minusMethods.AddMethod([]types.Type{float64Type, float64Type}, float64Type, false)

	// *
	mulMethods := a.dispatch.RegisterFunction("*")
	mulMethods.AddMethod([]types.Type{int64Type, int64Type}, int64Type, false)
	mulMethods.AddMethod([]types.Type{float64Type, float64Type}, float64Type, false)

	// /
	divMethods := a.dispatch.RegisterFunction("/")
	divMethods.AddMethod([]types.Type{int64Type, int64Type}, float64Type, false)
	divMethods.AddMethod([]types.Type{float64Type, float64Type}, float64Type, false)
}

// registerTypeConstructors registers type constructors
func (a *Analyzer) registerTypeConstructors() {
	// Vector{T}(n) constructor
	// Array{T}(dims...) constructor
	// etc.
}

// analyzeStatement analyzes a single statement
func (a *Analyzer) analyzeStatement(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		a.analyzeExprStmt(s)
	case *ast.Assignment:
		a.analyzeAssignment(s)
	case *ast.VarDecl:
		a.analyzeVarDecl(s)
	case *ast.ConstDecl:
		a.analyzeConstDecl(s)
	case *ast.FunctionDecl:
		a.analyzeFunctionDecl(s)
	case *ast.StructDecl:
		a.analyzeStructDecl(s)
	case *ast.IfStmt:
		a.analyzeIfStmt(s)
	case *ast.WhileStmt:
		a.analyzeWhileStmt(s)
	case *ast.ForStmt:
		a.analyzeForStmt(s)
	case *ast.ReturnStmt:
		a.analyzeReturnStmt(s)
	case *ast.TryStmt:
		a.analyzeTryStmt(s)
	case *ast.BreakStmt:
		// Break is valid in loops
	case *ast.ContinueStmt:
		// Continue is valid in loops
	}
}

// analyzeExprStmt analyzes an expression statement
func (a *Analyzer) analyzeExprStmt(stmt *ast.ExprStmt) {
	a.analyzeExpression(stmt.Expr)
}

// analyzeAssignment analyzes an assignment
func (a *Analyzer) analyzeAssignment(assign *ast.Assignment) {
	leftType := a.analyzeExpression(assign.Left)
	rightType := a.analyzeExpression(assign.Right)

	if leftType == nil || rightType == nil {
		return
	}

	// Check type compatibility
	if !rightType.IsSubtypeOf(leftType) {
		a.errors = append(a.errors, Error{
			Message: fmt.Sprintf("cannot assign %s to %s", rightType.String(), leftType.String()),
			Line:    assign.Equal.Pos.Line,
			Column:  assign.Equal.Pos.Column,
		})
	}
}

// analyzeVarDecl analyzes a variable declaration
func (a *Analyzer) analyzeVarDecl(decl *ast.VarDecl) {
	// Analyze initializer if present
	var valueType types.Type
	if decl.Value != nil {
		valueType = a.analyzeExpression(decl.Value)
	}

	// Create symbol
	sym := &Symbol{
		Name:    decl.Name,
		Type:    valueType,
		Kind:    KindVariable,
		Mutable: true,
	}

	if err := a.scopes.Define(decl.Name, sym); err != nil {
		a.errors = append(a.errors, Error{
			Message: err.Error(),
			Line:    decl.Let.Pos.Line,
			Column:  decl.Let.Pos.Column,
		})
	}
}

// analyzeConstDecl analyzes a constant declaration
func (a *Analyzer) analyzeConstDecl(decl *ast.ConstDecl) {
	valueType := a.analyzeExpression(decl.Value)

	sym := &Symbol{
		Name:    decl.Name,
		Type:    valueType,
		Kind:    KindConstant,
		Mutable: false,
	}

	if err := a.scopes.Define(decl.Name, sym); err != nil {
		a.errors = append(a.errors, Error{
			Message: err.Error(),
			Line:    decl.Const.Pos.Line,
			Column:  decl.Const.Pos.Column,
		})
	}
}

// analyzeFunctionDecl analyzes a function declaration
func (a *Analyzer) analyzeFunctionDecl(decl *ast.FunctionDecl) {
	// Create function context
	ctx := &FunctionContext{
		name:   decl.Name,
		params: make(map[string]types.Type),
	}

	a.currentFunc = ctx

	// Enter new scope
	a.scopes.Push(decl.Name)

	// Parse return type if specified
	var returnType types.Type
	if decl.ReturnType != "" {
		returnType = a.scopes.TypeRegistry().Get(decl.ReturnType)
		if returnType == nil {
			returnType = a.scopes.TypeRegistry().Get("Any")
		}
	} else {
		returnType = a.scopes.TypeRegistry().Get("Any")
	}

	ctx.returnType = returnType

	// Add parameters to scope
	for _, param := range decl.Parameters {
		var paramType types.Type
		if param.Type != "" {
			paramType = a.scopes.TypeRegistry().Get(param.Type)
			if paramType == nil {
				paramType = a.scopes.TypeRegistry().Get("Any")
			}
		} else {
			paramType = a.scopes.TypeRegistry().Get("Any")
		}

		sym := &Symbol{
			Name:    param.Name,
			Type:    paramType,
			Kind:    KindVariable,
			Mutable: true,
		}

		a.scopes.Define(param.Name, sym)
		ctx.params[param.Name] = paramType
	}

	// Analyze function body
	for _, stmt := range decl.Body {
		a.analyzeStatement(stmt)
	}

	// Leave function scope
	a.scopes.Pop()

	// Register function in dispatch
	paramTypes := make([]types.Type, len(decl.Parameters))
	for i, param := range decl.Parameters {
		paramType := a.scopes.TypeRegistry().Get(param.Type)
		if paramType == nil {
			paramType = a.scopes.TypeRegistry().Get("Any")
		}
		paramTypes[i] = paramType
	}

	methodTable := a.dispatch.GetMethodTable(decl.Name)
	if methodTable == nil {
		methodTable = a.dispatch.RegisterFunction(decl.Name)
	}
	methodTable.AddMethod(paramTypes, returnType, false)

	a.currentFunc = nil
}

// analyzeStructDecl analyzes a struct declaration
func (a *Analyzer) analyzeStructDecl(decl *ast.StructDecl) {
	// Register the struct as a type
	sym := &Symbol{
		Name:    decl.Name,
		Type:    nil, // Will be a struct type
		Kind:    KindType,
		Mutable: decl.IsMutable,
	}

	if err := a.scopes.Define(decl.Name, sym); err != nil {
		a.errors = append(a.errors, Error{
			Message: err.Error(),
			Line:    0,
			Column:  0,
		})
	}
}

// analyzeIfStmt analyzes an if statement
func (a *Analyzer) analyzeIfStmt(stmt *ast.IfStmt) {
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && condType.Kind() != types.KindBool {
		a.errors = append(a.errors, Error{
			Message: fmt.Sprintf("if condition must be Bool, got %s", condType.String()),
			Line:    0,
			Column:  0,
		})
	}

	for _, s := range stmt.Then {
		a.analyzeStatement(s)
	}

	for _, elseif := range stmt.ElseIfs {
		eType := a.analyzeExpression(elseif.Condition)
		if eType != nil && eType.Kind() != types.KindBool {
			a.errors = append(a.errors, Error{
				Message: fmt.Sprintf("elseif condition must be Bool, got %s", eType.String()),
				Line:    0,
				Column:  0,
			})
		}
		for _, s := range elseif.Body {
			a.analyzeStatement(s)
		}
	}

	for _, s := range stmt.Else {
		a.analyzeStatement(s)
	}
}

// analyzeWhileStmt analyzes a while statement
func (a *Analyzer) analyzeWhileStmt(stmt *ast.WhileStmt) {
	condType := a.analyzeExpression(stmt.Condition)
	if condType != nil && condType.Kind() != types.KindBool {
		a.errors = append(a.errors, Error{
			Message: fmt.Sprintf("while condition must be Bool, got %s", condType.String()),
			Line:    0,
			Column:  0,
		})
	}

	for _, s := range stmt.Body {
		a.analyzeStatement(s)
	}
}

// analyzeForStmt analyzes a for statement
func (a *Analyzer) analyzeForStmt(stmt *ast.ForStmt) {
	a.scopes.Push("for_loop")
	defer a.scopes.Pop()

	iterType := a.analyzeExpression(stmt.Iterator)

	// Add loop variable
	sym := &Symbol{
		Name:    stmt.Variable,
		Type:    iterType,
		Kind:    KindVariable,
		Mutable: true,
	}
	a.scopes.Define(stmt.Variable, sym)

	for _, s := range stmt.Body {
		a.analyzeStatement(s)
	}
}

// analyzeReturnStmt analyzes a return statement
func (a *Analyzer) analyzeReturnStmt(stmt *ast.ReturnStmt) {
	if a.currentFunc == nil {
		a.errors = append(a.errors, Error{
			Message: "return outside function",
			Line:    0,
			Column:  0,
		})
		return
	}

	if stmt.Value != nil {
		retType := a.analyzeExpression(stmt.Value)
		if retType != nil && !retType.IsSubtypeOf(a.currentFunc.returnType) {
			a.errors = append(a.errors, Error{
				Message: fmt.Sprintf("return type mismatch: got %s, expected %s",
					retType.String(), a.currentFunc.returnType.String()),
				Line: 0,
				Column: 0,
			})
		}
	}
}

// analyzeTryStmt analyzes a try statement
func (a *Analyzer) analyzeTryStmt(stmt *ast.TryStmt) {
	for _, s := range stmt.Try {
		a.analyzeStatement(s)
	}

	for _, c := range stmt.Catches {
		a.scopes.Push("catch_block")
		a.scopes.Define(c.Variable, &Symbol{
			Name:    c.Variable,
			Type:    a.scopes.TypeRegistry().Get("Any"),
			Kind:    KindVariable,
			Mutable: false,
		})
		for _, s := range c.Body {
			a.analyzeStatement(s)
		}
		a.scopes.Pop()
	}

	for _, s := range stmt.Finally {
		a.analyzeStatement(s)
	}
}

// analyzeExpression analyzes an expression and returns its type
func (a *Analyzer) analyzeExpression(expr ast.Expr) types.Type {
	switch e := expr.(type) {
	case *ast.Literal:
		return a.analyzeLiteral(e)
	case *ast.Identifier:
		return a.analyzeIdentifier(e)
	case *ast.BinaryOp:
		return a.analyzeBinaryOp(e)
	case *ast.UnaryOp:
		return a.analyzeUnaryOp(e)
	case *ast.Call:
		return a.analyzeCall(e)
	case *ast.Index:
		return a.analyzeIndex(e)
	case *ast.MemberAccess:
		return a.analyzeMemberAccess(e)
	case *ast.TypeAnnotation:
		return a.analyzeTypeAnnotation(e)
	case *ast.ArrayLiteral:
		return a.analyzeArrayLiteral(e)
	case *ast.TupleLiteral:
		return a.analyzeTupleLiteral(e)
	case *ast.DictLiteral:
		return a.analyzeDictLiteral(e)
	default:
		return a.scopes.TypeRegistry().Get("Any")
	}
}

// analyzeLiteral analyzes a literal and returns its type
func (a *Analyzer) analyzeLiteral(lit *ast.Literal) types.Type {
	switch lit.Token.Type {
	case lexer.TokenInteger:
		return a.scopes.TypeRegistry().Get("Int64")
	case lexer.TokenFloat:
		return a.scopes.TypeRegistry().Get("Float64")
	case lexer.TokenString:
		return a.scopes.TypeRegistry().Get("String")
	case lexer.TokenSymbol:
		return a.scopes.TypeRegistry().Get("Symbol")
	case lexer.TokenKeywordTrue, lexer.TokenKeywordFalse:
		return a.scopes.TypeRegistry().Get("Bool")
	case lexer.TokenKeywordNothing:
		return a.scopes.TypeRegistry().Get("Nothing")
	default:
		return a.scopes.TypeRegistry().Get("Any")
	}
}

// analyzeIdentifier analyzes an identifier and returns its type
func (a *Analyzer) analyzeIdentifier(id *ast.Identifier) types.Type {
	sym := a.scopes.Resolve(id.Name)
	if sym == nil {
		a.errors = append(a.errors, Error{
			Message: fmt.Sprintf("undefined variable: %s", id.Name),
			Line:    id.Token.Pos.Line,
			Column:  id.Token.Pos.Column,
		})
		return a.scopes.TypeRegistry().Get("Any")
	}
	return sym.Type
}

// analyzeBinaryOp analyzes a binary operation
func (a *Analyzer) analyzeBinaryOp(op *ast.BinaryOp) types.Type {
	leftType := a.analyzeExpression(op.Left)
	rightType := a.analyzeExpression(op.Right)

	if leftType == nil || rightType == nil {
		return a.scopes.TypeRegistry().Get("Any")
	}

	// Determine operation name
	opName := a.getOperatorName(op.Op)
	if opName == "" {
		return a.scopes.TypeRegistry().Get("Any")
	}

	// Look up method
	method, err := a.dispatch.LookupMethod(opName, []types.Type{leftType, rightType})
	if err != nil {
		a.errors = append(a.errors, Error{
			Message: err.Error(),
			Line:    op.OpToken.Pos.Line,
			Column:  op.OpToken.Pos.Column,
		})
		return a.scopes.TypeRegistry().Get("Any")
	}

	return method.ReturnType
}

// analyzeUnaryOp analyzes a unary operation
func (a *Analyzer) analyzeUnaryOp(op *ast.UnaryOp) types.Type {
	operandType := a.analyzeExpression(op.Operand)
	if operandType == nil {
		return a.scopes.TypeRegistry().Get("Any")
	}

	// Handle unary operators
	switch op.Op {
	case lexer.TokenPlus:
		return operandType
	case lexer.TokenMinus:
		return operandType
	case lexer.TokenKeywordNot:
		return a.scopes.TypeRegistry().Get("Bool")
	case lexer.TokenTilde:
		return operandType
	default:
		return operandType
	}
}

// analyzeCall analyzes a function call
func (a *Analyzer) analyzeCall(call *ast.Call) types.Type {
	// Get function name
	var funcName string
	if id, ok := call.Function.(*ast.Identifier); ok {
		funcName = id.Name
	} else {
		return a.scopes.TypeRegistry().Get("Any")
	}

	// Analyze arguments
	argTypes := make([]types.Type, len(call.Arguments))
	for i, arg := range call.Arguments {
		argTypes[i] = a.analyzeExpression(arg)
	}

	// Look up method
	method, err := a.dispatch.LookupMethod(funcName, argTypes)
	if err != nil {
		a.errors = append(a.errors, Error{
			Message: err.Error(),
			Line:    call.LParen.Pos.Line,
			Column:  call.LParen.Pos.Column,
		})
		return a.scopes.TypeRegistry().Get("Any")
	}

	return method.ReturnType
}

// analyzeIndex analyzes array indexing
func (a *Analyzer) analyzeIndex(idx *ast.Index) types.Type {
	arrayType := a.analyzeExpression(idx.Object)
	if arrayType == nil {
		return a.scopes.TypeRegistry().Get("Any")
	}

	// For now, assume indexing returns Any
	// TODO: track element types
	return a.scopes.TypeRegistry().Get("Any")
}

// analyzeMemberAccess analyzes member access
func (a *Analyzer) analyzeMemberAccess(ma *ast.MemberAccess) types.Type {
	// TODO: implement based on struct types
	return a.scopes.TypeRegistry().Get("Any")
}

// analyzeTypeAnnotation analyzes type annotation
func (a *Analyzer) analyzeTypeAnnotation(ta *ast.TypeAnnotation) types.Type {
	exprType := a.analyzeExpression(ta.Expr)
	annotatedType := a.scopes.TypeRegistry().Get(ta.Type)

	if annotatedType == nil {
		return exprType
	}

	if exprType != nil && !exprType.IsSubtypeOf(annotatedType) {
		a.errors = append(a.errors, Error{
			Message: fmt.Sprintf("expression type %s is not subtype of annotated type %s",
				exprType.String(), annotatedType.String()),
			Line:   ta.ColonColon.Pos.Line,
			Column: ta.ColonColon.Pos.Column,
		})
	}

	return annotatedType
}

// analyzeArrayLiteral analyzes array literals
func (a *Analyzer) analyzeArrayLiteral(al *ast.ArrayLiteral) types.Type {
	if len(al.Elements) == 0 {
		return a.scopes.TypeRegistry().Get("Array")
	}

	// Analyze element types
	elementType := a.analyzeExpression(al.Elements[0][0])
	return types.NewParametricType(
		a.scopes.TypeRegistry().Get("Array").(*types.BasicType),
		elementType,
	)
}

// analyzeTupleLiteral analyzes tuple literals
func (a *Analyzer) analyzeTupleLiteral(tl *ast.TupleLiteral) types.Type {
	elementTypes := make([]types.Type, len(tl.Elements))
	for i, elem := range tl.Elements {
		elementTypes[i] = a.analyzeExpression(elem)
	}
	return types.NewTupleType(elementTypes...)
}

// analyzeDictLiteral analyzes dict literals
func (a *Analyzer) analyzeDictLiteral(dl *ast.DictLiteral) types.Type {
	return a.scopes.TypeRegistry().Get("Dict")
}

// getOperatorName converts a token type to an operator name for dispatch
func (a *Analyzer) getOperatorName(tok lexer.TokenType) string {
	switch tok {
	case lexer.TokenPlus:
		return "+"
	case lexer.TokenMinus:
		return "-"
	case lexer.TokenStar:
		return "*"
	case lexer.TokenSlash:
		return "/"
	case lexer.TokenPercent:
		return "%"
	case lexer.TokenCaret:
		return "^"
	case lexer.TokenEqual:
		return "=="
	case lexer.TokenNotEqual:
		return "!="
	case lexer.TokenLess:
		return "<"
	case lexer.TokenLessEqual:
		return "<="
	case lexer.TokenGreater:
		return ">"
	case lexer.TokenGreaterEqual:
		return ">="
	default:
		return ""
	}
}
