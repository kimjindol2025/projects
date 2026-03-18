package codegen

import (
	"fmt"
)

// VM is a simple stack-based virtual machine
type VM struct {
	bytecode  *Bytecode
	stack     []interface{}
	variables map[string]interface{}
	pc        int // Program counter
}

// NewVM creates a new virtual machine
func NewVM(bytecode *Bytecode) *VM {
	return &VM{
		bytecode:  bytecode,
		stack:     []interface{}{},
		variables: make(map[string]interface{}),
		pc:        0,
	}
}

// Run executes the bytecode
func (vm *VM) Run() (interface{}, error) {
	for vm.pc < len(vm.bytecode.Code) {
		opByte := vm.bytecode.Code[vm.pc]
		op := BytecodeOp(opByte)

		switch op {
		case OpPush:
			if err := vm.executePush(); err != nil {
				return nil, err
			}
		case OpAdd:
			if err := vm.executeBinOp(OpAdd); err != nil {
				return nil, err
			}
		case OpSub:
			if err := vm.executeBinOp(OpSub); err != nil {
				return nil, err
			}
		case OpMul:
			if err := vm.executeBinOp(OpMul); err != nil {
				return nil, err
			}
		case OpDiv:
			if err := vm.executeBinOp(OpDiv); err != nil {
				return nil, err
			}
		case OpMod:
			if err := vm.executeBinOp(OpMod); err != nil {
				return nil, err
			}
		case OpEq:
			if err := vm.executeBinOp(OpEq); err != nil {
				return nil, err
			}
		case OpNe:
			if err := vm.executeBinOp(OpNe); err != nil {
				return nil, err
			}
		case OpLt:
			if err := vm.executeBinOp(OpLt); err != nil {
				return nil, err
			}
		case OpLe:
			if err := vm.executeBinOp(OpLe); err != nil {
				return nil, err
			}
		case OpGt:
			if err := vm.executeBinOp(OpGt); err != nil {
				return nil, err
			}
		case OpGe:
			if err := vm.executeBinOp(OpGe); err != nil {
				return nil, err
			}
		case OpAnd:
			if err := vm.executeBinOp(OpAnd); err != nil {
				return nil, err
			}
		case OpOr:
			if err := vm.executeBinOp(OpOr); err != nil {
				return nil, err
			}
		case OpNot:
			if err := vm.executeNot(); err != nil {
				return nil, err
			}
		case OpNeg:
			if err := vm.executeNeg(); err != nil {
				return nil, err
			}
		case OpCall:
			if err := vm.executeCall(); err != nil {
				return nil, err
			}
		case OpRet:
			if len(vm.stack) > 0 {
				return vm.stack[len(vm.stack)-1], nil
			}
			return nil, nil
		case OpLoad:
			if err := vm.executeLoad(); err != nil {
				return nil, err
			}
		case OpStore:
			if err := vm.executeStore(); err != nil {
				return nil, err
			}
		case OpHalt:
			if len(vm.stack) > 0 {
				return vm.stack[len(vm.stack)-1], nil
			}
			return nil, nil
		default:
			return nil, fmt.Errorf("unknown opcode: %d", op)
		}

		vm.pc++
	}

	if len(vm.stack) > 0 {
		return vm.stack[len(vm.stack)-1], nil
	}
	return nil, nil
}

// executePush pushes a constant onto the stack
func (vm *VM) executePush() error {
	vm.pc++
	if vm.pc >= len(vm.bytecode.Code) {
		return fmt.Errorf("invalid push instruction")
	}

	constIdx := vm.bytecode.Code[vm.pc]
	if int(constIdx) >= len(vm.bytecode.Constants) {
		return fmt.Errorf("constant index out of bounds: %d", constIdx)
	}

	vm.stack = append(vm.stack, vm.bytecode.Constants[constIdx])
	return nil
}

// executeBinOp executes a binary operation
func (vm *VM) executeBinOp(op BytecodeOp) error {
	if len(vm.stack) < 2 {
		return fmt.Errorf("insufficient stack for binary op")
	}

	right := vm.stack[len(vm.stack)-1]
	left := vm.stack[len(vm.stack)-2]
	vm.stack = vm.stack[:len(vm.stack)-2]

	var result interface{}
	var err error

	switch op {
	case OpAdd:
		result, err = vm.add(left, right)
	case OpSub:
		result, err = vm.sub(left, right)
	case OpMul:
		result, err = vm.mul(left, right)
	case OpDiv:
		result, err = vm.div(left, right)
	case OpMod:
		result, err = vm.mod(left, right)
	case OpEq:
		result = vm.eq(left, right)
	case OpNe:
		result = !vm.eq(left, right).(bool)
	case OpLt:
		result, err = vm.lt(left, right)
	case OpLe:
		result, err = vm.le(left, right)
	case OpGt:
		result, err = vm.gt(left, right)
	case OpGe:
		result, err = vm.ge(left, right)
	case OpAnd:
		result = vm.toBool(left) && vm.toBool(right)
	case OpOr:
		result = vm.toBool(left) || vm.toBool(right)
	}

	if err != nil {
		return err
	}

	vm.stack = append(vm.stack, result)
	return nil
}

// executeNot executes logical NOT
func (vm *VM) executeNot() error {
	if len(vm.stack) < 1 {
		return fmt.Errorf("insufficient stack for not")
	}

	val := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	vm.stack = append(vm.stack, !vm.toBool(val))
	return nil
}

// executeNeg executes negation
func (vm *VM) executeNeg() error {
	if len(vm.stack) < 1 {
		return fmt.Errorf("insufficient stack for negate")
	}

	val := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]

	switch v := val.(type) {
	case int64:
		vm.stack = append(vm.stack, -v)
	case float64:
		vm.stack = append(vm.stack, -v)
	default:
		return fmt.Errorf("cannot negate %T", val)
	}

	return nil
}

// executeCall executes a function call
func (vm *VM) executeCall() error {
	vm.pc++
	if vm.pc >= len(vm.bytecode.Code) {
		return fmt.Errorf("invalid call instruction")
	}

	constIdx := vm.bytecode.Code[vm.pc]
	if int(constIdx) >= len(vm.bytecode.Constants) {
		return fmt.Errorf("constant index out of bounds")
	}

	fnName, ok := vm.bytecode.Constants[constIdx].(string)
	if !ok {
		return fmt.Errorf("function name is not string")
	}

	// Handle built-in functions
	return vm.callBuiltin(fnName)
}

// executeLoad executes a load instruction
func (vm *VM) executeLoad() error {
	vm.pc++
	if vm.pc >= len(vm.bytecode.Code) {
		return fmt.Errorf("invalid load instruction")
	}

	constIdx := vm.bytecode.Code[vm.pc]
	if int(constIdx) >= len(vm.bytecode.Constants) {
		return fmt.Errorf("constant index out of bounds")
	}

	varName, ok := vm.bytecode.Constants[constIdx].(string)
	if !ok {
		return fmt.Errorf("variable name is not string")
	}

	val, ok := vm.variables[varName]
	if !ok {
		return fmt.Errorf("undefined variable: %s", varName)
	}

	vm.stack = append(vm.stack, val)
	return nil
}

// executeStore executes a store instruction
func (vm *VM) executeStore() error {
	vm.pc++
	if vm.pc >= len(vm.bytecode.Code) {
		return fmt.Errorf("invalid store instruction")
	}

	constIdx := vm.bytecode.Code[vm.pc]
	if int(constIdx) >= len(vm.bytecode.Constants) {
		return fmt.Errorf("constant index out of bounds")
	}

	varName, ok := vm.bytecode.Constants[constIdx].(string)
	if !ok {
		return fmt.Errorf("variable name is not string")
	}

	if len(vm.stack) < 1 {
		return fmt.Errorf("insufficient stack for store")
	}

	val := vm.stack[len(vm.stack)-1]
	vm.variables[varName] = val
	return nil
}

// callBuiltin executes a built-in function
func (vm *VM) callBuiltin(fnName string) error {
	switch fnName {
	case "print":
		if len(vm.stack) < 1 {
			return fmt.Errorf("print requires 1 argument")
		}
		val := vm.stack[len(vm.stack)-1]
		fmt.Println(val)
		vm.stack = vm.stack[:len(vm.stack)-1]
		vm.stack = append(vm.stack, nil)
		return nil
	case "length":
		if len(vm.stack) < 1 {
			return fmt.Errorf("length requires 1 argument")
		}
		val := vm.stack[len(vm.stack)-1]
		vm.stack = vm.stack[:len(vm.stack)-1]

		switch v := val.(type) {
		case string:
			vm.stack = append(vm.stack, int64(len(v)))
		default:
			return fmt.Errorf("cannot get length of %T", val)
		}
		return nil
	default:
		return fmt.Errorf("unknown built-in function: %s", fnName)
	}
}

// Arithmetic operations

func (vm *VM) add(left, right interface{}) (interface{}, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l + r, nil
		case float64:
			return float64(l) + r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l + float64(r), nil
		case float64:
			return l + r, nil
		}
	}
	return nil, fmt.Errorf("cannot add %T and %T", left, right)
}

func (vm *VM) sub(left, right interface{}) (interface{}, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l - r, nil
		case float64:
			return float64(l) - r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l - float64(r), nil
		case float64:
			return l - r, nil
		}
	}
	return nil, fmt.Errorf("cannot subtract %T and %T", left, right)
}

func (vm *VM) mul(left, right interface{}) (interface{}, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l * r, nil
		case float64:
			return float64(l) * r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l * float64(r), nil
		case float64:
			return l * r, nil
		}
	}
	return nil, fmt.Errorf("cannot multiply %T and %T", left, right)
}

func (vm *VM) div(left, right interface{}) (interface{}, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			if r == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return l / r, nil
		case float64:
			if r == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return float64(l) / r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			if r == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return l / float64(r), nil
		case float64:
			if r == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return l / r, nil
		}
	}
	return nil, fmt.Errorf("cannot divide %T and %T", left, right)
}

func (vm *VM) mod(left, right interface{}) (interface{}, error) {
	l, lok := left.(int64)
	r, rok := right.(int64)
	if !lok || !rok {
		return nil, fmt.Errorf("modulo requires integers")
	}
	if r == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}
	return l % r, nil
}

// Comparison operations

func (vm *VM) eq(left, right interface{}) bool {
	return left == right
}

func (vm *VM) lt(left, right interface{}) (bool, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l < r, nil
		case float64:
			return float64(l) < r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l < float64(r), nil
		case float64:
			return l < r, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

func (vm *VM) le(left, right interface{}) (bool, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l <= r, nil
		case float64:
			return float64(l) <= r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l <= float64(r), nil
		case float64:
			return l <= r, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

func (vm *VM) gt(left, right interface{}) (bool, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l > r, nil
		case float64:
			return float64(l) > r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l > float64(r), nil
		case float64:
			return l > r, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

func (vm *VM) ge(left, right interface{}) (bool, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l >= r, nil
		case float64:
			return float64(l) >= r, nil
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l >= float64(r), nil
		case float64:
			return l >= r, nil
		}
	}
	return false, fmt.Errorf("cannot compare %T and %T", left, right)
}

// Helper function to convert to bool
func (vm *VM) toBool(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case nil:
		return false
	default:
		return true
	}
}
