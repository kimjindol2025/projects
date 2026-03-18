package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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

	// 파일 읽기
	source, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: 파일을 읽을 수 없습니다: %v\n", err)
		os.Exit(1)
	}

	// 컴파일 파이프라인
	if err := compile(string(source), *debugFlag); err != nil {
		fmt.Fprintf(os.Stderr, "컴파일 오류: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 컴파일 성공: %s\n", *outputFile)
}

func compile(source string, debug bool) error {
	if debug {
		fmt.Println("🐛 디버그 모드 활성화")
	}

	// Phase 1: Lexing
	if debug {
		fmt.Println("Phase 1: Lexing...")
	}
	lex := lexer.New(source)
	tokens, err := lex.Tokenize()
	if err != nil {
		return fmt.Errorf("lexing error: %v", err)
	}
	if debug {
		fmt.Printf("  ✓ %d tokens\n", len(tokens))
	}

	// Phase 2: Parsing
	if debug {
		fmt.Println("Phase 2: Parsing...")
	}
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}
	if debug {
		fmt.Printf("  ✓ Parsed %d statements\n", len(program.Statements))
	}

	// Phase 3: Semantic Analysis
	if debug {
		fmt.Println("Phase 3: Semantic Analysis...")
	}
	analyzer := sema.NewAnalyzer()
	_, err = analyzer.Analyze(program)
	if err != nil {
		return fmt.Errorf("semantic analysis error: %v", err)
	}
	if debug {
		fmt.Println("  ✓ Semantic analysis passed")
	}

	// Phase 5: IR Generation
	if debug {
		fmt.Println("Phase 5: IR Generation...")
	}
	builder := ir.NewBuilder()
	irModule, err := builder.Build(program)
	if err != nil {
		return fmt.Errorf("IR generation error: %v", err)
	}
	if debug {
		fmt.Printf("  ✓ Generated IR module with %d functions\n", len(irModule.Functions))
	}

	// Phase 6: Type Inference
	if debug {
		fmt.Println("Phase 6: Type Inference...")
	}
	inferrer := typeinf.NewInferrer(irModule)
	if err := inferrer.Infer(); err != nil {
		return fmt.Errorf("type inference error: %v", err)
	}
	if debug {
		fmt.Println("  ✓ Type inference complete")
	}

	// Phase 7: Optimization
	if debug {
		fmt.Println("Phase 7: Optimization...")
	}
	opt := optimizer.NewOptimizer(irModule)
	if err := opt.Optimize(); err != nil {
		return fmt.Errorf("optimization error: %v", err)
	}
	if debug {
		fmt.Println("  ✓ Optimization complete")
	}

	// Phase 8: Codegen
	if debug {
		fmt.Println("Phase 8: Code Generation...")
	}
	cg := codegen.NewCodegen(irModule)
	bytecode, err := cg.Generate()
	if err != nil {
		return fmt.Errorf("code generation error: %v", err)
	}
	if debug {
		fmt.Printf("  ✓ Generated %d bytes of bytecode\n", len(bytecode.Code))
	}

	// Phase 8b: VM Execution
	if debug {
		fmt.Println("Phase 8b: Execution...")
	}
	vm := codegen.NewVM(bytecode)
	result, err := vm.Run()
	if err != nil {
		return fmt.Errorf("execution error: %v", err)
	}
	if debug {
		fmt.Printf("  ✓ Execution complete (result: %v)\n", result)
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
