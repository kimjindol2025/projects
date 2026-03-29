// Package typesys implements a type system for FreeLang
package typesys

// TypeKind represents the category of a type
type TypeKind int

const (
	TypeUnknown TypeKind = iota // 추론 불가 / 미결정
	TypeInt                     // 정수 타입
	TypeBool                    // 불리언 타입
	TypeString                  // 문자열 타입
	TypeUnit                    // 반환값 없음 (void)
	TypeStruct                  // struct 인스턴스
	TypeFn                      // 함수 타입
)

// TypeInfo represents complete type information
type TypeInfo struct {
	Kind       TypeKind
	StructName string     // TypeStruct일 때 struct 이름
	ParamTypes []TypeInfo // TypeFn일 때 파라미터 타입
	ReturnType *TypeInfo  // TypeFn일 때 반환 타입
}

// Common type constants
var (
	IntType     = TypeInfo{Kind: TypeInt}
	BoolType    = TypeInfo{Kind: TypeBool}
	StringType  = TypeInfo{Kind: TypeString}
	UnitType    = TypeInfo{Kind: TypeUnit}
	UnknownType = TypeInfo{Kind: TypeUnknown}
)

// StructType creates a struct type with given name
func StructType(name string) TypeInfo {
	return TypeInfo{Kind: TypeStruct, StructName: name}
}

// String returns a human-readable type name
func (t TypeInfo) String() string {
	switch t.Kind {
	case TypeInt:
		return "int"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypeUnit:
		return "unit"
	case TypeUnknown:
		return "unknown"
	case TypeStruct:
		return "struct:" + t.StructName
	case TypeFn:
		paramStr := "("
		for i, p := range t.ParamTypes {
			if i > 0 {
				paramStr += ","
			}
			paramStr += p.String()
		}
		paramStr += ")"
		retStr := "unit"
		if t.ReturnType != nil {
			retStr = t.ReturnType.String()
		}
		return paramStr + "->" + retStr
	default:
		return "unknown"
	}
}

// Equals checks if two types are equivalent
func (t TypeInfo) Equals(other TypeInfo) bool {
	if t.Kind != other.Kind {
		return false
	}
	if t.Kind == TypeStruct && t.StructName != other.StructName {
		return false
	}
	// For TypeFn, we'd need to check ParamTypes and ReturnType recursively
	// (simplified for Phase 1)
	return true
}

// TypeFromAnnotation converts a type annotation string to TypeInfo
func TypeFromAnnotation(s string) TypeInfo {
	switch s {
	case "int":
		return IntType
	case "bool":
		return BoolType
	case "string":
		return StringType
	case "unit":
		return UnitType
	default:
		// Assume it's a struct name
		return StructType(s)
	}
}
