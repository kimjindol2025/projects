package types

// Hierarchy defines the Julia type hierarchy
// Julia's type system is complex with abstract types and multiple inheritance
type Hierarchy struct {
	registry    *Registry
	supertypes  map[TypeKind][]TypeKind
	subtypes    map[TypeKind][]TypeKind
	abstractMap map[string]TypeKind
}

// NewHierarchy creates a new type hierarchy
func NewHierarchy(reg *Registry) *Hierarchy {
	h := &Hierarchy{
		registry:    reg,
		supertypes:  make(map[TypeKind][]TypeKind),
		subtypes:    make(map[TypeKind][]TypeKind),
		abstractMap: make(map[string]TypeKind),
	}

	// Build the Julia type hierarchy
	// In Julia:
	//   Any (abstract root)
	//   ├── Nothing (bottom type)
	//   ├── Type
	//   ├── Number (abstract)
	//   │   ├── Integer (abstract)
	//   │   │   ├── Signed (abstract)
	//   │   │   │   ├── Int8, Int16, Int32, Int64
	//   │   │   └── Unsigned (abstract)
	//   │   │       ├── UInt8, UInt16, UInt32, UInt64
	//   │   ├── Floating (abstract)
	//   │   │   ├── Float32, Float64
	//   │   ├── Complex (abstract)
	//   │   │   ├── Complex64, Complex128
	//   │   ├── Rational (abstract)
	//   ├── Bool
	//   ├── String
	//   ├── Char
	//   ├── Symbol
	//   ├── Array (generic container)
	//   │   ├── Vector (1D)
	//   │   ├── Matrix (2D)
	//   ├── Tuple
	//   ├── Dict
	//   ├── Function
	//   └── ...

	h.registerAbstract("Any", KindAny)
	h.registerAbstract("Number", KindNumber)
	h.registerAbstract("Integer", KindInteger)
	h.registerAbstract("Floating", KindFloating)
	h.registerAbstract("Signed", KindSigned)
	h.registerAbstract("Unsigned", KindUnsigned)

	// Establish hierarchy
	h.addSupertype(KindNothing, KindAny)
	h.addSupertype(KindType, KindAny)
	h.addSupertype(KindBool, KindAny)
	h.addSupertype(KindString, KindAny)
	h.addSupertype(KindChar, KindAny)
	h.addSupertype(KindSymbol, KindAny)
	h.addSupertype(KindArray, KindAny)
	h.addSupertype(KindVector, KindArray)
	h.addSupertype(KindMatrix, KindArray)
	h.addSupertype(KindTuple, KindAny)
	h.addSupertype(KindDict, KindAny)
	h.addSupertype(KindSet, KindAny)
	h.addSupertype(KindFunction, KindAny)

	// Number hierarchy
	h.addSupertype(KindNumber, KindAny)
	h.addSupertype(KindInteger, KindNumber)
	h.addSupertype(KindFloating, KindNumber)
	h.addSupertype(KindSigned, KindInteger)
	h.addSupertype(KindUnsigned, KindInteger)

	// Concrete integer types
	h.addSupertype(KindInt8, KindSigned)
	h.addSupertype(KindInt16, KindSigned)
	h.addSupertype(KindInt32, KindSigned)
	h.addSupertype(KindInt64, KindSigned)

	h.addSupertype(KindUInt8, KindUnsigned)
	h.addSupertype(KindUInt16, KindUnsigned)
	h.addSupertype(KindUInt32, KindUnsigned)
	h.addSupertype(KindUInt64, KindUnsigned)

	// Floating point types
	h.addSupertype(KindFloat32, KindFloating)
	h.addSupertype(KindFloat64, KindFloating)

	// Complex types
	h.addSupertype(KindComplex64, KindNumber)
	h.addSupertype(KindComplex128, KindNumber)

	return h
}

func (h *Hierarchy) registerAbstract(name string, kind TypeKind) {
	h.abstractMap[name] = kind
}

func (h *Hierarchy) addSupertype(subKind, superKind TypeKind) {
	if _, exists := h.supertypes[subKind]; !exists {
		h.supertypes[subKind] = []TypeKind{}
	}
	h.supertypes[subKind] = append(h.supertypes[subKind], superKind)

	if _, exists := h.subtypes[superKind]; !exists {
		h.subtypes[superKind] = []TypeKind{}
	}
	h.subtypes[superKind] = append(h.subtypes[superKind], subKind)
}

// IsAbstractType checks if a type is abstract
func (h *Hierarchy) IsAbstractType(kind TypeKind) bool {
	switch kind {
	case KindAny, KindNumber, KindInteger, KindFloating, KindSigned, KindUnsigned:
		return true
	default:
		return false
	}
}

// GetSupertypes returns all supertypes of a kind (direct and indirect)
func (h *Hierarchy) GetSupertypes(kind TypeKind) []TypeKind {
	var result []TypeKind
	visited := make(map[TypeKind]bool)
	h.getSupertypesRec(kind, &result, visited)
	return result
}

func (h *Hierarchy) getSupertypesRec(kind TypeKind, result *[]TypeKind, visited map[TypeKind]bool) {
	if visited[kind] {
		return
	}
	visited[kind] = true

	if supers, exists := h.supertypes[kind]; exists {
		for _, super := range supers {
			*result = append(*result, super)
			h.getSupertypesRec(super, result, visited)
		}
	}
}

// GetSubtypes returns all subtypes of a kind (direct and indirect)
func (h *Hierarchy) GetSubtypes(kind TypeKind) []TypeKind {
	var result []TypeKind
	visited := make(map[TypeKind]bool)
	h.getSubtypesRec(kind, &result, visited)
	return result
}

func (h *Hierarchy) getSubtypesRec(kind TypeKind, result *[]TypeKind, visited map[TypeKind]bool) {
	if visited[kind] {
		return
	}
	visited[kind] = true

	if subs, exists := h.subtypes[kind]; exists {
		for _, sub := range subs {
			*result = append(*result, sub)
			h.getSubtypesRec(sub, result, visited)
		}
	}
}

// IsSubtype checks if sub is a subtype of super in the hierarchy
func (h *Hierarchy) IsSubtype(sub, super TypeKind) bool {
	if sub == super {
		return true
	}

	visited := make(map[TypeKind]bool)
	return h.isSubtypeRec(sub, super, visited)
}

func (h *Hierarchy) isSubtypeRec(sub, super TypeKind, visited map[TypeKind]bool) bool {
	if visited[sub] {
		return false
	}
	visited[sub] = true

	if supers, exists := h.supertypes[sub]; exists {
		for _, s := range supers {
			if s == super {
				return true
			}
			if h.isSubtypeRec(s, super, visited) {
				return true
			}
		}
	}

	return false
}

// CommonSupertype finds the most specific common supertype of two kinds
func (h *Hierarchy) CommonSupertype(kind1, kind2 TypeKind) TypeKind {
	if kind1 == kind2 {
		return kind1
	}

	// Get all supertypes of kind1
	supers1 := make(map[TypeKind]bool)
	supers1[kind1] = true
	visited := make(map[TypeKind]bool)
	h.collectSupertypesRec(kind1, supers1, visited)

	// Find closest common supertype by checking kind2's hierarchy
	return h.findCommonRec(kind2, supers1, make(map[TypeKind]bool))
}

func (h *Hierarchy) collectSupertypesRec(kind TypeKind, result map[TypeKind]bool, visited map[TypeKind]bool) {
	if visited[kind] {
		return
	}
	visited[kind] = true

	if supers, exists := h.supertypes[kind]; exists {
		for _, super := range supers {
			result[super] = true
			h.collectSupertypesRec(super, result, visited)
		}
	}
}

func (h *Hierarchy) findCommonRec(kind TypeKind, common map[TypeKind]bool, visited map[TypeKind]bool) TypeKind {
	if visited[kind] {
		return KindAny
	}
	visited[kind] = true

	if common[kind] {
		return kind
	}

	if supers, exists := h.supertypes[kind]; exists {
		for _, super := range supers {
			result := h.findCommonRec(super, common, visited)
			if result != KindAny {
				return result
			}
		}
	}

	return KindAny
}

// Numeric type promotion
// In Julia, operations may promote types: Int64 + Float64 -> Float64
func (h *Hierarchy) PromoteTypes(kind1, kind2 TypeKind) TypeKind {
	// Same type, no promotion
	if kind1 == kind2 {
		return kind1
	}

	// Anything + Nothing -> Anything
	if kind2 == KindNothing {
		return kind1
	}
	if kind1 == KindNothing {
		return kind2
	}

	// Integer + Integer
	if h.IsSubtype(kind1, KindInteger) && h.IsSubtype(kind2, KindInteger) {
		return h.promoteIntTypes(kind1, kind2)
	}

	// Float + Anything numeric -> Float64 (or larger)
	if h.IsSubtype(kind1, KindFloating) || h.IsSubtype(kind2, KindFloating) {
		// Promote to largest float
		if kind1 == KindFloat64 || kind2 == KindFloat64 {
			return KindFloat64
		}
		// When mixing Float32 with integers, promote to Float64
		if (kind1 == KindFloat32 && h.IsSubtype(kind2, KindInteger)) ||
			(kind2 == KindFloat32 && h.IsSubtype(kind1, KindInteger)) {
			return KindFloat64
		}
		if kind1 == KindFloat32 || kind2 == KindFloat32 {
			return KindFloat32
		}
	}

	// Complex + Anything numeric -> Complex128
	if h.IsSubtype(kind1, KindComplex128) || h.IsSubtype(kind2, KindComplex128) {
		return KindComplex128
	}
	if h.IsSubtype(kind1, KindComplex64) || h.IsSubtype(kind2, KindComplex64) {
		return KindComplex64
	}

	// Default: Any
	return KindAny
}

func (h *Hierarchy) promoteIntTypes(kind1, kind2 TypeKind) TypeKind {
	// Get bit widths
	width1 := h.getIntWidth(kind1)
	width2 := h.getIntWidth(kind2)

	// Promote to wider type
	if width1 > width2 {
		return kind1
	}
	if width2 > width1 {
		return kind2
	}

	// Same width, prefer signed
	if h.IsSubtype(kind1, KindSigned) {
		return kind1
	}
	return kind2
}

func (h *Hierarchy) getIntWidth(kind TypeKind) int {
	switch kind {
	case KindInt8, KindUInt8:
		return 8
	case KindInt16, KindUInt16:
		return 16
	case KindInt32, KindUInt32:
		return 32
	case KindInt64, KindUInt64:
		return 64
	default:
		return 0
	}
}
