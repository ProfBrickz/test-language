package interpreter

import (
	"io"
	"math"
	"os"
	"strings"
	"testing"

	"lang-interpreter/ast"
	"lang-interpreter/lexer"
	"lang-interpreter/parser"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return strings.TrimSpace(string(out))
}

func TestVarDeclAndPrint(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 42;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "42") {
		t.Errorf("expected output to contain '42', got %q", output)
	}
}

func TestAssignment(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x = 20;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "20") {
		t.Errorf("expected output to contain '20', got %q", output)
	}
}

func TestCompoundAssignment(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x += 5;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "15") {
		t.Errorf("expected output to contain '15', got %q", output)
	}
}

func TestBinaryExpr(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
var y: integer{size: 32, signed: true, nullable: false} = 20;
print(x + y);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "30") {
		t.Errorf("expected output to contain '30', got %q", output)
	}
}

func TestNullAssign(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: true} = null;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "null") {
		t.Errorf("expected output to contain 'null', got %q", output)
	}
}

func TestTypeMismatch(t *testing.T) {
	input := `
var x: integer{size: 8, signed: true, nullable: false} = 1000;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestDivisionByZero(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x /= 0;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected division by zero error")
	}
}

func TestBinaryExprEval(t *testing.T) {
	i := New()

	// Create a binary expression: 10 + 20
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, Untyped: true},
	}

	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.Data != 30 {
		t.Errorf("expected 30, got %d", val.Data)
	}
}

func TestCanImplicitConvert(t *testing.T) {
	tests := []struct {
		src      ast.IntegerType
		dst      ast.IntegerType
		expected bool
	}{
		{ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 32, Signed: true}, true},
		{ast.IntegerType{Size: 8, Signed: true}, ast.IntegerType{Size: 32, Signed: true}, true},
		{ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 8, Signed: true}, false},
		{ast.IntegerType{Size: 32, Signed: false}, ast.IntegerType{Size: 32, Signed: false}, true},
		{ast.IntegerType{Size: 8, Signed: false}, ast.IntegerType{Size: 32, Signed: true}, true},
	}

	for _, tt := range tests {
		result := canImplicitConvert(tt.src, tt.dst)
		if result != tt.expected {
			t.Errorf("canImplicitConvert(%v, %v) = %v, expected %v", tt.src, tt.dst, result, tt.expected)
		}
	}
}

func TestCanFitInType(t *testing.T) {
	tests := []struct {
		val      int64
		itype    ast.IntegerType
		expected bool
	}{
		{100, ast.IntegerType{Size: 8, Signed: true}, true},
		{200, ast.IntegerType{Size: 8, Signed: true}, false},
		{255, ast.IntegerType{Size: 8, Signed: false}, true},
		{256, ast.IntegerType{Size: 8, Signed: false}, false},
		{32767, ast.IntegerType{Size: 16, Signed: true}, true},
		{32768, ast.IntegerType{Size: 16, Signed: true}, false},
		{65535, ast.IntegerType{Size: 16, Signed: false}, true},
		{65536, ast.IntegerType{Size: 16, Signed: false}, false},
		{2147483647, ast.IntegerType{Size: 32, Signed: true}, true},
		{-2147483648, ast.IntegerType{Size: 32, Signed: true}, true},
		{4294967295, ast.IntegerType{Size: 32, Signed: false}, true},
		{9223372036854775807, ast.IntegerType{Size: 64, Signed: true}, true},
	}

	for _, tt := range tests {
		result := canFitInType(tt.val, tt.itype)
		if result != tt.expected {
			t.Errorf("canFitInType(%d, %v) = %v, expected %v", tt.val, tt.itype, result, tt.expected)
		}
	}
}

func TestValueString(t *testing.T) {
	tests := []struct {
		val      Value
		expected string
	}{
		{Value{Null: true}, "null"},
		{Value{Untyped: true, Data: 42}, "42"},
		{Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 10}, "integer{size: 32, signed: true}(10)"},
		{Value{IType: ast.IntegerType{Size: 16, Signed: false}, Data: 5}, "integer{size: 16, signed: false}(5)"},
	}

	for _, tt := range tests {
		result := tt.val.String()
		if result != tt.expected {
			t.Errorf("Value.String() = %q, expected %q", result, tt.expected)
		}
	}
}

func TestTypeDesc(t *testing.T) {
	tests := []struct {
		itype    ast.IntegerType
		expected string
	}{
		{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, "32 bit signed integer"},
		{ast.IntegerType{Size: 16, Signed: false, Nullable: true}, "16 bit nullable unsigned integer"},
		{ast.IntegerType{Size: 64, Signed: true, Nullable: true}, "64 bit nullable signed integer"},
	}

	for _, tt := range tests {
		result := typeDesc(tt.itype)
		if result != tt.expected {
			t.Errorf("typeDesc(%v) = %q, expected %q", tt.itype, result, tt.expected)
		}
	}
}

func TestExecuteStmt(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.ExecuteStmt(program.Stmts[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnknownStmtType(t *testing.T) {
	i := New()
	err := i.executeStmt(nil)
	if err == nil {
		t.Errorf("expected error for nil statement")
	}
}

func TestEvalExprNull(t *testing.T) {
	i := New()
	expr := &ast.NullLit{}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Null {
		t.Errorf("expected null value")
	}
}

func TestEvalExprUnknown(t *testing.T) {
	i := New()
	// Create an expr that doesn't match any case
	expr := &ast.PrintStmt{}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for unknown expression type")
	}
}

func TestEvalBinaryDivByZero(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "/",
		Right: &ast.IntegerLit{Value: 0, Untyped: true},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected division by zero error")
	}
}

func TestEvalBinaryUnknownOp(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "%",
		Right: &ast.IntegerLit{Value: 3, Untyped: true},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected error for unknown operator")
	}
}

func TestEvalBinaryWithNull(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "+",
		Right: &ast.NullLit{},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Null {
		t.Errorf("expected null result")
	}
}

func TestEvalBinaryTypeMismatch(t *testing.T) {
	i := New()
	// Both conversions fail: 32-bit unsigned and 8-bit signed
	// 32-bit unsigned -> 8-bit signed fails (size reduction)
	// 8-bit signed -> 32-bit unsigned fails (signed to unsigned)
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, IType: ast.IntegerType{Size: 32, Signed: false}, Untyped: false},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, IType: ast.IntegerType{Size: 8, Signed: true}, Untyped: false},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestExecuteAssignmentNullToNonNullable(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x = null;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error assigning null to non-nullable")
	}
}

func TestExecuteAssignmentFromNullVar(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: true} = null;
x += 5;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error using null variable in += operation")
	}
}

func TestExecuteAssignmentUnknownOp(t *testing.T) {
	i := New()
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%",
		Expr: &ast.IntegerLit{Value: 5, Untyped: true},
	}
	i.env.Set("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 10})
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for unknown operator")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		val      int64
		itype    ast.IntegerType
		expected int64
	}{
		{200, ast.IntegerType{Size: 8, Signed: true}, 127},
		{-200, ast.IntegerType{Size: 8, Signed: true}, -128},
		{300, ast.IntegerType{Size: 8, Signed: false}, 255},
		{-10, ast.IntegerType{Size: 8, Signed: false}, 0},
		{100000, ast.IntegerType{Size: 16, Signed: true}, 32767},
	}

	for _, tt := range tests {
		result := clamp(tt.val, tt.itype)
		if result != tt.expected {
			t.Errorf("clamp(%d, %v) = %d, expected %d", tt.val, tt.itype, result, tt.expected)
		}
	}
}

func TestCanImplicitConvertMore(t *testing.T) {
	tests := []struct {
		src      ast.IntegerType
		dst      ast.IntegerType
		expected bool
	}{
		{ast.IntegerType{Size: 32, Signed: false}, ast.IntegerType{Size: 32, Signed: true}, false},
		{ast.IntegerType{Size: 32, Signed: true}, ast.IntegerType{Size: 32, Signed: false}, false},
		{ast.IntegerType{Size: 64, Signed: true}, ast.IntegerType{Size: 32, Signed: true}, false},
		{ast.IntegerType{Size: 8, Signed: false}, ast.IntegerType{Size: 16, Signed: false}, true},
		{ast.IntegerType{Size: 16, Signed: true}, ast.IntegerType{Size: 32, Signed: true}, true},
	}

	for _, tt := range tests {
		result := canImplicitConvert(tt.src, tt.dst)
		if result != tt.expected {
			t.Errorf("canImplicitConvert(%v, %v) = %v, expected %v", tt.src, tt.dst, result, tt.expected)
		}
	}
}

func TestAssignmentWithVarRef(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
var y: integer{size: 32, signed: true, nullable: false} = x;
print(y);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "10") {
		t.Errorf("expected output to contain '10', got %q", output)
	}
}

func TestTypeConversionOnAssign(t *testing.T) {
	input := `
var x: integer{size: 8, signed: true, nullable: false} = 10;
var y: integer{size: 32, signed: true, nullable: false} = x;
print(y);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "10") {
		t.Errorf("expected output to contain '10', got %q", output)
	}
}

func TestExecuteAssignmentWithLiteralOverflow(t *testing.T) {
	input := `
var x: integer{size: 8, signed: true, nullable: false} = 10;
x = 200;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestExecuteAssignmentWithVarRefAndConversion(t *testing.T) {
	input := `
var x: integer{size: 16, signed: true, nullable: false} = 1000;
var y: integer{size: 32, signed: true, nullable: false} = 0;
y = x;
print(y);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "1000") {
		t.Errorf("expected output to contain '1000', got %q", output)
	}
}

func TestEvalBinaryWithTypedOperands(t *testing.T) {
	i := New()
	// Both operands have types - test conversion logic
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 30 {
		t.Errorf("expected 30, got %d", val.Data)
	}
}

func TestEvalBinaryWithSizeClamp(t *testing.T) {
	i := New()
	// Right type converts to left type with clamping
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 200, IType: ast.IntegerType{Size: 8, Signed: false}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 210 {
		t.Errorf("expected 210, got %d", val.Data)
	}
}

func TestEvalBinaryResultOverflow(t *testing.T) {
	i := New()
	// Result overflows the type
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 200, IType: ast.IntegerType{Size: 8, Signed: true}, Untyped: false},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 200, IType: ast.IntegerType{Size: 8, Signed: true}, Untyped: false},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestCanFitInTypeEdgeCases(t *testing.T) {
	tests := []struct {
		val      int64
		itype    ast.IntegerType
		expected bool
	}{
		{math.MaxInt8, ast.IntegerType{Size: 8, Signed: true}, true},
		{math.MinInt8, ast.IntegerType{Size: 8, Signed: true}, true},
		{math.MaxUint8, ast.IntegerType{Size: 8, Signed: false}, true},
		{math.MaxInt16, ast.IntegerType{Size: 16, Signed: true}, true},
		{math.MinInt16, ast.IntegerType{Size: 16, Signed: true}, true},
		{math.MaxUint16, ast.IntegerType{Size: 16, Signed: false}, true},
		{math.MaxInt32, ast.IntegerType{Size: 32, Signed: true}, true},
		{math.MinInt32, ast.IntegerType{Size: 32, Signed: true}, true},
		{math.MaxUint32, ast.IntegerType{Size: 32, Signed: false}, true},
	}

	for _, tt := range tests {
		result := canFitInType(tt.val, tt.itype)
		if result != tt.expected {
			t.Errorf("canFitInType(%d, %v) = %v, expected %v", tt.val, tt.itype, result, tt.expected)
		}
	}
}

func TestClampEdgeCases(t *testing.T) {
	tests := []struct {
		val      int64
		itype    ast.IntegerType
		expected int64
	}{
		{math.MaxInt8 + 1, ast.IntegerType{Size: 8, Signed: true}, math.MaxInt8},
		{math.MinInt8 - 1, ast.IntegerType{Size: 8, Signed: true}, math.MinInt8},
		{math.MaxUint8 + 1, ast.IntegerType{Size: 8, Signed: false}, math.MaxUint8},
		{-1, ast.IntegerType{Size: 8, Signed: false}, 0},
		{math.MaxInt16 + 1, ast.IntegerType{Size: 16, Signed: true}, math.MaxInt16},
		{math.MinInt16 - 1, ast.IntegerType{Size: 16, Signed: true}, math.MinInt16},
		{math.MaxUint16 + 1, ast.IntegerType{Size: 16, Signed: false}, math.MaxUint16},
		{math.MaxInt32 + 1, ast.IntegerType{Size: 32, Signed: true}, math.MaxInt32},
		{math.MinInt32 - 1, ast.IntegerType{Size: 32, Signed: true}, math.MinInt32},
		{math.MaxUint32 + 1, ast.IntegerType{Size: 32, Signed: false}, math.MaxUint32},
	}

	for _, tt := range tests {
		result := clamp(tt.val, tt.itype)
		if result != tt.expected {
			t.Errorf("clamp(%d, %v) = %d, expected %d", tt.val, tt.itype, result, tt.expected)
		}
	}
}

func TestExecuteAssignmentNullAssign(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: true} = 10;
x = null;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "null") {
		t.Errorf("expected output to contain 'null', got %q", output)
	}
}

func TestEvalExprVarRef(t *testing.T) {
	i := New()
	i.env.Set("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 42})

	expr := &ast.VarRef{Name: "x"}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 42 {
		t.Errorf("expected 42, got %d", val.Data)
	}
}

func TestEvalExprUndefinedVar(t *testing.T) {
	i := New()
	expr := &ast.VarRef{Name: "undefined"}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for undefined variable")
	}
}

func TestEvalBinaryMixedTypes(t *testing.T) {
	i := New()
	// Untyped left, typed right
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 30 {
		t.Errorf("expected 30, got %d", val.Data)
	}
}

func TestEvalBinaryResultUntyped(t *testing.T) {
	i := New()
	// Both untyped - result should be untyped
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 100, Untyped: true},
		Op:    "*",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Untyped {
		t.Errorf("expected untyped result")
	}
	if val.Data != 200 {
		t.Errorf("expected 200, got %d", val.Data)
	}
}

func TestExecuteAssignmentWithLiteral(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x = 20;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "20") {
		t.Errorf("expected output to contain '20', got %q", output)
	}
}

func TestCanFitInType64BitUnsigned(t *testing.T) {
	// 64-bit unsigned uses math.MaxInt64 as max
	tests := []struct {
		val      int64
		itype    ast.IntegerType
		expected bool
	}{
		{math.MaxInt64, ast.IntegerType{Size: 64, Signed: false}, true},
		{math.MaxInt64, ast.IntegerType{Size: 64, Signed: false}, true},
		{-1, ast.IntegerType{Size: 64, Signed: false}, false},
	}

	for _, tt := range tests {
		result := canFitInType(tt.val, tt.itype)
		if result != tt.expected {
			t.Errorf("canFitInType(%d, %v) = %v, expected %v", tt.val, tt.itype, result, tt.expected)
		}
	}
}

func TestClampDefaultCase(t *testing.T) {
	// Test the default case in clamp (invalid size)
	// Since Go won't let us create invalid sizes easily,
	// let's just verify the function works for valid sizes
	result := clamp(0, ast.IntegerType{Size: 64, Signed: true})
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}

	// Test with invalid size (triggers default case)
	// Use reflection to set invalid size
	itype := ast.IntegerType{Size: 123, Signed: true} // invalid size
	result = clamp(100, itype)
	// Default case uses MinInt64/MaxInt64
	if result != 100 {
		t.Errorf("expected 100, got %d", result)
	}
}

func TestCanFitInTypeDefaultCase(t *testing.T) {
	// Test the default case in canFitInType (invalid size)
	// Use an invalid size to trigger default case
	itype := ast.IntegerType{Size: 123, Signed: true} // invalid size
	result := canFitInType(100, itype)
	// Default case uses MinInt64/MaxInt64
	if !result {
		t.Errorf("expected true for value within int64 range")
	}

	result = canFitInType(-100, itype)
	if !result {
		t.Errorf("expected true for negative value within int64 range")
	}
}

func TestExecuteAssignmentWithUntypedExpr(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x = x + 5;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "15") {
		t.Errorf("expected output to contain '15', got %q", output)
	}
}

func TestEvalBinaryWithLeftUntyped(t *testing.T) {
	i := New()
	// Left untyped, right typed - result type should be right's type
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 30 {
		t.Errorf("expected 30, got %d", val.Data)
	}
}

func TestEvalBinaryWithBothTypedLeftConvert(t *testing.T) {
	i := New()
	// Both typed, left type wins because size >= right size
	// Right converts to left type
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, IType: ast.IntegerType{Size: 16, Signed: true}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 30 {
		t.Errorf("expected 30, got %d", val.Data)
	}
}

func TestClamp64Bit(t *testing.T) {
	tests := []struct {
		val      int64
		itype    ast.IntegerType
		expected int64
	}{
		{math.MaxInt64, ast.IntegerType{Size: 64, Signed: true}, math.MaxInt64},
		{math.MinInt64, ast.IntegerType{Size: 64, Signed: true}, math.MinInt64},
		{math.MaxInt64, ast.IntegerType{Size: 64, Signed: false}, math.MaxInt64},
		{-1, ast.IntegerType{Size: 64, Signed: false}, 0},
	}

	for _, tt := range tests {
		result := clamp(tt.val, tt.itype)
		if result != tt.expected {
			t.Errorf("clamp(%d, %v) = %d, expected %d", tt.val, tt.itype, result, tt.expected)
		}
	}
}

func TestEvalBinaryRightConvertsToLeft(t *testing.T) {
	i := New()
	// Right can convert to left, left.IType.Size >= right.IType.Size
	// Right gets clamped and converted to left's type
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, IType: ast.IntegerType{Size: 16, Signed: true}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 30 {
		t.Errorf("expected 30, got %d", val.Data)
	}
}

func TestEvalBinaryLeftConvertsToRight(t *testing.T) {
	i := New()
	// Left can convert to right (but left.IType.Size < right.IType.Size)
	// Left gets clamped and converted to right's type
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, IType: ast.IntegerType{Size: 16, Signed: true}, Untyped: false},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 30 {
		t.Errorf("expected 30, got %d", val.Data)
	}
}

func TestEvalBinaryWithMultiplicationOverflow(t *testing.T) {
	i := New()
	// Result overflows the type
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 200, IType: ast.IntegerType{Size: 8, Signed: true}, Untyped: false},
		Op:    "*",
		Right: &ast.IntegerLit{Value: 2, IType: ast.IntegerType{Size: 8, Signed: true}, Untyped: false},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestExecuteAssignmentWithTypeMismatch(t *testing.T) {
	input := `
var x: integer{size: 8, signed: true, nullable: false} = 10;
var y: integer{size: 32, signed: true, nullable: false} = 1000;
x = y;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestExecuteAssignmentUndefinedVar(t *testing.T) {
	i := New()
	stmt := &ast.Assignment{
		Name: "undefined",
		Op:   "=",
		Expr: &ast.IntegerLit{Value: 5, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for undefined variable")
	}
}

func TestExecuteVarDeclWithNullToNonNullable(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false},
		Expr:  &ast.NullLit{},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error for null to non-nullable")
	}
}

func TestExecuteVarDeclWithTypeMismatch(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 8, Signed: true, Nullable: false},
		Expr:  &ast.IntegerLit{Value: 200, IType: ast.IntegerType{Size: 16, Signed: true}, Untyped: false},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestExecuteVarDeclWithVarRefAndTypeMismatch(t *testing.T) {
	i := New()
	i.env.Set("y", Value{IType: ast.IntegerType{Size: 16, Signed: true}, Data: 200})
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 8, Signed: true, Nullable: false},
		Expr:  &ast.VarRef{Name: "y"},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}


func TestExecuteVarDeclWithLiteralOverflow(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
	IType: ast.IntegerType{Size: 8, Signed: true, Nullable: false},
		Expr:  &ast.IntegerLit{Value: 200, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestExecutePrintStmtWithError(t *testing.T) {
	i := New()
	stmt := &ast.PrintStmt{
		Expr: &ast.VarRef{Name: "undefined"},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error for undefined variable in print")
	}
}

func TestExecuteAssignmentSubtraction(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x -= 5;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "5") {
		t.Errorf("expected output to contain '5', got %q", output)
	}
}

func TestExecuteAssignmentMultiplication(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x *= 5;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "50") {
		t.Errorf("expected output to contain '50', got %q", output)
	}
}

func TestExecuteAssignmentDivision(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: false} = 10;
x /= 2;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "5") {
		t.Errorf("expected output to contain '5', got %q", output)
	}
}

func TestEvalBinarySubtraction(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "-",
		Right: &ast.IntegerLit{Value: 3, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 7 {
		t.Errorf("expected 7, got %d", val.Data)
	}
}

func TestEvalBinaryDivision(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "/",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 5 {
		t.Errorf("expected 5, got %d", val.Data)
	}
}

func TestEvalBinaryLeftError(t *testing.T) {
	i := New()
	// Left expression returns error
	expr := &ast.BinaryExpr{
		Left:  &ast.PrintStmt{}, // invalid expression
		Op:    "+",
		Right: &ast.IntegerLit{Value: 20, Untyped: true},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected error from left expression")
	}
}

func TestEvalBinaryRightError(t *testing.T) {
	i := New()
	// Right expression returns error
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "+",
		Right: &ast.PrintStmt{}, // invalid expression
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected error from right expression")
	}
}

func TestExecuteAssignmentWithNullLiteral(t *testing.T) {
	input := `
var x: integer{size: 32, signed: true, nullable: true} = 10;
x = null;
print(x);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "null") {
		t.Errorf("expected output to contain 'null', got %q", output)
	}
}

func TestExecuteVarDeclWithExprError(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false},
		Expr:  &ast.PrintStmt{}, // invalid expression that will cause error
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error from invalid expression")
	}
}

func TestExecuteAssignmentWithExprError(t *testing.T) {
	i := New()
	i.env.Set("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 10})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.PrintStmt{}, // invalid expression
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error from invalid expression")
	}
}
