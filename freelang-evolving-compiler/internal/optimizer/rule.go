// Package optimizer implements adaptive optimization rules
package optimizer

import (
	"fmt"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/profiler"
)

// OptimizationRule defines a transformation rule
type OptimizationRule struct {
	Name            string
	TargetPattern   profiler.PatternKind
	Priority        int
	Description     string
	Apply           func(*ast.Node) *ast.Node
}

// ConstantFoldingRule evaluates constant expressions at compile time
var ConstantFoldingRule OptimizationRule

func initConstantFoldingRule() {
	ConstantFoldingRule = OptimizationRule{
		Name:          "ConstantFolding",
		TargetPattern: profiler.PatternConstantExpr,
		Priority:      100,
		Description:   "Evaluate constant expressions at compile time",
		Apply: func(node *ast.Node) *ast.Node {
			if node == nil {
				return node
			}

			// If it's a binary expression with two integer literals
			if node.Kind == ast.NodeBinaryExpr && len(node.Children) == 2 {
				left, right := node.Children[0], node.Children[1]

				if left.Kind == ast.NodeIntLit && right.Kind == ast.NodeIntLit {
					result := evaluateConstExpr(node)
					if result != nil {
						return result
					}
				}
			}

			// Recursively apply to children
			for i, child := range node.Children {
				node.Children[i] = ConstantFoldingRule.Apply(child)
			}

			return node
		},
	}
}

// DeadCodeEliminationRule removes unused variable assignments
var DeadCodeEliminationRule OptimizationRule

func initDeadCodeEliminationRule() {
	DeadCodeEliminationRule = OptimizationRule{
		Name:          "DeadCodeElimination",
		TargetPattern: profiler.PatternDeadAssign,
		Priority:      80,
		Description:   "Remove unused variable assignments",
		Apply: func(node *ast.Node) *ast.Node {
			if node == nil {
				return node
			}

			// Mark dead assignments by filtering out declarations that are never read
			// This would require dataflow analysis in a real implementation
			// For now, just mark them with a comment

			for i, child := range node.Children {
				node.Children[i] = DeadCodeEliminationRule.Apply(child)
			}

			return node
		},
	}
}

// InliningRule substitutes function calls with their body
var InliningRule OptimizationRule

func initInliningRule() {
	InliningRule = OptimizationRule{
		Name:          "FunctionInlining",
		TargetPattern: profiler.PatternInlinableCall,
		Priority:      70,
		Description:   "Inline simple function calls",
		Apply: func(node *ast.Node) *ast.Node {
			if node == nil {
				return node
			}

			// Function call inlining would require access to the program's function definitions
			// This is a placeholder implementation

			for i, child := range node.Children {
				node.Children[i] = InliningRule.Apply(child)
			}

			return node
		},
	}
}

// LoopInvariantMovementRule hoists loop-invariant expressions
var LoopInvariantMovementRule OptimizationRule

func initLoopInvariantMovementRule() {
	globalLicmCounter := 0 // 프로그램 전체에서 단조 증가하는 counter

	LoopInvariantMovementRule = OptimizationRule{
		Name:          "LoopInvariantMovement",
		TargetPattern: profiler.PatternLoopInvariant,
		Priority:      101, // ConstantFolding(100) 보다 먼저 실행
		Description:   "Move loop-invariant expressions outside loops",
		Apply: func(node *ast.Node) *ast.Node {
			if node == nil {
				return node
			}

			// 먼저 자식에 재귀 적용
			for i, child := range node.Children {
				node.Children[i] = LoopInvariantMovementRule.Apply(child)
			}

			// for 루프 노드만 LICM 처리
			if node.Kind != ast.NodeForStmt || len(node.Children) < 3 {
				return node
			}
			body := node.Children[2]
			if body.Kind != ast.NodeBlockStmt {
				return node
			}

			// 불변식 수집 및 본문 교체
			var hoisted []*ast.Node
			var newBody []*ast.Node

			for _, stmt := range body.Children {
				if isInvariantLetDecl(stmt) {
					licmName := fmt.Sprintf("__licm_%d", globalLicmCounter)
					globalLicmCounter++

					// 호이스팅: let __licm_N = <원래 상수 식>
					hoisted = append(hoisted, &ast.Node{
						Kind:     ast.NodeLetDecl,
						Line:     stmt.Line,
						Col:      stmt.Col,
						Children: []*ast.Node{
							{Kind: ast.NodeIdent, Value: licmName},
							stmt.Children[1],
						},
					})
					// 루프 내부 교체: let x = __licm_N
					newBody = append(newBody, &ast.Node{
						Kind:     ast.NodeLetDecl,
						Line:     stmt.Line,
						Col:      stmt.Col,
						Children: []*ast.Node{
							stmt.Children[0],
							{Kind: ast.NodeIdent, Value: licmName},
						},
					})
				} else {
					newBody = append(newBody, stmt)
				}
			}

			if len(hoisted) == 0 {
				return node
			}

			body.Children = newBody
			// 루프 노드를 wrapper로 감싸기
			wrapper := &ast.Node{Kind: ast.NodeBlockStmt, Line: node.Line, Col: node.Col}
			wrapper.Children = append(hoisted, node)
			return wrapper
		},
	}
}

// CommonSubexpressionRule eliminates redundant computations
var CommonSubexpressionRule OptimizationRule

func initCommonSubexpressionRule() {
	CommonSubexpressionRule = OptimizationRule{
		Name:          "CommonSubexpressionElimination",
		TargetPattern: profiler.PatternRepeatedSubExpr,
		Priority:      50,
		Description:   "Cache results of repeated sub-expressions",
		Apply: func(node *ast.Node) *ast.Node {
			if node == nil {
				return node
			}

			for i, child := range node.Children {
				node.Children[i] = CommonSubexpressionRule.Apply(child)
			}

			return node
		},
	}
}

// init initializes all rules
func init() {
	initConstantFoldingRule()
	initDeadCodeEliminationRule()
	initInliningRule()
	initLoopInvariantMovementRule()
	initCommonSubexpressionRule()
}

// evaluateConstExpr computes the result of a constant binary expression
func evaluateConstExpr(node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.NodeBinaryExpr {
		return nil
	}

	if len(node.Children) != 2 {
		return nil
	}

	left, right := node.Children[0], node.Children[1]
	if left.Kind != ast.NodeIntLit || right.Kind != ast.NodeIntLit {
		return nil
	}

	leftVal := parseIntLiteral(left.Value)
	rightVal := parseIntLiteral(right.Value)

	var result int64
	switch node.Value {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "/":
		if rightVal == 0 {
			return nil // Don't fold division by zero
		}
		result = leftVal / rightVal
	default:
		return nil // Not a binary operator we can fold
	}

	// Return a new constant node with the computed value
	return &ast.Node{
		Kind:  ast.NodeIntLit,
		Value: formatInt(result),
		Line:  node.Line,
		Col:   node.Col,
	}
}

// parseIntLiteral converts a string to int64
func parseIntLiteral(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}

// formatInt converts int64 back to string
func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	var result string
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		result = string(byte(n%10)+'0') + result
		n /= 10
	}
	if neg {
		result = "-" + result
	}
	return result
}

// DefaultRules returns all optimization rules in default order
func DefaultRules() []OptimizationRule {
	return []OptimizationRule{
		ConstantFoldingRule,
		DeadCodeEliminationRule,
		InliningRule,
		LoopInvariantMovementRule,
		CommonSubexpressionRule,
	}
}

// isInvariantLetDecl returns true if stmt is: let x = <const BinaryExpr>
func isInvariantLetDecl(stmt *ast.Node) bool {
	if stmt == nil || stmt.Kind != ast.NodeLetDecl || len(stmt.Children) != 2 {
		return false
	}
	val := stmt.Children[1]
	return val != nil &&
		val.Kind == ast.NodeBinaryExpr &&
		len(val.Children) == 2 &&
		val.Children[0].Kind == ast.NodeIntLit &&
		val.Children[1].Kind == ast.NodeIntLit
}
