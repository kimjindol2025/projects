package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"juliacc/internal/codegen"
	"juliacc/internal/ir"
	"juliacc/internal/lexer"
	"juliacc/internal/optimizer"
	"juliacc/internal/parser"
	"juliacc/internal/sema"
	"juliacc/internal/typeinf"
)

const (
	version = "0.1.0-alpha"
	banner  = `
╔═══════════════════════════════════╗
║   JuliaCC - Julia Compiler        ║
║   Version: %s                     ║
║   Go-based Implementation         ║
╚═══════════════════════════════════╝
`
)

// PhaseLogger 헬퍼: Phase 로깅 반복 제거 (Issue #1)
type PhaseLogger struct {
	debug bool
}

func NewPhaseLogger(debug bool) *PhaseLogger {
	return &PhaseLogger{debug: debug}
}

func (pl *PhaseLogger) Run(phaseName string, fn func() (string, error)) error {
	if pl.debug {
		fmt.Printf("Phase: %s...\n", phaseName)
	}

	msg, err := fn()
	if err != nil {
		return fmt.Errorf("%s error: %v", phaseName, err)
	}

	if pl.debug && msg != "" {
		fmt.Printf("  ✓ %s\n", msg)
	}
	return nil
}

// readSourceFile 헬퍼: 파일 읽기 에러 처리 중복 제거 (Issue #2)
func readSourceFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}
	return string(data), nil
}

func main() {
	// 플래그 정의
	versionFlag := flag.Bool("version", false, "버전 출력")
	helpFlag := flag.Bool("help", false, "도움말 출력")
	outputFile := flag.String("o", "", "출력 파일명")
	debugFlag := flag.Bool("debug", false, "디버그 모드")

	flag.Parse()

	// 플래그 처리
	if *versionFlag {
		fmt.Printf(banner, version)
		os.Exit(0)
	}

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	// 입력 파일 확인
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "오류: 입력 파일을 지정해주세요\n")
		printHelp()
		os.Exit(1)
	}

	inputFile := args[0]
	if _, err := os.Stat(inputFile); err != nil {
		fmt.Fprintf(os.Stderr, "오류: 파일을 열 수 없습니다: %s\n", inputFile)
		os.Exit(1)
	}

	// 출력 파일명 기본값 설정
	if *outputFile == "" {
		ext := filepath.Ext(inputFile)
		*outputFile = inputFile[:len(inputFile)-len(ext)] + ".out"
	}

	// 파일 읽기 (Issue #2: readSourceFile 헬퍼 사용)
	source, err := readSourceFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		os.Exit(1)
	}

	// 컴파일 파이프라인
	if err := compile(source, *debugFlag); err != nil {
		fmt.Fprintf(os.Stderr, "컴파일 오류: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 컴파일 성공: %s\n", *outputFile)
}

func compile(source string, debug bool) error {
	if debug {
		fmt.Println("🐛 디버그 모드 활성화")
	}

	logger := NewPhaseLogger(debug)
	var tokens []lexer.Token
	var program *parser.Program
	var irModule *ir.Module
	var bytecode *codegen.Bytecode

	// Phase 1: Lexing
	if err := logger.Run("Lexing", func() (string, error) {
		lex := lexer.New(source)
		var e error
		tokens, e = lex.Tokenize()
		return fmt.Sprintf("%d tokens", len(tokens)), e
	}); err != nil {
		return err
	}

	// Phase 2: Parsing
	if err := logger.Run("Parsing", func() (string, error) {
		p := parser.New(tokens)
		var e error
		program, e = p.Parse()
		if program == nil {
			return "", e
		}
		return fmt.Sprintf("Parsed %d statements", len(program.Statements)), e
	}); err != nil {
		return err
	}

	// Phase 3: Semantic Analysis
	if err := logger.Run("Semantic Analysis", func() (string, error) {
		analyzer := sema.NewAnalyzer()
		_, e := analyzer.Analyze(program)
		return "Semantic analysis passed", e
	}); err != nil {
		return err
	}

	// Phase 4: IR Generation
	if err := logger.Run("IR Generation", func() (string, error) {
		builder := ir.NewBuilder()
		var e error
		irModule, e = builder.Build(program)
		if irModule == nil {
			return "", e
		}
		return fmt.Sprintf("Generated IR module with %d functions", len(irModule.Functions)), e
	}); err != nil {
		return err
	}

	// Phase 5: Type Inference
	if err := logger.Run("Type Inference", func() (string, error) {
		inferrer := typeinf.NewInferrer(irModule)
		return "Type inference complete", inferrer.Infer()
	}); err != nil {
		return err
	}

	// Phase 6: Optimization
	if err := logger.Run("Optimization", func() (string, error) {
		opt := optimizer.NewOptimizer(irModule)
		return "Optimization complete", opt.Optimize()
	}); err != nil {
		return err
	}

	// Phase 7: Code Generation
	if err := logger.Run("Code Generation", func() (string, error) {
		cg := codegen.NewCodegen(irModule)
		var e error
		bytecode, e = cg.Generate()
		if bytecode == nil {
			return "", e
		}
		return fmt.Sprintf("Generated %d bytes of bytecode", len(bytecode.Code)), e
	}); err != nil {
		return err
	}

	// Phase 8: VM Execution
	if err := logger.Run("Execution", func() (string, error) {
		vm := codegen.NewVM(bytecode)
		result, e := vm.Run()
		if e != nil {
			return "", e
		}
		return fmt.Sprintf("Execution complete (result: %v)", result), nil
	}); err != nil {
		return err
	}

	return nil
}

func printHelp() {
	fmt.Printf(banner, version)
	fmt.Print(`
사용법: jcc [옵션] <input_file>

옵션:
  -o <file>      출력 파일명 지정 (기본: input.out)
  -debug         디버그 모드 활성화
  -version       버전 출력
  -help          도움말 출력

예시:
  jcc hello.jl           → hello.out 생성
  jcc -o prog hello.jl   → prog 생성
  jcc -debug hello.jl    → 디버그 출력 활성화

프로젝트: https://github.com/yourusername/juliacc
문서: https://docs.julialang.org
`)
}
