// Package typechecker implements type checking and inference for FV 2.0
package typechecker

import "fmt"

// Type represents a type in FV 2.0
type Type interface {
	TypeString() string
	Equal(other Type) bool
}

// PrimitiveType represents built-in types
type PrimitiveType struct {
	Name string // i64, f64, string, bool, none
}

func (p *PrimitiveType) TypeString() string { return p.Name }
func (p *PrimitiveType) Equal(other Type) bool {
	if o, ok := other.(*PrimitiveType); ok {
		return p.Name == o.Name
	}
	return false
}

// ArrayType represents an array type
type ArrayType struct {
	ElementType Type
}

func (a *ArrayType) TypeString() string {
	return "[" + a.ElementType.TypeString() + "]"
}
func (a *ArrayType) Equal(other Type) bool {
	if o, ok := other.(*ArrayType); ok {
		return a.ElementType.Equal(o.ElementType)
	}
	return false
}

// FunctionType represents a function type
type FunctionType struct {
	ParamTypes []Type
	ReturnType Type
}

func (f *FunctionType) TypeString() string {
	params := ""
	for i, p := range f.ParamTypes {
		if i > 0 {
			params += ", "
		}
		params += p.TypeString()
	}
	return fmt.Sprintf("fn(%s) %s", params, f.ReturnType.TypeString())
}
func (f *FunctionType) Equal(other Type) bool {
	if o, ok := other.(*FunctionType); ok {
		if len(f.ParamTypes) != len(o.ParamTypes) {
			return false
		}
		for i, p := range f.ParamTypes {
			if !p.Equal(o.ParamTypes[i]) {
				return false
			}
		}
		return f.ReturnType.Equal(o.ReturnType)
	}
	return false
}

// BuiltinFunctionType represents a built-in function (variadic or special)
type BuiltinFunctionType struct {
	Name       string
	IsVariadic bool
}

func (b *BuiltinFunctionType) TypeString() string {
	return fmt.Sprintf("builtin(%s)", b.Name)
}
func (b *BuiltinFunctionType) Equal(other Type) bool {
	if o, ok := other.(*BuiltinFunctionType); ok {
		return b.Name == o.Name
	}
	return false
}

// OptionType represents Option[T]
type OptionType struct {
	InnerType Type
}

func (o *OptionType) TypeString() string {
	return "Option[" + o.InnerType.TypeString() + "]"
}
func (o *OptionType) Equal(other Type) bool {
	if oo, ok := other.(*OptionType); ok {
		return o.InnerType.Equal(oo.InnerType)
	}
	return false
}

// ResultType represents Result[T, E]
type ResultType struct {
	OkType    Type
	ErrorType Type
}

func (r *ResultType) TypeString() string {
	return fmt.Sprintf("Result[%s, %s]", r.OkType.TypeString(), r.ErrorType.TypeString())
}
func (r *ResultType) Equal(other Type) bool {
	if ro, ok := other.(*ResultType); ok {
		return r.OkType.Equal(ro.OkType) && r.ErrorType.Equal(ro.ErrorType)
	}
	return false
}

// StructType represents a struct type
type StructType struct {
	Name   string
	Fields map[string]Type
}

func (s *StructType) TypeString() string {
	return "struct " + s.Name
}
func (s *StructType) Equal(other Type) bool {
	if o, ok := other.(*StructType); ok {
		return s.Name == o.Name
	}
	return false
}

// UnionType represents a union type (e.g., int | string)
type UnionType struct {
	Types []Type
}

func (u *UnionType) TypeString() string {
	result := ""
	for i, t := range u.Types {
		if i > 0 {
			result += " | "
		}
		result += t.TypeString()
	}
	return result
}
func (u *UnionType) Equal(other Type) bool {
	if o, ok := other.(*UnionType); ok {
		if len(u.Types) != len(o.Types) {
			return false
		}
		for i, t := range u.Types {
			if !t.Equal(o.Types[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// Symbol represents a variable or function binding
type Symbol struct {
	Name string
	Type Type
	Kind string // "var", "function", "const", "type"
}

// Scope represents a scope level
type Scope struct {
	Symbols map[string]*Symbol
	Parent  *Scope
}

// NewScope creates a new scope
func NewScope(parent *Scope) *Scope {
	return &Scope{
		Symbols: make(map[string]*Symbol),
		Parent:  parent,
	}
}

// Define adds a symbol to the scope
func (s *Scope) Define(name string, typ Type, kind string) {
	s.Symbols[name] = &Symbol{
		Name: name,
		Type: typ,
		Kind: kind,
	}
}

// Lookup finds a symbol in this scope or parent scopes
func (s *Scope) Lookup(name string) *Symbol {
	if sym, ok := s.Symbols[name]; ok {
		return sym
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

// Error represents a type checking error
type Error struct {
	Line    int
	Column  int
	Message string
}

func (e Error) String() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
}
