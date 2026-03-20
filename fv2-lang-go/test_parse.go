package main

import (
	"fmt"
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
	
	l, _ := lexer.New(input)
	tokens, _ := l.Tokenize()
	fmt.Printf("Tokens: %d\n", len(tokens))
	
	p := parser.New(tokens)
	prog, _ := p.Parse()
	
	fmt.Printf("Definitions: %d\n", len(prog.Definitions))
	for i, d := range prog.Definitions {
		fmt.Printf("%d: %T\n", i, d)
	}
}
