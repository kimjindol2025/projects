package main

import (
	"fmt"
	"fv2-lang/internal/ast"
	"fv2-lang/internal/codegen"
)

func main() {
	// for-in 루프 AST 생성
	forLoopStmt := &ast.ForStatement{
		Variable: "i",
		Iterator: &ast.Identifier{Name: "arr"},
		Body: []ast.Statement{
			&ast.ExpressionStatement{Expression: &ast.Identifier{Name: "i"}},
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

	fmt.Println("=== For-in 루프 생성된 C 코드 ===")
	fmt.Println(code)
}
