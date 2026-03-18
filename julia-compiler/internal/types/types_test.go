package types

import (
	"testing"
)

// Test basic type operations
func TestBasicTypes(t *testing.T) {
	registry := NewRegistry()

	// Test that all primitive types are registered
	typeNames := []string{"Int64", "Float64", "String", "Bool"}
	for _, name := range typeNames {
		if registry.Get(name) == nil {
			t.Errorf("Type %s not registered", name)
		}
	}
}

func TestBasicTypeSubtype(t *testing.T) {
	tests := []struct {
		subType    TypeKind
		superType  TypeKind
		shouldPass bool
	}{
		{KindInt64, KindInteger, true},
		{KindInt64, KindSigned, true},
		{KindInt64, KindNumber, true},
		{KindInt64, KindAny, true},
		{KindInt64, KindFloat64, false},
		{KindFloat64, KindFloating, true},
		{KindFloat64, KindNumber, true},
		{KindUInt32, KindUnsigned, true},
		{KindUInt32, KindInteger, true},
		{KindBool, KindAny, true},
		{KindBool, KindInteger, false},
	}

	for _, test := range tests {
		subType := NewBasicType(test.subType, test.subType.String(), 0)
		superType := NewBasicType(test.superType, test.superType.String(), 0)

		result := subType.IsSubtypeOf(superType)
		if result != test.shouldPass {
			t.Errorf("IsSubtypeOf(%s, %s) = %v, want %v",
				test.subType.String(), test.superType.String(), result, test.shouldPass)
		}
	}
}

func TestParametricType(t *testing.T) {
	vector := NewBasicType(KindVector, "Vector", 0)
	int64Type := NewBasicType(KindInt64, "Int64", 8)

	vectorInt := NewParametricType(vector, int64Type)

	if vectorInt.String() != "Vector{Int64}" {
		t.Errorf("ParametricType.String() = %s, want Vector{Int64}", vectorInt.String())
	}

	arrayType := NewBasicType(KindArray, "Array", 0)
	if vectorInt.IsSubtypeOf(arrayType) {
		t.Error("Vector{Int64} should not be subtype of Array (without matching parameters)")
	}
}

func TestUnionType(t *testing.T) {
	int64Type := NewBasicType(KindInt64, "Int64", 8)
	float64Type := NewBasicType(KindFloat64, "Float64", 8)

	unionType := NewUnionType(int64Type, float64Type)

	if unionType.String() != "Union{Int64,Float64}" {
		t.Errorf("UnionType.String() = %s, want Union{Int64,Float64}", unionType.String())
	}

	anyType := NewBasicType(KindAny, "Any", 0)
	if !unionType.IsSubtypeOf(anyType) {
		t.Error("Union{Int64,Float64} should be subtype of Any")
	}

	numberType := NewBasicType(KindNumber, "Number", 0)
	if !unionType.IsSubtypeOf(numberType) {
		t.Error("Union{Int64,Float64} should be subtype of Number")
	}
}

func TestTupleType(t *testing.T) {
	int64Type := NewBasicType(KindInt64, "Int64", 8)
	stringType := NewBasicType(KindString, "String", 0)

	tupleType := NewTupleType(int64Type, stringType)

	if tupleType.String() != "Tuple{Int64,String}" {
		t.Errorf("TupleType.String() = %s, want Tuple{Int64,String}", tupleType.String())
	}

	// Test tuple size calculation
	size := tupleType.Size()
	if size != 8 {
		t.Errorf("TupleType.Size() = %d, want 8 (Int64=8, String=0)", size)
	}
}

func TestFunctionType(t *testing.T) {
	int64Type := NewBasicType(KindInt64, "Int64", 8)
	float64Type := NewBasicType(KindFloat64, "Float64", 8)

	funcType := NewFunctionType([]Type{int64Type, int64Type}, float64Type, false)

	if funcType.String() != "(Int64,Int64) -> Float64" {
		t.Errorf("FunctionType.String() = %s, want (Int64,Int64) -> Float64", funcType.String())
	}
}

func TestTypeHierarchy(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)

	// Test IsSubtype
	if !hierarchy.IsSubtype(KindInt64, KindInteger) {
		t.Error("Int64 should be subtype of Integer")
	}

	if !hierarchy.IsSubtype(KindInteger, KindNumber) {
		t.Error("Integer should be subtype of Number")
	}

	if !hierarchy.IsSubtype(KindNumber, KindAny) {
		t.Error("Number should be subtype of Any")
	}

	if hierarchy.IsSubtype(KindInteger, KindFloating) {
		t.Error("Integer should not be subtype of Floating")
	}
}

func TestCommonSupertype(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)

	tests := []struct {
		kind1    TypeKind
		kind2    TypeKind
		expected TypeKind
	}{
		{KindInt64, KindInt32, KindSigned},
		{KindInt64, KindUInt64, KindInteger},
		{KindFloat32, KindFloat64, KindFloating},
		{KindInt64, KindFloat64, KindNumber},
	}

	for _, test := range tests {
		result := hierarchy.CommonSupertype(test.kind1, test.kind2)
		if result != test.expected {
			t.Errorf("CommonSupertype(%s, %s) = %s, want %s",
				test.kind1.String(), test.kind2.String(), result.String(), test.expected.String())
		}
	}
}

func TestNumericPromotion(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)

	tests := []struct {
		kind1    TypeKind
		kind2    TypeKind
		expected TypeKind
	}{
		{KindInt64, KindInt32, KindInt64},
		{KindInt32, KindInt64, KindInt64},
		{KindInt64, KindFloat64, KindFloat64},
		{KindFloat32, KindInt64, KindFloat64},
		{KindFloat32, KindFloat32, KindFloat32},
		{KindNothing, KindInt64, KindInt64},
	}

	for _, test := range tests {
		result := hierarchy.PromoteTypes(test.kind1, test.kind2)
		if result != test.expected {
			t.Errorf("PromoteTypes(%s, %s) = %s, want %s",
				test.kind1.String(), test.kind2.String(), result.String(), test.expected.String())
		}
	}
}

func TestMethodTable(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)

	mt := NewMethodTable("add", hierarchy)

	int64Type := registry.Get("Int64")
	float64Type := registry.Get("Float64")

	// Add some methods
	m1 := mt.AddMethod([]Type{int64Type, int64Type}, int64Type, false)
	m2 := mt.AddMethod([]Type{float64Type, float64Type}, float64Type, false)

	if m1.ID == m2.ID {
		t.Error("Methods should have different IDs")
	}

	// Find exact match
	method, score := mt.FindMethod([]Type{int64Type, int64Type})
	if method != m1 || score != 0 {
		t.Errorf("FindMethod for (Int64,Int64) returned wrong method or score")
	}
}

func TestMethodDispatch(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)
	dispatch := NewDispatch(hierarchy)

	// Register add function
	addMethods := dispatch.RegisterFunction("add")

	int64Type := registry.Get("Int64")
	float64Type := registry.Get("Float64")
	numberType := registry.Get("Number")

	// Add methods
	addMethods.AddMethod([]Type{int64Type, int64Type}, int64Type, false)
	addMethods.AddMethod([]Type{float64Type, float64Type}, float64Type, false)
	addMethods.AddMethod([]Type{numberType, numberType}, numberType, false)

	// Look up method
	method, err := dispatch.LookupMethod("add", []Type{int64Type, int64Type})
	if err != nil {
		t.Errorf("LookupMethod failed: %v", err)
	}
	if method == nil {
		t.Error("LookupMethod returned nil")
	}

	// Look up non-existent method
	stringType := registry.Get("String")
	_, err = dispatch.LookupMethod("add", []Type{stringType, stringType})
	if err == nil {
		t.Error("LookupMethod should fail for String,String")
	}
}

func TestConversionCost(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)
	dispatch := NewDispatch(hierarchy)

	int64Type := registry.Get("Int64")
	numberType := registry.Get("Number")
	stringType := registry.Get("String")

	// Int64 to Int64: no conversion
	cost1 := dispatch.ConversionCost(int64Type, int64Type)
	if cost1 != 0 {
		t.Errorf("ConversionCost(Int64, Int64) = %d, want 0", cost1)
	}

	// Int64 to Number: upcast
	cost2 := dispatch.ConversionCost(int64Type, numberType)
	if cost2 != 1 {
		t.Errorf("ConversionCost(Int64, Number) = %d, want 1", cost2)
	}

	// Int64 to String: impossible
	cost3 := dispatch.ConversionCost(int64Type, stringType)
	if cost3 < 100 {
		t.Errorf("ConversionCost(Int64, String) = %d, want >= 100", cost3)
	}
}

func TestGetSupertypes(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)

	supers := hierarchy.GetSupertypes(KindInt64)
	expectedTypes := map[TypeKind]bool{
		KindSigned:   true,
		KindInteger:  true,
		KindNumber:   true,
		KindAny:      true,
	}

	if len(supers) != len(expectedTypes) {
		t.Errorf("GetSupertypes(Int64) returned %d types, want %d", len(supers), len(expectedTypes))
	}

	for _, superType := range supers {
		if !expectedTypes[superType] {
			t.Errorf("Unexpected supertype: %s", superType.String())
		}
	}
}

func TestIsAbstractType(t *testing.T) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)

	tests := []struct {
		kind       TypeKind
		isAbstract bool
	}{
		{KindNumber, true},
		{KindInteger, true},
		{KindSigned, true},
		{KindAny, true},
		{KindInt64, false},
		{KindFloat64, false},
		{KindString, false},
	}

	for _, test := range tests {
		result := hierarchy.IsAbstractType(test.kind)
		if result != test.isAbstract {
			t.Errorf("IsAbstractType(%s) = %v, want %v",
				test.kind.String(), result, test.isAbstract)
		}
	}
}

func TestTypeRegistry(t *testing.T) {
	registry := NewRegistry()

	// Get by kind
	int64Type := registry.GetByKind(KindInt64)
	if int64Type == nil || int64Type.String() != "Int64" {
		t.Error("GetByKind(KindInt64) failed")
	}

	// Register custom type
	customType := NewBasicType(KindStruct, "MyStruct", 16)
	registry.Register("MyStruct", customType)

	retrieved := registry.Get("MyStruct")
	if retrieved != customType {
		t.Error("Failed to register and retrieve custom type")
	}

	// List all types
	allTypes := registry.AllTypes()
	if len(allTypes) == 0 {
		t.Error("AllTypes returned empty list")
	}
}

func BenchmarkTypeHierarchyIsSubtype(b *testing.B) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hierarchy.IsSubtype(KindInt64, KindNumber)
	}
}

func BenchmarkMethodDispatch(b *testing.B) {
	registry := NewRegistry()
	hierarchy := NewHierarchy(registry)
	dispatch := NewDispatch(hierarchy)

	addMethods := dispatch.RegisterFunction("add")
	int64Type := registry.Get("Int64")
	float64Type := registry.Get("Float64")

	addMethods.AddMethod([]Type{int64Type, int64Type}, int64Type, false)
	addMethods.AddMethod([]Type{float64Type, float64Type}, float64Type, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dispatch.LookupMethod("add", []Type{int64Type, int64Type})
	}
}
