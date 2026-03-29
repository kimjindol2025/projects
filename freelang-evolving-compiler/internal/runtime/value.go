// Package runtime implements the VM and runtime for FreeLang
package runtime

import "fmt"

// ValueKind represents the type of a value at runtime
type ValueKind int

const (
	KindNil ValueKind = iota
	KindInt
	KindBool
	KindString
	KindStruct
	KindArray
)

// Value represents a runtime value in FreeLang
type Value struct {
	Kind   ValueKind
	IVal   int64             // for KindInt
	BVal   bool              // for KindBool
	SVal   string            // for KindString
	Fields map[string]Value  // for KindStruct (keys: "f0", "f1", ...)
	Elems  []Value           // for KindArray
}

// Constructors
func NilVal() Value {
	return Value{Kind: KindNil}
}

func IntVal(n int64) Value {
	return Value{Kind: KindInt, IVal: n}
}

func BoolVal(b bool) Value {
	return Value{Kind: KindBool, BVal: b}
}

func StringVal(s string) Value {
	return Value{Kind: KindString, SVal: s}
}

func StructVal() Value {
	return Value{Kind: KindStruct, Fields: make(map[string]Value)}
}

func ArrayVal(elems []Value) Value {
	return Value{Kind: KindArray, Elems: elems}
}

// Truthy returns whether a value is truthy in a conditional context
func (v Value) Truthy() bool {
	switch v.Kind {
	case KindNil:
		return false
	case KindBool:
		return v.BVal
	case KindInt:
		return v.IVal != 0
	case KindString:
		return v.SVal != ""
	case KindArray:
		return len(v.Elems) > 0
	case KindStruct:
		return len(v.Fields) > 0
	default:
		return true
	}
}

// String returns a human-readable representation
func (v Value) String() string {
	switch v.Kind {
	case KindNil:
		return "nil"
	case KindInt:
		return fmt.Sprintf("%d", v.IVal)
	case KindBool:
		return fmt.Sprintf("%v", v.BVal)
	case KindString:
		return fmt.Sprintf("%q", v.SVal)
	case KindArray:
		result := "["
		for i, e := range v.Elems {
			if i > 0 {
				result += ", "
			}
			result += e.String()
		}
		result += "]"
		return result
	case KindStruct:
		return fmt.Sprintf("struct{%d fields}", len(v.Fields))
	default:
		return "unknown"
	}
}

// ToInterface converts a Value to interface{} for builtin function calls
func (v Value) ToInterface() interface{} {
	switch v.Kind {
	case KindNil:
		return nil
	case KindInt:
		return v.IVal
	case KindBool:
		return v.BVal
	case KindString:
		return v.SVal
	case KindArray:
		result := make([]interface{}, len(v.Elems))
		for i, e := range v.Elems {
			result[i] = e.ToInterface()
		}
		return result
	case KindStruct:
		return map[string]interface{}{"_type": "struct"}
	default:
		return nil
	}
}

// FromInterface creates a Value from interface{} (builtin return values)
func FromInterface(v interface{}) Value {
	if v == nil {
		return NilVal()
	}

	switch val := v.(type) {
	case int:
		return IntVal(int64(val))
	case int64:
		return IntVal(val)
	case bool:
		return BoolVal(val)
	case string:
		return StringVal(val)
	case []interface{}:
		elems := make([]Value, len(val))
		for i, e := range val {
			elems[i] = FromInterface(e)
		}
		return ArrayVal(elems)
	case []string:
		elems := make([]Value, len(val))
		for i, s := range val {
			elems[i] = StringVal(s)
		}
		return ArrayVal(elems)
	default:
		return NilVal()
	}
}
