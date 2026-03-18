package typeinf

import (
	"juliacc/internal/ir"
	"testing"
)

func TestInferrerCreation(t *testing.T) {
	mod := ir.NewModule()
	inf := NewInferrer(mod)
	if inf == nil {
		t.Fatal("inferrer is nil")
	}
}

func TestInferrerInfer(t *testing.T) {
	mod := ir.NewModule()
	inf := NewInferrer(mod)

	err := inf.Infer()
	if err != nil {
		t.Errorf("inference error: %v", err)
	}
}

func TestTypeMap(t *testing.T) {
	mod := ir.NewModule()
	inf := NewInferrer(mod)

	typeMap := inf.GetTypeMap()
	if typeMap == nil {
		t.Fatal("type map is nil")
	}

	if len(typeMap) != 0 {
		t.Errorf("expected empty type map, got %d entries", len(typeMap))
	}
}

func TestPromoteTypes(t *testing.T) {
	mod := ir.NewModule()
	inf := NewInferrer(mod)

	result := inf.promoteTypes("i64", "i64")
	if result != "i64" {
		t.Errorf("expected i64, got %s", result)
	}

	result = inf.promoteTypes("i64", "f64")
	if result != "f64" {
		t.Errorf("expected f64 from i64+f64, got %s", result)
	}

	result = inf.promoteTypes("bool", "bool")
	if result != "bool" {
		t.Errorf("expected bool, got %s", result)
	}
}

func TestGetType(t *testing.T) {
	mod := ir.NewModule()
	inf := NewInferrer(mod)

	// Should return unknown for non-existent type
	result := inf.GetType(999)
	if result != "unknown" {
		t.Errorf("expected unknown for non-existent ID, got %s", result)
	}
}
