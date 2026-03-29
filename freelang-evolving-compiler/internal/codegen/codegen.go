// Package codegen converts IR to pseudo-assembly text
package codegen

import (
	"fmt"
	"strings"

	"github.com/user/freelang-evolving-compiler/internal/ir"
)

// Result contains generated code and metrics
type Result struct {
	Code      string // Generated pseudo-assembly text
	ByteSize  int    // Length of generated code in bytes
	LineCount int    // Number of lines in output
}

// CodeGen generates pseudo-assembly from IR
type CodeGen struct {
	code      strings.Builder
	lineCount int
}

// New creates a new code generator
func New() *CodeGen {
	return &CodeGen{
		lineCount: 0,
	}
}

// Generate converts IR program to pseudo-assembly
func (cg *CodeGen) Generate(prog *ir.Program) Result {
	cg.code.Reset()
	cg.lineCount = 0

	// Generate functions
	for _, fn := range prog.Functions {
		cg.generateFunction(fn)
	}

	// Generate main
	if len(prog.Main) > 0 {
		cg.writeLine("; === main ===")
		cg.generateInstructions(prog.Main)
	}

	code := cg.code.String()
	return Result{
		Code:      code,
		ByteSize:  len(code),
		LineCount: cg.lineCount,
	}
}

// generateFunction generates code for a single function
func (cg *CodeGen) generateFunction(fn ir.Function) {
	cg.writeLine(fmt.Sprintf("; === function %s ===", fn.Name))
	cg.writeLine(fmt.Sprintf("ENTER %s", fn.Name))

	cg.generateInstructions(fn.Body)

	// Ensure LEAVE is present
	if len(fn.Body) == 0 || fn.Body[len(fn.Body)-1].Op != ir.OpLeave {
		cg.writeLine(fmt.Sprintf("LEAVE %s", fn.Name))
	}

	cg.writeLine("")
}

// generateInstructions generates code for a list of instructions
func (cg *CodeGen) generateInstructions(instrs []ir.Instruction) {
	for _, instr := range instrs {
		cg.generateInstruction(instr)
	}
}

// generateInstruction generates code for a single instruction
func (cg *CodeGen) generateInstruction(instr ir.Instruction) {
	switch instr.Op {
	case ir.OpConst:
		cg.writeLine(fmt.Sprintf("  LOAD %s, #%d", cg.formatOperand(instr.Dest), instr.Src1.ImmVal))

	case ir.OpCopy:
		cg.writeLine(fmt.Sprintf("  COPY %s, %s", cg.formatOperand(instr.Dest), cg.formatOperand(instr.Src1)))

	case ir.OpAdd:
		cg.writeLine(fmt.Sprintf("  ADD %s, %s, %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpSub:
		cg.writeLine(fmt.Sprintf("  SUB %s, %s, %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpMul:
		cg.writeLine(fmt.Sprintf("  MUL %s, %s, %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpDiv:
		cg.writeLine(fmt.Sprintf("  DIV %s, %s, %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpEq:
		cg.writeLine(fmt.Sprintf("  CMP %s, %s, == %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpNe:
		cg.writeLine(fmt.Sprintf("  CMP %s, %s, != %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpLt:
		cg.writeLine(fmt.Sprintf("  CMP %s, %s, < %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpGt:
		cg.writeLine(fmt.Sprintf("  CMP %s, %s, > %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpLe:
		cg.writeLine(fmt.Sprintf("  CMP %s, %s, <= %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpGe:
		cg.writeLine(fmt.Sprintf("  CMP %s, %s, >= %s",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpLabel:
		cg.writeLine(fmt.Sprintf("%s:", instr.Label))

	case ir.OpJump:
		cg.writeLine(fmt.Sprintf("  JUMP %s", instr.Label))

	case ir.OpJumpIf:
		cg.writeLine(fmt.Sprintf("  JIT %s, %s", cg.formatOperand(instr.Src1), instr.Label))

	case ir.OpJumpIfFalse:
		cg.writeLine(fmt.Sprintf("  JLF %s, %s", cg.formatOperand(instr.Src1), instr.Label))

	case ir.OpCall:
		cg.writeLine(fmt.Sprintf("  CALL %s, %s", cg.formatOperand(instr.Dest), instr.Fn))

	case ir.OpParam:
		cg.writeLine(fmt.Sprintf("  PARAM %s", cg.formatOperand(instr.Src1)))

	case ir.OpReturn:
		if instr.Src1.Name != "" || instr.Src1.IsImm {
			cg.writeLine(fmt.Sprintf("  RET %s", cg.formatOperand(instr.Src1)))
		} else {
			cg.writeLine("  RET")
		}

	case ir.OpEnter:
		cg.writeLine(fmt.Sprintf("ENTER %s", instr.Fn))

	case ir.OpLeave:
		cg.writeLine(fmt.Sprintf("LEAVE %s", instr.Fn))

	case ir.OpStructDef:
		cg.writeLine(fmt.Sprintf("; STRUCT %s size=%d", instr.Fn, instr.Src1.ImmVal))

	case ir.OpFieldLoad:
		cg.writeLine(fmt.Sprintf("  LOAD %s, [%s+%d]",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			instr.Src2.ImmVal))

	case ir.OpFieldStore:
		cg.writeLine(fmt.Sprintf("  STORE [%s+%d], %s",
			cg.formatOperand(instr.Dest),
			instr.Src1.ImmVal,
			cg.formatOperand(instr.Src2)))

	case ir.OpSyscall:
		cg.writeLine(fmt.Sprintf("  SVC %s", cg.formatOperand(instr.Dest)))

	case ir.OpArrayNew:
		cg.writeLine(fmt.Sprintf("  ARRAY_NEW %s, #%d", cg.formatOperand(instr.Dest), instr.Src1.ImmVal))

	case ir.OpArrayLoad:
		cg.writeLine(fmt.Sprintf("  LOAD_ELEM %s, %s[%s]",
			cg.formatOperand(instr.Dest),
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2)))

	case ir.OpArrayStore:
		cg.writeLine(fmt.Sprintf("  STORE_ELEM %s[%s], %s",
			cg.formatOperand(instr.Src1),
			cg.formatOperand(instr.Src2),
			cg.formatOperand(instr.Dest)))

	case ir.OpNoop:
		// Skip noop instructions
	}
}

// formatOperand converts an operand to text representation
func (cg *CodeGen) formatOperand(op ir.Operand) string {
	if op.IsImm {
		return fmt.Sprintf("#%d", op.ImmVal)
	}
	if op.IsLabel {
		return op.Name
	}
	return op.Name
}

// writeLine appends a line to the output
func (cg *CodeGen) writeLine(line string) {
	cg.code.WriteString(line)
	cg.code.WriteString("\n")
	cg.lineCount++
}
