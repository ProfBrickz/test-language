package static

import (
	"strings"
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
	program := parseProgram(t, `var x: int{size: 32, nullable: false} = 42; var y: int{size: 32, nullable: false} = x + 1; print(y.toString());`)
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

func TestAnalyzeStringDeclWithNullAssign(t *testing.T) {
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

func TestAnalyzeFuncValid(t *testing.T) {
	program := parseProgram(t, `
function add(x: int, y: int): int {
	return x + y;
}
var r: int{size: 32} = add(1, 2);
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeFuncVoidValid(t *testing.T) {
	program := parseProgram(t, `
function greet(name: string) {
	print(name);
}
greet("hi");
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeFuncReturnOutsideError(t *testing.T) {
	program := parseProgram(t, `return;`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for return outside function")
	}
}

func TestAnalyzeFuncVoidReturnValueError(t *testing.T) {
	program := parseProgram(t, `
function foo() {
	return 42;
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for return with value in void function")
	}
}

func TestAnalyzeFuncTypedBareReturnError(t *testing.T) {
	program := parseProgram(t, `
function foo(): int {
	return;
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for bare return in typed function")
	}
}

func TestAnalyzeFuncCallUndefinedError(t *testing.T) {
	program := parseProgram(t, `foo();`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for undefined function")
	}
}

func TestAnalyzeFuncCallArgCountError(t *testing.T) {
	program := parseProgram(t, `
function add(x: int, y: int): int {
	return x + y;
}
add(1);
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for wrong argument count")
	}
}

func TestAnalyzeFuncOverload(t *testing.T) {
	program := parseProgram(t, `
function foo(x: int): int { return x; }
function foo(x: float): int { return 1; }
foo(5);
foo(1.0);
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeFuncRecursion(t *testing.T) {
	program := parseProgram(t, `
function factorial(n: int): int {
	if (n <= 1) {
		return 1;
	}
	return n * factorial(n - 1);
}
print(factorial(5).toString());
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeFuncFloatParam(t *testing.T) {
	program := parseProgram(t, `
function foo(x: float): int {
	return 1;
}
foo(1.0);
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeFuncBoolParam(t *testing.T) {
	program := parseProgram(t, `
function foo(x: bool): int {
	return 1;
}
foo(true);
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeFuncArrayParam(t *testing.T) {
	program := parseProgram(t, `
function foo(x: array<int{size: 32}>): int {
	return 1;
}
var a: array{size: 3}<int{size: 32}> = [1, 2, 3];
foo(a);
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeFuncListParam(t *testing.T) {
	program := parseProgram(t, `
function foo(x: list<int{size: 32}>): int {
	return 1;
}
var a: list<int{size: 32}> = [1, 2, 3];
foo(a);
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

type unknownType struct{}

func (unknownType) Kind() string { return "custom" }

func TestDefaultValueForParamUnknownType(t *testing.T) {
	result := defaultValueForParam(unknownType{})
	if result.unionTypes != nil {
		t.Errorf("expected empty AbsValue for unknown type, got %+v", result)
	}
}

func TestAnalyzeFuncHoisting(t *testing.T) {
	program := parseProgram(t, `
function caller(): int {
	return callee();
}
function callee(): int {
	return 99;
}
caller();
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeSwitchBasic(t *testing.T) {
	program := parseProgram(t, `
var x: int{size: 32} = 5;
switch (x) {
	case (1) { print(10); }
	case (2) { print(20); }
	default { print(99); }
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeSwitchRelOps(t *testing.T) {
	program := parseProgram(t, `
var x: int{size: 32} = 5;
switch (x) {
	case (< 10) { print(1); }
	case (> 100) { print(2); }
	case (<= 50) { print(3); }
	case (>= 200) { print(4); }
	case (!= 0) { print(5); }
	case (== 5) { print(6); }
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeSwitchOnlyDefault(t *testing.T) {
	program := parseProgram(t, `
var x: int{size: 32} = 5;
switch (x) {
	default { print(99); }
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeSwitchEmpty(t *testing.T) {
	program := parseProgram(t, `switch (x) { }`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeSwitchNested(t *testing.T) {
	program := parseProgram(t, `
var x: int{size: 32} = 5;
switch (x) {
	case (1) {
		var y: int{size: 32} = 10;
		switch (y) {
			case (10) { print(100); }
			default { print(200); }
		}
	}
	default { print(99); }
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeSwitchString(t *testing.T) {
	program := parseProgram(t, `
var x: string{size: 10} = "hello";
switch (x) {
	case ("hello") { print(1); }
	case ("world") { print(2); }
	default { print(0); }
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got: %v", a.Errors())
	}
}

func TestAnalyzeSwitchRelOpWithBool(t *testing.T) {
	program := parseProgram(t, `
var x: int{size: 32} = 5;
switch (x) {
	case (< true) { print(1); }
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for relational case with bool expression")
	}
}

func TestAnalyzeSwitchRelOpNonNumeric(t *testing.T) {
	program := parseProgram(t, `
var x: string{size: 10} = "hello";
switch (x) {
	case (< "world") { print(1); }
}
`)
	a := New()
	a.Analyze(program)
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for relational comparison with non-numeric types")
	}
}

func TestTypeIsNullable(t *testing.T) {
	tests := []struct {
		name string
		typ  ast.Type
		want bool
	}{
		{"nullable int", ast.IntegerType{Nullable: true}, true},
		{"non-nullable int", ast.IntegerType{Nullable: false}, false},
		{"nullable float", ast.FloatType{Nullable: true}, true},
		{"non-nullable float", ast.FloatType{Nullable: false}, false},
		{"nullable bool", ast.BoolType{Nullable: true}, true},
		{"non-nullable bool", ast.BoolType{Nullable: false}, false},
		{"union all nullable", ast.UnionType{Types: []ast.Type{ast.IntegerType{Nullable: true}, ast.FloatType{Nullable: true}}}, true},
		{"union partially nullable", ast.UnionType{Types: []ast.Type{ast.IntegerType{Nullable: true}, ast.FloatType{Nullable: false}}}, true},
		{"union none nullable", ast.UnionType{Types: []ast.Type{ast.IntegerType{Nullable: false}, ast.FloatType{Nullable: false}}}, false},
		{"union empty", ast.UnionType{Types: nil}, false},
		{"string", ast.StringType{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeIsNullable(tt.typ); got != tt.want {
				t.Errorf("typeIsNullable(%v) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}

func TestTypeDescFromType(t *testing.T) {
	tests := []struct {
		name string
		typ  ast.Type
		want string
	}{
		{"int", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "int{size: 32, signed: true, nullable: false}"},
		{"int nullable", ast.IntegerType{Size: 8, Signed: false, Nullable: true}, "int{size: 8, signed: false, nullable: true}"},
		{"float", ast.FloatType{Size: 64, Nullable: true}, "float{size: 64, nullable: true}"},
		{"float non-nullable", ast.FloatType{Size: 32, Nullable: false}, "float{size: 32, nullable: false}"},
		{"bool", ast.BoolType{Nullable: false}, "bool{nullable: false}"},
		{"bool nullable", ast.BoolType{Nullable: true}, "bool{nullable: true}"},
		{"string", ast.StringType{}, "string"},
		{"union", ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: true}}}, "int{size: 32, signed: true, nullable: false} | float{size: 64, nullable: true}"},
		{"union single", ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 64, Signed: true, Nullable: false}}}, "int{size: 64, signed: true, nullable: false}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeDescFromType(tt.typ); got != tt.want {
				t.Errorf("typeDescFromType(%v) = %q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

func TestUnionDescFromTypes(t *testing.T) {
	tests := []struct {
		name  string
		types []ast.Type
		want  string
	}{
		{"nil", nil, ""},
		{"empty", []ast.Type{}, ""},
		{"single", []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}, "int{size: 32, signed: true, nullable: false}"},
		{"multiple", []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: true}}, "int{size: 32, signed: true, nullable: false} | float{size: 64, nullable: true}"},
		{"three", []ast.Type{ast.IntegerType{Size: 8, Signed: false, Nullable: true}, ast.BoolType{Nullable: false}, ast.FloatType{Size: 32, Nullable: true}}, "int{size: 8, signed: false, nullable: true} | bool{nullable: false} | float{size: 32, nullable: true}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unionDescFromTypes(tt.types); got != tt.want {
				t.Errorf("unionDescFromTypes(%v) = %q, want %q", tt.types, got, tt.want)
			}
		})
	}
}

func TestDeclTypeFromDecl(t *testing.T) {
	t.Run("union", func(t *testing.T) {
		d := &ast.VarDecl{IsUnion: true, UnionType: ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}}
		got := declTypeFromDecl(d)
		if _, ok := got.(ast.UnionType); !ok {
			t.Errorf("expected UnionType, got %T", got)
		}
	})
	t.Run("float", func(t *testing.T) {
		d := &ast.VarDecl{IsFloat: true, FType: ast.FloatType{Size: 64, Nullable: false}}
		got := declTypeFromDecl(d)
		ft, ok := got.(ast.FloatType)
		if !ok || ft.Size != 64 || ft.Nullable != false {
			t.Errorf("expected FloatType{64, false}, got %v", got)
		}
	})
	t.Run("bool", func(t *testing.T) {
		d := &ast.VarDecl{IsBool: true, BType: ast.BoolType{Nullable: true}}
		got := declTypeFromDecl(d)
		bt, ok := got.(ast.BoolType)
		if !ok || bt.Nullable != true {
			t.Errorf("expected BoolType{nullable: true}, got %v", got)
		}
	})
	t.Run("string", func(t *testing.T) {
		d := &ast.VarDecl{IsString: true, SType: ast.StringType{Size: 10}}
		got := declTypeFromDecl(d)
		if _, ok := got.(ast.StringType); !ok {
			t.Errorf("expected StringType, got %T", got)
		}
	})
	t.Run("int default", func(t *testing.T) {
		d := &ast.VarDecl{IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}}
		got := declTypeFromDecl(d)
		it, ok := got.(ast.IntegerType)
		if !ok || it.Size != 32 || it.Signed != true || it.Nullable != false {
			t.Errorf("expected IntegerType{32, true, false}, got %v", got)
		}
	})
}

func TestCanImplicitConvertStatic(t *testing.T) {
	tests := []struct {
		name       string
		srcInt     ast.IntegerType
		srcFloat   ast.FloatType
		srcIsFloat bool
		dstInt     ast.IntegerType
		dstFloat   ast.FloatType
		dstIsFloat bool
		want       bool
	}{
		{"float32 to float64", ast.IntegerType{}, ast.FloatType{Size: 32}, true, ast.IntegerType{}, ast.FloatType{Size: 64}, true, true},
		{"float64 to float32", ast.IntegerType{}, ast.FloatType{Size: 64}, true, ast.IntegerType{}, ast.FloatType{Size: 32}, true, false},
		{"float32 to float32", ast.IntegerType{}, ast.FloatType{Size: 32}, true, ast.IntegerType{}, ast.FloatType{Size: 32}, true, true},
		{"int8 to float16", ast.IntegerType{Size: 8, Signed: true}, ast.FloatType{}, false, ast.IntegerType{}, ast.FloatType{Size: 16}, true, true},
		{"int8 to float32", ast.IntegerType{Size: 8, Signed: true}, ast.FloatType{}, false, ast.IntegerType{}, ast.FloatType{Size: 32}, true, true},
		{"int16 to float32", ast.IntegerType{Size: 16, Signed: true}, ast.FloatType{}, false, ast.IntegerType{}, ast.FloatType{Size: 32}, true, true},
		{"int32 to float64", ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, ast.IntegerType{}, ast.FloatType{Size: 64}, true, true},
		{"int32 to float32", ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, ast.IntegerType{}, ast.FloatType{Size: 32}, true, false},
		{"int64 to float64", ast.IntegerType{Size: 64, Signed: true}, ast.FloatType{}, false, ast.IntegerType{}, ast.FloatType{Size: 64}, true, false},
		{"float to int always false", ast.IntegerType{}, ast.FloatType{Size: 64}, true, ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, false},
		{"int32 to int32 same", ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, true},
		{"int32 to int64 both signed", ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, ast.IntegerType{Size: 64, Signed: true}, ast.FloatType{}, false, true},
		{"int64 to int32 src larger", ast.IntegerType{Size: 64, Signed: true}, ast.FloatType{}, false, ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, false},
		{"uint32 to uint64", ast.IntegerType{Size: 32, Signed: false}, ast.FloatType{}, false, ast.IntegerType{Size: 64, Signed: false}, ast.FloatType{}, false, true},
		{"uint32 to int64", ast.IntegerType{Size: 32, Signed: false}, ast.FloatType{}, false, ast.IntegerType{Size: 64, Signed: true}, ast.FloatType{}, false, true},
		{"int32 to uint64", ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, ast.IntegerType{Size: 64, Signed: false}, ast.FloatType{}, false, false},
		{"int32 to uint32 diff signed same size", ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, ast.IntegerType{Size: 32, Signed: false}, ast.FloatType{}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canImplicitConvertStatic(tt.srcInt, tt.srcFloat, tt.srcIsFloat, tt.dstInt, tt.dstFloat, tt.dstIsFloat); got != tt.want {
				t.Errorf("canImplicitConvertStatic(%v, %v, %v, %v, %v, %v) = %v, want %v",
					tt.srcInt, tt.srcFloat, tt.srcIsFloat, tt.dstInt, tt.dstFloat, tt.dstIsFloat, got, tt.want)
			}
		})
	}
}

func TestArgMatchCategoryStatic(t *testing.T) {
	tests := []struct {
		name         string
		paramType    ast.Type
		arg          AbsValue
		sameCategory bool
		want         bool
	}{
		// Null arg
		{"null to nullable int", ast.IntegerType{Size: 32, Signed: true, Nullable: true}, nullAbsValue(), false, true},
		{"null to non-nullable int", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, nullAbsValue(), false, false},
		{"null to nullable float", ast.FloatType{Nullable: true}, nullAbsValue(), false, true},
		{"null to non-nullable bool", ast.BoolType{Nullable: false}, nullAbsValue(), false, false},
		{"null to nullable bool", ast.BoolType{Nullable: true}, nullAbsValue(), false, true},

		// IntegerType param with AbsInt arg
		{"int to same int", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, true},
		{"int to larger int", ast.IntegerType{Size: 64, Signed: true, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, true},
		{"int to smaller int", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: false}), false, false},
		{"int to different signed", ast.IntegerType{Size: 32, Signed: false, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, false},
		{"untyped int to any int", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, func() AbsValue { v := knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: false}); v.untyped = true; return v }(), false, true},

		// FloatType param with AbsFloat arg
		{"float to same float", ast.FloatType{Size: 64, Nullable: false}, knownFloatValue(1.0, ast.FloatType{Size: 64, Nullable: false}), false, true},
		{"float to larger float", ast.FloatType{Size: 64, Nullable: false}, knownFloatValue(1.0, ast.FloatType{Size: 32, Nullable: false}), false, true},
		{"float to smaller float", ast.FloatType{Size: 32, Nullable: false}, knownFloatValue(1.0, ast.FloatType{Size: 64, Nullable: false}), false, false},
		{"untyped float to any float", ast.FloatType{Size: 64, Nullable: false}, func() AbsValue { v := knownFloatValue(1.0, ast.FloatType{Size: 32, Nullable: false}); v.untyped = true; return v }(), false, true},

		// FloatType param with AbsInt arg (cross-category)
		{"int to float cross category", ast.FloatType{Size: 64, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, true},
		{"int to float same category false", ast.FloatType{Size: 64, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), true, false},
		{"int64 to float64 cannot convert", ast.FloatType{Size: 64, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: false}), false, false},
		{"untyped int to float cross category", ast.FloatType{Size: 64, Nullable: false}, func() AbsValue { v := knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: false}); v.untyped = true; return v }(), false, true},

		// FloatType param with AbsInt arg not matching - sameCategory=true case
		{"int arg same category with float param", ast.FloatType{Size: 64, Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), true, false},

		// BoolType
		{"bool param bool arg", ast.BoolType{Nullable: false}, knownBoolValue(true, ast.BoolType{Nullable: false}), false, true},
		{"bool param int arg", ast.BoolType{Nullable: false}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, false},

		// StringType
		{"string param string arg", ast.StringType{}, AbsValue{kind: AbsString, isAnyStr: true}, false, true},
		{"string param int arg", ast.StringType{}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, false},

		// ArrayType and ListType
		{"array param array arg", ast.ArrayType{}, AbsValue{kind: AbsArray}, false, true},
		{"list param list arg", ast.ListType{}, AbsValue{kind: AbsList}, false, true},

		// UnionType
		{"union param matching member", ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, true},
		{"union param no matching member", ast.UnionType{Types: []ast.Type{ast.BoolType{Nullable: false}}}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := argMatchCategoryStatic(tt.paramType, tt.arg, tt.sameCategory); got != tt.want {
				t.Errorf("argMatchCategoryStatic(%v, %v, %v) = %v, want %v", tt.paramType, tt.arg, tt.sameCategory, got, tt.want)
			}
		})
	}
}

func TestArgsMatchCategoryStatic(t *testing.T) {
	intParam := ast.Param{Name: "x", Type: ast.IntegerType{Size: 32, Signed: true, Nullable: false}}
	floatParam := ast.Param{Name: "y", Type: ast.FloatType{Size: 64, Nullable: false}}

	t.Run("both match same category", func(t *testing.T) {
		params := []ast.Param{intParam, intParam}
		args := []AbsValue{
			knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}),
			knownIntValue(10, ast.IntegerType{Size: 32, Signed: true, Nullable: false}),
		}
		if !argsMatchCategoryStatic(params, args, false) {
			t.Errorf("expected both to match")
		}
	})

	t.Run("one fails", func(t *testing.T) {
		params := []ast.Param{intParam, intParam}
		args := []AbsValue{
			knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: false}), // int64->int32 fails
			knownIntValue(10, ast.IntegerType{Size: 64, Signed: true, Nullable: false}),
		}
		if argsMatchCategoryStatic(params, args, false) {
			t.Errorf("expected false when one arg does not match")
		}
	})

	t.Run("cross category", func(t *testing.T) {
		params := []ast.Param{intParam, floatParam}
		args := []AbsValue{
			knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}),
			knownIntValue(10, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), // int->float cross category
		}
		// sameCategory=false: int matches int, int matches float cross-category
		if !argsMatchCategoryStatic(params, args, false) {
			t.Errorf("expected both to match with cross category")
		}
	})

	t.Run("cross category rejected with sameCategory=true", func(t *testing.T) {
		params := []ast.Param{floatParam}
		args := []AbsValue{knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})}
		if argsMatchCategoryStatic(params, args, true) {
			t.Errorf("expected false for int->float with sameCategory=true")
		}
	})
}

func TestBinaryOpResultType(t *testing.T) {
	tests := []struct {
		name  string
		left  ast.Type
		right ast.Type
		op    string
		want  ast.Type
	}{
		{"int32 + int32", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "+", ast.IntegerType{Size: 32, Signed: true, Nullable: false}},
		{"int32 + int64", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 64, Signed: true, Nullable: false}, "+", ast.IntegerType{Size: 64, Signed: true, Nullable: false}},
		{"int64 + int32", ast.IntegerType{Size: 64, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "+", ast.IntegerType{Size: 64, Signed: true, Nullable: false}},
		{"int32 + float64", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}, "+", ast.FloatType{Size: 64, Nullable: false}},
		{"float64 + int32", ast.FloatType{Size: 64, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "+", ast.FloatType{Size: 64, Nullable: false}},
		{"float64 + float32", ast.FloatType{Size: 64, Nullable: false}, ast.FloatType{Size: 32, Nullable: false}, "+", ast.FloatType{Size: 64, Nullable: false}},
		{"float32 + float64", ast.FloatType{Size: 32, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}, "+", ast.FloatType{Size: 64, Nullable: false}},
		{"int32 - int32", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "-", ast.IntegerType{Size: 32, Signed: true, Nullable: false}},
		{"int32 * int64", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 64, Signed: true, Nullable: false}, "*", ast.IntegerType{Size: 64, Signed: true, Nullable: false}},
		{"int32 / float64", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}, "/", ast.FloatType{Size: 64, Nullable: false}},
		{"int32 % int32", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "%", ast.IntegerType{Size: 32, Signed: true, Nullable: false}},
		// Incompatible: bool + int -> nil
		{"bool + int", ast.BoolType{Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "+", nil},
		// Incompatible: int8 cannot widen to float32
		{"int64 + float64 incompatible", ast.IntegerType{Size: 64, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}, "+", nil},
		// Non-arithmetic op returns nil
		{"int32 == int32", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "==", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := binaryOpResultType(tt.left, tt.right, tt.op)
			if tt.want == nil {
				if got != nil {
					t.Errorf("binaryOpResultType(%v, %v, %q) = %v, want nil", tt.left, tt.right, tt.op, got)
				}
				return
			}
			if !typesEqual(got, tt.want) {
				t.Errorf("binaryOpResultType(%v, %v, %q) = %v, want %v", tt.left, tt.right, tt.op, got, tt.want)
			}
		})
	}
}

func TestGetTypesFromAbs(t *testing.T) {
	tests := []struct {
		name string
		val  AbsValue
		want int // expected number of types
	}{
		{"union", AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}}}, 2},
		{"int", knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), 1},
		{"float", knownFloatValue(1.0, ast.FloatType{Size: 64, Nullable: false}), 1},
		{"bool", knownBoolValue(true, ast.BoolType{Nullable: false}), 0},
		{"string", AbsValue{kind: AbsString}, 0},
		{"null", nullAbsValue(), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTypesFromAbs(tt.val)
			if len(got) != tt.want {
				t.Errorf("getTypesFromAbs(%v) returned %d types, want %d: %v", tt.val, len(got), tt.want, got)
			}
		})
	}
}

func TestDedupTypes(t *testing.T) {
	tests := []struct {
		name  string
		types []ast.Type
		want  int
	}{
		{"empty", nil, 0},
		{"all unique", []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}}, 2},
		{"all same", []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: true}}, 1},
		{"some duplicates", []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: true}}, 2},
		{"differ by signed", []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: false, Nullable: false}}, 2},
		{"float different size", []ast.Type{ast.FloatType{Size: 32, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedupTypes(tt.types)
			if len(got) != tt.want {
				t.Errorf("dedupTypes(%v) returned %d types, want %d: %v", tt.types, len(got), tt.want, got)
			}
		})
	}
}

func TestTypesEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b ast.Type
		want bool
	}{
		{"same int", ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.IntegerType{Size: 32, Signed: true, Nullable: true}, true},
		{"diff size int", ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 64, Signed: true}, false},
		{"diff signed int", ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 32, Signed: false}, false},
		{"same float", ast.FloatType{Size: 64, Nullable: false}, ast.FloatType{Size: 64, Nullable: true}, true},
		{"diff size float", ast.FloatType{Size: 32}, ast.FloatType{Size: 64}, false},
		{"int vs float", ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{Size: 32}, false},
		{"int vs bool", ast.IntegerType{Size: 32, Signed: true}, ast.BoolType{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("typesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestOpToVerb(t *testing.T) {
	tests := []struct {
		op   string
		want string
	}{
		{"+", "add"},
		{"-", "subtract"},
		{"*", "multiply"},
		{"/", "divide"},
		{"%", "modulo"},
		{"unknown", "unknown"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			if got := opToVerb(tt.op); got != tt.want {
				t.Errorf("opToVerb(%q) = %q, want %q", tt.op, got, tt.want)
			}
		})
	}
}

func TestCountMatchingOverloads(t *testing.T) {
	intParam := ast.Param{Name: "x", Type: ast.IntegerType{Size: 32, Signed: true, Nullable: false}}
	floatParam := ast.Param{Name: "x", Type: ast.FloatType{Size: 64, Nullable: false}}

	matchingInt := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	int32ToFloat64 := ast.IntegerType{Size: 32, Signed: true, Nullable: false}

	t.Run("exact match only", func(t *testing.T) {
		// int param matches; bool param doesn't match int arg at all
		overloads := []*ast.FuncDecl{
			{Parameters: []ast.Param{intParam}},
			{Parameters: []ast.Param{{Name: "x", Type: ast.BoolType{Nullable: false}}}},
		}
		count := countMatchingOverloads(overloads, []AbsValue{matchingInt})
		if count != 1 {
			t.Errorf("expected 1 match for int arg, got %d", count)
		}
	})

	t.Run("match via category", func(t *testing.T) {
		// int32 can widen to float64 via implicit conversion
		overloads := []*ast.FuncDecl{
			{Parameters: []ast.Param{floatParam}},
		}
		// Can int(32,signed) match float(64) via cross-category? Yes.
		intArg := knownIntValue(5, int32ToFloat64)
		count := countMatchingOverloads(overloads, []AbsValue{intArg})
		if count != 1 {
			t.Errorf("expected 1 match (int->float cross category), got %d", count)
		}
	})

	t.Run("no match", func(t *testing.T) {
		overloads := []*ast.FuncDecl{
			{Parameters: []ast.Param{{Name: "x", Type: ast.BoolType{Nullable: false}}}},
		}
		count := countMatchingOverloads(overloads, []AbsValue{matchingInt})
		if count != 0 {
			t.Errorf("expected 0 matches, got %d", count)
		}
	})

	t.Run("multiple overloads match same args", func(t *testing.T) {
		overloads := []*ast.FuncDecl{
			{Parameters: []ast.Param{intParam}},
			{Parameters: []ast.Param{intParam}},
		}
		count := countMatchingOverloads(overloads, []AbsValue{matchingInt})
		if count != 2 {
			t.Errorf("expected 2 matches, got %d", count)
		}
	})

	t.Run("empty overloads", func(t *testing.T) {
		count := countMatchingOverloads(nil, []AbsValue{matchingInt})
		if count != 0 {
			t.Errorf("expected 0 matches for nil overloads, got %d", count)
		}
	})

	t.Run("wrong arg count", func(t *testing.T) {
		overloads := []*ast.FuncDecl{
			{Parameters: []ast.Param{intParam, floatParam}},
		}
		count := countMatchingOverloads(overloads, []AbsValue{matchingInt})
		if count != 0 {
			t.Errorf("expected 0 matches for wrong arg count, got %d", count)
		}
	})

	t.Run("untyped int matches int param", func(t *testing.T) {
		overloads := []*ast.FuncDecl{
			{Parameters: []ast.Param{intParam}},
			{Parameters: []ast.Param{floatParam}},
		}
		untyped := knownIntValue(5, ast.IntegerType{Size: 64, Signed: true, Nullable: false})
		untyped.untyped = true
		count := countMatchingOverloads(overloads, []AbsValue{untyped})
		// untyped int matches IntegerType param exactly (exact match), but not FloatType param via exact match
		// For FloatType param: argMatchCategoryStatic has arg.kind==AbsInt, sameCategory=false, untyped=true -> returns true
		// So both overloads match -> count = 2
		if count != 2 {
			t.Errorf("expected 2 matches (int via exact, float via cross-category), got %d", count)
		}
	})
}

func TestAnalyzeUnionBinaryExpr(t *testing.T) {
	t.Run("single result type", func(t *testing.T) {
		a := New()
		left := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "+")
		if result.kind != AbsInt {
			t.Errorf("expected AbsInt, got %v", result.kind)
		}
		if len(a.Errors()) > 0 {
			t.Errorf("expected no errors, got %v", a.Errors())
		}
	})

	t.Run("multiple result types", func(t *testing.T) {
		a := New()
		left := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{
			ast.IntegerType{Size: 32, Signed: true, Nullable: false},
			ast.FloatType{Size: 64, Nullable: false},
		}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "+")
		if result.kind != AbsUnion {
			t.Errorf("expected AbsUnion, got %v", result.kind)
		}
		if len(result.unionTypes) != 2 {
			t.Errorf("expected 2 union types, got %d: %v", len(result.unionTypes), result.unionTypes)
		}
		if len(a.Errors()) > 0 {
			t.Errorf("expected no errors, got %v", a.Errors())
		}
	})

	t.Run("incompatible types produce error", func(t *testing.T) {
		a := New()
		left := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.BoolType{Nullable: false}}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "+")
		if len(a.Errors()) == 0 {
			t.Errorf("expected error for incompatible types, got none")
		}
		if result.kind != AbsUnion {
			t.Errorf("expected AbsUnion with empty types on error")
		}
	})

	t.Run("sub and mul with unions", func(t *testing.T) {
		a := New()
		left := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 64, Signed: true, Nullable: false}}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "-")
		if result.kind != AbsInt {
			t.Errorf("expected AbsInt for subtraction, got %v", result.kind)
		}
		result = a.analyzeUnionBinaryExpr(1, left, right, "*")
		if result.kind != AbsInt {
			t.Errorf("expected AbsInt for multiplication, got %v", result.kind)
		}
	})

	t.Run("dedup eliminates duplicates", func(t *testing.T) {
		a := New()
		left := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{
			ast.IntegerType{Size: 32, Signed: true, Nullable: false},
		}}
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{
			ast.IntegerType{Size: 32, Signed: true, Nullable: false},
			ast.IntegerType{Size: 32, Signed: true, Nullable: true},
		}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "+")
		// int(32,signed,false) + int(32,signed,false) -> int(32,signed,false)
		// int(32,signed,false) + int(32,signed,true) -> int(32,signed,true)
		// typesEqual ignores Nullable, so both results are "equal" -> dedup to 1
		if result.kind != AbsInt {
			t.Errorf("expected single type after dedup (AbsInt), got kind=%v types=%v", result.kind, result.unionTypes)
		}
	})

	t.Run("non-arithmetic op returns empty union", func(t *testing.T) {
		a := New()
		left := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "==")
		// binaryOpResultType returns nil for "==", so for all pairs res==nil -> errors added
		if len(a.Errors()) == 0 {
			t.Errorf("expected errors for non-arithmetic op with union")
		}
		if result.kind != AbsUnion {
			t.Errorf("expected AbsUnion after all errors, got %v", result.kind)
		}
	})

	t.Run("mixed union with non-union left", func(t *testing.T) {
		a := New()
		// Non-union left (int) + union right
		left := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "+")
		if result.kind != AbsInt {
			t.Errorf("expected AbsInt, got %v", result.kind)
		}
	})

	t.Run("float union", func(t *testing.T) {
		a := New()
		left := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.FloatType{Size: 32, Nullable: false}}}
		right := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.FloatType{Size: 64, Nullable: false}}}
		result := a.analyzeUnionBinaryExpr(1, left, right, "+")
		if result.kind != AbsFloat {
			t.Errorf("expected AbsFloat, got %v", result.kind)
		}
	})
}

// -- Remaining branch coverage --

func TestDefaultValueForParamUnion(t *testing.T) {
	got := defaultValueForParam(ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}})
	if got.kind != AbsUnion {
		t.Errorf("expected AbsUnion, got %v", got.kind)
	}
}

func TestDefaultValueForParamString(t *testing.T) {
	got := defaultValueForParam(ast.StringType{})
	if got.kind != AbsString {
		t.Errorf("expected AbsString, got %v", got.kind)
	}
}

func TestDefaultValueForParamArray(t *testing.T) {
	got := defaultValueForParam(ast.ArrayType{ElemType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}, Size: 5})
	if got.kind != AbsArray {
		t.Errorf("expected AbsArray, got %v", got.kind)
	}
}

func TestDefaultValueForParamList(t *testing.T) {
	got := defaultValueForParam(ast.ListType{ElemType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}})
	if got.kind != AbsList {
		t.Errorf("expected AbsList, got %v", got.kind)
	}
}

func TestAbsValueFromTypeAllPaths(t *testing.T) {
	if got := absValueFromType(ast.IntegerType{Size: 8, Signed: true, Nullable: false}); got.kind != AbsInt {
		t.Errorf("expected AbsInt")
	}
	if got := absValueFromType(ast.FloatType{Size: 32, Nullable: false}); got.kind != AbsFloat {
		t.Errorf("expected AbsFloat")
	}
	if got := absValueFromType(ast.BoolType{Nullable: false}); got.kind != AbsBool {
		t.Errorf("expected AbsBool")
	}
	if got := absValueFromType(ast.StringType{}); got.kind != AbsString {
		t.Errorf("expected AbsString")
	}
	if got := absValueFromType(ast.ArrayType{ElemType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}, Size: 5}); got.kind != AbsArray {
		t.Errorf("expected AbsArray")
	}
	if got := absValueFromType(ast.ListType{ElemType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}}); got.kind != AbsList {
		t.Errorf("expected AbsList")
	}
	if got := absValueFromType(ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}); got.kind != AbsUnion {
		t.Errorf("expected AbsUnion")
	}
}

func TestDefaultValueAllPaths(t *testing.T) {
	a := New()
	if got := a.defaultValue(&ast.VarDecl{IsUnion: true}); got.kind != AbsUnion {
		t.Errorf("expected AbsUnion for union default")
	}
	if got := a.defaultValue(&ast.VarDecl{IsFloat: true, FType: ast.FloatType{Size: 32, Nullable: false}}); got.kind != AbsFloat {
		t.Errorf("expected AbsFloat for float default")
	}
	if got := a.defaultValue(&ast.VarDecl{IsBool: true, BType: ast.BoolType{Nullable: false}}); got.kind != AbsBool {
		t.Errorf("expected AbsBool for bool default")
	}
}

func TestTypeDescFromAbsUnion(t *testing.T) {
	v := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: true}}}
	got := typeDescFromAbs(v)
	if got != "int{size: 32, signed: true, nullable: false} | float{size: 64, nullable: true}" {
		t.Errorf("unexpected union desc: %s", got)
	}
}

func TestTypeDescFromDeclFloat(t *testing.T) {
	got := typeDescFromDecl(&ast.VarDecl{IsFloat: true, FType: ast.FloatType{Size: 64, Nullable: true}})
	if got != "float{size: 64, nullable: true}" {
		t.Errorf("unexpected float desc: %s", got)
	}
}

func TestTypeDescFromDeclUnion(t *testing.T) {
	got := typeDescFromDecl(&ast.VarDecl{IsUnion: true, UnionType: ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}})
	if got != "int{size: 32, signed: true, nullable: false}" {
		t.Errorf("unexpected union desc: %s", got)
	}
}

func TestCheckAssignmentTypeDeclIsNil(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x", knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), nil)
}

func TestCheckAssignmentTypeUnionAssign(t *testing.T) {
	a := New()
	unionRHS := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}}}
	decl := &ast.VarDecl{Name: "x", IsUnion: true, UnionType: ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}}
	a.checkAssignmentType(1, "x", unionRHS, decl)
}

func TestCheckAssignmentTypeUnionToNonUnion(t *testing.T) {
	a := New()
	unionRHS := AbsValue{kind: AbsUnion, unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.BoolType{Nullable: false}}}
	decl := &ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}}
	a.checkAssignmentType(1, "x", unionRHS, decl)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for union with incompatible type to non-union decl")
	}
}

func TestCheckReturnTypeNilRetType(t *testing.T) {
	a := New()
	a.checkReturnType(1, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false}), nil)
}

func TestCheckReturnTypeUnion(t *testing.T) {
	a := New()
	a.checkReturnType(1, AbsValue{kind: AbsUnion}, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
}

func TestCheckReturnTypeMismatch(t *testing.T) {
	a := New()
	a.checkReturnType(1, knownFloatValue(1.0, ast.FloatType{Size: 64, Nullable: false}), ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for return type mismatch")
	}
}

func TestIsTypeAssignableToUnion(t *testing.T) {
	u := ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}
	arg := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	if !isTypeAssignableTo(arg, u) {
		t.Errorf("expected int32 to be assignable to union(int32)")
	}
}

func TestIsTypeAssignableToNonUnion(t *testing.T) {
	arg := knownFloatValue(1.0, ast.FloatType{Size: 32, Nullable: false})
	if isTypeAssignableTo(arg, ast.IntegerType{Size: 32, Signed: true, Nullable: false}) {
		t.Errorf("expected float to not be assignable to int")
	}
}

func TestIsNullableDeclUnion(t *testing.T) {
	d := &ast.VarDecl{IsUnion: true, UnionType: ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: true}}}}
	if !isNullableDecl(d) {
		t.Errorf("expected union with nullable member to be nullable")
	}
}

func TestIsNullableDeclUnionAllNonNullable(t *testing.T) {
	d := &ast.VarDecl{IsUnion: true, UnionType: ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true, Nullable: false}}}}
	if isNullableDecl(d) {
		t.Errorf("expected union with all non-nullable to be non-nullable")
	}
}

func TestResolveOverloadStaticCategoryMatch(t *testing.T) {
	overloads := []*ast.FuncDecl{
		{Parameters: []ast.Param{{Name: "x", Type: ast.FloatType{Size: 64, Nullable: false}}}},
	}
	args := []AbsValue{knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})}
	decl := resolveOverloadStatic(overloads, args)
	if decl == nil {
		t.Errorf("expected int to match float param via cross-category")
	}
}

func TestAnalyzeBinaryExprFloatAdd(t *testing.T) {
	a := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.0, FType: ast.FloatType{Size: 64, Nullable: true}, Untyped: true, Line: 1},
		Op:    "+",
		Right: &ast.FloatLit{Value: 2.0, FType: ast.FloatType{Size: 64, Nullable: true}, Untyped: true, Line: 1},
		Line:  1,
	}
	result := a.analyzeBinaryExpr(expr)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got %v", a.Errors())
	}
	_ = result
}

func TestArgExactTypeMatchStaticUntypedFloat(t *testing.T) {
	arg := knownFloatValue(1.0, ast.FloatType{Size: 64, Nullable: false})
	arg.untyped = true
	if !argExactTypeMatchStatic(ast.FloatType{Size: 64, Nullable: false}, arg) {
		t.Errorf("expected untyped float to match float64")
	}
}

func TestArgExactTypeMatchStaticNullToNullable(t *testing.T) {
	nullVal := nullAbsValue()
	if !argExactTypeMatchStatic(ast.IntegerType{Size: 32, Signed: true, Nullable: true}, nullVal) {
		t.Errorf("expected null to match nullable int param")
	}
	if argExactTypeMatchStatic(ast.IntegerType{Size: 32, Signed: true, Nullable: false}, nullVal) {
		t.Errorf("expected null to NOT match non-nullable int param")
	}
}

func TestTypesEqualFloat(t *testing.T) {
	if !typesEqual(ast.FloatType{Size: 32, Nullable: false}, ast.FloatType{Size: 32, Nullable: true}) {
		t.Errorf("expected float32 == float32 (ignoring nullable)")
	}
	if typesEqual(ast.FloatType{Size: 32, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}) {
		t.Errorf("expected float32 != float64")
	}
}

func TestCheckNullArithHasGuards(t *testing.T) {
	a := New()
	a.guards = append(a.guards, guardInfo{name: "x", notNull: true})
	left := &ast.VarRef{Name: "x", Line: 1}
	right := &ast.VarRef{Name: "y", Line: 1}
	expr := &ast.BinaryExpr{Left: left, Op: "+", Right: right, Line: 1}
	leftVal := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: true})
	rightVal := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: true})
	a.checkNullArith(expr, leftVal, rightVal)
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for nullable right (y)")
	}
}

func TestCheckNullArithNullErrorMode(t *testing.T) {
	a := New()
	a.SetNullMode(NullError)
	left := &ast.VarRef{Name: "x", Line: 1}
	right := &ast.VarRef{Name: "y", Line: 1}
	expr := &ast.BinaryExpr{Left: left, Op: "+", Right: right, Line: 1}
	leftVal := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: true})
	rightVal := knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: true})
	a.checkNullArith(expr, leftVal, rightVal)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error in NullError mode")
	}
}

func TestAnalyzeBinaryExprNullArithBothInts(t *testing.T) {
	a := New()
	a.guards = append(a.guards, guardInfo{name: "x", notNull: true})
	expr := &ast.BinaryExpr{
		Left:  &ast.VarRef{Name: "x", Line: 1},
		Op:    "+",
		Right: &ast.VarRef{Name: "y", Line: 1},
		Line:  1,
	}
	a.env["x"] = knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	a.env["y"] = knownIntValue(5, ast.IntegerType{Size: 32, Signed: true, Nullable: false})
	a.analyzeBinaryExpr(expr)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors, got %v", a.Errors())
	}
}

func TestBinaryOpResultTypeIncompatible(t *testing.T) {
	result := binaryOpResultType(ast.IntegerType{Size: 64, Signed: true, Nullable: false}, ast.FloatType{Size: 32, Nullable: false}, "+")
	if result != nil {
		t.Errorf("expected nil for int64 + float32 (invalid)")
	}
}

func TestBinaryOpResultTypeIntAndFloatFromRight(t *testing.T) {
	result := binaryOpResultType(ast.IntegerType{Size: 8, Signed: true, Nullable: false}, ast.FloatType{Size: 64, Nullable: false}, "+")
	if result == nil {
		t.Errorf("expected a float type for int8 + float64")
	}
}

func TestAnalyzeExprStringLit(t *testing.T) {
	a := New()
	result := a.analyzeExpr(&ast.StringLit{Value: "hello", Untyped: true, Line: 1})
	if result.kind != AbsString {
		t.Errorf("expected AbsString")
	}
}

func TestAnalyzeExprUntypedBoolLit(t *testing.T) {
	a := New()
	result := a.analyzeExpr(&ast.BoolLit{Value: true, Untyped: true, Line: 1})
	if result.kind != AbsBool {
		t.Errorf("expected AbsBool")
	}
}

func TestAnalyzeExprUntypedIntLit(t *testing.T) {
	a := New()
	result := a.analyzeExpr(&ast.IntegerLit{Value: 42, Untyped: true, Line: 1})
	if result.kind != AbsInt {
		t.Errorf("expected AbsInt")
	}
}

// -- Coverage: typesEqual --
func TestTypesEqualFloatMismatch(t *testing.T) {
	if typesEqual(ast.FloatType{Size: 32}, ast.FloatType{Size: 64}) {
		t.Errorf("expected false for float32 != float64")
	}
}

func TestTypesEqualDifferentKinds(t *testing.T) {
	if typesEqual(ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{Size: 32}) {
		t.Errorf("expected false for int vs float")
	}
}

func TestTypesEqualNonIntFloat(t *testing.T) {
	if typesEqual(ast.StringType{Size: 10}, ast.IntegerType{Size: 32, Signed: true}) {
		t.Errorf("expected false for string vs int")
	}
}

// -- Coverage: hasNullLiteralArg --
func TestHasNullLiteralArgNoNull(t *testing.T) {
	if hasNullLiteralArg(
		[]ast.Expr{&ast.IntegerLit{Value: 1, Line: 1}},
		[]AbsValue{knownIntValue(1, ast.IntegerType{Size: 32, Signed: true})},
	) {
		t.Errorf("expected false when no null literal in args")
	}
}

func TestHasNullLiteralArgNoArgs(t *testing.T) {
	if hasNullLiteralArg([]ast.Expr{}, []AbsValue{}) {
		t.Errorf("expected false for empty args")
	}
}

// -- Coverage: isTypeAssignableTo --
func TestIsTypeAssignableToIntVal(t *testing.T) {
	if !isTypeAssignableTo(
		knownIntValue(5, ast.IntegerType{Size: 32, Signed: true}),
		ast.IntegerType{Size: 32, Signed: true},
	) {
		t.Errorf("expected int value assignable to int type")
	}
}

func TestIsTypeAssignableToIntValToFloatType(t *testing.T) {
	if !isTypeAssignableTo(
		knownIntValue(5, ast.IntegerType{Size: 32, Signed: true}),
		ast.FloatType{Size: 64},
	) {
		t.Errorf("expected true for int32 -> float64 via cross-category")
	}
}

func TestIsTypeAssignableToIntValToUnion(t *testing.T) {
	if !isTypeAssignableTo(
		knownIntValue(5, ast.IntegerType{Size: 32, Signed: true}),
		ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{Size: 64}}},
	) {
		t.Errorf("expected int value assignable to union containing int")
	}
}

func TestIsTypeAssignableToIntValToUnionCrossCategory(t *testing.T) {
	if !isTypeAssignableTo(
		knownIntValue(5, ast.IntegerType{Size: 32, Signed: true}),
		ast.UnionType{Types: []ast.Type{ast.FloatType{Size: 64}}},
	) {
		t.Errorf("expected int value assignable to union via cross-category to float")
	}
}

func TestIsTypeAssignableToFloatVal(t *testing.T) {
	if !isTypeAssignableTo(
		knownFloatValue(1.5, ast.FloatType{Size: 64}),
		ast.FloatType{Size: 64},
	) {
		t.Errorf("expected float value assignable to float type")
	}
}

func TestIsTypeAssignableToNullToNullable(t *testing.T) {
	if !isTypeAssignableTo(
		nullAbsValue(),
		ast.IntegerType{Size: 32, Signed: true, Nullable: true},
	) {
		t.Errorf("expected null value assignable to nullable int")
	}
}

func TestIsTypeAssignableToNullToNonNullable(t *testing.T) {
	if isTypeAssignableTo(
		nullAbsValue(),
		ast.IntegerType{Size: 32, Signed: true},
	) {
		t.Errorf("expected false for null value not assignable to non-nullable int")
	}
}

// -- Coverage: checkAssignmentType with various decls --
func TestCheckAssignmentTypeUnionToInt(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x",
		absValueFromType(ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true}}}),
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true}},
	)
	if len(a.Errors()) != 0 {
		t.Errorf("expected no errors for union{int} assigned to int, got %v", a.Errors())
	}
}

func TestCheckAssignmentTypeIntToUnion(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x",
		knownIntValue(5, ast.IntegerType{Size: 32, Signed: true}),
		&ast.VarDecl{Name: "x", IsUnion: true, UnionType: ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{Size: 64}}}},
	)
	if len(a.Errors()) != 0 {
		t.Errorf("expected no errors for int assigned to union{int|float}, got %v", a.Errors())
	}
}

func TestCheckAssignmentTypeNullToNullable(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x",
		nullAbsValue(),
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true}},
	)
	if len(a.Errors()) != 0 {
		t.Errorf("expected no errors for null assigned to nullable int, got %v", a.Errors())
	}
}

func TestCheckAssignmentTypeNullToNonNullable(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x",
		nullAbsValue(),
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true}},
	)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for null assigned to non-nullable int")
	}
}

func TestCheckAssignmentTypeNullToNonNullableBool(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "b",
		nullAbsValue(),
		&ast.VarDecl{Name: "b", IsBool: true, BType: ast.BoolType{}},
	)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for null assigned to non-nullable bool")
	}
}

func TestCheckAssignmentTypeNullToNullableBool(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "b",
		nullAbsValue(),
		&ast.VarDecl{Name: "b", IsBool: true, BType: ast.BoolType{Nullable: true}},
	)
	if len(a.Errors()) != 0 {
		t.Errorf("expected no errors for null assigned to nullable bool, got %v", a.Errors())
	}
}

// -- Coverage: checkNullArith --
func TestCheckNullArithNullNoneMode(t *testing.T) {
	a := New()
	a.nullMode = NullNone
	a.checkNullArith(
		&ast.BinaryExpr{Left: &ast.VarRef{Name: "x", Line: 1}, Right: &ast.VarRef{Name: "y", Line: 1}, Op: "+", Line: 1},
		AbsValue{kind: AbsInt, nullable: true},
		AbsValue{kind: AbsInt, nullable: true},
	)
	if len(a.Warnings()) != 0 || len(a.Errors()) != 0 {
		t.Errorf("expected no warnings or errors in NullNone mode")
	}
}

func TestCheckNullArithGuardedLeft(t *testing.T) {
	a := New()
	a.nullMode = NullWarn
	a.guards = append(a.guards, guardInfo{name: "x", notNull: true})
	a.checkNullArith(
		&ast.BinaryExpr{Left: &ast.VarRef{Name: "x", Line: 1}, Right: &ast.VarRef{Name: "y", Line: 1}, Op: "+", Line: 1},
		AbsValue{kind: AbsInt, nullable: false},
		AbsValue{kind: AbsInt, nullable: true},
	)
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for nullable right operand")
	}
}

func TestCheckNullArithGuardedRight(t *testing.T) {
	a := New()
	a.nullMode = NullWarn
	a.guards = append(a.guards, guardInfo{name: "y", notNull: true})
	a.checkNullArith(
		&ast.BinaryExpr{Left: &ast.VarRef{Name: "x", Line: 1}, Right: &ast.VarRef{Name: "y", Line: 1}, Op: "+", Line: 1},
		AbsValue{kind: AbsInt, nullable: true},
		AbsValue{kind: AbsInt, nullable: false},
	)
	if len(a.Warnings()) == 0 {
		t.Errorf("expected warning for nullable left operand")
	}
}

// -- Coverage: analyzeBinaryExpr with comparison/boolean ops --
func TestAnalyzeBinaryExprComparison(t *testing.T) {
	a := New()
	a.env["x"] = knownIntValue(5, ast.IntegerType{Size: 32, Signed: true})
	a.env["y"] = knownIntValue(10, ast.IntegerType{Size: 32, Signed: true})
	result := a.analyzeBinaryExpr(&ast.BinaryExpr{
		Left: &ast.VarRef{Name: "x", Line: 1},
		Op:   "<", Right: &ast.VarRef{Name: "y", Line: 1}, Line: 1,
	})
	if result.kind != AbsBool {
		t.Errorf("expected AbsBool from comparison, got kind=%d", result.kind)
	}
}

func TestAnalyzeBinaryExprLogical(t *testing.T) {
	a := New()
	a.env["x"] = knownBoolValue(true, ast.BoolType{})
	a.env["y"] = knownBoolValue(false, ast.BoolType{})
	result := a.analyzeBinaryExpr(&ast.BinaryExpr{
		Left: &ast.VarRef{Name: "x", Line: 1},
		Op:   "&&", Right: &ast.VarRef{Name: "y", Line: 1}, Line: 1,
	})
	if result.kind != AbsBool {
		t.Errorf("expected AbsBool from logical op, got kind=%d", result.kind)
	}
}

// -- Coverage: countMatchingOverloads --
func TestCountMatchingOverloadsNone(t *testing.T) {
	if count := countMatchingOverloads(nil, []AbsValue{knownIntValue(5, ast.IntegerType{Size: 32, Signed: true})}); count != 0 {
		t.Errorf("expected 0 matches for nil overloads, got %d", count)
	}
}

func TestCountMatchingOverloadsExactAndCross(t *testing.T) {
	int32Param := ast.Param{Type: ast.IntegerType{Size: 32, Signed: true}}
	float64Param := ast.Param{Type: ast.FloatType{Size: 64}}
	overloads := []*ast.FuncDecl{
		{Parameters: []ast.Param{int32Param}},
		{Parameters: []ast.Param{float64Param}},
	}
	count := countMatchingOverloads(overloads, []AbsValue{knownIntValue(5, ast.IntegerType{Size: 32, Signed: true})})
	if count != 2 {
		t.Errorf("expected 2 matches (exact + cross-category), got %d", count)
	}
}

// -- Coverage: resolveOverloadStatic --
func TestResolveOverloadStaticCrossCategory(t *testing.T) {
	int32Param := ast.Param{Type: ast.IntegerType{Size: 32, Signed: true}}
	float64Param := ast.Param{Type: ast.FloatType{Size: 64}}
	overloads := []*ast.FuncDecl{
		{Parameters: []ast.Param{int32Param}},
		{Parameters: []ast.Param{float64Param}},
	}
	result := resolveOverloadStatic(overloads, []AbsValue{knownIntValue(5, ast.IntegerType{Size: 8, Signed: true})})
	if result == nil {
		t.Errorf("expected int32 overload match for int8 arg")
	}
}

func TestResolveOverloadStaticNoMatch(t *testing.T) {
	if result := resolveOverloadStatic(nil, []AbsValue{knownIntValue(5, ast.IntegerType{Size: 32, Signed: true})}); result != nil {
		t.Errorf("expected nil for no overloads")
	}
}

// -- Coverage: argExactTypeMatchStatic --
func TestArgExactTypeMatchStaticNullWithNullable(t *testing.T) {
	if !argExactTypeMatchStatic(ast.IntegerType{Size: 32, Signed: true, Nullable: true}, nullAbsValue()) {
		t.Errorf("expected true for null with nullable int")
	}
}

func TestArgExactTypeMatchStaticNullWithNonNullable(t *testing.T) {
	if argExactTypeMatchStatic(ast.IntegerType{Size: 32, Signed: true}, nullAbsValue()) {
		t.Errorf("expected false for null with non-nullable int")
	}
}

func TestArgExactTypeMatchStaticAnyIntWithInt(t *testing.T) {
	if !argExactTypeMatchStatic(ast.IntegerType{Size: 32, Signed: true}, anyIntValue(ast.IntegerType{Size: 32, Signed: true})) {
		t.Errorf("expected true for anyInt matching int32")
	}
}

func TestArgExactTypeMatchStaticIntWithFloat(t *testing.T) {
	if argExactTypeMatchStatic(ast.IntegerType{Size: 32, Signed: true}, knownFloatValue(1.5, ast.FloatType{Size: 64})) {
		t.Errorf("expected false for int param with float arg")
	}
}

func TestArgExactTypeMatchStaticBoolWithBool(t *testing.T) {
	if !argExactTypeMatchStatic(ast.BoolType{}, knownBoolValue(true, ast.BoolType{})) {
		t.Errorf("expected true for bool with bool")
	}
}

func TestArgExactTypeMatchStaticBoolWithInt(t *testing.T) {
	if argExactTypeMatchStatic(ast.BoolType{}, knownIntValue(1, ast.IntegerType{Size: 32, Signed: true})) {
		t.Errorf("expected false for bool param with int arg")
	}
}

func TestArgExactTypeMatchStaticStringWithString(t *testing.T) {
	if !argExactTypeMatchStatic(ast.StringType{Size: 10}, AbsValue{kind: AbsString}) {
		t.Errorf("expected true for string with string")
	}
}

func TestArgExactTypeMatchStaticUnionWithInt(t *testing.T) {
	if !argExactTypeMatchStatic(ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true}}}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true})) {
		t.Errorf("expected true for union containing int with int arg")
	}
}

func TestArgExactTypeMatchStaticUnionNoMatch(t *testing.T) {
	if argExactTypeMatchStatic(ast.UnionType{Types: []ast.Type{ast.FloatType{Size: 64}}}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true})) {
		t.Errorf("expected false for union with only float not matching int arg")
	}
}

// -- Coverage: argMatchCategoryStatic --
func TestArgMatchCategoryStaticFloatWithIntSame(t *testing.T) {
	if argMatchCategoryStatic(ast.FloatType{Size: 64}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true}), true) {
		t.Errorf("expected false for float param with int arg (same category)")
	}
}

func TestArgMatchCategoryStaticFloatWithIntCross(t *testing.T) {
	if !argMatchCategoryStatic(ast.FloatType{Size: 64}, knownIntValue(5, ast.IntegerType{Size: 32, Signed: true}), false) {
		t.Errorf("expected true for float param with int arg (cross category)")
	}
}

func TestArgMatchCategoryStaticBoolWithInt(t *testing.T) {
	if argMatchCategoryStatic(ast.BoolType{}, knownIntValue(1, ast.IntegerType{Size: 32, Signed: true}), true) {
		t.Errorf("expected false for bool param with int arg")
	}
}

func TestArgMatchCategoryStaticNullWithNullable(t *testing.T) {
	if !argMatchCategoryStatic(ast.IntegerType{Size: 32, Signed: true, Nullable: true}, nullAbsValue(), true) {
		t.Errorf("expected true for null arg with nullable int param")
	}
}

// -- Coverage: typeDescFromDecl --
func TestTypeDescFromDeclBoolDesc(t *testing.T) {
	result := typeDescFromDecl(&ast.VarDecl{Name: "b", IsBool: true, BType: ast.BoolType{Nullable: true}})
	if result != "bool{nullable: true}" {
		t.Errorf("expected bool description, got %q", result)
	}
}

func TestTypeDescFromDeclFloatDesc(t *testing.T) {
	result := typeDescFromDecl(&ast.VarDecl{Name: "f", IsFloat: true, FType: ast.FloatType{Size: 64, Nullable: false}})
	if result != "float{size: 64, nullable: false}" {
		t.Errorf("expected float description, got %q", result)
	}
}

func TestTypeDescFromDeclIntDesc(t *testing.T) {
	result := typeDescFromDecl(&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}})
	if result != "int{size: 32, signed: true, nullable: false}" {
		t.Errorf("expected int description, got %q", result)
	}
}

func TestTypeDescFromDeclUnionDesc(t *testing.T) {
	result := typeDescFromDecl(&ast.VarDecl{
		Name: "u", IsUnion: true,
		UnionType: ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true}}},
	})
	if result != "int{size: 32, signed: true, nullable: false}" {
		t.Errorf("expected union description, got %q", result)
	}
}

// -- Coverage: hasNullLiteralArg with null literal --
func TestHasNullLiteralArgWithNull(t *testing.T) {
	if !hasNullLiteralArg(
		[]ast.Expr{&ast.NullLit{Line: 1}},
		[]AbsValue{nullAbsValue()},
	) {
		t.Errorf("expected true when null literal in args")
	}
}

// -- Coverage: absValueFromType --
func TestAbsValueFromTypeString(t *testing.T) {
	v := absValueFromType(ast.StringType{Size: 10})
	if v.kind != AbsString || !v.isAnyStr {
		t.Errorf("expected AbsString from StringType, got kind=%d isAnyStr=%v", v.kind, v.isAnyStr)
	}
}

func TestAbsValueFromTypeList(t *testing.T) {
	v := absValueFromType(ast.ListType{})
	if v.kind != AbsList {
		t.Errorf("expected AbsList from ListType, got kind=%d", v.kind)
	}
}

// -- Coverage: typeDescFromType --
func TestTypeDescFromTypeString(t *testing.T) {
	if s := typeDescFromType(ast.StringType{Size: 10}); s != "string" {
		t.Errorf("expected 'string', got %q", s)
	}
}

// -- Coverage: isTypeAssignableTo union no match --
func TestIsTypeAssignableToStringToUnion(t *testing.T) {
	if isTypeAssignableTo(
		AbsValue{kind: AbsString, isAnyStr: true},
		ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{Size: 64}}},
	) {
		t.Errorf("expected false for string not assignable to union{int, float}")
	}
}

// -- Coverage: checkAssignmentType union no error path (all types compatible) --
func TestCheckAssignmentTypeUnionAllCompatible(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x",
		AbsValue{
			kind:       AbsUnion,
			unionTypes: []ast.Type{ast.IntegerType{Size: 32, Signed: true}},
		},
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true}},
	)
	if len(a.Errors()) != 0 {
		t.Errorf("expected no error for union{int} assigned to int, got %v", a.Errors())
	}
}

// -- Coverage: typeDescFromType default case --
func TestTypeDescFromTypeArray(t *testing.T) {
	s := typeDescFromType(ast.ArrayType{})
	if s == "" {
		t.Errorf("expected non-empty description for ArrayType")
	}
}

func TestTypeDescFromTypeList(t *testing.T) {
	s := typeDescFromType(ast.ListType{})
	if s == "" {
		t.Errorf("expected non-empty description for ListType")
	}
}

// -- Coverage: typeDescFromDecl string desc --
func TestTypeDescFromDeclString(t *testing.T) {
	result := typeDescFromDecl(&ast.VarDecl{Name: "s", IsString: true, SType: ast.StringType{Size: 10}})
	if result != "string" {
		t.Errorf("expected 'string', got %q", result)
	}
}

// -- Coverage: absValueFromType array / union / default --
func TestAbsValueFromTypeArray(t *testing.T) {
	v := absValueFromType(ast.ArrayType{ElemType: ast.IntegerType{Size: 32, Signed: true}})
	if v.kind != AbsArray {
		t.Errorf("expected AbsArray, got kind=%d", v.kind)
	}
}

func TestAbsValueFromTypeUnion(t *testing.T) {
	v := absValueFromType(ast.UnionType{Types: []ast.Type{ast.IntegerType{Size: 32, Signed: true}}})
	if v.kind != AbsUnion {
		t.Errorf("expected AbsUnion, got kind=%d", v.kind)
	}
}

func TestAbsValueFromTypeDefault(t *testing.T) {
	v := absValueFromType(struct{ ast.Type }{})
	if v.kind != 0 || v.isAnyStr || v.definitelyNull {
		t.Errorf("expected zero AbsValue for unknown type, got %+v", v)
	}
}

// -- Coverage: analyzeCall non-VarRef --
func TestAnalyzeCallNonVarRef(t *testing.T) {
	a := New()
	program := parseProgram(t, `print(1);`)
	a.Analyze(program)
	// Test analyzeCall directly with a non-VarRef function
	result := a.analyzeExpr(&ast.CallExpr{
		Function: &ast.IntegerLit{Value: 42},
		Args:     []ast.Expr{},
		Line:     1,
	})
	if result.kind != 0 {
		t.Errorf("expected zero AbsValue for non-VarRef call, got %+v", result)
	}
}

// -- Coverage: countMatchingOverloads all three paths --
func TestCountMatchingOverloadsExactMatch(t *testing.T) {
	overloads := []*ast.FuncDecl{
		{Parameters: []ast.Param{{Name: "x", Type: ast.IntegerType{Size: 32, Signed: true}}}},
	}
	count := countMatchingOverloads(overloads, []AbsValue{knownIntValue(5, ast.IntegerType{Size: 32, Signed: true})})
	if count != 1 {
		t.Errorf("expected 1 exact match, got %d", count)
	}
}

func TestCountMatchingOverloadsCategoryMatch(t *testing.T) {
	overloads := []*ast.FuncDecl{
		{Parameters: []ast.Param{{Name: "x", Type: ast.FloatType{Size: 64}}}},
	}
	// Untyped int should match float via sameCategory
	count := countMatchingOverloads(overloads, []AbsValue{knownIntValue(5, ast.IntegerType{Size: 8, Signed: true})})
	if count != 1 {
		t.Errorf("expected 1 category match, got %d", count)
	}
}

// -- Coverage: argExactTypeMatchStatic ArrayType and ListType --
func TestArgExactTypeMatchStaticArray(t *testing.T) {
	if !argExactTypeMatchStatic(ast.ArrayType{}, AbsValue{kind: AbsArray}) {
		t.Errorf("expected true for array param with array arg")
	}
}

func TestArgExactTypeMatchStaticList(t *testing.T) {
	if !argExactTypeMatchStatic(ast.ListType{}, AbsValue{kind: AbsList}) {
		t.Errorf("expected true for list param with list arg")
	}
}

func TestArgExactTypeMatchStaticUntypedInt(t *testing.T) {
	if !argExactTypeMatchStatic(ast.IntegerType{Size: 32, Signed: true}, AbsValue{kind: AbsInt, untyped: true, intType: ast.IntegerType{Size: 32, Signed: true}}) {
		t.Errorf("expected true for untyped int matching int param")
	}
}

// -- Coverage: argMatchCategoryStatic FloatType float arg --
func TestArgMatchCategoryStaticFloatWithFloat(t *testing.T) {
	if !argMatchCategoryStatic(ast.FloatType{Size: 64}, AbsValue{kind: AbsFloat, untyped: true, floatType: ast.FloatType{Size: 64}}, true) {
		t.Errorf("expected true for float param with untyped float arg (same category)")
	}
}

func TestArgMatchCategoryStaticFloatWithFloatTyped(t *testing.T) {
	if !argMatchCategoryStatic(ast.FloatType{Size: 64}, knownFloatValue(1.5, ast.FloatType{Size: 64}), true) {
		t.Errorf("expected true for float param with typed float arg")
	}
}

func TestArgMatchCategoryStaticBoolMatch(t *testing.T) {
	if !argMatchCategoryStatic(ast.BoolType{}, knownBoolValue(true, ast.BoolType{}), true) {
		t.Errorf("expected true for bool param with bool arg")
	}
}

func TestArgMatchCategoryStaticStringMatch(t *testing.T) {
	if !argMatchCategoryStatic(ast.StringType{}, AbsValue{kind: AbsString}, true) {
		t.Errorf("expected true for string param with string arg")
	}
}

func TestArgMatchCategoryStaticArrayMatch(t *testing.T) {
	if !argMatchCategoryStatic(ast.ArrayType{}, AbsValue{kind: AbsArray}, true) {
		t.Errorf("expected true for array param with array arg")
	}
}

func TestArgMatchCategoryStaticListMatch(t *testing.T) {
	if !argMatchCategoryStatic(ast.ListType{}, AbsValue{kind: AbsList}, true) {
		t.Errorf("expected true for list param with list arg")
	}
}

// -- Coverage: analyzeBinaryExpr union ops --
func TestAnalyzeBinaryExprUnionAdd(t *testing.T) {
	a := New()
	program := parseProgram(t, `var x: int{size: 32} | float{size: 64} = 1; var y: int{size: 32} | float{size: 64} = 2; print(x + y);`)
	a.Analyze(program)
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for union add, got %v", a.Errors())
	}
}

// -- Coverage: binaryOpResultType all paths --
func TestBinaryOpResultTypeIntIntMatch(t *testing.T) {
	result := binaryOpResultType(ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 64, Signed: true}, "+")
	if result == nil {
		t.Errorf("expected non-nil result for int+int")
	}
}

func TestBinaryOpResultTypeIntFloat(t *testing.T) {
	result := binaryOpResultType(ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{Size: 64}, "+")
	if result == nil {
		t.Errorf("expected non-nil result for int+float")
	}
}

func TestBinaryOpResultTypeFloatInt(t *testing.T) {
	result := binaryOpResultType(ast.FloatType{Size: 64}, ast.IntegerType{Size: 32, Signed: true}, "+")
	if result == nil {
		t.Errorf("expected non-nil result for float+int")
	}
}

func TestBinaryOpResultTypeFloatFloat(t *testing.T) {
	result := binaryOpResultType(ast.FloatType{Size: 32}, ast.FloatType{Size: 64}, "+")
	if result == nil {
		t.Errorf("expected non-nil result for float+float")
	}
}

func TestBinaryOpResultTypeFloatFloatLarger(t *testing.T) {
	result := binaryOpResultType(ast.FloatType{Size: 64}, ast.FloatType{Size: 32}, "+")
	if result == nil {
		t.Errorf("expected non-nil result for larger float+smaller float")
	}
}

func TestBinaryOpResultTypeNoMatch(t *testing.T) {
	result := binaryOpResultType(ast.BoolType{}, ast.BoolType{}, "+")
	if result != nil {
		t.Errorf("expected nil for bool+bool, got %v", result)
	}
}

// -- Coverage: hasNullLiteralArg false path --
func TestHasNullLiteralArgEmpty(t *testing.T) {
	if hasNullLiteralArg(nil, nil) {
		t.Errorf("expected false for empty args")
	}
}

// -- Coverage: checkAssignmentType union with incompatible types (error path) --
func TestCheckAssignmentTypeUnionWithIncompatible(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x",
		AbsValue{
			kind:       AbsUnion,
			unionTypes: []ast.Type{ast.FloatType{Size: 64}},
		},
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true}},
	)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for union{float} assigned to int, got none")
	}
}

// -- Coverage: argMatchCategoryStatic FloatType with typed int cross-category through canImplicitConvert --
func TestArgMatchCategoryStaticIntCrossToFloatConvertible(t *testing.T) {
	// int8 can fit in float64 -> should match cross-category
	if !argMatchCategoryStatic(ast.FloatType{Size: 64}, knownIntValue(5, ast.IntegerType{Size: 8, Signed: true}), false) {
		t.Errorf("expected true for int8 arg to float64 param (cross category)")
	}
}

func TestArgMatchCategoryStaticIntCrossToFloatNotConvertible(t *testing.T) {
	// uint64 cannot fit in float64 -> should NOT match cross-category
	if argMatchCategoryStatic(ast.FloatType{Size: 64}, knownIntValue(5, ast.IntegerType{Size: 64, Signed: false}), false) {
		t.Errorf("expected false for uint64 arg to float64 param (cross category)")
	}
}

// -- Coverage: binaryOpResultType leftInt && rightInt conversion first then second --
func TestBinaryOpResultTypeIntIntFirstConvert(t *testing.T) {
	// 32-bit fits in 64-bit, so ri (64) should win
	result := binaryOpResultType(ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 64, Signed: true}, "+")
	if result == nil {
		t.Errorf("expected non-nil for 32+64")
	}
}

func TestBinaryOpResultTypeIntIntNoConversion(t *testing.T) {
	// uint64 can't implicitly convert to int32 and vice versa
	result := binaryOpResultType(ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 64, Signed: false}, "+")
	if result != nil {
		t.Errorf("expected nil for incompatible int+int, got %v", result)
	}
}

// -- Coverage: binaryOpResultType leftFloat && rightInt --
func TestBinaryOpResultTypeFloatIntNoConvert(t *testing.T) {
	// uint64 can't implicitly convert to float32
	result := binaryOpResultType(ast.FloatType{Size: 32}, ast.IntegerType{Size: 64, Signed: false}, "+")
	if result != nil {
		t.Errorf("expected nil for float32+uint64")
	}
}

// -- Coverage: checkDivisionByZero with string or non-number arg --
func TestDivisionByZeroNonNumeric(t *testing.T) {
	a := New()
	program := parseProgram(t, `var x: string = "hello"; var y: int{size: 32} = 10 / x;`)
	a.Analyze(program)
	// Division by zero check should not produce false positive for non-numeric types
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for division by string, got: %v", a.Errors())
	}
}

// -- Coverage: argExactTypeMatchStatic with union of various types --
func TestArgExactTypeMatchStaticUnionMultipleTypes(t *testing.T) {
	if !argExactTypeMatchStatic(
		ast.UnionType{Types: []ast.Type{ast.BoolType{}, ast.IntegerType{Size: 32, Signed: true}}},
		knownBoolValue(true, ast.BoolType{}),
	) {
		t.Errorf("expected true for bool arg matching union{bool, int}")
	}
}

// -- Coverage: typeDescFromType union with multiple types --
func TestTypeDescFromTypeUnionMulti(t *testing.T) {
	s := typeDescFromType(ast.UnionType{
		Types: []ast.Type{
			ast.IntegerType{Size: 32, Signed: true, Nullable: false},
			ast.FloatType{Size: 64, Nullable: true},
		},
	})
	if s == "" {
		t.Errorf("expected non-empty union description")
	}
}

// -- Coverage: typeDescFromType with BoolType nullable --
func TestTypeDescFromTypeBool(t *testing.T) {
	s := typeDescFromType(ast.BoolType{Nullable: true})
	if s != "bool{nullable: true}" {
		t.Errorf("expected 'bool{nullable: true}', got %q", s)
	}
}

// -- Coverage: checkAssignmentType with 2+ incompatible types (,  separator) --
func TestCheckAssignmentTypeUnionWithTwoIncompatible(t *testing.T) {
	a := New()
	a.checkAssignmentType(1, "x",
		AbsValue{
			kind:       AbsUnion,
			unionTypes: []ast.Type{ast.BoolType{}, ast.StringType{Size: 10}},
		},
		&ast.VarDecl{Name: "x", IType: ast.IntegerType{Size: 32, Signed: true}},
	)
	if len(a.Errors()) == 0 {
		t.Errorf("expected error for union{bool, string} assigned to int")
	}
	msg := a.Errors()[0]
	if !strings.Contains(msg, ", ") {
		t.Errorf("expected comma-separated incompatible types, got %q", msg)
	}
}

// -- Coverage: typeDescFromDecl with 2+ union types (| separator) --
func TestTypeDescFromDeclUnionMultiType(t *testing.T) {
	s := typeDescFromDecl(&ast.VarDecl{
		IsUnion: true,
		UnionType: ast.UnionType{
			Types: []ast.Type{
				ast.IntegerType{Size: 32, Signed: true, Nullable: false},
				ast.FloatType{Size: 64, Nullable: true},
			},
		},
	})
	if !strings.Contains(s, " | ") {
		t.Errorf("expected pipe-separated union types, got %q", s)
	}
}

// -- Coverage: countMatchingOverloads same-category path (typed int to different-sized int) --
func TestCountMatchingOverloadsSameCategoryIntToInt(t *testing.T) {
	overloads := []*ast.FuncDecl{
		{Parameters: []ast.Param{{Name: "x", Type: ast.IntegerType{Size: 64, Signed: true}}}},
	}
	// int8 arg: exact fails (8!=64), same-category succeeds (8<=64)
	count := countMatchingOverloads(overloads, []AbsValue{knownIntValue(5, ast.IntegerType{Size: 8, Signed: true})})
	if count != 1 {
		t.Errorf("expected 1 same-category match, got %d", count)
	}
}

// -- Coverage: argExactTypeMatchStatic with nil paramType (default branch) --
func TestArgExactTypeMatchStaticDefaultNil(t *testing.T) {
	// nil paramType does not match any switch case -> hits default return false
	if argExactTypeMatchStatic(nil, AbsValue{kind: AbsInt}) {
		t.Errorf("expected false for nil paramType")
	}
}

// -- Coverage: argMatchCategoryStatic with nil paramType (default branch) --
func TestArgMatchCategoryStaticDefaultNil(t *testing.T) {
	if argMatchCategoryStatic(nil, AbsValue{kind: AbsInt}, true) {
		t.Errorf("expected false for nil paramType")
	}
}

// -- Coverage: analyzeBinaryExpr union branch (union vars without initializer) --
func TestAnalyzeBinaryExprUnionNoInit(t *testing.T) {
	a := New()
	program := parseProgram(t, `var x: int{size: 32} | float{size: 64}; var y: int{size: 32} | float{size: 64}; print(x + y);`)
	a.Analyze(program)
	// Should not produce errors for union ops
	if len(a.Errors()) > 0 {
		t.Errorf("expected no errors for union ops without init, got %v", a.Errors())
	}
}

// -- Coverage: analyzeCall ambiguous null --
func TestAnalyzeCallAmbiguousNull(t *testing.T) {
	a := New()
	program := parseProgram(t, `
function foo(x: int{size: 32, nullable: true}): int { return x; }
function foo(x: int{size: 64, nullable: true}): int { return x; }
foo(null);
`)
	a.Analyze(program)
	found := false
	for _, err := range a.Errors() {
		if strings.Contains(err, "ambiguous") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ambiguous call error, got: %v", a.Errors())
	}
}


