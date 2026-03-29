package builtin

import (
	"fmt"
)

// registerIOBuiltins registers I/O functions (print, println)
func registerIOBuiltins() {
	// print(value: any) -> unit
	// Output: prints value without newline
	Register(BuiltinDef{
		Name:           "print",
		ParamTypeNames: []string{"any"},
		ReturnTypeName: "unit",
		Impl: func(args ...interface{}) interface{} {
			if len(args) > 0 {
				fmt.Print(args[0])
			}
			return nil
		},
	})

	// println(value: any) -> unit
	// Output: prints value with newline
	Register(BuiltinDef{
		Name:           "println",
		ParamTypeNames: []string{"any"},
		ReturnTypeName: "unit",
		Impl: func(args ...interface{}) interface{} {
			if len(args) > 0 {
				fmt.Println(args[0])
			} else {
				fmt.Println()
			}
			return nil
		},
	})
}
