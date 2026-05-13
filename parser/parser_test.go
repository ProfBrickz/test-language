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

func TestParseRefDecl(t *testing.T) {
	input := "ref a = b;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	stmt, ok := program.Stmts[0].(*ast.RefDecl)
	if !ok {
		t.Fatalf("expected *ast.RefDecl, got %T", program.Stmts[0])
	}
	if stmt.Name != "a" {
		t.Errorf("expected name 'a', got %q", stmt.Name)
	}
	if stmt.Type != nil {
		t.Errorf("expected nil type, got %T", stmt.Type)
	}
	varRef, ok := stmt.Expr.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected *ast.VarRef, got %T", stmt.Expr)
	}
	if varRef.Name != "b" {
		t.Errorf("expected VarRef 'b', got %q", varRef.Name)
	}
}

func TestParseRefDeclWithType(t *testing.T) {
	input := "ref a: int{size: 32, signed: true, nullable: false} = b;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	stmt, ok := program.Stmts[0].(*ast.RefDecl)
	if !ok {
		t.Fatalf("expected *ast.RefDecl, got %T", program.Stmts[0])
	}
	if stmt.Name != "a" {
		t.Errorf("expected name 'a', got %q", stmt.Name)
	}
	if stmt.Type == nil {
		t.Fatal("expected non-nil type")
	}
	it, ok := stmt.Type.(ast.IntegerType)
	if !ok {
		t.Fatalf("expected ast.IntegerType, got %T", stmt.Type)
	}
	if it.Size != 32 {
		t.Errorf("expected size 32, got %d", it.Size)
	}
	if !it.Signed {
		t.Errorf("expected signed true")
	}
	if it.Nullable {
		t.Errorf("expected nullable false")
	}
}

func TestParseRefDeclErrorNoName(t *testing.T) {
	input := "ref = b;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing name")
	}
}

func TestParseRefDeclErrorNoEq(t *testing.T) {
	input := "ref a b;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '='")
	}
}

func TestParseRefDeclErrorLiteral(t *testing.T) {
	input := "ref a = 5;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected parse error for ref with literal")
	}
}

func TestParseRefDeclErrorNoSemicolon(t *testing.T) {
	input := "ref a = b"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ';'")
	}
}

func TestParseAssignmentWithRef(t *testing.T) {
	input := "a = ref b;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	stmt, ok := program.Stmts[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected *ast.Assignment, got %T", program.Stmts[0])
	}
	if stmt.Name != "a" {
		t.Errorf("expected name 'a', got %q", stmt.Name)
	}
	if stmt.Op != "=" {
		t.Errorf("expected op '=', got %q", stmt.Op)
	}
	if !stmt.IsRef {
		t.Errorf("expected IsRef to be true")
	}
	varRef, ok := stmt.Expr.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected *ast.VarRef, got %T", stmt.Expr)
	}
	if varRef.Name != "b" {
		t.Errorf("expected VarRef 'b', got %q", varRef.Name)
	}
}

func TestParseAssignmentRefErrorNoName(t *testing.T) {
	input := "a = ref ;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing name after ref")
	}
}

func TestParseIsExpr(t *testing.T) {
	input := "print(a is b);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	ise, ok := stmt.Expr.(*ast.IsExpr)
	if !ok {
		t.Fatalf("expected *ast.IsExpr, got %T", stmt.Expr)
	}
	left, ok := ise.Left.(*ast.VarRef)
	if !ok || left.Name != "a" {
		t.Errorf("expected left VarRef 'a', got %T %+v", ise.Left, ise.Left)
	}
	right, ok := ise.Right.(*ast.VarRef)
	if !ok || right.Name != "b" {
		t.Errorf("expected right VarRef 'b', got %T %+v", ise.Right, ise.Right)
	}
}

func TestParseCopyExpr(t *testing.T) {
	input := "print(copy a);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	ce, ok := stmt.Expr.(*ast.CopyExpr)
	if !ok {
		t.Fatalf("expected *ast.CopyExpr, got %T", stmt.Expr)
	}
	varRef, ok := ce.Right.(*ast.VarRef)
	if !ok || varRef.Name != "a" {
		t.Errorf("expected VarRef 'a', got %T %+v", ce.Right, ce.Right)
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

func TestParseIfStmt(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
if (x) {
	print(x);
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Stmts))
	}

	ifStmt, ok := program.Stmts[1].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt, got %T", program.Stmts[1])
	}

	if ifStmt.Condition == nil {
		t.Fatal("expected condition, got nil")
	}

	if ifStmt.Then == nil {
		t.Fatal("expected then block, got nil")
	}

	if len(ifStmt.Then.Stmts) != 1 {
		t.Fatalf("expected 1 statement in then block, got %d", len(ifStmt.Then.Stmts))
	}

	if ifStmt.Else != nil {
		t.Errorf("expected nil else, got %T", ifStmt.Else)
	}
}

func TestParseIfElseStmt(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
if (x) {
	print(1);
} else {
	print(2);
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Stmts))
	}

	ifStmt, ok := program.Stmts[1].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt, got %T", program.Stmts[1])
	}

	elseBlock, ok := ifStmt.Else.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("expected *ast.BlockStmt for else, got %T", ifStmt.Else)
	}

	if len(elseBlock.Stmts) != 1 {
		t.Fatalf("expected 1 statement in else block, got %d", len(elseBlock.Stmts))
	}
}

func TestParseIfElseIfElseStmt(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
if (x) {
	print(1);
} else if (x) {
	print(2);
} else {
	print(3);
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	if len(program.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Stmts))
	}

	ifStmt, ok := program.Stmts[1].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt, got %T", program.Stmts[1])
	}

	elseIfStmt, ok := ifStmt.Else.(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt for else if, got %T", ifStmt.Else)
	}

	elseBlock, ok := elseIfStmt.Else.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("expected *ast.BlockStmt for final else, got %T", elseIfStmt.Else)
	}

	if len(elseBlock.Stmts) != 1 {
		t.Fatalf("expected 1 statement in final else block, got %d", len(elseBlock.Stmts))
	}
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
	if tr.Type.Kind() != "int" {
		t.Errorf("expected Kind 'int', got %q", tr.Type.Kind())
	}
	if tr.Type.(ast.IntegerType).Size != 64 {
		t.Errorf("expected Size 64, got %d", tr.Type.(ast.IntegerType).Size)
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
	if tr.Type.Kind() != "float" {
		t.Errorf("expected Kind 'float', got %q", tr.Type.Kind())
	}
	if tr.Type.(ast.FloatType).Size != 64 {
		t.Errorf("expected Size 64, got %d", tr.Type.(ast.FloatType).Size)
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
	if tr.Type.Kind() != "bool" {
		t.Errorf("expected Kind 'bool', got %q", tr.Type.Kind())
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
	if tr.Type.Kind() != "bool" {
		t.Errorf("expected Kind 'bool', got %q", tr.Type.Kind())
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
	if tr.Type.Kind() != "int" {
		t.Errorf("expected Kind 'int', got %q", tr.Type.Kind())
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
	if tr.Type.Kind() != "int" {
		t.Errorf("expected Kind 'int', got %q", tr.Type.Kind())
	}
	if tr.Type.(ast.IntegerType).Size != 8 {
		t.Errorf("expected Size 8, got %d", tr.Type.(ast.IntegerType).Size)
	}
}

func TestParseTypeOf(t *testing.T) {
	input := "print(typeof(a));"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	te, ok := stmt.Expr.(*ast.TypeOfExpr)
	if !ok {
		t.Fatalf("expected *ast.TypeOfExpr, got %T", stmt.Expr)
	}
	vr, ok := te.Expr.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected Object *ast.VarRef, got %T", te.Expr)
	}
	if vr.Name != "a" {
		t.Errorf("expected Name 'a', got %q", vr.Name)
	}
}

func TestParseTypeOfStmt(t *testing.T) {
	input := "typeof(x);"
	l := lexer.New(input)
	p := New(l)
	stmt, errs, _ := p.ParseSingleStmt()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	es, ok := stmt.(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected *ast.ExprStmt, got %T", stmt)
	}
	te, ok := es.Expr.(*ast.TypeOfExpr)
	if !ok {
		t.Fatalf("expected *ast.TypeOfExpr, got %T", es.Expr)
	}
	vr, ok := te.Expr.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected *ast.VarRef, got %T", te.Expr)
	}
	if vr.Name != "x" {
		t.Errorf("expected Name 'x', got %q", vr.Name)
	}
}

func TestParseTypeOfStmtMissingSemi(t *testing.T) {
	input := "typeof(x)"
	l := lexer.New(input)
	p := New(l)
	_, errs, _ := p.ParseSingleStmt()
	if len(errs) == 0 {
		t.Errorf("expected error for missing semicolon")
	}
}

func TestParseTypeOfMissingLParen(t *testing.T) {
	input := "typeof x;"
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '(' after typeof")
	}
}

func TestParseTypeOfMissingRParen(t *testing.T) {
	input := "typeof(x;"
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ')'")
	}
}

func TestParseEquality(t *testing.T) {
	input := "print(1 == 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != "==" {
		t.Errorf("expected Op '==', got %q", be.Op)
	}
}

func TestParseNotEqual(t *testing.T) {
	input := "print(1 != 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != "!=" {
		t.Errorf("expected Op '!=', got %q", be.Op)
	}
}

func TestParseComparison(t *testing.T) {
	input := "print(1 < 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != "<" {
		t.Errorf("expected Op '<', got %q", be.Op)
	}
}

func TestParseLTE(t *testing.T) {
	input := "print(1 <= 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != "<=" {
		t.Errorf("expected Op '<=', got %q", be.Op)
	}
}

func TestParseGT(t *testing.T) {
	input := "print(1 > 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != ">" {
		t.Errorf("expected Op '>', got %q", be.Op)
	}
}

func TestParseGTE(t *testing.T) {
	input := "print(1 >= 2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != ">=" {
		t.Errorf("expected Op '>=', got %q", be.Op)
	}
}

func TestParseLogicalAnd(t *testing.T) {
	input := "print(true && false);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != "&&" {
		t.Errorf("expected Op '&&', got %q", be.Op)
	}
}

func TestParseLogicalOr(t *testing.T) {
	input := "print(true || false);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	be, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if be.Op != "||" {
		t.Errorf("expected Op '||', got %q", be.Op)
	}
}

func TestParseNot(t *testing.T) {
	input := "print(!true);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	ue, ok := stmt.Expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected *ast.UnaryExpr, got %T", stmt.Expr)
	}
	if ue.Op != "!" {
		t.Errorf("expected Op '!', got %q", ue.Op)
	}
}

func TestParsePrecedenceOrVsAnd(t *testing.T) {
	input := "print(true || false && false);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	outer, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if outer.Op != "||" {
		t.Errorf("expected outer Op '||', got %q", outer.Op)
	}
	inner, ok := outer.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected inner *ast.BinaryExpr, got %T", outer.Right)
	}
	if inner.Op != "&&" {
		t.Errorf("expected inner Op '&&', got %q", inner.Op)
	}
}

func TestParsePrecedenceComparisonVsAdd(t *testing.T) {
	input := "print(1 + 2 < 3);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	outer, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpr, got %T", stmt.Expr)
	}
	if outer.Op != "<" {
		t.Errorf("expected outer Op '<', got %q", outer.Op)
	}
	left, ok := outer.Left.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected left *ast.BinaryExpr, got %T", outer.Left)
	}
	if left.Op != "+" {
		t.Errorf("expected left Op '+', got %q", left.Op)
	}
}

func TestParsePrefixOnlyLiterals(t *testing.T) {
	tests := []struct {
		input  string
		errMsg string
	}{
		{"print(0x);", "invalid hexadecimal literal"},
		{"print(0X);", "invalid hexadecimal literal"},
		{"print(0b);", "invalid binary literal"},
		{"print(0B);", "invalid binary literal"},
		{"print(0o);", "invalid octal literal"},
		{"print(0O);", "invalid octal literal"},
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

func TestParseDoubleUnaryNot(t *testing.T) {
	input := "print(!!true);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.PrintStmt)
	outer, ok := stmt.Expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected outer *ast.UnaryExpr, got %T", stmt.Expr)
	}
	if outer.Op != "!" {
		t.Errorf("expected outer Op '!', got %q", outer.Op)
	}
	inner, ok := outer.Right.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected inner *ast.UnaryExpr, got %T", outer.Right)
	}
	if inner.Op != "!" {
		t.Errorf("expected inner Op '!', got %q", inner.Op)
	}
}

func TestParseDecrementExpr(t *testing.T) {
	input := "print(--5);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for prefix --, got none")
	}
}

func TestParseEmptyTypeBraces(t *testing.T) {
	tests := []string{
		"var x: int{};",
		"var x: float{};",
		"var x: bool{};",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Errorf("input %q: unexpected errors: %v", input, p.Errors())
		}
		if len(program.Stmts) != 1 {
			t.Errorf("input %q: expected 1 statement, got %d", input, len(program.Stmts))
		}
	}
}

func TestParseEmptyProgram(t *testing.T) {
	input := ""
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 0 {
		t.Errorf("expected 0 statements, got %d", len(program.Stmts))
	}
}

func TestParseNullLiteralWarnings(t *testing.T) {
	tests := []struct {
		input       string
		warningPart string
	}{
		{"print(null || true);", "|| operator"},
		{"print(true || null);", "|| operator"},
		{"print(null && true);", "&& operator"},
		{"print(true && null);", "&& operator"},
		{"print(null < 1);", "comparison operator"},
		{"print(1 < null);", "comparison operator"},
		{"print(!null);", "! operator"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		p.ParseProgram()
		if len(p.Warnings()) == 0 {
			t.Errorf("input %q: expected warning, got none", tt.input)
		}
		found := false
		for _, w := range p.Warnings() {
			if strings.Contains(w, tt.warningPart) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("input %q: expected warning containing %q, got %v", tt.input, tt.warningPart, p.Warnings())
		}
	}
}

func TestParseUnaryMinusOnFloat(t *testing.T) {
	input := "print(-1.5);"
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
	if lit.Value != -1.5 {
		t.Errorf("expected -1.5, got %g", lit.Value)
	}
}

func TestParseUnaryMinusOnNonLiteral(t *testing.T) {
	tests := []string{
		"print(-(1 + 2));",
		"print(-x);",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		stmt := program.Stmts[0].(*ast.PrintStmt)
		bin, ok := stmt.Expr.(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("input %q: expected *ast.BinaryExpr, got %T", input, stmt.Expr)
		}
		if bin.Op != "-" {
			t.Errorf("input %q: expected Op '-', got %q", input, bin.Op)
		}
	}
}

func TestParseIfMissingLParen(t *testing.T) {
	input := "if true) { print(1); }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '('")
	}
}

func TestParseIfMissingRParen(t *testing.T) {
	input := "if (true { print(1); }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ')'")
	}
}

func TestParseIfSingleStmt(t *testing.T) {
	input := "if (true) print(1);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	ifStmt, ok := program.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt, got %T", program.Stmts[0])
	}
	if ifStmt.Then == nil || len(ifStmt.Then.Stmts) != 1 {
		t.Fatalf("expected then block with 1 statement")
	}
}

func TestParseIfMissingCloseBrace(t *testing.T) {
	input := "if (true) { print(1);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '}'")
	}
}

func TestParseIfElseSingleStmt(t *testing.T) {
	input := "if (true) { print(1); } else print(2);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	ifStmt, ok := program.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt, got %T", program.Stmts[0])
	}
	elseBlock, ok := ifStmt.Else.(*ast.BlockStmt)
	if !ok {
		t.Fatalf("expected *ast.BlockStmt for else, got %T", ifStmt.Else)
	}
	if len(elseBlock.Stmts) != 1 {
		t.Fatalf("expected 1 statement in else block, got %d", len(elseBlock.Stmts))
	}
}

func TestParseForStmt(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (i = 0; i < 10; i = i + 1) {
	print(i);
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Stmts))
	}
	forStmt, ok := program.Stmts[1].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected *ast.ForStmt, got %T", program.Stmts[1])
	}
	if forStmt.Condition == nil {
		t.Errorf("expected condition, got nil")
	}
	if forStmt.Body == nil {
		t.Errorf("expected body, got nil")
	}
	if len(forStmt.Body.Stmts) != 1 {
		t.Errorf("expected 1 statement in body, got %d", len(forStmt.Body.Stmts))
	}
}

func TestParseForWithVarDecl(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 10; i = i + 1) {
	print(i);
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	_, ok := program.Stmts[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected *ast.ForStmt, got %T", program.Stmts[0])
	}
}

func TestParseForMissingLParen(t *testing.T) {
	input := "for i = 0; i < 10; i = i + 1) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '('")
	}
}

func TestParseForBadInit(t *testing.T) {
	input := "for (print(1); i < 10; i = i + 1) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for bad init")
	}
}

func TestParseForMissingSemicolonAfterCondition(t *testing.T) {
	input := "for (i = 0; i < 10 i = i + 1) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ';'")
	}
}

func TestParseForBadUpdate(t *testing.T) {
	input := "for (i = 0; i < 10; print(1)) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for bad update")
	}
}

func TestParseForMissingRParen(t *testing.T) {
	input := "for (i = 0; i < 10; i = i + 1 { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ')'")
	}
}

func TestParseWhileStmt(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (i < 10) {
	print(i);
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Stmts))
	}
	whileStmt, ok := program.Stmts[1].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected *ast.WhileStmt, got %T", program.Stmts[1])
	}
	if whileStmt.Condition == nil {
		t.Errorf("expected condition, got nil")
	}
	if whileStmt.Body == nil {
		t.Errorf("expected body, got nil")
	}
	if len(whileStmt.Body.Stmts) != 1 {
		t.Errorf("expected 1 statement in body, got %d", len(whileStmt.Body.Stmts))
	}
}

func TestParseWhileMissingLParen(t *testing.T) {
	input := "while i < 10) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '('")
	}
}

func TestParseWhileMissingRParen(t *testing.T) {
	input := "while (i < 10 { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ')'")
	}
}

func TestParseWhileUnclosedBlock(t *testing.T) {
	input := "while (i < 10) {"
	p := New(lexer.New(input))
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for unclosed block")
	}
}

func TestParseBreakStmt(t *testing.T) {
	input := "break;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseBreakMissingSemicolon(t *testing.T) {
	input := "break"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ';'")
	}
}

func TestParseSkipStmt(t *testing.T) {
	input := "skip;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseSkipMissingSemicolon(t *testing.T) {
	input := "skip"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ';'")
	}
}

func TestParseForUpdatePlusEq(t *testing.T) {
	input := "for (i = 0; i < 10; i += 1) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseForUpdateMinusEq(t *testing.T) {
	input := "for (i = 10; i > 0; i -= 1) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseForUpdateStarEq(t *testing.T) {
	input := "for (i = 0; i < 10; i *= 2) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseForUpdateSlashEq(t *testing.T) {
	input := "for (i = 10; i > 0; i /= 2) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseForUpdateInvalidOp(t *testing.T) {
	input := "for (i = 0; i < 10; i =* 1) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid update operator")
	}
}

func TestParseForSingleStmt(t *testing.T) {
	input := "for (i = 0; i < 10; i = i + 1) print(i);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	forStmt, ok := program.Stmts[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected *ast.ForStmt, got %T", program.Stmts[0])
	}
	if forStmt.Body == nil || len(forStmt.Body.Stmts) != 1 {
		t.Fatalf("expected body with 1 statement")
	}
}

func TestParseWhileSingleStmt(t *testing.T) {
	input := "while (i < 10) print(i);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	whileStmt, ok := program.Stmts[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected *ast.WhileStmt, got %T", program.Stmts[0])
	}
	if whileStmt.Body == nil || len(whileStmt.Body.Stmts) != 1 {
		t.Fatalf("expected body with 1 statement")
	}
}

func TestParseIncDecStmt(t *testing.T) {
	input := "i++;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	inc, ok := program.Stmts[0].(*ast.IncDecStmt)
	if !ok {
		t.Fatalf("expected *ast.IncDecStmt, got %T", program.Stmts[0])
	}
	if inc.Name != "i" {
		t.Errorf("expected name 'i', got %q", inc.Name)
	}
	if inc.Op != "++" {
		t.Errorf("expected op '++', got %q", inc.Op)
	}
}

func TestParseDecStmt(t *testing.T) {
	input := "i--;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	dec, ok := program.Stmts[0].(*ast.IncDecStmt)
	if !ok {
		t.Fatalf("expected *ast.IncDecStmt, got %T", program.Stmts[0])
	}
	if dec.Name != "i" {
		t.Errorf("expected name 'i', got %q", dec.Name)
	}
	if dec.Op != "--" {
		t.Errorf("expected op '--', got %q", dec.Op)
	}
}

func TestParseIncDecMissingSemicolon(t *testing.T) {
	input := "i++"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing semicolon")
	}
}

func TestParseForUpdateIncDec(t *testing.T) {
	input := "for (i = 0; i < 10; i++) { }"
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseForUpdateDec(t *testing.T) {
	input := "for (i = 10; i > 0; i--) { }"
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}

func TestParseForUpdateDefaultError(t *testing.T) {
	input := "for (i = 0; i < 10; i + 1) { }"
	p := New(lexer.New(input))
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid update")
	}
}

func TestParseForUnclosedBlock(t *testing.T) {
	input := "for (i = 0; i < 10; i = i + 1) {"
	p := New(lexer.New(input))
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for unclosed block")
	}
}

func TestParsePrefixIncrement(t *testing.T) {
	input := "++i;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for prefix ++")
	}
}

func TestParsePrefixDecrement(t *testing.T) {
	input := "--i;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for prefix --")
	}
}

func TestParseIncDecInForInit(t *testing.T) {
	input := "for (i++; i < 10; i = i + 1) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for inc/dec in for init")
	}
}

func TestParseForWithRefInit(t *testing.T) {
	input := "for (ref x = y; x; ) { print(x); }"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Stmts))
	}
	forStmt, ok := program.Stmts[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected *ast.ForStmt, got %T", program.Stmts[0])
	}
	refDecl, ok := forStmt.Init.(*ast.RefDecl)
	if !ok {
		t.Fatalf("expected *ast.RefDecl as init, got %T", forStmt.Init)
	}
	if refDecl.Name != "x" {
		t.Errorf("expected ref name 'x', got %q", refDecl.Name)
	}
	varRef, ok := refDecl.Expr.(*ast.VarRef)
	if !ok || varRef.Name != "y" {
		t.Errorf("expected VarRef 'y', got %T %+v", refDecl.Expr, refDecl.Expr)
	}
	if forStmt.Condition == nil {
		t.Fatal("expected condition, got nil")
	}
	condRef, ok := forStmt.Condition.(*ast.VarRef)
	if !ok || condRef.Name != "x" {
		t.Errorf("expected condition VarRef 'x', got %T %+v", forStmt.Condition, forStmt.Condition)
	}
	if forStmt.Update != nil {
		t.Errorf("expected nil update, got %T", forStmt.Update)
	}
	if forStmt.Body == nil || len(forStmt.Body.Stmts) != 1 {
		t.Fatalf("expected body with 1 statement")
	}
}

func TestParseMultipleIncDec(t *testing.T) {
	input := "i++; j--;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(program.Stmts) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Stmts))
	}
	inc, ok := program.Stmts[0].(*ast.IncDecStmt)
	if !ok {
		t.Fatalf("expected *ast.IncDecStmt, got %T", program.Stmts[0])
	}
	if inc.Name != "i" || inc.Op != "++" {
		t.Errorf("expected i++, got %s%s", inc.Name, inc.Op)
	}
	dec, ok := program.Stmts[1].(*ast.IncDecStmt)
	if !ok {
		t.Fatalf("expected *ast.IncDecStmt, got %T", program.Stmts[1])
	}
	if dec.Name != "j" || dec.Op != "--" {
		t.Errorf("expected j--, got %s%s", dec.Name, dec.Op)
	}
}

func TestParseDoubleUnaryMinusWithSpace(t *testing.T) {
	input := "print(- -5);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for double unary minus, got none")
	}
	found := false
	for _, err := range p.Errors() {
		if strings.Contains(err, "unary minus") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error containing 'unary minus', got %v", p.Errors())
	}
}

func TestParseArrayTypeDecl(t *testing.T) {
	input := "var arr: array{size: 5}<int>;"
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
	at, ok := stmt.Type.(ast.ArrayType)
	if !ok {
		t.Fatalf("expected ast.ArrayType, got %T", stmt.Type)
	}
	if at.Size != 5 {
		t.Errorf("expected size 5, got %d", at.Size)
	}
	_, ok = at.ElemType.(ast.IntegerType)
	if !ok {
		t.Fatalf("expected int element type, got %T", at.ElemType)
	}
}

func TestParseArrayTypeFloatElem(t *testing.T) {
	input := "var arr: array{size: 3}<float>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	at := stmt.Type.(ast.ArrayType)
	if at.Size != 3 {
		t.Errorf("expected size 3, got %d", at.Size)
	}
	_, ok := at.ElemType.(ast.FloatType)
	if !ok {
		t.Fatalf("expected float element type, got %T", at.ElemType)
	}
}

func TestParseArrayTypeNoSize(t *testing.T) {
	input := "var arr: array{}<int>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	at := stmt.Type.(ast.ArrayType)
	if at.Size != 0 {
		t.Errorf("expected size 0, got %d", at.Size)
	}
}

func TestParseArrayTypeMissingGT(t *testing.T) {
	input := "var arr: array{size: 5}<int;"
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '>'")
	}
}

func TestParseListTypeDecl(t *testing.T) {
	input := "var lst: list<int>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	lt, ok := stmt.Type.(ast.ListType)
	if !ok {
		t.Fatalf("expected ast.ListType, got %T", stmt.Type)
	}
	if lt.HasMin || lt.HasMax {
		t.Errorf("expected no min/max bounds")
	}
	_, ok = lt.ElemType.(ast.IntegerType)
	if !ok {
		t.Fatalf("expected int element type, got %T", lt.ElemType)
	}
}

func TestParseListTypeWithMinMax(t *testing.T) {
	input := "var lst: list{min: 1, max: 10}<int>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	lt := stmt.Type.(ast.ListType)
	if !lt.HasMin {
		t.Errorf("expected HasMin true")
	}
	if lt.MinSize != 1 {
		t.Errorf("expected MinSize 1, got %d", lt.MinSize)
	}
	if !lt.HasMax {
		t.Errorf("expected HasMax true")
	}
	if lt.MaxSize != 10 {
		t.Errorf("expected MaxSize 10, got %d", lt.MaxSize)
	}
}

func TestParseListTypeAutoBounds(t *testing.T) {
	input := "var lst: list{min: auto, max: 10}<int>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	lt := stmt.Type.(ast.ListType)
	if lt.HasMin {
		t.Errorf("expected HasMin false for auto")
	}
	if !lt.HasMax {
		t.Errorf("expected HasMax true")
	}
	if lt.MaxSize != 10 {
		t.Errorf("expected MaxSize 10, got %d", lt.MaxSize)
	}
}

func TestParseListTypeOnlyMin(t *testing.T) {
	input := "var lst: list{min: 3}<int>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	lt := stmt.Type.(ast.ListType)
	if !lt.HasMin || lt.MinSize != 3 {
		t.Errorf("expected HasMin true, MinSize 3")
	}
	if lt.HasMax {
		t.Errorf("expected HasMax false")
	}
}

func TestParseListTypeMissingGT(t *testing.T) {
	input := "var lst: list<int;"
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '>'")
	}
}

func TestParseIndexedAssign(t *testing.T) {
	input := "arr[0] = 42;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt, ok := program.Stmts[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected *ast.Assignment, got %T", program.Stmts[0])
	}
	if stmt.Name != "arr" {
		t.Errorf("expected name 'arr', got %q", stmt.Name)
	}
	if stmt.Op != "=" {
		t.Errorf("expected op '=', got %q", stmt.Op)
	}
	if stmt.Index == nil {
		t.Fatal("expected non-nil Index")
	}
	intLit, ok := stmt.Index.(*ast.IntegerLit)
	if !ok {
		t.Fatalf("expected *ast.IntegerLit index, got %T", stmt.Index)
	}
	if intLit.Value != 0 {
		t.Errorf("expected index 0, got %d", intLit.Value)
	}
	valLit := stmt.Expr.(*ast.IntegerLit)
	if valLit.Value != 42 {
		t.Errorf("expected value 42, got %d", valLit.Value)
	}
}

func TestParseIndexedAssignCompound(t *testing.T) {
	input := "arr[0] += 5;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.Assignment)
	if stmt.Op != "+=" {
		t.Errorf("expected op '+=', got %q", stmt.Op)
	}
	if stmt.Index == nil {
		t.Fatal("expected non-nil Index")
	}
}

func TestParseIndexedAssignWithIdentIndex(t *testing.T) {
	input := "arr[i] = val;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.Assignment)
	if stmt.Name != "arr" {
		t.Errorf("expected name 'arr', got %q", stmt.Name)
	}
	varRef, ok := stmt.Index.(*ast.VarRef)
	if !ok {
		t.Fatalf("expected *ast.VarRef index, got %T", stmt.Index)
	}
	if varRef.Name != "i" {
		t.Errorf("expected index var 'i', got %q", varRef.Name)
	}
	valRef := stmt.Expr.(*ast.VarRef)
	if valRef.Name != "val" {
		t.Errorf("expected value var 'val', got %q", valRef.Name)
	}
}

func TestParseArrayLit(t *testing.T) {
	input := "print([1, 2, 3]);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	al, ok := ps.Expr.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected *ast.ArrayLit, got %T", ps.Expr)
	}
	if len(al.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(al.Elements))
	}
	e0 := al.Elements[0].(*ast.IntegerLit)
	if e0.Value != 1 {
		t.Errorf("expected 1, got %d", e0.Value)
	}
	e1 := al.Elements[1].(*ast.IntegerLit)
	if e1.Value != 2 {
		t.Errorf("expected 2, got %d", e1.Value)
	}
	e2 := al.Elements[2].(*ast.IntegerLit)
	if e2.Value != 3 {
		t.Errorf("expected 3, got %d", e2.Value)
	}
}

func TestParseEmptyArrayLit(t *testing.T) {
	input := "print([]);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	al, ok := ps.Expr.(*ast.ArrayLit)
	if !ok {
		t.Fatalf("expected *ast.ArrayLit, got %T", ps.Expr)
	}
	if len(al.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(al.Elements))
	}
}

func TestParseIndexExprInPrint(t *testing.T) {
	input := "print(arr[0]);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	ie, ok := ps.Expr.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected *ast.IndexExpr, got %T", ps.Expr)
	}
	varRef, ok := ie.Object.(*ast.VarRef)
	if !ok || varRef.Name != "arr" {
		t.Errorf("expected VarRef 'arr', got %T %+v", ie.Object, ie.Object)
	}
	intLit := ie.Index.(*ast.IntegerLit)
	if intLit.Value != 0 {
		t.Errorf("expected index 0, got %d", intLit.Value)
	}
}

func TestParseIndexExprWithVarIndex(t *testing.T) {
	input := "print(arr[i]);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	ie := ps.Expr.(*ast.IndexExpr)
	varRef := ie.Index.(*ast.VarRef)
	if varRef.Name != "i" {
		t.Errorf("expected index var 'i', got %q", varRef.Name)
	}
}

func TestParseMemberAccessLength(t *testing.T) {
	input := "print(arr.length);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	ma, ok := ps.Expr.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected *ast.MemberAccess, got %T", ps.Expr)
	}
	varRef := ma.Object.(*ast.VarRef)
	if varRef.Name != "arr" {
		t.Errorf("expected VarRef 'arr', got %q", varRef.Name)
	}
	if ma.Member != "length" {
		t.Errorf("expected member 'length', got %q", ma.Member)
	}
	if len(ma.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(ma.Args))
	}
}

func TestParseMemberCallAdd(t *testing.T) {
	input := "a.add(4);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	es := program.Stmts[0].(*ast.ExprStmt)
	ma, ok := es.Expr.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("expected *ast.MemberAccess, got %T", es.Expr)
	}
	varRef := ma.Object.(*ast.VarRef)
	if varRef.Name != "a" {
		t.Errorf("expected VarRef 'a', got %q", varRef.Name)
	}
	if ma.Member != "add" {
		t.Errorf("expected member 'add', got %q", ma.Member)
	}
	if len(ma.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(ma.Args))
	}
	intLit := ma.Args[0].(*ast.IntegerLit)
	if intLit.Value != 4 {
		t.Errorf("expected arg 4, got %d", intLit.Value)
	}
}

func TestParseMemberCallRemove(t *testing.T) {
	input := "a.remove(0);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	es := program.Stmts[0].(*ast.ExprStmt)
	ma := es.Expr.(*ast.MemberAccess)
	if ma.Member != "remove" {
		t.Errorf("expected member 'remove', got %q", ma.Member)
	}
	if len(ma.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(ma.Args))
	}
}

func TestParseTypeRefArray(t *testing.T) {
	input := "print(array{size: 5}<int>);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	tr, ok := ps.Expr.(*ast.TypeRef)
	if !ok {
		t.Fatalf("expected *ast.TypeRef, got %T", ps.Expr)
	}
	if !tr.IsType {
		t.Errorf("expected IsType true")
	}
	at, ok := tr.Type.(ast.ArrayType)
	if !ok {
		t.Fatalf("expected ast.ArrayType, got %T", tr.Type)
	}
	if at.Size != 5 {
		t.Errorf("expected size 5, got %d", at.Size)
	}
}

func TestParseTypeRefList(t *testing.T) {
	input := "print(list{min: 1, max: 10}<int>);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	tr := ps.Expr.(*ast.TypeRef)
	lt := tr.Type.(ast.ListType)
	if !lt.HasMin || lt.MinSize != 1 {
		t.Errorf("expected HasMin, MinSize 1")
	}
	if !lt.HasMax || lt.MaxSize != 10 {
		t.Errorf("expected HasMax, MaxSize 10")
	}
}

func TestParseTypeRefListNoBounds(t *testing.T) {
	input := "print(list<int>);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	ps := program.Stmts[0].(*ast.PrintStmt)
	tr := ps.Expr.(*ast.TypeRef)
	lt := tr.Type.(ast.ListType)
	if lt.HasMin || lt.HasMax {
		t.Errorf("expected no bounds")
	}
}

// ---- Coverage edge cases for uncovered branches ----

func TestParseStmtMemberAccessNoSemicolon(t *testing.T) {
	input := "a.add(4)"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing semicolon after member access call")
	}
}

func TestParseTypeParamsInvalidKey(t *testing.T) {
	input := "var x: int{invalid_key: 42};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid type param key")
	}
}

func TestParseTypeParamsUnexpectedParam(t *testing.T) {
	input := "var x: int{size: bogus};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid integer param value")
	}
}

func TestParseTypeParamsOptionalParam(t *testing.T) {
	input := "var x: int{signed: wrong};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid signed value")
	}
}

func TestParseTypeParamsInvalidKeyForContext(t *testing.T) {
	input := "var x: int{min: 1};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid key in context")
	}
}

func TestParseTypeParamsInvalidIntegerValue(t *testing.T) {
	input := "var x: int{size: 99999999999999999999};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid integer")
	}
}

func TestParseTypeParamsMissingColon(t *testing.T) {
	input := "var x: int{size 42};"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing colon")
	}
}

func TestForceTypeNotNullableFloat(t *testing.T) {
	result := forceTypeNotNullable(ast.FloatType{Size: 32, Nullable: true})
	ft, ok := result.(ast.FloatType)
	if !ok {
		t.Fatalf("expected FloatType")
	}
	if ft.Nullable {
		t.Errorf("expected nullable false")
	}
	if ft.Size != 32 {
		t.Errorf("expected size 32, got %d", ft.Size)
	}
}

func TestForceTypeNotNullableArrayType(t *testing.T) {
	result := forceTypeNotNullable(ast.ArrayType{ElemType: ast.IntegerType{Size: 64, Signed: true}, Size: 3})
	_, ok := result.(ast.ArrayType)
	if !ok {
		t.Fatalf("expected ArrayType passthrough")
	}
}

func TestForceTypeNotNullableBool(t *testing.T) {
	result := forceTypeNotNullable(ast.BoolType{Nullable: true})
	bt, ok := result.(ast.BoolType)
	if !ok {
		t.Fatalf("expected BoolType")
	}
	if bt.Nullable {
		t.Errorf("expected nullable false")
	}
}

func TestParseArrayTypeMissingLT(t *testing.T) {
	input := "var a: array{size: 5}int;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '<' in array type")
	}
}

func TestParseListTypeMissingLT(t *testing.T) {
	input := "var a: list int;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '<' in list type")
	}
}

func TestParseArrayTypeInvalidAfterElem(t *testing.T) {
	input := "var a: array<int!;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '>' in array type")
	}
}

func TestParseListTypeInvalidAfterElem(t *testing.T) {
	input := "var a: list<int!;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing '>' in list type")
	}
}

func TestParseIndexedAssignMissingRBracket(t *testing.T) {
	input := "arr[0 = 42;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ']'")
	}
}

func TestParseIndexedAssignNotAssignOp(t *testing.T) {
	input := "arr[0] == 42;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for non-assignment operator")
	}
}

func TestParseIndexedAssignMissingSemicolon(t *testing.T) {
	input := "arr[0] = 42"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing semicolon")
	}
}

func TestParsePostfixNonMemberToken(t *testing.T) {
	input := "a.42;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for non-member token after dot")
	}
}

func TestParseIndexExprMissingRBracket(t *testing.T) {
	input := "print(arr[0);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ']' in index expression")
	}
}

func TestParseArrayLitMissingRBracket(t *testing.T) {
	input := "print([1, 2);"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ']' in array literal")
	}
}

func TestParsePostfixMultipleArgs(t *testing.T) {
	input := "a.add(1, 2, 3);"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	es := program.Stmts[0].(*ast.ExprStmt)
	ma := es.Expr.(*ast.MemberAccess)
	if len(ma.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(ma.Args))
	}
}

func TestParseMemberCallMissingRParen(t *testing.T) {
	input := "a.add(4;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for missing ')' in member call")
	}
}

func TestParseArrayBracketSyntax(t *testing.T) {
	input := "var a: array[5]<int>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	at := stmt.Type.(ast.ArrayType)
	if at.Size != 5 {
		t.Errorf("expected size 5, got %d", at.Size)
	}
}

func TestParseArrayBracketSyntaxFloatElem(t *testing.T) {
	input := "var a: array[3]<float>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	at := stmt.Type.(ast.ArrayType)
	if at.Size != 3 {
		t.Errorf("expected size 3, got %d", at.Size)
	}
	_, ok := at.ElemType.(ast.FloatType)
	if !ok {
		t.Errorf("expected float element type")
	}
}

func TestParseArrayNoSize(t *testing.T) {
	input := "var a: array<int>;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	stmt := program.Stmts[0].(*ast.VarDecl)
	at := stmt.Type.(ast.ArrayType)
	if at.Size != 0 {
		t.Errorf("expected size 0, got %d", at.Size)
	}
}

func TestParseArrayNoSpaceBeforeAssign(t *testing.T) {
	input := "var a: array<int>=[1,2,3];"
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
	at, ok := stmt.Type.(ast.ArrayType)
	if !ok {
		t.Fatalf("expected ast.ArrayType, got %T", stmt.Type)
	}
	if at.Size != 0 {
		t.Errorf("expected size 0, got %d", at.Size)
	}
	_, ok = at.ElemType.(ast.IntegerType)
	if !ok {
		t.Fatalf("expected int element type, got %T", at.ElemType)
	}
}

func TestParseListNoSpaceBeforeAssign(t *testing.T) {
	input := "var a: list<int>=[1,2,3];"
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
	lt, ok := stmt.Type.(ast.ListType)
	if !ok {
		t.Fatalf("expected ast.ListType, got %T", stmt.Type)
	}
	_, ok = lt.ElemType.(ast.IntegerType)
	if !ok {
		t.Fatalf("expected int element type, got %T", lt.ElemType)
	}
}

func TestParseArrayBracketEmpty(t *testing.T) {
	input := "var a: array[]<int>;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for empty brackets")
	}
}

func TestParseArrayBracketNonInt(t *testing.T) {
	input := "var a: array[abc]<int>;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for non-integer in brackets")
	}
}

func TestParseArrayBracketInvalidInt(t *testing.T) {
	input := "var a: array[999999999999999999999]<int>;"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected error for invalid integer in brackets")
	}
}

func TestParseStringTypeFixed(t *testing.T) {
	input := `var s: string{size: 5} = "hello";`
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

	if !stmt.IsString {
		t.Fatalf("expected IsString to be true")
	}

	st, ok := stmt.Type.(ast.StringType)
	if !ok {
		t.Fatalf("expected StringType, got %T", stmt.Type)
	}

	if st.Size != 5 {
		t.Errorf("expected size 5, got %d", st.Size)
	}

	if stmt.Expr == nil {
		t.Fatalf("expected expression, got nil")
	}

	strLit, ok := stmt.Expr.(*ast.StringLit)
	if !ok {
		t.Fatalf("expected *ast.StringLit, got %T", stmt.Expr)
	}

	if strLit.Value != "hello" {
		t.Errorf("expected value 'hello', got %q", strLit.Value)
	}
}

func TestParseStringTypeDynamic(t *testing.T) {
	input := `var s: string{min: 1, max: 10} = "hello";`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	st, ok := stmt.Type.(ast.StringType)
	if !ok {
		t.Fatalf("expected StringType, got %T", stmt.Type)
	}

	if !st.HasMin || st.MinSize != 1 {
		t.Errorf("expected min: 1, got min: %v (hasMin: %v)", st.MinSize, st.HasMin)
	}
	if !st.HasMax || st.MaxSize != 10 {
		t.Errorf("expected max: 10, got max: %v (hasMax: %v)", st.MaxSize, st.HasMax)
	}
}

func TestParseStringTypeBasic(t *testing.T) {
	input := `var s: string = "hello";`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	if !stmt.IsString {
		t.Fatalf("expected IsString to be true")
	}

	_, ok := stmt.Type.(ast.StringType)
	if !ok {
		t.Fatalf("expected StringType, got %T", stmt.Type)
	}
}

func TestParseStringLit(t *testing.T) {
	input := `print("hello world");`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.PrintStmt)
	strLit, ok := stmt.Expr.(*ast.StringLit)
	if !ok {
		t.Fatalf("expected *ast.StringLit, got %T", stmt.Expr)
	}

	if strLit.Value != "hello world" {
		t.Errorf("expected 'hello world', got %q", strLit.Value)
	}

	if !strLit.Untyped {
		t.Errorf("expected untyped true")
	}
}

func TestParseStringTypeRef(t *testing.T) {
	input := `print(string);`
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

	if tr.Type.Kind() != "string" {
		t.Errorf("expected Kind 'string', got %q", tr.Type.Kind())
	}
}

func TestParseStringDeclWithoutInit(t *testing.T) {
	input := `var s: string;`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	if !stmt.IsString {
		t.Fatalf("expected IsString to be true")
	}

	if stmt.Expr != nil {
		t.Errorf("expected nil expression")
	}
}

func TestParseStringDeclFixedWithoutInit(t *testing.T) {
	input := `var s: string{size: 5};`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	stmt := program.Stmts[0].(*ast.VarDecl)
	st := stmt.Type.(ast.StringType)
	if st.Size != 5 {
		t.Errorf("expected size 5, got %d", st.Size)
	}
}

func TestParseForUpdateModEq(t *testing.T) {
	input := "for (i = 0; i < 10; i %= 3) { }"
	l := lexer.New(input)
	p := New(l)
	p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
}
