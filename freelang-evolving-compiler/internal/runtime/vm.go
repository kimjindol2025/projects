package runtime

import (
	"fmt"

	"github.com/user/freelang-evolving-compiler/internal/builtin"
	"github.com/user/freelang-evolving-compiler/internal/ir"
)

// VM represents the FreeLang virtual machine
type VM struct {
	program    *ir.Program
	fnIndex    map[string]*ir.Function
	callStack  []*CallFrame
	paramQueue []Value
	globals    map[string]Value
}

// New creates a new VM for the given IR program
func New(prog *ir.Program) *VM {
	vm := &VM{
		program:    prog,
		fnIndex:    make(map[string]*ir.Function),
		callStack:  make([]*CallFrame, 0),
		paramQueue: make([]Value, 0),
		globals:    make(map[string]Value),
	}

	// Build function index
	for i := range prog.Functions {
		vm.fnIndex[prog.Functions[i].Name] = &prog.Functions[i]
	}

	return vm
}

// Run executes the program and returns the result
func (vm *VM) Run() (Value, error) {
	// Build label map for main
	labelMap := buildLabelMap(vm.program.Main)

	// Execute main instructions
	_, err := vm.execInstrs(vm.program.Main, labelMap, nil)
	if err != nil {
		return NilVal(), err
	}

	return NilVal(), nil
}

// DumpGlobals returns the global variables (for testing)
func (vm *VM) DumpGlobals() map[string]Value {
	return vm.globals
}

// execInstrs executes a sequence of instructions and returns the result
func (vm *VM) execInstrs(instrs []ir.Instruction, labelMap map[string]int, frame *CallFrame) (Value, error) {
	pc := 0
	var result Value

	for pc < len(instrs) {
		instr := instrs[pc]
		pc++

		switch instr.Op {
		case ir.OpNoop:
			// No operation

		case ir.OpConst:
			// Load constant
			if instr.Src1.IsImm {
				result = IntVal(instr.Src1.ImmVal)
				vm.storeResult(instr.Dest, result, frame)
			}

		case ir.OpCopy:
			// Copy value
			src := vm.loadOperand(instr.Src1, frame)
			vm.storeResult(instr.Dest, src, frame)

		case ir.OpAdd, ir.OpSub, ir.OpMul, ir.OpDiv:
			left := vm.loadOperand(instr.Src1, frame)
			right := vm.loadOperand(instr.Src2, frame)

			// String concatenation
			if instr.Op == ir.OpAdd && left.Kind == KindString && right.Kind == KindString {
				result = StringVal(left.SVal + right.SVal)
				vm.storeResult(instr.Dest, result, frame)
				break
			}

			// Integer arithmetic
			if left.Kind != KindInt || right.Kind != KindInt {
				return NilVal(), fmt.Errorf("arithmetic operands must be int")
			}

			var resultVal int64
			switch instr.Op {
			case ir.OpAdd:
				resultVal = left.IVal + right.IVal
			case ir.OpSub:
				resultVal = left.IVal - right.IVal
			case ir.OpMul:
				resultVal = left.IVal * right.IVal
			case ir.OpDiv:
				if right.IVal == 0 {
					return NilVal(), fmt.Errorf("division by zero")
				}
				resultVal = left.IVal / right.IVal
			}
			result = IntVal(resultVal)
			vm.storeResult(instr.Dest, result, frame)

		case ir.OpEq, ir.OpNe, ir.OpLt, ir.OpGt, ir.OpLe, ir.OpGe:
			left := vm.loadOperand(instr.Src1, frame)
			right := vm.loadOperand(instr.Src2, frame)

			var cmpResult bool

			// String comparison
			if left.Kind == KindString && right.Kind == KindString {
				switch instr.Op {
				case ir.OpEq:
					cmpResult = left.SVal == right.SVal
				case ir.OpNe:
					cmpResult = left.SVal != right.SVal
				case ir.OpLt:
					cmpResult = left.SVal < right.SVal
				case ir.OpGt:
					cmpResult = left.SVal > right.SVal
				case ir.OpLe:
					cmpResult = left.SVal <= right.SVal
				case ir.OpGe:
					cmpResult = left.SVal >= right.SVal
				}
				result = BoolVal(cmpResult)
				vm.storeResult(instr.Dest, result, frame)
				break
			}

			// Bool comparison
			if left.Kind == KindBool && right.Kind == KindBool {
				switch instr.Op {
				case ir.OpEq:
					cmpResult = left.BVal == right.BVal
				case ir.OpNe:
					cmpResult = left.BVal != right.BVal
				default:
					return NilVal(), fmt.Errorf("invalid comparison operator for bool")
				}
				result = BoolVal(cmpResult)
				vm.storeResult(instr.Dest, result, frame)
				break
			}

			// Integer comparison (default)
			if left.Kind != KindInt || right.Kind != KindInt {
				return NilVal(), fmt.Errorf("comparison operands must be same type (int, string, or bool)")
			}

			switch instr.Op {
			case ir.OpEq:
				cmpResult = left.IVal == right.IVal
			case ir.OpNe:
				cmpResult = left.IVal != right.IVal
			case ir.OpLt:
				cmpResult = left.IVal < right.IVal
			case ir.OpGt:
				cmpResult = left.IVal > right.IVal
			case ir.OpLe:
				cmpResult = left.IVal <= right.IVal
			case ir.OpGe:
				cmpResult = left.IVal >= right.IVal
			}
			result = BoolVal(cmpResult)
			vm.storeResult(instr.Dest, result, frame)

		case ir.OpLabel:
			// Label definition (no-op, already in labelMap)

		case ir.OpJump:
			if idx, ok := labelMap[instr.Label]; ok {
				pc = idx
			} else {
				return NilVal(), fmt.Errorf("undefined label: %s", instr.Label)
			}

		case ir.OpJumpIf:
			condition := vm.loadOperand(instr.Src1, frame)
			if condition.Truthy() {
				if idx, ok := labelMap[instr.Label]; ok {
					pc = idx
				} else {
					return NilVal(), fmt.Errorf("undefined label: %s", instr.Label)
				}
			}

		case ir.OpJumpIfFalse:
			condition := vm.loadOperand(instr.Src1, frame)
			if !condition.Truthy() {
				if idx, ok := labelMap[instr.Label]; ok {
					pc = idx
				} else {
					return NilVal(), fmt.Errorf("undefined label: %s", instr.Label)
				}
			}

		case ir.OpParam:
			arg := vm.loadOperand(instr.Src1, frame)
			vm.paramQueue = append(vm.paramQueue, arg)

		case ir.OpCall:
			args := vm.drainParamQueue()
			retVal, err := vm.callFunction(instr.Fn, args)
			if err != nil {
				return NilVal(), err
			}
			vm.storeResult(instr.Dest, retVal, frame)

		case ir.OpReturn:
			if instr.Src1.IsImm || instr.Src1.IsStr || instr.Src1.IsBool || instr.Src1.Name != "" {
				return vm.loadOperand(instr.Src1, frame), nil
			}
			return NilVal(), nil

		case ir.OpEnter, ir.OpLeave:
			// No-op (callFunction handles frame management)

		case ir.OpStructDef:
			// No-op (type definition, no runtime effect)

		case ir.OpFieldLoad:
			obj := vm.loadOperand(instr.Src1, frame)
			if obj.Kind != KindStruct {
				return NilVal(), fmt.Errorf("field access on non-struct")
			}
			offset := instr.Src2.ImmVal / 8
			fieldKey := fmt.Sprintf("f%d", offset)
			if val, ok := obj.Fields[fieldKey]; ok {
				vm.storeResult(instr.Dest, val, frame)
			} else {
				vm.storeResult(instr.Dest, NilVal(), frame)
			}

		case ir.OpFieldStore:
			obj := vm.loadOperand(instr.Src1, frame)
			if obj.Kind != KindStruct {
				return NilVal(), fmt.Errorf("field store on non-struct")
			}
			offset := instr.Src2.ImmVal / 8
			fieldKey := fmt.Sprintf("f%d", offset)
			value := vm.loadOperand(instr.Dest, frame)
			obj.Fields[fieldKey] = value
			if frame != nil {
				frame.Set(instr.Src1.Name, obj)
			} else {
				vm.globals[instr.Src1.Name] = obj
			}

		case ir.OpSyscall:
			// No-op (not implemented)

		case ir.OpArrayNew:
			count := int(instr.Src1.ImmVal)
			args := vm.drainParamQueue()
			elems := make([]Value, count)
			for i := 0; i < count && i < len(args); i++ {
				elems[i] = args[i]
			}
			result = ArrayVal(elems)
			vm.storeResult(instr.Dest, result, frame)

		case ir.OpArrayLoad:
			arr := vm.loadOperand(instr.Src1, frame)
			idx := vm.loadOperand(instr.Src2, frame)
			if arr.Kind != KindArray {
				return NilVal(), fmt.Errorf("index on non-array")
			}
			i := int(idx.IVal)
			if i < 0 || i >= len(arr.Elems) {
				return NilVal(), fmt.Errorf("index out of bounds: %d", i)
			}
			result = arr.Elems[i]
			vm.storeResult(instr.Dest, result, frame)

		case ir.OpArrayStore:
			arr := vm.loadOperand(instr.Src1, frame)
			idx := vm.loadOperand(instr.Src2, frame)
			val := vm.loadOperand(instr.Dest, frame)
			if arr.Kind != KindArray {
				return NilVal(), fmt.Errorf("field store on non-array")
			}
			i := int(idx.IVal)
			if i < 0 || i >= len(arr.Elems) {
				return NilVal(), fmt.Errorf("index out of bounds: %d", i)
			}
			arr.Elems[i] = val
			vm.storeResult(instr.Src1, arr, frame)

		default:
			return NilVal(), fmt.Errorf("unknown opcode: %v", instr.Op)
		}
	}

	return result, nil
}

// callFunction calls either a builtin or user-defined function
func (vm *VM) callFunction(name string, args []Value) (Value, error) {
	// Check for builtin function
	if builtin.IsBuiltin(name) {
		def, _ := builtin.Lookup(name)
		rawArgs := make([]interface{}, len(args))
		for i, a := range args {
			rawArgs[i] = a.ToInterface()
		}
		result := def.Impl(rawArgs...)
		return FromInterface(result), nil
	}

	// Check for user-defined function
	fn, ok := vm.fnIndex[name]
	if !ok {
		return NilVal(), fmt.Errorf("undefined function: %s", name)
	}

	// Create new frame and push to call stack
	newFrame := NewFrame(fn.Name, fn.Params, args)
	vm.callStack = append(vm.callStack, newFrame)
	defer func() {
		vm.callStack = vm.callStack[:len(vm.callStack)-1]
	}()

	// Build label map for function body
	labelMap := buildLabelMap(fn.Body)

	// Execute function body
	retVal, err := vm.execInstrs(fn.Body, labelMap, newFrame)
	if err != nil {
		return NilVal(), err
	}

	return retVal, nil
}

// loadOperand loads a value from an operand
func (vm *VM) loadOperand(op ir.Operand, frame *CallFrame) Value {
	// Immediate integer value
	if op.IsImm {
		return IntVal(op.ImmVal)
	}

	// Immediate string value
	if op.IsStr {
		return StringVal(op.SVal)
	}

	// Immediate bool value
	if op.IsBool {
		return BoolVal(op.BVal)
	}

	// Named value (variable or temporary)
	if op.Name != "" {
		// Try frame first (if in function)
		if frame != nil {
			if v, ok := frame.Get(op.Name); ok {
				return v
			}
		}

		// Try globals
		if v, ok := vm.globals[op.Name]; ok {
			return v
		}

		return NilVal()
	}

	return NilVal()
}

// storeResult stores a value to a destination operand
func (vm *VM) storeResult(dest ir.Operand, v Value, frame *CallFrame) {
	if dest.Name == "" {
		return
	}

	// Store in frame (if in function)
	if frame != nil {
		frame.Set(dest.Name, v)
	} else {
		// Store in globals
		vm.globals[dest.Name] = v
	}
}

// buildLabelMap builds a map from label names to instruction indices
func buildLabelMap(instrs []ir.Instruction) map[string]int {
	m := make(map[string]int)
	for i, instr := range instrs {
		if instr.Op == ir.OpLabel {
			m[instr.Label] = i
		}
	}
	return m
}

// drainParamQueue returns and clears the parameter queue
func (vm *VM) drainParamQueue() []Value {
	args := make([]Value, len(vm.paramQueue))
	copy(args, vm.paramQueue)
	vm.paramQueue = vm.paramQueue[:0]
	return args
}
