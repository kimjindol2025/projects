package main

import (
	"fmt"
	"fv2-lang/internal/ast"
	"fv2-lang/internal/codegen"
)

func main() {
	// 다양한 타입의 let 문
	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "a",
				Init: &ast.IntegerLiteral{Value: 42},
				Type: &ast.Type{Name: "i64"},
			},
			&ast.LetStatement{
				Name: "b",
				Init: &ast.FloatLiteral{Value: 3.14},
				Type: &ast.Type{Name: "f64"},
			},
			&ast.LetStatement{
				Name: "s",
				Init: &ast.StringLiteral{Value: "hello"},
				Type: &ast.Type{Name: "string"},
			},
			// 타입 없이 (auto로 추론)
			&ast.LetStatement{
				Name: "x",
				Init: &ast.IntegerLiteral{Value: 100},
				Type: nil,
			},
		},
	}

	gen := codegen.New()
	code, err := gen.Generate(program)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
		return
	}

	fmt.Println("=== 타입 선언 생성된 C 코드 ===")
	fmt.Println(code)
}
