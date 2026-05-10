package interpreter

import (
	"fmt"
	"math"
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
	IsType  bool
}

func intToStr(data int64, itype ast.IntegerType, untyped bool) string {
	if !untyped && !itype.Signed && itype.Size == 64 {
		return fmt.Sprintf("%d", uint64(data))
	}
	return fmt.Sprintf("%d", data)
}

func formatFloat(val float64) string {
	if math.IsInf(val, 1) {
		return "infinity"
	}
	if math.IsInf(val, -1) {
		return "-infinity"
	}
	if math.IsNaN(val) {
		return "NaN"
	}
	return strconv.FormatFloat(val, 'g', -1, 64)
}

func intTypeMin(it ast.IntegerType) int64 {
	if !it.Signed {
		return 0
	}
	switch it.Size {
	case 8:
		return -128
	case 16:
		return -32768
	case 32:
		return -2147483648
	default:
		return math.MinInt64
	}
}

func intTypeMax(it ast.IntegerType) int64 {
	if !it.Signed {
		switch it.Size {
		case 8:
			return 255
		case 16:
			return 65535
		case 32:
			return 4294967295
		default:
			return -1 // int64(-1) has same bit pattern as math.MaxUint64
		}
	}
	switch it.Size {
	case 8:
		return 127
	case 16:
		return 32767
	case 32:
		return 2147483647
	default:
		return math.MaxInt64
	}
}

func floatTypeMax(size int) float64 {
	switch size {
	case 16:
		return 65504
	case 32:
		return math.MaxFloat32
	default:
		return math.MaxFloat64
	}
}

func floatTypeMinSubnormal(size int) float64 {
	switch size {
	case 16:
		return math.Exp2(-24)
	case 32:
		return math.SmallestNonzeroFloat32
	default:
		return math.SmallestNonzeroFloat64
	}
}

func floatTypeMinNormal(size int) float64 {
	switch size {
	case 16:
		return math.Exp2(-14)
	case 32:
		return math.Exp2(-126)
	default:
		return math.Exp2(-1022)
	}
}

func floatTypePrecision(size int) int {
	switch size {
	case 16:
		return 3
	case 32:
		return 7
	default:
		return 15
	}
}

func floatTypeMinExponent(size int) int {
	switch size {
	case 16:
		return -14
	case 32:
		return -126
	default:
		return -1022
	}
}

func floatTypeMaxExponent(size int) int {
	switch size {
	case 16:
		return 15
	case 32:
		return 127
	default:
		return 1023
	}
}

func valueToTypeDesc(val Value) Value {
	td := Value{IsType: true}
	if val.IsFloat {
		td.IsFloat = true
		td.FType = val.FType
		td.FType.Nullable = false
	} else if val.IsBool {
		td.IsBool = true
		td.BType = val.BType
		td.BType.Nullable = false
	} else {
		td.IType = val.IType
		td.IType.Nullable = false
	}
	return td
}

func (i *Interpreter) evalTypeMember(td Value, member string) (Value, error) {
	if td.IsFloat {
		size := td.FType.Size
		switch member {
		case "min":
			return Value{FType: td.FType, FData: -floatTypeMax(size), IsFloat: true}, nil
		case "max":
			return Value{FType: td.FType, FData: floatTypeMax(size), IsFloat: true}, nil
		case "min_subnormal":
			return Value{FType: td.FType, FData: floatTypeMinSubnormal(size), IsFloat: true}, nil
		case "min_normal":
			return Value{FType: td.FType, FData: floatTypeMinNormal(size), IsFloat: true}, nil
		case "precision":
			return Value{IType: td.IType, Data: int64(floatTypePrecision(size))}, nil
		case "min_exponent":
			return Value{IType: td.IType, Data: int64(floatTypeMinExponent(size))}, nil
		case "max_exponent":
			return Value{IType: td.IType, Data: int64(floatTypeMaxExponent(size))}, nil
		case "size":
			return Value{IType: td.IType, Data: int64(size)}, nil
		}
	} else if td.IsBool {
		switch member {
		case "size":
			return Value{IType: td.IType, Data: 8}, nil
		}
	} else {
		switch member {
		case "min":
			return Value{IType: td.IType, Data: intTypeMin(td.IType)}, nil
		case "max":
			return Value{IType: td.IType, Data: intTypeMax(td.IType)}, nil
		case "size":
			return Value{IType: td.IType, Data: int64(td.IType.Size)}, nil
		}
	}
	return Value{}, fmt.Errorf("type %s has no member %q", typeDescForVal(td), member)
}

func (v Value) String() string {
	if v.IsType {
		return typeDescForVal(v)
	}
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
			return formatFloat(v.FData)
		}
		return fmt.Sprintf("%s(%s)", typeDesc(v.FType, false), formatFloat(v.FData))
	}
	if v.Untyped {
		return intToStr(v.Data, v.IType, true)
	}
	return fmt.Sprintf("%s(%s)", typeDesc(v.IType, false), intToStr(v.Data, v.IType, false))
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
		return "untyped int literal"
	}
	sign := "signed"
	if !it.Signed {
		sign = "unsigned"
	}
	return fmt.Sprintf("%s%d-bit %s int", nullable, it.Size, sign)
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
			return "untyped int literal"
		}
		sign := "signed"
		if !typ.Signed {
			sign = "unsigned"
		}
		nullable := ""
		if typ.Nullable {
			nullable = "nullable "
		}
		return fmt.Sprintf("%s%d-bit %s int", nullable, typ.Size, sign)
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

func checkIntFits(val int64, itype ast.IntegerType) error {
	if itype.Signed {
		switch itype.Size {
		case 8:
			if val < math.MinInt8 || val > math.MaxInt8 {
				return fmt.Errorf("value %d does not fit in 8-bit signed int", val)
			}
		case 16:
			if val < math.MinInt16 || val > math.MaxInt16 {
				return fmt.Errorf("value %d does not fit in 16-bit signed int", val)
			}
		case 32:
			if val < math.MinInt32 || val > math.MaxInt32 {
				return fmt.Errorf("value %d does not fit in 32-bit signed int", val)
			}
		}
	} else {
		switch itype.Size {
		case 8:
			if val < 0 || val > math.MaxUint8 {
				return fmt.Errorf("value %d does not fit in 8-bit unsigned int", val)
			}
		case 16:
			if val < 0 || val > math.MaxUint16 {
				return fmt.Errorf("value %d does not fit in 16-bit unsigned int", val)
			}
		case 32:
			if val < 0 || val > math.MaxUint32 {
				return fmt.Errorf("value %d does not fit in 32-bit unsigned int", val)
			}
		case 64:
			if val < 0 {
				return fmt.Errorf("value %d does not fit in 64-bit unsigned int", val)
			}
		}
	}
	return nil
}

func checkFloatFits(val float64, ftype ast.FloatType) error {
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return nil
	}
	switch ftype.Size {
	case 16:
		if val > 65504 || val < -65504 {
			return fmt.Errorf("value %g does not fit in 16-bit float", val)
		}
	case 32:
		if val > math.MaxFloat32 || val < -math.MaxFloat32 {
			return fmt.Errorf("value %g does not fit in 32-bit float", val)
		}
	}
	return nil
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
	outer     *Environment
}

func NewEnv() *Environment {
	return &Environment{variables: make(map[string]Value)}
}

func NewEnclosedEnv(outer *Environment) *Environment {
	return &Environment{variables: make(map[string]Value), outer: outer}
}

func (e *Environment) Set(name string, val Value) {
	e.variables[name] = val
}

func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.variables[name]
	if ok {
		return val, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return Value{}, false
}

func (e *Environment) LookupEnv(name string) *Environment {
	_, ok := e.variables[name]
	if ok {
		return e
	}
	if e.outer != nil {
		return e.outer.LookupEnv(name)
	}
	return nil
}

type Interpreter struct {
	env *Environment
}

func New() *Interpreter {
	return &Interpreter{env: NewEnv()}
}

func (i *Interpreter) pushScope() {
	i.env = NewEnclosedEnv(i.env)
}

func (i *Interpreter) popScope() {
	if i.env.outer != nil {
		i.env = i.env.outer
	}
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
			if lit, ok := s.Expr.(*ast.IntegerLit); ok && lit.Untyped && !s.IsFloat && !s.IsBool {
				if err := checkIntFits(lit.Value, s.IType); err != nil {
					return err
				}
			}
			if lit, ok := s.Expr.(*ast.FloatLit); ok && lit.Untyped && s.IsFloat {
				if err := checkFloatFits(lit.Value, s.FType); err != nil {
					return err
				}
			}
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
						return fmt.Errorf("cannot assign %s to int variable", typeDescForVal(rightVal))
					}
					if !rightVal.Untyped && !canImplicitConvert(rightVal.IType, rightVal.FType, true, s.IType, s.FType, true) {
						return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDescForVal(rightVal), typeDescFromVar(s.IType, s.FType, s.BType, true, false))
					}
					val.FData = convertFloat(rightVal.FData, s.FType)
				} else {
					if s.IsBool {
						return fmt.Errorf("cannot assign %s to %s", typeDescForVal(rightVal), typeDescFromVar(s.IType, s.FType, s.BType, s.IsFloat, s.IsBool))
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
	case *ast.BlockStmt:
		for _, inner := range s.Stmts {
			if err := i.executeStmt(inner); err != nil {
				return err
			}
		}
	case *ast.IfStmt:
		return i.execIf(s)
	case *ast.PrintStmt:
		val, err := i.evalExpr(s.Expr)
		if err != nil {
			return err
		}
		if val.IsType {
			fmt.Println(typeDescForVal(val))
		} else if val.Null {
			fmt.Println("null")
		} else if val.IsBool {
			fmt.Println(val.BData)
		} else if val.IsFloat {
			fmt.Println(formatFloat(val.FData))
		} else {
			fmt.Println(intToStr(val.Data, val.IType, val.Untyped))
		}
	default:
		return fmt.Errorf("unknown statement type")
	}
	return nil
}

func (i *Interpreter) execIf(s *ast.IfStmt) error {
	condVal, err := i.evalExpr(s.Condition)
	if err != nil {
		return err
	}
	if condVal.Null {
		if s.Else != nil {
			return i.execElse(s.Else)
		}
		return nil
	}
	if !condVal.IsBool {
		return fmt.Errorf("if condition must be a bool, got %s", typeDescForVal(condVal))
	}
	if condVal.BData {
		i.pushScope()
		err = i.executeStmt(s.Then)
		i.popScope()
		return err
	}
	if s.Else != nil {
		return i.execElse(s.Else)
	}
	return nil
}

func (i *Interpreter) execElse(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		i.pushScope()
		err := i.executeStmt(s)
		i.popScope()
		return err
	case *ast.IfStmt:
		return i.execIf(s)
	default:
		return fmt.Errorf("unexpected else type: %T", stmt)
	}
}

func (i *Interpreter) executeAssignment(stmt *ast.Assignment) error {
	env := i.env.LookupEnv(stmt.Name)
	if env == nil {
		return fmt.Errorf("undefined variable: %s", stmt.Name)
	}
	val, _ := env.Get(stmt.Name)

	if stmt.Op == "=" {
		if lit, ok := stmt.Expr.(*ast.IntegerLit); ok && lit.Untyped && !val.IsFloat && !val.IsBool {
			if err := checkIntFits(lit.Value, val.IType); err != nil {
				return err
			}
		}
		if lit, ok := stmt.Expr.(*ast.FloatLit); ok && lit.Untyped && val.IsFloat {
			if err := checkFloatFits(lit.Value, val.FType); err != nil {
				return err
			}
		}
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
		env.Set(stmt.Name, val)
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
		return fmt.Errorf("cannot assign %s to int variable", typeDescForVal(rightVal))
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
	env.Set(stmt.Name, val)
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
	case *ast.UnaryExpr:
		return i.evalUnary(e)
	case *ast.NullLit:
		return Value{Null: true}, nil
	case *ast.TypeRef:
		val := Value{IsType: true}
		switch e.Kind {
		case "float":
			val.IsFloat = true
			val.FType = e.FType
			val.FType.Nullable = false
		case "bool":
			val.IsBool = true
			val.BType = e.BType
			val.BType.Nullable = false
		default:
			val.IType = e.IType
			val.IType.Nullable = false
		}
		return val, nil
	case *ast.MemberAccess:
		obj, err := i.evalExpr(e.Object)
		if err != nil {
			return Value{}, err
		}
		if e.Member == "type" {
			return valueToTypeDesc(obj), nil
		}
		if obj.IsType {
			return i.evalTypeMember(obj, e.Member)
		}
		return Value{}, fmt.Errorf("value of type %s has no member %q", typeDescForVal(obj), e.Member)
	default:
		return Value{}, fmt.Errorf("unknown expression type")
	}
}

func (i *Interpreter) evalUnary(expr *ast.UnaryExpr) (Value, error) {
	right, err := i.evalExpr(expr.Right)
	if err != nil {
		return Value{}, err
	}
	switch expr.Op {
	case "!":
		if right.Null {
			return Value{Untyped: true, BData: true, IsBool: true}, nil
		}
		if !right.IsBool {
			return Value{}, fmt.Errorf("operator ! requires a bool, got %s", typeDescForVal(right))
		}
		return Value{Untyped: true, BData: !right.BData, IsBool: true}, nil
	default:
		return Value{}, fmt.Errorf("unknown unary operator: %s", expr.Op)
	}
}

func (i *Interpreter) evalBinary(expr *ast.BinaryExpr) (Value, error) {
	if expr.Op == "&&" || expr.Op == "||" {
		return i.evalShortCircuit(expr)
	}

	left, err := i.evalExpr(expr.Left)
	if err != nil {
		return Value{}, err
	}
	right, err := i.evalExpr(expr.Right)
	if err != nil {
		return Value{}, err
	}

	switch expr.Op {
	case "==", "!=":
		return i.evalEquality(left, right, expr.Op)
	case "<", ">", "<=", ">=":
		return i.evalComparison(left, right, expr.Op)
	}

	if left.Null || right.Null {
		return Value{Null: true}, nil
	}

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

func (i *Interpreter) evalShortCircuit(expr *ast.BinaryExpr) (Value, error) {
	left, err := i.evalExpr(expr.Left)
	if err != nil {
		return Value{}, err
	}

	if !left.Null && !left.IsBool {
		return Value{}, fmt.Errorf("operator %s requires bools, got %s", expr.Op, typeDescForVal(left))
	}

	if expr.Op == "&&" {
		if left.Null || (left.IsBool && !left.BData) {
			return Value{Untyped: true, BData: false, IsBool: true}, nil
		}
	} else {
		if left.IsBool && left.BData {
			return Value{Untyped: true, BData: true, IsBool: true}, nil
		}
	}

	right, err := i.evalExpr(expr.Right)
	if err != nil {
		return Value{}, err
	}

	if !right.Null && !right.IsBool {
		return Value{}, fmt.Errorf("operator %s requires bools, got %s", expr.Op, typeDescForVal(right))
	}

	if expr.Op == "&&" {
		if right.Null {
			return Value{Untyped: true, BData: false, IsBool: true}, nil
		}
		return Value{Untyped: true, BData: right.BData, IsBool: true}, nil
	} else {
		if right.Null {
			return Value{Untyped: true, BData: false, IsBool: true}, nil
		}
		return Value{Untyped: true, BData: right.BData, IsBool: true}, nil
	}
}

func (i *Interpreter) evalEquality(left, right Value, op string) (Value, error) {
	if left.Null && right.Null {
		if op == "==" {
			return Value{Untyped: true, BData: true, IsBool: true}, nil
		}
		return Value{Untyped: true, BData: false, IsBool: true}, nil
	}

	if left.Null || right.Null {
		if op == "==" {
			return Value{Untyped: true, BData: false, IsBool: true}, nil
		}
		return Value{Untyped: true, BData: true, IsBool: true}, nil
	}

	if left.IsBool != right.IsBool && (left.IsBool || right.IsBool) {
		return Value{}, fmt.Errorf("type mismatch: cannot compare %s %s %s", typeDescForVal(left), op, typeDescForVal(right))
	}
	if left.IsFloat != right.IsFloat && !left.Untyped && !right.Untyped {
		if !canImplicitConvert(right.IType, right.FType, false, ast.IntegerType{}, left.FType, true) &&
			!canImplicitConvert(left.IType, left.FType, false, ast.IntegerType{}, right.FType, true) {
			return Value{}, fmt.Errorf("type mismatch: cannot compare %s %s %s", typeDescForVal(left), op, typeDescForVal(right))
		}
	}

	getFloat := func(v Value) float64 {
		if v.IsFloat {
			return v.FData
		}
		return float64(v.Data)
	}

	var eq bool
	if left.IsBool && right.IsBool {
		eq = left.BData == right.BData
	} else if left.IsFloat || right.IsFloat {
		eq = getFloat(left) == getFloat(right)
	} else {
		eq = left.Data == right.Data
	}

	if op == "!=" {
		eq = !eq
	}
	return Value{Untyped: true, BData: eq, IsBool: true}, nil
}

func (i *Interpreter) evalComparison(left, right Value, op string) (Value, error) {
	if left.Null || right.Null {
		return Value{Untyped: true, BData: false, IsBool: true}, nil
	}

	if left.IsBool || right.IsBool {
		return Value{}, fmt.Errorf("operator %s cannot be used with booleans", op)
	}

	if left.IsFloat != right.IsFloat && !left.Untyped && !right.Untyped {
		if !canImplicitConvert(right.IType, right.FType, false, ast.IntegerType{}, left.FType, true) &&
			!canImplicitConvert(left.IType, left.FType, false, ast.IntegerType{}, right.FType, true) {
			return Value{}, fmt.Errorf("type mismatch: cannot compare %s %s %s", typeDescForVal(left), op, typeDescForVal(right))
		}
	}

	getFloat := func(v Value) float64 {
		if v.IsFloat {
			return v.FData
		}
		return float64(v.Data)
	}

	var cmp bool
	if left.IsFloat || right.IsFloat {
		leftF := getFloat(left)
		rightF := getFloat(right)
		switch op {
		case "<":
			cmp = leftF < rightF
		case ">":
			cmp = leftF > rightF
		case "<=":
			cmp = leftF <= rightF
		case ">=":
			cmp = leftF >= rightF
		}
	} else {
		switch op {
		case "<":
			cmp = left.Data < right.Data
		case ">":
			cmp = left.Data > right.Data
		case "<=":
			cmp = left.Data <= right.Data
		case ">=":
			cmp = left.Data >= right.Data
		}
	}
	return Value{Untyped: true, BData: cmp, IsBool: true}, nil
}
