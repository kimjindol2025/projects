package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/codegen"
	"github.com/user/freelang-evolving-compiler/internal/evolution"
	"github.com/user/freelang-evolving-compiler/internal/ir"
	"github.com/user/freelang-evolving-compiler/internal/lexer"
	"github.com/user/freelang-evolving-compiler/internal/optimizer"
	"github.com/user/freelang-evolving-compiler/internal/parser"
	"github.com/user/freelang-evolving-compiler/internal/profiler"
	"github.com/user/freelang-evolving-compiler/internal/typesys"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: freelang-evolving-compiler <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  lex <code>           - Tokenize code")
		fmt.Println("  parse <code>         - Parse code to AST")
		fmt.Println("  profile <code>       - Profile patterns in code")
		fmt.Println("  report               - Show evolution report")
		fmt.Println("  compile <code>       - Full pipeline: parse->optimize->IR->codegen (soft mode)")
		fmt.Println("  compile-strict <code> - Full pipeline with hard type checking")
		return
	}

	cmd := os.Args[1]

	if cmd == "report" {
		profileReport()
		return
	}

	if len(os.Args) < 3 {
		fmt.Printf("Command %s requires code argument\n", cmd)
		return
	}

	code := os.Args[2]

	switch cmd {
	case "lex":
		lexCode(code)
	case "parse":
		parseCode(code)
	case "profile":
		profileCode(code)
	case "compile":
		compileCode(code)
	case "compile-strict":
		compileCodeStrict(code)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func lexCode(code string) {
	l := lexer.New(code)
	fmt.Println("=== Tokens ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Token(type=%v, value=%q, line=%d, col=%d)\n",
			tok.Type, tok.Value, tok.Line, tok.Col)
		if tok.Type == 0 { // TokenEOF
			break
		}
	}
}

func parseCode(code string) {
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	fmt.Println("=== AST ===")
	printAST(prog, 0)
}

func printAST(node *ast.Node, indent int) {
	if node == nil {
		return
	}
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}
	fmt.Printf("%sNode(kind=%v, value=%q)\n", indentStr, node.Kind, node.Value)
	for _, child := range node.Children {
		printAST(child, indent+1)
	}
}

func profileCode(code string) {
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	collector := profiler.NewCollector()
	collector.CollectFromAST(prog)

	patterns := collector.GetPatterns()
	fmt.Println("=== Patterns Collected ===")
	fmt.Printf("Total patterns: %d\n", len(patterns))
	fmt.Printf("Build time: %d ns\n", collector.GetBuildTimeNs())

	topPatterns := collector.GetTopPatterns(10)
	fmt.Println("\nTop Patterns:")
	for i, p := range topPatterns {
		fmt.Printf("%d. %s (count=%d, saved=%d ns)\n",
			i+1, p.Signature, p.Count, p.SavedNs)
	}

	// Save to database
	dbPath := filepath.Join(".", "pattern-db.json")
	db, _ := profiler.LoadFromFile(dbPath)
	db.UpdateFromCollector(collector, code)
	db.SaveToFile(dbPath)
	fmt.Printf("\nDatabase saved to %s\n", dbPath)
	fmt.Printf("Total builds: %d\n", db.TotalBuilds)
}

func profileReport() {
	dbPath := filepath.Join(".", "pattern-db.json")
	db, err := profiler.LoadFromFile(dbPath)
	if err != nil {
		fmt.Printf("Error loading database: %v\n", err)
		return
	}

	fmt.Println("=== Evolution Report ===")
	fmt.Printf("Total builds: %d\n", db.TotalBuilds)
	fmt.Printf("Total patterns learned: %d\n", len(db.Patterns))
	fmt.Println("\nTop Patterns:")

	topPatterns := db.TopPatterns(10)
	for i, p := range topPatterns {
		fmt.Printf("%d. %s (count=%d, saved=%d ns)\n",
			i+1, p.Signature, p.Count, p.SavedNs)
	}

	if len(db.BuildHistory) > 0 {
		fmt.Println("\nBuild History:")
		avgTime := db.AverageBuildTime()
		latestTime := db.LatestBuildTime()

		fmt.Printf("Average build time: %d ns\n", avgTime)
		fmt.Printf("Latest build time: %d ns\n", latestTime)

		if isRegression, ratio := db.DetectRegression(1.2); isRegression {
			fmt.Printf("⚠️  Performance Regression Detected: %.2fx slower\n", ratio)
		} else {
			fmt.Println("✅ No performance regression")
		}
	}
}

func compileCode(code string) {
	start := time.Now()

	// Step 1: Parse
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	// Step 1.5: Type check
	tc := typesys.NewTypeChecker()
	if typeErrs := tc.Check(prog); len(typeErrs) > 0 {
		fmt.Println("=== Type Errors ===")
		for _, e := range typeErrs {
			fmt.Printf("  line %d, col %d: %s\n", e.Line, e.Col, e.Message)
		}
		// Soft mode: warn but continue compilation
		fmt.Println("(Continuing compilation in soft mode...)")
	}

	// Step 2: Collect patterns
	collector := profiler.NewCollector()
	collector.CollectFromAST(prog)

	// Step 3: Load DB and update priorities
	dbPath := filepath.Join(".", "pattern-db.json")
	db, _ := profiler.LoadFromFile(dbPath)

	// Step 4: Optimize
	opt := optimizer.NewAdaptiveOptimizer()
	opt.UpdatePriorities(db)
	optimized, stats := opt.OptimizeWithStats(prog)

	// Step 5: Generate IR
	gen := ir.NewGenerator()
	irProg, err := gen.Generate(optimized)
	if err != nil {
		fmt.Printf("IR generation error: %v\n", err)
		return
	}

	// Step 6: Generate code
	cg := codegen.New()
	result := cg.Generate(irProg)

	// Step 7: Calculate build metrics
	buildTimeNs := time.Since(start).Nanoseconds()

	// Step 8: Record build
	recorder := evolution.NewEvolutionRecorder()
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))[:8]
	metric := recorder.RecordBuild(buildTimeNs, stats.RulesApplied, result.ByteSize, sourceHash)

	// Step 9: Check health
	detector := evolution.NewRegressionDetector(recorder)
	health := detector.GetHealthStatus()

	// Step 10: Update and save DB
	db.UpdateFromCollector(collector, code)
	db.SaveToFile(dbPath)

	// Output
	fmt.Println("=== Generated Code ===")
	fmt.Println(result.Code)
	fmt.Printf("\n=== Build Metrics ===\n")
	fmt.Printf("Build ID: %s\n", metric.BuildID)
	fmt.Printf("Build time: %d ns (%.2f ms)\n", metric.BuildTimeNs, float64(metric.BuildTimeNs)/1000000)
	fmt.Printf("Optimizations applied: %d\n", metric.OptsPassed)
	fmt.Printf("Code size: %d bytes\n", metric.CodeSizeBy)
	fmt.Printf("Health status: %s\n", health)
	fmt.Printf("Optimization rules: %v\n", stats.RulesApplied)
}

func compileCodeStrict(code string) {
	start := time.Now()

	// Step 1: Parse
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	// Step 1.5: Type check (hard mode - errors are fatal)
	tc := typesys.NewTypeCheckerHard()
	if typeErrs := tc.Check(prog); len(typeErrs) > 0 {
		fmt.Println("=== Type Errors (Strict Mode) ===")
		for _, e := range typeErrs {
			fmt.Printf("  line %d, col %d: %s\n", e.Line, e.Col, e.Message)
		}
		// Hard mode: stop compilation on type errors
		return
	}

	// Step 2: Collect patterns
	collector := profiler.NewCollector()
	collector.CollectFromAST(prog)

	// Step 3: Load DB and update priorities
	dbPath := filepath.Join(".", "pattern-db.json")
	db, _ := profiler.LoadFromFile(dbPath)

	// Step 4: Optimize
	opt := optimizer.NewAdaptiveOptimizer()
	opt.UpdatePriorities(db)
	optimized, stats := opt.OptimizeWithStats(prog)

	// Step 5: Generate IR
	gen := ir.NewGenerator()
	irProg, err := gen.Generate(optimized)
	if err != nil {
		fmt.Printf("IR generation error: %v\n", err)
		return
	}

	// Step 6: Generate code
	cg := codegen.New()
	result := cg.Generate(irProg)

	// Step 7: Calculate build metrics
	buildTimeNs := time.Since(start).Nanoseconds()

	// Step 8: Record build
	recorder := evolution.NewEvolutionRecorder()
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(code)))[:8]
	metric := recorder.RecordBuild(buildTimeNs, stats.RulesApplied, result.ByteSize, sourceHash)

	// Step 9: Check health
	detector := evolution.NewRegressionDetector(recorder)
	health := detector.GetHealthStatus()

	// Step 10: Update and save DB
	db.UpdateFromCollector(collector, code)
	db.SaveToFile(dbPath)

	// Output
	fmt.Println("=== Generated Code (Strict Mode) ===")
	fmt.Println(result.Code)
	fmt.Printf("\n=== Build Metrics ===\n")
	fmt.Printf("Build ID: %s\n", metric.BuildID)
	fmt.Printf("Build time: %d ns (%.2f ms)\n", metric.BuildTimeNs, float64(metric.BuildTimeNs)/1000000)
	fmt.Printf("Optimizations applied: %d\n", metric.OptsPassed)
	fmt.Printf("Code size: %d bytes\n", metric.CodeSizeBy)
	fmt.Printf("Health status: %s\n", health)
	fmt.Printf("Optimization rules: %v\n", stats.RulesApplied)
}
