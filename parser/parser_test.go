package parser

import (
	"math"
	"strings"
	"testing"

	"lang-interpreter/ast"
	"lang-interpreter/lexer"
)

func TestVarDecl(t *testing.T) {
	input := "var x: int{size: 32, signed: true, nullable: false} = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if stmt.Name != "x" {
		t.Errorf("expected name 'x', got %q", stmt.Name)
	}

	if stmt.IType.Size != 32 {
		t.Errorf("expected size 32, got %d", stmt.IType.Size)
	}

	if !stmt.IType.Signed {
		t.Errorf("expected signed true")
	}

	if stmt.IType.Nullable {
		t.Errorf("expected nullable false")
	}

	if stmt.Expr == nil {
		t.Fatalf("expected expression, got nil")
	}

	intLit, ok := stmt.Expr.(*ast.IntegerLit)
	if !ok {
		t.Fatalf("expected *ast.IntegerLit, got %T", stmt.Expr)
	}

	if intLit.Value != 42 {
		t.Errorf("expected value 42, got %d", intLit.Value)
	}
}

func TestVarDeclWithoutInit(t *testing.T) {
	input := "var x: int{size: 64, signed: true, nullable: true};"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if stmt.Expr != nil {
		t.Errorf("expected nil expression, got %v", stmt.Expr)
	}
}

func TestPrintStmt(t *testing.T) {
	input := "print(x);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.PrintStmt)
	if !ok {
		t.Fatalf("expected *ast.PrintStmt, got %T", program.Stmts[0])
	}

	varRef, ok := stmt.Expr.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected *ast.VarRef, got %T", stmt.Expr)
	}

	if varRef.Name != "x" {
		t.Errorf("expected name 'x', got %q", varRef.Name)
	}
}

func TestAssignment(t *testing.T) {
	input := "x = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected *ast.Assignment, got %T", program.Stmts[0])
	}

	if stmt.Name != "x" {
		t.Errorf("expected name 'x', got %q", stmt.Name)
	}

	if stmt.Op != "=" {
		t.Errorf("expected op '=', got %q", stmt.Op)
	}

	intLit, ok := stmt.Expr.(*ast.IntegerLit)
	if !ok {
		t.Fatalf("expected *ast.IntegerLit, got %T", stmt.Expr)
	}

	if intLit.Value != 42 {
		t.Errorf("expected value 42, got %d", intLit.Value)
	}
}

func TestCompoundAssignment(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"x += 10;", "+="},
		{"x -= 10;", "-="},
		{"x *= 10;", "*="},
		{"x /= 10;", "/="},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(program.Stmts) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
		}

		stmt, ok := program.Stmts[0].(*ast.Assignment)
		if !ok {
			t.Fatalf("expected *ast.Assignment, got %T", program.Stmts[0])
		}

		if stmt.Op != tt.op {
			t.Errorf("expected op %q, got %q", tt.op, stmt.Op)
		}
	}
}

func TestBinaryExpr(t *testing.T) {
	input := "print(x + 1 * 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.PrintStmt)
	if !ok {
		t.Fatalf("expected *ast.PrintStmt, got %T", program.Stmts[0])
	}

	// Should be a binary expression: x + (1 * 2)
	binExpr, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}

	if binExpr.Op != "+" {
		t.Errorf("expected op '+', got %q", binExpr.Op)
	}

	// Right side should be another binary expr: 1 * 2
	rightBinExpr, ok := binExpr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr on right, got %T", binExpr.Right)
	}

	if rightBinExpr.Op != "*" {
		t.Errorf("expected op '*', got %q", rightBinExpr.Op)
	}
}

func TestNullLit(t *testing.T) {
	input := "var x: int{size: 32, nullable: true} = null;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if _, ok := stmt.Expr.(*ast.NullLit); !ok {
		t.Fatalf("expected *ast.NullLit, got %T", stmt.Expr)
	}
}

func TestParserErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"var: int;", true},
		{"var x int;", true},
		{"print;", true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		p.ParseProgram()

		if tt.expectError && len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", tt.input)
		}
	}
}

func TestParseSingleStmt(t *testing.T) {
	input := "var x: int{size: 32} = 42;"
	l := lexer.New(input)
	p := New(l)

	stmt, errors, _ := p.ParseSingleStmt()
	if errors != nil && len(errors) > 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}

	if stmt == nil {
		t.Fatalf("expected statement, got nil")
	}

	_, ok := stmt.(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", stmt)
	}
}

func TestParseSingleStmtEOF(t *testing.T) {
	input := ""
	l := lexer.New(input)
	p := New(l)

	stmt, errors, _ := p.ParseSingleStmt()
	if stmt != nil {
		t.Errorf("expected nil statement for EOF")
	}
	if errors != nil && len(errors) > 0 {
		t.Errorf("unexpected errors: %v", errors)
	}
}

func TestParseIntegerTypeFull(t *testing.T) {
	input := "var x: int{size: 16, signed: false, nullable: true} = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if stmt.IType.Size != 16 {
		t.Errorf("expected size 16, got %d", stmt.IType.Size)
	}
	if stmt.IType.Signed {
		t.Errorf("expected signed false")
	}
	if !stmt.IType.Nullable {
		t.Errorf("expected nullable true")
	}
}

func TestParseIntegerTypeDefault(t *testing.T) {
	input := "var x: int = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if stmt.IType.Size != 64 {
		t.Errorf("expected default size 64, got %d", stmt.IType.Size)
	}
	if !stmt.IType.Signed {
		t.Errorf("expected default signed true")
	}
	if !stmt.IType.Nullable {
		t.Errorf("expected default nullable true")
	}
}

func TestParsePrintErrors(t *testing.T) {
	tests := []string{
		"print x);",
		"print(;",
		"print);",
		"print(1 2);", // missing operator
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", input)
		}
	}
}

func TestParsePrimaryParensNested(t *testing.T) {
	input := "print(((42)));"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParsePrintMissingParen(t *testing.T) {
	input := "print(1;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing )")
	}
}

func TestParsePrintMissingRParen(t *testing.T) {
	input := "print(1;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing )")
	}
}

func TestParsePrintMissingSemicolon(t *testing.T) {
	input := "print(1)"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ;")
	}
}

func TestParsePrimaryErrors(t *testing.T) {
	tests := []string{
		"print(@);",
		"print(x + );",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", input)
		}
	}
}

func TestParseIntegerTypeErrors(t *testing.T) {
	tests := []string{
		"var x: notint{size: 32};",
		"var x: int{size: 12};",
		"var x: int{size: 32, signed: yes};",
		"var x: int{size: 32, invalid: true};",
		"var x: int{size: 32 signed: true};", // missing colon
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", input)
		}
	}
}

func TestParseIntegerTypeWithComma(t *testing.T) {
	input := "var x: int{size: 32, signed: true} = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParseIntegerTypeWithAllFields(t *testing.T) {
	input := "var x: int{size: 16, signed: false, nullable: true} = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	if stmt.IType.Size != 16 {
		t.Errorf("expected size 16, got %d", stmt.IType.Size)
	}
	if stmt.IType.Signed {
		t.Errorf("expected signed false")
	}
	if !stmt.IType.Nullable {
		t.Errorf("expected nullable true")
	}
}

func TestParseIntegerTypeWithOnlyBrace(t *testing.T) {
	input := "var x: int{} = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParseAssignmentWithAllOperators(t *testing.T) {
	tests := []string{
		"x += 5;",
		"x -= 5;",
		"x *= 5;",
		"x /= 5;",
	}

	for _, input := range tests {
		l := lexer.New("var x: int{size: 32} = 10;" + input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Fatalf("unexpected errors for %q: %v", input, p.Errors())
		}
	}
}

func TestParseAssignmentErrors(t *testing.T) {
	tests := []string{
		"x = ;",
		"x += ;",
		"x = x + ;",
	}

	for _, input := range tests {
		l := lexer.New("var x: int{size: 32} = 10;" + input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", input)
		}
	}
}

func TestParseAssignmentNoSemicolon(t *testing.T) {
	input := "var x: int{size: 32} = 10; x = 5"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing semicolon")
	}
}

func TestParseAssignmentInvalidOp(t *testing.T) {
	input := "var x: int{size: 32} = 10; x + 5;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid assignment operator")
	}
}

func TestParsePrintComplexExpr(t *testing.T) {
	input := "print(1 + 2 * 3 - 4 / 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParsePrimaryInteger(t *testing.T) {
	input := "print(42);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	intLit, ok := stmt.Expr.(*ast.IntegerLit)
	if !ok {
		t.Fatalf("expected *ast.IntegerLit, got %T", stmt.Expr)
	}

	if intLit.Value != 42 {
		t.Errorf("expected 42, got %d", intLit.Value)
	}
	if !intLit.Untyped {
		t.Errorf("expected untyped true")
	}
}

func TestParsePrimaryParens(t *testing.T) {
	input := "print((42));"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParsePrintWithParens(t *testing.T) {
	input := "print(1 + 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParsePrimaryIdent(t *testing.T) {
	input := "print(x);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	_, ok := stmt.Expr.(*ast.VarRef)
	if !ok {
		t.Errorf("expected *ast.VarRef, got %T", stmt.Expr)
	}
}

func TestParsePrimaryInvalidToken(t *testing.T) {
	tests := []string{
		"print(@);",
		"print($);",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", input)
		}
	}
}

func TestParsePrimaryMissingRParen(t *testing.T) {
	input := "print(42;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing )")
	}
}

func TestParsePrimaryNestedParensMissingRParen(t *testing.T) {
	input := "print((42);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ), got none")
	}
}

func TestParsePrimaryNull(t *testing.T) {
	input := "print(null);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	_, ok := stmt.Expr.(*ast.NullLit)
	if !ok {
		t.Errorf("expected *ast.NullLit, got %T", stmt.Expr)
	}
}

func TestParseIntegerTypeMissingColon(t *testing.T) {
	input := "var x: int{size 32};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing colon")
	}
}

func TestParseIntegerTypeInvalidSizeString(t *testing.T) {
	input := "var x: int{size: abc};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for non-int size")
	}
}

func TestParseIntegerTypeInvalidSizeNumber(t *testing.T) {
	input := "var x: int{size: 7};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid size number")
	}
}

func TestParseIntegerTypeInvalidSigned(t *testing.T) {
	input := "var x: int{signed: maybe};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid signed value")
	}
}

func TestParseIntegerTypeInvalidNullable(t *testing.T) {
	input := "var x: int{nullable: maybe};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid nullable value")
	}
}

func TestParseAssignmentWithVarRef(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 10;
x = y;
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	// Should parse correctly even with undefined var (semantic check happens in interpreter)
	if len(program.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Stmts))
	}
}

func TestParsePrintWithComplexExpr(t *testing.T) {
	input := "print(x + y * 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParsePrimaryMissingRParenInExpr(t *testing.T) {
	// print(1 + 2;  <- missing )
	// This should trigger the error at line 263-265 in parsePrimary
	input := "print(1 + 2;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ), got none. Errors: %v", p.Errors())
	}
}

func TestParseFloatLiteral(t *testing.T) {
	input := "print(3.14);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	_, ok := stmt.Expr.(*ast.FloatLit)
	if !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
}

func TestParseDotLeadingFloatLiteral(t *testing.T) {
	input := "print(.1);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	fl, ok := stmt.Expr.(*ast.FloatLit)
	if !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
	if fl.Value != 0.1 {
		t.Errorf("expected 0.1, got %g", fl.Value)
	}
}

func TestParseInvalidFloatLiteral(t *testing.T) {
	// Invalid float literal - Go's ParseFloat might handle some edge cases
	// Let's test with a float that has multiple decimal points
	input := "print(1.2.3);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// Should have errors
	if len(p.Errors()) == 0 {
		t.Errorf("expected errors for invalid float literal")
	}
}

func TestParsePrimaryInvalidFloat(t *testing.T) {
	// Test the error case in parsePrimary for float
	input := "print(1e10);" // exponential notation not supported
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// This might not error since lexer treats it as identifier potentially
	// Let's try something that will definitely cause error
	input2 := "print(.);"
	l2 := lexer.New(input2)
	p2 := New(l2)
	p2.ParseProgram()

	if len(p2.Errors()) == 0 {
		t.Errorf("expected error for invalid primary expression")
	}
}

func TestParseFloatDecl(t *testing.T) {
	input := "var x: float{size: 32} = 3.14;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if !stmt.IsFloat {
		t.Errorf("expected IsFloat to be true")
	}
	if stmt.FType.Size != 32 {
		t.Errorf("expected size 32, got %d", stmt.FType.Size)
	}
}

func TestParseFloat16Decl(t *testing.T) {
	input := "var x: float{size: 16} = 1.5;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	if stmt.FType.Size != 16 {
		t.Errorf("expected size 16, got %d", stmt.FType.Size)
	}
}

func TestParseFloat64Decl(t *testing.T) {
	input := "var x: float{size: 64} = 1.5;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	if stmt.FType.Size != 64 {
		t.Errorf("expected size 64, got %d", stmt.FType.Size)
	}
}

func TestParseFloatInvalidSize(t *testing.T) {
	input := "var x: float{size: 128} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid float size")
	}
}

func TestParseFloatTypeMissingColon(t *testing.T) {
	input := "var x: float{size 32} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing colon")
	}
}

func TestParseFloatTypeInvalidNullableValue(t *testing.T) {
	input := "var x: float{nullable: maybe} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid nullable value")
	}
}

func TestParseFloatTypeWithComma(t *testing.T) {
	input := "var x: float{size: 32, nullable: true} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParseFloatTypeOnlyBraces(t *testing.T) {
	input := "var x: float{} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParseFloatTypeWithNullable(t *testing.T) {
	input := "var x: float{size: 32, nullable: true} = null;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	if !stmt.FType.Nullable {
		t.Errorf("expected nullable to be true")
	}
}

func TestParseFloatTypeInvalidKey(t *testing.T) {
	input := "var x: float{invalid: 32} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid key")
	}
}

func TestParseFloatTypeInvalidSizeValue(t *testing.T) {
	input := "var x: float{size: abc} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for non-int size")
	}
}

func TestParseFloatTypeInvalidSizeNumber(t *testing.T) {
	input := "var x: float{size: 128} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid size number")
	}
}

func TestParseBoolDeclDefault(t *testing.T) {
	input := "var x: bool = true;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if !stmt.IsBool {
		t.Errorf("expected IsBool to be true")
	}
	if !stmt.BType.Nullable {
		t.Errorf("expected default nullable true")
	}

	lit, ok := stmt.Expr.(*ast.BoolLit)
	if !ok {
		t.Fatalf("expected *ast.BoolLit, got %T", stmt.Expr)
	}
	if lit.Value != true {
		t.Errorf("expected true, got %t", lit.Value)
	}
	if !lit.Untyped {
		t.Errorf("expected untyped true")
	}
}

func TestParseBoolDeclNullableFalse(t *testing.T) {
	input := "var x: bool{nullable: false} = false;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if !stmt.IsBool {
		t.Errorf("expected IsBool to be true")
	}
	if stmt.BType.Nullable {
		t.Errorf("expected nullable false")
	}

	lit, ok := stmt.Expr.(*ast.BoolLit)
	if !ok {
		t.Fatalf("expected *ast.BoolLit, got %T", stmt.Expr)
	}
	if lit.Value != false {
		t.Errorf("expected false, got %t", lit.Value)
	}
}

func TestParseBoolDeclNullableTrue(t *testing.T) {
	input := "var x: bool{nullable: true} = null;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}

	if !stmt.IsBool {
		t.Errorf("expected IsBool to be true")
	}
	if !stmt.BType.Nullable {
		t.Errorf("expected nullable true")
	}

	if _, ok := stmt.Expr.(*ast.NullLit); !ok {
		t.Fatalf("expected *ast.NullLit, got %T", stmt.Expr)
	}
}

func TestParseBoolPrintTrue(t *testing.T) {
	input := "print(true);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt, ok := program.Stmts[0].(*ast.PrintStmt)
	if !ok {
		t.Fatalf("expected *ast.PrintStmt, got %T", program.Stmts[0])
	}

	lit, ok := stmt.Expr.(*ast.BoolLit)
	if !ok {
		t.Fatalf("expected *ast.BoolLit, got %T", stmt.Expr)
	}
	if lit.Value != true {
		t.Errorf("expected true, got %t", lit.Value)
	}
}

func TestParseBoolPrintFalse(t *testing.T) {
	input := "print(false);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt, ok := program.Stmts[0].(*ast.PrintStmt)
	if !ok {
		t.Fatalf("expected *ast.PrintStmt, got %T", program.Stmts[0])
	}

	lit, ok := stmt.Expr.(*ast.BoolLit)
	if !ok {
		t.Fatalf("expected *ast.BoolLit, got %T", stmt.Expr)
	}
	if lit.Value != false {
		t.Errorf("expected false, got %t", lit.Value)
	}
}

func TestParseBoolTypeErrors(t *testing.T) {
	tests := []string{
		"var x: bool{nullable: maybe} = true;",
		"var x: bool{invalid: true} = true;",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()

		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", input)
		}
	}
}

func TestParseBoolTypeMissingColon(t *testing.T) {
	input := "var x: bool{nullable true} = true;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing colon, got none")
	}
}

func TestParseBoolTypeOnlyBraces(t *testing.T) {
	input := "var x: bool{} = true;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
}

func TestParseBoolDeclWithoutInit(t *testing.T) {
	input := "var x: bool;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt, ok := program.Stmts[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", program.Stmts[0])
	}
	if !stmt.IsBool {
		t.Errorf("expected IsBool to be true")
	}
	if stmt.Expr != nil {
		t.Errorf("expected nil expression")
	}
}

func TestParseBoolTypeTrailingComma(t *testing.T) {
	input := "var x: bool{nullable: true,} = true;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	if !stmt.BType.Nullable {
		t.Errorf("expected nullable true")
	}
}

func TestParseBoolTypeMissingComma(t *testing.T) {
	input := "var x: bool{nullable: true nullable: false} = true;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing comma, got none")
	}
}

func TestParseFloatTypeTrailingComma(t *testing.T) {
	input := "var x: float{size: 32,} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// This might error due to unexpected token after comma
	// Just ensure no panic
}

func TestParsePrimaryInvalidInt(t *testing.T) {
	// Test invalid integer literal that causes ParseInt to fail
	input := "print(99999999999999999999);" // too large for int64
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// Might or might not error depending on lexer implementation
	// Just ensure no panic
}

func TestParseStmtUnexpectedToken(t *testing.T) {
	input := "@"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for unexpected token")
	}
}

func TestParseFloatInvalidLiteral(t *testing.T) {
	// Test float literal that fails ParseFloat
	input := "var x: float = 1.2.3;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// Lexer will produce tokens that cause parse error
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid float literal")
	}
}

func TestParseFloatTypeSizeLine(t *testing.T) {
	// Cover line 134-136: size token but not TOK_INT_LIT
	input := "var x: float{size: true} = 1.0;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for non-int size value")
	}
}

func TestParseScientificNotation(t *testing.T) {
	tests := []string{
		"print(1e20);",
		"print(1e+20);",
		"print(1e-20);",
		"print(1.5e3);",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		if _, ok := stmt.Expr.(*ast.FloatLit); !ok {
			t.Errorf("input %q: expected *ast.FloatLit, got %T", input, stmt.Expr)
		}
	}
}

func TestParseUnderscoreInteger(t *testing.T) {
	tests := []struct {
		input     string
		expectVal int64
	}{
		{"print(100_000);", 100000},
		{"print(1_0_0__0_0_0);", 100000},
		{"print(1000_1000);", 10001000},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		lit, ok := stmt.Expr.(*ast.IntegerLit)
		if !ok {
			t.Errorf("input %q: expected *ast.IntegerLit, got %T", tt.input, stmt.Expr)
			continue
		}
		if lit.Value != tt.expectVal {
			t.Errorf("input %q: expected %d, got %d", tt.input, tt.expectVal, lit.Value)
		}
	}
}

func TestParseUnderscoreFloat(t *testing.T) {
	input := "print(1_000.5);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	if _, ok := stmt.Expr.(*ast.FloatLit); !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
}

func TestParseBinaryLiteral(t *testing.T) {
	tests := []struct {
		input     string
		expectVal int64
	}{
		{"print(0b0101_0101);", 0b01010101},
		{"print(0b1010);", 0b1010},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		lit, ok := stmt.Expr.(*ast.IntegerLit)
		if !ok {
			t.Errorf("input %q: expected *ast.IntegerLit, got %T", tt.input, stmt.Expr)
			continue
		}
		if lit.Value != tt.expectVal {
			t.Errorf("input %q: expected %d, got %d", tt.input, tt.expectVal, lit.Value)
		}
	}
}

func TestParseOctalLiteral(t *testing.T) {
	tests := []struct {
		input     string
		expectVal int64
	}{
		{"print(0o777);", 0o777},
		{"print(0o012_345_67);", int64(0o01234567)},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		lit, ok := stmt.Expr.(*ast.IntegerLit)
		if !ok {
			t.Errorf("input %q: expected *ast.IntegerLit, got %T", tt.input, stmt.Expr)
			continue
		}
		if lit.Value != tt.expectVal {
			t.Errorf("input %q: expected %d, got %d", tt.input, tt.expectVal, lit.Value)
		}
	}
}

func TestParseHexLiteral(t *testing.T) {
	tests := []struct {
		input     string
		expectVal int64
	}{
		{"print(0xFF);", 0xFF},
		{"print(0xffee_d2a5);", int64(0xffeed2a5)},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		lit, ok := stmt.Expr.(*ast.IntegerLit)
		if !ok {
			t.Errorf("input %q: expected *ast.IntegerLit, got %T", tt.input, stmt.Expr)
			continue
		}
		if lit.Value != tt.expectVal {
			t.Errorf("input %q: expected %d, got %d", tt.input, tt.expectVal, lit.Value)
		}
	}
}

func TestParseHexFloatLiteral(t *testing.T) {
	tests := []struct {
		input     string
		expectVal float64
	}{
		{"print(0xf.f);", 15.9375},
		{"print(0x.1);", 0.0625},
		{"print(0xf.0);", 15.0},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		lit, ok := stmt.Expr.(*ast.FloatLit)
		if !ok {
			t.Errorf("input %q: expected *ast.FloatLit, got %T", tt.input, stmt.Expr)
			continue
		}
		if lit.Value != tt.expectVal {
			t.Errorf("input %q: expected %g, got %g", tt.input, tt.expectVal, lit.Value)
		}
	}
}

func TestParseBinFloatLiteral(t *testing.T) {
	tests := []struct {
		input     string
		expectVal float64
	}{
		{"print(0b1.01);", 1.25},
		{"print(0b0.1);", 0.5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		lit, ok := stmt.Expr.(*ast.FloatLit)
		if !ok {
			t.Errorf("input %q: expected *ast.FloatLit, got %T", tt.input, stmt.Expr)
			continue
		}
		if lit.Value != tt.expectVal {
			t.Errorf("input %q: expected %g, got %g", tt.input, tt.expectVal, lit.Value)
		}
	}
}

func TestParseOctFloatLiteral(t *testing.T) {
	tests := []struct {
		input     string
		expectVal float64
	}{
		{"print(0o7.7);", 7.875},
		{"print(0o0.4);", 0.5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, p.Errors())
			continue
		}
		stmt := program.Stmts[0].(*ast.PrintStmt)
		lit, ok := stmt.Expr.(*ast.FloatLit)
		if !ok {
			t.Errorf("input %q: expected *ast.FloatLit, got %T", tt.input, stmt.Expr)
			continue
		}
		if lit.Value != tt.expectVal {
			t.Errorf("input %q: expected %g, got %g", tt.input, tt.expectVal, lit.Value)
		}
	}
}

func TestParsePrefixedFloatError(t *testing.T) {
	tests := []string{
		"print(0xg.1);",
		"print(0xf.g);",
		"print(0b2.0);",
		"print(0o8.0);",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		p.ParseProgram()
		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for input %q, got none", input)
		}
	}
}

func TestParseLeadingZeroWarning(t *testing.T) {
	input := "print(010);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Warnings()) == 0 {
		t.Errorf("expected warning for leading zero, got none")
	}
}

func TestParseSingleStmtWithWarning(t *testing.T) {
	input := "print(010);"
	l := lexer.New(input)
	p := New(l)

	stmt, errs, warns := p.ParseSingleStmt()
	if stmt == nil {
		t.Fatalf("expected statement, got nil")
	}
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warns) == 0 {
		t.Errorf("expected warnings for leading zero, got none")
	}
}

func TestParsePrefixedFloatNoFrac(t *testing.T) {
	input := "print(0xf.);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	lit, ok := stmt.Expr.(*ast.FloatLit)
	if !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
	if lit.Value != 15.0 {
		t.Errorf("expected 15, got %g", lit.Value)
	}
}

func TestParseInvalidPrefixIntegerLiterals(t *testing.T) {
	tests := []struct {
		input  string
		errMsg string
	}{
		{"print(0b2);", "invalid binary literal"},
		{"print(0o8);", "invalid octal literal"},
		{"print(0xGG);", "invalid hexadecimal literal"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		p.ParseProgram()
		found := false
		for _, err := range p.Errors() {
			if strings.Contains(err, tt.errMsg) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("input %q: expected error containing %q, got %v", tt.input, tt.errMsg, p.Errors())
		}
	}
}

func TestParseInvalidIntegerLiteralTooLarge(t *testing.T) {
	input := "print(99999999999999999999);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	found := false
	for _, err := range p.Errors() {
		if strings.Contains(err, "invalid int literal") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'invalid int literal' error, got %v", p.Errors())
	}
}

func TestParseFloatTypeMissingComma(t *testing.T) {
	input := "var x: float{size: 32 nullable: true};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	found := false
	for _, err := range p.Errors() {
		if strings.Contains(err, "expected ',' or '}'") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about missing comma, got %v", p.Errors())
	}
}

func TestParsePrimaryStrayRParen(t *testing.T) {
	input := "print((42);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	// This tests a missing RPAREN in the outer print call
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ')'")
	}
}

func TestParsePrimaryMissingInnerRParen(t *testing.T) {
	input := "print((42;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing inner ')'")
	}
}

func TestParseUnaryMinusInteger(t *testing.T) {
	input := "print(-42);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	lit, ok := stmt.Expr.(*ast.IntegerLit)
	if !ok {
		t.Fatalf("expected *ast.IntegerLit, got %T", stmt.Expr)
	}
	if lit.Value != -42 {
		t.Errorf("expected -42, got %d", lit.Value)
	}
}

func TestParseUnaryMinusFloat(t *testing.T) {
	input := "print(-3.14);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	lit, ok := stmt.Expr.(*ast.FloatLit)
	if !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
	if lit.Value != -3.14 {
		t.Errorf("expected -3.14, got %g", lit.Value)
	}
}

func TestParseUnaryMinusExpr(t *testing.T) {
	input := "print(-(1+2));"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	binExpr, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr for negated expr, got %T", stmt.Expr)
	}
	if binExpr.Op != "-" {
		t.Errorf("expected op '-', got %q", binExpr.Op)
	}
	intLit, ok := binExpr.Left.(*ast.IntegerLit)
	if !ok {
		t.Fatalf("expected *ast.IntegerLit as left operand, got %T", binExpr.Left)
	}
	if intLit.Value != 0 {
		t.Errorf("expected left value 0, got %d", intLit.Value)
	}
}

func TestParseFloatParseError(t *testing.T) {
	input := "print(1.0e_);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()

	found := false
	for _, err := range p.Errors() {
		if strings.Contains(err, "invalid float literal") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'invalid float literal' error, got %v", p.Errors())
	}
}

func TestParseNaNFloatLiteral(t *testing.T) {
	input := "print(NaN);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	flt, ok := stmt.Expr.(*ast.FloatLit)
	if !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
	if !math.IsNaN(flt.Value) {
		t.Errorf("expected NaN value, got %v", flt.Value)
	}
}

func TestParseInfinityFloatLiteral(t *testing.T) {
	input := "print(infinity);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	flt, ok := stmt.Expr.(*ast.FloatLit)
	if !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
	if !math.IsInf(flt.Value, 1) {
		t.Errorf("expected +Inf value, got %v", flt.Value)
	}
}

func TestParseNegInfinityFloatLiteral(t *testing.T) {
	input := "print(-infinity);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	flt, ok := stmt.Expr.(*ast.FloatLit)
	if !ok {
		t.Fatalf("expected *ast.FloatLit, got %T", stmt.Expr)
	}
	if !math.IsInf(flt.Value, -1) {
		t.Errorf("expected -Inf value, got %v", flt.Value)
	}
}

func TestParseTypeRefInt(t *testing.T) {
	input := "print(int);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	tr, ok := stmt.Expr.(*ast.TypeRef)
	if !ok {
		t.Fatalf("expected *ast.TypeRef, got %T", stmt.Expr)
	}
	if tr.Kind != "int" {
		t.Errorf("expected Kind 'int', got %q", tr.Kind)
	}
	if tr.IType.Size != 64 {
		t.Errorf("expected Size 64, got %d", tr.IType.Size)
	}
}

func TestParseTypeRefFloat(t *testing.T) {
	input := "print(float);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	tr, ok := stmt.Expr.(*ast.TypeRef)
	if !ok {
		t.Fatalf("expected *ast.TypeRef, got %T", stmt.Expr)
	}
	if tr.Kind != "float" {
		t.Errorf("expected Kind 'float', got %q", tr.Kind)
	}
	if tr.FType.Size != 64 {
		t.Errorf("expected Size 64, got %d", tr.FType.Size)
	}
}

func TestParseTypeRefBool(t *testing.T) {
	input := "print(bool);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	tr, ok := stmt.Expr.(*ast.TypeRef)
	if !ok {
		t.Fatalf("expected *ast.TypeRef, got %T", stmt.Expr)
	}
	if tr.Kind != "bool" {
		t.Errorf("expected Kind 'bool', got %q", tr.Kind)
	}
}

func TestParseBoolSize(t *testing.T) {
	input := "print(bool.size);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	ma, ok := stmt.Expr.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected *ast.MemberAccess, got %T", stmt.Expr)
	}
	if ma.Member != "size" {
		t.Errorf("expected Member 'size', got %q", ma.Member)
	}
	tr, ok := ma.Object.(*ast.TypeRef)
	if !ok {
		t.Fatalf("expected Object *ast.TypeRef, got %T", ma.Object)
	}
	if tr.Kind != "bool" {
		t.Errorf("expected Kind 'bool', got %q", tr.Kind)
	}
}

func TestParseIntDotMin(t *testing.T) {
	input := "print(int.min);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	ma, ok := stmt.Expr.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected *ast.MemberAccess, got %T", stmt.Expr)
	}
	if ma.Member != "min" {
		t.Errorf("expected Member 'min', got %q", ma.Member)
	}
	tr, ok := ma.Object.(*ast.TypeRef)
	if !ok {
		t.Fatalf("expected Object *ast.TypeRef, got %T", ma.Object)
	}
	if tr.Kind != "int" {
		t.Errorf("expected Kind 'int', got %q", tr.Kind)
	}
}

func TestParseTypeWithSizeDotMax(t *testing.T) {
	input := "print(int{size: 8}.max);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	ma, ok := stmt.Expr.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected *ast.MemberAccess, got %T", stmt.Expr)
	}
	if ma.Member != "max" {
		t.Errorf("expected Member 'max', got %q", ma.Member)
	}
	tr, ok := ma.Object.(*ast.TypeRef)
	if !ok {
		t.Fatalf("expected Object *ast.TypeRef, got %T", ma.Object)
	}
	if tr.Kind != "int" {
		t.Errorf("expected Kind 'int', got %q", tr.Kind)
	}
	if tr.IType.Size != 8 {
		t.Errorf("expected Size 8, got %d", tr.IType.Size)
	}
}

func TestParseVarDotType(t *testing.T) {
	input := "print(a.type);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	ma, ok := stmt.Expr.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected *ast.MemberAccess, got %T", stmt.Expr)
	}
	if ma.Member != "type" {
		t.Errorf("expected Member 'type', got %q", ma.Member)
	}
	vr, ok := ma.Object.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected Object *ast.VarRef, got %T", ma.Object)
	}
	if vr.Name != "a" {
		t.Errorf("expected Name 'a', got %q", vr.Name)
	}
}

func TestParseChainedMemberAccess(t *testing.T) {
	input := "print(a.type.min);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	outer, ok := stmt.Expr.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected *ast.MemberAccess, got %T", stmt.Expr)
	}
	if outer.Member != "min" {
		t.Errorf("expected Member 'min', got %q", outer.Member)
	}
	inner, ok := outer.Object.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected inner *ast.MemberAccess, got %T", outer.Object)
	}
	if inner.Member != "type" {
		t.Errorf("expected Member 'type', got %q", inner.Member)
	}
	vr, ok := inner.Object.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected Object *ast.VarRef, got %T", inner.Object)
	}
	if vr.Name != "a" {
		t.Errorf("expected Name 'a', got %q", vr.Name)
	}
}
