// Package builtin implements the standard library built-in functions for FreeLang
package builtin

// BuiltinDef represents a built-in function definition
type BuiltinDef struct {
	Name           string
	ParamTypeNames []string                             // parameter type names (e.g., "string", "int")
	ReturnTypeName string                               // return type name
	Impl           func(args ...interface{}) interface{} // implementation
}

// registry holds all registered built-in functions
var registry = make(map[string]BuiltinDef)

// Register adds a built-in function to the registry
func Register(b BuiltinDef) {
	registry[b.Name] = b
}

// Lookup finds a built-in function by name
func Lookup(name string) (BuiltinDef, bool) {
	def, exists := registry[name]
	return def, exists
}

// IsBuiltin checks if a name is a built-in function
func IsBuiltin(name string) bool {
	_, exists := registry[name]
	return exists
}

// AllDefs returns all registered built-in definitions
func AllDefs() []BuiltinDef {
	defs := make([]BuiltinDef, 0, len(registry))
	for _, def := range registry {
		defs = append(defs, def)
	}
	return defs
}

// init initializes the built-in registry by calling the register functions
// from io, string, and array packages
func init() {
	registerIOBuiltins()
	registerStringBuiltins()
	registerArrayBuiltins()
}
