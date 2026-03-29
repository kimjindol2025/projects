package typesys

import (
	"fmt"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/builtin"
)

// TypeError represents a type checking error
type TypeError struct {
	Message string
	Line    int
	Col     int
}

// TypeChecker validates types in an AST
type TypeChecker struct {
	env      *TypeEnv
	errors   []TypeError
	hardMode bool // if true, type errors are fatal
}

// NewTypeChecker creates a new type checker in soft mode
func NewTypeChecker() *TypeChecker {
	env := NewTypeEnv()
	registerBuiltinsInEnv(env)
	return &TypeChecker{
		env:      env,
		errors:   []TypeError{},
		hardMode: false,
	}
}

// NewTypeCheckerHard creates a new type checker in hard mode
func NewTypeCheckerHard() *TypeChecker {
	env := NewTypeEnv()
	registerBuiltinsInEnv(env)
	return &TypeChecker{
		env:      env,
		errors:   []TypeError{},
		hardMode: true,
	}
}

// registerBuiltinsInEnv registers all built-in function signatures in the environment
func registerBuiltinsInEnv(env *TypeEnv) {
	for _, def := range builtin.AllDefs() {
		// Convert type names to TypeInfo
		paramTypes := make([]TypeInfo, len(def.ParamTypeNames))
		for i, typeName := range def.ParamTypeNames {
			paramTypes[i] = TypeFromAnnotation(typeName)
		}
		returnType := TypeFromAnnotation(def.ReturnTypeName)

		env.RegisterFunc(def.Name, FuncDef{
			Name:       def.Name,
			ParamTypes: paramTypes,
			ReturnType: returnType,
		})
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
		// Check else branch if present
		if len(n.Children) > 2 {
			tc.checkNode(n.Children[2])
		}
		return UnitType

	case ast.NodeReturn:
		if len(n.Children) > 0 {
			return tc.checkNode(n.Children[0])
		}
		return UnitType

	case ast.NodeBinaryExpr:
		return tc.checkBinaryExpr(n)

	case ast.NodeUnaryExpr:
		return tc.checkUnaryExpr(n)

	case ast.NodeLogicalExpr:
		return tc.checkLogicalExpr(n)

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

	case ast.NodeBoolLit:
		return BoolType

	case ast.NodeStringLit:
		return StringType

	case ast.NodeStructLit:
		return tc.checkStructLit(n)

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
		// Check value type matches declared type
		inferredType := tc.inferType(valueNode)
		if inferredType.Kind != TypeUnknown && !varType.Equals(inferredType) {
			tc.addError(
				fmt.Sprintf("type mismatch: expected %v, got %v", varType, inferredType),
				n.Line, n.Col,
			)
		}
	} else {
		// No annotation: infer from the value expression
		varType = tc.inferType(valueNode)
	}

	// Register the variable
	tc.env.Define(nameNode.Value, varType)
	return UnitType
}

// checkFnDecl validates function declaration and registers function signature
func (tc *TypeChecker) checkFnDecl(n *ast.Node) TypeInfo {
	fnName := n.Value

	// Collect parameter types
	paramTypes := []TypeInfo{}
	tc.env.EnterScope()

	for _, child := range n.Children {
		if child.Kind == ast.NodeIdent {
			// This is a parameter
			paramType := UnknownType
			if child.TypeAnnotation != "" {
				paramType = TypeFromAnnotation(child.TypeAnnotation)
			}
			paramTypes = append(paramTypes, paramType)
			tc.env.Define(child.Value, paramType)
		}
	}

	// Get return type from fn node's TypeAnnotation
	returnType := UnitType
	if n.TypeAnnotation != "" {
		returnType = TypeFromAnnotation(n.TypeAnnotation)
	}

	// Check function body
	for _, child := range n.Children {
		if child.Kind == ast.NodeBlockStmt {
			tc.checkNode(child)
			break
		}
	}

	tc.env.ExitScope()

	// Register function signature
	fnDef := FuncDef{
		Name:       fnName,
		ParamTypes: paramTypes,
		ReturnType: returnType,
	}
	tc.env.RegisterFunc(fnName, fnDef)

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

	// Logical operators: both operands must be bool, result is bool
	if op == "&&" || op == "||" {
		if leftType.Kind != TypeBool && leftType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("left operand of %s must be bool, got %v", op, leftType),
				n.Line, n.Col,
			)
		}
		if rightType.Kind != TypeBool && rightType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("right operand of %s must be bool, got %v", op, rightType),
				n.Line, n.Col,
			)
		}
		return BoolType
	}

	return UnknownType
}

// checkUnaryExpr validates unary expression (!, -)
func (tc *TypeChecker) checkUnaryExpr(n *ast.Node) TypeInfo {
	if len(n.Children) < 1 {
		return UnknownType
	}

	operandType := tc.checkNode(n.Children[0])
	op := n.Value

	switch op {
	case "!":
		// Logical NOT: operand must be bool, result is bool
		if operandType.Kind != TypeBool && operandType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("operand of ! must be bool, got %v", operandType),
				n.Line, n.Col,
			)
		}
		return BoolType

	case "-":
		// Arithmetic negation: operand must be int, result is int
		if operandType.Kind != TypeInt && operandType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("operand of - must be int, got %v", operandType),
				n.Line, n.Col,
			)
		}
		return IntType

	default:
		return UnknownType
	}
}

// checkLogicalExpr validates logical expression (&&, ||)
func (tc *TypeChecker) checkLogicalExpr(n *ast.Node) TypeInfo {
	if len(n.Children) < 2 {
		return UnknownType
	}

	leftType := tc.checkNode(n.Children[0])
	rightType := tc.checkNode(n.Children[1])

	op := n.Value

	// Both && and || require bool operands and return bool
	if op == "&&" || op == "||" {
		if leftType.Kind != TypeBool && leftType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("left operand of %s must be bool, got %v", op, leftType),
				n.Line, n.Col,
			)
		}
		if rightType.Kind != TypeBool && rightType.Kind != TypeUnknown {
			tc.addError(
				fmt.Sprintf("right operand of %s must be bool, got %v", op, rightType),
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

// checkCallExpr validates function call and returns return type
func (tc *TypeChecker) checkCallExpr(n *ast.Node) TypeInfo {
	fnName := n.Value

	// Check for user-defined or built-in function
	if fnDef, found := tc.env.LookupFunc(fnName); found {
		// Validate argument count
		if len(n.Children) != len(fnDef.ParamTypes) {
			tc.addError(
				fmt.Sprintf("function '%s' expects %d arguments, got %d", fnName, len(fnDef.ParamTypes), len(n.Children)),
				n.Line, n.Col,
			)
		}
		// Return the function's return type
		return fnDef.ReturnType
	}

	// Unknown function
	tc.addError(fmt.Sprintf("unknown function '%s'", fnName), n.Line, n.Col)
	return UnknownType
}

// inferType infers the type of a node without modifying the environment
func (tc *TypeChecker) inferType(n *ast.Node) TypeInfo {
	if n == nil {
		return UnknownType
	}

	switch n.Kind {
	case ast.NodeIntLit:
		return IntType
	case ast.NodeBoolLit:
		return BoolType
	case ast.NodeStringLit:
		return StringType
	case ast.NodeStructLit:
		// Infer struct type from the struct name
		return StructType(n.Value)
	case ast.NodeIdent:
		// Look up variable type
		if t, found := tc.env.Lookup(n.Value); found {
			return t
		}
		return UnknownType
	case ast.NodeBinaryExpr:
		return tc.checkBinaryExpr(n)
	case ast.NodeFieldAccess:
		return tc.checkFieldAccess(n)
	case ast.NodeCallExpr:
		return tc.checkCallExpr(n)
	default:
		return UnknownType
	}
}

// checkStructLit validates struct initialization
func (tc *TypeChecker) checkStructLit(n *ast.Node) TypeInfo {
	structName := n.Value

	// Look up struct definition
	if structDef, found := tc.env.LookupStruct(structName); found {
		// Validate field initializations
		for _, child := range n.Children {
			if child.Kind == ast.NodeFieldDecl {
				fieldName := child.Value
				if _, fieldExists := structDef.Fields[fieldName]; !fieldExists {
					tc.addError(
						fmt.Sprintf("unknown field '%s' in struct '%s'", fieldName, structName),
						child.Line, child.Col,
					)
				}
				// Check field value type if we have field type info
				if len(child.Children) > 0 {
					fieldValueType := tc.inferType(child.Children[0])
					expectedType := structDef.Fields[fieldName]
					if fieldValueType.Kind != TypeUnknown && expectedType.Kind != TypeUnknown && !fieldValueType.Equals(expectedType) {
						tc.addError(
							fmt.Sprintf("field '%s' type mismatch: expected %v, got %v", fieldName, expectedType, fieldValueType),
							child.Line, child.Col,
						)
					}
				}
			}
		}
		return StructType(structName)
	}

	// Struct not found
	tc.addError(fmt.Sprintf("unknown struct '%s'", structName), n.Line, n.Col)
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
