package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"fv2-lang/internal/lexer"
	"fv2-lang/internal/parser"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: fv2 [options] <file.fv>\n")
		flag.PrintDefaults()
	}

	tokenize := flag.Bool("tokenize", false, "Show tokens only")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "os.Args: %v\n", os.Args)
		fmt.Fprintf(os.Stderr, "flag.Args(): %v\n", flag.Args())
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	args := flag.Args()

	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "args[0]: %q, os.Args[0]: %q, match: %v\n", args[0], os.Args[0], args[0] == os.Args[0])
	}

	// Skip first arg if it's the program itself (Termux quirk)
	startIdx := 0
	if len(args) > 0 && (args[0] == os.Args[0] || strings.HasSuffix(args[0], "/bin/fv2")) {
		startIdx = 1
	}

	if len(args)-startIdx <= 0 {
		fmt.Fprintf(os.Stderr, "Error: No input file specified\n")
		flag.Usage()
		os.Exit(1)
	}

	filename := args[startIdx]

	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "Opening file: %q (%d bytes in arg)\n", filename, len(filename))
	}

	// Read source file
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", filename, err)
		os.Exit(1)
	}

	// Compile
	sourceStr := string(source)

	// Debug: Check for null bytes
	if os.Getenv("DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "File size: %d bytes\n", len(source))
		for i, b := range source {
			if b == 0 {
				fmt.Fprintf(os.Stderr, "Warning: NULL byte at position %d\n", i)
			}
		}
	}
	if err := compile(sourceStr, *tokenize); err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}
}

func compile(source string, tokensOnly bool) error {
	// Step 1: Lexing
	lex, err := lexer.New(source)
	if err != nil {
		return fmt.Errorf("lexer initialization failed: %w", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		return fmt.Errorf("tokenization failed: %w", err)
	}

	if tokensOnly {
		fmt.Printf("=== Tokens ===\n")
		for _, token := range tokens {
			fmt.Printf("%v\n", token)
		}
		return nil
	}

	// Step 2: Parser
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}

	// Step 3: Type Checker (TODO)
	// Step 4: Code Generator (TODO)

	fmt.Printf("// FV 2.0 Compiler\n")
	fmt.Printf("// Tokenized %d tokens\n", len(tokens))
	fmt.Printf("// Parsed: %d definitions, %d statements in main\n", len(program.Definitions), len(program.MainBody))
	fmt.Printf("// Type checking: NOT YET IMPLEMENTED\n")
	fmt.Printf("// C code generation: NOT YET IMPLEMENTED\n")

	return nil
}
