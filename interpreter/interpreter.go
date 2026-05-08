package interpreter

import (
	"fmt"
	"strconv"

	"github.com/x448/float16"

	"lang-interpreter/ast"
)

type Value struct {
	IType   ast.IntegerType
	FType   ast.FloatType
	BType   ast.BoolType
	Data    int64
	FData   float64
	BData   bool
	Untyped bool
	IsFloat bool
	IsBool  bool
	Null    bool
}

func (v Value) String() string {
	if v.Null {
		return "null"
	}
	if v.IsBool {
		if v.Untyped {
			return fmt.Sprintf("%t", v.BData)
		}
		return fmt.Sprintf("%s(%t)", typeDesc(v.BType, true), v.BData)
	}
	if v.IsFloat {
		if v.Untyped {
			return fmt.Sprintf("%g", v.FData)
		}
		return fmt.Sprintf("%s(%g)", typeDesc(v.FType, false), v.FData)
	}
	if v.Untyped {
		return fmt.Sprintf("%d", v.Data)
	}
	return fmt.Sprintf("%s(%d)", typeDesc(v.IType, false), v.Data)
}

func typeDescFromVar(it ast.IntegerType, ft ast.FloatType, bt ast.BoolType, isFloat, isBool bool) string {
	if isBool {
		nullable := ""
		if bt.Nullable {
			nullable = "nullable "
		}
		return fmt.Sprintf("%sbool", nullable)
	}
	if isFloat {
		nullable := ""
		if ft.Nullable {
			nullable = "nullable "
		}
		if ft.Size == 0 {
			return "untyped float literal"
		}
		return fmt.Sprintf("%s%d-bit float", nullable, ft.Size)
	}
	nullable := ""
	if it.Nullable {
		nullable = "nullable "
	}
	if it.Size == 0 {
		return "untyped integer literal"
	}
	sign := "signed"
	if !it.Signed {
		sign = "unsigned"
	}
	return fmt.Sprintf("%s%d-bit %s integer", nullable, it.Size, sign)
}

func typeDescForVal(v Value) string {
	return typeDescFromVar(v.IType, v.FType, v.BType, v.IsFloat, v.IsBool)
}

func typeDesc(t interface{}, isBool bool) string {
	switch typ := t.(type) {
	case ast.BoolType:
		if typ.Nullable {
			return "nullable bool"
		}
		return "bool"
	case ast.FloatType:
		if typ.Size == 0 {
			return "untyped float literal"
		}
		nullable := ""
		if typ.Nullable {
			nullable = "nullable "
		}
		return fmt.Sprintf("%s%d-bit float", nullable, typ.Size)
	case ast.IntegerType:
		if typ.Size == 0 {
			return "untyped integer literal"
		}
		sign := "signed"
		if !typ.Signed {
			sign = "unsigned"
		}
		nullable := ""
		if typ.Nullable {
			nullable = "nullable "
		}
		return fmt.Sprintf("%s%d-bit %s integer", nullable, typ.Size, sign)
	}
	return "unknown"
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

func convertInt(val int64, itype ast.IntegerType) int64 {
	if itype.Signed {
		switch itype.Size {
		case 8:
			return int64(int8(val))
		case 16:
			return int64(int16(val))
		case 32:
			return int64(int32(val))
		case 64:
			return val
		}
	} else {
		switch itype.Size {
		case 8:
			return int64(uint8(val))
		case 16:
			return int64(uint16(val))
		case 32:
			return int64(uint32(val))
		case 64:
			return int64(uint64(val))
		}
	}
	return val
}

func convertFloat(val float64, ftype ast.FloatType) float64 {
	switch ftype.Size {
	case 16:
		return float64(float16.Fromfloat32(float32(val)).Float32())
	case 32:
		return float64(float32(val))
	case 64:
		return val
	}
	return val
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
		val := Value{IType: s.IType, FType: s.FType, BType: s.BType, IsFloat: s.IsFloat, IsBool: s.IsBool}
		if s.Expr != nil {
			rightVal, err := i.evalExpr(s.Expr)
			if err != nil {
				return err
			}
			if rightVal.Null {
				if !s.IsBool && !s.IType.Nullable && !s.FType.Nullable || (s.IsBool && !s.BType.Nullable) {
					return fmt.Errorf("cannot assign null to %s", typeDescFromVar(s.IType, s.FType, s.BType, s.IsFloat, s.IsBool))
				}
				val.Null = true
			} else {
				if rightVal.IsBool {
					if !s.IsBool {
						return fmt.Errorf("cannot assign %s to %s", typeDescForVal(rightVal), typeDescFromVar(s.IType, s.FType, s.BType, s.IsFloat, s.IsBool))
					}
					if rightVal.BType.Nullable && !s.BType.Nullable {
						return fmt.Errorf("cannot assign nullable %s to non-nullable %s", typeDescForVal(rightVal), typeDescFromVar(s.IType, s.FType, s.BType, s.IsFloat, s.IsBool))
					}
					val.BData = rightVal.BData
				} else if rightVal.IsFloat {
					if !s.IsFloat {
						return fmt.Errorf("cannot assign %s to integer variable", typeDescForVal(rightVal))
					}
					if !rightVal.Untyped && !canImplicitConvert(rightVal.IType, rightVal.FType, true, s.IType, s.FType, true) {
						return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDescForVal(rightVal), typeDescFromVar(s.IType, s.FType, s.BType, true, false))
					}
					val.FData = convertFloat(rightVal.FData, s.FType)
				} else {
					if s.IsBool {
						val.BData = rightVal.Data != 0
					} else if s.IsFloat {
						// Integer to float conversion
						if !rightVal.Untyped && !canImplicitConvert(rightVal.IType, ast.FloatType{}, false, ast.IntegerType{}, s.FType, true) {
							return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDescForVal(rightVal), typeDescFromVar(ast.IntegerType{}, s.FType, ast.BoolType{}, true, false))
						}
						val.FData = convertFloat(float64(rightVal.Data), s.FType)
					} else {
						// Integer to integer
						if !rightVal.Untyped && !canImplicitConvert(rightVal.IType, rightVal.FType, false, s.IType, s.FType, false) {
							return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDescForVal(rightVal), typeDescFromVar(s.IType, s.FType, ast.BoolType{}, false, false))
						}
						val.Data = convertInt(rightVal.Data, s.IType)
					}
				}
			}
		} else if s.IsBool {
			if s.BType.Nullable {
				val.Null = true
			} else {
				val.BData = false
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
		} else if val.IsBool {
			fmt.Println(val.BData)
		} else if val.IsFloat {
			fmt.Println(strconv.FormatFloat(val.FData, 'g', 20, 64))
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
		if !val.IsBool && !val.IType.Nullable && !val.FType.Nullable || (val.IsBool && !val.BType.Nullable) {
			return fmt.Errorf("cannot assign null to %s", typeDescForVal(val))
		}
		val.Null = true
		val.Data = 0
		val.FData = 0
		val.BData = false
		i.env.Set(stmt.Name, val)
		return nil
	}

	if val.Null && stmt.Op != "=" {
		return fmt.Errorf("cannot use null variable in %s operation", stmt.Op)
	}

	var result int64
	var fresult float64

	if val.IsBool {
		if !rightVal.IsBool {
			return fmt.Errorf("cannot assign %s to %s", typeDescForVal(rightVal), typeDescForVal(val))
		}
		if rightVal.BType.Nullable && !val.BType.Nullable {
			return fmt.Errorf("cannot assign nullable %s to non-nullable %s", typeDescForVal(rightVal), typeDescForVal(val))
		}
		switch stmt.Op {
		case "=":
			val.BData = rightVal.BData
		default:
			return fmt.Errorf("unknown operator: %s", stmt.Op)
		}
	} else if val.IsFloat && !rightVal.IsFloat {
		// Integer to float assignment
		if !rightVal.Untyped && !canImplicitConvert(rightVal.IType, rightVal.FType, false, val.IType, val.FType, true) {
			return fmt.Errorf("type mismatch: cannot assign %s to %s", typeDescForVal(rightVal), typeDescFromVar(val.IType, val.FType, ast.BoolType{}, true, false))
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
			fresult = val.FData / float64(rightVal.Data)
		default:
			return fmt.Errorf("unknown operator: %s", stmt.Op)
		}
		val.FData = convertFloat(fresult, val.FType)
	} else if !val.IsFloat && rightVal.IsFloat {
		// Float to integer - not allowed implicitly
		return fmt.Errorf("cannot assign %s to integer variable", typeDescForVal(rightVal))
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
			fresult = fresult / rightVal.FData
		default:
			return fmt.Errorf("unknown operator: %s", stmt.Op)
		}
		val.FData = convertFloat(fresult, val.FType)
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
		val.Data = convertInt(result, val.IType)
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
	case *ast.BoolLit:
		if e.Untyped {
			return Value{Untyped: true, BData: e.Value, IsBool: true}, nil
		}
		return Value{BType: e.BType, BData: e.Value, IsBool: true}, nil
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
		if left.Untyped && right.Untyped {
			resultType = ast.IntegerType{Size: 64, Signed: true}
		} else if left.Untyped {
			resultType = right.IType
		} else if right.Untyped {
			resultType = left.IType
		} else {
			resultType = left.IType
			if !canImplicitConvert(right.IType, right.FType, false, left.IType, left.FType, false) && !canImplicitConvert(left.IType, left.FType, false, right.IType, right.FType, false) {
				return Value{}, fmt.Errorf("type mismatch: cannot compute %s %s %s", typeDescForVal(left), expr.Op, typeDescForVal(right))
			}
			if canImplicitConvert(right.IType, right.FType, false, left.IType, left.FType, false) && left.IType.Size >= right.IType.Size {
				right.Data = convertInt(right.Data, left.IType)
				right.IType = left.IType
				right.Untyped = false
			} else if canImplicitConvert(left.IType, left.FType, false, right.IType, right.FType, false) {
				left.Data = convertInt(left.Data, right.IType)
				left.IType = right.IType
				left.Untyped = false
				resultType = right.IType
			}
		}
	}

	var result int64
	var fresult float64

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
		return Value{FType: resultFType, FData: convertFloat(fresult, resultFType), IsFloat: true}, nil
	}

	if left.Untyped || right.Untyped {
		return Value{Untyped: true, Data: result}, nil
	}
	return Value{IType: resultType, Data: convertInt(result, resultType)}, nil
}
