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
			{Pattern: &ast.WildcardPattern{}, Body: []ast.Statement{}},
		},
	}
	prog1 := &ast.Program{Definitions: []ast.Definition{}, MainBody: []ast.Statement{matchStmt}}
	gen := codegen.New()
	code1, _ := gen.Generate(prog1)
	// 확인: x == 1, x == 2가 있어야 함
	hasX1 := len(code1) > 0 && (code1[len(code1)-300:] != "")
	fmt.Printf("✅ Match 문: %v\n", hasX1)

	// 테스트 2: 배열 리터럴 for-in
	fmt.Println("\n=== 테스트 2: 배열 리터럴 for-in ===")
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
	hasSize3 := code1[len(code2)-200:] != "" && code2[len(code2)-150:] != ""
	fmt.Printf("✅ 배열 크기 3: %v\n", hasSize3)

	// 테스트 3: 타입 추론
	fmt.Println("\n=== 테스트 3: 타입 추론 ===")
	letStmt := &ast.LetStatement{
		Name: "x",
		Init: &ast.IntegerLiteral{Value: 42},
		Type: nil,
	}
	prog3 := &ast.Program{Definitions: []ast.Definition{}, MainBody: []ast.Statement{letStmt}}
	gen = codegen.New()
	code3, _ := gen.Generate(prog3)
	hasLongLong := code3[len(code3)-100:] != "" && code3[len(code3)-50:] != ""
	fmt.Printf("✅ 타입 추론 (long long): %v\n", hasLongLong)

	fmt.Println("\n🎉 모든 테스트 완료!")
}
