package types

import (
	"fmt"
	"strings"
)

// Method represents a Julia method with its signature and implementation info
type Method struct {
	Name         string
	ParameterTypes []Type
	ReturnType   Type
	Variadic     bool
	ID           int
	Module       string // Which module this method belongs to
}

// Signature returns the method signature as a string
func (m *Method) Signature() string {
	parts := make([]string, len(m.ParameterTypes))
	for i, pt := range m.ParameterTypes {
		parts[i] = pt.String()
	}
	variadicStr := ""
	if m.Variadic {
		variadicStr = "..."
	}
	return fmt.Sprintf("%s(%s%s) -> %s", m.Name, strings.Join(parts, ","), variadicStr, m.ReturnType.String())
}

// MethodTable stores all methods for a function, supporting multiple dispatch
type MethodTable struct {
	name            string
	methods         []*Method
	hierarchy       *Hierarchy
	nextMethodID    int
	specificity     map[int]int // method ID -> specificity score
}

// NewMethodTable creates a new method table for a function
func NewMethodTable(name string, hierarchy *Hierarchy) *MethodTable {
	return &MethodTable{
		name:         name,
		methods:      make([]*Method, 0),
		hierarchy:    hierarchy,
		nextMethodID: 0,
		specificity: make(map[int]int),
	}
}

// AddMethod adds a new method to the table
func (mt *MethodTable) AddMethod(paramTypes []Type, returnType Type, variadic bool) *Method {
	method := &Method{
		Name:           mt.name,
		ParameterTypes: paramTypes,
		ReturnType:     returnType,
		Variadic:       variadic,
		ID:             mt.nextMethodID,
	}
	mt.nextMethodID++

	// Calculate and store specificity
	specificity := mt.calculateSpecificity(paramTypes)
	mt.specificity[method.ID] = specificity

	mt.methods = append(mt.methods, method)
	return method
}

// FindMethod finds the best matching method for given argument types (multiple dispatch)
// Returns the method and a confidence score (higher = better)
func (mt *MethodTable) FindMethod(argTypes []Type) (*Method, int) {
	if len(mt.methods) == 0 {
		return nil, 0
	}

	var bestMethod *Method
	bestScore := -1

	for _, method := range mt.methods {
		if score := mt.matchScore(method, argTypes); score > bestScore {
			bestScore = score
			bestMethod = method
		}
	}

	return bestMethod, bestScore
}

// FindAllMatching returns all methods that match the given argument types, ordered by specificity
func (mt *MethodTable) FindAllMatching(argTypes []Type) []*Method {
	var matches []*Method

	for _, method := range mt.methods {
		if mt.matchScore(method, argTypes) >= 0 {
			matches = append(matches, method)
		}
	}

	// Sort by specificity (most specific first)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if mt.specificity[matches[i].ID] < mt.specificity[matches[j].ID] {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}

// matchScore returns a match score:
//   0 = exact match
//   1-99 = increasing levels of subtype matching
//   -1 = no match
func (mt *MethodTable) matchScore(method *Method, argTypes []Type) int {
	// Check variadic handling
	if method.Variadic {
		if len(argTypes) < len(method.ParameterTypes)-1 {
			return -1
		}
	} else {
		if len(argTypes) != len(method.ParameterTypes) {
			return -1
		}
	}

	score := 0

	// Check each argument
	for i, argType := range argTypes {
		paramType := method.ParameterTypes[i]

		if argType.String() == paramType.String() {
			// Exact match
			score += 0
		} else if argType.IsSubtypeOf(paramType) {
			// Subtype match
			score += 10
		} else {
			// No match
			return -1
		}
	}

	return score
}

// calculateSpecificity calculates how specific a method signature is
// More specific = higher score
func (mt *MethodTable) calculateSpecificity(paramTypes []Type) int {
	score := 0

	for _, paramType := range paramTypes {
		// Concrete types are more specific than abstract
		if basicType, ok := paramType.(*BasicType); ok {
			if !mt.hierarchy.IsAbstractType(basicType.kind) {
				score += 10
			}
		}

		// Parametric types are more specific
		if _, ok := paramType.(*ParametricType); ok {
			score += 5
		}

		// Union types are less specific
		if _, ok := paramType.(*UnionType); ok {
			score -= 5
		}
	}

	return score
}

// Dispatch represents a multi-method dispatch system
type Dispatch struct {
	methodTables map[string]*MethodTable
	hierarchy    *Hierarchy
}

// NewDispatch creates a new dispatch system
func NewDispatch(hierarchy *Hierarchy) *Dispatch {
	return &Dispatch{
		methodTables: make(map[string]*MethodTable),
		hierarchy:    hierarchy,
	}
}

// RegisterFunction creates a method table for a function
func (d *Dispatch) RegisterFunction(name string) *MethodTable {
	mt := NewMethodTable(name, d.hierarchy)
	d.methodTables[name] = mt
	return mt
}

// GetMethodTable retrieves the method table for a function
func (d *Dispatch) GetMethodTable(name string) *MethodTable {
	return d.methodTables[name]
}

// LookupMethod finds the best method for a function call
func (d *Dispatch) LookupMethod(name string, argTypes []Type) (*Method, error) {
	mt, exists := d.methodTables[name]
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", name)
	}

	method, score := mt.FindMethod(argTypes)
	if score < 0 {
		argTypeStrs := make([]string, len(argTypes))
		for i, t := range argTypes {
			argTypeStrs[i] = t.String()
		}
		return nil, fmt.Errorf("no method found for %s(%s)", name, strings.Join(argTypeStrs, ","))
	}

	return method, nil
}

// ConversionCost estimates the cost of converting from one type to another
// Used for determining method preference when multiple methods match
func (d *Dispatch) ConversionCost(from, to Type) int {
	if from.String() == to.String() {
		return 0 // No conversion needed
	}

	if from.IsSubtypeOf(to) {
		return 1 // Simple upcast
	}

	// Check for numeric promotions
	if basicFrom, okFrom := from.(*BasicType); okFrom {
		if basicTo, okTo := to.(*BasicType); okTo {
			promoted := d.hierarchy.PromoteTypes(basicFrom.kind, basicTo.kind)
			if promoted == basicTo.kind {
				return 5 // Numeric promotion
			}
		}
	}

	return 100 // No conversion possible
}

// MethodResolution represents the result of method resolution
type MethodResolution struct {
	Method           *Method
	ExactMatch       bool
	AllowableMatches []*Method
	Ambiguous        bool
}

// ResolveMethod performs full method resolution with detailed results
func (d *Dispatch) ResolveMethod(name string, argTypes []Type) *MethodResolution {
	mt, exists := d.methodTables[name]
	if !exists {
		return &MethodResolution{Method: nil}
	}

	allMatches := mt.FindAllMatching(argTypes)
	if len(allMatches) == 0 {
		return &MethodResolution{Method: nil}
	}

	bestMethod := allMatches[0]
	isExact := len(allMatches) == 1 && mt.matchScore(bestMethod, argTypes) == 0

	ambiguous := false
	if len(allMatches) > 1 {
		// Check if top two methods have same score
		score1 := mt.matchScore(allMatches[0], argTypes)
		score2 := mt.matchScore(allMatches[1], argTypes)
		ambiguous = score1 == score2
	}

	return &MethodResolution{
		Method:           bestMethod,
		ExactMatch:       isExact,
		AllowableMatches: allMatches,
		Ambiguous:        ambiguous,
	}
}

// ListMethods returns all methods for a function
func (d *Dispatch) ListMethods(name string) []*Method {
	mt, exists := d.methodTables[name]
	if !exists {
		return nil
	}
	// Return a copy
	result := make([]*Method, len(mt.methods))
	copy(result, mt.methods)
	return result
}

// ListAllFunctions returns all registered function names
func (d *Dispatch) ListAllFunctions() []string {
	var names []string
	for name := range d.methodTables {
		names = append(names, name)
	}
	return names
}
