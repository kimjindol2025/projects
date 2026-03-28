// Package profiler implements AST pattern collection and analysis
package profiler

import (
	"time"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

// Collector accumulates pattern statistics during AST analysis
type Collector struct {
	patterns         map[string]*Pattern // key: signature, value: Pattern
	startTime        time.Time
	buildTimeNs      int64
	totalConstExprs  int64
	totalDeadAssigns int64
	totalInlCalls    int64
}

// NewCollector creates a new pattern collector
func NewCollector() *Collector {
	return &Collector{
		patterns:  make(map[string]*Pattern),
		startTime: time.Now(),
	}
}

// CollectFromAST traverses the AST and collects all patterns
func (c *Collector) CollectFromAST(program *ast.Node) {
	if program == nil {
		return
	}

	ctx := NewPatternContext(program)
	ctx.AnalyzeNode(program)

	// Collect constant expressions
	constExprs := ctx.FindConstantExprs(program)
	for _, p := range constExprs {
		c.recordPattern(p.Signature, PatternConstantExpr, 100000) // 100us saved per const eval
		c.totalConstExprs++
	}

	// Collect dead assigns
	deadAssigns := ctx.DetectDeadAssigns()
	for _, varName := range deadAssigns {
		sig := "DeadAssign:" + varName
		c.recordPattern(sig, PatternDeadAssign, 10000) // 10us saved per unused var
		c.totalDeadAssigns++
	}

	// Collect inlinable calls
	for fnName, callCount := range ctx.FunctionCalls {
		if fnDef, ok := findFunctionDef(fnName, program); ok {
			if isFunctionInlinable(fnDef) {
				sig := "CallExpr:" + fnName
				c.recordPattern(sig, PatternInlinableCall, int64(callCount)*50000) // 50us per inline call
				c.totalInlCalls++
			}
		}
	}

	// Collect repeated expressions (optimization candidate)
	topExprs := ctx.TopPatterns(10)
	for _, expr := range topExprs {
		if expr.Count > 1 {
			c.recordPattern(expr.Signature, PatternRepeatedSubExpr, expr.Count*1000)
		}
	}

	c.buildTimeNs = time.Since(c.startTime).Nanoseconds()
}

// recordPattern updates or creates a pattern entry
func (c *Collector) recordPattern(sig string, kind PatternKind, savedNs int64) {
	if p, exists := c.patterns[sig]; exists {
		p.Count++
		p.SavedNs += savedNs
	} else {
		c.patterns[sig] = &Pattern{
			Kind:      kind,
			Signature: sig,
			Count:     1,
			SavedNs:   savedNs,
		}
	}
}

// GetPatterns returns all collected patterns
func (c *Collector) GetPatterns() []Pattern {
	patterns := make([]Pattern, 0, len(c.patterns))
	for _, p := range c.patterns {
		patterns = append(patterns, *p)
	}
	return patterns
}

// GetBuildTimeNs returns the build duration in nanoseconds
func (c *Collector) GetBuildTimeNs() int64 {
	return c.buildTimeNs
}

// GetTopPatterns returns the top N patterns by count
func (c *Collector) GetTopPatterns(n int) []Pattern {
	patterns := c.GetPatterns()

	// Sort by count descending
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Count > patterns[i].Count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	if n > len(patterns) {
		n = len(patterns)
	}
	return patterns[:n]
}

// isFunctionInlinable checks if a function is simple enough to inline
func isFunctionInlinable(fnDef *ast.Node) bool {
	if fnDef == nil || fnDef.Kind != ast.NodeFnDecl {
		return false
	}

	// Function is inlinable if its body is simple (single return, or short block)
	if len(fnDef.Children) == 0 {
		return false
	}

	body := fnDef.Children[len(fnDef.Children)-1]
	if body == nil {
		return false
	}

	// Single return statement is inlinable
	if body.Kind == ast.NodeReturn {
		return true
	}

	// Block with single statement is inlinable
	if body.Kind == ast.NodeBlockStmt {
		return len(body.Children) <= 1
	}

	return false
}

// Summary returns a text summary of collected patterns
func (c *Collector) Summary() string {
	total := len(c.patterns)
	return "Collected " + string(rune(total)) + " unique patterns in " +
		formatNs(c.buildTimeNs)
}

// formatNs formats nanoseconds as human-readable duration
func formatNs(ns int64) string {
	if ns < 1000 {
		return string(rune(ns)) + "ns"
	}
	if ns < 1000000 {
		return string(rune(ns/1000)) + "us"
	}
	if ns < 1000000000 {
		return string(rune(ns/1000000)) + "ms"
	}
	return string(rune(ns/1000000000)) + "s"
}
