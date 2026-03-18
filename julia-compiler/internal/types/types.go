package types

import (
	"fmt"
	"sort"
	"strings"
)

// TypeKind represents the kind of a Julia type
type TypeKind int

const (
	// Primitive types
	KindAny TypeKind = iota
	KindNothing
	KindBool
	KindType

	// Integer types
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUInt8
	KindUInt16
	KindUInt32
	KindUInt64

	// Float types
	KindFloat32
	KindFloat64

	// Complex types
	KindComplex64
	KindComplex128

	// Numeric union types
	KindInteger
	KindNumber
	KindFloating
	KindSigned
	KindUnsigned

	// String and chars
	KindString
	KindChar
	KindSymbol

	// Container types
	KindArray
	KindVector
	KindMatrix
	KindTuple
	KindDict
	KindSet

	// Function types
	KindFunction

	// User-defined types
	KindStruct
	KindMutableStruct
	KindUnion

	// Abstract types
	KindAbstract
)

// TypeString returns string representation of TypeKind
func (k TypeKind) String() string {
	switch k {
	case KindAny:
		return "Any"
	case KindNothing:
		return "Nothing"
	case KindBool:
		return "Bool"
	case KindType:
		return "Type"
	case KindInt8:
		return "Int8"
	case KindInt16:
		return "Int16"
	case KindInt32:
		return "Int32"
	case KindInt64:
		return "Int64"
	case KindUInt8:
		return "UInt8"
	case KindUInt16:
		return "UInt16"
	case KindUInt32:
		return "UInt32"
	case KindUInt64:
		return "UInt64"
	case KindFloat32:
		return "Float32"
	case KindFloat64:
		return "Float64"
	case KindComplex64:
		return "Complex64"
	case KindComplex128:
		return "Complex128"
	case KindInteger:
		return "Integer"
	case KindNumber:
		return "Number"
	case KindFloating:
		return "Floating"
	case KindSigned:
		return "Signed"
	case KindUnsigned:
		return "Unsigned"
	case KindString:
		return "String"
	case KindChar:
		return "Char"
	case KindSymbol:
		return "Symbol"
	case KindArray:
		return "Array"
	case KindVector:
		return "Vector"
	case KindMatrix:
		return "Matrix"
	case KindTuple:
		return "Tuple"
	case KindDict:
		return "Dict"
	case KindSet:
		return "Set"
	case KindFunction:
		return "Function"
	case KindStruct:
		return "Struct"
	case KindMutableStruct:
		return "MutableStruct"
	case KindUnion:
		return "Union"
	case KindAbstract:
		return "Abstract"
	default:
		return "Unknown"
	}
}

// Type is the base interface for all Julia types
type Type interface {
	// Kind returns the kind of this type
	Kind() TypeKind

	// String returns the string representation of the type
	String() string

	// IsSubtypeOf checks if this type is a subtype of another
	IsSubtypeOf(other Type) bool

	// Underlying returns the underlying type (for type aliases, etc.)
	Underlying() Type

	// Size returns the memory size in bytes (0 for abstract types)
	Size() int
}

// BasicType represents a basic/primitive Julia type
type BasicType struct {
	kind TypeKind
	name string
	size int
}

func NewBasicType(kind TypeKind, name string, size int) *BasicType {
	return &BasicType{kind: kind, name: name, size: size}
}

func (b *BasicType) Kind() TypeKind {
	return b.kind
}

func (b *BasicType) String() string {
	return b.name
}

func (b *BasicType) Underlying() Type {
	return b
}

func (b *BasicType) Size() int {
	return b.size
}

func (b *BasicType) IsSubtypeOf(other Type) bool {
	if other == nil {
		return false
	}

	otherKind := other.Kind()

	// Direct match
	if b.kind == otherKind {
		return true
	}

	// Subtype relationships
	switch b.kind {
	// Integer subtypes
	case KindInt8, KindInt16, KindInt32, KindInt64:
		return otherKind == KindInteger || otherKind == KindSigned || otherKind == KindNumber || otherKind == KindAny
	case KindUInt8, KindUInt16, KindUInt32, KindUInt64:
		return otherKind == KindInteger || otherKind == KindUnsigned || otherKind == KindNumber || otherKind == KindAny

	// Float subtypes
	case KindFloat32, KindFloat64:
		return otherKind == KindFloating || otherKind == KindNumber || otherKind == KindAny

	// Complex subtypes
	case KindComplex64, KindComplex128:
		return otherKind == KindNumber || otherKind == KindAny

	// Number categories
	case KindInteger:
		return otherKind == KindNumber || otherKind == KindAny
	case KindFloating:
		return otherKind == KindNumber || otherKind == KindAny
	case KindSigned:
		return otherKind == KindInteger || otherKind == KindNumber || otherKind == KindAny
	case KindUnsigned:
		return otherKind == KindInteger || otherKind == KindNumber || otherKind == KindAny
	case KindNumber:
		return otherKind == KindAny

	// Everything is subtype of Any
	default:
		return otherKind == KindAny
	}
}

// ParametricType represents a parametric type like Vector{Int64} or Dict{String,Int64}
type ParametricType struct {
	base       *BasicType
	parameters []Type
}

func NewParametricType(base *BasicType, parameters ...Type) *ParametricType {
	return &ParametricType{
		base:       base,
		parameters: parameters,
	}
}

func (p *ParametricType) Kind() TypeKind {
	return p.base.kind
}

func (p *ParametricType) String() string {
	if len(p.parameters) == 0 {
		return p.base.String()
	}
	params := make([]string, len(p.parameters))
	for i, param := range p.parameters {
		params[i] = param.String()
	}
	return fmt.Sprintf("%s{%s}", p.base.String(), strings.Join(params, ","))
}

func (p *ParametricType) Underlying() Type {
	return p.base
}

func (p *ParametricType) Size() int {
	// Parametric types are runtime-dependent
	return 0
}

func (p *ParametricType) IsSubtypeOf(other Type) bool {
	if other == nil {
		return false
	}

	// Try to match with the base type
	if p.base.IsSubtypeOf(other) {
		return true
	}

	// If other is also parametric, both must match
	if otherParam, ok := other.(*ParametricType); ok {
		if !p.base.IsSubtypeOf(otherParam.base) {
			return false
		}
		if len(p.parameters) != len(otherParam.parameters) {
			return false
		}
		for i := range p.parameters {
			if !p.parameters[i].IsSubtypeOf(otherParam.parameters[i]) {
				return false
			}
		}
		return true
	}

	return false
}

// UnionType represents a union of multiple types (Union{Int64,Float64})
type UnionType struct {
	members []Type
}

func NewUnionType(members ...Type) *UnionType {
	return &UnionType{members: members}
}

func (u *UnionType) Kind() TypeKind {
	return KindUnion
}

func (u *UnionType) String() string {
	parts := make([]string, len(u.members))
	for i, member := range u.members {
		parts[i] = member.String()
	}
	return fmt.Sprintf("Union{%s}", strings.Join(parts, ","))
}

func (u *UnionType) Underlying() Type {
	return u
}

func (u *UnionType) Size() int {
	return 0 // Runtime dependent
}

func (u *UnionType) IsSubtypeOf(other Type) bool {
	if other == nil {
		return false
	}

	otherKind := other.Kind()

	// If other is Any, union is subtype
	if otherKind == KindAny {
		return true
	}

	// If other is also Union, all members of u must be subtype of some member of other
	if otherUnion, ok := other.(*UnionType); ok {
		for _, uMember := range u.members {
			found := false
			for _, otherMember := range otherUnion.members {
				if uMember.IsSubtypeOf(otherMember) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	// If other is not union, all members of u must be subtype of other
	for _, member := range u.members {
		if !member.IsSubtypeOf(other) {
			return false
		}
	}
	return true
}

// TupleType represents a tuple type with fixed element types
type TupleType struct {
	elementTypes []Type
}

func NewTupleType(elementTypes ...Type) *TupleType {
	return &TupleType{elementTypes: elementTypes}
}

func (t *TupleType) Kind() TypeKind {
	return KindTuple
}

func (t *TupleType) String() string {
	parts := make([]string, len(t.elementTypes))
	for i, elem := range t.elementTypes {
		parts[i] = elem.String()
	}
	return fmt.Sprintf("Tuple{%s}", strings.Join(parts, ","))
}

func (t *TupleType) Underlying() Type {
	return t
}

func (t *TupleType) Size() int {
	sum := 0
	for _, elem := range t.elementTypes {
		sum += elem.Size()
	}
	return sum
}

func (t *TupleType) IsSubtypeOf(other Type) bool {
	if other == nil {
		return false
	}

	if other.Kind() == KindAny {
		return true
	}

	if otherTuple, ok := other.(*TupleType); ok {
		if len(t.elementTypes) != len(otherTuple.elementTypes) {
			return false
		}
		for i := range t.elementTypes {
			if !t.elementTypes[i].IsSubtypeOf(otherTuple.elementTypes[i]) {
				return false
			}
		}
		return true
	}

	return false
}

// FunctionType represents a function signature
type FunctionType struct {
	paramTypes []Type
	returnType Type
	variadic   bool
}

func NewFunctionType(paramTypes []Type, returnType Type, variadic bool) *FunctionType {
	return &FunctionType{
		paramTypes: paramTypes,
		returnType: returnType,
		variadic:   variadic,
	}
}

func (f *FunctionType) Kind() TypeKind {
	return KindFunction
}

func (f *FunctionType) String() string {
	parts := make([]string, len(f.paramTypes))
	for i, param := range f.paramTypes {
		parts[i] = param.String()
	}
	variadicStr := ""
	if f.variadic {
		variadicStr = "..."
	}
	return fmt.Sprintf("(%s%s) -> %s", strings.Join(parts, ","), variadicStr, f.returnType.String())
}

func (f *FunctionType) Underlying() Type {
	return f
}

func (f *FunctionType) Size() int {
	return 0 // Function types are not sized values
}

func (f *FunctionType) IsSubtypeOf(other Type) bool {
	if other == nil {
		return false
	}

	if other.Kind() == KindAny || other.Kind() == KindFunction {
		return true
	}

	if otherFunc, ok := other.(*FunctionType); ok {
		if len(f.paramTypes) != len(otherFunc.paramTypes) {
			return false
		}
		// Function subtyping: contravariant in parameters, covariant in return
		for i := range f.paramTypes {
			if !otherFunc.paramTypes[i].IsSubtypeOf(f.paramTypes[i]) {
				return false
			}
		}
		return f.returnType.IsSubtypeOf(otherFunc.returnType)
	}

	return false
}

// Registry stores all known types
type Registry struct {
	types      map[string]Type
	primitives map[TypeKind]Type
}

func NewRegistry() *Registry {
	reg := &Registry{
		types:      make(map[string]Type),
		primitives: make(map[TypeKind]Type),
	}

	// Register all primitive types
	primitiveTypes := []*BasicType{
		NewBasicType(KindAny, "Any", 0),
		NewBasicType(KindNothing, "Nothing", 0),
		NewBasicType(KindBool, "Bool", 1),
		NewBasicType(KindType, "Type", 8),
		NewBasicType(KindInt8, "Int8", 1),
		NewBasicType(KindInt16, "Int16", 2),
		NewBasicType(KindInt32, "Int32", 4),
		NewBasicType(KindInt64, "Int64", 8),
		NewBasicType(KindUInt8, "UInt8", 1),
		NewBasicType(KindUInt16, "UInt16", 2),
		NewBasicType(KindUInt32, "UInt32", 4),
		NewBasicType(KindUInt64, "UInt64", 8),
		NewBasicType(KindFloat32, "Float32", 4),
		NewBasicType(KindFloat64, "Float64", 8),
		NewBasicType(KindComplex64, "Complex64", 8),
		NewBasicType(KindComplex128, "Complex128", 16),
		NewBasicType(KindInteger, "Integer", 0),
		NewBasicType(KindNumber, "Number", 0),
		NewBasicType(KindFloating, "Floating", 0),
		NewBasicType(KindSigned, "Signed", 0),
		NewBasicType(KindUnsigned, "Unsigned", 0),
		NewBasicType(KindString, "String", 0),
		NewBasicType(KindChar, "Char", 4),
		NewBasicType(KindSymbol, "Symbol", 0),
	}

	for _, t := range primitiveTypes {
		reg.types[t.String()] = t
		reg.primitives[t.Kind()] = t
	}

	return reg
}

// Get retrieves a type by name
func (r *Registry) Get(name string) Type {
	return r.types[name]
}

// GetByKind retrieves a primitive type by kind
func (r *Registry) GetByKind(kind TypeKind) Type {
	return r.primitives[kind]
}

// Register stores a new type in the registry
func (r *Registry) Register(name string, t Type) {
	r.types[name] = t
}

// AllTypes returns all registered types (sorted by name)
func (r *Registry) AllTypes() []string {
	var names []string
	for name := range r.types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
