package codegen

import (
	"fmt"
	"juliacc/internal/ir"
)

// BytecodeOp represents a bytecode operation
type BytecodeOp uint8

const (
	OpPush BytecodeOp = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe
	OpAnd
	OpOr
	OpNot
	OpNeg
	OpCall
	OpRet
	OpLoad
	OpStore
	OpBranch
	OpCondBranch
	OpLabel
	OpHalt
)

// Bytecode represents compiled bytecode
type Bytecode struct {
	Code      []uint8
	Constants []interface{}
	Labels    map[string]int
}

// Codegen compiles IR to bytecode
type Codegen struct {
	module    *ir.Module
	bytecode  *Bytecode
	nextConst int
	labelMap  map[string]int
}

// NewCodegen creates a new code generator
func NewCodegen(module *ir.Module) *Codegen {
	return &Codegen{
		module:   module,
		bytecode: &Bytecode{Code: []uint8{}, Constants: []interface{}{}, Labels: make(map[string]int)},
		labelMap: make(map[string]int),
	}
}

// Generate compiles IR to bytecode
func (cg *Codegen) Generate() (*Bytecode, error) {
	for _, fn := range cg.module.Functions {
		if err := cg.compileFunction(fn); err != nil {
			return nil, err
		}
	}

	// Add halt instruction at the end
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpHalt))

	return cg.bytecode, nil
}

// compileFunction compiles a function to bytecode
func (cg *Codegen) compileFunction(fn *ir.Function) error {
	// Mark function entry
	fnLabel := fmt.Sprintf("fn_%s", fn.Name)
	cg.labelMap[fnLabel] = len(cg.bytecode.Code)

	for _, block := range fn.Blocks {
		if err := cg.compileBlock(block); err != nil {
			return err
		}
	}

	return nil
}

// compileBlock compiles a basic block to bytecode
func (cg *Codegen) compileBlock(block *ir.BasicBlock) error {
	blockLabel := fmt.Sprintf("block_%s", block.Name)
	cg.labelMap[blockLabel] = len(cg.bytecode.Code)

	for _, inst := range block.Insts {
		if err := cg.compileInstruction(inst); err != nil {
			return err
		}
	}

	return nil
}

// compileInstruction compiles a single instruction
func (cg *Codegen) compileInstruction(inst *ir.Instruction) error {
	switch inst.Type {
	case ir.InstLiteral:
		return cg.compileLiteral(inst)
	case ir.InstBinOp:
		return cg.compileBinOp(inst)
	case ir.InstUnaryOp:
		return cg.compileUnaryOp(inst)
	case ir.InstCall:
		return cg.compileCall(inst)
	case ir.InstReturn:
		return cg.compileReturn(inst)
	case ir.InstLoad:
		return cg.compileLoad(inst)
	case ir.InstStore:
		return cg.compileStore(inst)
	case ir.InstBranch:
		return cg.compileBranch(inst)
	case ir.InstCondBranch:
		return cg.compileCondBranch(inst)
	case ir.InstLabel:
		cg.labelMap[fmt.Sprintf("%v", inst.Meta)] = len(cg.bytecode.Code)
		return nil
	default:
		return nil
	}
}

// compileLiteral compiles a literal value
func (cg *Codegen) compileLiteral(inst *ir.Instruction) error {
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpPush))
	constIdx := cg.addConstant(inst.Meta)
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(constIdx))
	return nil
}

// compileBinOp compiles a binary operation
func (cg *Codegen) compileBinOp(inst *ir.Instruction) error {
	op, ok := opToBytecodeOp(inst.OpType)
	if !ok {
		return fmt.Errorf("unknown binary operator: %s", inst.OpType)
	}

	cg.bytecode.Code = append(cg.bytecode.Code, uint8(op))
	return nil
}

// compileUnaryOp compiles a unary operation
func (cg *Codegen) compileUnaryOp(inst *ir.Instruction) error {
	var op BytecodeOp
	switch inst.OpType {
	case "not":
		op = OpNot
	case "-":
		op = OpNeg
	default:
		op = OpNeg
	}

	cg.bytecode.Code = append(cg.bytecode.Code, uint8(op))
	return nil
}

// compileCall compiles a function call
func (cg *Codegen) compileCall(inst *ir.Instruction) error {
	fnName, ok := inst.Meta.(string)
	if !ok {
		return fmt.Errorf("invalid function name in call")
	}

	// For built-in functions, use special handling
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpCall))
	constIdx := cg.addConstant(fnName)
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(constIdx))
	return nil
}

// compileReturn compiles a return instruction
func (cg *Codegen) compileReturn(inst *ir.Instruction) error {
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpRet))
	return nil
}

// compileLoad compiles a load instruction
func (cg *Codegen) compileLoad(inst *ir.Instruction) error {
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpLoad))
	if len(inst.Ops) > 0 {
		// Store variable name as constant
		constIdx := cg.addConstant(inst.Ops[0].Name)
		cg.bytecode.Code = append(cg.bytecode.Code, uint8(constIdx))
	}
	return nil
}

// compileStore compiles a store instruction
func (cg *Codegen) compileStore(inst *ir.Instruction) error {
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpStore))
	if len(inst.Ops) > 0 {
		constIdx := cg.addConstant(inst.Ops[0].Name)
		cg.bytecode.Code = append(cg.bytecode.Code, uint8(constIdx))
	}
	return nil
}

// compileBranch compiles an unconditional branch
func (cg *Codegen) compileBranch(inst *ir.Instruction) error {
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpBranch))
	// Placeholder for label address (will be patched)
	cg.bytecode.Code = append(cg.bytecode.Code, 0)
	return nil
}

// compileCondBranch compiles a conditional branch
func (cg *Codegen) compileCondBranch(inst *ir.Instruction) error {
	cg.bytecode.Code = append(cg.bytecode.Code, uint8(OpCondBranch))
	// Placeholder for label addresses
	cg.bytecode.Code = append(cg.bytecode.Code, 0)
	return nil
}

// addConstant adds a constant to the constant pool
func (cg *Codegen) addConstant(val interface{}) int {
	idx := len(cg.bytecode.Constants)
	cg.bytecode.Constants = append(cg.bytecode.Constants, val)
	return idx
}

// opToBytecodeOp converts IR operator to bytecode operator
func opToBytecodeOp(op string) (BytecodeOp, bool) {
	switch op {
	case "add":
		return OpAdd, true
	case "sub":
		return OpSub, true
	case "mul":
		return OpMul, true
	case "div":
		return OpDiv, true
	case "mod":
		return OpMod, true
	case "eq":
		return OpEq, true
	case "ne":
		return OpNe, true
	case "lt":
		return OpLt, true
	case "le":
		return OpLe, true
	case "gt":
		return OpGt, true
	case "ge":
		return OpGe, true
	case "and":
		return OpAnd, true
	case "or":
		return OpOr, true
	default:
		return 0, false
	}
}

// GetBytecode returns the generated bytecode
func (cg *Codegen) GetBytecode() *Bytecode {
	return cg.bytecode
}
