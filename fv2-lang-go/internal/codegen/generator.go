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
	g.writeLine("")

	// Forward declarations for functions
	for _, def := range program.Definitions {
		if fn, ok := def.(*ast.FunctionDef); ok {
			g.writeFunctionDeclaration(fn)
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
	for _, def := range program.Definitions {
		if fn, ok := def.(*ast.FunctionDef); ok {
			g.writeFunctionDefinition(fn)
			g.writeLine("")
		}
	}

	// Main function
	g.writeLine("int main() {")
	g.indent++

	for _, stmt := range program.MainBody {
		g.generateStatement(stmt)
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
	varType := "auto" // C11 auto-type inference
	if let.Type != nil {
		varType = g.generateType(let.Type)
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
	// Use a simple counter-based loop for arrays
	iterator := g.generateExpression(forStmt.Iterator)

	g.writeLine(fmt.Sprintf("// for %s in %s", forStmt.Variable, iterator))
	g.writeLine(fmt.Sprintf("for (int _i = 0; _i < sizeof(%s)/sizeof(%s[0]); _i++) {", iterator, iterator))
	g.indent++

	// Declare loop variable
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
	_ = g.generateExpression(match.Expression)

	for i, arm := range match.Arms {
		// Generate pattern matching condition
		// Pattern is an interface, so we always use default condition
		condition := "1" // default: always true (simplified implementation)

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

	return fmt.Sprintf("%s(%s)", fn, strings.Join(args, ", "))
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

// escapeString escapes special characters in strings
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
