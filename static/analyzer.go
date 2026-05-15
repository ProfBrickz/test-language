package static

import (
	"fmt"
	"math"

	"lang-interpreter/ast"
)

type AbsKind int

const (
	AbsInt AbsKind = iota
	AbsFloat
	AbsBool
	AbsString
	AbsArray
	AbsList
	AbsNull
)

type AbsValue struct {
	kind AbsKind

	knownInt bool
	exactInt int64
	minInt   int64
	maxInt   int64
	isAnyInt bool

	knownFloat bool
	exactFloat float64
	isAnyFloat bool

	knownBool bool
	exactBool bool
	isAnyBool bool

	knownStr bool
	exactStr string
	isAnyStr bool

	nullable       bool
	definitelyNull bool

	intType    ast.IntegerType
	floatType  ast.FloatType
	boolType   ast.BoolType
	stringType ast.StringType
	arrayType  ast.ArrayType
	listType   ast.ListType
}

func knownIntValue(v int64, t ast.IntegerType) AbsValue {
	return AbsValue{
		kind:     AbsInt,
		knownInt: true,
		exactInt: v,
		intType:  t,
		nullable: t.Nullable,
	}
}

func anyIntValue(t ast.IntegerType) AbsValue {
	min, max := intRange(t)
	return AbsValue{
		kind:     AbsInt,
		isAnyInt: true,
		minInt:   min,
		maxInt:   max,
		intType:  t,
		nullable: t.Nullable,
	}
}

func rangeIntValue(min, max int64, t ast.IntegerType) AbsValue {
	return AbsValue{
		kind:     AbsInt,
		minInt:   min,
		maxInt:   max,
		intType:  t,
		nullable: t.Nullable,
	}
}

func knownFloatValue(v float64, t ast.FloatType) AbsValue {
	return AbsValue{
		kind:       AbsFloat,
		knownFloat: true,
		exactFloat: v,
		floatType:  t,
		nullable:   t.Nullable,
	}
}

func anyFloatValue(t ast.FloatType) AbsValue {
	return AbsValue{
		kind:       AbsFloat,
		isAnyFloat: true,
		floatType:  t,
		nullable:   t.Nullable,
	}
}

func knownBoolValue(v bool, t ast.BoolType) AbsValue {
	return AbsValue{
		kind:      AbsBool,
		knownBool: true,
		exactBool: v,
		boolType:  t,
		nullable:  t.Nullable,
	}
}

func anyBoolValue(t ast.BoolType) AbsValue {
	return AbsValue{
		kind:      AbsBool,
		isAnyBool: true,
		boolType:  t,
		nullable:  t.Nullable,
	}
}

func nullAbsValue() AbsValue {
	return AbsValue{
		kind:           AbsNull,
		definitelyNull: true,
		nullable:       true,
	}
}

func (a AbsValue) definitelyZero() bool {
	switch a.kind {
	case AbsInt:
		return a.knownInt && a.exactInt == 0
	}
	return false
}

func (a AbsValue) couldBeZero() bool {
	switch a.kind {
	case AbsInt:
		if a.knownInt {
			return a.exactInt == 0
		}
		return a.isAnyInt || a.nullable || (a.minInt <= 0 && a.maxInt >= 0)
	case AbsNull:
		return true
	}
	return false
}

func (a AbsValue) definitelyNotNull() bool {
	return !a.nullable && !a.definitelyNull
}

func (a AbsValue) canBeNull() bool {
	return a.nullable || a.definitelyNull
}

func intRange(t ast.IntegerType) (int64, int64) {
	if !t.Signed {
		switch t.Size {
		case 8:
			return 0, math.MaxUint8
		case 16:
			return 0, math.MaxUint16
		case 32:
			return 0, math.MaxUint32
		default:
			return 0, math.MaxInt64
		}
	}
	switch t.Size {
	case 8:
		return math.MinInt8, math.MaxInt8
	case 16:
		return math.MinInt16, math.MaxInt16
	case 32:
		return math.MinInt32, math.MaxInt32
	default:
		return math.MinInt64, math.MaxInt64
	}
}

type NullMode int

const (
	NullWarn NullMode = iota
	NullError
	NullNone
)

type guardInfo struct {
	name    string
	notZero bool
	notNull bool
}

type Analyzer struct {
	env        map[string]AbsValue
	errors     []string
	warnings   []string
	guards     []guardInfo
	nullMode   NullMode
	funcs      map[string]*ast.FuncDecl
	insideFunc *ast.FuncDecl
}

func New() *Analyzer {
	return &Analyzer{
		env:      make(map[string]AbsValue),
		nullMode: NullWarn,
		funcs:    make(map[string]*ast.FuncDecl),
	}
}

func (a *Analyzer) SetNullMode(mode NullMode) {
	a.nullMode = mode
}

func (a *Analyzer) Errors() []string {
	return a.errors
}

func (a *Analyzer) Warnings() []string {
	return a.warnings
}

func (a *Analyzer) Analyze(program *ast.Program) {
	for _, stmt := range program.Stmts {
		a.collectFunc(stmt)
	}
	for _, stmt := range program.Stmts {
		a.analyzeStmt(stmt)
	}
	for _, fn := range a.funcs {
		a.analyzeFunc(fn)
	}
}

func (a *Analyzer) collectFunc(stmt ast.Stmt) {
	if fn, ok := stmt.(*ast.FuncDecl); ok {
		if _, exists := a.funcs[fn.Name]; exists {
			a.addError(fn.Line, "duplicate function: "+fn.Name)
			return
		}
		a.funcs[fn.Name] = fn
	}
}

func (a *Analyzer) analyzeFunc(fn *ast.FuncDecl) {
	a.insideFunc = fn
	a.pushScope()
	for _, param := range fn.Parameters {
		a.env[param.Name] = defaultValueForParam(param.Type)
	}
	a.analyzeBlock(fn.Body.Stmts)
	a.popScope()
	a.insideFunc = nil
}

func defaultValueForParam(t ast.Type) AbsValue {
	switch typ := t.(type) {
	case ast.IntegerType:
		return anyIntValue(typ)
	case ast.FloatType:
		return anyFloatValue(typ)
	case ast.BoolType:
		return anyBoolValue(typ)
	case ast.StringType:
		return AbsValue{kind: AbsString, isAnyStr: true}
	case ast.ArrayType:
		return AbsValue{kind: AbsArray}
	case ast.ListType:
		return AbsValue{kind: AbsList}
	}
	return AbsValue{}
}

func (a *Analyzer) addError(line int, msg string) {
	a.errors = append(a.errors, fmt.Sprintf("line %d: %s", line, msg))
}

func (a *Analyzer) addWarning(line int, msg string) {
	a.warnings = append(a.warnings, fmt.Sprintf("line %d: %s", line, msg))
}

func (a *Analyzer) analyzeStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		a.analyzeVarDecl(s)
	case *ast.Assignment:
		a.analyzeAssignment(s)
	case *ast.RefDecl:
		a.analyzeRefDecl(s)
	case *ast.IfStmt:
		a.analyzeIfStmt(s)
	case *ast.ForStmt:
		a.analyzeForStmt(s)
	case *ast.ForInStmt:
		a.analyzeForInStmt(s)
	case *ast.ForAtStmt:
		a.analyzeForAtStmt(s)
	case *ast.ForOfStmt:
		a.analyzeForOfStmt(s)
	case *ast.WhileStmt:
		a.analyzeWhileStmt(s)
	case *ast.BlockStmt:
		a.analyzeBlock(s.Stmts)
	case *ast.ExprStmt:
		a.analyzeExpr(s.Expr)
	case *ast.PrintStmt:
		a.analyzeExpr(s.Expr)
	case *ast.IncDecStmt:
		a.analyzeIncDec(s)
	case *ast.BreakStmt:
	case *ast.SkipStmt:
	case *ast.FuncDecl:
	case *ast.ReturnStmt:
		a.analyzeReturn(s)
	}
}

func (a *Analyzer) analyzeReturn(s *ast.ReturnStmt) {
	if a.insideFunc == nil {
		a.addError(s.Line, "return outside function")
		return
	}
	if a.insideFunc.ReturnType == nil && s.Value != nil {
		a.addError(s.Line, "return with value in void function")
		return
	}
	if a.insideFunc.ReturnType != nil && s.Value == nil {
		a.addError(s.Line, "bare return in typed function")
		return
	}
	if s.Value != nil {
		a.analyzeExpr(s.Value)
	}
}

func (a *Analyzer) pushScope() {
	newEnv := make(map[string]AbsValue)
	for k, v := range a.env {
		newEnv[k] = v
	}
	a.env = newEnv
}

func (a *Analyzer) popScope() {
	_ = 0
}

func (a *Analyzer) resolveVar(name string) (AbsValue, bool) {
	v, ok := a.env[name]
	return v, ok
}

func (a *Analyzer) analyzeVarDecl(d *ast.VarDecl) {
	var val AbsValue
	if d.Expr != nil {
		val = a.analyzeExpr(d.Expr)
	} else {
		val = a.defaultValue(d)
	}
	a.checkAssignmentType(d.Line, d.Name, val, d)
	val.nullable = isNullableDecl(d)
	a.env[d.Name] = val
}

func (a *Analyzer) analyzeRefDecl(d *ast.RefDecl) {
	val := a.analyzeExpr(d.Expr)
	a.env[d.Name] = val
}

func (a *Analyzer) analyzeAssignment(s *ast.Assignment) {
	rhs := a.analyzeExpr(s.Expr)
	existing, ok := a.resolveVar(s.Name)
	if !ok {
		return
	}

	if existing.kind == AbsInt && rhs.kind == AbsFloat {
		a.addError(s.Line, "cannot assign float value to int variable")
		return
	}

	if rhs.definitelyNull && !existing.nullable {
		a.addError(s.Line, fmt.Sprintf("cannot assign null to %s", typeDescFromAbs(existing)))
		return
	}

	a.env[s.Name] = rhs
	_ = existing
}

func (a *Analyzer) analyzeIncDec(s *ast.IncDecStmt) {
	existing, ok := a.resolveVar(s.Name)
	if !ok {
		return
	}
	if existing.canBeNull() {
		return
	}
}

func (a *Analyzer) checkAssignmentType(line int, name string, rhs AbsValue, decl *ast.VarDecl) {
	if rhs.definitelyNull {
		if decl != nil {
			if !isNullableDecl(decl) {
				a.addError(line, fmt.Sprintf("cannot assign null to %s", typeDescFromDecl(decl)))
				return
			}
		}
	}
}

func isNullableDecl(d *ast.VarDecl) bool {
	if d.IType.Nullable {
		return true
	}
	if d.IsFloat && d.FType.Nullable {
		return true
	}
	if d.IsBool && d.BType.Nullable {
		return true
	}
	if d.IsString {
		return false
	}
	return false
}

func typeDescFromAbs(v AbsValue) string {
	switch v.kind {
	case AbsInt:
		return fmt.Sprintf("int{size: %d, signed: %v, nullable: %v}", v.intType.Size, v.intType.Signed, v.nullable)
	case AbsFloat:
		return fmt.Sprintf("float{size: %d, nullable: %v}", v.floatType.Size, v.nullable)
	case AbsBool:
		return fmt.Sprintf("bool{nullable: %v}", v.nullable)
	case AbsString:
		return "string"
	default:
		return "unknown"
	}
}

func typeDescFromDecl(d *ast.VarDecl) string {
	if d.IsFloat {
		return fmt.Sprintf("float{size: %d, nullable: %v}", d.FType.Size, d.FType.Nullable)
	}
	if d.IsBool {
		return fmt.Sprintf("bool{nullable: %v}", d.BType.Nullable)
	}
	if d.IsString {
		return fmt.Sprintf("string")
	}
	return fmt.Sprintf("int{size: %d, signed: %v, nullable: %v}", d.IType.Size, d.IType.Signed, d.IType.Nullable)
}

func (a *Analyzer) defaultValue(d *ast.VarDecl) AbsValue {
	if d.IsFloat {
		return anyFloatValue(d.FType)
	}
	if d.IsBool {
		return anyBoolValue(d.BType)
	}
	if d.IType.Size > 0 {
		return anyIntValue(d.IType)
	}
	return anyIntValue(ast.IntegerType{Size: 64, Signed: true, Nullable: true})
}

func (a *Analyzer) isGuardedNotZero(name string) bool {
	for _, g := range a.guards {
		if g.name == name && g.notZero {
			return true
		}
	}
	return false
}

func (a *Analyzer) isGuardedNotNull(name string) bool {
	for _, g := range a.guards {
		if g.name == name && g.notNull {
			return true
		}
	}
	return false
}

func (a *Analyzer) analyzeBlock(stmts []ast.Stmt) {
	a.pushScope()
	for _, stmt := range stmts {
		a.analyzeStmt(stmt)
	}
	a.popScope()
}

func (a *Analyzer) analyzeIfStmt(s *ast.IfStmt) {
	cond := a.analyzeExpr(s.Condition)

	if cond.kind == AbsBool && cond.knownBool {
		if cond.exactBool {
			a.analyzeBlock(s.Then.Stmts)
		} else if s.Else != nil {
			if block, ok := s.Else.(*ast.BlockStmt); ok {
				a.analyzeBlock(block.Stmts)
			} else {
				a.analyzeStmt(s.Else)
			}
		}
		return
	}

	a.guards = append(a.guards, a.extractGuards(s.Condition)...)
	a.analyzeBlock(s.Then.Stmts)
	a.guards = a.guards[:len(a.guards)-len(a.extractGuards(s.Condition))]

	if s.Else != nil {
		if block, ok := s.Else.(*ast.BlockStmt); ok {
			a.analyzeBlock(block.Stmts)
		} else {
			a.analyzeStmt(s.Else)
		}
	}
}

func (a *Analyzer) extractGuards(cond ast.Expr) []guardInfo {
	var guards []guardInfo
	switch e := cond.(type) {
	case *ast.BinaryExpr:
		if e.Op == "!=" {
			if ref, ok := e.Left.(*ast.VarRef); ok {
				if lit, ok := e.Right.(*ast.IntegerLit); ok && lit.Value == 0 {
					guards = append(guards, guardInfo{name: ref.Name, notZero: true})
				}
				if _, ok := e.Right.(*ast.NullLit); ok {
					guards = append(guards, guardInfo{name: ref.Name, notNull: true})
				}
			}
			if ref, ok := e.Right.(*ast.VarRef); ok {
				if lit, ok := e.Left.(*ast.IntegerLit); ok && lit.Value == 0 {
					guards = append(guards, guardInfo{name: ref.Name, notZero: true})
				}
				if _, ok := e.Left.(*ast.NullLit); ok {
					guards = append(guards, guardInfo{name: ref.Name, notNull: true})
				}
			}
		}
		if e.Op == "==" {
			if ref, ok := e.Left.(*ast.VarRef); ok {
				if lit, ok := e.Right.(*ast.IntegerLit); ok {
					guards = append(guards, guardInfo{name: ref.Name, notZero: lit.Value != 0, notNull: true})
				}
				if _, ok := e.Right.(*ast.NullLit); ok {
					guards = append(guards, guardInfo{name: ref.Name, notNull: false})
				}
			}
			if ref, ok := e.Right.(*ast.VarRef); ok {
				if lit, ok := e.Left.(*ast.IntegerLit); ok {
					guards = append(guards, guardInfo{name: ref.Name, notZero: lit.Value != 0, notNull: true})
				}
				if _, ok := e.Left.(*ast.NullLit); ok {
					guards = append(guards, guardInfo{name: ref.Name, notNull: false})
				}
			}
		}
	}
	return guards
}

func (a *Analyzer) analyzeForStmt(s *ast.ForStmt) {
	a.pushScope()
	if s.Init != nil {
		a.analyzeStmt(s.Init)
	}
	if s.Condition != nil {
		a.analyzeExpr(s.Condition)
	}
	a.analyzeBlock(s.Body.Stmts)
	if s.Update != nil {
		a.analyzeStmt(s.Update)
	}
	a.popScope()
}

func (a *Analyzer) analyzeForInStmt(s *ast.ForInStmt) {
	a.analyzeExpr(s.Iter)
	a.analyzeBlock(s.Body.Stmts)
}

func (a *Analyzer) analyzeForAtStmt(s *ast.ForAtStmt) {
	a.analyzeExpr(s.Iter)
	a.analyzeBlock(s.Body.Stmts)
}

func (a *Analyzer) analyzeForOfStmt(s *ast.ForOfStmt) {
	a.analyzeExpr(s.Iter)
	a.analyzeBlock(s.Body.Stmts)
}

func (a *Analyzer) analyzeWhileStmt(s *ast.WhileStmt) {
	if s.Condition != nil {
		a.analyzeExpr(s.Condition)
	}
	a.analyzeBlock(s.Body.Stmts)
}

func (a *Analyzer) analyzeExpr(expr ast.Expr) AbsValue {
	switch e := expr.(type) {
	case *ast.IntegerLit:
		t := e.IType
		if e.Untyped {
			t = ast.IntegerType{Size: 64, Signed: true, Nullable: true}
		}
		v := knownIntValue(e.Value, t)
		v.nullable = false
		return v
	case *ast.FloatLit:
		t := e.FType
		if e.Untyped {
			t = ast.FloatType{Size: 64, Nullable: true}
		}
		v := knownFloatValue(e.Value, t)
		v.nullable = false
		return v
	case *ast.BoolLit:
		t := e.BType
		if e.Untyped {
			t = ast.BoolType{Nullable: true}
		}
		v := knownBoolValue(e.Value, t)
		v.nullable = false
		return v
	case *ast.StringLit:
		return AbsValue{kind: AbsString, knownStr: true, exactStr: e.Value}
	case *ast.NullLit:
		return nullAbsValue()
	case *ast.VarRef:
		if v, ok := a.resolveVar(e.Name); ok {
			return v
		}
		return AbsValue{}
	case *ast.BinaryExpr:
		return a.analyzeBinaryExpr(e)
	case *ast.UnaryExpr:
		return a.analyzeUnaryExpr(e)
	case *ast.ArrayLit:
		return AbsValue{kind: AbsArray}
	case *ast.IndexExpr:
		a.analyzeExpr(e.Index)
		return AbsValue{}
	case *ast.MemberAccess:
		return AbsValue{}
	case *ast.TypeOfExpr:
		return AbsValue{}
	case *ast.CopyExpr:
		return a.analyzeExpr(e.Right)
	case *ast.RefExpr:
		return a.analyzeExpr(e.Right)
	case *ast.IsExpr:
		return AbsValue{}
	case *ast.CallExpr:
		return a.analyzeCall(e)
	}
	return AbsValue{}
}

func (a *Analyzer) analyzeCall(e *ast.CallExpr) AbsValue {
	if ref, ok := e.Function.(*ast.VarRef); ok {
		decl, exists := a.funcs[ref.Name]
		if !exists {
			a.addError(e.Line, "undefined function: "+ref.Name)
			return AbsValue{}
		}
		if len(e.Args) != len(decl.Parameters) {
			a.addError(e.Line, fmt.Sprintf("function %s expects %d arguments, got %d", ref.Name, len(decl.Parameters), len(e.Args)))
			return AbsValue{}
		}
		for _, arg := range e.Args {
			a.analyzeExpr(arg)
		}
		if decl.ReturnType != nil {
			return defaultValueForParam(decl.ReturnType)
		}
	}
	return AbsValue{}
}

func (a *Analyzer) analyzeBinaryExpr(e *ast.BinaryExpr) AbsValue {
	left := a.analyzeExpr(e.Left)
	right := a.analyzeExpr(e.Right)

	switch e.Op {
	case "/", "%":
		a.checkDivisionByZero(e.Line, e.Right, right, e.Op)
	case "+", "-", "*":
		if left.kind == AbsInt && right.kind == AbsInt {
			return a.foldIntArith(left, right, e.Op)
		}
	case "<", ">", "<=", ">=", "==", "!=":
		return AbsValue{kind: AbsBool, isAnyBool: true}
	case "&&", "||":
		return AbsValue{kind: AbsBool, isAnyBool: true}
	}

	if a.nullMode != NullNone && isArithmeticOp(e.Op) {
		leftNullable := left.canBeNull()
		rightNullable := right.canBeNull()
		if ref, ok := e.Left.(*ast.VarRef); ok && leftNullable {
			leftNullable = !a.isGuardedNotNull(ref.Name)
		}
		if ref, ok := e.Right.(*ast.VarRef); ok && rightNullable {
			rightNullable = !a.isGuardedNotNull(ref.Name)
		}
		if leftNullable || rightNullable {
			a.addNullMessage(e.Line, e.Op)
		}
	}

	return AbsValue{}
}

func isArithmeticOp(op string) bool {
	return op == "+" || op == "-" || op == "*" || op == "/" || op == "%"
}

func (a *Analyzer) addNullMessage(line int, op string) {
	msg := fmt.Sprintf("value may be null when used with operator %s", op)
	switch a.nullMode {
	case NullError:
		a.addError(line, msg)
	case NullWarn:
		a.addWarning(line, msg)
	}
}

func (a *Analyzer) analyzeUnaryExpr(e *ast.UnaryExpr) AbsValue {
	right := a.analyzeExpr(e.Right)
	switch e.Op {
	case "-":
		if right.kind == AbsInt && right.knownInt {
			return knownIntValue(-right.exactInt, right.intType)
		}
		if right.kind == AbsFloat && right.knownFloat {
			return knownFloatValue(-right.exactFloat, right.floatType)
		}
	case "!":
		return AbsValue{kind: AbsBool, isAnyBool: true}
	}
	return AbsValue{}
}

func (a *Analyzer) foldIntArith(left, right AbsValue, op string) AbsValue {
	if left.knownInt && right.knownInt {
		switch op {
		case "+":
			return knownIntValue(left.exactInt+right.exactInt, left.intType)
		case "-":
			return knownIntValue(left.exactInt-right.exactInt, left.intType)
		case "*":
			return knownIntValue(left.exactInt*right.exactInt, left.intType)
		}
	}
	return anyIntValue(left.intType)
}

func (a *Analyzer) checkDivisionByZero(line int, denomExpr ast.Expr, denom AbsValue, op string) {
	if denom.definitelyZero() {
		if op == "/" {
			a.addError(line, "division by zero will occur")
		} else {
			a.addError(line, "modulo by zero will occur")
		}
		return
	}

	guarded := false
	if ref, ok := denomExpr.(*ast.VarRef); ok {
		guarded = a.isGuardedNotZero(ref.Name) || a.isGuardedNotNull(ref.Name)
	}

	if denom.couldBeZero() && !guarded && !denom.canBeNull() {
		if op == "/" {
			a.addWarning(line, "division by zero can occur")
		} else {
			a.addWarning(line, "modulo by zero can occur")
		}
	}
}
