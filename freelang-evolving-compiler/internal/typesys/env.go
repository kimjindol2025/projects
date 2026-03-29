package typesys

// Scope represents a single block's variable type environment
type Scope struct {
	vars   map[string]TypeInfo
	parent *Scope
}

func newScope(parent *Scope) *Scope {
	return &Scope{
		vars:   make(map[string]TypeInfo),
		parent: parent,
	}
}

// Define adds a variable to this scope
func (s *Scope) define(name string, t TypeInfo) {
	s.vars[name] = t
}

// Lookup searches for a variable type in this scope and parent scopes
func (s *Scope) lookup(name string) (TypeInfo, bool) {
	if t, exists := s.vars[name]; exists {
		return t, true
	}
	if s.parent != nil {
		return s.parent.lookup(name)
	}
	return UnknownType, false
}

// FuncDef represents a function signature
type FuncDef struct {
	Name       string
	ParamTypes []TypeInfo // parameter types in order
	ReturnType TypeInfo   // return type
}

// StructDef represents a struct definition
type StructDef struct {
	Name   string
	Fields map[string]TypeInfo // fieldname -> type
}

// TypeEnv is the complete type environment
type TypeEnv struct {
	current   *Scope
	structs   map[string]StructDef
	functions map[string]FuncDef
}

// NewTypeEnv creates a new type environment
func NewTypeEnv() *TypeEnv {
	global := newScope(nil)
	return &TypeEnv{
		current:   global,
		structs:   make(map[string]StructDef),
		functions: make(map[string]FuncDef),
	}
}

// EnterScope creates a new nested scope
func (e *TypeEnv) EnterScope() {
	e.current = newScope(e.current)
}

// ExitScope pops to parent scope
func (e *TypeEnv) ExitScope() {
	if e.current.parent != nil {
		e.current = e.current.parent
	}
}

// Define adds a variable to current scope
func (e *TypeEnv) Define(name string, t TypeInfo) {
	e.current.define(name, t)
}

// Lookup searches for a variable in current scope and parents
func (e *TypeEnv) Lookup(name string) (TypeInfo, bool) {
	return e.current.lookup(name)
}

// RegisterStruct adds a struct definition
func (e *TypeEnv) RegisterStruct(name string, def StructDef) {
	e.structs[name] = def
}

// LookupStruct finds a struct definition
func (e *TypeEnv) LookupStruct(name string) (StructDef, bool) {
	s, exists := e.structs[name]
	return s, exists
}

// RegisterFunc adds a function signature
func (e *TypeEnv) RegisterFunc(name string, def FuncDef) {
	e.functions[name] = def
}

// LookupFunc finds a function signature
func (e *TypeEnv) LookupFunc(name string) (FuncDef, bool) {
	f, exists := e.functions[name]
	return f, exists
}
