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
	AbsUnion
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
	unionTypes []ast.Type
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
	funcs      map[string][]*ast.FuncDecl
	insideFunc *ast.FuncDecl
}

func New() *Analyzer {
	return &Analyzer{
		env:      make(map[string]AbsValue),
		nullMode: NullWarn,
		funcs:    make(map[string][]*ast.FuncDecl),
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
	for _, overloads := range a.funcs {
		for _, fn := range overloads {
			a.analyzeFunc(fn)
		}
	}
}

func (a *Analyzer) collectFunc(stmt ast.Stmt) {
	if fn, ok := stmt.(*ast.FuncDecl); ok {
		a.funcs[fn.Name] = append(a.funcs[fn.Name], fn)
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
	case ast.UnionType:
		return AbsValue{kind: AbsUnion, unionTypes: typ.Types}
	}
	return AbsValue{}
}

func (a *Analyzer) addError(line int, msg string) {
	full := fmt.Sprintf("line %d: %s", line, msg)
	for _, e := range a.errors {
		if e == full {
			return
		}
	}
	a.errors = append(a.errors, full)
}

func (a *Analyzer) addWarning(line int, msg string) {
	full := fmt.Sprintf("line %d: %s", line, msg)
	for _, w := range a.warnings {
		if w == full {
			return
		}
	}
	a.warnings = append(a.warnings, full)
}

func (a *Analyzer) analyzeStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		a.analyzeVarDecl(s)
	case *ast.Assignment:
		a.analyzeAssignment(s)
	case *ast.RefDecl:
		a.analyzeRefDecl(s)
	case *ast.SwitchStmt:
		a.analyzeSwitchStmt(s)
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
	// Use declared type to improve kind resolution (e.g., AbsArray -> AbsList)
	if d.Type != nil && val.kind == AbsArray {
		if _, isList := d.Type.(ast.ListType); isList {
			val = absValueFromType(d.Type)
		}
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
	if d.IsUnion {
		for _, t := range d.UnionType.Types {
			if typeIsNullable(t) {
				return true
			}
		}
		return false
	}
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

func typeIsNullable(t ast.Type) bool {
	switch typ := t.(type) {
	case ast.IntegerType:
		return typ.Nullable
	case ast.FloatType:
		return typ.Nullable
	case ast.BoolType:
		return typ.Nullable
	case ast.UnionType:
		for _, m := range typ.Types {
			if typeIsNullable(m) {
				return true
			}
		}
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
	case AbsUnion:
		s := ""
		for i, t := range v.unionTypes {
			if i > 0 {
				s += " | "
			}
			s += typeDescFromType(t)
		}
		return s
	default:
		return "unknown"
	}
}

func typeDescFromType(t ast.Type) string {
	switch typ := t.(type) {
	case ast.IntegerType:
		return fmt.Sprintf("int{size: %d, signed: %v, nullable: %v}", typ.Size, typ.Signed, typ.Nullable)
	case ast.FloatType:
		return fmt.Sprintf("float{size: %d, nullable: %v}", typ.Size, typ.Nullable)
	case ast.BoolType:
		return fmt.Sprintf("bool{nullable: %v}", typ.Nullable)
	case ast.StringType:
		return "string"
	default:
		return fmt.Sprintf("%T", t)
	}
}

func typeDescFromDecl(d *ast.VarDecl) string {
	if d.IsUnion {
		s := ""
		for i, t := range d.UnionType.Types {
			if i > 0 {
				s += " | "
			}
			s += typeDescFromType(t)
		}
		return s
	}
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
	if d.IsUnion {
		return AbsValue{kind: AbsUnion, unionTypes: d.UnionType.Types}
	}
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

func absValueFromType(t ast.Type) AbsValue {
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
		return AbsValue{kind: AbsArray, arrayType: typ}
	case ast.ListType:
		return AbsValue{kind: AbsList, listType: typ}
	case ast.UnionType:
		return AbsValue{kind: AbsUnion, unionTypes: typ.Types}
	}
	return AbsValue{}
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

func (a *Analyzer) analyzeSwitchStmt(s *ast.SwitchStmt) {
	switchExpr := a.analyzeExpr(s.Value)

	for _, c := range s.Cases {
		if c.Default {
			a.analyzeBlock(c.Body.Stmts)
			continue
		}

		if c.Op == "==" || c.Op == "!=" || c.Op == "" {
			a.analyzeExpr(c.Value)
		} else {
			caseVal := a.analyzeExpr(c.Value)
			if caseVal.kind == AbsBool {
				a.addError(c.Line, "relational case expression cannot be bool")
			}
			if switchExpr.kind != AbsInt && switchExpr.kind != AbsFloat && caseVal.kind != AbsInt && caseVal.kind != AbsFloat {
				a.addWarning(c.Line, "relational comparison may not work with non-numeric types")
			}
		}

		a.analyzeBlock(c.Body.Stmts)
	}
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
		a.analyzeExpr(e.Object)
		return AbsValue{}
	case *ast.TypeOfExpr:
		a.analyzeExpr(e.Expr)
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
		overloads, exists := a.funcs[ref.Name]
		if !exists || len(overloads) == 0 {
			a.addError(e.Line, "undefined function: "+ref.Name)
			return AbsValue{}
		}

		argTypes := make([]AbsValue, len(e.Args))
		for i, arg := range e.Args {
			argTypes[i] = a.analyzeExpr(arg)
		}

		// Check for ambiguous null literal calls
		if hasNullLiteralArg(e.Args, argTypes) && countMatchingOverloads(overloads, argTypes) > 1 {
			a.addError(e.Line, fmt.Sprintf("ambiguous call to %s: multiple overloads accept null", ref.Name))
			return AbsValue{}
		}

		decl := resolveOverloadStatic(overloads, argTypes)
		if decl == nil {
			a.addError(e.Line, fmt.Sprintf("no matching overload for %s", ref.Name))
			return AbsValue{}
		}

		if decl.ReturnType != nil {
			return defaultValueForParam(decl.ReturnType)
		}
	}
	return AbsValue{}
}

func hasNullLiteralArg(args []ast.Expr, argTypes []AbsValue) bool {
	for i, arg := range args {
		if _, ok := arg.(*ast.NullLit); ok {
			_ = argTypes[i]
			return true
		}
	}
	return false
}

func countMatchingOverloads(overloads []*ast.FuncDecl, argTypes []AbsValue) int {
	count := 0
	for _, fn := range overloads {
		if len(fn.Parameters) != len(argTypes) {
			continue
		}
		if argsExactMatchStatic(fn.Parameters, argTypes) {
			count++
		} else if argsMatchCategoryStatic(fn.Parameters, argTypes, true) {
			count++
		} else if argsMatchCategoryStatic(fn.Parameters, argTypes, false) {
			count++
		}
	}
	return count
}

func resolveOverloadStatic(overloads []*ast.FuncDecl, argTypes []AbsValue) *ast.FuncDecl {
	for _, fn := range overloads {
		if len(fn.Parameters) != len(argTypes) {
			continue
		}
		if argsExactMatchStatic(fn.Parameters, argTypes) {
			return fn
		}
	}
	for _, fn := range overloads {
		if len(fn.Parameters) != len(argTypes) {
			continue
		}
		if argsMatchCategoryStatic(fn.Parameters, argTypes, true) {
			return fn
		}
	}
	for _, fn := range overloads {
		if len(fn.Parameters) != len(argTypes) {
			continue
		}
		if argsMatchCategoryStatic(fn.Parameters, argTypes, false) {
			return fn
		}
	}
	return nil
}

func argsExactMatchStatic(params []ast.Param, argTypes []AbsValue) bool {
	for i, p := range params {
		if !argExactTypeMatchStatic(p.Type, argTypes[i]) {
			return false
		}
	}
	return true
}

func argExactTypeMatchStatic(paramType ast.Type, arg AbsValue) bool {
	if arg.kind == AbsNull || arg.definitelyNull {
		return typeIsNullable(paramType)
	}
	switch pt := paramType.(type) {
	case ast.IntegerType:
		return arg.kind == AbsInt
	case ast.FloatType:
		return arg.kind == AbsFloat
	case ast.BoolType:
		return arg.kind == AbsBool
	case ast.StringType:
		return arg.kind == AbsString
	case ast.ArrayType:
		return arg.kind == AbsArray
	case ast.ListType:
		return arg.kind == AbsList
	case ast.UnionType:
		for _, t := range pt.Types {
			if argExactTypeMatchStatic(t, arg) {
				return true
			}
		}
		return false
	}
	return false
}

func argsMatchCategoryStatic(params []ast.Param, argTypes []AbsValue, sameCategory bool) bool {
	for i, p := range params {
		if !argMatchCategoryStatic(p.Type, argTypes[i], sameCategory) {
			return false
		}
	}
	return true
}

func argMatchCategoryStatic(paramType ast.Type, arg AbsValue, sameCategory bool) bool {
	if arg.kind == AbsNull || arg.definitelyNull {
		return typeIsNullable(paramType)
	}
	switch pt := paramType.(type) {
	case ast.IntegerType:
		if sameCategory {
			return arg.kind == AbsInt
		}
		return arg.kind == AbsInt || arg.kind == AbsFloat
	case ast.FloatType:
		if sameCategory {
			return arg.kind == AbsFloat
		}
		return arg.kind == AbsFloat || arg.kind == AbsInt
	case ast.BoolType:
		return arg.kind == AbsBool
	case ast.StringType:
		return arg.kind == AbsString
	case ast.ArrayType:
		return arg.kind == AbsArray
	case ast.ListType:
		return arg.kind == AbsList
	case ast.UnionType:
		for _, t := range pt.Types {
			if argMatchCategoryStatic(t, arg, sameCategory) {
				return true
			}
		}
		return false
	}
	return false
}

func (a *Analyzer) analyzeBinaryExpr(e *ast.BinaryExpr) AbsValue {
	left := a.analyzeExpr(e.Left)
	right := a.analyzeExpr(e.Right)

	switch e.Op {
	case "/", "%":
		a.checkDivisionByZero(e.Line, e.Right, right, e.Op)
		a.checkNullArith(e, left, right)
	case "+", "-", "*":
		a.checkNullArith(e, left, right)
		if left.kind == AbsInt && right.kind == AbsInt {
			return a.foldIntArith(left, right, e.Op)
		}
	case "<", ">", "<=", ">=", "==", "!=":
		return AbsValue{kind: AbsBool, isAnyBool: true}
	case "&&", "||":
		return AbsValue{kind: AbsBool, isAnyBool: true}
	}

	if a.nullMode != NullNone && isArithmeticOp(e.Op) {
		a.checkNullArith(e, left, right)
	}

	return AbsValue{}
}

func (a *Analyzer) checkNullArith(e *ast.BinaryExpr, left, right AbsValue) {
	if a.nullMode == NullNone {
		return
	}
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
