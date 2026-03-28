package profiler

import (
	"testing"
	"time"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/parser"
)

// TestPatternSignature validates signature generation
func TestPatternSignature(t *testing.T) {
	node := &ast.Node{
		Kind:  ast.NodeIntLit,
		Value: "42",
	}
	sig := PatternSignature(node)
	if sig == "" {
		t.Errorf("expected non-empty signature, got empty")
	}

	// Same node should generate same signature
	sig2 := PatternSignature(node)
	if sig != sig2 {
		t.Errorf("signatures differ for same node: %q vs %q", sig, sig2)
	}
}

// TestPatternContextAnalysis validates AST traversal and variable tracking
func TestPatternContextAnalysis(t *testing.T) {
	code := `
let x = 10
let y = x + 5
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ctx := NewPatternContext(program)
	ctx.AnalyzeNode(program)

	if _, ok := ctx.VariableUsage["x"]; !ok {
		t.Errorf("expected variable x to be tracked")
	}
	if _, ok := ctx.AssignedVars["x"]; !ok {
		t.Errorf("expected variable x in assigned vars")
	}
	if _, ok := ctx.ReadVars["x"]; !ok {
		t.Errorf("expected variable x in read vars")
	}
}

// TestDeadAssignDetection finds unused variables
func TestDeadAssignDetection(t *testing.T) {
	code := `
let x = 10
let y = 20
let z = x + y
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ctx := NewPatternContext(program)
	ctx.AnalyzeNode(program)

	deadAssigns := ctx.DetectDeadAssigns()
	if len(deadAssigns) > 0 {
		t.Errorf("expected no dead assigns, got: %v", deadAssigns)
	}

	// Test with actual dead assign
	code2 := `
let unused = 42
let x = 10
`
	p2 := parser.New(code2)
	program2, err := p2.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ctx2 := NewPatternContext(program2)
	ctx2.AnalyzeNode(program2)
	deadAssigns2 := ctx2.DetectDeadAssigns()
	if len(deadAssigns2) == 0 {
		t.Errorf("expected dead assign detection, got none")
	}
}

// TestCollectorBasic validates pattern collection
func TestCollectorBasic(t *testing.T) {
	code := `
let x = 10
let y = 20
let z = x + y
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	collector := NewCollector()
	collector.CollectFromAST(program)

	patterns := collector.GetPatterns()
	if len(patterns) == 0 {
		t.Errorf("expected collected patterns, got none")
	}

	buildTime := collector.GetBuildTimeNs()
	if buildTime <= 0 {
		t.Errorf("expected positive build time, got %d", buildTime)
	}
}

// TestCollectorConstantExpressions detects constant expressions
func TestCollectorConstantExpressions(t *testing.T) {
	code := `
let x = 10 + 5
let y = 2 * 3
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	collector := NewCollector()
	collector.CollectFromAST(program)

	patterns := collector.GetPatterns()
	foundConst := false
	for _, pattern := range patterns {
		if pattern.Kind == PatternConstantExpr {
			foundConst = true
			break
		}
	}

	if !foundConst {
		t.Errorf("expected to find constant expressions")
	}
}

// TestDatabaseCreation validates database initialization
func TestDatabaseCreation(t *testing.T) {
	db := NewDatabase()
	if db == nil {
		t.Errorf("failed to create database")
	}
	if len(db.Patterns) != 0 {
		t.Errorf("expected empty pattern list, got %d", len(db.Patterns))
	}
	if db.TotalBuilds != 0 {
		t.Errorf("expected 0 builds, got %d", db.TotalBuilds)
	}
}

// TestDatabaseUpdate validates pattern merging
func TestDatabaseUpdate(t *testing.T) {
	code := `
let x = 10
let y = x + 5
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	collector := NewCollector()
	collector.CollectFromAST(program)

	db := NewDatabase()
	db.UpdateFromCollector(collector, code)

	if db.TotalBuilds != 1 {
		t.Errorf("expected 1 build, got %d", db.TotalBuilds)
	}
	if len(db.Patterns) == 0 {
		t.Errorf("expected patterns in database")
	}
	if len(db.BuildHistory) != 1 {
		t.Errorf("expected 1 build record, got %d", len(db.BuildHistory))
	}
}

// TestDatabaseTopPatterns validates sorting
func TestDatabaseTopPatterns(t *testing.T) {
	code1 := `
let x = 10 + 5
let y = 10 + 5
let z = 2 * 3
`
	p := parser.New(code1)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	collector := NewCollector()
	collector.CollectFromAST(program)

	db := NewDatabase()
	db.UpdateFromCollector(collector, code1)

	topPatterns := db.TopPatterns(5)
	if len(topPatterns) == 0 {
		t.Errorf("expected top patterns")
	}

	// Verify descending order
	for i := 0; i < len(topPatterns)-1; i++ {
		if topPatterns[i].Count < topPatterns[i+1].Count {
			t.Errorf("patterns not sorted by count: %d >= %d",
				topPatterns[i].Count, topPatterns[i+1].Count)
		}
	}
}

// TestDatabasePersistence validates save/load round-trip
func TestDatabasePersistence(t *testing.T) {
	code := `
let x = 10
let y = x + 5
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	collector := NewCollector()
	collector.CollectFromAST(program)

	db := NewDatabase()
	db.UpdateFromCollector(collector, code)

	// Save to temporary file (in-memory test)
	tempFile := "/tmp/test_profiler_db.json"
	err = db.SaveToFile(tempFile)
	if err != nil {
		t.Fatalf("failed to save database: %v", err)
	}

	// Load from file
	loadedDb, err := LoadFromFile(tempFile)
	if err != nil {
		t.Fatalf("failed to load database: %v", err)
	}

	if loadedDb.TotalBuilds != db.TotalBuilds {
		t.Errorf("builds mismatch: %d vs %d", loadedDb.TotalBuilds, db.TotalBuilds)
	}
	if len(loadedDb.Patterns) != len(db.Patterns) {
		t.Errorf("pattern count mismatch: %d vs %d", len(loadedDb.Patterns), len(db.Patterns))
	}
}

// TestAverageBuildTime validates computation
func TestAverageBuildTime(t *testing.T) {
	db := &Database{
		BuildHistory: []BuildRecord{
			{BuildTimeNs: 1000000},
			{BuildTimeNs: 2000000},
			{BuildTimeNs: 3000000},
		},
	}

	avg := db.AverageBuildTime()
	expected := int64(2000000)
	if avg != expected {
		t.Errorf("expected average %d, got %d", expected, avg)
	}
}

// TestRegressionDetection validates performance regression detection
func TestRegressionDetection(t *testing.T) {
	db := &Database{
		BuildHistory: []BuildRecord{
			{BuildTimeNs: 1000000},
			{BuildTimeNs: 1100000},
			{BuildTimeNs: 1050000},
			{BuildTimeNs: 1200000},
			{BuildTimeNs: 1150000},
			{BuildTimeNs: 3000000}, // 2.5x slower - regression
		},
	}

	isRegression, ratio := db.DetectRegression(1.5)
	if !isRegression {
		t.Errorf("expected regression detection")
	}
	if ratio < 2.0 {
		t.Errorf("expected ratio > 2.0, got %f", ratio)
	}
}

// TestGetPatternStats validates pattern lookup
func TestGetPatternStats(t *testing.T) {
	db := &Database{
		Patterns: []PatternEntry{
			{
				Kind:      "ConstantExpr",
				Signature: "test123",
				Count:     5,
				SavedNs:   100000,
			},
		},
	}

	stats, found := db.GetPatternStats("test123")
	if !found {
		t.Errorf("expected to find pattern")
	}
	if stats.Count != 5 {
		t.Errorf("expected count 5, got %d", stats.Count)
	}
}

// TestMultipleBuildHistory validates build history accumulation
func TestMultipleBuildHistory(t *testing.T) {
	code := "let x = 10"

	db := NewDatabase()
	for i := 0; i < 3; i++ {
		p := parser.New(code)
		program, _ := p.ParseProgram()
		collector := NewCollector()
		collector.CollectFromAST(program)
		db.UpdateFromCollector(collector, code)
		time.Sleep(10 * time.Millisecond) // ensure different timestamps
	}

	if db.TotalBuilds != 3 {
		t.Errorf("expected 3 builds, got %d", db.TotalBuilds)
	}
	if len(db.BuildHistory) != 3 {
		t.Errorf("expected 3 build records, got %d", len(db.BuildHistory))
	}
}

// TestPatternTypeDetection validates pattern classification
func TestPatternTypeDetection(t *testing.T) {
	code := `
fn add(a, b) {
    return a + b
}
let x = add(1, 2)
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ctx := NewPatternContext(program)
	ctx.AnalyzeNode(program)

	// Check that call is recorded
	if callCount, ok := ctx.FunctionCalls["add"]; !ok || callCount == 0 {
		t.Errorf("expected function call tracking")
	}
}

// TestEmptyProgram handles edge case
func TestEmptyProgram(t *testing.T) {
	program := &ast.Node{
		Kind:     ast.NodeProgram,
		Children: []*ast.Node{},
	}

	collector := NewCollector()
	collector.CollectFromAST(program)

	patterns := collector.GetPatterns()
	if len(patterns) != 0 {
		t.Errorf("expected empty patterns for empty program, got %d", len(patterns))
	}
}
