package runtime

import (
	"fmt"
	"math"
)

// Builtins defines built-in functions for the runtime
type Builtins struct {
	functions map[string]BuiltinFunc
}

// BuiltinFunc is the signature for built-in functions
type BuiltinFunc func(args ...interface{}) (interface{}, error)

// NewBuiltins creates a new built-in functions registry
func NewBuiltins() *Builtins {
	b := &Builtins{
		functions: make(map[string]BuiltinFunc),
	}

	// Register built-in functions
	b.functions["print"] = b.builtinPrint
	b.functions["println"] = b.builtinPrintln
	b.functions["length"] = b.builtinLength
	b.functions["size"] = b.builtinSize
	b.functions["int"] = b.builtinInt
	b.functions["float"] = b.builtinFloat
	b.functions["string"] = b.builtinString
	b.functions["abs"] = b.builtinAbs
	b.functions["sqrt"] = b.builtinSqrt
	b.functions["sin"] = b.builtinSin
	b.functions["cos"] = b.builtinCos
	b.functions["tan"] = b.builtinTan
	b.functions["floor"] = b.builtinFloor
	b.functions["ceil"] = b.builtinCeil
	b.functions["round"] = b.builtinRound
	b.functions["min"] = b.builtinMin
	b.functions["max"] = b.builtinMax

	return b
}

// Call invokes a built-in function
func (b *Builtins) Call(name string, args ...interface{}) (interface{}, error) {
	fn, ok := b.functions[name]
	if !ok {
		return nil, fmt.Errorf("unknown built-in function: %s", name)
	}
	return fn(args...)
}

// Register registers a custom built-in function
func (b *Builtins) Register(name string, fn BuiltinFunc) {
	b.functions[name] = fn
}

// Built-in function implementations

func (b *Builtins) builtinPrint(args ...interface{}) (interface{}, error) {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg)
	}
	return nil, nil
}

func (b *Builtins) builtinPrintln(args ...interface{}) (interface{}, error) {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg)
	}
	fmt.Println()
	return nil, nil
}

func (b *Builtins) builtinLength(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length expects 1 argument, got %d", len(args))
	}

	switch v := args[0].(type) {
	case string:
		return int64(len(v)), nil
	case []interface{}:
		return int64(len(v)), nil
	default:
		return nil, fmt.Errorf("cannot get length of %T", v)
	}
}

func (b *Builtins) builtinSize(args ...interface{}) (interface{}, error) {
	return b.builtinLength(args...)
}

func (b *Builtins) builtinInt(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("int expects 1 argument, got %d", len(args))
	}

	switch v := args[0].(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		var i int64
		_, err := fmt.Sscanf(v, "%d", &i)
		if err != nil {
			return nil, fmt.Errorf("cannot convert string to int: %v", err)
		}
		return i, nil
	case bool:
		if v {
			return int64(1), nil
		}
		return int64(0), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to int", v)
	}
}

func (b *Builtins) builtinFloat(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("float expects 1 argument, got %d", len(args))
	}

	switch v := args[0].(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		if err != nil {
			return nil, fmt.Errorf("cannot convert string to float: %v", err)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to float", v)
	}
}

func (b *Builtins) builtinString(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("string expects 1 argument, got %d", len(args))
	}

	return fmt.Sprintf("%v", args[0]), nil
}

func (b *Builtins) builtinAbs(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("abs expects 1 argument, got %d", len(args))
	}

	switch v := args[0].(type) {
	case int64:
		if v < 0 {
			return -v, nil
		}
		return v, nil
	case float64:
		return math.Abs(v), nil
	default:
		return nil, fmt.Errorf("cannot get abs of %T", v)
	}
}

func (b *Builtins) builtinSqrt(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sqrt expects 1 argument, got %d", len(args))
	}

	f, err := b.toFloat(args[0])
	if err != nil {
		return nil, err
	}

	return math.Sqrt(f), nil
}

func (b *Builtins) builtinSin(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sin expects 1 argument, got %d", len(args))
	}

	f, err := b.toFloat(args[0])
	if err != nil {
		return nil, err
	}

	return math.Sin(f), nil
}

func (b *Builtins) builtinCos(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cos expects 1 argument, got %d", len(args))
	}

	f, err := b.toFloat(args[0])
	if err != nil {
		return nil, err
	}

	return math.Cos(f), nil
}

func (b *Builtins) builtinTan(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tan expects 1 argument, got %d", len(args))
	}

	f, err := b.toFloat(args[0])
	if err != nil {
		return nil, err
	}

	return math.Tan(f), nil
}

func (b *Builtins) builtinFloor(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("floor expects 1 argument, got %d", len(args))
	}

	f, err := b.toFloat(args[0])
	if err != nil {
		return nil, err
	}

	return math.Floor(f), nil
}

func (b *Builtins) builtinCeil(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ceil expects 1 argument, got %d", len(args))
	}

	f, err := b.toFloat(args[0])
	if err != nil {
		return nil, err
	}

	return math.Ceil(f), nil
}

func (b *Builtins) builtinRound(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("round expects 1 argument, got %d", len(args))
	}

	f, err := b.toFloat(args[0])
	if err != nil {
		return nil, err
	}

	return math.Round(f), nil
}

func (b *Builtins) builtinMin(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("min expects at least 1 argument")
	}

	min := args[0]
	for i := 1; i < len(args); i++ {
		switch v := min.(type) {
		case int64:
			if other, ok := args[i].(int64); ok && other < v {
				min = other
			}
		case float64:
			if other, ok := args[i].(float64); ok && other < v {
				min = other
			}
		}
	}

	return min, nil
}

func (b *Builtins) builtinMax(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("max expects at least 1 argument")
	}

	max := args[0]
	for i := 1; i < len(args); i++ {
		switch v := max.(type) {
		case int64:
			if other, ok := args[i].(int64); ok && other > v {
				max = other
			}
		case float64:
			if other, ok := args[i].(float64); ok && other > v {
				max = other
			}
		}
	}

	return max, nil
}

// Helper functions

func (b *Builtins) toFloat(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float", v)
	}
}
