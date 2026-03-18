package runtime

import (
	"testing"
)

func TestBuiltinsCreation(t *testing.T) {
	builtins := NewBuiltins()
	if builtins == nil {
		t.Fatal("builtins is nil")
	}
}

func TestBuiltinsPrint(t *testing.T) {
	builtins := NewBuiltins()
	_, err := builtins.Call("print", 42)
	if err != nil {
		t.Errorf("print failed: %v", err)
	}
}

func TestBuiltinsPrintln(t *testing.T) {
	builtins := NewBuiltins()
	_, err := builtins.Call("println", "hello")
	if err != nil {
		t.Errorf("println failed: %v", err)
	}
}

func TestBuiltinsLength(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("length", "hello")
	if err != nil {
		t.Errorf("length failed: %v", err)
	}

	if val, ok := result.(int64); ok && val == 5 {
		// Correct
	} else {
		t.Errorf("expected 5, got %v", result)
	}
}

func TestBuiltinsInt(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("int", 42.7)
	if err != nil {
		t.Errorf("int conversion failed: %v", err)
	}

	if val, ok := result.(int64); ok && val == 42 {
		// Correct
	} else {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestBuiltinsFloat(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("float", int64(42))
	if err != nil {
		t.Errorf("float conversion failed: %v", err)
	}

	if val, ok := result.(float64); ok && val == 42.0 {
		// Correct
	} else {
		t.Errorf("expected 42.0, got %v", result)
	}
}

func TestBuiltinsString(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("string", 42)
	if err != nil {
		t.Errorf("string conversion failed: %v", err)
	}

	if val, ok := result.(string); ok && val == "42" {
		// Correct
	} else {
		t.Errorf("expected '42', got %v", result)
	}
}

func TestBuiltinsAbs(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("abs", int64(-42))
	if err != nil {
		t.Errorf("abs failed: %v", err)
	}

	if val, ok := result.(int64); ok && val == 42 {
		// Correct
	} else {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestBuiltinsSqrt(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("sqrt", 16.0)
	if err != nil {
		t.Errorf("sqrt failed: %v", err)
	}

	if val, ok := result.(float64); ok && val == 4.0 {
		// Correct
	} else {
		t.Errorf("expected 4.0, got %v", result)
	}
}

func TestBuiltinsMin(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("min", int64(5), int64(3), int64(7))
	if err != nil {
		t.Errorf("min failed: %v", err)
	}

	if val, ok := result.(int64); ok && val == 3 {
		// Correct
	} else {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestBuiltinsMax(t *testing.T) {
	builtins := NewBuiltins()
	result, err := builtins.Call("max", int64(5), int64(3), int64(7))
	if err != nil {
		t.Errorf("max failed: %v", err)
	}

	if val, ok := result.(int64); ok && val == 7 {
		// Correct
	} else {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestRegisterCustomFunction(t *testing.T) {
	builtins := NewBuiltins()

	custom := func(args ...interface{}) (interface{}, error) {
		return int64(100), nil
	}

	builtins.Register("custom", custom)
	result, err := builtins.Call("custom")
	if err != nil {
		t.Errorf("custom function failed: %v", err)
	}

	if val, ok := result.(int64); ok && val == 100 {
		// Correct
	} else {
		t.Errorf("expected 100, got %v", result)
	}
}

func TestUnknownFunction(t *testing.T) {
	builtins := NewBuiltins()
	_, err := builtins.Call("nonexistent", 42)
	if err == nil {
		t.Error("should return error for unknown function")
	}
}
