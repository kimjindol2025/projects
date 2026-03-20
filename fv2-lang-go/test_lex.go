package main

import (
	"fmt"
	"fv2-lang/internal/lexer"
)

func main() {
	input := `import "math"

fn main() {
    let x = 3
}`
	
	l, _ := lexer.New(input)
	tokens, _ := l.Tokenize()
	
	for i, t := range tokens {
		fmt.Printf("%d: Type=%d Text=%q\n", i, t.Type, t.Text)
		if i > 10 {
			break
		}
	}
}
