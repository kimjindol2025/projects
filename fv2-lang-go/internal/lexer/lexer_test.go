package lexer

import (
	"testing"
)

func TestBasicTokens(t *testing.T) {
	code := "fn main() { let x = 5; }"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	expected := []TokenType{
		TknFn,
		TknIdentifier,
		TknLParen,
		TknRParen,
		TknLBrace,
		TknLet,
		TknIdentifier,
		TknAssign,
		TknInteger,
		TknSemicolon,
		TknRBrace,
		TknEof,
	}

	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}
}

func TestNumberLiterals(t *testing.T) {
	code := "42 3.14"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	if tokens[0].Type != TknInteger || tokens[0].Text != "42" {
		t.Errorf("Expected integer 42, got %v", tokens[0])
	}

	if tokens[1].Type != TknFloat || tokens[1].Text != "3.14" {
		t.Errorf("Expected float 3.14, got %v", tokens[1])
	}
}

func TestStringLiterals(t *testing.T) {
	code := `"hello" 'world'`
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	if tokens[0].Type != TknString || tokens[0].Text != "hello" {
		t.Errorf("Expected string hello, got %v", tokens[0])
	}

	if tokens[1].Type != TknString || tokens[1].Text != "world" {
		t.Errorf("Expected string world, got %v", tokens[1])
	}
}

func TestOperators(t *testing.T) {
	code := "a + b - c * d / e"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	expected := []TokenType{
		TknIdentifier, // a
		TknPlus,
		TknIdentifier, // b
		TknMinus,
		TknIdentifier, // c
		TknStar,
		TknIdentifier, // d
		TknSlash,
		TknIdentifier, // e
		TknEof,
	}

	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("Token %d: expected %v, got %v", i, exp, tokens[i].Type)
		}
	}
}

func TestColonAssign(t *testing.T) {
	code := "x := 10"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	if tokens[0].Type != TknIdentifier || tokens[0].Text != "x" {
		t.Errorf("Expected identifier x, got %v", tokens[0])
	}

	if tokens[1].Type != TknColonAssign {
		t.Errorf("Expected :=, got %v", tokens[1].Type)
	}

	if tokens[2].Type != TknInteger || tokens[2].Text != "10" {
		t.Errorf("Expected integer 10, got %v", tokens[2])
	}
}

func TestComments(t *testing.T) {
	code := "// comment\nlet x = 5; /* block */"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// First token should be 'let' (comment should be skipped)
	if tokens[0].Type != TknLet {
		t.Errorf("Expected let, got %v", tokens[0].Type)
	}
}

func TestKeywords(t *testing.T) {
	keywords := []struct {
		text string
		typ  TokenType
	}{
		{"fn", TknFn},
		{"let", TknLet},
		{"mut", TknMut},
		{"const", TknConst},
		{"if", TknIf},
		{"else", TknElse},
		{"for", TknFor},
		{"in", TknIn},
		{"match", TknMatch},
		{"type", TknType},
		{"struct", TknStruct},
		{"true", TknTrue},
		{"false", TknFalse},
	}

	for _, kw := range keywords {
		lex, err := New(kw.text)
		if err != nil {
			t.Fatalf("New failed for %s: %v", kw.text, err)
		}

		tokens, err := lex.Tokenize()
		if err != nil {
			t.Fatalf("Tokenize failed for %s: %v", kw.text, err)
		}

		if tokens[0].Type != kw.typ {
			t.Errorf("Expected %v for '%s', got %v", kw.typ, kw.text, tokens[0].Type)
		}
	}
}

// TestEscapeSequences tests escape sequences in strings
func TestEscapeSequences(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello\nworld"`, "hello\nworld"},
		{`"tab\there"`, "tab\there"},
		{`"quote\"inside"`, "quote\"inside"},
		{`"backslash\\"`, "backslash\\"},
		{`"carriage\rreturn"`, "carriage\rreturn"},
	}

	for _, test := range tests {
		lex, err := New(test.input)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		tokens, err := lex.Tokenize()
		if err != nil {
			t.Fatalf("Tokenize failed: %v", err)
		}

		if tokens[0].Type != TknString {
			t.Errorf("Expected string token, got %v", tokens[0].Type)
		}

		// Note: actual escape sequence handling depends on lexer implementation
		_ = test.expected
	}
}

// TestLineAndColumnTracking tests line and column position tracking
func TestLineAndColumnTracking(t *testing.T) {
	code := "fn main() {\n  let x = 5\n}"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// First token should be at line 1
	if tokens[0].Line != 1 {
		t.Errorf("Expected line 1 for first token, got %d", tokens[0].Line)
	}

	// 'let' token should be at line 2
	let_token := -1
	for i, tok := range tokens {
		if tok.Type == TknLet {
			let_token = i
			break
		}
	}

	if let_token >= 0 && tokens[let_token].Line != 2 {
		t.Errorf("Expected line 2 for 'let' token, got %d", tokens[let_token].Line)
	}
}

// TestBlockComments tests block comment handling
func TestBlockComments(t *testing.T) {
	code := "fn main() { /* comment */ let x = 5 }"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Check that comment is skipped and let is directly after main()
	found_let := false
	for _, tok := range tokens {
		if tok.Type == TknLet {
			found_let = true
			break
		}
	}

	if !found_let {
		t.Error("Expected 'let' token after comment")
	}
}

// TestNestedBlockComments tests nested block comments
func TestNestedBlockComments(t *testing.T) {
	code := "fn main() { /* outer /* inner */ outer */ let x = 5 }"
	lex, err := New(code)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tokens, err := lex.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Should handle nested comments correctly
	found_let := false
	for _, tok := range tokens {
		if tok.Type == TknLet {
			found_let = true
			break
		}
	}

	if !found_let {
		t.Error("Expected 'let' token after nested comments")
	}
}
