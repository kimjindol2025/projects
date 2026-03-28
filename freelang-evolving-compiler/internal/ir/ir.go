// Package ir defines intermediate representation for optimized AST
package ir

// Opcode represents an intermediate representation instruction operation
type Opcode int

const (
	OpNoop Opcode = iota
	OpConst
	OpCopy
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpEq
	OpNe
	OpLt
	OpGt
	OpLe
	OpGe
	OpLabel
	OpJump
	OpJumpIf
	OpJumpIfFalse
	OpCall
	OpParam
	OpReturn
	OpEnter
	OpLeave
	OpStructDef
	OpFieldLoad
	OpFieldStore
)

// Operand represents a value in an instruction (register, immediate, or label)
type Operand struct {
	IsTemp  bool   // true if this is a temporary (t0, t1, ...)
	Name    string // variable name or temp name
	IsImm   bool   // true if this is an immediate integer
	ImmVal  int64  // immediate value
	IsLabel bool   // true if this is a label reference
}

// Instruction is a three-address code instruction
type Instruction struct {
	Op    Opcode
	Dest  Operand
	Src1  Operand
	Src2  Operand
	Label string // for OpLabel, OpJump, OpJumpIf, OpJumpIfFalse
	Fn    string // for OpCall, OpEnter, OpLeave
	Line  int    // source line number
}

// Function represents a function in IR
type Function struct {
	Name   string
	Params []string
	Body   []Instruction
}

// Program represents a complete program in IR
type Program struct {
	Functions []Function
	Main      []Instruction
}

// ByteSize returns the estimated byte size of the program (4 bytes per instruction)
func (p *Program) ByteSize() int {
	count := len(p.Main)
	for _, fn := range p.Functions {
		count += len(fn.Body)
	}
	return count * 4
}
