package lexer

import (
	"testing"
)

// TestPhase1Keywords - Phase 1 모든 키워드 테스트
func TestPhase1Keywords(t *testing.T) {
	input := `
	abstract and as begin break catch
	continue const do else elseif
	end export false final finally
	for function global if import
	in isa let local macro
	module mutable not nothing or
	quote return struct true try
	using while
	`
	lexer := NewLexer(input)
	
	expectedKeywords := []TokenType{
		TokenKeywordAbstract, TokenKeywordAnd, TokenKeywordAs, TokenKeywordBegin,
		TokenKeywordBreak, TokenKeywordCatch, TokenKeywordContinue, TokenKeywordConst,
		TokenKeywordDo, TokenKeywordElse, TokenKeywordElseif, TokenKeywordEnd,
		TokenKeywordExport, TokenKeywordFalse, TokenKeywordFinal, TokenKeywordFinally,
		TokenKeywordFor, TokenKeywordFunction, TokenKeywordGlobal, TokenKeywordIf,
		TokenKeywordImport, TokenKeywordIn, TokenKeywordIsA, TokenKeywordLet,
		TokenKeywordLocal, TokenKeywordMacro, TokenKeywordModule, TokenKeywordMutable,
		TokenKeywordNot, TokenKeywordNothing, TokenKeywordOr, TokenKeywordQuote,
		TokenKeywordReturn, TokenKeywordStruct, TokenKeywordTrue, TokenKeywordTry,
		TokenKeywordUsing, TokenKeywordWhile,
	}
	
	tokens := []Token{}
	for {
		token := lexer.NextToken()
		if token.Type == TokenNewline {
			continue
		}
		if token.Type == TokenEOF {
			break
		}
		tokens = append(tokens, token)
	}
	
	if len(tokens) != len(expectedKeywords) {
		t.Errorf("예상 %d개 키워드, 얻은 %d개", len(expectedKeywords), len(tokens))
	}
	
	for i, expected := range expectedKeywords {
		if i < len(tokens) && tokens[i].Type != expected {
			t.Errorf("Token %d: 예상 %v, 얻은 %v (lexeme: %q)", i, expected, tokens[i].Type, tokens[i].Lexeme)
		}
	}
}

// TestPhase1Operators - Phase 1 연산자 테스트
func TestPhase1Operators(t *testing.T) {
	input := `+ - * / % ^ == != < <= > >= && || ! ~ & | . .. ... :: -> => |> << >> >>>
	         += -= *= /= %= ^= &= |= <<= >>= >>>= &&= ||=`
	lexer := NewLexer(input)
	
	tokens := []Token{}
	for {
		token := lexer.NextToken()
		if token.Type == TokenNewline {
			continue
		}
		if token.Type == TokenEOF {
			break
		}
		tokens = append(tokens, token)
	}
	
	if len(tokens) < 30 {
		t.Errorf("연산자가 너무 적음: %d개", len(tokens))
	}
}

// TestPhase1Comments - Phase 1 주석 테스트
func TestPhase1Comments(t *testing.T) {
	input := `x = 5  # 이것은 주석
	y = 10 #= 블록
	주석 =#
	z = 15
	`
	lexer := NewLexer(input)
	
	tokens := []Token{}
	for {
		token := lexer.NextToken()
		if token.Type == TokenNewline {
			continue
		}
		if token.Type == TokenEOF {
			break
		}
		tokens = append(tokens, token)
	}
	
	// x = 5, y = 10, z = 15 총 9개 토큰
	if len(tokens) < 9 {
		t.Errorf("주석이 제대로 제거되지 않음: %d개 토큰", len(tokens))
	}
}

// TestPhase1Numbers - Phase 1 숫자 테스트
func TestPhase1Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"123", TokenInteger},
		{"123.456", TokenFloat},
		{"1e-10", TokenFloat},
		{"3.14159", TokenFloat},
		{"1im", TokenComplexNumber},
		{"2.5e-10im", TokenComplexNumber},
		{"1//2", TokenRationalNumber},
		{"3//4", TokenRationalNumber},
	}
	
	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		token := lexer.NextToken()
		if token.Type != tt.expected {
			t.Errorf("입력 %q: 예상 %v, 얻은 %v", tt.input, tt.expected, token.Type)
		}
	}
}

// TestPhase1Symbols - Phase 1 심볼 테스트
func TestPhase1Symbols(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{":symbol", TokenSymbol},
		{":MyVar", TokenSymbol},
		{":test123", TokenSymbol},
	}
	
	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		token := lexer.NextToken()
		if token.Type != tt.expected {
			t.Errorf("입력 %q: 예상 %v, 얻은 %v", tt.input, tt.expected, token.Type)
		}
	}
}

// TestPhase1IdentifierWithSpecialChars - Phase 1 특수 식별자 테스트
func TestPhase1IdentifierWithSpecialChars(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"test!", TokenIdentifier},
		{"test?", TokenIdentifier},
		{"push!", TokenIdentifier},
		{"isempty?", TokenIdentifier},
	}
	
	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		token := lexer.NextToken()
		if token.Type != tt.expected {
			t.Errorf("입력 %q: 예상 %v, 얻은 %v (lexeme: %q)", tt.input, tt.expected, token.Type, token.Lexeme)
		}
	}
}

// TestPhase1Strings - Phase 1 문자열 테스트
func TestPhase1Strings(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{`"hello"`, TokenString},
		{`'world'`, TokenString},
		{`"with\nescape"`, TokenString},
	}
	
	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		token := lexer.NextToken()
		if token.Type != tt.expected {
			t.Errorf("입력 %q: 예상 %v, 얻은 %v", tt.input, tt.expected, token.Type)
		}
	}
}

// BenchmarkLexer - Lexer 성능 벤치마크
func BenchmarkLexer(b *testing.B) {
	input := `
	function fibonacci(n::Int)::Int
		if n <= 1
			return n
		else
			return fibonacci(n-1) + fibonacci(n-2)
		end
	end
	
	result = fibonacci(10)
	println("Result: $result")
	`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(input)
		for {
			token := lexer.NextToken()
			if token.Type == TokenEOF {
				break
			}
		}
	}
}

// TestPhase1ComplexCode - Phase 1 복잡한 코드 테스트
func TestPhase1ComplexCode(t *testing.T) {
	input := `
	struct Point{T<:Real}
		x::T
		y::T
	end
	
	function distance(p1::Point, p2::Point)::Float64
		return sqrt((p1.x - p2.x)^2 + (p1.y - p2.y)^2)
	end
	
	p1 = Point(3.0, 4.0)
	p2 = Point(0.0, 0.0)
	d = distance(p1, p2)
	@printf("Distance: %.2f\\n", d)
	`
	
	lexer := NewLexer(input)
	tokens := lexer.ScanAll()
	
	// 최소 30개 토큰 확인
	if len(tokens) < 30 {
		t.Errorf("복잡한 코드에서 토큰이 부족함: %d개", len(tokens))
	}
	
	// EOF 확인
	if tokens[len(tokens)-1].Type != TokenEOF {
		t.Errorf("마지막 토큰이 EOF가 아님: %v", tokens[len(tokens)-1].Type)
	}
}
