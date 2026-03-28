package lexer

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input  string
		tokens []ast.Token
	}{
		{
			input: "let x = 10",
			tokens: []ast.Token{
				{Type: ast.TokenLet, Value: "let"},
				{Type: ast.TokenIdent, Value: "x"},
				{Type: ast.TokenAssign, Value: "="},
				{Type: ast.TokenInt, Value: "10"},
				{Type: ast.TokenEOF},
			},
		},
		{
			input: "fn add(a, b) { return a + b }",
			tokens: []ast.Token{
				{Type: ast.TokenFn, Value: "fn"},
				{Type: ast.TokenIdent, Value: "add"},
				{Type: ast.TokenLParen, Value: "("},
				{Type: ast.TokenIdent, Value: "a"},
				{Type: ast.TokenComma, Value: ","},
				{Type: ast.TokenIdent, Value: "b"},
				{Type: ast.TokenRParen, Value: ")"},
				{Type: ast.TokenLBrace, Value: "{"},
				{Type: ast.TokenReturn, Value: "return"},
				{Type: ast.TokenIdent, Value: "a"},
				{Type: ast.TokenPlus, Value: "+"},
				{Type: ast.TokenIdent, Value: "b"},
				{Type: ast.TokenRBrace, Value: "}"},
				{Type: ast.TokenEOF},
			},
		},
		{
			input: "if x > 5 { let y = x * 2 }",
			tokens: []ast.Token{
				{Type: ast.TokenIf, Value: "if"},
				{Type: ast.TokenIdent, Value: "x"},
				{Type: ast.TokenGt, Value: ">"},
				{Type: ast.TokenInt, Value: "5"},
				{Type: ast.TokenLBrace, Value: "{"},
				{Type: ast.TokenLet, Value: "let"},
				{Type: ast.TokenIdent, Value: "y"},
				{Type: ast.TokenAssign, Value: "="},
				{Type: ast.TokenIdent, Value: "x"},
				{Type: ast.TokenStar, Value: "*"},
				{Type: ast.TokenInt, Value: "2"},
				{Type: ast.TokenRBrace, Value: "}"},
				{Type: ast.TokenEOF},
			},
		},
		{
			input: "for i in 0..10 { let s = s + i }",
			tokens: []ast.Token{
				{Type: ast.TokenFor, Value: "for"},
				{Type: ast.TokenIdent, Value: "i"},
				{Type: ast.TokenIn, Value: "in"},
				{Type: ast.TokenInt, Value: "0"},
				{Type: ast.TokenDotDot, Value: ".."},
				{Type: ast.TokenInt, Value: "10"},
				{Type: ast.TokenLBrace, Value: "{"},
				{Type: ast.TokenLet, Value: "let"},
				{Type: ast.TokenIdent, Value: "s"},
				{Type: ast.TokenAssign, Value: "="},
				{Type: ast.TokenIdent, Value: "s"},
				{Type: ast.TokenPlus, Value: "+"},
				{Type: ast.TokenIdent, Value: "i"},
				{Type: ast.TokenRBrace, Value: "}"},
				{Type: ast.TokenEOF},
			},
		},
	}

	for _, tt := range tests {
		l := New(tt.input)
		for i, expectedTok := range tt.tokens {
			tok := l.NextToken()
			if tok.Type != expectedTok.Type {
				t.Errorf("Test %d Token %d: expected type %v, got %v", 0, i, expectedTok.Type, tok.Type)
			}
			if tok.Value != expectedTok.Value {
				t.Errorf("Test %d Token %d: expected value %q, got %q", 0, i, expectedTok.Value, tok.Value)
			}
		}
	}
}

func TestOperators(t *testing.T) {
	tests := []struct {
		input  string
		opType ast.TokenType
	}{
		{"==", ast.TokenEq},
		{"!=", ast.TokenNe},
		{"<=", ast.TokenLe},
		{">=", ast.TokenGe},
		{"..", ast.TokenDotDot},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.opType {
			t.Errorf("Input %q: expected %v, got %v", tt.input, tt.opType, tok.Type)
		}
	}
}
