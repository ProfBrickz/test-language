package interpreter

import (
	"fmt"
	"math"

	"lang-interpreter/ast"
)

type Value struct {
	IType   ast.IntegerType
	FType   ast.FloatType
	Data    int64
	FData   float64
	Untyped bool
	IsFloat bool
	Null    bool
}

func (v Value) String() string {
	if v.Null {
		return "null"
	}
	if v.IsFloat {
		if v.Untyped {
			return fmt.Sprintf("%g", v.FData)
		}
		return fmt.Sprintf("%s(%g)", typeDesc(ast.IntegerType{}, v.FType, true), v.FData)
	}
	if v.Untyped {
		return fmt.Sprintf("%d", v.Data)
	}
	return fmt.Sprintf("%s(%d)", typeDesc(v.IType, ast.FloatType{}, false), v.Data)
}

func typeDesc(it ast.IntegerType, ft ast.FloatType, isFloat bool) string {
	if isFloat {
		nullable := ""
		if ft.Nullable {
			nullable = "nullable "
		}
		return fmt.Sprintf("%s%d-bit float", nullable, ft.Size)
	}
	sign := "signed"
	if !it.Signed {
		sign = "unsigned"
	}
	nullable := ""
	if it.Nullable {
		nullable = "nullable "
	}
	return fmt.Sprintf("%s%d-bit %s integer", nullable, it.Size, sign)
}

func canImplicitConvert(srcInt ast.IntegerType, srcFloat ast.FloatType, srcIsFloat bool,
	dstInt ast.IntegerType, dstFloat ast.FloatType, dstIsFloat bool) bool {

	// Nullability check: nullable cannot convert to non-nullable
	srcNullable := srcIsFloat && srcFloat.Nullable || !srcIsFloat && srcInt.Nullable
	dstNullable := dstIsFloat && dstFloat.Nullable || !dstIsFloat && dstInt.Nullable
	if srcNullable && !dstNullable {
		return false
	}

	// Float to float
	if srcIsFloat && dstIsFloat {
		return srcFloat.Size <= dstFloat.Size
	}

	// Integer to float
	if !srcIsFloat && dstIsFloat {
		// int8, uint8 -> float16
		if srcInt.Size <= 8 {
			return dstFloat.Size >= 16
		}
		// int16, uint16 -> float32
		if srcInt.Size <= 16 {
			return dstFloat.Size >= 32
		}
		// int32, uint32 -> float64
		if srcInt.Size <= 32 {
			return dstFloat.Size >= 64
		}
		// int64, uint64 -> no implicit float conversion
		return false
	}

	// Float to integer - not allowed implicitly
	if srcIsFloat && !dstIsFloat {
		return false
	}

	// Integer to integer
	if !srcIsFloat && !dstIsFloat {
		if srcInt.Size == dstInt.Size && srcInt.Signed == dstInt.Signed {
			return true
		}
		if srcInt.Size > dstInt.Size {
			return false
		}
		if srcInt.Signed && dstInt.Signed {
			return true
		}
		if !srcInt.Signed && !dstInt.Signed {
			return true
		}
		if !srcInt.Signed && dstInt.Signed && srcInt.Size < dstInt.Size {
			return true
		}
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

func canFitInFloat(val float64, ftype ast.FloatType) bool {
	switch ftype.Size {
	case 16:
		return val >= -65504 && val <= 65504 // float16 max
	case 32:
		return val >= -math.MaxFloat32 && val <= math.MaxFloat32
	case 64:
		return true // float64 can hold anything Go float64 can
	}
	return false
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
		val := Value{IType: s.IType, FType: s.FType, IsFloat: s.IsFloat}
		if s.Expr != nil {
			rightVal, err := i.evalExpr(s.Expr)
			if err != nil {
				return err
			}
			if rightVal.Null {
				if !s.IType.Nullable && !s.FType.Nullable {
					return fmt.Errorf("cannot assign null to %s", typeDesc(s.IType, s.FType, s.IsFloat))
				}
				val.Null = true
			} else {
				if rightVal.IsFloat {
					if !s.IsFloat {
						return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDesc(rightVal.IType, rightVal.FType, true), typeDesc(s.IType, s.FType, false))
					}
					if !canImplicitConvert(rightVal.IType, rightVal.FType, true, s.IType, s.FType, true) {
						return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDesc(rightVal.IType, rightVal.FType, true), typeDesc(s.IType, s.FType, true))
					}
					if !canFitInFloat(rightVal.FData, s.FType) {
						return fmt.Errorf("overflow: value %g cannot fit in %s", rightVal.FData, typeDesc(ast.IntegerType{}, s.FType, true))
					}
					val.FData = rightVal.FData
				} else {
					if s.IsFloat {
						// Integer to float conversion
						if !canImplicitConvert(rightVal.IType, ast.FloatType{}, false, ast.IntegerType{}, s.FType, true) {
							return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDesc(rightVal.IType, ast.FloatType{}, false), typeDesc(ast.IntegerType{}, s.FType, true))
						}
						fresult := float64(rightVal.Data)
						if !canFitInFloat(fresult, s.FType) {
							return fmt.Errorf("overflow: value %g cannot fit in %s", fresult, typeDesc(ast.IntegerType{}, s.FType, true))
						}
						val.FData = fresult
					} else {
						// Integer to integer
						if rightVal.Untyped {
							if !canFitInType(rightVal.Data, s.IType) {
								return fmt.Errorf("overflow: value %d cannot fit in %s", rightVal.Data, typeDesc(s.IType, s.FType, false))
							}
						} else if !canImplicitConvert(rightVal.IType, rightVal.FType, false, s.IType, s.FType, false) {
							return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDesc(rightVal.IType, rightVal.FType, false), typeDesc(s.IType, s.FType, false))
						}
						val.Data = rightVal.Data
					}
				}
			}
		} else if (s.IsFloat && s.FType.Nullable) || (!s.IsFloat && s.IType.Nullable) {
			val.Null = true
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
		} else if val.IsFloat {
			fmt.Println(val.FData)
		} else {
			fmt.Println(val.Data)
		}
	default:
		return fmt.Errorf("unknown statement type")
	}
	return nil
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
		if !val.IType.Nullable && !val.FType.Nullable {
			return fmt.Errorf("cannot assign null to %s", typeDesc(val.IType, val.FType, val.IsFloat))
		}
		val.Null = true
		val.Data = 0
		val.FData = 0
		i.env.Set(stmt.Name, val)
		return nil
	}

	if val.Null && stmt.Op != "=" {
		return fmt.Errorf("cannot use null variable in %s operation", stmt.Op)
	}

	var result int64
	var fresult float64

	if val.IsFloat && !rightVal.IsFloat {
		// Integer to float assignment
		if !canImplicitConvert(rightVal.IType, rightVal.FType, false, val.IType, val.FType, true) {
			return fmt.Errorf("type mismatch: cannot assign %s to %s", typeDesc(rightVal.IType, rightVal.FType, false), typeDesc(val.IType, val.FType, true))
		}
		switch stmt.Op {
		case "=":
			fresult = float64(rightVal.Data)
		case "+=":
			fresult = val.FData + float64(rightVal.Data)
		case "-=":
			fresult = val.FData - float64(rightVal.Data)
		case "*=":
			fresult = val.FData * float64(rightVal.Data)
		case "/=":
			if rightVal.Data == 0 {
				return fmt.Errorf("division by zero")
			}
			fresult = val.FData / float64(rightVal.Data)
		default:
			return fmt.Errorf("unknown operator: %s", stmt.Op)
		}
		if !canFitInFloat(fresult, val.FType) {
			return fmt.Errorf("overflow: value %g cannot fit in %s", fresult, typeDesc(val.IType, val.FType, true))
		}
		val.FData = fresult
	} else if !val.IsFloat && rightVal.IsFloat {
		// Float to integer - not allowed implicitly
		return fmt.Errorf("type mismatch: cannot assign %s to %s", typeDesc(rightVal.IType, rightVal.FType, true), typeDesc(val.IType, val.FType, false))
	} else if val.IsFloat {
		// Float to float
		fresult = val.FData
		switch stmt.Op {
		case "=":
			fresult = rightVal.FData
		case "+=":
			fresult = fresult + rightVal.FData
		case "-=":
			fresult = fresult - rightVal.FData
		case "*=":
			fresult = fresult * rightVal.FData
		case "/=":
			if rightVal.FData == 0 {
				return fmt.Errorf("division by zero")
			}
			fresult = fresult / rightVal.FData
		default:
			return fmt.Errorf("unknown operator: %s", stmt.Op)
		}
		if !canFitInFloat(fresult, val.FType) {
			return fmt.Errorf("overflow: value %g cannot fit in %s", fresult, typeDesc(val.IType, val.FType, true))
		}
		val.FData = fresult
	} else {
		// Integer to integer
		result = val.Data
		switch stmt.Op {
		case "=":
			result = rightVal.Data
		case "+=":
			result = result + rightVal.Data
		case "-=":
			result = result - rightVal.Data
		case "*=":
			result = result * rightVal.Data
		case "/=":
			if rightVal.Data == 0 {
				return fmt.Errorf("division by zero")
			}
			result = result / rightVal.Data
		default:
			return fmt.Errorf("unknown operator: %s", stmt.Op)
		}
		if !canFitInType(result, val.IType) {
			return fmt.Errorf("overflow: value %d cannot fit in %s", result, typeDesc(val.IType, val.FType, false))
		}
		val.Data = result
	}
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
		return Value{IType: e.IType, Data: e.Value, IsFloat: false}, nil
	case *ast.FloatLit:
		if e.Untyped {
			return Value{Untyped: true, FData: e.Value, IsFloat: true}, nil
		}
		return Value{FType: e.FType, FData: e.Value, IsFloat: true}, nil
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

	// Determine result type
	isFloat := left.IsFloat || right.IsFloat
	var resultType ast.IntegerType
	var resultFType ast.FloatType

	if isFloat {
		// Float result
		if left.IsFloat && right.IsFloat {
			resultFType.Size = left.FType.Size
			if right.FType.Size > left.FType.Size {
				resultFType.Size = right.FType.Size
			}
		} else if left.IsFloat {
			resultFType = left.FType
		} else {
			resultFType = right.FType
		}
	} else {
		// Integer result
		if left.Untyped && right.Untyped {
			resultType = ast.IntegerType{Size: 64, Signed: true}
		} else if left.Untyped {
			resultType = right.IType
		} else if right.Untyped {
			resultType = left.IType
		} else {
			resultType = left.IType
			if !canImplicitConvert(right.IType, right.FType, false, left.IType, left.FType, false) && !canImplicitConvert(left.IType, left.FType, false, right.IType, right.FType, false) {
				return Value{}, fmt.Errorf("type mismatch: cannot compute %s %s %s", typeDesc(left.IType, left.FType, false), expr.Op, typeDesc(right.IType, right.FType, false))
			}
			if canImplicitConvert(right.IType, right.FType, false, left.IType, left.FType, false) && left.IType.Size >= right.IType.Size {
				right.Data = clamp(right.Data, left.IType)
				right.IType = left.IType
				right.Untyped = false
			} else if canImplicitConvert(left.IType, left.FType, false, right.IType, right.FType, false) {
				left.Data = clamp(left.Data, right.IType)
				left.IType = right.IType
				left.Untyped = false
				resultType = right.IType
			}
		}
	}

	var result int64
	var fresult float64

	// Get float values for both operands (convert int to float if needed)
	getFloat := func(v Value) float64 {
		if v.IsFloat {
			return v.FData
		}
		return float64(v.Data)
	}
	leftF := getFloat(left)
	rightF := getFloat(right)

	switch expr.Op {
	case "+":
		if isFloat {
			fresult = leftF + rightF
		} else {
			result = left.Data + right.Data
		}
	case "-":
		if isFloat {
			fresult = leftF - rightF
		} else {
			result = left.Data - right.Data
		}
	case "*":
		if isFloat {
			fresult = leftF * rightF
		} else {
			result = left.Data * right.Data
		}
	case "/":
		if isFloat {
			if rightF == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			fresult = leftF / rightF
		} else {
			if right.Data == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			result = left.Data / right.Data
		}
	default:
		return Value{}, fmt.Errorf("unknown operator: %s", expr.Op)
	}

	if isFloat {
		if left.Untyped || right.Untyped {
			return Value{Untyped: true, FData: fresult, IsFloat: true}, nil
		}
		if !canFitInFloat(fresult, resultFType) {
			return Value{}, fmt.Errorf("overflow: result %g cannot fit in %s", fresult, typeDesc(ast.IntegerType{}, resultFType, true))
		}
		return Value{FType: resultFType, FData: fresult, IsFloat: true}, nil
	}

	if left.Untyped || right.Untyped {
		return Value{Untyped: true, Data: result}, nil
	}
	if !canFitInType(result, resultType) {
		return Value{}, fmt.Errorf("overflow: result %d cannot fit in %s", result, typeDesc(resultType, ast.FloatType{}, false))
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
