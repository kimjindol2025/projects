package main

import (
	"fmt"
	"io/ioutil"
	"fv2-lang/internal/lexer"
)

func main() {
	source, _ := ioutil.ReadFile("examples/hello.fv")
	fmt.Printf("File size: %d bytes\n", len(source))
	
	lex, err := lexer.New(string(source))
	if err != nil {
		fmt.Printf("Lexer error: %v\n", err)
		return
	}
	
	tokens, err := lex.Tokenize()
	if err != nil {
		fmt.Printf("Tokenize error: %v\n", err)
		return
	}
	
	fmt.Printf("Tokens: %d\n", len(tokens))
	for i, tok := range tokens[:10] {
		fmt.Printf("%d: %v\n", i, tok)
	}
}
