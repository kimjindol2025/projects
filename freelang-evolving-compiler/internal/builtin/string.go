package builtin

import (
	"strings"
)

// registerStringBuiltins registers string manipulation functions
func registerStringBuiltins() {
	// len_str(s: string) -> int
	// Returns the length of a string
	Register(BuiltinDef{
		Name:           "len_str",
		ParamTypeNames: []string{"string"},
		ReturnTypeName: "int",
		Impl: func(args ...interface{}) interface{} {
			if len(args) == 0 {
				return 0
			}
			s, ok := args[0].(string)
			if !ok {
				return 0
			}
			return len(s)
		},
	})

	// concat(a: string, b: string) -> string
	// Concatenates two strings
	Register(BuiltinDef{
		Name:           "concat",
		ParamTypeNames: []string{"string", "string"},
		ReturnTypeName: "string",
		Impl: func(args ...interface{}) interface{} {
			if len(args) < 2 {
				return ""
			}
			a, ok1 := args[0].(string)
			b, ok2 := args[1].(string)
			if !ok1 || !ok2 {
				return ""
			}
			return a + b
		},
	})

	// substring(s: string, start: int, end: int) -> string
	// Returns a substring from start to end (exclusive)
	Register(BuiltinDef{
		Name:           "substring",
		ParamTypeNames: []string{"string", "int", "int"},
		ReturnTypeName: "string",
		Impl: func(args ...interface{}) interface{} {
			if len(args) < 3 {
				return ""
			}
			s, ok1 := args[0].(string)
			start, ok2 := args[1].(int)
			end, ok3 := args[2].(int)
			if !ok1 || !ok2 || !ok3 {
				return ""
			}
			if start < 0 || end > len(s) || start > end {
				return ""
			}
			return s[start:end]
		},
	})

	// upper(s: string) -> string
	// Converts string to uppercase
	Register(BuiltinDef{
		Name:           "upper",
		ParamTypeNames: []string{"string"},
		ReturnTypeName: "string",
		Impl: func(args ...interface{}) interface{} {
			if len(args) == 0 {
				return ""
			}
			s, ok := args[0].(string)
			if !ok {
				return ""
			}
			return strings.ToUpper(s)
		},
	})

	// lower(s: string) -> string
	// Converts string to lowercase
	Register(BuiltinDef{
		Name:           "lower",
		ParamTypeNames: []string{"string"},
		ReturnTypeName: "string",
		Impl: func(args ...interface{}) interface{} {
			if len(args) == 0 {
				return ""
			}
			s, ok := args[0].(string)
			if !ok {
				return ""
			}
			return strings.ToLower(s)
		},
	})

	// split(s: string, sep: string) -> string[]
	// Splits a string by separator
	Register(BuiltinDef{
		Name:           "split",
		ParamTypeNames: []string{"string", "string"},
		ReturnTypeName: "string", // simplified: would be array in full implementation
		Impl: func(args ...interface{}) interface{} {
			if len(args) < 2 {
				return []string{}
			}
			s, ok1 := args[0].(string)
			sep, ok2 := args[1].(string)
			if !ok1 || !ok2 {
				return []string{}
			}
			return strings.Split(s, sep)
		},
	})
}
