// Package codegen implements code generation from AST to C
package codegen

import (
	"fmt"
	"strings"

	"fv2-lang/internal/ast"
)

// Generator generates C code from an AST
type Generator struct {
	code          strings.Builder
	indent        int
	VarCounter    int
	functionStack []string // 현재 함수 스택
}

// New creates a new code generator
func New() *Generator {
	return &Generator{
		VarCounter:    0,
		functionStack: []string{},
	}
}

// Generate generates C code from a program AST
func (g *Generator) Generate(program *ast.Program) (string, error) {
	// Header files
	g.writeLine("#include <stdio.h>")
	g.writeLine("#include <stdlib.h>")
	g.writeLine("#include <string.h>")
	g.writeLine("#include <stdbool.h>")

	// Process imports and write additional headers
	for _, def := range program.Definitions {
		if imp, ok := def.(*ast.ImportStatement); ok {
			switch imp.Module {
			case "math":
				g.writeLine("#include <math.h>")
			case "stdio":
				// Already included
			case "string":
				// string functions in string.h already included
			case "stdlib":
				// stdlib.h already included
			}
		}
	}
	g.writeLine("")

	// Forward declarations for functions (except main) and extern functions
	for _, def := range program.Definitions {
		switch d := def.(type) {
		case *ast.FunctionDef:
			if d.Name != "main" {
				g.writeFunctionDeclaration(d)
			}
		case *ast.ExternDef:
			g.writeExternDeclaration(d)
		}
	}
	g.writeLine("")

	// Type definitions (structs, etc)
	for _, def := range program.Definitions {
		switch d := def.(type) {
		case *ast.StructDef:
			g.writeStructDefinition(d)
		case *ast.TypeDef:
			g.writeTypeDefinition(d)
		}
	}
	g.writeLine("")

	// Function implementations
	var mainFunc *ast.FunctionDef
	for _, def := range program.Definitions {
		if fn, ok := def.(*ast.FunctionDef); ok {
			if fn.Name == "main" {
				mainFunc = fn
			} else {
				g.writeFunctionDefinition(fn)
				g.writeLine("")
			}
		}
	}

	// Main function (either from definition or generated)
	g.writeLine("int main() {")
	g.indent++

	if mainFunc != nil {
		// Use main function body from definition
		for _, stmt := range mainFunc.Body {
			g.generateStatement(stmt)
		}
	} else {
		// Use main body statements
		for _, stmt := range program.MainBody {
			g.generateStatement(stmt)
		}
	}

	g.writeLine("return 0;")
	g.indent--
	g.writeLine("}")

	return g.code.String(), nil
}

// writeFunctionDeclaration writes forward declaration
func (g *Generator) writeFunctionDeclaration(fn *ast.FunctionDef) {
	params := g.generateParameterList(fn.Parameters)
	returnType := g.generateType(fn.ReturnType)
	g.writeLine(fmt.Sprintf("%s %s(%s);", returnType, fn.Name, params))
}

// writeExternDeclaration writes extern function declaration
func (g *Generator) writeExternDeclaration(ext *ast.ExternDef) {
	params := g.generateParameterList(ext.Parameters)
	returnType := g.generateType(ext.ReturnType)
	g.writeLine(fmt.Sprintf("extern %s %s(%s);", returnType, ext.Name, params))
}

// writeStructDefinition writes struct definition
func (g *Generator) writeStructDefinition(str *ast.StructDef) {
	g.writeLine(fmt.Sprintf("struct %s {", str.Name))
	g.indent++

	for _, field := range str.Fields {
		fieldType := g.generateType(field.Type)
		g.writeLine(fmt.Sprintf("%s %s;", fieldType, field.Name))
	}

	g.indent--
	g.writeLine("};")
}

// writeTypeDefinition writes type alias
func (g *Generator) writeTypeDefinition(td *ast.TypeDef) {
	// For now, just skip type aliases (they're compile-time constructs)
	_ = td
}

// writeFunctionDefinition writes complete function
func (g *Generator) writeFunctionDefinition(fn *ast.FunctionDef) {
	g.functionStack = append(g.functionStack, fn.Name)

	params := g.generateParameterList(fn.Parameters)
	returnType := g.generateType(fn.ReturnType)

	g.writeLine(fmt.Sprintf("%s %s(%s) {", returnType, fn.Name, params))
	g.indent++

	for _, stmt := range fn.Body {
		g.generateStatement(stmt)
	}

	// Add return statement if not present
	if len(fn.Body) == 0 || !g.endsWithReturn(fn.Body) {
		if returnType == "void" {
			g.writeLine("return;")
		} else {
			g.writeLine(fmt.Sprintf("return 0; // %s", returnType))
		}
	}

	g.indent--
	g.writeLine("}")

	g.functionStack = g.functionStack[:len(g.functionStack)-1]
}

// generateParameterList generates parameter list for C
func (g *Generator) generateParameterList(params []ast.Parameter) string {
	if len(params) == 0 {
		return "void"
	}

	var parts []string
	for _, param := range params {
		paramType := g.generateType(param.Type)
		parts = append(parts, fmt.Sprintf("%s %s", paramType, param.Name))
	}

	return strings.Join(parts, ", ")
}

// generateType converts AST type to C type
func (g *Generator) generateType(t *ast.Type) string {
	if t == nil {
		return "void"
	}

	if t.IsArray {
		elemType := g.generateType(t.ElementType)
		return fmt.Sprintf("%s*", elemType) // Array as pointer
	}

	switch t.Name {
	case "i64":
		return "long long"
	case "f64":
		return "double"
	case "string":
		return "char*"
	case "bool":
		return "bool"
	case "none":
		return "void"
	default:
		return t.Name // struct name or custom type
	}
}

// generateStatement generates C code for a statement
func (g *Generator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		g.generateLetStatement(s)
	case *ast.ConstStatement:
		g.generateConstStatement(s)
	case *ast.IfStatement:
		g.generateIfStatement(s)
	case *ast.ForStatement:
		g.generateForStatement(s)
	case *ast.ForRangeStatement:
		g.generateForRangeStatement(s)
	case *ast.ReturnStatement:
		g.generateReturnStatement(s)
	case *ast.ExpressionStatement:
		g.generateExpressionStatement(s)
	case *ast.MatchStatement:
		g.generateMatchStatement(s)
	case *ast.BlockStatement:
		g.generateBlockStatement(s)
	}
}

// generateLetStatement generates let binding
func (g *Generator) generateLetStatement(let *ast.LetStatement) {
	varType := ""
	if let.Type != nil {
		varType = g.generateType(let.Type)
	} else {
		// Infer type from initial value
		varType = g.inferTypeFromExpression(let.Init)
	}

	initValue := g.generateExpression(let.Init)
	g.writeLine(fmt.Sprintf("%s %s = %s;", varType, let.Name, initValue))
}

// generateConstStatement generates const binding
func (g *Generator) generateConstStatement(const_ *ast.ConstStatement) {
	varType := "const int" // C const
	if const_.Type != nil {
		varType = fmt.Sprintf("const %s", g.generateType(const_.Type))
	}

	initValue := g.generateExpression(const_.Value)
	g.writeLine(fmt.Sprintf("%s %s = %s;", varType, const_.Name, initValue))
}

// generateIfStatement generates if-else statement
func (g *Generator) generateIfStatement(ifStmt *ast.IfStatement) {
	cond := g.generateExpression(ifStmt.Condition)
	g.writeLine(fmt.Sprintf("if (%s) {", cond))
	g.indent++

	for _, stmt := range ifStmt.ThenBody {
		g.generateStatement(stmt)
	}

	g.indent--

	if len(ifStmt.ElseBody) > 0 {
		g.writeLine("} else {")
		g.indent++

		for _, stmt := range ifStmt.ElseBody {
			g.generateStatement(stmt)
		}

		g.indent--
	}

	g.writeLine("}")
}

// generateForStatement generates for loop
func (g *Generator) generateForStatement(forStmt *ast.ForStatement) {
	// for i in iterator -> convert to C for loop
	// Handle array literals with known size vs. variables
	var loopCondition string

	if arrExpr, ok := forStmt.Iterator.(*ast.ArrayExpression); ok {
		// Array literal with known size at compile time
		loopCondition = fmt.Sprintf("_i < %d", len(arrExpr.Elements))
	} else {
		// Iterator is a variable - length must be provided at runtime
		iterator := g.generateExpression(forStmt.Iterator)
		loopCondition = fmt.Sprintf("_i < sizeof(%s)/sizeof(*%s)", iterator, iterator)
		g.writeLine(fmt.Sprintf("// for %s in %s (array length calculation)", forStmt.Variable, iterator))
	}

	g.writeLine(fmt.Sprintf("for (int _i = 0; %s; _i++) {", loopCondition))
	g.indent++

	// Declare loop variable as array index
	g.writeLine(fmt.Sprintf("int %s = _i;", forStmt.Variable))

	// Loop body
	for _, stmt := range forStmt.Body {
		g.generateStatement(stmt)
	}

	g.indent--
	g.writeLine("}")
}

// generateForRangeStatement generates for-range loop
func (g *Generator) generateForRangeStatement(forStmt *ast.ForRangeStatement) {
	start := g.generateExpression(forStmt.Start)
	end := g.generateExpression(forStmt.End)

	g.writeLine(fmt.Sprintf("for (long long %s = %s; %s < %s; %s++) {",
		forStmt.Variable, start, forStmt.Variable, end, forStmt.Variable))
	g.indent++

	for _, stmt := range forStmt.Body {
		g.generateStatement(stmt)
	}

	g.indent--
	g.writeLine("}")
}

// generateReturnStatement generates return statement
func (g *Generator) generateReturnStatement(ret *ast.ReturnStatement) {
	if ret.Value != nil {
		value := g.generateExpression(ret.Value)
		g.writeLine(fmt.Sprintf("return %s;", value))
	} else {
		g.writeLine("return;")
	}
}

// generateExpressionStatement generates expression statement
func (g *Generator) generateExpressionStatement(expr *ast.ExpressionStatement) {
	code := g.generateExpression(expr.Expression)
	g.writeLine(fmt.Sprintf("%s;", code))
}

// generateMatchStatement generates match statement (as if-else chain)
func (g *Generator) generateMatchStatement(match *ast.MatchStatement) {
	// Convert match to if-else chain
	expr := g.generateExpression(match.Expression)

	for i, arm := range match.Arms {
		// Generate pattern matching condition based on pattern type
		var condition string

		if arm.Pattern == nil {
			// No pattern, always true
			condition = "1"
		} else {
			// Type assert pattern to get correct type
			switch p := arm.Pattern.(type) {
			case *ast.LiteralPattern:
				// Literal pattern: compare with expression
				if p.Value != nil {
					patternValue := g.generateExpression(p.Value)
					condition = fmt.Sprintf("%s == %s", expr, patternValue)
				} else {
					condition = "1"
				}
			case *ast.IdentifierPattern:
				// Identifier pattern: variable binding
				// For now, treat as wildcard (would bind in full impl)
				condition = "1"
			case *ast.WildcardPattern:
				// Wildcard: always matches (default case)
				condition = "1"
			default:
				// Unknown pattern, treat as wildcard
				condition = "1"
			}
		}

		// Generate if/else if branch
		if i == 0 {
			g.writeLine(fmt.Sprintf("if (%s) {", condition))
		} else {
			g.writeLine(fmt.Sprintf("} else if (%s) {", condition))
		}
		g.indent++

		for _, stmt := range arm.Body {
			g.generateStatement(stmt)
		}

		g.indent--
	}

	if len(match.Arms) > 0 {
		g.writeLine("}")
	}
}

// generateBlockStatement generates block statement
func (g *Generator) generateBlockStatement(block *ast.BlockStatement) {
	g.writeLine("{")
	g.indent++

	for _, stmt := range block.Statements {
		g.generateStatement(stmt)
	}

	g.indent--
	g.writeLine("}")
}

// generateExpression generates C code for expression
func (g *Generator) generateExpression(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", e.Value)
	case *ast.FloatLiteral:
		return fmt.Sprintf("%g", e.Value)
	case *ast.StringLiteral:
		return fmt.Sprintf("\"%s\"", escapeString(e.Value))
	case *ast.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *ast.NoneLiteral:
		return "NULL"
	case *ast.Identifier:
		return e.Name
	case *ast.BinaryExpression:
		return g.generateBinaryExpression(e)
	case *ast.UnaryExpression:
		return g.generateUnaryExpression(e)
	case *ast.CallExpression:
		return g.generateCallExpression(e)
	case *ast.ArrayExpression:
		return g.generateArrayExpression(e)
	case *ast.FieldExpression:
		return g.generateFieldExpression(e)
	case *ast.IndexExpression:
		return g.generateIndexExpression(e)
	case *ast.IfExpression:
		return g.generateIfExpression(e)
	}

	return "0"
}

// generateBinaryExpression generates binary operation
func (g *Generator) generateBinaryExpression(bin *ast.BinaryExpression) string {
	left := g.generateExpression(bin.Left)
	right := g.generateExpression(bin.Right)

	switch bin.Operator {
	case "..":
		// Range operator - not directly used in C expressions
		return fmt.Sprintf("range(%s, %s)", left, right)
	default:
		return fmt.Sprintf("(%s %s %s)", left, bin.Operator, right)
	}
}

// generateUnaryExpression generates unary operation
func (g *Generator) generateUnaryExpression(unary *ast.UnaryExpression) string {
	operand := g.generateExpression(unary.Operand)
	return fmt.Sprintf("(%s%s)", unary.Operator, operand)
}

// generateCallExpression generates function call
func (g *Generator) generateCallExpression(call *ast.CallExpression) string {
	fn := g.generateExpression(call.Function)

	var args []string
	for _, arg := range call.Arguments {
		args = append(args, g.generateExpression(arg))
	}

	// Handle builtin functions
	if fnIdent, ok := call.Function.(*ast.Identifier); ok {
		switch fnIdent.Name {
		case "println":
			if len(args) == 0 {
				return `printf("\n")`
			}
			// println can accept i64, f64, string, bool, etc.
			// Try to detect the type from the argument expression
			return g.generatePrintln(call.Arguments[0])
		case "print":
			if len(args) == 0 {
				return `printf("")`
			}
			// print can accept i64, f64, string, bool, etc.
			return g.generatePrint(call.Arguments[0])
		case "len":
			if len(args) == 1 {
				// len(array) → simplified as array length check
				// For arrays: would need runtime length info
				return fmt.Sprintf(`(sizeof(%s)/sizeof(*%s))`, args[0], args[0])
			}
		}
	}

	return fmt.Sprintf("%s(%s)", fn, strings.Join(args, ", "))
}

// generatePrintln generates a type-aware println call
func (g *Generator) generatePrintln(arg ast.Expression) string {
	argStr := g.generateExpression(arg)

	// Detect type from AST expression
	switch arg.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf(`printf("%%lld\n", %s)`, argStr)
	case *ast.FloatLiteral:
		return fmt.Sprintf(`printf("%%f\n", %s)`, argStr)
	case *ast.StringLiteral:
		return fmt.Sprintf(`printf("%%s\n", %s)`, argStr)
	case *ast.BoolLiteral:
		return fmt.Sprintf(`printf("%%s\n", %s ? "true" : "false")`, argStr)
	case *ast.Identifier:
		// For identifiers, we don't know the type, so assume string
		// In a full implementation, we'd use TypeChecker info
		return fmt.Sprintf(`printf("%%s\n", %s)`, argStr)
	default:
		// Default to treating as i64
		return fmt.Sprintf(`printf("%%lld\n", %s)`, argStr)
	}
}

// generatePrint generates a type-aware print call
func (g *Generator) generatePrint(arg ast.Expression) string {
	argStr := g.generateExpression(arg)

	// Detect type from AST expression
	switch arg.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf(`printf("%%lld", %s)`, argStr)
	case *ast.FloatLiteral:
		return fmt.Sprintf(`printf("%%f", %s)`, argStr)
	case *ast.StringLiteral:
		return fmt.Sprintf(`printf("%%s", %s)`, argStr)
	case *ast.BoolLiteral:
		return fmt.Sprintf(`printf("%%s", %s ? "true" : "false")`, argStr)
	case *ast.Identifier:
		// For identifiers, we don't know the type, so assume string
		return fmt.Sprintf(`printf("%%s", %s)`, argStr)
	default:
		// Default to treating as i64
		return fmt.Sprintf(`printf("%%lld", %s)`, argStr)
	}
}

// generateArrayExpression generates array literal
func (g *Generator) generateArrayExpression(arr *ast.ArrayExpression) string {
	// Arrays in C: create as pointer allocation
	if len(arr.Elements) == 0 {
		return "NULL"
	}

	// For now, return a simple representation
	var elems []string
	for _, elem := range arr.Elements {
		elems = append(elems, g.generateExpression(elem))
	}

	return fmt.Sprintf("{%s}", strings.Join(elems, ", "))
}

// generateFieldExpression generates struct field access
func (g *Generator) generateFieldExpression(field *ast.FieldExpression) string {
	obj := g.generateExpression(field.Object)
	return fmt.Sprintf("%s.%s", obj, field.Field)
}

// generateIndexExpression generates array indexing
func (g *Generator) generateIndexExpression(index *ast.IndexExpression) string {
	obj := g.generateExpression(index.Object)
	idx := g.generateExpression(index.Index)
	return fmt.Sprintf("%s[%s]", obj, idx)
}

// generateIfExpression generates if as expression
func (g *Generator) generateIfExpression(ifExpr *ast.IfExpression) string {
	cond := g.generateExpression(ifExpr.Condition)
	then := g.generateExpression(ifExpr.ThenExpr)

	if ifExpr.ElseExpr != nil {
		else_ := g.generateExpression(ifExpr.ElseExpr)
		return fmt.Sprintf("(%s ? %s : %s)", cond, then, else_)
	}

	return fmt.Sprintf("(%s ? %s : 0)", cond, then)
}

// Helper functions

// writeLine writes a line with proper indentation
func (g *Generator) writeLine(line string) {
	indent := strings.Repeat("  ", g.indent)
	g.code.WriteString(indent + line + "\n")
}

// endsWithReturn checks if statements end with return
func (g *Generator) endsWithReturn(stmts []ast.Statement) bool {
	if len(stmts) == 0 {
		return false
	}

	_, ok := stmts[len(stmts)-1].(*ast.ReturnStatement)
	return ok
}

// inferTypeFromExpression infers C type from AST expression
func (g *Generator) inferTypeFromExpression(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return "long long"
	case *ast.FloatLiteral:
		return "double"
	case *ast.StringLiteral:
		return "char*"
	case *ast.BoolLiteral:
		return "bool"
	case *ast.NoneLiteral:
		return "void*"
	case *ast.ArrayExpression:
		if len(e.Elements) > 0 {
			elemType := g.inferTypeFromExpression(e.Elements[0])
			return fmt.Sprintf("%s*", elemType)
		}
		return "void*"
	default:
		// Default to long long for unknown types
		return "long long"
	}
}

// escapeString escapes special characters in strings
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
