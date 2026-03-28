// Package optimizer implements adaptive optimization selection
package optimizer

import (
	"sort"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/profiler"
)

// AdaptiveOptimizer prioritizes optimization rules based on pattern frequency
type AdaptiveOptimizer struct {
	rules []OptimizationRule
}

// NewAdaptiveOptimizer creates a new optimizer with default rules
func NewAdaptiveOptimizer() *AdaptiveOptimizer {
	return &AdaptiveOptimizer{
		rules: DefaultRules(),
	}
}

// UpdatePriorities adjusts rule priorities based on pattern database
func (ao *AdaptiveOptimizer) UpdatePriorities(db *profiler.Database) {
	// Get top patterns from database
	topPatterns := db.TopPatterns(10)

	// Map pattern kind to rule indices
	patternToRuleIndex := map[profiler.PatternKind]int{
		profiler.PatternConstantExpr:      0, // ConstantFoldingRule
		profiler.PatternDeadAssign:        1, // DeadCodeEliminationRule
		profiler.PatternInlinableCall:     2, // InliningRule
		profiler.PatternLoopInvariant:     3, // LoopInvariantMovementRule
		profiler.PatternRepeatedSubExpr:   4, // CommonSubexpressionRule
	}

	// Reset priorities to defaults
	defaultRules := DefaultRules()
	for i := range ao.rules {
		ao.rules[i].Priority = defaultRules[i].Priority
	}

	// Boost priorities for frequently seen patterns
	for i, pattern := range topPatterns {
		// Parse pattern kind from the entry
		kind := parsePatternKind(pattern.Kind)
		if ruleIdx, ok := patternToRuleIndex[kind]; ok && ruleIdx < len(ao.rules) {
			// Priority boost: rank position * 10
			boost := (10 - i) * 10
			ao.rules[ruleIdx].Priority += boost
		}
	}

	// Sort rules by priority descending
	sort.Slice(ao.rules, func(i, j int) bool {
		return ao.rules[i].Priority > ao.rules[j].Priority
	})
}

// OptimizeAST applies optimization rules in priority order
func (ao *AdaptiveOptimizer) OptimizeAST(node *ast.Node) *ast.Node {
	if node == nil {
		return node
	}

	// Apply each rule in priority order
	for _, rule := range ao.rules {
		node = rule.Apply(node)
	}

	return node
}

// OptimizeWithStats applies optimizations and returns statistics
func (ao *AdaptiveOptimizer) OptimizeWithStats(node *ast.Node) (*ast.Node, OptimizationStats) {
	stats := OptimizationStats{
		RulesApplied: []string{},
	}

	if node == nil {
		return node, stats
	}

	// Apply each rule and track which rules actually changed something
	for _, rule := range ao.rules {
		before := nodeSignature(node)
		node = rule.Apply(node)
		after := nodeSignature(node)

		if before != after {
			stats.RulesApplied = append(stats.RulesApplied, rule.Name)
		}
	}

	stats.NodeCount = countNodes(node)
	return node, stats
}

// OptimizationStats tracks the results of optimization
type OptimizationStats struct {
	RulesApplied []string
	NodeCount    int
}

// parsePatternKind converts PatternEntry kind string back to PatternKind
func parsePatternKind(kindStr string) profiler.PatternKind {
	switch kindStr {
	case "ConstantExpr":
		return profiler.PatternConstantExpr
	case "DeadAssign":
		return profiler.PatternDeadAssign
	case "InlinableCall":
		return profiler.PatternInlinableCall
	case "LoopInvariant":
		return profiler.PatternLoopInvariant
	case "RepeatedSubExpr":
		return profiler.PatternRepeatedSubExpr
	default:
		return 0
	}
}

// nodeSignature generates a quick hash of AST structure
func nodeSignature(node *ast.Node) string {
	if node == nil {
		return ""
	}

	var sig string
	sig += string(rune(node.Kind))
	sig += "|" + node.Value
	for _, child := range node.Children {
		sig += "|" + nodeSignature(child)
	}
	return sig
}

// countNodes returns the total number of nodes in the AST
func countNodes(node *ast.Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}

// GetRules returns the current rule list
func (ao *AdaptiveOptimizer) GetRules() []OptimizationRule {
	return ao.rules
}

// GetRulePriorities returns all rules with their current priorities
func (ao *AdaptiveOptimizer) GetRulePriorities() map[string]int {
	priorities := make(map[string]int)
	for _, rule := range ao.rules {
		priorities[rule.Name] = rule.Priority
	}
	return priorities
}
