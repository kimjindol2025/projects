package main

import (
	"fmt"
	"fv2-lang/internal/ast"
	"fv2-lang/internal/codegen"
)

func main() {
	prog := &ast.Program{
		Definitions: []ast.Definition{
			&ast.ImportStatement{Module: "math"},
			&ast.FunctionDef{
				Name: "main",
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.IntegerLiteral{Value: 42},
					},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	gen := codegen.New()
	code, _ := gen.Generate(prog)
	fmt.Println(code)
}
