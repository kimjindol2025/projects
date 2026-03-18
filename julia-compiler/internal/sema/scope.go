package sema

import (
	"fmt"
	"juliacc/internal/types"
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name      string
	Type      types.Type
	Kind      SymbolKind
	Mutable   bool
	Module    string
	Depth     int
	Declared  bool
}

// SymbolKind indicates what kind of symbol this is
type SymbolKind int

const (
	KindVariable SymbolKind = iota
	KindConstant
	KindFunction
	KindType
	KindModule
	KindMacro
)

func (k SymbolKind) String() string {
	switch k {
	case KindVariable:
		return "Variable"
	case KindConstant:
		return "Constant"
	case KindFunction:
		return "Function"
	case KindType:
		return "Type"
	case KindModule:
		return "Module"
	case KindMacro:
		return "Macro"
	default:
		return "Unknown"
	}
}

// Scope represents a lexical scope
type Scope struct {
	parent    *Scope
	symbols   map[string]*Symbol
	depth     int
	namespace string // module name or scope identifier
}

// NewScope creates a new scope
func NewScope(parent *Scope, namespace string) *Scope {
	depth := 0
	if parent != nil {
		depth = parent.depth + 1
	}
	return &Scope{
		parent:    parent,
		symbols:   make(map[string]*Symbol),
		depth:     depth,
		namespace: namespace,
	}
}

// Define adds a symbol to this scope
func (s *Scope) Define(name string, sym *Symbol) error {
	if _, exists := s.symbols[name]; exists {
		return fmt.Errorf("symbol %q already defined in this scope", name)
	}
	sym.Depth = s.depth
	sym.Module = s.namespace
	s.symbols[name] = sym
	return nil
}

// Declare declares a symbol without initializing
func (s *Scope) Declare(name string, kind SymbolKind, typT types.Type) error {
	sym := &Symbol{
		Name:     name,
		Type:     typT,
		Kind:     kind,
		Depth:    s.depth,
		Module:   s.namespace,
		Declared: true,
	}
	return s.Define(name, sym)
}

// Resolve looks up a symbol in this scope and parent scopes
func (s *Scope) Resolve(name string) *Symbol {
	if sym, exists := s.symbols[name]; exists {
		return sym
	}
	if s.parent != nil {
		return s.parent.Resolve(name)
	}
	return nil
}

// ResolveLocal looks up a symbol only in this scope
func (s *Scope) ResolveLocal(name string) *Symbol {
	return s.symbols[name]
}

// Set updates the type of an existing symbol
func (s *Scope) Set(name string, typ types.Type) error {
	sym := s.Resolve(name)
	if sym == nil {
		return fmt.Errorf("undefined symbol %q", name)
	}
	sym.Type = typ
	sym.Declared = true
	return nil
}

// All returns all symbols in this scope (not parent scopes)
func (s *Scope) All() map[string]*Symbol {
	result := make(map[string]*Symbol)
	for k, v := range s.symbols {
		result[k] = v
	}
	return result
}

// ScopeStack manages a stack of scopes (represents the scope chain)
type ScopeStack struct {
	current *Scope
	root    *Scope
	types   *types.Registry
}

// NewScopeStack creates a new scope stack
func NewScopeStack(typeRegistry *types.Registry) *ScopeStack {
	root := NewScope(nil, "global")
	return &ScopeStack{
		current: root,
		root:    root,
		types:   typeRegistry,
	}
}

// Push creates a new child scope
func (ss *ScopeStack) Push(namespace string) {
	ss.current = NewScope(ss.current, namespace)
}

// Pop returns to the parent scope
func (ss *ScopeStack) Pop() error {
	if ss.current.parent == nil {
		return fmt.Errorf("cannot pop root scope")
	}
	ss.current = ss.current.parent
	return nil
}

// Define adds a symbol to the current scope
func (ss *ScopeStack) Define(name string, sym *Symbol) error {
	return ss.current.Define(name, sym)
}

// Declare declares a symbol in the current scope
func (ss *ScopeStack) Declare(name string, kind SymbolKind, typT types.Type) error {
	return ss.current.Declare(name, kind, typT)
}

// Resolve looks up a symbol in the current scope chain
func (ss *ScopeStack) Resolve(name string) *Symbol {
	return ss.current.Resolve(name)
}

// Current returns the current scope
func (ss *ScopeStack) Current() *Scope {
	return ss.current
}

// Root returns the root scope
func (ss *ScopeStack) Root() *Scope {
	return ss.root
}

// Depth returns the current scope depth
func (ss *ScopeStack) Depth() int {
	return ss.current.depth
}

// TypeRegistry returns the type registry
func (ss *ScopeStack) TypeRegistry() *types.Registry {
	return ss.types
}
