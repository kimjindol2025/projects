package ir

import (
	"fmt"
	"juliacc/internal/ast"
	"juliacc/internal/lexer"
)

// Builder converts analyzed AST to IR
type Builder struct {
	module    *Module
	currBlock *BasicBlock
	currFunc  *Function
	nextValID uint32
}

// NewBuilder creates a new IR builder
func NewBuilder() *Builder {
	return &Builder{
		module:    NewModule(),
		nextValID: 0,
	}
}

// Build converts an AST program to IR
func (b *Builder) Build(statements []ast.Stmt) (*Module, error) {
	for _, stmt := range statements {
		if err := b.buildStmt(stmt); err != nil {
			return nil, err
		}
	}
	return b.module, nil
}

// buildStmt builds a statement
func (b *Builder) buildStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.FunctionDecl:
		return b.buildFunctionDecl(s)
	case *ast.Assignment:
		_, err := b.buildExpr(s.Expression)
		return err
	default:
		// Try as expression statement
		if exprStmt, ok := stmt.(ast.Expr); ok {
			_, err := b.buildExpr(exprStmt)
			return err
		}
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

// buildFunctionDecl builds a function declaration
func (b *Builder) buildFunctionDecl(fd *ast.FunctionDecl) error {
	// Create function
	params := []Value{}
	for _, param := range fd.Parameters {
		params = append(params, Value{
			ID:   b.nextValID,
			Name: param.Name,
			Type: "i64", // Simplified: assume i64
		})
		b.nextValID++
	}

	fn := NewFunction(fd.Name, "i64", params)
	b.currFunc = fn

	// Create entry block
	entryBlock := NewBasicBlock("entry")
	fn.AddBlock(entryBlock)
	b.currBlock = entryBlock

	// Build function body
	for _, stmt := range fd.Body {
		if err := b.buildStmt(stmt); err != nil {
			return err
		}
	}

	// Add return instruction if not present
	if len(b.currBlock.Insts) == 0 || b.currBlock.Insts[len(b.currBlock.Insts)-1].Type != InstReturn {
		// Default return 0
		lit := &Instruction{
			Type: InstLiteral,
			Meta: int64(0),
			Result: Value{
				ID:   b.nextValID,
				Type: "i64",
			},
		}
		b.currBlock.AddInstruction(lit)
		b.nextValID++

		retInst := &Instruction{
			Type: InstReturn,
			Ops:  []Value{lit.Result},
		}
		b.currBlock.AddInstruction(retInst)
	}

	b.module.Functions = append(b.module.Functions, fn)
	return nil
}

// buildExpr builds an expression
func (b *Builder) buildExpr(expr ast.Expr) (Value, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		return b.buildLiteral(e)

	case *ast.Identifier:
		loadInst := &Instruction{
			Type: InstLoad,
			Ops:  []Value{{Name: e.Name}},
			Result: Value{
				ID:   b.nextValID,
				Type: "i64",
			},
		}
		b.currBlock.AddInstruction(loadInst)
		b.nextValID++
		return loadInst.Result, nil

	case *ast.BinaryOp:
		left, err := b.buildExpr(e.Left)
		if err != nil {
			return Value{}, err
		}
		right, err := b.buildExpr(e.Right)
		if err != nil {
			return Value{}, err
		}

		opType := typeFromOp(e.Op)
		binOp := &Instruction{
			Type:   InstBinOp,
			OpType: opType,
			Ops:    []Value{left, right},
			Result: Value{
				ID:   b.nextValID,
				Type: "i64",
			},
		}
		b.currBlock.AddInstruction(binOp)
		b.nextValID++
		return binOp.Result, nil

	case *ast.UnaryOp:
		operand, err := b.buildExpr(e.Operand)
		if err != nil {
			return Value{}, err
		}

		opStr := e.OpToken.Lexeme
		unaryOp := &Instruction{
			Type:   InstUnaryOp,
			OpType: opStr,
			Ops:    []Value{operand},
			Result: Value{
				ID:   b.nextValID,
				Type: "i64",
			},
		}
		b.currBlock.AddInstruction(unaryOp)
		b.nextValID++
		return unaryOp.Result, nil

	case *ast.Call:
		args := []Value{}
		for _, arg := range e.Arguments {
			val, err := b.buildExpr(arg)
			if err != nil {
				return Value{}, err
			}
			args = append(args, val)
		}

		fnName := ""
		if id, ok := e.Function.(*ast.Identifier); ok {
			fnName = id.Name
		}

		callInst := &Instruction{
			Type:   InstCall,
			Meta:   fnName,
			Ops:    args,
			Result: Value{
				ID:   b.nextValID,
				Type: "i64",
			},
		}
		b.currBlock.AddInstruction(callInst)
		b.nextValID++
		return callInst.Result, nil

	default:
		return Value{}, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// buildLiteral builds a literal expression
func (b *Builder) buildLiteral(lit *ast.Literal) (Value, error) {
	litType := ""
	switch lit.Value.(type) {
	case int64:
		litType = "i64"
	case float64:
		litType = "f64"
	case bool:
		litType = "bool"
	case string:
		litType = "string"
	default:
		litType = "unknown"
	}

	inst := &Instruction{
		Type: InstLiteral,
		Meta: lit.Value,
		Result: Value{
			ID:   b.nextValID,
			Type: litType,
		},
	}
	b.currBlock.AddInstruction(inst)
	b.nextValID++
	return inst.Result, nil
}

// typeFromOp converts operator token type to IR op type
func typeFromOp(op lexer.TokenType) string {
	switch op {
	case lexer.TokenPlus:
		return "add"
	case lexer.TokenMinus:
		return "sub"
	case lexer.TokenStar:
		return "mul"
	case lexer.TokenSlash:
		return "div"
	case lexer.TokenPercent:
		return "mod"
	case lexer.TokenEqualEqual:
		return "eq"
	case lexer.TokenNotEqual:
		return "ne"
	case lexer.TokenLess:
		return "lt"
	case lexer.TokenLessEqual:
		return "le"
	case lexer.TokenGreater:
		return "gt"
	case lexer.TokenGreaterEqual:
		return "ge"
	case lexer.TokenAnd:
		return "and"
	case lexer.TokenOr:
		return "or"
	case lexer.TokenAmpersand:
		return "bitand"
	case lexer.TokenPipe:
		return "bitor"
	case lexer.TokenCaret:
		return "bitxor"
	default:
		return "unknown"
	}
}
