package test

import (
	"testing"

	"juliacc/internal/codegen"
	"juliacc/internal/ir"
	"juliacc/internal/lexer"
	"juliacc/internal/optimizer"
	"juliacc/internal/parser"
	"juliacc/internal/sema"
	"juliacc/internal/typeinf"
)

// TestE2ECompilation 테스트: 전체 컴파일 파이프라인
func TestE2ECompilation(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   interface{}
	}{
		{
			name:   "simple_literal",
			source: "42",
			want:   int64(42),
		},
		{
			name:   "arithmetic_operation",
			source: "2 + 3",
			want:   int64(5),
		},
		{
			name:   "nested_arithmetic",
			source: "2 * 3 + 4",
			want:   int64(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Phase 1: Lexing
			lex := lexer.New(tt.source)
			tokens, err := lex.Tokenize()
			if err != nil {
				t.Fatalf("lexing failed: %v", err)
			}

			// Phase 2: Parsing
			p := parser.New(tokens)
			program, err := p.Parse()
			if err != nil {
				t.Fatalf("parsing failed: %v", err)
			}

			// Phase 3: Semantic Analysis
			analyzer := sema.NewAnalyzer(nil, nil)
			_, err = analyzer.Analyze(program)
			if err != nil {
				t.Fatalf("semantic analysis failed: %v", err)
			}

			// Phase 4: IR Generation
			builder := ir.NewBuilder()
			irModule, err := builder.Build(program.Statements)
			if err != nil {
				t.Fatalf("IR generation failed: %v", err)
			}

			// Phase 5: Type Inference
			inferrer := typeinf.NewInferrer(irModule)
			if err := inferrer.Infer(); err != nil {
				t.Fatalf("type inference failed: %v", err)
			}

			// Phase 6: Optimization
			opt := optimizer.NewOptimizer(irModule)
			if err := opt.Optimize(); err != nil {
				t.Fatalf("optimization failed: %v", err)
			}

			// Phase 7: Code Generation
			cg := codegen.NewCodegen(irModule)
			bytecode, err := cg.Generate()
			if err != nil {
				t.Fatalf("code generation failed: %v", err)
			}

			// Phase 8: VM Execution
			vm := codegen.NewVM(bytecode)
			result, err := vm.Run()
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			if result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

// BenchmarkLexer 벤치마크: 렉서 성능
func BenchmarkLexer(b *testing.B) {
	source := `
function fibonacci(n)
    if n <= 1
        return n
    end
    return fibonacci(n-1) + fibonacci(n-2)
end
fibonacci(10)
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lex := lexer.New(source)
		_, _ = lex.Tokenize()
	}
}

// BenchmarkParser 벤치마크: 파서 성능
func BenchmarkParser(b *testing.B) {
	source := "2 * 3 + 4 * 5"

	lex := lexer.New(source)
	tokens, _ := lex.Tokenize()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.New(tokens)
		_, _ = p.Parse()
	}
}

// BenchmarkFullCompilation 벤치마크: 전체 컴파일 파이프라인
func BenchmarkFullCompilation(b *testing.B) {
	source := "2 + 3 * 4"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Phase 1
		lex := lexer.New(source)
		tokens, _ := lex.Tokenize()

		// Phase 2
		p := parser.New(tokens)
		program, _ := p.Parse()

		// Phase 3
		analyzer := sema.NewAnalyzer(nil, nil)
		analyzer.Analyze(program)

		// Phase 4
		builder := ir.NewBuilder()
		irModule, _ := builder.Build(program.Statements)

		// Phase 5
		inferrer := typeinf.NewInferrer(irModule)
		inferrer.Infer()

		// Phase 6
		opt := optimizer.NewOptimizer(irModule)
		opt.Optimize()

		// Phase 7
		cg := codegen.NewCodegen(irModule)
		cg.Generate()
	}
}
