package builtin

// registerArrayBuiltins registers array manipulation functions
func registerArrayBuiltins() {
	// len_arr(arr: array) -> int
	// Returns the length of an array
	Register(BuiltinDef{
		Name:           "len_arr",
		ParamTypeNames: []string{"any"}, // array type
		ReturnTypeName: "int",
		Impl: func(args ...interface{}) interface{} {
			if len(args) == 0 {
				return 0
			}
			// Handle []interface{} type
			switch v := args[0].(type) {
			case []interface{}:
				return len(v)
			case []int:
				return len(v)
			case []string:
				return len(v)
			default:
				return 0
			}
		},
	})

	// append(arr: array, value: any) -> array
	// Appends a value to an array
	Register(BuiltinDef{
		Name:           "append",
		ParamTypeNames: []string{"any", "any"},
		ReturnTypeName: "any",
		Impl: func(args ...interface{}) interface{} {
			if len(args) < 2 {
				return []interface{}{}
			}
			// Handle []interface{} type
			arr, ok := args[0].([]interface{})
			if !ok {
				return []interface{}{args[1]}
			}
			return append(arr, args[1])
		},
	})

	// get(arr: array, index: int) -> any
	// Gets element at index from array
	Register(BuiltinDef{
		Name:           "get",
		ParamTypeNames: []string{"any", "int"},
		ReturnTypeName: "any",
		Impl: func(args ...interface{}) interface{} {
			if len(args) < 2 {
				return nil
			}
			arr, ok1 := args[0].([]interface{})
			idx, ok2 := args[1].(int)
			if !ok1 || !ok2 {
				return nil
			}
			if idx < 0 || idx >= len(arr) {
				return nil
			}
			return arr[idx]
		},
	})

	// set(arr: array, index: int, value: any) -> void
	// Sets element at index in array
	Register(BuiltinDef{
		Name:           "set",
		ParamTypeNames: []string{"any", "int", "any"},
		ReturnTypeName: "unit",
		Impl: func(args ...interface{}) interface{} {
			if len(args) < 3 {
				return nil
			}
			arr, ok1 := args[0].([]interface{})
			idx, ok2 := args[1].(int)
			if !ok1 || !ok2 {
				return nil
			}
			if idx < 0 || idx >= len(arr) {
				return nil
			}
			arr[idx] = args[2]
			return nil
		},
	})

	// slice(arr: array, start: int, end: int) -> array
	// Returns a slice of the array from start to end (exclusive)
	Register(BuiltinDef{
		Name:           "slice",
		ParamTypeNames: []string{"any", "int", "int"},
		ReturnTypeName: "any",
		Impl: func(args ...interface{}) interface{} {
			if len(args) < 3 {
				return []interface{}{}
			}
			arr, ok1 := args[0].([]interface{})
			start, ok2 := args[1].(int)
			end, ok3 := args[2].(int)
			if !ok1 || !ok2 || !ok3 {
				return []interface{}{}
			}
			if start < 0 || end > len(arr) || start > end {
				return []interface{}{}
			}
			return arr[start:end]
		},
	})
}
