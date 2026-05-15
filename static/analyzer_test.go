package static

import (
	"testing"

	"lang-interpreter/ast"
	"lang-interpreter/lexer"
	"lang-interpreter/parser"
)

func parseProgram(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Logf("parse error: %s", err)
		}
	}
	return program
}

func parseSingleStmt(t *testing.T, input string) ast.Stmt {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	stmt, errs, _ := p.ParseSingleStmt()
	if stmt == nil {
		for _, err := range errs {
			t.Logf("parse error: %s", err)
		}
		t.Fatalf("failed to parse: %s", input)
	}
	return stmt
}

func TestDivisionByZeroDefinite(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false} = 0; var y: int{size: 32} = 10 / x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected division by zero error")
	}
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: division by zero will occur" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'division by zero will occur', got: %v", a.Errors())
	}
}

func TestDivisionByZeroPossible(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: false}; var y: int{size: 32} = 10 / x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for possible division by zero, got: %v", a.Errors())
	}
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for possible division by zero")
	}
	found := false
	for _, w := range a.Warnings() {
		if w == "line 1: division by zero can occur" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'division by zero can occur', got: %v", a.Warnings())
	}
}

func TestDivisionByZeroGuarded(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: false}; if (x != 0) { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for guarded division, got: %v", a.Errors())
	}
	if len(a.Warnings()) > 0 {
		t.Errorf("expected no warnings for guarded division, got: %v", a.Warnings())
	}
}

func TestModuloByZeroDefinite(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false} = 0; var y: int{size: 32} = 10 % x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected modulo by zero error")
	}
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: modulo by zero will occur" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'modulo by zero will occur', got: %v", a.Errors())
	}
}

func TestDivisionByKnownNonZero(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false} = 5; var y: int{size: 32} = 10 / x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for division by known non-zero, got: %v", a.Errors())
	}
	if len(a.Warnings()) > 0 {
		t.Errorf("expected no warnings for division by known non-zero, got: %v", a.Warnings())
	}
}

func TestNoErrorsCleanProgram(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 42; var y: int{size: 32} = x + 1; print(y.toString());`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
	if len(a.Warnings()) > 0 {
		t.Errorf("expected no warnings, got: %v", a.Warnings())
	}
}

func TestTypeMismatchFloatToIntAssignment(t *testing.T) {
	stmt := parseSingleStmt(t, `a = 0.5;`)
	program := &ast.Program{Stmts: []ast.Stmt{
		&ast.VarDecl{
			Name:  "a",
			IType: ast.IntegerType{Size: 64, Signed: true, Nullable: true},
			Line:  1,
		},
		stmt,
	}}
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for float to int assignment")
	}
}

func TestAnalyzerErrorsMethod(t *testing.T) {
	a := New()
	if len(a.Errors()) != 0 {
		t.Errorf("expected empty errors initially")
	}
}

func TestAnalyzerWarningsMethod(t *testing.T) {
	a := New()
	if len(a.Warnings()) != 0 {
		t.Errorf("expected empty warnings initially")
	}
}

func TestDivisionByZeroWithNull(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true} = null; var y: int{size: 32} = 10 / x;`)
	a := New()
	a.Analyze(program)
	for _, e := range a.Errors() {
		if e == "line 1: division by zero will occur" {
			t.Errorf("should NOT report division by zero for null denominator, got errors: %v, warnings: %v", a.Errors(), a.Warnings())
		}
	}
}

func TestUnaryNegationInt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false} = -5; var y: int{size: 32} = 10 / x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for known negated non-zero, got: %v", a.Errors())
	}
	if len(a.Warnings()) > 0 {
		t.Errorf("expected no warnings, got: %v", a.Warnings())
	}
}

func TestUnaryNotBool(t *testing.T) {
	program := parseProgram(t, `var x: bool{nullable: false} = true; var y: bool = !x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestForLoop(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; for (var i: int{size: 32} = 0; i < 10; i++) { x = 5; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestWhileLoop(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; while (true) { x = 5; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestRefDecl(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 5; ref y = x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestFloatLiteralInExpr(t *testing.T) {
	program := parseProgram(t, `var x: float{size: 64} = 3.14;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for float literal, got: %v", a.Errors())
	}
}

func TestBoolLiteralInExpr(t *testing.T) {
	program := parseProgram(t, `var x: bool{nullable: false} = false;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for bool literal, got: %v", a.Errors())
	}
}

func TestStringLiteralInExpr(t *testing.T) {
	program := parseProgram(t, `var x: string = "hello";`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for string literal, got: %v", a.Errors())
	}
}

func TestComparisonExpr(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 5; var y: bool = x > 3;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for comparison, got: %v", a.Errors())
	}
}

func TestLogicalExpr(t *testing.T) {
	program := parseProgram(t, `var x: bool = true; var y: bool = x && false || true;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for logical expr, got: %v", a.Errors())
	}
}

func TestMemberAccessExpr(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 42; print(x.toString());`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for member access, got: %v", a.Errors())
	}
}

func TestIntDefaultValue(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32};`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestFloatDefaultValue(t *testing.T) {
	program := parseProgram(t, `var x: float{size: 64};`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestBoolDefaultValue(t *testing.T) {
	program := parseProgram(t, `var x: bool;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestNullLiteralExpr(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true} = null;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestUndefinedVarRef(t *testing.T) {
	program := parseProgram(t, `print(x.toString());`)
	a := New()
	a.Analyze(program)
}

func TestIncDecStmt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 5; x++; x--;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestElseBranch(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; if (true) { x = 5; } else { x = 10; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestElseSingleStmt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; if (true) { x = 5; } else x = 10;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAssignmentToNonExistentVar(t *testing.T) {
	program := parseProgram(t, `x = 5;`)
	a := New()
	a.Analyze(program)
}

func TestIndexExpr(t *testing.T) {
	program := parseProgram(t, `var a: array{size: 3}<int{size: 32}> = [1, 2, 3]; var x: int{size: 32} = a[0];`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for index expr, got: %v", a.Errors())
	}
}

func TestCopyAndRefExpr(t *testing.T) {
	program := parseProgram(t, `var x: array{size: 3}<int{size: 32}> = [1, 2, 3]; var y: array{size: 3}<int{size: 32}> = copy x; ref z = x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestIsExpr(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 5; var y: int{size: 32} = 5; var z: bool = x is y;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestTypeOfExpr(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 5; print(typeof(x).toString());`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestArrayLitExpr(t *testing.T) {
	program := parseProgram(t, `var a: array{size: 3}<int{size: 32}> = [1, 2, 3];`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestNullAssignmentToNonNullable(t *testing.T) {
	stmt := parseSingleStmt(t, `a = null;`)
	program := &ast.Program{Stmts: []ast.Stmt{
		&ast.VarDecl{Name: "a", IType: ast.IntegerType{Size: 64, Signed: true, Nullable: false}, Line: 1},
		stmt,
	}}
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for null assignment to non-nullable")
	}
}

func TestIntConstantsFolded(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 2 + 3; var y: int{size: 32} = x * 10;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for folded constants, got: %v", a.Errors())
	}
}

func TestUnsignedIntRange(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 8, signed: false} = 5; var y: int{size: 8, signed: false} = 10 / x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestVerifyScopePushPop(t *testing.T) {
	program := parseProgram(t, `if (true) { var x: int{size: 32} = 5; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestDivisionByZeroPossibleFollowingAssignment(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; x = 0; var y: int{size: 32} = 10 / x;`)
	a := New()
	a.Analyze(program)
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: division by zero will occur" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'division by zero will occur', got errors: %v", a.Errors())
	}
}

func TestDivisionByZeroLiteralExpression(t *testing.T) {
	program := parseProgram(t, `var y: int{size: 32} = 10 / 0;`)
	a := New()
	a.Analyze(program)
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: division by zero will occur" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'division by zero will occur', got errors: %v", a.Errors())
	}
}

func TestArrayAndListTypeAnalysis(t *testing.T) {
	program := parseProgram(t, `var a: array{size: 5}<int{size: 32}> = [1, 2, 3, 4, 5]; var b: list{min: 1, max: 10}<int{size: 32}> = [1, 2, 3];`)
	a := New()
	a.Analyze(program)
}

func TestGuardWithNullCheck(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true}; if (x != null) { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
}

func TestGuardWithEqNullElse(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true}; if (x == null) { } else { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
}

func TestBoolNegationUnary(t *testing.T) {
	program := parseProgram(t, `var x: bool = true; var y: bool = !x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestIncDecWithNullVar(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true} = null; x++;`)
	a := New()
	a.Analyze(program)
}

func TestSingleStmtIf(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; if (x != 0) x = 10 / x;`)
	a := New()
	a.Analyze(program)
}

func TestForLoopWithUpdateAssignment(t *testing.T) {
	program := parseProgram(t, `var i: int{size: 32} = 0; for (; i < 10; i = i + 1) { var x: int{size: 32} = 5; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestWhileWithCondition(t *testing.T) {
	program := parseProgram(t, `while (true) { var x: int{size: 32} = 5; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestUnsignedInt32Division(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: false} = 5; var y: int{size: 32, signed: false} = 10 / x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestUnaryNegateFloat(t *testing.T) {
	program := parseProgram(t, `var x: float{size: 64} = -3.14;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestNullDeclarationToNonNullableInt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: false} = null;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for null assignment to non-nullable int")
	}
}

func TestNestedBlocks(t *testing.T) {
	program := parseProgram(t, "var x: int{size: 32} = 5; if (true) { var y: int{size: 32} = 10; if (y > 0) { var z: int{size: 32} = 15; } }")
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestSetNullModeError(t *testing.T) {
	a := New()
	a.SetNullMode(NullError)
	if a.nullMode != NullError {
		t.Errorf("expected NullError")
	}
}

func TestSetNullModeNone(t *testing.T) {
	a := New()
	a.SetNullMode(NullNone)
	if a.nullMode != NullNone {
		t.Errorf("expected NullNone")
	}
}

func TestNullModeErrorAddsError(t *testing.T) {
	input := "var x: int{size: 64, signed: true, nullable: true} = 10; var y: int = 10 / x;"
	program := parseProgram(t, input)
	a := New()
	a.SetNullMode(NullError)
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected errors with NullError mode, got none")
	}
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: value may be null when used with operator /" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected null error, got: %v", a.Errors())
	}
}

func TestForLoopNoInit(t *testing.T) {
	program := parseProgram(t, `var i: int{size: 32} = 0; for (; i < 10; i++) { var x: int{size: 32} = 5; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestModuloByZeroPossible(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: false}; var y: int{size: 32} = 10 % x;`)
	a := New()
	a.Analyze(program)
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for possible modulo by zero")
	}
	found := false
	for _, w := range a.Warnings() {
		if w == "line 1: modulo by zero can occur" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'modulo by zero can occur', got: %v", a.Warnings())
	}
}

func TestKnownBoolLiteral(t *testing.T) {
	program := parseProgram(t, `var x: bool{nullable: false} = true; var y: bool{nullable: false} = false;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

// --- Coverage edge cases ---

func TestRangeIntValue(t *testing.T) {
	v := rangeIntValue(0, 100, ast.IntegerType{Size: 64, Signed: true, Nullable: false})
	if v.minInt != 0 || v.maxInt != 100 {
		t.Errorf("expected range 0-100, got %d-%d", v.minInt, v.maxInt)
	}
}

func TestDefinitelyZeroDefault(t *testing.T) {
	v := AbsValue{kind: AbsFloat}
	if v.definitelyZero() {
		t.Errorf("expected float to not be definitely zero")
	}
}

func TestCouldBeZeroAllPaths(t *testing.T) {
	nullV := nullAbsValue()
	if !nullV.couldBeZero() {
		t.Errorf("expected null to be couldBeZero")
	}
	floatV := AbsValue{kind: AbsFloat}
	if floatV.couldBeZero() {
		t.Errorf("expected float to not be couldBeZero")
	}
}

func TestDefinitelyNotNull(t *testing.T) {
	nonNull := knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: false})
	if !nonNull.definitelyNotNull() {
		t.Errorf("expected non-nullable to be definitelyNotNull")
	}
	nullable := knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: true})
	if nullable.definitelyNotNull() {
		t.Errorf("expected nullable to not be definitelyNotNull")
	}
}

func TestIntRangeUnsigned(t *testing.T) {
	min, max := intRange(ast.IntegerType{Size: 8, Signed: false})
	if min != 0 || max != 255 {
		t.Errorf("expected uint8 range 0-255, got %d-%d", min, max)
	}
	min, max = intRange(ast.IntegerType{Size: 16, Signed: false})
	if min != 0 || max != 65535 {
		t.Errorf("expected uint16 range 0-65535, got %d-%d", min, max)
	}
	min, max = intRange(ast.IntegerType{Size: 32, Signed: false})
	if min != 0 || max != 4294967295 {
		t.Errorf("expected uint32 range 0-4294967295, got %d-%d", min, max)
	}
	min, max = intRange(ast.IntegerType{Size: 64, Signed: false})
	if min != 0 || max != 9223372036854775807 {
		t.Errorf("expected uint64 range 0-MaxInt64, got %d-%d", min, max)
	}
}

func TestIntRangeSigned(t *testing.T) {
	min, max := intRange(ast.IntegerType{Size: 16, Signed: true})
	if min != -32768 || max != 32767 {
		t.Errorf("expected int16 range -32768-32767, got %d-%d", min, max)
	}
	min, max = intRange(ast.IntegerType{Size: 32, Signed: true})
	if min != -2147483648 || max != 2147483647 {
		t.Errorf("expected int32 range -2147483648-2147483647, got %d-%d", min, max)
	}
	min, max = intRange(ast.IntegerType{Size: 64, Signed: true})
	if min != -9223372036854775808 || max != 9223372036854775807 {
		t.Errorf("expected int64 range, got %d-%d", min, max)
	}
}

func TestPopScope(t *testing.T) {
	a := New()
	a.popScope()
}

func TestUnaryExprUnknownOp(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = -5;`)
	a := New()
	a.Analyze(program)
}

func TestStringDeclWithNull(t *testing.T) {
	program := parseProgram(t, `var x: string = "hello";`)
	a := New()
	a.Analyze(program)
}

func TestDivisionByFloatDenominator(t *testing.T) {
	program := parseProgram(t, `var x: float{size: 64} = 5.0; var y: float{size: 64} = 10.0 / x;`)
	a := New()
	a.Analyze(program)
}

func TestModuloByZeroWithRange(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false}; var y: int{size: 32} = 10 % x;`)
	a := New()
	a.Analyze(program)
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for possible modulo by zero with range")
	}
}

func TestBreakAndSkipStmt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; while (x < 5) { break; skip; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestIncDecUndefinedVar(t *testing.T) {
	program := parseProgram(t, `x++;`)
	a := New()
	a.Analyze(program)
}

func TestIncDecNonNullable(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false} = 5; x++; x--;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for inc/dec on non-nullable, got: %v", a.Errors())
	}
}

func TestIsNullableDeclFloatPath(t *testing.T) {
	if !isNullableDecl(&ast.VarDecl{IsFloat: true, FType: ast.FloatType{Nullable: true}}) {
		t.Errorf("expected float nullable decl to be nullable")
	}
}

func TestIsGuardedNotNullReturnTrue(t *testing.T) {
	a := New()
	a.guards = append(a.guards, guardInfo{name: "x", notNull: true})
	if !a.isGuardedNotNull("x") {
		t.Errorf("expected x to be guarded not null")
	}
}

func TestKnownFalseWithElseBlock(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; if (false) { } else { x = 10; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestKnownFalseWithElseSingleStmt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; if (false) { } else x = 10;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestUnknownConditionWithElseSingleStmt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; if (x > 0) { } else x = 10;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestNegateNonKnownInt(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; var y: int{size: 32} = -x;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestIntRangeSigned8(t *testing.T) {
	min, max := intRange(ast.IntegerType{Size: 8, Signed: true})
	if min != -128 || max != 127 {
		t.Errorf("expected int8 range -128-127, got %d-%d", min, max)
	}
}

func TestIsNullableDeclBoolPath(t *testing.T) {
	if !isNullableDecl(&ast.VarDecl{IsBool: true, BType: ast.BoolType{Nullable: true}}) {
		t.Errorf("expected bool nullable decl to be nullable")
	}
}

func TestIsNullableDeclStringPath(t *testing.T) {
	if isNullableDecl(&ast.VarDecl{IsString: true}) {
		t.Errorf("expected string decl to not be nullable")
	}
}

func TestIsNullableDeclFallthrough(t *testing.T) {
	if isNullableDecl(&ast.VarDecl{}) {
		t.Errorf("expected empty decl to not be nullable")
	}
}

func TestTypeDescFromAbsFloat(t *testing.T) {
	program := parseProgram(t, `var x: float{size: 32, nullable: false}; x = null;`)
	a := New()
	a.Analyze(program)
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: cannot assign null to float{size: 32, nullable: false}" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected float null assignment error, got: %v", a.Errors())
	}
}

func TestTypeDescFromAbsBool(t *testing.T) {
	program := parseProgram(t, `var x: bool{nullable: false}; x = null;`)
	a := New()
	a.Analyze(program)
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: cannot assign null to bool{nullable: false}" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected bool null assignment error, got: %v", a.Errors())
	}
}

func TestTypeDescFromAbsString(t *testing.T) {
	program := parseProgram(t, `var x: string = "hello"; x = null;`)
	a := New()
	a.Analyze(program)
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: cannot assign null to string" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected string null assignment error, got: %v", a.Errors())
	}
}

func TestDefaultValueFallback(t *testing.T) {
	a := New()
	val := a.defaultValue(&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 0}})
	if val.kind != AbsInt {
		t.Errorf("expected AbsInt, got %v", val.kind)
	}
}

func TestIsGuardedNotNullFallthrough(t *testing.T) {
	a := New()
	if a.isGuardedNotNull("nonexistent") {
		t.Errorf("expected false for unguarded name")
	}
}

func TestExtractGuardsNotNullCondition(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true, signed: true}; if (x != null) { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
}

func TestExtractGuardsEqNullCondition(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true, signed: true}; if (x == null) { } else { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
}

func TestExtractGuardsNonBinaryExpr(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; if (true) { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for unguarded possible division")
	}
}

func TestTypeDescFromDeclBool(t *testing.T) {
	program := parseProgram(t, `var x: bool{nullable: false} = null;`)
	a := New()
	a.Analyze(program)
	found := false
	for _, e := range a.Errors() {
		if e == "line 1: cannot assign null to bool{nullable: false}" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected bool null decl error, got: %v", a.Errors())
	}
}

func TestTypeDescFromDeclString(t *testing.T) {
	program := parseProgram(t, `var x: string; a = null;`)
	a := New()
	a.Analyze(program)
}

func TestAnalyzeIfStmtElseIfChain(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 0; if (x > 0) { } else if (x < 0) { }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeExprNullLit(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, nullable: true} = null;`)
	a := New()
	a.Analyze(program)
}

func TestAnalyzeStmtExprStmt(t *testing.T) {
	// expression statement (e.g. variable reference as statement)
	program := &ast.Program{Stmts: []ast.Stmt{
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true}, Line: 1},
		&ast.ExprStmt{Expr: &ast.VarRef{Name: "x", Line: 1}, Line: 1},
	}}
	a := New()
	a.Analyze(program)
}

func TestFoldIntArithKnownMul(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false} = 2 * 3;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestFoldIntArithUnknownOp(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32, signed: true, nullable: false}; var y: int{size: 32} = x + 1; var z: int{size: 32} = x - 1; var w: int{size: 32} = x * 1;`)
	a := New()
	a.Analyze(program)
}

func TestUnaryBooleanNotOp(t *testing.T) {
	program := parseProgram(t, `var x: bool = !true;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeExprMemberAccess(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 42; print(x.toString());`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeExprIndexExpr(t *testing.T) {
	program := parseProgram(t, `var a: array{size: 3}<int{size: 32}> = [1, 2, 3]; var x: int{size: 32} = a[0];`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeExprIsExpr(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32} = 5; var y: int{size: 32} = 5; var z: bool = x is y;`)
	a := New()
	a.Analyze(program)
}

func TestAssignNullToNonNullableFloatViaDecl(t *testing.T) {
	program := parseProgram(t, `var x: float{size: 64, nullable: false} = null;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for null to non-nullable float")
	}
}

func TestCompoundAssignExpr(t *testing.T) {
	// an assignment with an Op != "" like x += 1
	program := parseProgram(t, `var x: int{size: 32} = 5; x += 1;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestBlockStmtAsStmt(t *testing.T) {
	a := New()
	a.analyzeStmt(&ast.BlockStmt{Stmts: []ast.Stmt{
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true}, Line: 1},
	}, Line: 1})
}

func TestTypeDescFromDeclStringDirect(t *testing.T) {
	result := typeDescFromDecl(&ast.VarDecl{Name: "x", IsString: true, SType: ast.StringType{}})
	if result != "string" {
		t.Errorf("expected 'string', got %s", result)
	}
}

func TestIfStmtElseNonBlockNonIf(t *testing.T) {
	a := New()
	a.analyzeStmt(&ast.IfStmt{
		Condition: &ast.BoolLit{Value: false, BType: ast.BoolType{Nullable: true}, Line: 1},
		Then:      &ast.BlockStmt{Stmts: []ast.Stmt{}, Line: 1},
		Else:      &ast.ExprStmt{Expr: &ast.IntegerLit{Value: 1, IType: ast.IntegerType{Size: 64, Signed: true, Nullable: true}, Untyped: true, Line: 1}, Line: 1},
		Line:      1,
	})
}

func TestExtractGuardsNotNullRightSide(t *testing.T) {
	a := New()
	// guard: null != x
	conds := []ast.Expr{
		&ast.BinaryExpr{Left: &ast.NullLit{Line: 1}, Op: "!=", Right: &ast.VarRef{Name: "x", Line: 1}, Line: 1},
		&ast.BinaryExpr{Left: &ast.NullLit{Line: 1}, Op: "==", Right: &ast.VarRef{Name: "y", Line: 1}, Line: 1},
		&ast.BinaryExpr{Left: &ast.IntegerLit{Value: 0, IType: ast.IntegerType{Size: 64, Signed: true, Nullable: true}, Untyped: true, Line: 1}, Op: "!=", Right: &ast.VarRef{Name: "z", Line: 1}, Line: 1},
	}
	for _, cond := range conds {
		_ = a.extractGuards(cond)
	}
}

func TestExtractGuardsEqNullRightSide(t *testing.T) {
	a := New()
	cond := &ast.BinaryExpr{Left: &ast.NullLit{Line: 1}, Op: "==", Right: &ast.VarRef{Name: "x", Line: 1}, Line: 1}
	_ = a.extractGuards(cond)
}

func TestAnalyzeExprEdgeCases(t *testing.T) {
	a := New()
	// VarRef not found
	a.analyzeExpr(&ast.VarRef{Name: "nonexistent", Line: 1})
	// MemberAccess
	a.analyzeExpr(&ast.MemberAccess{Object: &ast.VarRef{Name: "x", Line: 1}, Member: "toString", Args: nil, Line: 1})
	// TypeOfExpr
	a.analyzeExpr(&ast.TypeOfExpr{Expr: &ast.VarRef{Name: "x", Line: 1}, Line: 1})
	// CopyExpr
	a.analyzeExpr(&ast.CopyExpr{Right: &ast.VarRef{Name: "x", Line: 1}, Line: 1})
	// RefExpr
	a.analyzeExpr(&ast.RefExpr{Right: &ast.VarRef{Name: "x", Line: 1}, Line: 1})
	// IsExpr
	a.analyzeExpr(&ast.IsExpr{Left: &ast.VarRef{Name: "x", Line: 1}, Right: &ast.VarRef{Name: "y", Line: 1}, Line: 1})
	// NullLit
	a.analyzeExpr(&ast.NullLit{Line: 1})
}

func TestAnalyzeUnaryNegateKnownInt(t *testing.T) {
	a := New()
	a.analyzeExpr(&ast.UnaryExpr{Op: "-", Right: &ast.IntegerLit{Value: 5, IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}, Line: 1}, Line: 1})
}

func TestAnalyzeUnaryNegateKnownFloat(t *testing.T) {
	a := New()
	a.analyzeExpr(&ast.UnaryExpr{Op: "-", Right: &ast.FloatLit{Value: 3.14, FType: ast.FloatType{Size: 64, Nullable: true}, Untyped: true, Line: 1}, Line: 1})
}

func TestAnalyzeUnaryNegateUnknownInt(t *testing.T) {
	a := New()
	a.env["x"] = anyIntValue(ast.IntegerType{Size: 32, Signed: true, Nullable: true})
	a.analyzeExpr(&ast.UnaryExpr{Op: "-", Right: &ast.VarRef{Name: "x", Line: 1}, Line: 1})
}

func TestFoldIntArithSubtract(t *testing.T) {
	a := New()
	left := knownIntValue(10, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	right := knownIntValue(3, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	result := a.foldIntArith(left, right, "-")
	if result.knownInt && result.exactInt != 7 {
		t.Errorf("expected 7, got %d", result.exactInt)
	}
}

func TestFoldIntArithMul(t *testing.T) {
	a := New()
	left := knownIntValue(4, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	right := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	result := a.foldIntArith(left, right, "*")
	if result.knownInt && result.exactInt != 20 {
		t.Errorf("expected 20, got %d", result.exactInt)
	}
}

func TestIsGuardedNotZeroNotFound(t *testing.T) {
	a := New()
	if a.isGuardedNotZero("nonexistent") {
		t.Errorf("expected false")
	}
}

func TestTypeDescFromAbsDefault(t *testing.T) {
	result := typeDescFromAbs(AbsValue{kind: AbsArray})
	if result != "unknown" {
		t.Errorf("expected 'unknown', got %s", result)
	}
}

func TestAnalyzeExprFallthrough(t *testing.T) {
	a := New()
	result := a.analyzeExpr(nil)
	if result.kind != 0 || result.isAnyInt {
		t.Errorf("expected empty AbsValue")
	}
}

func TestGuardEqualsLiteralLeft(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; if (x == 5) { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
	if len(a.Warnings()) > 0 {
		t.Errorf("expected no warnings, got: %v", a.Warnings())
	}
}

func TestGuardEqualsLiteralRight(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; if (5 == x) { var y: int{size: 32} = 10 / x; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
	if len(a.Warnings()) > 0 {
		t.Errorf("expected no warnings, got: %v", a.Warnings())
	}
}

func TestNullGuardSuppressesNullWarningBinaryExprLeft(t *testing.T) {
	program := parseProgram(t, `var x: int{size: 32}; if (x != null) { var y: int{size: 32} = x + true; }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
	if len(a.Warnings()) > 0 {
		t.Errorf("expected no warnings, got: %v", a.Warnings())
	}
}

func TestAnalyzeForInStmt(t *testing.T) {
	program := parseProgram(t, `
var arr: array{size: 3}<int> = [1, 2, 3];
for (var x in arr) {
	print(x.toString());
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeForAtStmt(t *testing.T) {
	program := parseProgram(t, `
var arr: array{size: 3}<int> = [1, 2, 3];
for (var i at arr) {
	print(arr[i].toString());
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeForOfStmt(t *testing.T) {
	program := parseProgram(t, `
var arr: array{size: 3}<int> = [1, 2, 3];
for (var i, v of arr) {
	print((i).toString());
	print((v).toString());
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}
