package main

import (
	"fmt"
	"fv2-lang/internal/ast"
	"fv2-lang/internal/codegen"
)

func main() {
	// for-in 루프 (배열 리터럴 사용)
	forLoopStmt := &ast.ForStatement{
		Variable: "elem",
		Iterator: &ast.ArrayExpression{
			Elements: []ast.Expression{
				&ast.IntegerLiteral{Value: 1},
				&ast.IntegerLiteral{Value: 2},
				&ast.IntegerLiteral{Value: 3},
			},
		},
		Body: []ast.Statement{
			&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "elem"}},
		},
	}

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			forLoopStmt,
		},
	}

	gen := codegen.New()
	code, err := gen.Generate(program)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
		return
	}

	fmt.Println("=== 배열 리터럴 for-in 루프 (크기 알려짐) ===")
	fmt.Println(code)
}
