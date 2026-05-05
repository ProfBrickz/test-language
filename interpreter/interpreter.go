package interpreter

import (
	"fmt"
	"math"

	"lang-interpreter/ast"
)

type Value struct {
	IType   ast.IntegerType
	Data    int64
	Untyped bool
	Null    bool
}

func (v Value) String() string {
	if v.Null {
		return "null"
	}
	if v.Untyped {
		return fmt.Sprintf("%d", v.Data)
	}
	return fmt.Sprintf("integer{size: %d, signed: %t}(%d)", v.IType.Size, v.IType.Signed, v.Data)
}

func typeDesc(it ast.IntegerType) string {
	sign := "signed"
	if !it.Signed {
		sign = "unsigned"
	}
	nullable := ""
	if it.Nullable {
		nullable = " nullable"
	}
	return fmt.Sprintf("%d bit%s %s integer", it.Size, nullable, sign)
}

func canImplicitConvert(src ast.IntegerType, dst ast.IntegerType) bool {
	if src.Size == dst.Size && src.Signed == dst.Signed {
		return true
	}
	if src.Size > dst.Size {
		return false
	}
	if src.Signed && dst.Signed {
		return true
	}
	if !src.Signed && !dst.Signed {
		return true
	}
	if !src.Signed && dst.Signed && src.Size < dst.Size {
		return true
	}
	return false
}

func canFitInType(val int64, itype ast.IntegerType) bool {
	var min, max int64
	switch itype.Size {
	case 8:
		if itype.Signed {
			min = math.MinInt8
			max = math.MaxInt8
		} else {
			min = 0
			max = math.MaxUint8
		}
	case 16:
		if itype.Signed {
			min = math.MinInt16
			max = math.MaxInt16
		} else {
			min = 0
			max = math.MaxUint16
		}
	case 32:
		if itype.Signed {
			min = math.MinInt32
			max = math.MaxInt32
		} else {
			min = 0
			max = math.MaxUint32
		}
	case 64:
		if itype.Signed {
			min = math.MinInt64
			max = math.MaxInt64
		} else {
			min = 0
			max = math.MaxInt64
		}
	default:
		min = math.MinInt64
		max = math.MaxInt64
	}
	return val >= min && val <= max
}

type Environment struct {
	variables map[string]Value
}

func NewEnv() *Environment {
	return &Environment{variables: make(map[string]Value)}
}

func (e *Environment) Set(name string, val Value) {
	e.variables[name] = val
}

func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.variables[name]
	return val, ok
}

type Interpreter struct {
	env *Environment
}

func New() *Interpreter {
	return &Interpreter{env: NewEnv()}
}

func (i *Interpreter) Run(program *ast.Program) error {
	for _, stmt := range program.Stmts {
		if err := i.executeStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) ExecuteStmt(stmt ast.Stmt) error {
	return i.executeStmt(stmt)
}

func (i *Interpreter) executeStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		val := Value{IType: s.IType}
		if s.IType.Nullable {
			val.Null = true
		}
		if s.Expr != nil {
			rightVal, err := i.evalExpr(s.Expr)
			if err != nil {
				return err
			}
			if rightVal.Null {
				if !s.IType.Nullable {
					return fmt.Errorf("cannot assign null to %s", typeDesc(s.IType))
				}
				val.Null = true
			} else {
				if isVarRef(s.Expr) {
					if !canImplicitConvert(rightVal.IType, s.IType) {
						return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDesc(rightVal.IType), typeDesc(s.IType))
					}
				}
				if rightVal.Untyped {
					if !canFitInType(rightVal.Data, s.IType) {
						return fmt.Errorf("overflow: value %d cannot fit in %s", rightVal.Data, typeDesc(s.IType))
					}
				} else if !canImplicitConvert(rightVal.IType, s.IType) {
					return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDesc(rightVal.IType), typeDesc(s.IType))
				}
				val.Data = rightVal.Data
			}
		}
		i.env.Set(s.Name, val)
	case *ast.Assignment:
		return i.executeAssignment(s)
	case *ast.PrintStmt:
		val, err := i.evalExpr(s.Expr)
		if err != nil {
			return err
		}
		if val.Null {
			fmt.Println("null")
		} else {
			fmt.Println(val.Data)
		}
	default:
		return fmt.Errorf("unknown statement type")
	}
	return nil
}

func isLiteralExpr(expr ast.Expr) bool {
	_, ok := expr.(*ast.IntegerLit)
	return ok
}

func isVarRef(expr ast.Expr) bool {
	_, ok := expr.(*ast.VarRef)
	return ok
}

func (i *Interpreter) executeAssignment(stmt *ast.Assignment) error {
	val, ok := i.env.Get(stmt.Name)
	if !ok {
		return fmt.Errorf("undefined variable: %s", stmt.Name)
	}

	rightVal, err := i.evalExpr(stmt.Expr)
	if err != nil {
		return err
	}

	if rightVal.Null {
		if !val.IType.Nullable {
			return fmt.Errorf("cannot assign null to %s", typeDesc(val.IType))
		}
		val.Null = true
		val.Data = 0
		i.env.Set(stmt.Name, val)
		return nil
	}

	if val.Null && stmt.Op != "=" {
		return fmt.Errorf("cannot use null variable in %s operation", stmt.Op)
	}

	if isLiteralExpr(stmt.Expr) {
		// literals are handled by canFitInType check below
	} else if rightVal.Untyped {
		// untyped expression result - just check it fits
	} else {
		if !canImplicitConvert(rightVal.IType, val.IType) {
			return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDesc(rightVal.IType), typeDesc(val.IType))
		}
	}

	var result int64
	switch stmt.Op {
	case "=":
		result = rightVal.Data
	case "+=":
		result = val.Data + rightVal.Data
	case "-=":
		result = val.Data - rightVal.Data
	case "*=":
		result = val.Data * rightVal.Data
	case "/=":
		if rightVal.Data == 0 {
			return fmt.Errorf("division by zero")
		}
		result = val.Data / rightVal.Data
	default:
		return fmt.Errorf("unknown operator: %s", stmt.Op)
	}

	if !canFitInType(result, val.IType) {
		return fmt.Errorf("overflow: value %d cannot fit in %s", result, typeDesc(val.IType))
	}

	val.Data = result
	val.Null = false
	i.env.Set(stmt.Name, val)
	return nil
}

func (i *Interpreter) evalExpr(expr ast.Expr) (Value, error) {
	switch e := expr.(type) {
	case *ast.IntegerLit:
		if e.Untyped {
			return Value{Untyped: true, Data: e.Value}, nil
		}
		return Value{IType: e.IType, Data: e.Value}, nil
	case *ast.VarRef:
		val, ok := i.env.Get(e.Name)
		if !ok {
			return Value{}, fmt.Errorf("undefined variable: %s", e.Name)
		}
		return val, nil
	case *ast.BinaryExpr:
		return i.evalBinary(e)
	case *ast.NullLit:
		return Value{Null: true}, nil
	default:
		return Value{}, fmt.Errorf("unknown expression type")
	}
}

func (i *Interpreter) evalBinary(expr *ast.BinaryExpr) (Value, error) {
	left, err := i.evalExpr(expr.Left)
	if err != nil {
		return Value{}, err
	}
	right, err := i.evalExpr(expr.Right)
	if err != nil {
		return Value{}, err
	}

	if left.Null || right.Null {
		return Value{Null: true}, nil
	}

	var resultType ast.IntegerType
	if left.Untyped && right.Untyped {
		resultType = ast.IntegerType{Size: 64, Signed: true}
	} else if left.Untyped {
		resultType = right.IType
	} else if right.Untyped {
		resultType = left.IType
	} else {
		resultType = left.IType
		if !canImplicitConvert(right.IType, left.IType) && !canImplicitConvert(left.IType, right.IType) {
			return Value{}, fmt.Errorf("type mismatch: cannot compute %s %s %s", typeDesc(left.IType), expr.Op, typeDesc(right.IType))
		}
		if canImplicitConvert(right.IType, left.IType) && left.IType.Size >= right.IType.Size {
			right.Data = clamp(right.Data, left.IType)
			right.IType = left.IType
			right.Untyped = false
		} else if canImplicitConvert(left.IType, right.IType) {
			left.Data = clamp(left.Data, right.IType)
			left.IType = right.IType
			left.Untyped = false
			resultType = right.IType
		}
	}

	var result int64
	switch expr.Op {
	case "+":
		result = left.Data + right.Data
	case "-":
		result = left.Data - right.Data
	case "*":
		result = left.Data * right.Data
	case "/":
		if right.Data == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		result = left.Data / right.Data
	default:
		return Value{}, fmt.Errorf("unknown operator: %s", expr.Op)
	}

	if left.Untyped || right.Untyped {
		return Value{Untyped: true, IType: resultType, Data: result}, nil
	}

	if !canFitInType(result, resultType) {
		return Value{}, fmt.Errorf("overflow: result %d cannot fit in %s", result, typeDesc(resultType))
	}

	return Value{IType: resultType, Data: result}, nil
}

func clamp(val int64, itype ast.IntegerType) int64 {
	var min, max int64
	switch itype.Size {
	case 8:
		if itype.Signed {
			min = math.MinInt8
			max = math.MaxInt8
		} else {
			min = 0
			max = math.MaxUint8
		}
	case 16:
		if itype.Signed {
			min = math.MinInt16
			max = math.MaxInt16
		} else {
			min = 0
			max = math.MaxUint16
		}
	case 32:
		if itype.Signed {
			min = math.MinInt32
			max = math.MaxInt32
		} else {
			min = 0
			max = math.MaxUint32
		}
	case 64:
		if itype.Signed {
			min = math.MinInt64
			max = math.MaxInt64
		} else {
			min = 0
			max = math.MaxInt64
		}
	default:
		min = math.MinInt64
		max = math.MaxInt64
	}

	if val < min {
		val = min
	}
	if val > max {
		val = max
	}
	return val
}
