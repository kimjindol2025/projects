package typesys

import (
	"fmt"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

// TypeError represents a type checking error
type TypeError struct {
	Message string
	Line    int
	Col     int
}

// TypeChecker validates types in an AST
type TypeChecker struct {
	env    *TypeEnv
	errors []TypeError
}

// NewTypeChecker creates a new type checker
func NewTypeChecker() *TypeChecker {
	return &TypeChecker{
		env:    NewTypeEnv(),
		errors: []TypeError{},
	}
}

// Check walks the AST and validates types, returning any errors found
func (tc *TypeChecker) Check(root *ast.Node) []TypeError {
	if root != nil {
		tc.checkNode(root)
	}
	return tc.errors
}

// checkNode recursively checks a node and returns its inferred type
func (tc *TypeChecker) checkNode(n *ast.Node) TypeInfo {
	if n == nil {
		return UnknownType
	}

	switch n.Kind {
	case ast.NodeProgram:
		// Check all top-level statements
		for _, child := range n.Children {
			tc.checkNode(child)
		}
		return UnitType

	case ast.NodeStructDecl:
		return tc.checkStructDecl(n)

	case ast.NodeLetDecl:
		return tc.checkLetDecl(n)

	case ast.NodeFnDecl:
		return tc.checkFnDecl(n)

	case ast.NodeBlockStmt:
		tc.env.EnterScope()
		for _, child := range n.Children {
			tc.checkNode(child)
		}
		tc.env.ExitScope()
		return UnitType

	case ast.NodeIfStmt:
		if len(n.Children) > 0 {
			condType := tc.checkNode(n.Children[0])
			if condType.Kind != TypeBool && condType.Kind != TypeUnknown {
				tc.addError(fmt.Sprintf("if condition must be bool, got %v", condType), n.Line, n.Col)
			}
		}
		if len(n.Children) > 1 {
			tc.checkNode(n.Children[1])
		}
		return UnitType

	case ast.NodeReturn:
		if len(n.Children) > 0 {
			return tc.checkNode(n.Children[0])
		}
		return UnitType

	case ast.NodeBinaryExpr:
		return tc.checkBinaryExpr(n)

	case ast.NodeFieldAccess:
		return tc.checkFieldAccess(n)

	case ast.NodeCallExpr:
		return tc.checkCallExpr(n)

	case ast.NodeIdent:
		if t, found := tc.env.Lookup(n.Value); found {
			return t
		}
		tc.addError(fmt.Sprintf("unknown variable '%s'", n.Value), n.Line, n.Col)
		return UnknownType

	case ast.NodeIntLit:
		return IntType

	case ast.NodeForStmt:
		// Register iterator variable as int (simplified)
		if len(n.Children) > 0 && n.Children[0].Kind == ast.NodeIdent {
			tc.env.Define(n.Children[0].Value, IntType)
		}
		if len(n.Children) > 2 {
			tc.checkNode(n.Children[2]) // body
		}
		return UnitType

	default:
		return UnknownType
	}
}

// checkStructDecl registers a struct definition
func (tc *TypeChecker) checkStructDecl(n *ast.Node) TypeInfo {
	structName := n.Value
	def := StructDef{
		Name:   structName,
		Fields: make(map[string]TypeInfo),
	}

	// Extract field types from children
	for _, child := range n.Children {
		if child.Kind == ast.NodeFieldDecl {
			fieldName := child.Value
			// Field type is in TypeAnnotation or Children[0].Value
			fieldType := UnknownType
			if child.TypeAnnotation != "" {
				fieldType = TypeFromAnnotation(child.TypeAnnotation)
			} else if len(child.Children) > 0 {
				// Type is in Children[0].Value
				fieldType = TypeFromAnnotation(child.Children[0].Value)
			}
			def.Fields[fieldName] = fieldType
		}
	}

	tc.env.RegisterStruct(structName, def)
	return UnitType
}

// checkLetDecl validates variable declaration and registration
func (tc *TypeChecker) checkLetDecl(n *ast.Node) TypeInfo {
	if len(n.Children) < 2 {
		return UnknownType
	}

	nameNode := n.Children[0]
	valueNode := n.Children[1]

	// Determine the type of the variable
	var varType TypeInfo

	// If there's a type annotation, use it
	if nameNode.TypeAnnotation != "" {
		varType = TypeFromAnnotation(nameNode.TypeAnnotation)
	} else {
		// Infer from the value expression
		varType = tc.checkNode(valueNode)
	}

	// Check value type matches declared type
	valueType := tc.checkNode(valueNode)
	if nameNode.TypeAnnotation != "" && valueType.Kind != TypeUnknown {
		if !varType.Equals(valueType) {
			tc.addError(
				fmt.Sprintf("type mismatch: expected %v, got %v", varType, valueType),
				n.Line, n.Col,
			)
		}
	}

	// Register the variable
	tc.env.Define(nameNode.Value, varType)
	return UnitType
}

// checkFnDecl validates function declaration
func (tc *TypeChecker) checkFnDecl(n *ast.Node) TypeInfo {
	// Register function parameters
	tc.env.EnterScope()
	for _, child := range n.Children {
		if child.Kind == ast.NodeIdent {
			// This is a parameter
			paramType := UnknownType
			if child.TypeAnnotation != "" {
				paramType = TypeFromAnnotation(child.TypeAnnotation)
			}
			tc.env.Define(child.Value, paramType)
		}
	}

	// Check function body
	for _, child := range n.Children {
		if child.Kind == ast.NodeBlockStmt {
			tc.checkNode(child)
			break
		}
	}

	tc.env.ExitScope()

	// Register function type (simplified for Phase 1)
	return UnitType
}

// checkBinaryExpr validates binary expression type correctness
func (tc *TypeChecker) checkBinaryExpr(n *ast.Node) TypeInfo {
	if len(n.Children) < 2 {
		return UnknownType
	}

	leftType := tc.checkNode(n.Children[0])
	rightType := tc.checkNode(n.Children[1])

	// For now, both operands must be int for arithmetic
	op := n.Value
	if op == "+" || op == "-" || op == "*" || op == "/" {
		if leftType.Kind != TypeInt && leftType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("left operand of %s must be int, got %v", op, leftType),
				n.Line, n.Col,
			)
		}
		if rightType.Kind != TypeInt && rightType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("right operand of %s must be int, got %v", op, rightType),
				n.Line, n.Col,
			)
		}
		return IntType
	}

	// Comparison operators return bool
	if op == "==" || op == "!=" || op == "<" || op == ">" || op == "<=" || op == ">=" {
		if !leftType.Equals(rightType) && leftType.Kind != TypeUnknown && rightType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("type mismatch in %s: %v vs %v", op, leftType, rightType),
				n.Line, n.Col,
			)
		}
		return BoolType
	}

	return UnknownType
}

// checkFieldAccess validates field access and returns field type
func (tc *TypeChecker) checkFieldAccess(n *ast.Node) TypeInfo {
	if len(n.Children) == 0 {
		return UnknownType
	}

	// Get the object type
	objType := tc.checkNode(n.Children[0])
	fieldName := n.Value

	// Look up struct definition
	if objType.Kind == TypeStruct {
		if structDef, found := tc.env.LookupStruct(objType.StructName); found {
			if fieldType, fieldExists := structDef.Fields[fieldName]; fieldExists {
				return fieldType
			}
			tc.addError(
				fmt.Sprintf("unknown field '%s' on struct '%s'", fieldName, objType.StructName),
				n.Line, n.Col,
			)
			return UnknownType
		}
		tc.addError(fmt.Sprintf("unknown struct '%s'", objType.StructName), n.Line, n.Col)
		return UnknownType
	}

	// Object is not a struct
	if objType.Kind != TypeUnknown {
		tc.addError(
			fmt.Sprintf("cannot access field '%s' on non-struct type %v", fieldName, objType),
			n.Line, n.Col,
		)
	}
	return UnknownType
}

// checkCallExpr validates function call (simplified for Phase 1)
func (tc *TypeChecker) checkCallExpr(n *ast.Node) TypeInfo {
	// Phase 1: Just validate that the function exists
	name := n.Value
	if _, found := tc.env.Lookup(name); !found && name != "print" && name != "len" && name != "str" {
		tc.addError(fmt.Sprintf("unknown function '%s'", name), n.Line, n.Col)
	}
	return UnknownType
}

// addError records a type error
func (tc *TypeChecker) addError(msg string, line, col int) {
	tc.errors = append(tc.errors, TypeError{
		Message: msg,
		Line:    line,
		Col:     col,
	})
}
