package typechecker

import (
	"fmt"
	"fv2-lang/internal/ast"
)

// Checker performs type checking on an AST
type Checker struct {
	GlobalScope *Scope
	CurrentScope *Scope
	Errors      []Error
}

// New creates a new type checker
func New() *Checker {
	globalScope := NewScope(nil)

	// Add built-in types
	globalScope.Define("i64", &PrimitiveType{Name: "i64"}, "type")
	globalScope.Define("f64", &PrimitiveType{Name: "f64"}, "type")
	globalScope.Define("string", &PrimitiveType{Name: "string"}, "type")
	globalScope.Define("bool", &PrimitiveType{Name: "bool"}, "type")
	globalScope.Define("none", &PrimitiveType{Name: "none"}, "type")

	// Add built-in functions
	// Note: println/print are variadic and accept any type
	// For simplicity, we register them as accepting "any" and let codegen handle conversion
	globalScope.Define("println", &BuiltinFunctionType{
		Name: "println",
		IsVariadic: true,
	}, "function")
	globalScope.Define("print", &BuiltinFunctionType{
		Name: "print",
		IsVariadic: true,
	}, "function")
	globalScope.Define("len", &BuiltinFunctionType{
		Name: "len",
		IsVariadic: false,
	}, "function")

	// Math stdlib functions
	globalScope.Define("abs", &BuiltinFunctionType{
		Name: "abs",
		IsVariadic: false,
	}, "function")
	globalScope.Define("min", &BuiltinFunctionType{
		Name: "min",
		IsVariadic: false,
	}, "function")
	globalScope.Define("max", &BuiltinFunctionType{
		Name: "max",
		IsVariadic: false,
	}, "function")

	// Conversion functions
	globalScope.Define("to_string", &BuiltinFunctionType{
		Name: "to_string",
		IsVariadic: false,
	}, "function")
	globalScope.Define("to_int", &BuiltinFunctionType{
		Name: "to_int",
		IsVariadic: false,
	}, "function")
	globalScope.Define("to_float", &BuiltinFunctionType{
		Name: "to_float",
		IsVariadic: false,
	}, "function")

	return &Checker{
		GlobalScope: globalScope,
		CurrentScope: globalScope,
		Errors:      []Error{},
	}
}

// Check performs type checking on the program
func (c *Checker) Check(program *ast.Program) ([]Error, error) {
	// Check definitions first (function, type, struct)
	for _, def := range program.Definitions {
		c.checkDefinition(def)
	}

	// Check main body
	for _, stmt := range program.MainBody {
		c.checkStatement(stmt)
	}

	if len(c.Errors) > 0 {
		return c.Errors, fmt.Errorf("type checking failed with %d errors", len(c.Errors))
	}

	return c.Errors, nil
}

func (c *Checker) checkDefinition(def ast.Definition) {
	switch d := def.(type) {
	case *ast.FunctionDef:
		c.checkFunctionDef(d)
	case *ast.ExternDef:
		c.checkExternDef(d)
	case *ast.StructDef:
		c.checkStructDef(d)
	case *ast.TypeDef:
		c.checkTypeDef(d)
	case *ast.InterfaceDef:
		c.checkInterfaceDef(d)
	case *ast.EnumDef:
		c.checkEnumDef(d)
	}
}

func (c *Checker) checkFunctionDef(fn *ast.FunctionDef) {
	// Create function scope
	fnScope := NewScope(c.CurrentScope)
	prevScope := c.CurrentScope
	c.CurrentScope = fnScope

	// Add parameters to scope
	paramTypes := []Type{}
	for _, param := range fn.Parameters {
		paramType := c.astTypeToCheckerType(param.Type)
		fnScope.Define(param.Name, paramType, "var")
		paramTypes = append(paramTypes, paramType)
	}

	// Check return type
	var returnType Type
	if fn.ReturnType != nil {
		returnType = c.astTypeToCheckerType(fn.ReturnType)
	} else {
		returnType = &PrimitiveType{Name: "none"}
	}

	// Check function body
	for _, stmt := range fn.Body {
		c.checkStatement(stmt)
	}

	// Register function in parent scope
	fnType := &FunctionType{
		ParamTypes: paramTypes,
		ReturnType: returnType,
	}
	prevScope.Define(fn.Name, fnType, "function")

	c.CurrentScope = prevScope
}

func (c *Checker) checkExternDef(ext *ast.ExternDef) {
	// Register extern function in scope
	paramTypes := []Type{}
	for _, param := range ext.Parameters {
		paramType := c.astTypeToCheckerType(param.Type)
		paramTypes = append(paramTypes, paramType)
	}

	var returnType Type = &PrimitiveType{Name: "none"}
	if ext.ReturnType != nil {
		returnType = c.astTypeToCheckerType(ext.ReturnType)
	}

	fnType := &FunctionType{
		ParamTypes: paramTypes,
		ReturnType: returnType,
	}
	c.CurrentScope.Define(ext.Name, fnType, "extern_function")
}

func (c *Checker) checkStructDef(str *ast.StructDef) {
	fields := make(map[string]Type)
	for _, field := range str.Fields {
		fieldType := c.astTypeToCheckerType(field.Type)
		fields[field.Name] = fieldType
	}

	structType := &StructType{
		Name:   str.Name,
		Fields: fields,
	}

	c.CurrentScope.Define(str.Name, structType, "type")
}

func (c *Checker) checkTypeDef(td *ast.TypeDef) {
	c.CurrentScope.Define(td.Name, &PrimitiveType{Name: td.Name}, "type")
}

func (c *Checker) checkInterfaceDef(iface *ast.InterfaceDef) {
	c.CurrentScope.Define(iface.Name, &PrimitiveType{Name: "interface " + iface.Name}, "type")
}

func (c *Checker) checkEnumDef(enum *ast.EnumDef) {
	c.CurrentScope.Define(enum.Name, &PrimitiveType{Name: "enum " + enum.Name}, "type")
}

func (c *Checker) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		c.checkLetStatement(s)
	case *ast.ConstStatement:
		c.checkConstStatement(s)
	case *ast.IfStatement:
		c.checkIfStatement(s)
	case *ast.ForStatement:
		c.checkForStatement(s)
	case *ast.ForRangeStatement:
		c.checkForRangeStatement(s)
	case *ast.ReturnStatement:
		c.checkReturnStatement(s)
	case *ast.MatchStatement:
		c.checkMatchStatement(s)
	case *ast.ExpressionStatement:
		c.checkExpression(s.Expression)
	case *ast.BlockStatement:
		c.checkBlockStatement(s)
	}
}

func (c *Checker) checkLetStatement(let *ast.LetStatement) {
	initType := c.checkExpression(let.Init)

	var declaredType Type
	if let.Type != nil {
		declaredType = c.astTypeToCheckerType(let.Type)

		// Verify init matches declared type
		if !initType.Equal(declaredType) {
			c.addError(0, 0, fmt.Sprintf(
				"let %s: expected type %s, got %s",
				let.Name, declaredType.TypeString(), initType.TypeString(),
			))
		}
	} else {
		declaredType = initType
	}

	c.CurrentScope.Define(let.Name, declaredType, "var")
}

func (c *Checker) checkConstStatement(const_ *ast.ConstStatement) {
	initType := c.checkExpression(const_.Value)

	var declaredType Type
	if const_.Type != nil {
		declaredType = c.astTypeToCheckerType(const_.Type)

		if !initType.Equal(declaredType) {
			c.addError(0, 0, fmt.Sprintf(
				"const %s: expected type %s, got %s",
				const_.Name, declaredType.TypeString(), initType.TypeString(),
			))
		}
	} else {
		declaredType = initType
	}

	c.CurrentScope.Define(const_.Name, declaredType, "const")
}

func (c *Checker) checkIfStatement(ifStmt *ast.IfStatement) {
	condType := c.checkExpression(ifStmt.Condition)

	// Condition must be bool
	if !condType.Equal(&PrimitiveType{Name: "bool"}) {
		c.addError(0, 0, fmt.Sprintf(
			"if condition: expected bool, got %s",
			condType.TypeString(),
		))
	}

	// Check then branch
	for _, stmt := range ifStmt.ThenBody {
		c.checkStatement(stmt)
	}

	// Check else branch (if any)
	for _, stmt := range ifStmt.ElseBody {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkForStatement(forStmt *ast.ForStatement) {
	loopScope := NewScope(c.CurrentScope)
	prevScope := c.CurrentScope
	c.CurrentScope = loopScope

	c.checkExpression(forStmt.Iterator)

	// Check body
	for _, stmt := range forStmt.Body {
		c.checkStatement(stmt)
	}

	c.CurrentScope = prevScope
}

func (c *Checker) checkForRangeStatement(forStmt *ast.ForRangeStatement) {
	// Check start and end are numeric
	startType := c.checkExpression(forStmt.Start)
	endType := c.checkExpression(forStmt.End)

	isNumeric := func(t Type) bool {
		if p, ok := t.(*PrimitiveType); ok {
			return p.Name == "i64" || p.Name == "f64"
		}
		return false
	}

	if !isNumeric(startType) {
		c.addError(0, 0, fmt.Sprintf("range start must be numeric, got %s", startType.TypeString()))
	}
	if !isNumeric(endType) {
		c.addError(0, 0, fmt.Sprintf("range end must be numeric, got %s", endType.TypeString()))
	}

	// Add loop variable to scope
	loopScope := NewScope(c.CurrentScope)
	prevScope := c.CurrentScope
	c.CurrentScope = loopScope

	loopScope.Define(forStmt.Variable, startType, "var")

	// Check body
	for _, stmt := range forStmt.Body {
		c.checkStatement(stmt)
	}

	c.CurrentScope = prevScope
}

func (c *Checker) checkReturnStatement(ret *ast.ReturnStatement) {
	if ret.Value != nil {
		c.checkExpression(ret.Value)
	}
}

func (c *Checker) checkMatchStatement(match *ast.MatchStatement) {
	_ = c.checkExpression(match.Expression)

	for _, arm := range match.Arms {
		for _, stmt := range arm.Body {
			c.checkStatement(stmt)
		}
	}
}

func (c *Checker) checkBlockStatement(block *ast.BlockStatement) {
	blockScope := NewScope(c.CurrentScope)
	prevScope := c.CurrentScope
	c.CurrentScope = blockScope

	for _, stmt := range block.Statements {
		c.checkStatement(stmt)
	}

	c.CurrentScope = prevScope
}

func (c *Checker) checkExpression(expr ast.Expression) Type {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return &PrimitiveType{Name: "i64"}
	case *ast.FloatLiteral:
		return &PrimitiveType{Name: "f64"}
	case *ast.StringLiteral:
		return &PrimitiveType{Name: "string"}
	case *ast.BoolLiteral:
		return &PrimitiveType{Name: "bool"}
	case *ast.NoneLiteral:
		return &PrimitiveType{Name: "none"}
	case *ast.Identifier:
		return c.checkIdentifier(e)
	case *ast.BinaryExpression:
		return c.checkBinaryExpression(e)
	case *ast.UnaryExpression:
		return c.checkUnaryExpression(e)
	case *ast.CallExpression:
		return c.checkCallExpression(e)
	case *ast.ArrayExpression:
		return c.checkArrayExpression(e)
	case *ast.FieldExpression:
		return c.checkFieldExpression(e)
	case *ast.IndexExpression:
		return c.checkIndexExpression(e)
	case *ast.IfExpression:
		return c.checkIfExpression(e)
	case *ast.MatchExpression:
		return c.checkMatchExpression(e)
	}
	return &PrimitiveType{Name: "none"}
}

func (c *Checker) checkIdentifier(ident *ast.Identifier) Type {
	sym := c.CurrentScope.Lookup(ident.Name)
	if sym == nil {
		c.addError(0, 0, fmt.Sprintf("undefined variable: %s", ident.Name))
		return &PrimitiveType{Name: "unknown"}
	}
	return sym.Type
}

func (c *Checker) checkBinaryExpression(bin *ast.BinaryExpression) Type {
	leftType := c.checkExpression(bin.Left)
	rightType := c.checkExpression(bin.Right)

	// Type-check operator
	return c.inferBinaryOpType(leftType, rightType, bin.Operator)
}

func (c *Checker) inferBinaryOpType(left, right Type, op string) Type {
	// Arithmetic operators
	if op == "+" || op == "-" || op == "*" || op == "/" || op == "%" {
		if !c.isNumeric(left) || !c.isNumeric(right) {
			c.addError(0, 0, fmt.Sprintf("arithmetic operator %s requires numeric types", op))
			return left
		}
		if !left.Equal(right) {
			c.addError(0, 0, fmt.Sprintf("type mismatch: %s vs %s", left.TypeString(), right.TypeString()))
		}
		return left
	}

	// Comparison operators
	if op == "==" || op == "!=" || op == "<" || op == ">" || op == "<=" || op == ">=" {
		return &PrimitiveType{Name: "bool"}
	}

	// Logical operators
	if op == "&&" || op == "||" {
		if !left.Equal(&PrimitiveType{Name: "bool"}) {
			c.addError(0, 0, "logical operator requires bool")
		}
		return &PrimitiveType{Name: "bool"}
	}

	return left
}

func (c *Checker) isNumeric(t Type) bool {
	if p, ok := t.(*PrimitiveType); ok {
		return p.Name == "i64" || p.Name == "f64"
	}
	return false
}

func (c *Checker) checkUnaryExpression(unary *ast.UnaryExpression) Type {
	exprType := c.checkExpression(unary.Operand)

	if unary.Operator == "-" {
		if !c.isNumeric(exprType) {
			c.addError(0, 0, "unary minus requires numeric type")
		}
		return exprType
	}

	if unary.Operator == "!" {
		if !exprType.Equal(&PrimitiveType{Name: "bool"}) {
			c.addError(0, 0, "logical not requires bool")
		}
		return &PrimitiveType{Name: "bool"}
	}

	return exprType
}

func (c *Checker) checkCallExpression(call *ast.CallExpression) Type {
	fnType := c.checkExpression(call.Function)

	if ft, ok := fnType.(*FunctionType); ok {
		// Check argument count
		if len(call.Arguments) != len(ft.ParamTypes) {
			c.addError(0, 0, fmt.Sprintf(
				"function expects %d arguments, got %d",
				len(ft.ParamTypes), len(call.Arguments),
			))
		}

		// Check argument types
		for i, arg := range call.Arguments {
			if i < len(ft.ParamTypes) {
				argType := c.checkExpression(arg)
				expectedType := ft.ParamTypes[i]
				if !argType.Equal(expectedType) {
					c.addError(0, 0, fmt.Sprintf(
						"argument %d: expected %s, got %s",
						i, expectedType.TypeString(), argType.TypeString(),
					))
				}
			}
		}

		return ft.ReturnType
	}

	// Handle built-in functions
	if bt, ok := fnType.(*BuiltinFunctionType); ok {
		// Built-in functions like println, print accept any arguments
		if bt.IsVariadic {
			// Just check that arguments are valid expressions
			for _, arg := range call.Arguments {
				c.checkExpression(arg)
			}
			return &PrimitiveType{Name: "none"}
		}
		return &PrimitiveType{Name: "none"}
	}

	c.addError(0, 0, "cannot call non-function")
	return &PrimitiveType{Name: "none"}
}

func (c *Checker) checkArrayExpression(arr *ast.ArrayExpression) Type {
	if len(arr.Elements) == 0 {
		// Empty array - type is ambiguous
		return &ArrayType{ElementType: &PrimitiveType{Name: "unknown"}}
	}

	firstType := c.checkExpression(arr.Elements[0])

	// Check all elements have same type
	for i := 1; i < len(arr.Elements); i++ {
		elemType := c.checkExpression(arr.Elements[i])
		if !elemType.Equal(firstType) {
			c.addError(0, 0, fmt.Sprintf(
				"array element %d: expected %s, got %s",
				i, firstType.TypeString(), elemType.TypeString(),
			))
		}
	}

	return &ArrayType{ElementType: firstType}
}

func (c *Checker) checkFieldExpression(field *ast.FieldExpression) Type {
	exprType := c.checkExpression(field.Object)

	if st, ok := exprType.(*StructType); ok {
		if fieldType, ok := st.Fields[field.Field]; ok {
			return fieldType
		}
		c.addError(0, 0, fmt.Sprintf("struct %s has no field %s", st.Name, field.Field))
		return &PrimitiveType{Name: "unknown"}
	}

	c.addError(0, 0, fmt.Sprintf("cannot access field on %s", exprType.TypeString()))
	return &PrimitiveType{Name: "none"}
}

func (c *Checker) checkIndexExpression(index *ast.IndexExpression) Type {
	exprType := c.checkExpression(index.Object)
	indexType := c.checkExpression(index.Index)

	// Index must be numeric
	if !c.isNumeric(indexType) {
		c.addError(0, 0, "array index must be numeric")
	}

	if at, ok := exprType.(*ArrayType); ok {
		return at.ElementType
	}

	c.addError(0, 0, fmt.Sprintf("cannot index %s", exprType.TypeString()))
	return &PrimitiveType{Name: "none"}
}

func (c *Checker) checkIfExpression(ifExpr *ast.IfExpression) Type {
	condType := c.checkExpression(ifExpr.Condition)

	if !condType.Equal(&PrimitiveType{Name: "bool"}) {
		c.addError(0, 0, "if condition must be bool")
	}

	thenType := c.checkExpression(ifExpr.ThenExpr)

	if ifExpr.ElseExpr != nil {
		elseType := c.checkExpression(ifExpr.ElseExpr)
		if !thenType.Equal(elseType) {
			c.addError(0, 0, fmt.Sprintf(
				"if/else branch types mismatch: %s vs %s",
				thenType.TypeString(), elseType.TypeString(),
			))
		}
		return thenType
	}

	return thenType
}

func (c *Checker) checkMatchExpression(match *ast.MatchExpression) Type {
	_ = c.checkExpression(match.Expression)

	if len(match.Arms) == 0 {
		return &PrimitiveType{Name: "none"}
	}

	var firstType Type
	for i, arm := range match.Arms {
		armType := c.checkExpression(arm.Value)
		if i == 0 {
			firstType = armType
		} else if !armType.Equal(firstType) {
			c.addError(0, 0, fmt.Sprintf(
				"match arm %d: expected %s, got %s",
				i, firstType.TypeString(), armType.TypeString(),
			))
		}
	}

	if firstType == nil {
		return &PrimitiveType{Name: "none"}
	}
	return firstType
}

func (c *Checker) astTypeToCheckerType(t *ast.Type) Type {
	if t == nil {
		return &PrimitiveType{Name: "none"}
	}

	if t.IsArray {
		return &ArrayType{ElementType: c.astTypeToCheckerType(t.ElementType)}
	}

	return &PrimitiveType{Name: t.Name}
}

func (c *Checker) addError(line, column int, msg string) {
	c.Errors = append(c.Errors, Error{
		Line:    line,
		Column:  column,
		Message: msg,
	})
}
