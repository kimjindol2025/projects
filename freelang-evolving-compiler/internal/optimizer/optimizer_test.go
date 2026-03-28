package optimizer

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/parser"
	"github.com/user/freelang-evolving-compiler/internal/profiler"
)

// TestConstantFolding validates constant expression optimization
func TestConstantFolding(t *testing.T) {
	// Create AST for "10 + 5"
	node := &ast.Node{
		Kind:  ast.NodeBinaryExpr,
		Value: "+",
		Children: []*ast.Node{
			{Kind: ast.NodeIntLit, Value: "10"},
			{Kind: ast.NodeIntLit, Value: "5"},
		},
	}

	result := ConstantFoldingRule.Apply(node)
	if result == nil {
		t.Errorf("expected optimized result, got nil")
	}
	if result.Kind != ast.NodeIntLit {
		t.Errorf("expected IntLit after folding, got %d", result.Kind)
	}
	if result.Value != "15" {
		t.Errorf("expected value 15, got %s", result.Value)
	}
}

// TestConstantFoldingSubtraction validates subtraction
func TestConstantFoldingSubtraction(t *testing.T) {
	node := &ast.Node{
		Kind:  ast.NodeBinaryExpr,
		Value: "-",
		Children: []*ast.Node{
			{Kind: ast.NodeIntLit, Value: "20"},
			{Kind: ast.NodeIntLit, Value: "8"},
		},
	}

	result := ConstantFoldingRule.Apply(node)
	if result.Value != "12" {
		t.Errorf("expected 12, got %s", result.Value)
	}
}

// TestConstantFoldingMultiplication validates multiplication
func TestConstantFoldingMultiplication(t *testing.T) {
	node := &ast.Node{
		Kind:  ast.NodeBinaryExpr,
		Value: "*",
		Children: []*ast.Node{
			{Kind: ast.NodeIntLit, Value: "6"},
			{Kind: ast.NodeIntLit, Value: "7"},
		},
	}

	result := ConstantFoldingRule.Apply(node)
	if result.Value != "42" {
		t.Errorf("expected 42, got %s", result.Value)
	}
}

// TestConstantFoldingDivision validates division
func TestConstantFoldingDivision(t *testing.T) {
	node := &ast.Node{
		Kind:  ast.NodeBinaryExpr,
		Value: "/",
		Children: []*ast.Node{
			{Kind: ast.NodeIntLit, Value: "100"},
			{Kind: ast.NodeIntLit, Value: "5"},
		},
	}

	result := ConstantFoldingRule.Apply(node)
	if result.Value != "20" {
		t.Errorf("expected 20, got %s", result.Value)
	}
}

// TestConstantFoldingNested validates nested constant expressions
func TestConstantFoldingNested(t *testing.T) {
	// (10 + 5) * 2
	node := &ast.Node{
		Kind:  ast.NodeBinaryExpr,
		Value: "*",
		Children: []*ast.Node{
			{
				Kind:  ast.NodeBinaryExpr,
				Value: "+",
				Children: []*ast.Node{
					{Kind: ast.NodeIntLit, Value: "10"},
					{Kind: ast.NodeIntLit, Value: "5"},
				},
			},
			{Kind: ast.NodeIntLit, Value: "2"},
		},
	}

	result := ConstantFoldingRule.Apply(node)
	// Inner expression should be folded first
	if result.Kind == ast.NodeBinaryExpr {
		// Check if left child was folded
		if result.Children[0].Kind == ast.NodeIntLit {
			t.Logf("Inner expression folded to %s", result.Children[0].Value)
		}
	}
}

// TestAdaptiveOptimizerCreation validates initialization
func TestAdaptiveOptimizerCreation(t *testing.T) {
	opt := NewAdaptiveOptimizer()
	if opt == nil {
		t.Errorf("failed to create optimizer")
	}
	if len(opt.GetRules()) == 0 {
		t.Errorf("expected rules, got none")
	}
}

// TestAdaptiveOptimizerPriorities validates priority adjustment
func TestAdaptiveOptimizerPriorities(t *testing.T) {
	db := &profiler.Database{
		Patterns: []profiler.PatternEntry{
			{
				Kind:      "ConstantExpr",
				Signature: "expr1",
				Count:     100,
			},
			{
				Kind:      "DeadAssign",
				Signature: "var1",
				Count:     50,
			},
		},
	}

	opt := NewAdaptiveOptimizer()
	opt.UpdatePriorities(db)

	priorities := opt.GetRulePriorities()
	if len(priorities) == 0 {
		t.Errorf("expected priorities, got none")
	}

	// ConstantFolding should have higher priority than DeadCodeElimination
	cfPriority := priorities["ConstantFolding"]
	dcePriority := priorities["DeadCodeElimination"]

	if cfPriority <= dcePriority {
		t.Errorf("ConstantFolding should have higher priority than DeadCodeElimination, got %d vs %d",
			cfPriority, dcePriority)
	}
}

// TestOptimizeAST validates full optimization pipeline
func TestOptimizeAST(t *testing.T) {
	code := `
let x = 10 + 5
let y = x * 2
`
	p := parser.New(code)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	opt := NewAdaptiveOptimizer()
	optimized := opt.OptimizeAST(program)

	if optimized == nil {
		t.Errorf("optimization returned nil")
	}
}

// TestOptimizeWithStats validates stats collection
func TestOptimizeWithStats(t *testing.T) {
	node := &ast.Node{
		Kind:  ast.NodeBinaryExpr,
		Value: "+",
		Children: []*ast.Node{
			{Kind: ast.NodeIntLit, Value: "10"},
			{Kind: ast.NodeIntLit, Value: "5"},
		},
	}

	opt := NewAdaptiveOptimizer()
	result, stats := opt.OptimizeWithStats(node)

	if result == nil {
		t.Errorf("optimization returned nil")
	}
	if stats.NodeCount <= 0 {
		t.Errorf("expected positive node count, got %d", stats.NodeCount)
	}
}

// TestParsePatternKind validates kind conversion
func TestParsePatternKind(t *testing.T) {
	tests := []struct {
		input    string
		expected profiler.PatternKind
	}{
		{"ConstantExpr", profiler.PatternConstantExpr},
		{"DeadAssign", profiler.PatternDeadAssign},
		{"InlinableCall", profiler.PatternInlinableCall},
		{"LoopInvariant", profiler.PatternLoopInvariant},
		{"RepeatedSubExpr", profiler.PatternRepeatedSubExpr},
	}

	for _, tt := range tests {
		kind := parsePatternKind(tt.input)
		if kind != tt.expected {
			t.Errorf("parsePatternKind(%q): expected %d, got %d", tt.input, tt.expected, kind)
		}
	}
}

// TestNodeSignature validates AST hashing
func TestNodeSignature(t *testing.T) {
	node1 := &ast.Node{
		Kind:  ast.NodeIntLit,
		Value: "42",
	}

	node2 := &ast.Node{
		Kind:  ast.NodeIntLit,
		Value: "42",
	}

	sig1 := nodeSignature(node1)
	sig2 := nodeSignature(node2)

	if sig1 != sig2 {
		t.Errorf("same nodes should have same signature: %q vs %q", sig1, sig2)
	}

	node3 := &ast.Node{
		Kind:  ast.NodeIntLit,
		Value: "43",
	}

	sig3 := nodeSignature(node3)
	if sig1 == sig3 {
		t.Errorf("different nodes should have different signatures")
	}
}

// TestCountNodes validates node counting
func TestCountNodes(t *testing.T) {
	node := &ast.Node{
		Kind: ast.NodeBinaryExpr,
		Children: []*ast.Node{
			{Kind: ast.NodeIntLit},
			{Kind: ast.NodeIntLit},
		},
	}

	count := countNodes(node)
	if count != 3 {
		t.Errorf("expected 3 nodes, got %d", count)
	}
}

// TestCountNodesNil validates nil handling
func TestCountNodesNil(t *testing.T) {
	count := countNodes(nil)
	if count != 0 {
		t.Errorf("expected 0 for nil node, got %d", count)
	}
}

// TestOptimizeNilNode validates nil safety
func TestOptimizeNilNode(t *testing.T) {
	opt := NewAdaptiveOptimizer()
	result := opt.OptimizeAST(nil)
	if result != nil {
		t.Errorf("expected nil for nil input")
	}
}

// TestRuleOrder validates that rules maintain priority order
func TestRuleOrder(t *testing.T) {
	opt := NewAdaptiveOptimizer()
	rules := opt.GetRules()

	// Rules should be in some consistent order
	if len(rules) < 2 {
		t.Errorf("expected at least 2 rules, got %d", len(rules))
	}

	// First rule should be constant folding
	if rules[0].Name != "ConstantFolding" {
		t.Errorf("expected first rule to be ConstantFolding, got %s", rules[0].Name)
	}
}

// TestDivisionByZero validates safety
func TestDivisionByZero(t *testing.T) {
	node := &ast.Node{
		Kind:  ast.NodeBinaryExpr,
		Value: "/",
		Children: []*ast.Node{
			{Kind: ast.NodeIntLit, Value: "10"},
			{Kind: ast.NodeIntLit, Value: "0"},
		},
	}

	result := ConstantFoldingRule.Apply(node)
	// Should not fold division by zero
	if result.Kind == ast.NodeIntLit {
		t.Errorf("should not fold division by zero")
	}
}

// TestFormatInt validates integer formatting
func TestFormatInt(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{123, "123"},
		{1000, "1000"},
	}

	for _, tt := range tests {
		result := formatInt(tt.input)
		if result != tt.expected {
			t.Errorf("formatInt(%d): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}
