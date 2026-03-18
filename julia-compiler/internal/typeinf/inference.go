package typeinf

import (
	"fmt"
	"juliacc/internal/ir"
)

// Inferrer performs type inference on IR
type Inferrer struct {
	module *ir.Module
	typeMap map[uint32]string // Maps value ID to inferred type
}

// NewInferrer creates a new type inferencer
func NewInferrer(module *ir.Module) *Inferrer {
	return &Inferrer{
		module:  module,
		typeMap: make(map[uint32]string),
	}
}

// Infer performs type inference on the IR module
func (inf *Inferrer) Infer() error {
	for _, fn := range inf.module.Functions {
		if err := inf.inferFunction(fn); err != nil {
			return err
		}
	}
	return nil
}

// inferFunction infers types in a function
func (inf *Inferrer) inferFunction(fn *ir.Function) error {
	for _, block := range fn.Blocks {
		if err := inf.inferBlock(block); err != nil {
			return err
		}
	}
	return nil
}

// inferBlock infers types in a basic block
func (inf *Inferrer) inferBlock(block *ir.BasicBlock) error {
	for _, inst := range block.Insts {
		if err := inf.inferInstruction(inst); err != nil {
			return err
		}
	}
	return nil
}

// inferInstruction infers types for an instruction
func (inf *Inferrer) inferInstruction(inst *ir.Instruction) error {
	switch inst.Type {
	case ir.InstLiteral:
		return inf.inferLiteral(inst)
	case ir.InstBinOp:
		return inf.inferBinOp(inst)
	case ir.InstUnaryOp:
		return inf.inferUnaryOp(inst)
	case ir.InstCall:
		return inf.inferCall(inst)
	case ir.InstLoad:
		return inf.inferLoad(inst)
	case ir.InstStore:
		return nil // Store doesn't produce a value
	case ir.InstReturn:
		return nil // Return doesn't need type inference
	case ir.InstBranch, ir.InstCondBranch, ir.InstLabel:
		return nil // Control flow doesn't need inference
	default:
		return nil
	}
}

// inferLiteral infers type from literal value
func (inf *Inferrer) inferLiteral(inst *ir.Instruction) error {
	switch v := inst.Meta.(type) {
	case int64:
		inst.Result.Type = "i64"
		inf.typeMap[inst.Result.ID] = "i64"
	case float64:
		inst.Result.Type = "f64"
		inf.typeMap[inst.Result.ID] = "f64"
	case bool:
		inst.Result.Type = "bool"
		inf.typeMap[inst.Result.ID] = "bool"
	case string:
		inst.Result.Type = "string"
		inf.typeMap[inst.Result.ID] = "string"
	default:
		inst.Result.Type = "unknown"
		inf.typeMap[inst.Result.ID] = "unknown"
	}
	return nil
}

// inferBinOp infers type from binary operation
func (inf *Inferrer) inferBinOp(inst *ir.Instruction) error {
	if len(inst.Ops) < 2 {
		return fmt.Errorf("binary op requires 2 operands")
	}

	leftType := inf.typeMap[inst.Ops[0].ID]
	rightType := inf.typeMap[inst.Ops[1].ID]

	// Type promotion rules
	resultType := inf.promoteTypes(leftType, rightType)

	// Special cases for comparison operators
	switch inst.OpType {
	case "eq", "ne", "lt", "le", "gt", "ge":
		resultType = "bool"
	}

	inst.Result.Type = resultType
	inf.typeMap[inst.Result.ID] = resultType
	return nil
}

// inferUnaryOp infers type from unary operation
func (inf *Inferrer) inferUnaryOp(inst *ir.Instruction) error {
	if len(inst.Ops) < 1 {
		return fmt.Errorf("unary op requires 1 operand")
	}

	operandType := inf.typeMap[inst.Ops[0].ID]

	switch inst.OpType {
	case "not":
		inst.Result.Type = "bool"
	default:
		inst.Result.Type = operandType
	}

	inf.typeMap[inst.Result.ID] = inst.Result.Type
	return nil
}

// inferCall infers return type from function call
func (inf *Inferrer) inferCall(inst *ir.Instruction) error {
	fnName, ok := inst.Meta.(string)
	if !ok {
		inst.Result.Type = "unknown"
		return nil
	}

	// Built-in function types
	switch fnName {
	case "print", "println":
		inst.Result.Type = "void"
	case "length", "size":
		inst.Result.Type = "i64"
	case "sqrt", "sin", "cos", "tan":
		inst.Result.Type = "f64"
	case "string":
		inst.Result.Type = "string"
	case "int":
		inst.Result.Type = "i64"
	case "float":
		inst.Result.Type = "f64"
	default:
		// User-defined function: assume i64 for now
		inst.Result.Type = "i64"
	}

	inf.typeMap[inst.Result.ID] = inst.Result.Type
	return nil
}

// inferLoad infers type from load instruction
func (inf *Inferrer) inferLoad(inst *ir.Instruction) error {
	// Simplified: assume i64
	inst.Result.Type = "i64"
	inf.typeMap[inst.Result.ID] = "i64"
	return nil
}

// promoteTypes returns the promoted type for two types
func (inf *Inferrer) promoteTypes(t1, t2 string) string {
	if t1 == t2 {
		return t1
	}

	// Promotion rules: f64 > i64 > bool
	if t1 == "f64" || t2 == "f64" {
		return "f64"
	}
	if t1 == "i64" || t2 == "i64" {
		return "i64"
	}
	if t1 == "string" || t2 == "string" {
		return "string"
	}

	return "unknown"
}

// GetType returns the inferred type of a value
func (inf *Inferrer) GetType(valID uint32) string {
	if t, ok := inf.typeMap[valID]; ok {
		return t
	}
	return "unknown"
}

// GetTypeMap returns the complete type map
func (inf *Inferrer) GetTypeMap() map[uint32]string {
	return inf.typeMap
}
