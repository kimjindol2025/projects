package main

import (
	"fmt"
	"fv2-lang/internal/ast"
	"fv2-lang/internal/lexer"
	"fv2-lang/internal/parser"
)

func main() {
	input := `import "math"

extern fn sqrt(x: f64) f64

fn main() {
    let result = sqrt(16.0)
    println(result)
}`
	
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	
	p := parser.New(tokens)
	prog, _ := p.Parse()
	
	fmt.Printf("Definitions: %d\n", len(prog.Definitions))
	for i, d := range prog.Definitions {
		if imp, ok := d.(*ast.ImportStatement); ok {
			fmt.Printf("%d: ImportStatement(Module=%q)\n", i, imp.Module)
		} else {
			fmt.Printf("%d: %T\n", i, d)
		}
	}
}
