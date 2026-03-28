// Package profiler implements pattern detection and frequency tracking
package profiler

import (
	"fmt"
	"hash/fnv"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

// PatternKind identifies the type of pattern
type PatternKind int

const (
	PatternConstantExpr PatternKind = iota
	PatternDeadAssign
	PatternInlinableCall
	PatternLoopInvariant
	PatternRepeatedSubExpr
)

func (pk PatternKind) String() string {
	switch pk {
	case PatternConstantExpr:
		return "ConstantExpr"
	case PatternDeadAssign:
		return "DeadAssign"
	case PatternInlinableCall:
		return "InlinableCall"
	case PatternLoopInvariant:
		return "LoopInvariant"
	case PatternRepeatedSubExpr:
		return "RepeatedSubExpr"
	default:
		return "Unknown"
	}
}

// Pattern represents a detected pattern with frequency and optimization benefit
type Pattern struct {
	Kind      PatternKind `json:"kind"`
	Signature string      `json:"signature"`
	Count     int64       `json:"count"`
	SavedNs   int64       `json:"saved_ns"`
}

// PatternSignature generates a unique identifier for an AST node
func PatternSignature(node *ast.Node) string {
	if node == nil {
		return "nil"
	}

	h := fnv.New64a()
	h.Write([]byte(nodeKindStr(node.Kind)))

	if node.Value != "" {
		h.Write([]byte("|"))
		h.Write([]byte(node.Value))
	}

	for _, child := range node.Children {
		h.Write([]byte("|"))
		h.Write([]byte(PatternSignature(child)))
	}

	return fmt.Sprintf("%x", h.Sum64())
}

// PatternType determines what kind of pattern this node represents
func PatternType(node *ast.Node, context *PatternContext) PatternKind {
	if node == nil {
		return 0
	}

	// Constant expression: both operands are literals
	if node.Kind == ast.NodeBinaryExpr && len(node.Children) == 2 {
		left, right := node.Children[0], node.Children[1]
		if isLiteral(left) && isLiteral(right) {
			return PatternConstantExpr
		}
	}

	// Inlinable call: function body is simple (single expression)
	if node.Kind == ast.NodeCallExpr {
		if fnName, ok := findFunctionDef(node.Value, context.Program); ok {
			if fnName.Kind == ast.NodeFnDecl && len(fnName.Children) > 0 {
				body := fnName.Children[len(fnName.Children)-1]
				if body.Kind == ast.NodeBlockStmt && len(body.Children) <= 1 {
					return PatternInlinableCall
				}
			}
		}
	}

	return 0
}

// isLiteral checks if node is a literal (int, bool, string)
func isLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return node.Kind == ast.NodeIntLit
}

// PatternContext holds information for pattern analysis
type PatternContext struct {
	Program           *ast.Node
	VariableUsage     map[string]int    // usage count per variable
	InLoop            bool              // whether we're inside a loop
	LoopInvariants    map[string]bool   // expressions that don't depend on loop var
	RepeatedExprs     map[string]int    // frequency of sub-expressions
	FunctionCalls     map[string]int    // frequency of function calls
	AssignedVars      map[string]bool   // variables that are assigned
	ReadVars          map[string]bool   // variables that are read
}

// NewPatternContext creates a new pattern analysis context
func NewPatternContext(program *ast.Node) *PatternContext {
	return &PatternContext{
		Program:        program,
		VariableUsage:  make(map[string]int),
		LoopInvariants: make(map[string]bool),
		RepeatedExprs:  make(map[string]int),
		FunctionCalls:  make(map[string]int),
		AssignedVars:   make(map[string]bool),
		ReadVars:       make(map[string]bool),
	}
}

// AnalyzeNode recursively analyzes a node for patterns
func (pc *PatternContext) AnalyzeNode(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.NodeLetDecl:
		if len(node.Children) > 0 {
			varName := node.Children[0].Value
			pc.AssignedVars[varName] = true
		}
		for _, child := range node.Children {
			pc.AnalyzeNode(child)
		}

	case ast.NodeIdent:
		pc.VariableUsage[node.Value]++
		pc.ReadVars[node.Value] = true

	case ast.NodeBinaryExpr:
		sig := PatternSignature(node)
		pc.RepeatedExprs[sig]++
		for _, child := range node.Children {
			pc.AnalyzeNode(child)
		}

	case ast.NodeCallExpr:
		pc.FunctionCalls[node.Value]++
		for _, child := range node.Children {
			pc.AnalyzeNode(child)
		}

	case ast.NodeForStmt:
		oldInLoop := pc.InLoop
		pc.InLoop = true
		for _, child := range node.Children {
			pc.AnalyzeNode(child)
		}
		pc.InLoop = oldInLoop

	default:
		for _, child := range node.Children {
			pc.AnalyzeNode(child)
		}
	}
}

// DetectDeadAssigns finds variables assigned but never read
func (pc *PatternContext) DetectDeadAssigns() []string {
	var deadAssigns []string
	for varName := range pc.AssignedVars {
		if !pc.ReadVars[varName] {
			deadAssigns = append(deadAssigns, varName)
		}
	}
	return deadAssigns
}

// findFunctionDef searches for a function definition by name
func findFunctionDef(fnName string, program *ast.Node) (*ast.Node, bool) {
	if program == nil || program.Kind != ast.NodeProgram {
		return nil, false
	}
	for _, stmt := range program.Children {
		if stmt.Kind == ast.NodeFnDecl && stmt.Value == fnName {
			return stmt, true
		}
	}
	return nil, false
}

// nodeKindStr converts NodeKind to string for signature generation
func nodeKindStr(kind ast.NodeKind) string {
	switch kind {
	case ast.NodeProgram:
		return "Program"
	case ast.NodeLetDecl:
		return "LetDecl"
	case ast.NodeFnDecl:
		return "FnDecl"
	case ast.NodeIfStmt:
		return "IfStmt"
	case ast.NodeForStmt:
		return "ForStmt"
	case ast.NodeBinaryExpr:
		return "BinaryExpr"
	case ast.NodeCallExpr:
		return "CallExpr"
	case ast.NodeIdent:
		return "Ident"
	case ast.NodeIntLit:
		return "IntLit"
	case ast.NodeReturn:
		return "Return"
	case ast.NodeBlockStmt:
		return "BlockStmt"
	default:
		return "Unknown"
	}
}

// SignaturePattern is a pattern with its count per build
type SignaturePattern struct {
	Signature string
	Kind      PatternKind
	Count     int64
	SavedNs   int64
}

// TopPatterns returns the top N most frequent patterns
func (pc *PatternContext) TopPatterns(n int) []SignaturePattern {
	patterns := make([]SignaturePattern, 0)

	// Collect repeated expressions
	for sig, count := range pc.RepeatedExprs {
		if count > 1 {
			patterns = append(patterns, SignaturePattern{
				Signature: sig,
				Kind:      PatternRepeatedSubExpr,
				Count:     int64(count),
				SavedNs:   0,
			})
		}
	}

	// Collect function calls
	for fnName, count := range pc.FunctionCalls {
		patterns = append(patterns, SignaturePattern{
			Signature: "CallExpr:" + fnName,
			Kind:      PatternInlinableCall,
			Count:     int64(count),
			SavedNs:   0,
		})
	}

	// Sort by count (simple bubble sort for small N)
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Count > patterns[i].Count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	// Return top N
	if n > len(patterns) {
		n = len(patterns)
	}
	return patterns[:n]
}

// FindConstantExprs finds all binary expressions with constant operands
func (pc *PatternContext) FindConstantExprs(node *ast.Node) []SignaturePattern {
	var patterns []SignaturePattern
	findConstExpr(node, &patterns)
	return patterns
}

func findConstExpr(node *ast.Node, patterns *[]SignaturePattern) {
	if node == nil {
		return
	}

	if node.Kind == ast.NodeBinaryExpr && len(node.Children) == 2 {
		if isLiteral(node.Children[0]) && isLiteral(node.Children[1]) {
			sig := PatternSignature(node)
			*patterns = append(*patterns, SignaturePattern{
				Signature: sig,
				Kind:      PatternConstantExpr,
				Count:     1,
				SavedNs:   0,
			})
		}
	}

	for _, child := range node.Children {
		findConstExpr(child, patterns)
	}
}
