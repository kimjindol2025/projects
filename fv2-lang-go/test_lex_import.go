package main

import (
	"fmt"
	"fv2-lang/internal/lexer"
)

func main() {
	input := `import "math"`
	
	l, _ := lexer.New(input)
	tokens, _ := l.Tokenize()
	
	for i, t := range tokens[:3] {
		fmt.Printf("%d: Type=%d Text=%q\n", i, t.Type, t.Text)
	}
}
