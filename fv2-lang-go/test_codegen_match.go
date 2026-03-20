package main

import (
	"fmt"
	"fv2-lang/internal/ast"
	"fv2-lang/internal/codegen"
)

func main() {
	// Match statement AST 생성
	matchStmt := &ast.MatchStatement{
		Expression: &ast.Identifier{Name: "x"},
		Arms: []ast.MatchArm{
			{
				Pattern: &ast.LiteralPattern{Value: &ast.IntegerLiteral{Value: 1}},
				Body: []ast.Statement{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "y"}},
				},
			},
			{
				Pattern: &ast.LiteralPattern{Value: &ast.IntegerLiteral{Value: 2}},
				Body: []ast.Statement{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "z"}},
				},
			},
			{
				Pattern: &ast.WildcardPattern{},
				Body: []ast.Statement{
					&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "other"}},
				},
			},
		},
	}

	// Program 생성
	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			matchStmt,
		},
	}

	// 코드 생성
	gen := codegen.New()
	code, err := gen.Generate(program)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
		return
	}

	fmt.Println("=== 생성된 C 코드 ===")
	fmt.Println(code)
}
