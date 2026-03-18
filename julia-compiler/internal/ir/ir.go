package ir

import (
	"fmt"
	"strings"
)

// Instruction types for the Intermediate Representation
type InstructionType int

const (
	// Literals
	InstLiteral InstructionType = iota
	// Variables
	InstLoad
	InstStore
	// Operators
	InstBinOp
	InstUnaryOp
	// Control flow
	InstLabel
	InstBranch
	InstCondBranch
	InstReturn
	// Functions
	InstCall
	InstFunctionDef
	// Other
	InstNop
)

// Instruction represents a single IR instruction
type Instruction struct {
	Type   InstructionType
	ID     uint32
	OpType string      // For BinOp, UnaryOp
	Ops    []Value     // Operands
	Result Value       // Result value
	Meta   interface{} // Additional metadata
}

// Value represents a value in IR (SSA form)
type Value struct {
	ID    uint32
	Type  string // "i64", "f64", "bool", "void", etc.
	Name  string
	IsRef bool // true if this is a reference to an instruction
}

// BasicBlock contains a sequence of instructions
type BasicBlock struct {
	ID    uint32
	Name  string
	Insts []*Instruction
	Succ  []*BasicBlock // Successors (for control flow)
	Pred  []*BasicBlock // Predecessors
}

// Function represents a function in IR
type Function struct {
	Name       string
	Params     []Value
	ReturnType string
	Blocks     []*BasicBlock
	EntryBlock *BasicBlock
}

// Module contains all functions
type Module struct {
	Functions []*Function
	Globals   []Value
}

// NewModule creates a new IR module
func NewModule() *Module {
	return &Module{
		Functions: []*Function{},
		Globals:   []Value{},
	}
}

// NewFunction creates a new function
func NewFunction(name, returnType string, params []Value) *Function {
	return &Function{
		Name:       name,
		Params:     params,
		ReturnType: returnType,
		Blocks:     []*BasicBlock{},
	}
}

// NewBasicBlock creates a new basic block
func NewBasicBlock(name string) *BasicBlock {
	return &BasicBlock{
		Name:  name,
		Insts: []*Instruction{},
		Succ:  []*BasicBlock{},
		Pred:  []*BasicBlock{},
	}
}

// AddInstruction adds an instruction to a basic block
func (bb *BasicBlock) AddInstruction(inst *Instruction) *Instruction {
	inst.ID = uint32(len(bb.Insts))
	bb.Insts = append(bb.Insts, inst)
	return inst
}

// AddBlock adds a basic block to a function
func (f *Function) AddBlock(bb *BasicBlock) {
	bb.ID = uint32(len(f.Blocks))
	f.Blocks = append(f.Blocks, bb)
	if len(f.Blocks) == 1 {
		f.EntryBlock = bb
	}
}

// String implementations for debugging

func (i *Instruction) String() string {
	switch i.Type {
	case InstLiteral:
		return fmt.Sprintf("v%d = literal %v", i.ID, i.Meta)
	case InstLoad:
		return fmt.Sprintf("v%d = load %s", i.ID, i.Ops[0].Name)
	case InstStore:
		return fmt.Sprintf("store %s = v%d", i.Ops[0].Name, i.Ops[1].ID)
	case InstBinOp:
		if len(i.Ops) >= 2 {
			return fmt.Sprintf("v%d = %s v%d, v%d", i.ID, i.OpType, i.Ops[0].ID, i.Ops[1].ID)
		}
	case InstUnaryOp:
		if len(i.Ops) >= 1 {
			return fmt.Sprintf("v%d = %s v%d", i.ID, i.OpType, i.Ops[0].ID)
		}
	case InstCall:
		args := []string{}
		for _, op := range i.Ops {
			args = append(args, fmt.Sprintf("v%d", op.ID))
		}
		return fmt.Sprintf("v%d = call %s(%s)", i.ID, i.Meta, strings.Join(args, ", "))
	case InstReturn:
		if len(i.Ops) > 0 {
			return fmt.Sprintf("ret v%d", i.Ops[0].ID)
		}
		return "ret void"
	case InstLabel:
		return fmt.Sprintf("label %s", i.Meta)
	case InstBranch:
		return fmt.Sprintf("br %s", i.Meta)
	case InstCondBranch:
		if len(i.Ops) >= 1 {
			return fmt.Sprintf("br.cond v%d, %v", i.Ops[0].ID, i.Meta)
		}
	case InstNop:
		return "nop"
	}
	return fmt.Sprintf("v%d = unknown", i.ID)
}

// PrintModule prints the entire module in a readable format
func (m *Module) PrintModule() string {
	var sb strings.Builder
	sb.WriteString("=== IR Module ===\n")

	for _, fn := range m.Functions {
		sb.WriteString(fmt.Sprintf("\nfunction %s(%d params) -> %s:\n", fn.Name, len(fn.Params), fn.ReturnType))
		for i, param := range fn.Params {
			sb.WriteString(fmt.Sprintf("  param[%d]: %s (%s)\n", i, param.Name, param.Type))
		}

		for _, block := range fn.Blocks {
			sb.WriteString(fmt.Sprintf("\nblock %s:\n", block.Name))
			for _, inst := range block.Insts {
				sb.WriteString(fmt.Sprintf("  %s\n", inst.String()))
			}
		}
	}

	return sb.String()
}
