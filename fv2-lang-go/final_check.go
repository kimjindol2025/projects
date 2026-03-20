package main

import (
	"fmt"
	"fv2-lang/internal/ast"
	"fv2-lang/internal/codegen"
)

func main() {
	// 테스트 1: Match 문
	fmt.Println("=== 테스트 1: Match 문 ===")
	matchStmt := &ast.MatchStatement{
		Expression: &ast.Identifier{Name: "x"},
		Arms: []ast.MatchArm{
			{Pattern: &ast.LiteralPattern{Value: &ast.IntegerLiteral{Value: 1}}, Body: []ast.Statement{}},
			{Pattern: &ast.LiteralPattern{Value: &ast.IntegerLiteral{Value: 2}}, Body: []ast.Statement{}},
		},
	}
	prog1 := &ast.Program{Definitions: []ast.Definition{}, MainBody: []ast.Statement{matchStmt}}
	gen := codegen.New()
	code1, _ := gen.Generate(prog1)
	fmt.Println(code1)

	// 테스트 2: 배열 리터럴 for-in
	fmt.Println("\n=== 테스트 2: 배열 리터럴 for-in (크기 3) ===")
	forStmt := &ast.ForStatement{
		Variable: "i",
		Iterator: &ast.ArrayExpression{
			Elements: []ast.Expression{
				&ast.IntegerLiteral{Value: 1},
				&ast.IntegerLiteral{Value: 2},
				&ast.IntegerLiteral{Value: 3},
			},
		},
		Body: []ast.Statement{},
	}
	prog2 := &ast.Program{Definitions: []ast.Definition{}, MainBody: []ast.Statement{forStmt}}
	gen = codegen.New()
	code2, _ := gen.Generate(prog2)
	fmt.Println(code2)

	// 테스트 3: 타입 추론
	fmt.Println("\n=== 테스트 3: 타입 추론 (명시 없음) ===")
	letStmt := &ast.LetStatement{
		Name: "x",
		Init: &ast.IntegerLiteral{Value: 42},
		Type: nil,
	}
	prog3 := &ast.Program{Definitions: []ast.Definition{}, MainBody: []ast.Statement{letStmt}}
	gen = codegen.New()
	code3, _ := gen.Generate(prog3)
	fmt.Println(code3)
}
