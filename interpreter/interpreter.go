package interpreter

import (
	"fmt"
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/x448/float16"

	"lang-interpreter/ast"
)

type Value struct {
	Type       ast.Type
	IType      ast.IntegerType
	FType      ast.FloatType
	BType      ast.BoolType
	SType      ast.StringType
	Data       int64
	FData      float64
	BData      bool
	StringData string
	ArrayData  []Value
	Untyped    bool
	IsFloat    bool
	IsBool     bool
	IsArray    bool
	IsString   bool
	Null       bool
	IsType     bool
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
			return -1
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
	if val.IsArray {
		td.Type = val.Type
		td.IsArray = true
	} else if val.IsString {
		td.IsString = true
		td.SType = val.SType
	} else if val.IsFloat {
		td.IsFloat = true
		td.FType = val.FType
	} else if val.IsBool {
		td.IsBool = true
		td.BType = val.BType
	} else {
		td.IType = val.IType
	}
	return td
}

func (i *Interpreter) evalTypeMember(td Value, member string) (Value, error) {
	if td.IsString {
		switch member {
		case "size":
			return Value{Data: int64(td.SType.Size)}, nil
		default:
			return Value{}, fmt.Errorf("type string has no attribute %q", member)
		}
	}
	if td.IsArray {
		switch member {
		case "length":
			if at, ok := td.Type.(ast.ArrayType); ok {
				return Value{Data: int64(at.Size)}, nil
			}
			return Value{}, fmt.Errorf("type has no attribute %q", member)
		case "size":
			if at, ok := td.Type.(ast.ArrayType); ok {
				return Value{Data: int64(at.Size)}, nil
			}
			return Value{}, fmt.Errorf("type has no attribute %q", member)
		case "elem_type":
			if at, ok := td.Type.(ast.ArrayType); ok {
				val := Value{IsType: true, Type: at.ElemType}
				switch et := at.ElemType.(type) {
				case ast.IntegerType:
					val.IType = et
				case ast.FloatType:
					val.IsFloat = true
					val.FType = et
				case ast.BoolType:
					val.IsBool = true
					val.BType = et
				case ast.ArrayType:
					val.IsArray = true
				case ast.ListType:
					val.IsArray = true
				}
				return val, nil
			}
			return Value{}, fmt.Errorf("type has no attribute %q", member)
		default:
			return Value{}, fmt.Errorf("type %s has no attribute %q", typeDescForVal(td), member)
		}
	}
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
	return Value{}, fmt.Errorf("type %s has no attribute %q", typeDescForVal(td), member)
}

func (v Value) String() string {
	if v.IsArray {
		s := "["
		for i, elem := range v.ArrayData {
			if i > 0 {
				s += ", "
			}
			s += elem.String()
		}
		s += "]"
		return s
	}
	if v.IsType {
		return typeDescForVal(v)
	}
	if v.Null {
		return "null"
	}
	if v.IsString {
		return v.StringData
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

func typeDescForType(t ast.Type) string {
	switch typ := t.(type) {
	case ast.ArrayType:
		return fmt.Sprintf("array{size: %d}<%s>", typ.Size, typeDescForType(typ.ElemType))
	case ast.ListType:
		s := "list"
		if typ.HasMin || typ.HasMax {
			s += "{"
			if typ.HasMin {
				s += fmt.Sprintf("min: %d", typ.MinSize)
				if typ.HasMax {
					s += ", "
				}
			}
			if typ.HasMax {
				s += fmt.Sprintf("max: %d", typ.MaxSize)
			}
			s += "}"
		}
		return s + "<" + typeDescForType(typ.ElemType) + ">"
	case ast.StringType:
		s := "string"
		if typ.Size > 0 {
			s += fmt.Sprintf("{size: %d}", typ.Size)
		} else if typ.HasMin || typ.HasMax {
			s += "{"
			if typ.HasMin {
				s += fmt.Sprintf("min: %d", typ.MinSize)
				if typ.HasMax {
					s += ", "
				}
			}
			if typ.HasMax {
				s += fmt.Sprintf("max: %d", typ.MaxSize)
			}
			s += "}"
		}
		return s
	default:
		return typeDesc(typ, false)
	}
}

func typeDescFromVar(it ast.IntegerType, ft ast.FloatType, bt ast.BoolType, isFloat, isBool bool) string {
	if isBool {
		nullability := "non-nullable "
		if bt.Nullable {
			nullability = "nullable "
		}
		return fmt.Sprintf("%sbool", nullability)
	}
	if isFloat {
		if ft.Size == 0 {
			return "untyped float literal"
		}
		nullability := "non-nullable "
		if ft.Nullable {
			nullability = "nullable "
		}
		return fmt.Sprintf("%s%d-bit float", nullability, ft.Size)
	}
	if it.Size == 0 {
		return "untyped int literal"
	}
	nullability := "non-nullable "
	if it.Nullable {
		nullability = "nullable "
	}
	sign := "signed"
	if !it.Signed {
		sign = "unsigned"
	}
	return fmt.Sprintf("%s%d-bit %s int", nullability, it.Size, sign)
}

func typeDescForVal(v Value) string {
	if v.Type != nil {
		return typeDescForType(v.Type)
	}
	return typeDescFromVar(v.IType, v.FType, v.BType, v.IsFloat, v.IsBool)
}

func typeDesc(t interface{}, isBool bool) string {
	switch typ := t.(type) {
	case ast.BoolType:
		if typ.Nullable {
			return "nullable bool"
		}
		return "non-nullable bool"
	case ast.FloatType:
		if typ.Size == 0 {
			return "untyped float literal"
		}
		nullability := "non-nullable "
		if typ.Nullable {
			nullability = "nullable "
		}
		return fmt.Sprintf("%s%d-bit float", nullability, typ.Size)
	case ast.IntegerType:
		if typ.Size == 0 {
			return "untyped int literal"
		}
		sign := "signed"
		if !typ.Signed {
			sign = "unsigned"
		}
		nullability := "non-nullable "
		if typ.Nullable {
			nullability = "nullable "
		}
		return fmt.Sprintf("%s%d-bit %s int", nullability, typ.Size, sign)
	}
	return "unknown"
}

func canImplicitConvert(srcInt ast.IntegerType, srcFloat ast.FloatType, srcIsFloat bool,
	dstInt ast.IntegerType, dstFloat ast.FloatType, dstIsFloat bool) bool {

	srcNullable := srcIsFloat && srcFloat.Nullable || !srcIsFloat && srcInt.Nullable
	dstNullable := dstIsFloat && dstFloat.Nullable || !dstIsFloat && dstInt.Nullable
	if srcNullable && !dstNullable {
		return false
	}

	if srcIsFloat && dstIsFloat {
		return srcFloat.Size <= dstFloat.Size
	}

	if !srcIsFloat && dstIsFloat {
		if srcInt.Size <= 8 {
			return dstFloat.Size >= 16
		}
		if srcInt.Size <= 16 {
			return dstFloat.Size >= 32
		}
		if srcInt.Size <= 32 {
			return dstFloat.Size >= 64
		}
		return false
	}

	if srcIsFloat && !dstIsFloat {
		return false
	}

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
	default:
		return val
	}
}

type LoopSignal int

const (
	BreakSignal LoopSignal = iota
	SkipSignal
)

func (s LoopSignal) Error() string {
	switch s {
	case BreakSignal:
		return "break outside loop"
	case SkipSignal:
		return "skip outside loop"
	}
	return "unknown loop control"
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
		return i.execVarDecl(s)
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
	case *ast.ForStmt:
		return i.execFor(s)
	case *ast.WhileStmt:
		return i.execWhile(s)
	case *ast.BreakStmt:
		return BreakSignal
	case *ast.SkipStmt:
		return SkipSignal
	case *ast.IncDecStmt:
		env := i.env.LookupEnv(s.Name)
		if env == nil {
			return fmt.Errorf("undefined variable: %s", s.Name)
		}
		val, _ := env.Get(s.Name)
		if val.Null {
			return fmt.Errorf("cannot increment/decrement null")
		}
		if val.IsBool || val.IsArray {
			return fmt.Errorf("cannot increment/decrement %s", typeDescForVal(val))
		}
		if s.Op == "++" {
			if val.IsFloat {
				val.FData = convertFloat(val.FData+1, val.FType)
			} else {
				newVal := val.Data + 1
				if err := checkIntFits(newVal, val.IType); err != nil {
					return err
				}
				val.Data = newVal
			}
		} else {
			if val.IsFloat {
				val.FData = convertFloat(val.FData-1, val.FType)
			} else {
				newVal := val.Data - 1
				if err := checkIntFits(newVal, val.IType); err != nil {
					return err
				}
				val.Data = newVal
			}
		}
		env.Set(s.Name, val)
	case *ast.PrintStmt:
		val, err := i.evalExpr(s.Expr)
		if err != nil {
			return err
		}
		if !val.IsString {
			return fmt.Errorf("print requires a string argument, got %s", typeDescForVal(val))
		}
		fmt.Println(val.StringData)
	case *ast.ExprStmt:
		_, err := i.evalExpr(s.Expr)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown statement type")
	}
	return nil
}

func (i *Interpreter) execVarDecl(s *ast.VarDecl) error {
	// Handle array/list/string types
	if at, ok := s.Type.(ast.ArrayType); ok {
		return i.execArrayDecl(s, at)
	}
	if lt, ok := s.Type.(ast.ListType); ok {
		return i.execListDecl(s, lt)
	}
	if st, ok := s.Type.(ast.StringType); ok {
		return i.execStringDecl(s, st)
	}

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
					if !rightVal.Untyped && !canImplicitConvert(rightVal.IType, ast.FloatType{}, false, ast.IntegerType{}, s.FType, true) {
						return fmt.Errorf("type mismatch: cannot convert %s to %s", typeDescForVal(rightVal), typeDescFromVar(ast.IntegerType{}, s.FType, ast.BoolType{}, true, false))
					}
					val.FData = convertFloat(float64(rightVal.Data), s.FType)
				} else {
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
	return nil
}

func (i *Interpreter) execArrayDecl(s *ast.VarDecl, at ast.ArrayType) error {
	if s.Expr != nil {
		rightVal, err := i.evalExpr(s.Expr)
		if err != nil {
			return err
		}
		if !rightVal.IsArray {
			return fmt.Errorf("cannot assign %s to array variable", typeDescForVal(rightVal))
		}
		effectiveSize := at.Size
		if effectiveSize == 0 {
			effectiveSize = len(rightVal.ArrayData)
		}
		if len(rightVal.ArrayData) != effectiveSize {
			return fmt.Errorf("expected %d elements, got %d", effectiveSize, len(rightVal.ArrayData))
		}
		at.Size = effectiveSize
		rightVal.Type = at
		i.env.Set(s.Name, rightVal)
		return nil
	}
	data := make([]Value, at.Size)
	for idx := range data {
		data[idx] = Value{Data: 0, Untyped: true}
	}
	val := Value{Type: at, IsArray: true, ArrayData: data}
	i.env.Set(s.Name, val)
	return nil
}

func (i *Interpreter) execListDecl(s *ast.VarDecl, lt ast.ListType) error {
	if s.Expr != nil {
		rightVal, err := i.evalExpr(s.Expr)
		if err != nil {
			return err
		}
		if !rightVal.IsArray {
			return fmt.Errorf("cannot assign %s to list variable", typeDescForVal(rightVal))
		}
		if lt.HasMin && len(rightVal.ArrayData) < lt.MinSize {
			return fmt.Errorf("expected at least %d elements, got %d", lt.MinSize, len(rightVal.ArrayData))
		}
		if lt.HasMax && len(rightVal.ArrayData) > lt.MaxSize {
			return fmt.Errorf("expected at most %d elements, got %d", lt.MaxSize, len(rightVal.ArrayData))
		}
		rightVal.Type = lt
		i.env.Set(s.Name, rightVal)
		return nil
	}
	if lt.HasMin && lt.MinSize > 0 {
		return fmt.Errorf("list requires at least %d elements, but no initializer given", lt.MinSize)
	}
	val := Value{Type: lt, IsArray: true, ArrayData: []Value{}}
	i.env.Set(s.Name, val)
	return nil
}

func (i *Interpreter) execStringDecl(s *ast.VarDecl, st ast.StringType) error {
	if s.Expr != nil {
		rightVal, err := i.evalExpr(s.Expr)
		if err != nil {
			return err
		}
		if !rightVal.IsString {
			return fmt.Errorf("cannot assign %s to string variable", typeDescForVal(rightVal))
		}
		runeCount := utf8.RuneCountInString(rightVal.StringData)
		if st.Size > 0 && runeCount != st.Size {
			return fmt.Errorf("expected %d characters, got %d", st.Size, runeCount)
		}
		if st.HasMin && runeCount < st.MinSize {
			return fmt.Errorf("expected at least %d characters, got %d", st.MinSize, runeCount)
		}
		if st.HasMax && runeCount > st.MaxSize {
			return fmt.Errorf("expected at most %d characters, got %d", st.MaxSize, runeCount)
		}
		rightVal.Type = st
		rightVal.SType = st
		i.env.Set(s.Name, rightVal)
		return nil
	}
	if st.Size > 0 {
		val := Value{IsString: true, Type: st, SType: st, StringData: string(make([]byte, st.Size))}
		i.env.Set(s.Name, val)
		return nil
	}
	if st.HasMin && st.MinSize > 0 {
		return fmt.Errorf("string requires at least %d characters, but no initializer given", st.MinSize)
	}
	val := Value{IsString: true, Type: st, SType: st, StringData: ""}
	i.env.Set(s.Name, val)
	return nil
}

func (i *Interpreter) executeAssignment(stmt *ast.Assignment) error {
	if stmt.Index != nil {
		return i.executeIndexedAssign(stmt)
	}

	env := i.env.LookupEnv(stmt.Name)
	if env == nil {
		return fmt.Errorf("undefined variable: %s", stmt.Name)
	}
	val, _ := env.Get(stmt.Name)

	if val.IsString {
		if stmt.Op != "=" {
			return fmt.Errorf("cannot use operator %s on string variable", stmt.Op)
		}
		rightVal, err := i.evalExpr(stmt.Expr)
		if err != nil {
			return err
		}
		if !rightVal.IsString {
			return fmt.Errorf("cannot assign %s to string", typeDescForVal(rightVal))
		}
		runeCount := utf8.RuneCountInString(rightVal.StringData)
		if st, ok := val.Type.(ast.StringType); ok {
			if st.Size > 0 && runeCount != st.Size {
				return fmt.Errorf("expected %d characters, got %d", st.Size, runeCount)
			}
			if st.HasMin && runeCount < st.MinSize {
				return fmt.Errorf("expected at least %d characters, got %d", st.MinSize, runeCount)
			}
			if st.HasMax && runeCount > st.MaxSize {
				return fmt.Errorf("expected at most %d characters, got %d", st.MaxSize, runeCount)
			}
		}
		val.StringData = rightVal.StringData
		val.Null = false
		env.Set(stmt.Name, val)
		return nil
	}

	if val.IsArray {
		typeName := "array"
		if _, ok := val.Type.(ast.ListType); ok {
			typeName = "list"
		}
		return fmt.Errorf("cannot use operator %s on %s variable", stmt.Op, typeName)
	}

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
		return fmt.Errorf("cannot assign %s to int variable", typeDescForVal(rightVal))
	} else if val.IsFloat {
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

func (i *Interpreter) executeIndexedAssign(stmt *ast.Assignment) error {
	env := i.env.LookupEnv(stmt.Name)
	if env == nil {
		return fmt.Errorf("undefined variable: %s", stmt.Name)
	}
	val, _ := env.Get(stmt.Name)
	if val.IsString {
		if stmt.Op != "=" {
			return fmt.Errorf("operator %s not supported for string indexed assignment", stmt.Op)
		}
		idxVal, err := i.evalExpr(stmt.Index)
		if err != nil {
			return err
		}
		if idxVal.Null {
			return fmt.Errorf("index cannot be null")
		}
		idx := int(idxVal.Data)
		runes := []rune(val.StringData)
		if idx < 0 || idx >= len(runes) {
			return fmt.Errorf("index %d out of bounds for length %d", idx, len(runes))
		}
		rightVal, err := i.evalExpr(stmt.Expr)
		if err != nil {
			return err
		}
		if !rightVal.IsString || utf8.RuneCountInString(rightVal.StringData) != 1 {
			return fmt.Errorf("string index assignment requires a single character string")
		}
		runes[idx] = []rune(rightVal.StringData)[0]
		val.StringData = string(runes)
		env.Set(stmt.Name, val)
		return nil
	}
	if !val.IsArray {
		return fmt.Errorf("cannot index non-array variable")
	}
	if stmt.Op != "=" {
		return fmt.Errorf("operator %s not supported for indexed assignment", stmt.Op)
	}

	idxVal, err := i.evalExpr(stmt.Index)
	if err != nil {
		return err
	}
	if idxVal.Null {
		return fmt.Errorf("index cannot be null")
	}
	idx := int(idxVal.Data)
	if idx < 0 || idx >= len(val.ArrayData) {
		return fmt.Errorf("index %d out of bounds for length %d", idx, len(val.ArrayData))
	}

	rightVal, err := i.evalExpr(stmt.Expr)
	if err != nil {
		return err
	}
	val.ArrayData[idx] = rightVal
	env.Set(stmt.Name, val)
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

func (i *Interpreter) execFor(s *ast.ForStmt) error {
	i.pushScope()
	defer i.popScope()

	if s.Init != nil {
		if err := i.executeStmt(s.Init); err != nil {
			return err
		}
	}

	for {
		if s.Condition != nil {
			condVal, err := i.evalExpr(s.Condition)
			if err != nil {
				return err
			}
			if !condVal.IsBool {
				return fmt.Errorf("for condition must be a bool, got %s", typeDescForVal(condVal))
			}
			if !condVal.BData {
				return nil
			}
		}

		i.pushScope()
		err := i.executeStmt(s.Body)
		i.popScope()

		if err != nil {
			if sig, ok := err.(LoopSignal); ok {
				switch sig {
				case SkipSignal:
					goto doUpdate
				case BreakSignal:
					return nil
				}
			}
			return err
		}

	doUpdate:
		if s.Update != nil {
			if err := i.executeStmt(s.Update); err != nil {
				return err
			}
		}
	}
}

func (i *Interpreter) execWhile(s *ast.WhileStmt) error {
	for {
		condVal, err := i.evalExpr(s.Condition)
		if err != nil {
			return err
		}
		if !condVal.IsBool {
			return fmt.Errorf("while condition must be a bool, got %s", typeDescForVal(condVal))
		}
		if !condVal.BData {
			return nil
		}

		i.pushScope()
		err = i.executeStmt(s.Body)
		i.popScope()

		if err != nil {
			if sig, ok := err.(LoopSignal); ok {
				switch sig {
				case SkipSignal:
					continue
				case BreakSignal:
					return nil
				}
			}
			return err
		}
	}
}

func (i *Interpreter) evalExpr(expr ast.Expr) (Value, error) {
	switch e := expr.(type) {
	case *ast.StringLit:
		if e.Untyped {
			return Value{IsString: true, Untyped: true, StringData: e.Value}, nil
		}
		return Value{IsString: true, SType: e.SType, StringData: e.Value}, nil
	case *ast.IntegerLit:
		if e.Untyped {
			return Value{Untyped: true, Data: e.Value}, nil
		}
		return Value{IType: e.IType, Data: e.Value}, nil
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
		switch t := e.Type.(type) {
		case ast.FloatType:
			val.Type = t
			val.IsFloat = true
			val.FType = t
		case ast.BoolType:
			val.Type = t
			val.IsBool = true
			val.BType = t
		case ast.IntegerType:
			val.Type = t
			val.IType = t
		case ast.ArrayType:
			val.Type = t
			val.IsArray = true
		case ast.ListType:
			val.Type = t
			val.IsArray = true
		case ast.StringType:
			val.Type = t
			val.IsString = true
			val.SType = t
		}
		return val, nil
	case *ast.ArrayLit:
		var elements []Value
		for _, el := range e.Elements {
			ev, err := i.evalExpr(el)
			if err != nil {
				return Value{}, err
			}
			elements = append(elements, ev)
		}
		return Value{IsArray: true, ArrayData: elements}, nil
	case *ast.TypeOfExpr:
		val, err := i.evalExpr(e.Expr)
		if err != nil {
			return Value{}, err
		}
		return valueToTypeDesc(val), nil
	case *ast.IndexExpr:
		obj, err := i.evalExpr(e.Object)
		if err != nil {
			return Value{}, err
		}
		idxVal, err := i.evalExpr(e.Index)
		if err != nil {
			return Value{}, err
		}
		if idxVal.Null {
			return Value{}, fmt.Errorf("index cannot be null")
		}
		idx := int(idxVal.Data)
		if obj.IsString {
			runes := []rune(obj.StringData)
			if idx < 0 || idx >= len(runes) {
				return Value{}, fmt.Errorf("index %d out of bounds for length %d", idx, len(runes))
			}
			return Value{IsString: true, StringData: string(runes[idx])}, nil
		}
		if !obj.IsArray {
			return Value{}, fmt.Errorf("cannot index non-array value")
		}
		if idx < 0 || idx >= len(obj.ArrayData) {
			return Value{}, fmt.Errorf("index %d out of bounds for length %d", idx, len(obj.ArrayData))
		}
		return obj.ArrayData[idx], nil
	case *ast.MemberAccess:
		obj, err := i.evalExpr(e.Object)
		if err != nil {
			return Value{}, err
		}
		if e.Member == "toString" && e.Args != nil {
			if len(e.Args) != 0 {
				return Value{}, fmt.Errorf("toString takes no arguments")
			}
			if obj.IsType {
				return Value{IsString: true, StringData: typeDescForVal(obj)}, nil
			}
			if obj.Null {
				return Value{IsString: true, StringData: "null"}, nil
			}
			if obj.IsString {
				return Value{IsString: true, StringData: obj.StringData}, nil
			}
			if obj.IsBool {
				return Value{IsString: true, StringData: strconv.FormatBool(obj.BData)}, nil
			}
			if obj.IsFloat {
				return Value{IsString: true, StringData: formatFloat(obj.FData)}, nil
			}
			if obj.IsArray {
				return Value{IsString: true, StringData: obj.String()}, nil
			}
			return Value{IsString: true, StringData: intToStr(obj.Data, obj.IType, obj.Untyped)}, nil
		}
		if obj.IsType {
			return i.evalTypeMember(obj, e.Member)
		}
		if obj.IsArray {
			result, err := i.evalArrayMember(&obj, e)
			if err != nil {
				return Value{}, err
			}
			if e.Member == "add" || e.Member == "remove" {
				if ref, ok := e.Object.(*ast.VarRef); ok {
					env := i.env.LookupEnv(ref.Name)
					if env != nil {
						env.Set(ref.Name, obj)
					}
				}
			}
			return result, nil
		}
		if obj.IsString {
			result, err := i.evalStringMember(&obj, e)
			if err != nil {
				return Value{}, err
			}
			if e.Member == "add" || e.Member == "remove" {
				if ref, ok := e.Object.(*ast.VarRef); ok {
					env := i.env.LookupEnv(ref.Name)
					if env != nil {
						env.Set(ref.Name, obj)
					}
				}
			}
			return result, nil
		}
		return Value{}, fmt.Errorf("value of type %s has no attribute %q", typeDescForVal(obj), e.Member)
	default:
		return Value{}, fmt.Errorf("unknown expression type")
	}
}

func (i *Interpreter) evalArrayMember(obj *Value, e *ast.MemberAccess) (Value, error) {
	switch e.Member {
	case "length":
		return Value{Data: int64(len(obj.ArrayData))}, nil
	case "add":
		if len(e.Args) == 0 || len(e.Args) > 2 {
			return Value{}, fmt.Errorf("add requires 1 or 2 arguments")
		}
		val, err := i.evalExpr(e.Args[0])
		if err != nil {
			return Value{}, err
		}
		idx := len(obj.ArrayData)
		if len(e.Args) == 2 {
			idxVal, err := i.evalExpr(e.Args[1])
			if err != nil {
				return Value{}, err
			}
			if idxVal.Null {
				return Value{}, fmt.Errorf("index cannot be null")
			}
			idx = int(idxVal.Data)
		}
		// Check bounds for list type
		if lt, ok := obj.Type.(ast.ListType); ok {
			if lt.HasMax && len(obj.ArrayData) >= lt.MaxSize {
				return Value{}, fmt.Errorf("list is at maximum capacity (%d)", lt.MaxSize)
			}
		}
		if idx < 0 || idx > len(obj.ArrayData) {
			return Value{}, fmt.Errorf("index %d out of bounds for length %d", idx, len(obj.ArrayData))
		}
		// Insert at index
		newData := make([]Value, len(obj.ArrayData)+1)
		copy(newData, obj.ArrayData[:idx])
		newData[idx] = val
		copy(newData[idx+1:], obj.ArrayData[idx:])
		obj.ArrayData = newData
		return Value{Null: true}, nil
	case "remove":
		if len(e.Args) != 1 {
			return Value{}, fmt.Errorf("remove requires 1 argument")
		}
		idxVal, err := i.evalExpr(e.Args[0])
		if err != nil {
			return Value{}, err
		}
		if idxVal.Null {
			return Value{}, fmt.Errorf("index cannot be null")
		}
		idx := int(idxVal.Data)
		if idx < 0 || idx >= len(obj.ArrayData) {
			return Value{}, fmt.Errorf("index %d out of bounds for length %d", idx, len(obj.ArrayData))
		}
		// Check min bound for list type
		if lt, ok := obj.Type.(ast.ListType); ok {
			if lt.HasMin && len(obj.ArrayData) <= lt.MinSize {
				return Value{}, fmt.Errorf("list is at minimum capacity (%d)", lt.MinSize)
			}
		}
		removed := obj.ArrayData[idx]
		newData := make([]Value, len(obj.ArrayData)-1)
		copy(newData, obj.ArrayData[:idx])
		copy(newData[idx:], obj.ArrayData[idx+1:])
		obj.ArrayData = newData
		return removed, nil
	default:
		typeName := "array"
		if _, ok := obj.Type.(ast.ListType); ok {
			typeName = "list"
		}
		return Value{}, fmt.Errorf("%s has no attribute %q", typeName, e.Member)
	}
}

func (i *Interpreter) evalStringMember(obj *Value, e *ast.MemberAccess) (Value, error) {
	switch e.Member {
	case "length":
		runeCount := utf8.RuneCountInString(obj.StringData)
		return Value{Data: int64(runeCount)}, nil
	case "concat":
		if len(e.Args) != 1 {
			return Value{}, fmt.Errorf("concat requires 1 argument")
		}
		val, err := i.evalExpr(e.Args[0])
		if err != nil {
			return Value{}, err
		}
		if !val.IsString {
			return Value{}, fmt.Errorf("cannot concat %s to string", typeDescForVal(val))
		}
		return Value{IsString: true, StringData: obj.StringData + val.StringData}, nil
	case "add":
		if len(e.Args) == 0 || len(e.Args) > 2 {
			return Value{}, fmt.Errorf("add requires 1 or 2 arguments")
		}
		if st, ok := obj.Type.(ast.StringType); ok {
			if st.Size > 0 {
				return Value{}, fmt.Errorf("cannot add to fixed-size string")
			}
		}
		val, err := i.evalExpr(e.Args[0])
		if err != nil {
			return Value{}, err
		}
		if !val.IsString || utf8.RuneCountInString(val.StringData) != 1 {
			return Value{}, fmt.Errorf("add requires a single character string")
		}
		runes := []rune(obj.StringData)
		idx := len(runes)
		if len(e.Args) == 2 {
			idxVal, err := i.evalExpr(e.Args[1])
			if err != nil {
				return Value{}, err
			}
			if idxVal.Null {
				return Value{}, fmt.Errorf("index cannot be null")
			}
			idx = int(idxVal.Data)
		}
		// Check max bound for dynamic string type
		if st, ok := obj.Type.(ast.StringType); ok {
			if st.HasMax && len(runes) >= st.MaxSize {
				return Value{}, fmt.Errorf("string is at maximum capacity (%d)", st.MaxSize)
			}
		}
		if idx < 0 || idx > len(runes) {
			return Value{}, fmt.Errorf("index %d out of bounds for length %d", idx, len(runes))
		}
		newRunes := make([]rune, 0, len(runes)+1)
		newRunes = append(newRunes, runes[:idx]...)
		newRunes = append(newRunes, []rune(val.StringData)...)
		newRunes = append(newRunes, runes[idx:]...)
		obj.StringData = string(newRunes)
		return Value{Null: true}, nil
	case "remove":
		if len(e.Args) != 1 {
			return Value{}, fmt.Errorf("remove requires 1 argument")
		}
		if st, ok := obj.Type.(ast.StringType); ok {
			if st.Size > 0 {
				return Value{}, fmt.Errorf("cannot remove from fixed-size string")
			}
		}
		idxVal, err := i.evalExpr(e.Args[0])
		if err != nil {
			return Value{}, err
		}
		if idxVal.Null {
			return Value{}, fmt.Errorf("index cannot be null")
		}
		idx := int(idxVal.Data)
		runes := []rune(obj.StringData)
		if idx < 0 || idx >= len(runes) {
			return Value{}, fmt.Errorf("index %d out of bounds for length %d", idx, len(runes))
		}
		if st, ok := obj.Type.(ast.StringType); ok {
			if st.HasMin && len(runes) <= st.MinSize {
				return Value{}, fmt.Errorf("string is at minimum capacity (%d)", st.MinSize)
			}
		}
		removed := string(runes[idx])
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:idx]...)
		newRunes = append(newRunes, runes[idx+1:]...)
		obj.StringData = string(newRunes)
		return Value{IsString: true, StringData: removed}, nil
	default:
		return Value{}, fmt.Errorf("string has no attribute %q", e.Member)
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

	// String concatenation
	if expr.Op == "+" && (left.IsString || right.IsString) {
		if !left.IsString || !right.IsString {
			return Value{}, fmt.Errorf("cannot concatenate %s and %s", typeDescForVal(left), typeDescForVal(right))
		}
		return Value{IsString: true, StringData: left.StringData + right.StringData}, nil
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

	if left.IsString && right.IsString {
		eq := left.StringData == right.StringData
		if op == "!=" {
			eq = !eq
		}
		return Value{Untyped: true, BData: eq, IsBool: true}, nil
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
