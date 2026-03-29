// Package ir implements IR generation from optimized AST
package ir

import (
	"fmt"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

// Generator converts AST to intermediate representation
type Generator struct {
	tempCount    int
	labelCount   int
	currentFn    *Function
	program      *Program
	structFields map[string][]string
}

// NewGenerator creates a new IR generator
func NewGenerator() *Generator {
	return &Generator{
		tempCount:  0,
		labelCount: 0,
		program: &Program{
			Functions: []Function{},
			Main:      []Instruction{},
		},
		structFields: make(map[string][]string),
	}
}

// Generate converts an AST to an IR program
func (g *Generator) Generate(root *ast.Node) (*Program, error) {
	if root == nil {
		return g.program, nil
	}

	// Process program statements
	if root.Kind == ast.NodeProgram {
		for _, child := range root.Children {
			if child.Kind == ast.NodeFnDecl {
				if err := g.genFnDecl(child); err != nil {
					return nil, err
				}
			} else if child.Kind == ast.NodeStructDecl {
				if err := g.genStructDecl(child); err != nil {
					return nil, err
				}
			} else {
				if err := g.genStmt(child); err != nil {
					return nil, err
				}
			}
		}
	} else {
		if err := g.genStmt(root); err != nil {
			return nil, err
		}
	}

	return g.program, nil
}

// genFnDecl generates IR for a function declaration
func (g *Generator) genFnDecl(node *ast.Node) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("function declaration must have name and body")
	}

	fnName := node.Children[0].Value
	body := node.Children[len(node.Children)-1]

	fn := Function{
		Name:   fnName,
		Params: []string{},
		Body:   []Instruction{},
	}

	// Extract parameter names if they exist
	if len(node.Children) > 2 {
		for i := 1; i < len(node.Children)-1; i++ {
			if node.Children[i].Kind == ast.NodeIdent {
				fn.Params = append(fn.Params, node.Children[i].Value)
			}
		}
	}

	// Save current function and switch context
	prevFn := g.currentFn
	g.currentFn = &fn

	// Generate OpEnter
	g.emit(Instruction{Op: OpEnter, Fn: fnName})

	// Generate body
	if body.Kind == ast.NodeBlockStmt {
		for _, stmt := range body.Children {
			if err := g.genStmt(stmt); err != nil {
				return err
			}
		}
	} else {
		if err := g.genStmt(body); err != nil {
			return err
		}
	}

	// Generate OpLeave if not already there
	g.emit(Instruction{Op: OpLeave, Fn: fnName})

	// Add function to program
	g.program.Functions = append(g.program.Functions, fn)
	g.currentFn = prevFn

	return nil
}

// genStmt generates IR for a statement
func (g *Generator) genStmt(node *ast.Node) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.NodeLetDecl:
		return g.genLetDecl(node)
	case ast.NodeIfStmt:
		return g.genIfStmt(node)
	case ast.NodeForStmt:
		return g.genForStmt(node)
	case ast.NodeReturn:
		return g.genReturn(node)
	case ast.NodeStructDecl:
		return g.genStructDecl(node)
	case ast.NodeBlockStmt:
		for _, child := range node.Children {
			if err := g.genStmt(child); err != nil {
				return err
			}
		}
		return nil
	default:
		// Expression statement
		_, err := g.genExpr(node)
		return err
	}
}

// genLetDecl generates IR for a let declaration
func (g *Generator) genLetDecl(node *ast.Node) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("let declaration must have name and value")
	}

	varName := node.Children[0].Value
	valueExpr := node.Children[1]

	// Generate code for the expression
	operand, err := g.genExpr(valueExpr)
	if err != nil {
		return err
	}

	// Copy result to variable
	g.emit(Instruction{
		Op:   OpCopy,
		Dest: Operand{Name: varName},
		Src1: operand,
	})

	return nil
}

// genIfStmt generates IR for an if statement
func (g *Generator) genIfStmt(node *ast.Node) error {
	if len(node.Children) < 2 {
		return fmt.Errorf("if statement must have condition and body")
	}

	condExpr := node.Children[0]
	bodyStmt := node.Children[1]

	// Generate condition
	condOp, err := g.genExpr(condExpr)
	if err != nil {
		return err
	}

	// Create end label
	endLabel := g.newLabel("end")

	// Jump if false
	g.emit(Instruction{
		Op:    OpJumpIfFalse,
		Src1:  condOp,
		Label: endLabel,
	})

	// Generate body
	if err := g.genStmt(bodyStmt); err != nil {
		return err
	}

	// Emit end label
	g.emit(Instruction{
		Op:    OpLabel,
		Label: endLabel,
	})

	return nil
}

// genForStmt generates IR for a for statement
func (g *Generator) genForStmt(node *ast.Node) error {
	if len(node.Children) < 3 {
		return fmt.Errorf("for statement must have iterator, range, and body")
	}

	iterName := node.Children[0].Value
	rangeExpr := node.Children[1]
	bodyStmt := node.Children[2]

	// Extract range start and end
	if rangeExpr.Kind != ast.NodeBinaryExpr || rangeExpr.Value != ".." {
		return fmt.Errorf("for loop requires range expression (..)")
	}

	if len(rangeExpr.Children) < 2 {
		return fmt.Errorf("range expression requires start and end")
	}

	// 상수 범위 감지: 언롤링 시도
	startNode := rangeExpr.Children[0]
	endNode := rangeExpr.Children[1]
	if sv, ok1 := isConstIntLit(startNode); ok1 {
		if ev, ok2 := isConstIntLit(endNode); ok2 {
			count := ev - sv
			if count >= 0 && count <= loopUnrollThreshold {
				return g.genUnrolledFor(iterName, sv, ev, bodyStmt)
			}
		}
	}

	startOp, err := g.genExpr(rangeExpr.Children[0])
	if err != nil {
		return err
	}

	endOp, err := g.genExpr(rangeExpr.Children[1])
	if err != nil {
		return err
	}

	// Initialize iterator
	g.emit(Instruction{
		Op:   OpCopy,
		Dest: Operand{Name: iterName},
		Src1: startOp,
	})

	// Create loop and end labels
	loopLabel := g.newLabel("loop")
	endLabel := g.newLabel("end")

	// Emit loop label
	g.emit(Instruction{
		Op:    OpLabel,
		Label: loopLabel,
	})

	// Check condition: iterator < end
	cmpTemp := g.newTemp()
	g.emit(Instruction{
		Op:   OpLt,
		Dest: cmpTemp,
		Src1: Operand{Name: iterName},
		Src2: endOp,
	})

	// Jump if false (exit loop)
	g.emit(Instruction{
		Op:    OpJumpIfFalse,
		Src1:  cmpTemp,
		Label: endLabel,
	})

	// Generate body
	if err := g.genStmt(bodyStmt); err != nil {
		return err
	}

	// Increment iterator
	incrTemp := g.newTemp()
	g.emit(Instruction{
		Op:   OpAdd,
		Dest: incrTemp,
		Src1: Operand{Name: iterName},
		Src2: Operand{IsImm: true, ImmVal: 1},
	})
	g.emit(Instruction{
		Op:   OpCopy,
		Dest: Operand{Name: iterName},
		Src1: incrTemp,
	})

	// Jump back to loop
	g.emit(Instruction{
		Op:    OpJump,
		Label: loopLabel,
	})

	// Emit end label
	g.emit(Instruction{
		Op:    OpLabel,
		Label: endLabel,
	})

	return nil
}

// genReturn generates IR for a return statement
func (g *Generator) genReturn(node *ast.Node) error {
	if len(node.Children) == 0 {
		g.emit(Instruction{Op: OpReturn})
		return nil
	}

	returnOp, err := g.genExpr(node.Children[0])
	if err != nil {
		return err
	}

	g.emit(Instruction{
		Op:   OpReturn,
		Src1: returnOp,
	})

	return nil
}

// genExpr generates IR for an expression and returns the operand
func (g *Generator) genExpr(node *ast.Node) (Operand, error) {
	if node == nil {
		return Operand{}, fmt.Errorf("nil expression")
	}

	switch node.Kind {
	case ast.NodeIntLit:
		// Parse integer literal
		var val int64
		fmt.Sscanf(node.Value, "%d", &val)
		return Operand{IsImm: true, ImmVal: val}, nil

	case ast.NodeStringLit:
		// String literal
		return Operand{IsStr: true, SVal: node.Value}, nil

	case ast.NodeBoolLit:
		// Boolean literal
		return Operand{IsBool: true, BVal: node.Value == "true"}, nil

	case ast.NodeIdent:
		return Operand{Name: node.Value}, nil

	case ast.NodeBinaryExpr:
		return g.genBinaryExpr(node)

	case ast.NodeUnaryExpr:
		return g.genUnaryExpr(node)

	case ast.NodeCallExpr:
		return g.genCallExpr(node)

	case ast.NodeFieldAccess:
		return g.genFieldAccess(node)

	case ast.NodeArrayLit:
		return g.genArrayLit(node)

	case ast.NodeIndexExpr:
		return g.genIndexExpr(node)

	default:
		return Operand{}, fmt.Errorf("unsupported expression kind: %d", node.Kind)
	}
}

// genBinaryExpr generates IR for a binary expression
func (g *Generator) genBinaryExpr(node *ast.Node) (Operand, error) {
	if len(node.Children) < 2 {
		return Operand{}, fmt.Errorf("binary expression requires 2 operands")
	}

	left, err := g.genExpr(node.Children[0])
	if err != nil {
		return Operand{}, err
	}

	right, err := g.genExpr(node.Children[1])
	if err != nil {
		return Operand{}, err
	}

	result := g.newTemp()
	op := g.opcodeFromOp(node.Value)

	g.emit(Instruction{
		Op:   op,
		Dest: result,
		Src1: left,
		Src2: right,
	})

	return result, nil
}

// genUnaryExpr generates IR for a unary expression (!, -)
func (g *Generator) genUnaryExpr(node *ast.Node) (Operand, error) {
	if len(node.Children) < 1 {
		return Operand{}, fmt.Errorf("unary expression requires 1 operand")
	}

	operand, err := g.genExpr(node.Children[0])
	if err != nil {
		return Operand{}, err
	}

	result := g.newTemp()
	op := node.Value

	switch op {
	case "!":
		// Logical NOT
		g.emit(Instruction{
			Op:   OpNot,
			Dest: result,
			Src1: operand,
		})
	case "-":
		// Arithmetic negation: negate by subtracting from 0
		g.emit(Instruction{
			Op:   OpSub,
			Dest: result,
			Src1: Operand{IsImm: true, ImmVal: 0},
			Src2: operand,
		})
	default:
		return Operand{}, fmt.Errorf("unsupported unary operator: %s", op)
	}

	return result, nil
}

// genUnrolledFor inlines a constant-range loop without labels or jumps
func (g *Generator) genUnrolledFor(iterName string, start, end int64, body *ast.Node) error {
	for i := start; i < end; i++ {
		g.emit(Instruction{
			Op:   OpCopy,
			Dest: Operand{Name: iterName},
			Src1: Operand{IsImm: true, ImmVal: i},
		})
		if err := g.genStmt(body); err != nil {
			return err
		}
	}
	return nil
}

// genCallExpr generates IR for a function call
func (g *Generator) genCallExpr(node *ast.Node) (Operand, error) {
	if len(node.Children) == 0 {
		return Operand{}, fmt.Errorf("call expression requires function name")
	}

	fnName := node.Children[0].Value

	// Special handling for syscall
	if fnName == "syscall" {
		if len(node.Children) < 2 {
			return Operand{}, fmt.Errorf("syscall requires at least syscall number")
		}

		// syscall number (always first arg after function name)
		numOp, err := g.genExpr(node.Children[1])
		if err != nil {
			return Operand{}, err
		}

		// Additional args (starting from index 2)
		for i := 2; i < len(node.Children); i++ {
			argOp, err := g.genExpr(node.Children[i])
			if err != nil {
				return Operand{}, err
			}
			g.emit(Instruction{
				Op:   OpParam,
				Src1: argOp,
			})
		}

		// Emit syscall instruction
		result := g.newTemp()
		g.emit(Instruction{
			Op:   OpSyscall,
			Dest: result,
			Src1: numOp,
		})

		return result, nil
	}

	// Emit parameters for regular function calls
	for i := 1; i < len(node.Children); i++ {
		argOp, err := g.genExpr(node.Children[i])
		if err != nil {
			return Operand{}, err
		}
		g.emit(Instruction{
			Op:   OpParam,
			Src1: argOp,
		})
	}

	// Emit call
	result := g.newTemp()
	g.emit(Instruction{
		Op:   OpCall,
		Dest: result,
		Fn:   fnName,
	})

	return result, nil
}

// Helper functions

func (g *Generator) newTemp() Operand {
	temp := Operand{
		IsTemp: true,
		Name:   fmt.Sprintf("t%d", g.tempCount),
	}
	g.tempCount++
	return temp
}

func (g *Generator) newLabel(prefix string) string {
	label := fmt.Sprintf("L_%s_%d", prefix, g.labelCount)
	g.labelCount++
	return label
}

func (g *Generator) emit(instr Instruction) {
	if g.currentFn != nil {
		g.currentFn.Body = append(g.currentFn.Body, instr)
	} else {
		g.program.Main = append(g.program.Main, instr)
	}
}

func (g *Generator) opcodeFromOp(op string) Opcode {
	switch op {
	case "+":
		return OpAdd
	case "-":
		return OpSub
	case "*":
		return OpMul
	case "/":
		return OpDiv
	case "==":
		return OpEq
	case "!=":
		return OpNe
	case "<":
		return OpLt
	case ">":
		return OpGt
	case "<=":
		return OpLe
	case ">=":
		return OpGe
	case "&&":
		return OpAnd
	case "||":
		return OpOr
	default:
		return OpNoop
	}
}

// genStructDecl generates IR for a struct declaration
func (g *Generator) genStructDecl(node *ast.Node) error {
	if node.Value == "" {
		return fmt.Errorf("struct declaration must have a name")
	}

	name := node.Value
	fields := []string{}

	// Extract field names from children
	for _, field := range node.Children {
		if field.Kind == ast.NodeFieldDecl {
			fields = append(fields, field.Value)
		}
	}

	// Store struct definition
	g.structFields[name] = fields

	// Calculate struct size (8 bytes per field)
	size := len(fields) * 8

	// Save current function context
	prevFn := g.currentFn
	g.currentFn = nil // Emit to Main context

	// Emit OpStructDef instruction
	g.emit(Instruction{
		Op:   OpStructDef,
		Fn:   name,
		Src1: Operand{IsImm: true, ImmVal: int64(size)},
	})

	// Restore function context
	g.currentFn = prevFn

	return nil
}

// genFieldAccess generates IR for a field access expression (obj.field)
func (g *Generator) genFieldAccess(node *ast.Node) (Operand, error) {
	if len(node.Children) == 0 {
		return Operand{}, fmt.Errorf("field access requires object expression")
	}

	// Generate IR for object expression
	objOp, err := g.genExpr(node.Children[0])
	if err != nil {
		return Operand{}, err
	}

	fieldName := node.Value

	// Search for field offset in known structs
	var offset int64
	found := false
	for _, fields := range g.structFields {
		for i, f := range fields {
			if f == fieldName {
				offset = int64(i * 8)
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	// If not found, use offset 0 as fallback

	// Emit OpFieldLoad instruction
	result := g.newTemp()
	g.emit(Instruction{
		Op:   OpFieldLoad,
		Dest: result,
		Src1: objOp,
		Src2: Operand{IsImm: true, ImmVal: offset},
	})

	return result, nil
}

// genArrayLit generates IR for array literals: [1, 2, 3]
func (g *Generator) genArrayLit(node *ast.Node) (Operand, error) {
	// Generate IR for each element and emit as parameter
	for _, elem := range node.Children {
		op, err := g.genExpr(elem)
		if err != nil {
			return Operand{}, err
		}
		g.emit(Instruction{
			Op:   OpParam,
			Src1: op,
		})
	}

	// Create array with element count
	result := g.newTemp()
	g.emit(Instruction{
		Op:   OpArrayNew,
		Dest: result,
		Src1: Operand{IsImm: true, ImmVal: int64(len(node.Children))},
	})

	return result, nil
}

// genIndexExpr generates IR for array indexing: arr[i]
func (g *Generator) genIndexExpr(node *ast.Node) (Operand, error) {
	if len(node.Children) < 2 {
		return Operand{}, fmt.Errorf("index expression requires array and index")
	}

	arrOp, err := g.genExpr(node.Children[0])
	if err != nil {
		return Operand{}, err
	}

	idxOp, err := g.genExpr(node.Children[1])
	if err != nil {
		return Operand{}, err
	}

	result := g.newTemp()
	g.emit(Instruction{
		Op:   OpArrayLoad,
		Dest: result,
		Src1: arrOp,
		Src2: idxOp,
	})

	return result, nil
}

const loopUnrollThreshold = 8

// isConstIntLit checks if a node is a NodeIntLit and returns its int64 value
func isConstIntLit(node *ast.Node) (int64, bool) {
	if node.Kind != ast.NodeIntLit {
		return 0, false
	}
	var val int64
	fmt.Sscanf(node.Value, "%d", &val)
	return val, true
}
