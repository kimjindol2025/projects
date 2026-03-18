package optimizer

import (
	"juliacc/internal/ir"
)

// Optimizer performs IR optimizations
type Optimizer struct {
	module *ir.Module
}

// NewOptimizer creates a new optimizer
func NewOptimizer(module *ir.Module) *Optimizer {
	return &Optimizer{
		module: module,
	}
}

// Optimize runs all optimization passes
func (opt *Optimizer) Optimize() error {
	// Run passes in order
	opt.constantFold()
	opt.deadCodeElimination()
	opt.commonSubexpressionElimination()
	return nil
}

// constantFold performs constant folding optimization
func (opt *Optimizer) constantFold() {
	for _, fn := range opt.module.Functions {
		for _, block := range fn.Blocks {
			opt.foldBlockConstants(block)
		}
	}
}

// foldBlockConstants folds constants in a block
func (opt *Optimizer) foldBlockConstants(block *ir.BasicBlock) {
	newInsts := []*ir.Instruction{}

	for i, inst := range block.Insts {
		if inst.Type == ir.InstBinOp && len(inst.Ops) >= 2 {
			left := inst.Ops[0]
			right := inst.Ops[1]

			// Check if both operands are literals
			leftLit := opt.findLiteralValue(block.Insts[:i], left.ID)
			rightLit := opt.findLiteralValue(block.Insts[:i], right.ID)

			if leftLit != nil && rightLit != nil {
				result := opt.evalBinOp(inst.OpType, leftLit, rightLit)
				if result != nil {
					// Replace with constant
					inst.Type = ir.InstLiteral
					inst.Meta = result
					inst.Ops = []ir.Value{}
				}
			}
		}
		newInsts = append(newInsts, inst)
	}

	block.Insts = newInsts
}

// findLiteralValue finds a literal value by ID in instruction list
func (opt *Optimizer) findLiteralValue(insts []*ir.Instruction, valID uint32) interface{} {
	for _, inst := range insts {
		if inst.Type == ir.InstLiteral && inst.Result.ID == valID {
			return inst.Meta
		}
	}
	return nil
}

// evalBinOp evaluates a binary operation on constants
func (opt *Optimizer) evalBinOp(op string, left, right interface{}) interface{} {
	leftInt, leftOK := toInt64(left)
	rightInt, rightOK := toInt64(right)

	if leftOK && rightOK {
		switch op {
		case "add":
			return leftInt + rightInt
		case "sub":
			return leftInt - rightInt
		case "mul":
			return leftInt * rightInt
		case "div":
			if rightInt != 0 {
				return leftInt / rightInt
			}
		case "mod":
			if rightInt != 0 {
				return leftInt % rightInt
			}
		case "eq":
			return leftInt == rightInt
		case "ne":
			return leftInt != rightInt
		case "lt":
			return leftInt < rightInt
		case "le":
			return leftInt <= rightInt
		case "gt":
			return leftInt > rightInt
		case "ge":
			return leftInt >= rightInt
		}
	}

	// Float operations
	leftFloat, leftFOK := toFloat64(left)
	rightFloat, rightFOK := toFloat64(right)

	if leftFOK && rightFOK {
		switch op {
		case "add":
			return leftFloat + rightFloat
		case "sub":
			return leftFloat - rightFloat
		case "mul":
			return leftFloat * rightFloat
		case "div":
			if rightFloat != 0 {
				return leftFloat / rightFloat
			}
		case "eq":
			return leftFloat == rightFloat
		case "ne":
			return leftFloat != rightFloat
		case "lt":
			return leftFloat < rightFloat
		case "le":
			return leftFloat <= rightFloat
		case "gt":
			return leftFloat > rightFloat
		case "ge":
			return leftFloat >= rightFloat
		}
	}

	return nil
}

// deadCodeElimination removes dead instructions
func (opt *Optimizer) deadCodeElimination() {
	for _, fn := range opt.module.Functions {
		for _, block := range fn.Blocks {
			opt.eliminateDeadCode(block)
		}
	}
}

// eliminateDeadCode removes unreachable instructions
func (opt *Optimizer) eliminateDeadCode(block *ir.BasicBlock) {
	// Remove code after unconditional returns/branches
	for i, inst := range block.Insts {
		if inst.Type == ir.InstReturn || inst.Type == ir.InstBranch {
			block.Insts = block.Insts[:i+1]
			break
		}
	}
}

// commonSubexpressionElimination (CSE) - simplified version
func (opt *Optimizer) commonSubexpressionElimination() {
	for _, fn := range opt.module.Functions {
		for _, block := range fn.Blocks {
			opt.eliminateCSE(block)
		}
	}
}

// eliminateCSE eliminates common subexpressions in a block
func (opt *Optimizer) eliminateCSE(block *ir.BasicBlock) {
	// Track identical expressions
	exprMap := make(map[string]uint32)

	newInsts := []*ir.Instruction{}
	for _, inst := range block.Insts {
		if inst.Type == ir.InstBinOp {
			// Create expression signature
			sig := opt.createExprSignature(inst)
			if prevID, exists := exprMap[sig]; exists {
				// Replace with previous result
				inst.Type = ir.InstLoad
				inst.Ops = []ir.Value{{ID: prevID}}
				inst.OpType = ""
			} else {
				exprMap[sig] = inst.Result.ID
			}
		}
		newInsts = append(newInsts, inst)
	}

	block.Insts = newInsts
}

// createExprSignature creates a signature for an expression
func (opt *Optimizer) createExprSignature(inst *ir.Instruction) string {
	if inst.Type == ir.InstBinOp && len(inst.Ops) >= 2 {
		// Simple signature: "op-op1-op2"
		return inst.OpType + "-" + string(rune(inst.Ops[0].ID)) + "-" + string(rune(inst.Ops[1].ID))
	}
	return ""
}

// Utility functions

func toInt64(v interface{}) (int64, bool) {
	if i, ok := v.(int64); ok {
		return i, true
	}
	return 0, false
}

func toFloat64(v interface{}) (float64, bool) {
	if f, ok := v.(float64); ok {
		return f, true
	}
	if i, ok := v.(int64); ok {
		return float64(i), true
	}
	return 0, false
}
