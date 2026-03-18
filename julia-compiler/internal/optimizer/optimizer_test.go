package optimizer

import (
	"juliacc/internal/ir"
	"testing"
)

func TestOptimizerCreation(t *testing.T) {
	mod := ir.NewModule()
	opt := NewOptimizer(mod)
	if opt == nil {
		t.Fatal("optimizer is nil")
	}
}

func TestOptimizePass(t *testing.T) {
	mod := ir.NewModule()
	opt := NewOptimizer(mod)

	err := opt.Optimize()
	if err != nil {
		t.Errorf("optimize error: %v", err)
	}
}

func TestConstantEvaluation(t *testing.T) {
	opt := NewOptimizer(ir.NewModule())

	// Test addition
	result := opt.evalBinOp("add", int64(5), int64(3))
	if val, ok := result.(int64); !ok || val != 8 {
		t.Errorf("expected 8, got %v", result)
	}

	// Test subtraction
	result = opt.evalBinOp("sub", int64(10), int64(3))
	if val, ok := result.(int64); !ok || val != 7 {
		t.Errorf("expected 7, got %v", result)
	}

	// Test multiplication
	result = opt.evalBinOp("mul", int64(6), int64(7))
	if val, ok := result.(int64); !ok || val != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestComparisonEval(t *testing.T) {
	opt := NewOptimizer(ir.NewModule())

	// Test equal
	result := opt.evalBinOp("eq", int64(5), int64(5))
	if val, ok := result.(bool); !ok || !val {
		t.Errorf("expected true, got %v", result)
	}

	// Test less than
	result = opt.evalBinOp("lt", int64(3), int64(5))
	if val, ok := result.(bool); !ok || !val {
		t.Errorf("expected true, got %v", result)
	}

	// Test greater than
	result = opt.evalBinOp("gt", int64(5), int64(3))
	if val, ok := result.(bool); !ok || !val {
		t.Errorf("expected true, got %v", result)
	}
}

func TestFloatOperations(t *testing.T) {
	opt := NewOptimizer(ir.NewModule())

	result := opt.evalBinOp("add", 2.5, 3.5)
	if val, ok := result.(float64); !ok || val != 6.0 {
		t.Errorf("expected 6.0, got %v", result)
	}
}

func TestDivisionByZero(t *testing.T) {
	opt := NewOptimizer(ir.NewModule())

	result := opt.evalBinOp("div", int64(10), int64(0))
	if result != nil {
		t.Errorf("division by zero should return nil, got %v", result)
	}
}
