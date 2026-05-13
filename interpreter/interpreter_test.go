package interpreter

import (
	"fmt"
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
var x: int{size: 32, signed: true, nullable: false} = 42;
print((x).toString());
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
var x: int{size: 32, signed: true, nullable: false} = 10;
x = 20;
print((x).toString());
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
var x: int{size: 32, signed: true, nullable: false} = 10;
x += 5;
print((x).toString());
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
var x: int{size: 32, signed: true, nullable: false} = 10;
var y: int{size: 32, signed: true, nullable: false} = 20;
print((x + y).toString());
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
var x: int{size: 32, signed: true, nullable: true} = null;
print((x).toString());
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
var x: int{size: 8, signed: true, nullable: false} = 1000;
print((x).toString());
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

func TestDivisionByZero(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 10;
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
		srcInt     ast.IntegerType
		srcFloat   ast.FloatType
		srcIsFloat bool
		dstInt     ast.IntegerType
		dstFloat   ast.FloatType
		dstIsFloat bool
		expected   bool
	}{
		{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, true},
		{ast.IntegerType{Size: 8, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, true},
		{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{Size: 8, Signed: true}, ast.FloatType{}, false, false},
		{ast.IntegerType{Size: 32, Signed: false}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: false}, ast.FloatType{}, false, true},
		{ast.IntegerType{Size: 8, Signed: false}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, true},
		// Int to float conversions
		{ast.IntegerType{Size: 8, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{}, ast.FloatType{Size: 16}, true, true},
		{ast.IntegerType{Size: 16, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{}, ast.FloatType{Size: 32}, true, true},
		{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{}, ast.FloatType{Size: 64}, true, true},
		// Float to float conversions
		{ast.IntegerType{}, ast.FloatType{Size: 16}, true,
			ast.IntegerType{}, ast.FloatType{Size: 32}, true, true},
		{ast.IntegerType{}, ast.FloatType{Size: 32}, true,
			ast.IntegerType{}, ast.FloatType{Size: 64}, true, true},
		// Float to int - not allowed
		{ast.IntegerType{}, ast.FloatType{Size: 32}, true,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, false},
	}

	for _, tt := range tests {
		result := canImplicitConvert(tt.srcInt, tt.srcFloat, tt.srcIsFloat, tt.dstInt, tt.dstFloat, tt.dstIsFloat)
		if result != tt.expected {
			t.Errorf("canImplicitConvert(%v, %v, %v, %v, %v, %v) = %v, expected %v",
				tt.srcInt, tt.srcFloat, tt.srcIsFloat, tt.dstInt, tt.dstFloat, tt.dstIsFloat,
				result, tt.expected)
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
		{Value{IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}, Data: 10, IsFloat: false}, "non-nullable 32-bit signed int(10)"},
		{Value{IType: ast.IntegerType{Size: 16, Signed: false, Nullable: false}, Data: 5, IsFloat: false}, "non-nullable 16-bit unsigned int(5)"},
		{Value{FType: ast.FloatType{Size: 32, Nullable: false}, FData: 3.14, IsFloat: true}, "non-nullable 32-bit float(3.14)"},
		{Value{Untyped: true, FData: 2.5, IsFloat: true}, "2.5"},
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
		ftype    ast.FloatType
		isFloat  bool
		expected string
	}{
		{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{}, false, "non-nullable 32-bit signed int"},
		{ast.IntegerType{Size: 16, Signed: false, Nullable: true}, ast.FloatType{}, false, "nullable 16-bit unsigned int"},
		{ast.IntegerType{Size: 64, Signed: true, Nullable: true}, ast.FloatType{}, false, "nullable 64-bit signed int"},
		{ast.IntegerType{}, ast.FloatType{Size: 32, Nullable: false}, true, "non-nullable 32-bit float"},
		{ast.IntegerType{}, ast.FloatType{Size: 64, Nullable: false}, true, "non-nullable 64-bit float"},
	}

	for _, tt := range tests {
		result := typeDescFromVar(tt.itype, tt.ftype, ast.BoolType{}, tt.isFloat, false)
		if result != tt.expected {
			t.Errorf("typeDesc(%v, %v, %v) = %q, expected %q", tt.itype, tt.ftype, tt.isFloat, result, tt.expected)
		}
	}
}

func TestExecuteStmt(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 10;
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
		Op:    "^",
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
var x: int{size: 32, signed: true, nullable: false} = 10;
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
var x: int{size: 32, signed: true, nullable: true} = null;
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

func TestCanImplicitConvertMore(t *testing.T) {
	tests := []struct {
		srcInt     ast.IntegerType
		srcFloat   ast.FloatType
		srcIsFloat bool
		dstInt     ast.IntegerType
		dstFloat   ast.FloatType
		dstIsFloat bool
		expected   bool
	}{
		{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, true},
		{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: false}, ast.FloatType{}, false, false},
		{ast.IntegerType{Size: 64, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, false},
		{ast.IntegerType{Size: 8, Signed: false}, ast.FloatType{}, false,
			ast.IntegerType{Size: 16, Signed: false}, ast.FloatType{}, false, true},
		{ast.IntegerType{Size: 16, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, true},
		// Int to float
		{ast.IntegerType{Size: 8, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{}, ast.FloatType{Size: 16}, true, true},
		{ast.IntegerType{Size: 16, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{}, ast.FloatType{Size: 32}, true, true},
		{ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false,
			ast.IntegerType{}, ast.FloatType{Size: 64}, true, true},
		// Float to float
		{ast.IntegerType{}, ast.FloatType{Size: 16}, true,
			ast.IntegerType{}, ast.FloatType{Size: 32}, true, true},
		{ast.IntegerType{}, ast.FloatType{Size: 32}, true,
			ast.IntegerType{}, ast.FloatType{Size: 64}, true, true},
		// Float to int - not allowed
		{ast.IntegerType{}, ast.FloatType{Size: 32}, true,
			ast.IntegerType{Size: 32, Signed: true}, ast.FloatType{}, false, false},
	}

	for _, tt := range tests {
		result := canImplicitConvert(tt.srcInt, tt.srcFloat, tt.srcIsFloat, tt.dstInt, tt.dstFloat, tt.dstIsFloat)
		if result != tt.expected {
			t.Errorf("canImplicitConvert(%v, %v, %v, %v, %v, %v) = %v, expected %v",
				tt.srcInt, tt.srcFloat, tt.srcIsFloat, tt.dstInt, tt.dstFloat, tt.dstIsFloat,
				result, tt.expected)
		}
	}
}

func TestAssignmentWithVarRef(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 10;
var y: int{size: 32, signed: true, nullable: false} = x;
print((y).toString());
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
var x: int{size: 8, signed: true, nullable: false} = 10;
var y: int{size: 32, signed: true, nullable: false} = x;
print((y).toString());
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
var x: int{size: 8, signed: true, nullable: false} = 10;
x = 200;
print((x).toString());
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
var x: int{size: 16, signed: true, nullable: false} = 1000;
var y: int{size: 32, signed: true, nullable: false} = 0;
y = x;
print((y).toString());
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
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != -112 {
		t.Errorf("expected -112, got %d", val.Data)
	}
}

func TestExecuteAssignmentNullAssign(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: true} = 10;
x = null;
print((x).toString());
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
	i.env.Define("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 42})

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
var x: int{size: 32, signed: true, nullable: false} = 10;
x = 20;
print((x).toString());
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

func TestExecuteAssignmentWithUntypedExpr(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 10;
x = x + 5;
print((x).toString());
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
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != -112 {
		t.Errorf("expected -112, got %d", val.Data)
	}
}

func TestExecuteAssignmentWithTypeMismatch(t *testing.T) {
	input := `
var x: int{size: 8, signed: true, nullable: false} = 10;
var y: int{size: 32, signed: true, nullable: false} = 1000;
x = y;
print((x).toString());
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
	if !strings.Contains(output, "-24") {
		t.Errorf("expected output to contain '-24', got %q", output)
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
	i.env.Define("y", Value{IType: ast.IntegerType{Size: 16, Signed: true}, Data: 200})
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
var x: int{size: 32, signed: true, nullable: false} = 10;
x -= 5;
print((x).toString());
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
var x: int{size: 32, signed: true, nullable: false} = 10;
x *= 5;
print((x).toString());
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
var x: int{size: 32, signed: true, nullable: false} = 10;
x /= 2;
print((x).toString());
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
var x: int{size: 32, signed: true, nullable: true} = 10;
x = null;
print((x).toString());
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
	i.env.Define("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 10})
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

func TestFloatVarDecl(t *testing.T) {
	input := `
var a: float{size: 32} = 3.14;
print((a).toString());
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

	if !strings.Contains(output, "3.14") {
		t.Errorf("expected output to contain '3.14', got %q", output)
	}
}

func TestFloat16VarDecl(t *testing.T) {
	input := `
var a: float{size: 16} = 3.14;
print((a).toString());
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

	if !strings.Contains(output, "3.14") {
		t.Errorf("expected output to contain '3.14', got %q", output)
	}
}

func TestFloat64VarDecl(t *testing.T) {
	input := `
var a: float{size: 64} = 3.14;
print((a).toString());
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

	if !strings.Contains(output, "3.14") {
		t.Errorf("expected output to contain '3.14', got %q", output)
	}
}

func TestIntToFloatConversion(t *testing.T) {
	input := `
var a: int{size: 32} = 42;
var b: float{size: 64} = a;
print((b).toString());
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

func TestFloatToFloatConversion(t *testing.T) {
	input := `
var a: float{size: 16} = 1.5;
var b: float{size: 32} = a;
print((b).toString());
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

	if !strings.Contains(output, "1.5") {
		t.Errorf("expected output to contain '1.5', got %q", output)
	}
}

func TestFloatToFloatUpConversion(t *testing.T) {
	input := `
var a: float{size: 32} = 1.5;
var b: float{size: 64} = a;
print((b).toString());
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

	if !strings.Contains(output, "1.5") {
		t.Errorf("expected output to contain '1.5', got %q", output)
	}
}

func TestFloatAssignment(t *testing.T) {
	input := `
var a: float{size: 32} = 1.5;
a = 2.5;
print((a).toString());
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

	if !strings.Contains(output, "2.5") {
		t.Errorf("expected output to contain '2.5', got %q", output)
	}
}

func TestFloatBinaryExpr(t *testing.T) {
	input := `
var a: float{size: 32} = 1.5;
var b: float{size: 32} = 2.5;
print((a + b).toString());
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

	if !strings.Contains(output, "4") {
		t.Errorf("expected output to contain '4', got %q", output)
	}
}

func TestFloatNullAssign(t *testing.T) {
	input := `
var a: float{size: 32, nullable: true} = null;
print((a).toString());
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

func TestFloatInvalidSize(t *testing.T) {
	input := `
var a: float{size: 128} = 1.0;
`
	l := lexer.New(input)
	p := parser.New(l)
	p.ParseProgram()

	// Check if parser caught the error
	if len(p.Errors()) == 0 {
		t.Errorf("expected parser error for invalid float size")
	}
}

func TestFloatTypeDesc(t *testing.T) {
	tests := []struct {
		ftype    ast.FloatType
		expected string
	}{
		{ast.FloatType{Size: 16}, "16-bit float"},
		{ast.FloatType{Size: 32}, "32-bit float"},
		{ast.FloatType{Size: 64}, "64-bit float"},
	}

	for _, tt := range tests {
		// Test via Value.String()
		v := Value{FType: tt.ftype, IsFloat: true, FData: 1.5}
		result := v.String()
		if !strings.Contains(result, tt.expected) {
			t.Errorf("Value.String() = %q, expected to contain %q", result, tt.expected)
		}
	}
}

func TestEvalFloatLit(t *testing.T) {
	i := New()
	expr := &ast.FloatLit{Value: 3.14, Untyped: true}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsFloat {
		t.Errorf("expected IsFloat to be true")
	}
	if val.FData != 3.14 {
		t.Errorf("expected 3.14, got %g", val.FData)
	}
}

func TestDotLeadingFloat(t *testing.T) {
	input := "print((.1).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "0.1" {
		t.Errorf("expected 0.1, got %q", output)
	}
}

func TestDotLeadingFloatNeg(t *testing.T) {
	input := "print((-.5).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "-0.5" {
		t.Errorf("expected -0.5, got %q", output)
	}
}

func TestDotLeadingFloatAdd(t *testing.T) {
	input := "print((.15 + .1).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "0.25" {
		t.Errorf("expected 0.25, got %q", output)
	}
}

func TestDotLeadingFloatExp(t *testing.T) {
	input := "print((.1e2).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "10" {
		t.Errorf("expected 10, got %q", output)
	}
}

func TestEvalFloatBinaryExpr(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.5, Untyped: true},
		Op:    "+",
		Right: &ast.FloatLit{Value: 2.5, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsFloat {
		t.Errorf("expected IsFloat to be true")
	}
	if val.FData != 4.0 {
		t.Errorf("expected 4.0, got %g", val.FData)
	}
}

func TestFloatCompoundAssignmentAdd(t *testing.T) {
	input := `
var a: float{size: 32} = 1.5;
a += 2.5;
print((a).toString());
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

	if !strings.Contains(output, "4") {
		t.Errorf("expected output to contain '4', got %q", output)
	}
}

func TestFloatCompoundAssignmentSub(t *testing.T) {
	input := `
var a: float{size: 32} = 5.5;
a -= 2.0;
print((a).toString());
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

	if !strings.Contains(output, "3.5") {
		t.Errorf("expected output to contain '3.5', got %q", output)
	}
}

func TestFloatCompoundAssignmentMul(t *testing.T) {
	input := `
var a: float{size: 32} = 2.0;
a *= 3.5;
print((a).toString());
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

	if !strings.Contains(output, "7") {
		t.Errorf("expected output to contain '7', got %q", output)
	}
}

func TestFloatCompoundAssignmentDiv(t *testing.T) {
	input := `
var a: float{size: 32} = 10.0;
a /= 2.0;
print((a).toString());
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

func TestNaNLiteral(t *testing.T) {
	input := `
var a: float{size: 64} = NaN;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "NaN" {
		t.Errorf("expected 'NaN', got %q", output)
	}
}

func TestInfinityLiteral(t *testing.T) {
	input := `
var a: float{size: 64} = infinity;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "infinity" {
		t.Errorf("expected 'infinity', got %q", output)
	}
}

func TestNegInfinityLiteral(t *testing.T) {
	input := `
var a: float{size: 64} = -infinity;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "-infinity" {
		t.Errorf("expected '-infinity', got %q", output)
	}
}

func TestFloatDivisionByZero(t *testing.T) {
	input := `
var a: float{size: 32} = 10.0;
a /= 0.0;
print((a).toString());
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
	if !strings.Contains(output, "infinity") {
		t.Errorf("expected output to contain 'infinity', got %q", output)
	}
}

func TestFloatBinaryExprSub(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 5.5, Untyped: true},
		Op:    "-",
		Right: &ast.FloatLit{Value: 2.0, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FData != 3.5 {
		t.Errorf("expected 3.5, got %g", val.FData)
	}
}

func TestFloatBinaryExprMul(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 2.5, Untyped: true},
		Op:    "*",
		Right: &ast.FloatLit{Value: 3.0, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FData != 7.5 {
		t.Errorf("expected 7.5, got %g", val.FData)
	}
}

func TestFloatBinaryExprDiv(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 10.0, Untyped: true},
		Op:    "/",
		Right: &ast.FloatLit{Value: 4.0, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FData != 2.5 {
		t.Errorf("expected 2.5, got %g", val.FData)
	}
}

func TestFloatBinaryExprDivByZero(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 10.0, Untyped: true},
		Op:    "/",
		Right: &ast.FloatLit{Value: 0.0, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsInf(val.FData, 1) {
		t.Errorf("expected +Inf, got %g", val.FData)
	}
}

func TestMixedIntFloatExpr(t *testing.T) {
	input := `
var a: int{size: 32} = 10;
var b: float{size: 32} = 2.5;
print((a + b).toString());
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

	if !strings.Contains(output, "12.5") {
		t.Errorf("expected output to contain '12.5', got %q", output)
	}
}

func TestMixedFloatIntExpr(t *testing.T) {
	input := `
var a: float{size: 32} = 2.5;
var b: int{size: 32} = 10;
print((a + b).toString());
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

	if !strings.Contains(output, "12.5") {
		t.Errorf("expected output to contain '12.5', got %q", output)
	}
}

func TestIntToFloatCompoundAssignment(t *testing.T) {
	input := `
var a: float{size: 32} = 2.5;
a += 3;
print((a).toString());
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

	if !strings.Contains(output, "5.5") {
		t.Errorf("expected output to contain '5.5', got %q", output)
	}
}

func TestFloatOverflowDetection(t *testing.T) {
	input := `
var a: float{size: 16} = 65504.0;
a += 1.0;
print((a).toString());
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
	if !strings.Contains(output, "65504") {
		t.Errorf("expected output to contain '65504', got %q", output)
	}
}

func TestFloatAssignmentOverflow(t *testing.T) {
	input := `
var a: float{size: 16} = 70000.0;
print((a).toString());
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

func TestNullableFloatAssignment(t *testing.T) {
	input := `
var a: float{size: 32, nullable: true} = null;
a = 3.14;
print((a).toString());
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

	if !strings.Contains(output, "3.14") {
		t.Errorf("expected output to contain '3.14', got %q", output)
	}
}

func TestNullableFloatCompoundAssignment(t *testing.T) {
	input := `
var a: float{size: 32, nullable: true} = 1.5;
a += 2.5;
print((a).toString());
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

	if !strings.Contains(output, "4") {
		t.Errorf("expected output to contain '4', got %q", output)
	}
}

func TestCanImplicitConvertInt64ToFloat(t *testing.T) {
	// int64 should not implicitly convert to float
	result := canImplicitConvert(
		ast.IntegerType{Size: 64, Signed: true}, ast.FloatType{}, false,
		ast.IntegerType{}, ast.FloatType{Size: 64}, true,
	)
	if result {
		t.Errorf("expected int64 to not convert to float")
	}
}

func TestCanImplicitConvertFloatToFloatDowngrade(t *testing.T) {
	// float64 to float32 should not be allowed
	result := canImplicitConvert(
		ast.IntegerType{}, ast.FloatType{Size: 64}, true,
		ast.IntegerType{}, ast.FloatType{Size: 32}, true,
	)
	if result {
		t.Errorf("expected float64 to not convert to float32")
	}
}

func TestCanImplicitConvertNullableToInt(t *testing.T) {
	// nullable integer should not convert to non-nullable integer
	result := canImplicitConvert(
		ast.IntegerType{Size: 32, Signed: true, Nullable: true}, ast.FloatType{}, false,
		ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{}, false,
	)
	if result {
		t.Errorf("expected nullable int to not convert to non-nullable int")
	}
}

func TestCanImplicitConvertNullableFloatToFloat(t *testing.T) {
	// nullable float should not convert to non-nullable float
	result := canImplicitConvert(
		ast.IntegerType{}, ast.FloatType{Size: 32, Nullable: true}, true,
		ast.IntegerType{}, ast.FloatType{Size: 32, Nullable: false}, true,
	)
	if result {
		t.Errorf("expected nullable float to not convert to non-nullable float")
	}
}

func TestCanImplicitConvertNullableToIntSameType(t *testing.T) {
	// nullable to non-nullable should fail even if same size/signedness
	result := canImplicitConvert(
		ast.IntegerType{Size: 32, Signed: true, Nullable: true}, ast.FloatType{}, false,
		ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{}, false,
	)
	if result {
		t.Errorf("expected nullable to non-nullable to fail")
	}
}

func TestCanImplicitConvertNonNullableToNullable(t *testing.T) {
	// non-nullable to nullable should be allowed
	result := canImplicitConvert(
		ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{}, false,
		ast.IntegerType{Size: 32, Signed: true, Nullable: true}, ast.FloatType{}, false,
	)
	if !result {
		t.Errorf("expected non-nullable int to convert to nullable int")
	}
}

func TestCanImplicitConvertNonNullableFloatToNullableFloat(t *testing.T) {
	// non-nullable float to nullable float should be allowed
	result := canImplicitConvert(
		ast.IntegerType{}, ast.FloatType{Size: 32, Nullable: false}, true,
		ast.IntegerType{}, ast.FloatType{Size: 32, Nullable: true}, true,
	)
	if !result {
		t.Errorf("expected non-nullable float to convert to nullable float")
	}
}

func TestEnvironmentNewAndSetGet(t *testing.T) {
	env := NewEnv()
	val := Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 42}
	env.Define("x", val)

	got, ok := env.Get("x")
	if !ok {
		t.Fatalf("expected to find variable 'x'")
	}
	if got.Data != 42 {
		t.Errorf("expected 42, got %d", got.Data)
	}
}

func TestEnvironmentUndefinedVar(t *testing.T) {
	env := NewEnv()
	_, ok := env.Get("undefined")
	if ok {
		t.Errorf("expected variable 'undefined' to not exist")
	}
}

func TestRunMultipleStatements(t *testing.T) {
	input := `
var x: int{size: 32} = 10;
var y: int{size: 32} = 20;
print((x + y).toString());
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

func TestValueStringNullableFloat(t *testing.T) {
	v := Value{FType: ast.FloatType{Size: 32, Nullable: true}, FData: 3.14, IsFloat: true}
	result := v.String()
	if !strings.Contains(result, "nullable") {
		t.Errorf("expected string to contain 'nullable', got %q", result)
	}
}

func TestValueStringNullableInt(t *testing.T) {
	v := Value{IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true}, Data: 42, IsFloat: false}
	result := v.String()
	if !strings.Contains(result, "nullable") {
		t.Errorf("expected string to contain 'nullable', got %q", result)
	}
}

func TestExecuteStmtNullableWithoutExpr(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true},
		Expr:  nil,
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that variable was set with null
	val, ok := i.env.Get("x")
	if !ok {
		t.Fatalf("expected variable 'x' to exist")
	}
	if !val.Null {
		t.Errorf("expected null value")
	}
}

func TestExecuteAssignmentWithNullToNullable(t *testing.T) {
	i := New()
	i.env.Define("x", Value{IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true}, Data: 10})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.NullLit{},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, _ := i.env.Get("x")
	if !val.Null {
		t.Errorf("expected null value after assignment")
	}
}

func TestEvalBinaryDefaultOp(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "^", // unknown operator
		Right: &ast.IntegerLit{Value: 3, Untyped: true},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected error for unknown operator")
	}
}

func TestValueStringNonNullableFloat(t *testing.T) {
	v := Value{FType: ast.FloatType{Size: 32}, FData: 3.14, IsFloat: true}
	result := v.String()
	if !strings.Contains(result, "32-bit float") {
		t.Errorf("expected '32-bit float' in %q", result)
	}
}

func TestExecuteStmtVarDeclWithNullableNoExpr(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true},
		Expr:  nil,
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := i.env.Get("x")
	if !ok || !val.Null {
		t.Errorf("expected null variable")
	}
}

func TestExecuteAssignmentNoOp(t *testing.T) {
	i := New()
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "?",
		Expr: &ast.IntegerLit{Value: 5},
	}
	i.env.Define("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 10})
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for unknown operator")
	}
}

func TestStringNonNullableFloat(t *testing.T) {
	v := Value{FType: ast.FloatType{Size: 32}, FData: 3.14, IsFloat: true}
	result := v.String()
	if !strings.Contains(result, "32-bit float") {
		t.Errorf("expected '32-bit float' in %q", result)
	}
}

func TestExecuteAssignmentFloatAdd(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "+=",
		Expr: &ast.FloatLit{Value: 2.5, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 4.0 {
		t.Errorf("expected 4.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentFloatSub(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 5.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "-=",
		Expr: &ast.FloatLit{Value: 2.0, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 3.5 {
		t.Errorf("expected 3.5, got %g", val.FData)
	}
}

func TestExecuteAssignmentFloatMul(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 2.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "*=",
		Expr: &ast.FloatLit{Value: 3.0, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 6.0 {
		t.Errorf("expected 6.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentFloatDiv(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "/=",
		Expr: &ast.FloatLit{Value: 2.0, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 5.0 {
		t.Errorf("expected 5.0, got %g", val.FData)
	}
}

func TestEvalExprDefaultCase(t *testing.T) {
	// The default case in evalExpr handles unknown expression types
	// This is difficult to test in Go without creating a type that implements ast.Expr
	// but isn't one of the handled types
	// Skip this test as it requires modifying the AST package
}

func TestVarDeclWithFloatExpr(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 3.14, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if math.Abs(val.FData-3.14) > 1e-6 {
		t.Errorf("expected ~3.14, got %g", val.FData)
	}
}

func TestVarDeclWithIntToFloatConversion(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.IntegerLit{Value: 42, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 42.0 {
		t.Errorf("expected 42.0, got %g", val.FData)
	}
}

func TestVarDeclWithFloatTypeMismatch(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 32, Signed: true},
		Expr:  &ast.FloatLit{Value: 3.14, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestExecuteAssignmentIntToFloatAdd(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "+=",
		Expr: &ast.IntegerLit{Value: 2},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 3.5 {
		t.Errorf("expected 3.5, got %g", val.FData)
	}
}

func TestExecuteAssignmentIntToFloatSub(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 5.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "-=",
		Expr: &ast.IntegerLit{Value: 2},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 3.5 {
		t.Errorf("expected 3.5, got %g", val.FData)
	}
}

func TestExecuteAssignmentIntToFloatMul(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 2.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "*=",
		Expr: &ast.IntegerLit{Value: 3},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 6.0 {
		t.Errorf("expected 6.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentIntToFloatDiv(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "/=",
		Expr: &ast.IntegerLit{Value: 2},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 5.0 {
		t.Errorf("expected 5.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentFloatUnknownOp(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%",
		Expr: &ast.FloatLit{Value: 2.5},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for unknown operator")
	}
}

func TestEvalBinaryFloatTypedBoth(t *testing.T) {
	i := New()
	// Both operands are typed floats
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.5, FType: ast.FloatType{Size: 32}, Untyped: false},
		Op:    "+",
		Right: &ast.FloatLit{Value: 2.5, FType: ast.FloatType{Size: 64}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FType.Size != 64 {
		t.Errorf("expected FType.Size 64, got %d", val.FType.Size)
	}
}

func TestEvalBinaryFloatTypedLeft(t *testing.T) {
	i := New()
	// Left is typed float, right is untyped - result should be untyped
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.5, FType: ast.FloatType{Size: 32}, Untyped: false},
		Op:    "+",
		Right: &ast.FloatLit{Value: 2.5, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Untyped {
		t.Errorf("expected Untyped to be true")
	}
}

func TestEvalBinaryFloatTypedRight(t *testing.T) {
	i := New()
	// Left is untyped, right is typed float - result should be untyped
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.5, Untyped: true},
		Op:    "+",
		Right: &ast.FloatLit{Value: 2.5, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Untyped {
		t.Errorf("expected Untyped to be true")
	}
}

func TestEvalBinaryFloatOverflow(t *testing.T) {
	i := New()
	// Create a float result that overflows float16
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 65504.0, FType: ast.FloatType{Size: 16}, Untyped: false},
		Op:    "+",
		Right: &ast.FloatLit{Value: 1.0, FType: ast.FloatType{Size: 16}, Untyped: false},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FData != 65504.0 {
		t.Errorf("expected 65504, got %g", val.FData)
	}
}

func TestVarDeclFloatToFloatConversion(t *testing.T) {
	i := New()
	// Assign float to float variable with implicit conversion
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 64},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 1.5, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVarDeclFloatOverflow(t *testing.T) {
	i := New()
	// Float value overflows the target type
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 16},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 70000.0, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestVarDeclIntToFloatConversion(t *testing.T) {
	i := New()
	// Assign integer to float variable
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.IntegerLit{Value: 42, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVarDeclIntToFloatTypeMismatch(t *testing.T) {
	i := New()
	// Integer (64-bit unsigned) cannot implicitly convert to float32
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.IntegerLit{Value: 42, IType: ast.IntegerType{Size: 64}, Untyped: false},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestExecuteAssignmentIntToFloatEq(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.IntegerLit{Value: 2},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 2.0 {
		t.Errorf("expected 2.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentIntToFloatAddEq(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "+=",
		Expr: &ast.IntegerLit{Value: 2},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 3.5 {
		t.Errorf("expected 3.5, got %g", val.FData)
	}
}

func TestExecuteAssignmentIntToFloatSubEq(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 5.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "-=",
		Expr: &ast.IntegerLit{Value: 2},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 3.5 {
		t.Errorf("expected 3.5, got %g", val.FData)
	}
}

func TestExecuteAssignmentIntToFloatMulEq(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 2.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "*=",
		Expr: &ast.IntegerLit{Value: 3},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 6.0 {
		t.Errorf("expected 6.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentIntToFloatDivEq(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "/=",
		Expr: &ast.IntegerLit{Value: 2},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 5.0 {
		t.Errorf("expected 5.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentFloatDivByZero(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "/=",
		Expr: &ast.FloatLit{Value: 0.0, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if !math.IsInf(val.FData, 1) {
		t.Errorf("expected +Inf, got %g", val.FData)
	}
}

func TestExecuteAssignmentFloatDefaultOp(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%",
		Expr: &ast.FloatLit{Value: 2.5, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected unknown operator error")
	}
}

func TestVarDeclFloatWithIntegerUntyped(t *testing.T) {
	i := New()
	// rightVal is float but Untyped, so rightVal.IsFloat is true
	// This should go through the rightVal.IsFloat path
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.IntegerLit{Value: 42, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 42.0 {
		t.Errorf("expected 42.0, got %g", val.FData)
	}
}

func TestVarDeclFloatTypeMismatch(t *testing.T) {
	i := New()
	// Integer (typed) cannot convert to float
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.IntegerLit{Value: 42, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestVarDeclTypedIntToFloatError(t *testing.T) {
	i := New()
	// Integer (typed) cannot convert to float
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.IntegerLit{Value: 42, IType: ast.IntegerType{Size: 32, Signed: true}, Untyped: false},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestVarDeclRightValUntypedIntToFloat(t *testing.T) {
	i := New()
	// rightVal is untyped integer, rightVal.IsFloat is false
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.IntegerLit{Value: 42, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.FData != 42.0 {
		t.Errorf("expected 42.0, got %g", val.FData)
	}
}

func TestExecuteAssignmentDivByZeroInt(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
	// rightVal is integer with Data=0
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "/=",
		Expr: &ast.IntegerLit{Value: 0},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if !math.IsInf(val.FData, 1) {
		t.Errorf("expected +Inf, got %g", val.FData)
	}
}

func TestVarDeclFloatOverflowInConversion(t *testing.T) {
	i := New()
	// rightVal is float, convert to float with smaller size
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 16},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 70000.0, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestVarDeclFloatToFloatOverflow(t *testing.T) {
	i := New()
	// Float value too large for target type
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 16},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 70000.0, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestExecuteAssignmentIntToFloatTypeMismatch(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, IsFloat: true})
	// Integer that cannot convert to float
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.IntegerLit{Value: 42, IType: ast.IntegerType{Size: 64}, Untyped: false},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestExecuteAssignmentFloatToFloatOverflow(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 16}, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.FloatLit{Value: 70000.0, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if !math.IsInf(val.FData, 1) {
		t.Errorf("expected +Inf, got %g", val.FData)
	}
}

func TestExecuteAssignmentFloatAddOverflow(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 16}, FData: 60000.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "+=",
		Expr: &ast.FloatLit{Value: 10000.0, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAssignmentFloatDivOverflow(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 16}, FData: 60000.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "/=",
		Expr: &ast.FloatLit{Value: 0.5, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAssignmentFloatToIntError(t *testing.T) {
	i := New()
	i.env.Define("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, IsFloat: false})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.FloatLit{Value: 3.14, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestVarDeclFloatOverflowCheck(t *testing.T) {
	i := New()
	// rightVal is float with value that overflows target FType
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 16},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 70000.0, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestExecuteAssignmentFloatEqOverflow(t *testing.T) {
	i := New()
	// Assign float value that overflows target type
	i.env.Define("x", Value{FType: ast.FloatType{Size: 16}, FData: 100.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.FloatLit{Value: 70000.0, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVarDeclFloatOverflowInCheck(t *testing.T) {
	i := New()
	// rightVal is float with value 70000, target is float16
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 16},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 70000.0, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

// Test for executeStmt - VarDecl with float conversion that overflows
func TestVarDeclFloatOverflowCoverage(t *testing.T) {
	i := New()
	// rightVal is float with value 70000, target is float16
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 16},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 70000.0, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

// Test for executeAssignment - default operator case (line 329-330)
func TestExecuteAssignmentDefaultOpCoverage(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%", // unknown operator
		Expr: &ast.FloatLit{Value: 2.5, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected unknown operator error")
	}
}

func TestExecuteAssignmentFloatEqOverflowCoverage(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 16}, FData: 1.0, IsFloat: true})
	// rightVal is float32 with value 70000, result overflows float16
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.FloatLit{Value: 70000.0, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteAssignmentFloatOverflowCoverage(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 16}, FData: 1.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.FloatLit{Value: 70000.0, FType: ast.FloatType{Size: 32}, Untyped: false},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScientificNotationLiteral(t *testing.T) {
	input := `
var a: float{size: 64} = 1e20;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "1e+20") {
		t.Errorf("expected output to contain '1e+20', got %q", output)
	}
}

func TestUnderscoreIntegerLiteral(t *testing.T) {
	input := `
var x: int{size: 64} = 100_000;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "100000") {
		t.Errorf("expected output to contain '100000', got %q", output)
	}
}

func TestBinaryLiteralInInterp(t *testing.T) {
	input := `
var x: int{size: 64} = 0b1010;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

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

func TestOctalLiteralInInterp(t *testing.T) {
	input := `
var x: int{size: 64} = 0o777;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "511") {
		t.Errorf("expected output to contain '511', got %q", output)
	}
}

func TestHexLiteralInInterp(t *testing.T) {
	input := `
var x: int{size: 64} = 0xFF;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "255") {
		t.Errorf("expected output to contain '255', got %q", output)
	}
}

func TestUnsignedInt8Overflow(t *testing.T) {
	input := `
var x: int{size: 8, signed: false} = 256;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestUnsignedInt16Overflow(t *testing.T) {
	input := `
var x: int{size: 16, signed: false} = 65536;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestUnsignedInt32Overflow(t *testing.T) {
	input := `
var x: int{size: 32, signed: false} = 4294967296;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestUnsignedIntUnderflowSub(t *testing.T) {
	input := `
var x: int{size: 8, signed: false} = 0;
x = x - 1;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "255") {
		t.Errorf("expected output to contain '255', got %q", output)
	}
}

func TestPrefixedHexFloatEval(t *testing.T) {
	input := `
var a: float{size: 64} = 0xf.f;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "15.9375") {
		t.Errorf("expected output to contain '15.9375', got %q", output)
	}
}

func TestPrefixedBinFloatEval(t *testing.T) {
	input := `
var a: float{size: 64} = 0b1.01;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "1.25") {
		t.Errorf("expected output to contain '1.25', got %q", output)
	}
}

func TestPrefixedOctFloatEval(t *testing.T) {
	input := `
var a: float{size: 64} = 0o7.7;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "7.875") {
		t.Errorf("expected output to contain '7.875', got %q", output)
	}
}

func TestPrefixedHexFloatNoIntPartEval(t *testing.T) {
	input := `
var a: float{size: 64} = 0x.1;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "0.0625") {
		t.Errorf("expected output to contain '0.0625', got %q", output)
	}
}

func TestConvertFloatInvalidSize(t *testing.T) {
	// convertFloat with invalid size should return the value unchanged (default case)
	result := convertFloat(3.14, ast.FloatType{Size: 123})
	if result != 3.14 {
		t.Errorf("expected 3.14, got %g", result)
	}
}

func TestConvertFloatAllSizes(t *testing.T) {
	// float16
	result := convertFloat(1.0, ast.FloatType{Size: 16})
	if result != 1.0 {
		t.Errorf("float16: expected 1.0, got %g", result)
	}

	// float32
	result = convertFloat(3.14, ast.FloatType{Size: 32})
	if result != float64(float32(3.14)) {
		t.Errorf("float32: expected %g, got %g", float64(float32(3.14)), result)
	}

	// float64
	result = convertFloat(3.14, ast.FloatType{Size: 64})
	if result != 3.14 {
		t.Errorf("float64: expected 3.14, got %g", result)
	}
}

func TestTypeDescUntypedInteger(t *testing.T) {
	result := typeDescFromVar(ast.IntegerType{Size: 0}, ast.FloatType{}, ast.BoolType{}, false, false)
	if result != "untyped int literal" {
		t.Errorf("expected 'untyped integer literal', got %q", result)
	}
}

func TestExecuteStmtTypedFloatMismatch(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:    "x",
		FType:   ast.FloatType{Size: 32},
		IsFloat: true,
		Expr:    &ast.FloatLit{Value: 1.5, FType: ast.FloatType{Size: 64}, Untyped: false},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type mismatch error")
	}
}

func TestExecuteAssignmentUnknownOpIntToFloat(t *testing.T) {
	i := New()
	i.env.Define("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.0, IsFloat: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%=",
		Expr: &ast.IntegerLit{Value: 5, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected unknown operator error")
	}
}

func TestConvertIntAllSizes(t *testing.T) {
	// signed int8
	result := convertInt(42, ast.IntegerType{Size: 8, Signed: true})
	if result != 42 {
		t.Errorf("int8: expected 42, got %d", result)
	}

	// unsigned uint8
	result = convertInt(255, ast.IntegerType{Size: 8, Signed: false})
	if result != 255 {
		t.Errorf("uint8: expected 255, got %d", result)
	}

	// signed int16
	result = convertInt(42, ast.IntegerType{Size: 16, Signed: true})
	if result != 42 {
		t.Errorf("int16: expected 42, got %d", result)
	}

	// unsigned uint16
	result = convertInt(42, ast.IntegerType{Size: 16, Signed: false})
	if result != 42 {
		t.Errorf("uint16: expected 42, got %d", result)
	}

	// signed int32
	result = convertInt(42, ast.IntegerType{Size: 32, Signed: true})
	if result != 42 {
		t.Errorf("int32: expected 42, got %d", result)
	}

	// unsigned uint32
	result = convertInt(42, ast.IntegerType{Size: 32, Signed: false})
	if result != 42 {
		t.Errorf("uint32: expected 42, got %d", result)
	}

	// signed int64
	result = convertInt(42, ast.IntegerType{Size: 64, Signed: true})
	if result != 42 {
		t.Errorf("int64: expected 42, got %d", result)
	}

	// unsigned uint64
	result = convertInt(42, ast.IntegerType{Size: 64, Signed: false})
	if result != 42 {
		t.Errorf("uint64: expected 42, got %d", result)
	}

	// invalid size (default case)
	result = convertInt(42, ast.IntegerType{Size: 123, Signed: true})
	if result != 42 {
		t.Errorf("default: expected 42, got %d", result)
	}
}

func TestBoolVarDeclTrue(t *testing.T) {
	input := `var x: bool = true;
print((x).toString());`
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
	if !strings.Contains(output, "true") {
		t.Errorf("expected output 'true', got %q", output)
	}
}

func TestBoolVarDeclFalse(t *testing.T) {
	input := `var x: bool = false;
print((x).toString());`
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
	if !strings.Contains(output, "false") {
		t.Errorf("expected output 'false', got %q", output)
	}
}

func TestBoolVarDeclDefaultFalse(t *testing.T) {
	input := `var x: bool{nullable: false};
print((x).toString());`
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
	if !strings.Contains(output, "false") {
		t.Errorf("expected output 'false', got %q", output)
	}
}

func TestBoolVarDeclNullableDefaultNull(t *testing.T) {
	input := `var x: bool;
print((x).toString());`
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
		t.Errorf("expected output 'null', got %q", output)
	}
}

func TestBoolVarDeclAssignNull(t *testing.T) {
	input := `var x: bool{nullable: true} = null;
print((x).toString());`
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
		t.Errorf("expected output 'null', got %q", output)
	}
}

func TestBoolAssignNullToNonNullableError(t *testing.T) {
	input := `var x: bool{nullable: false} = null;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error assigning null to non-nullable bool")
	}
}

func TestBoolAssignNullableVarToNonNullableError(t *testing.T) {
	input := `var a: bool{nullable: true} = true;
var b: bool{nullable: false} = a;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error assigning nullable bool to non-nullable bool")
	}
}

func TestBoolAssignNullableToNonNullableViaAssignmentError(t *testing.T) {
	input := `var a: bool{nullable: true} = true;
var b: bool{nullable: false} = false;
b = a;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error assigning nullable bool to non-nullable bool")
	}
}

func TestBoolAssignment(t *testing.T) {
	input := `var x: bool{nullable: false} = true;
x = false;
print((x).toString());`
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
	if !strings.Contains(output, "false") {
		t.Errorf("expected output 'false', got %q", output)
	}
}

func TestBoolRejectCompoundAssignment(t *testing.T) {
	input := `var x: bool{nullable: false} = true;
x += false;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for compound assignment on bool")
	}
}

func TestBoolAssignNoIntToBoolConversion(t *testing.T) {
	input := `var x: bool{nullable: false} = 42;
print((x).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for int to bool conversion, got none")
	}
}

func TestPrintBoolLiteral(t *testing.T) {
	input := `print((true).toString());
print((false).toString());`
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
	lines := strings.Split(output, "\n")
	if len(lines) != 2 || lines[0] != "true" || lines[1] != "false" {
		t.Errorf("expected 'true\\nfalse', got %q", output)
	}
}

func TestBoolValueString(t *testing.T) {
	tests := []struct {
		val      Value
		expected string
	}{
		{Value{BType: ast.BoolType{Nullable: false}, BData: true, IsBool: true}, "non-nullable bool(true)"},
		{Value{BType: ast.BoolType{Nullable: true}, BData: false, IsBool: true}, "nullable bool(false)"},
		{Value{Untyped: true, BData: true, IsBool: true}, "true"},
		{Value{Untyped: true, BData: false, IsBool: true}, "false"},
	}

	for _, tt := range tests {
		result := tt.val.String()
		if result != tt.expected {
			t.Errorf("Value.String() = %q, expected %q", result, tt.expected)
		}
	}
}

func TestBoolNullableAssignTrue(t *testing.T) {
	input := `var x: bool{nullable: true} = true;
print((x).toString());`
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
	if !strings.Contains(output, "true") {
		t.Errorf("expected output 'true', got %q", output)
	}
}

func TestBoolVarRefAssign(t *testing.T) {
	input := `var a: bool{nullable: false} = true;
var b: bool{nullable: false} = a;
print((b).toString());`
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
	if !strings.Contains(output, "true") {
		t.Errorf("expected output 'true', got %q", output)
	}
}

func TestBoolLitEval(t *testing.T) {
	i := New()
	expr := &ast.BoolLit{Value: true, Untyped: true}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsBool {
		t.Errorf("expected IsBool to be true")
	}
	if val.BData != true {
		t.Errorf("expected true, got %t", val.BData)
	}
	if !val.Untyped {
		t.Errorf("expected untyped")
	}
}

func TestBoolLitTypedEval(t *testing.T) {
	i := New()
	expr := &ast.BoolLit{Value: false, BType: ast.BoolType{Nullable: false}, Untyped: false}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsBool {
		t.Errorf("expected IsBool to be true")
	}
	if val.BData != false {
		t.Errorf("expected false, got %t", val.BData)
	}
	if val.Untyped {
		t.Errorf("expected typed")
	}
	if val.BType.Nullable {
		t.Errorf("expected nullable false")
	}
}

func TestBoolDeclWithExprError(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:   "x",
		BType:  ast.BoolType{Nullable: false},
		IsBool: true,
		Expr:   &ast.PrintStmt{},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error from invalid expression")
	}
}

func TestBoolAssignToNonNullableFromNonNullable(t *testing.T) {
	input := `var a: bool{nullable: false} = true;
var b: bool{nullable: false} = a;
print((b).toString());`
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
	if !strings.Contains(output, "true") {
		t.Errorf("expected output 'true', got %q", output)
	}
}

func TestTypeDescFromVarNullableFloat(t *testing.T) {
	result := typeDescFromVar(ast.IntegerType{}, ast.FloatType{Size: 32, Nullable: true}, ast.BoolType{}, true, false)
	if !strings.Contains(result, "nullable") {
		t.Errorf("expected 'nullable' in %q", result)
	}
}

func TestTypeDescUntypedFloat(t *testing.T) {
	result := typeDesc(ast.FloatType{Size: 0}, false)
	if result != "untyped float literal" {
		t.Errorf("expected 'untyped float literal', got %q", result)
	}
}

func TestTypeDescUntypedIntegerNew(t *testing.T) {
	result := typeDesc(ast.IntegerType{Size: 0}, false)
	if result != "untyped int literal" {
		t.Errorf("expected 'untyped integer literal', got %q", result)
	}
}

func TestTypeDescUnknown(t *testing.T) {
	result := typeDesc("unknown", false)
	if result != "unknown" {
		t.Errorf("expected 'unknown', got %q", result)
	}
}

func TestExecuteStmtBoolToIntegerError(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "x",
		IType: ast.IntegerType{Size: 32, Signed: true},
		Expr:  &ast.BoolLit{Value: true, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error assigning bool to int")
	}
}

func TestExecuteAssignmentBoolDefaultOp(t *testing.T) {
	i := New()
	i.env.Define("x", Value{BType: ast.BoolType{Nullable: false}, BData: true, IsBool: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%=",
		Expr: &ast.BoolLit{Value: false, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for unknown operator on bool")
	}
}

func TestExecuteAssignmentBoolTypeMismatch(t *testing.T) {
	i := New()
	i.env.Define("x", Value{BType: ast.BoolType{Nullable: false}, BData: true, IsBool: true})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "=",
		Expr: &ast.IntegerLit{Value: 42, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error assigning int to bool variable")
	}
}

func TestCheckIntFitsSigned16Overflow(t *testing.T) {
	err := checkIntFits(99999, ast.IntegerType{Size: 16, Signed: true})
	if err == nil {
		t.Errorf("expected overflow error for signed 16-bit")
	}
}

func TestCheckIntFitsSigned32Overflow(t *testing.T) {
	err := checkIntFits(9999999999, ast.IntegerType{Size: 32, Signed: true})
	if err == nil {
		t.Errorf("expected overflow error for signed 32-bit")
	}
}

func TestCheckIntFitsUnsigned64Overflow(t *testing.T) {
	err := checkIntFits(-1, ast.IntegerType{Size: 64, Signed: false})
	if err == nil {
		t.Errorf("expected overflow error for unsigned 64-bit")
	}
}

func TestCheckFloatFitsInf(t *testing.T) {
	err := checkFloatFits(math.Inf(1), ast.FloatType{Size: 16})
	if err != nil {
		t.Errorf("expected no error for Inf, got %v", err)
	}
}

func TestCheckFloatFitsNaN(t *testing.T) {
	err := checkFloatFits(math.NaN(), ast.FloatType{Size: 16})
	if err != nil {
		t.Errorf("expected no error for NaN, got %v", err)
	}
}

func TestCheckFloatFits32Overflow(t *testing.T) {
	err := checkFloatFits(3.5e40, ast.FloatType{Size: 32})
	if err == nil {
		t.Errorf("expected overflow error for 32-bit float")
	}
}

func TestExecuteAssignmentFloatOverflow(t *testing.T) {
	input := `
var a: float{size: 16} = 1.0;
a = 70000.0;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected float overflow error on assignment")
	}
}

func TestPrintTypeRefInt(t *testing.T) {
	input := "print((int).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "nullable 64-bit signed int" {
		t.Errorf("expected 'nullable 64-bit signed int', got %q", output)
	}
}

func TestPrintTypeRefFloat(t *testing.T) {
	input := "print((float).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "nullable 64-bit float" {
		t.Errorf("expected 'nullable 64-bit float', got %q", output)
	}
}

func TestPrintTypeRefBool(t *testing.T) {
	input := "print((bool).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "nullable bool" {
		t.Errorf("expected 'nullable bool', got %q", output)
	}
}

func TestIntMinMaxDefault(t *testing.T) {
	input := `
print((int.min).toString());
print((int.max).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(output, "\n")
	if len(lines) >= 1 && lines[0] != fmt.Sprintf("%d", math.MinInt64) {
		t.Errorf("expected %d for int.min, got %q", math.MinInt64, lines[0])
	}
	if len(lines) >= 2 && lines[1] != fmt.Sprintf("%d", math.MaxInt64) {
		t.Errorf("expected %d for int.max, got %q", math.MaxInt64, lines[1])
	}
}

func TestInt8MinMax(t *testing.T) {
	input := `
print((int{size: 8}.min).toString());
print((int{size: 8}.max).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(output, "\n")
	if lines[0] != "-128" {
		t.Errorf("expected -128 for int8.min, got %q", lines[0])
	}
	if lines[1] != "127" {
		t.Errorf("expected 127 for int8.max, got %q", lines[1])
	}
}

func TestUint8MinMax(t *testing.T) {
	input := `
print((int{size: 8, signed: false}.min).toString());
print((int{size: 8, signed: false}.max).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(output, "\n")
	if lines[0] != "0" {
		t.Errorf("expected 0 for uint8.min, got %q", lines[0])
	}
	if lines[1] != "255" {
		t.Errorf("expected 255 for uint8.max, got %q", lines[1])
	}
}

func TestFloatMinMax(t *testing.T) {
	input := `
print((float.min).toString());
print((float.max).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(output, "\n")
	if lines[0] != formatFloat(-math.MaxFloat64) {
		t.Errorf("expected %s for float.min, got %q", formatFloat(-math.MaxFloat64), lines[0])
	}
	if lines[1] != formatFloat(math.MaxFloat64) {
		t.Errorf("expected %s for float.max, got %q", formatFloat(math.MaxFloat64), lines[1])
	}
}

func TestFloatMinSubnormalMinNormal(t *testing.T) {
	input := `
print((float.min_subnormal).toString());
print((float.min_normal).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(output, "\n")
	if lines[0] != formatFloat(math.SmallestNonzeroFloat64) {
		t.Errorf("expected %s for float.min_subnormal, got %q", formatFloat(math.SmallestNonzeroFloat64), lines[0])
	}
	if lines[1] != formatFloat(math.Exp2(-1022)) {
		t.Errorf("expected %s for float.min_normal, got %q", formatFloat(math.Exp2(-1022)), lines[1])
	}
}

func TestFloatPrecisionExponentSize(t *testing.T) {
	input := `
print((float.precision).toString());
print((float.min_exponent).toString());
print((float.max_exponent).toString());
print((float.size).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(output, "\n")
	if lines[0] != "15" {
		t.Errorf("expected 15 for float.precision, got %q", lines[0])
	}
	if lines[1] != "-1022" {
		t.Errorf("expected -1022 for float.min_exponent, got %q", lines[1])
	}
	if lines[2] != "1023" {
		t.Errorf("expected 1023 for float.max_exponent, got %q", lines[2])
	}
	if lines[3] != "64" {
		t.Errorf("expected 64 for float.size, got %q", lines[3])
	}
}

func TestFloat32Properties(t *testing.T) {
	input := `
print((float{size: 32}.min).toString());
print((float{size: 32}.max).toString());
print((float{size: 32}.min_subnormal).toString());
print((float{size: 32}.min_normal).toString());
print((float{size: 32}.precision).toString());
print((float{size: 32}.min_exponent).toString());
print((float{size: 32}.max_exponent).toString());
print((float{size: 32}.size).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(output, "\n")
	if lines[0] != formatFloat(-math.MaxFloat32) {
		t.Errorf("expected %s for float32.min, got %q", formatFloat(-math.MaxFloat32), lines[0])
	}
	if lines[1] != formatFloat(math.MaxFloat32) {
		t.Errorf("expected %s for float32.max, got %q", formatFloat(math.MaxFloat32), lines[1])
	}
	if lines[2] != formatFloat(math.SmallestNonzeroFloat32) {
		t.Errorf("expected %s for float32.min_subnormal, got %q", formatFloat(math.SmallestNonzeroFloat32), lines[2])
	}
	if lines[3] != formatFloat(math.Exp2(-126)) {
		t.Errorf("expected %s for float32.min_normal, got %q", formatFloat(math.Exp2(-126)), lines[3])
	}
	if lines[4] != "7" {
		t.Errorf("expected 7 for float32.precision, got %q", lines[4])
	}
	if lines[5] != "-126" {
		t.Errorf("expected -126 for float32.min_exponent, got %q", lines[5])
	}
	if lines[6] != "127" {
		t.Errorf("expected 127 for float32.max_exponent, got %q", lines[6])
	}
	if lines[7] != "32" {
		t.Errorf("expected 32 for float32.size, got %q", lines[7])
	}
}

func TestTypeOf(t *testing.T) {
	input := `
var a: int{size: 8} = 42;
print((typeof(a)).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "nullable 8-bit signed int" {
		t.Errorf("expected 'nullable 8-bit signed int', got %q", output)
	}
}

func TestTypeOfDotMin(t *testing.T) {
	input := `
var a: int{size: 8} = 42;
print((typeof(a).min).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "-128" {
		t.Errorf("expected '-128', got %q", output)
	}
}

func TestTypeOfFloat(t *testing.T) {
	input := `
var a: float{size: 32} = 3.14;
print((typeof(a)).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "nullable 32-bit float" {
		t.Errorf("expected 'nullable 32-bit float', got %q", output)
	}
}

func TestTypeOfFloatDotMax(t *testing.T) {
	input := `
var a: float{size: 32} = 3.14;
print((typeof(a).max).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != formatFloat(math.MaxFloat32) {
		t.Errorf("expected %s, got %q", formatFloat(math.MaxFloat32), output)
	}
}

func TestNegIntMin(t *testing.T) {
	input := "print((-int{size: 8}.min).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "128" {
		t.Errorf("expected '128', got %q", output)
	}
}

func TestInt16MinMax(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 16, Signed: true}},
		Member: "min",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != -32768 {
		t.Errorf("expected -32768, got %d", val.Data)
	}

	expr2 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 16, Signed: true}},
		Member: "max",
	}
	val2, err2 := i.evalExpr(expr2)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if val2.Data != 32767 {
		t.Errorf("expected 32767, got %d", val2.Data)
	}
}

func TestInt32MinMax(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 32, Signed: true}},
		Member: "min",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != -2147483648 {
		t.Errorf("expected -2147483648, got %d", val.Data)
	}

	expr2 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 32, Signed: true}},
		Member: "max",
	}
	val2, err2 := i.evalExpr(expr2)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if val2.Data != 2147483647 {
		t.Errorf("expected 2147483647, got %d", val2.Data)
	}
}

func TestUint16MinMax(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 16, Signed: false}},
		Member: "min",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 0 {
		t.Errorf("expected 0, got %d", val.Data)
	}

	expr2 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 16, Signed: false}},
		Member: "max",
	}
	val2, err2 := i.evalExpr(expr2)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if val2.Data != 65535 {
		t.Errorf("expected 65535, got %d", val2.Data)
	}
}

func TestUint32MinMax(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 32, Signed: false}},
		Member: "max",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 4294967295 {
		t.Errorf("expected 4294967295, got %d", val.Data)
	}
}

func TestUint64MinMax(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 64, Signed: false}},
		Member: "min",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 0 {
		t.Errorf("expected 0, got %d", val.Data)
	}
}

func TestFloat16Properties(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "max",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FData != 65504 {
		t.Errorf("expected 65504, got %v", val.FData)
	}

	expr2 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "min_exponent",
	}
	val2, err2 := i.evalExpr(expr2)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if val2.Data != -14 {
		t.Errorf("expected -14, got %d", val2.Data)
	}

	expr3 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "max_exponent",
	}
	val3, err3 := i.evalExpr(expr3)
	if err3 != nil {
		t.Fatalf("unexpected error: %v", err3)
	}
	if val3.Data != 15 {
		t.Errorf("expected 15, got %d", val3.Data)
	}

	expr4 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "size",
	}
	val4, err4 := i.evalExpr(expr4)
	if err4 != nil {
		t.Fatalf("unexpected error: %v", err4)
	}
	if val4.Data != 16 {
		t.Errorf("expected 16, got %d", val4.Data)
	}

	expr5 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "min_subnormal",
	}
	val5, err5 := i.evalExpr(expr5)
	if err5 != nil {
		t.Fatalf("unexpected error: %v", err5)
	}
	if val5.FData != math.Exp2(-24) {
		t.Errorf("expected %v, got %v", math.Exp2(-24), val5.FData)
	}

	expr6 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "min_normal",
	}
	val6, err6 := i.evalExpr(expr6)
	if err6 != nil {
		t.Fatalf("unexpected error: %v", err6)
	}
	if val6.FData != math.Exp2(-14) {
		t.Errorf("expected %v, got %v", math.Exp2(-14), val6.FData)
	}

	expr7 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "precision",
	}
	val7, err7 := i.evalExpr(expr7)
	if err7 != nil {
		t.Fatalf("unexpected error: %v", err7)
	}
	if val7.Data != 3 {
		t.Errorf("expected 3, got %d", val7.Data)
	}

	expr8 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "min_exponent",
	}
	val8, err8 := i.evalExpr(expr8)
	if err8 != nil {
		t.Fatalf("unexpected error: %v", err8)
	}
	if val8.Data != -14 {
		t.Errorf("expected -14, got %d", val8.Data)
	}

	expr9 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "max_exponent",
	}
	val9, err9 := i.evalExpr(expr9)
	if err9 != nil {
		t.Fatalf("unexpected error: %v", err9)
	}
	if val9.Data != 15 {
		t.Errorf("expected 15, got %d", val9.Data)
	}

	expr10 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 16}},
		Member: "size",
	}
	val10, err10 := i.evalExpr(expr10)
	if err10 != nil {
		t.Fatalf("unexpected error: %v", err10)
	}
	if val10.Data != 16 {
		t.Errorf("expected 16, got %d", val10.Data)
	}
}

func TestFloat32MinSubnormal(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 32}},
		Member: "min_subnormal",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FData != float64(math.SmallestNonzeroFloat32) {
		t.Errorf("expected %v, got %v", math.SmallestNonzeroFloat32, val.FData)
	}
}

func TestFloat32MinNormal(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 32}},
		Member: "min_normal",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.FData != math.Exp2(-126) {
		t.Errorf("expected %v, got %v", math.Exp2(-126), val.FData)
	}
}

func TestIntSize(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 8, Signed: true}},
		Member: "size",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 8 {
		t.Errorf("expected 8, got %d", val.Data)
	}

	expr2 := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 32, Signed: false}},
		Member: "size",
	}
	val2, err2 := i.evalExpr(expr2)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if val2.Data != 32 {
		t.Errorf("expected 32, got %d", val2.Data)
	}
}

func TestBoolTypeRef(t *testing.T) {
	i := New()
	expr := &ast.TypeRef{Type: ast.BoolType{Nullable: false}}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsType || !val.IsBool {
		t.Errorf("expected IsType and IsBool")
	}
}

func TestTypeOfOnBool(t *testing.T) {
	input := `
var a: bool = true;
print((typeof(a)).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "nullable bool" {
		t.Errorf("expected 'nullable bool', got %q", output)
	}
}

func TestBoolSize(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.BoolType{Nullable: false}},
		Member: "size",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 8 {
		t.Errorf("expected 8, got %d", val.Data)
	}
}

func TestPrintBoolSize(t *testing.T) {
	input := "print((bool.size).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "8" {
		t.Errorf("expected 8, got %q", output)
	}
}

func TestBoolTypePrint(t *testing.T) {
	input := "print((bool).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "nullable bool" {
		t.Errorf("expected 'nullable bool', got %q", output)
	}
}

func TestTypeOfOnTypedVar(t *testing.T) {
	input := `
var a: int{size: 32} = 100;
print((typeof(a).min).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "-2147483648" {
		t.Errorf("expected '-2147483648', got %q", output)
	}
}

func TestFloatPrecisionDirect(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 64}},
		Member: "precision",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 15 {
		t.Errorf("expected 15, got %d", val.Data)
	}
}

func TestFloatMinExponentDirect(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 64}},
		Member: "min_exponent",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != -1022 {
		t.Errorf("expected -1022, got %d", val.Data)
	}
}

func TestFloatMaxExponentDirect(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 64}},
		Member: "max_exponent",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 1023 {
		t.Errorf("expected 1023, got %d", val.Data)
	}
}

func TestFloatSizeDirect(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 64}},
		Member: "size",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 64 {
		t.Errorf("expected 64, got %d", val.Data)
	}
}

func TestFloatMemberNotFound(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.FloatType{Size: 64}},
		Member: "bogus",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for nonexistent float member")
	}
}

func TestUint64Max(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 64, Signed: false}},
		Member: "max",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uint64(val.Data) != math.MaxUint64 {
		t.Errorf("expected MaxUint64, got %d", uint64(val.Data))
	}
}

func TestPrintUint64Max(t *testing.T) {
	input := "print((int{size: 64, signed: false}.max).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "18446744073709551615" {
		t.Errorf("expected '18446744073709551615', got %q", output)
	}
}

func TestPrintUint64Min(t *testing.T) {
	input := "print((int{size: 64, signed: false}.min).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "0" {
		t.Errorf("expected '0', got %q", output)
	}
}

func TestValueStringTypeDesc(t *testing.T) {
	v := Value{IsType: true, IType: ast.IntegerType{Size: 32, Signed: true, Nullable: false}}
	result := v.String()
	if result != "non-nullable 32-bit signed int" {
		t.Errorf("expected 'non-nullable 32-bit signed int', got %q", result)
	}

	v2 := Value{IsType: true, IsFloat: true, FType: ast.FloatType{Size: 64, Nullable: false}}
	result2 := v2.String()
	if result2 != "non-nullable 64-bit float" {
		t.Errorf("expected 'non-nullable 64-bit float', got %q", result2)
	}
}

func TestMemberAccessUndefinedVar(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.VarRef{Name: "nonexistent"},
		Member: "foo",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for member access on undefined var")
	}
}

func TestBoolTypeMemberError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.BoolType{Nullable: false}},
		Member: "nonexistent",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for nonexistent bool member")
	}
}

func TestTypeMemberError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.IntegerType{Size: 64, Signed: true}},
		Member: "nonexistent",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for nonexistent type member")
	}
}

func TestValueMemberError(t *testing.T) {
	i := New()
	i.env.Define("x", Value{Untyped: true, Data: 42})
	expr := &ast.MemberAccess{
		Object: &ast.VarRef{Name: "x"},
		Member: "foo",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for nonexistent value member")
	}
}

func TestEqInt(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "==",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsBool {
		t.Errorf("expected IsBool")
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestEqIntFalse(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "==",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != false {
		t.Errorf("expected false, got %v", val.BData)
	}
}

func TestEqFloat(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.5, Untyped: true},
		Op:    "==",
		Right: &ast.FloatLit{Value: 1.5, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestEqBool(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.BoolLit{Value: true, Untyped: true},
		Op:    "==",
		Right: &ast.BoolLit{Value: true, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestNotEq(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "!=",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestNotEqFalse(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "!=",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != false {
		t.Errorf("expected false, got %v", val.BData)
	}
}

func TestLtInt(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "<",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestLtFloat(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.0, Untyped: true},
		Op:    "<",
		Right: &ast.FloatLit{Value: 2.0, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestGtInt(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 2, Untyped: true},
		Op:    ">",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestLteInt(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "<=",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestLteIntEqual(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "<=",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestGteInt(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 2, Untyped: true},
		Op:    ">=",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestGteIntEqual(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    ">=",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestAnd(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.BoolLit{Value: true, Untyped: true},
		Op:    "&&",
		Right: &ast.BoolLit{Value: true, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsBool {
		t.Errorf("expected IsBool")
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestAndFalse(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.BoolLit{Value: true, Untyped: true},
		Op:    "&&",
		Right: &ast.BoolLit{Value: false, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != false {
		t.Errorf("expected false, got %v", val.BData)
	}
}

func TestOr(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.BoolLit{Value: true, Untyped: true},
		Op:    "||",
		Right: &ast.BoolLit{Value: false, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestOrFalse(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.BoolLit{Value: false, Untyped: true},
		Op:    "||",
		Right: &ast.BoolLit{Value: false, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != false {
		t.Errorf("expected false, got %v", val.BData)
	}
}

func TestNot(t *testing.T) {
	i := New()
	expr := &ast.UnaryExpr{
		Op:    "!",
		Right: &ast.BoolLit{Value: false, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsBool {
		t.Errorf("expected IsBool")
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestNotTrue(t *testing.T) {
	i := New()
	expr := &ast.UnaryExpr{
		Op:    "!",
		Right: &ast.BoolLit{Value: true, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != false {
		t.Errorf("expected false, got %v", val.BData)
	}
}

func TestNotOnNonBool(t *testing.T) {
	i := New()
	expr := &ast.UnaryExpr{
		Op:    "!",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for ! on non-bool")
	}
}

func TestAndOnNonBool(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "&&",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for && on non-bool")
	}
}

func TestOrOnNonBool(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "||",
		Right: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for || on non-bool")
	}
}

func TestEqNull(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 1, Untyped: true},
		Op:    "==",
		Right: &ast.NullLit{},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Null {
		t.Errorf("expected non-null result")
	}
	if val.BData != false {
		t.Errorf("expected false, got %v", val.BData)
	}
}

func TestNullNot(t *testing.T) {
	i := New()
	expr := &ast.UnaryExpr{
		Op:    "!",
		Right: &ast.NullLit{},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Null {
		t.Errorf("expected non-null result")
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestPrintEq(t *testing.T) {
	input := "print((1 == 1).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestPrintGt(t *testing.T) {
	input := "print((2 > 1).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestPrintLtFloat(t *testing.T) {
	input := "print((1.5 < 2.5).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestPrintAnd(t *testing.T) {
	input := "print((true && false).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "false" {
		t.Errorf("expected 'false', got %q", output)
	}
}

func TestPrintOr(t *testing.T) {
	input := "print((true || false).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestPrintNot(t *testing.T) {
	input := "print((!true).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "false" {
		t.Errorf("expected 'false', got %q", output)
	}
}

func TestPrintNotNot(t *testing.T) {
	input := "print((!!true).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestPrintPrecedenceOrAnd(t *testing.T) {
	input := "print((true || false && false).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// true || (false && false) = true || false = true
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestPrintPrecedenceCmpAdd(t *testing.T) {
	input := "print((1 + 2 < 3).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// (1 + 2) < 3 = 3 < 3 = false
	if output != "false" {
		t.Errorf("expected 'false', got %q", output)
	}
}

func TestPrintNotEq(t *testing.T) {
	input := "print((1 != 2).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestUnknownUnaryOp(t *testing.T) {
	i := New()
	expr := &ast.UnaryExpr{
		Op:    "~",
		Right: &ast.IntegerLit{Value: 1, Untyped: true},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for unknown unary operator")
	}
}

func TestUnaryUndefinedVar(t *testing.T) {
	i := New()
	expr := &ast.UnaryExpr{
		Op:    "!",
		Right: &ast.VarRef{Name: "nonexistent"},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for undefined var in unary")
	}
}

func TestGtFloat(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 2.0, Untyped: true},
		Op:    ">",
		Right: &ast.FloatLit{Value: 1.0, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestLteFloat(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.0, Untyped: true},
		Op:    "<=",
		Right: &ast.FloatLit{Value: 2.0, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestGteFloat(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 2.0, Untyped: true},
		Op:    ">=",
		Right: &ast.FloatLit{Value: 1.0, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestLteFloatEqual(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.0, Untyped: true},
		Op:    "<=",
		Right: &ast.FloatLit{Value: 1.0, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestGteFloatEqual(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 1.0, Untyped: true},
		Op:    ">=",
		Right: &ast.FloatLit{Value: 1.0, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestTypedInt64EqFloat64Error(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IType: ast.IntegerType{Size: 64, Signed: true}, Data: 5})
	i.env.Define("b", Value{FType: ast.FloatType{Size: 64}, FData: 5.0, IsFloat: true})
	expr := &ast.BinaryExpr{
		Left:  &ast.VarRef{Name: "a"},
		Op:    "==",
		Right: &ast.VarRef{Name: "b"},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for int64 == float64")
	}
}

func TestTypedInt32EqFloat64(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 5})
	i.env.Define("b", Value{FType: ast.FloatType{Size: 64}, FData: 5.0, IsFloat: true})
	expr := &ast.BinaryExpr{
		Left:  &ast.VarRef{Name: "a"},
		Op:    "==",
		Right: &ast.VarRef{Name: "b"},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestUntypedIntEqFloat(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 5, Untyped: true},
		Op:    "==",
		Right: &ast.FloatLit{Value: 5.0, Untyped: true},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.BData != true {
		t.Errorf("expected true, got %v", val.BData)
	}
}

func TestTypedInt64LtFloat64Error(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IType: ast.IntegerType{Size: 64, Signed: true}, Data: 5})
	i.env.Define("b", Value{FType: ast.FloatType{Size: 64}, FData: 10.0, IsFloat: true})
	expr := &ast.BinaryExpr{
		Left:  &ast.VarRef{Name: "a"},
		Op:    "<",
		Right: &ast.VarRef{Name: "b"},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for int64 < float64")
	}
}

func TestBoolEqIntError(t *testing.T) {
	i := New()
	i.env.Define("a", Value{Untyped: true, BData: true, IsBool: true})
	i.env.Define("b", Value{Untyped: true, Data: 1})
	expr := &ast.BinaryExpr{
		Left:  &ast.VarRef{Name: "a"},
		Op:    "==",
		Right: &ast.VarRef{Name: "b"},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for bool == int")
	}
}

func TestPrintTypedInt64EqFloat64Error(t *testing.T) {
	input := `
var a: int = 5;
var b: float = 5;
print((a == b).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for int64 == float64 comparison")
	}
}

func TestNullComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print((null == null).toString());", "true"},
		{"print((null != null).toString());", "false"},
		{"print((null == 1).toString());", "false"},
		{"print((null != 1).toString());", "true"},
		{"print((1 == null).toString());", "false"},
		{"print((1 != null).toString());", "true"},
		{"print((null == 1.0).toString());", "false"},
		{"print((null != 1.0).toString());", "true"},
		{"print((null < 1).toString());", "false"},
		{"print((null > 1).toString());", "false"},
		{"print((null <= 1).toString());", "false"},
		{"print((null >= 1).toString());", "false"},
		{"print((1 < null).toString());", "false"},
		{"print((1 > null).toString());", "false"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", tt.input, p.Errors())
		}
		i := New()
		output := captureOutput(func() {
			err := i.Run(program)
			if err != nil {
				t.Fatalf("input %q: unexpected runtime error: %v", tt.input, err)
			}
		})
		if output != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestNullInBooleanOps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print((null && true).toString());", "false"},
		{"print((null && false).toString());", "false"},
		{"print((true && null).toString());", "false"},
		{"print((false && null).toString());", "false"},
		{"print((null || true).toString());", "true"},
		{"print((null || false).toString());", "false"},
		{"print((true || null).toString());", "true"},
		{"print((false || null).toString());", "false"},
		{"print((!null).toString());", "true"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", tt.input, p.Errors())
		}
		i := New()
		output := captureOutput(func() {
			err := i.Run(program)
			if err != nil {
				t.Fatalf("input %q: unexpected runtime error: %v", tt.input, err)
			}
		})
		if output != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestNaNComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print((NaN == NaN).toString());", "false"},
		{"print((NaN != NaN).toString());", "true"},
		{"print((NaN < 0).toString());", "false"},
		{"print((NaN > 0).toString());", "false"},
		{"print((NaN <= 0).toString());", "false"},
		{"print((NaN >= 0).toString());", "false"},
		{"print((0 < NaN).toString());", "false"},
		{"print((0 > NaN).toString());", "false"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", tt.input, p.Errors())
		}
		i := New()
		output := captureOutput(func() {
			err := i.Run(program)
			if err != nil {
				t.Fatalf("input %q: unexpected runtime error: %v", tt.input, err)
			}
		})
		if output != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestInfinityComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print((infinity == infinity).toString());", "true"},
		{"print((-infinity < 0).toString());", "true"},
		{"print((-infinity > 0).toString());", "false"},
		{"print((infinity > 1000).toString());", "true"},
		{"print((-infinity < -1000).toString());", "true"},
		{"print((infinity == 1.0).toString());", "false"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", tt.input, p.Errors())
		}
		i := New()
		output := captureOutput(func() {
			err := i.Run(program)
			if err != nil {
				t.Fatalf("input %q: unexpected runtime error: %v", tt.input, err)
			}
		})
		if output != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestFloatDivByZeroInline(t *testing.T) {
	input := `
print((1.0 / 0.0).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected runtime error: %v", err)
		}
	})
	if !strings.Contains(output, "infinity") {
		t.Errorf("expected output to contain 'infinity', got %q", output)
	}
}

func TestDoubleNegation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print((!!true).toString());", "true"},
		{"print((!!false).toString());", "false"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", tt.input, p.Errors())
		}
		i := New()
		output := captureOutput(func() {
			err := i.Run(program)
			if err != nil {
				t.Fatalf("input %q: unexpected runtime error: %v", tt.input, err)
			}
		})
		if output != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestNoImplicitIntToBoolConversion(t *testing.T) {
	tests := []string{
		"var x: bool = 0;",
		"var x: bool = 1;",
		"var x: bool = -1;",
		"var x: bool = 42;",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", input, p.Errors())
		}
		i := New()
		err := i.Run(program)
		if err == nil {
			t.Errorf("input %q: expected error for int to bool conversion, got none", input)
		}
	}
}

func TestBoolComparisonOperatorsError(t *testing.T) {
	tests := []string{
		"print((true < false).toString());",
		"print((true > false).toString());",
		"print((true <= false).toString());",
		"print((true >= false).toString());",
		"print((1 < true).toString());",
		"print((true < 1).toString());",
		"print((1.0 < true).toString());",
		"print((true < 1.0).toString());",
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", input, p.Errors())
		}
		i := New()
		err := i.Run(program)
		if err == nil {
			t.Errorf("input %q: expected error for boolean comparison operator, got none", input)
		}
	}
}

func TestShortCircuitRightOperandNonBool(t *testing.T) {
	tests := []struct {
		left  ast.Expr
		op    string
		right ast.Expr
	}{
		{&ast.BoolLit{Value: true, Untyped: true}, "&&", &ast.IntegerLit{Value: 1, Untyped: true}},
		{&ast.BoolLit{Value: false, Untyped: true}, "||", &ast.IntegerLit{Value: 1, Untyped: true}},
	}

	for _, tt := range tests {
		i := New()
		expr := &ast.BinaryExpr{
			Left:  tt.left,
			Op:    tt.op,
			Right: tt.right,
		}
		_, err := i.evalExpr(expr)
		if err == nil {
			t.Errorf("expected error for non-bool right operand with op %q, got none", tt.op)
		}
	}
}

func TestShortCircuitOperandEvalError(t *testing.T) {
	tests := []struct {
		left  ast.Expr
		op    string
		right ast.Expr
	}{
		{&ast.VarRef{Name: "undefined"}, "&&", &ast.BoolLit{Value: true, Untyped: true}},
		{&ast.BoolLit{Value: true, Untyped: true}, "&&", &ast.VarRef{Name: "undefined"}},
		{&ast.VarRef{Name: "undefined"}, "||", &ast.BoolLit{Value: false, Untyped: true}},
		{&ast.BoolLit{Value: false, Untyped: true}, "||", &ast.VarRef{Name: "undefined"}},
	}

	for _, tt := range tests {
		i := New()
		expr := &ast.BinaryExpr{
			Left:  tt.left,
			Op:    tt.op,
			Right: tt.right,
		}
		_, err := i.evalExpr(expr)
		if err == nil {
			t.Errorf("expected error for undefined variable with op %q, got none", tt.op)
		}
	}
}

func TestShortCircuitEvaluation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print((false && true).toString());", "false"},
		{"print((true || false).toString());", "true"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			t.Fatalf("input %q: unexpected parse errors: %v", tt.input, p.Errors())
		}
		i := New()
		output := captureOutput(func() {
			err := i.Run(program)
			if err != nil {
				t.Fatalf("input %q: unexpected runtime error: %v", tt.input, err)
			}
		})
		if output != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestIfTrue(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
if (true) {
	x = 1;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestIfFalse(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
if (false) {
	x = 1;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0" {
		t.Errorf("expected '0', got %q", output)
	}
}

func TestIfElseTrue(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
if (true) {
	x = 1;
} else {
	x = 2;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestIfElseFalse(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
if (false) {
	x = 1;
} else {
	x = 2;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestIfElseIfElse(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
if (x == 0) {
	print((0).toString());
} else if (x == 1) {
	print((1).toString());
} else {
	print((2).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestIfConditionNotBool(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
if (x) {
	print((1).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for non-bool condition")
	}
}

func TestNestedIf(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
var y: int{size: 32, signed: true, nullable: false} = 2;
if (true) {
	if (true) {
		x = 3;
	}
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "3" {
		t.Errorf("expected '3', got %q", output)
	}
}

func TestIfVariableScopedToBlock(t *testing.T) {
	input := `
var a: int{size: 32, signed: true, nullable: true} = null;
if (a == null) {
	var b: int{size: 32, signed: true, nullable: false} = 1;
}
print((b).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: b is not defined outside if block")
	}
}

func TestIfVariableVisibleInInnerScope(t *testing.T) {
	input := `
var a: int{size: 32, signed: true, nullable: false} = 1;
if (true) {
	print((a).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestIfModifiesOuterVar(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
if (true) {
	x = 1;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestElseVariableScoped(t *testing.T) {
	input := `
var a: int{size: 32, signed: true, nullable: false} = 1;
if (a == 0) {
	var b: int{size: 32, signed: true, nullable: false} = 2;
} else {
	var c: int{size: 32, signed: true, nullable: false} = 3;
}
print((b).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: b is not defined outside if/else block")
	}
}

func TestElseIfVariableScoped(t *testing.T) {
	input := `
var a: int{size: 32, signed: true, nullable: false} = 2;
if (a == 0) {
	var b: int{size: 32, signed: true, nullable: false} = 1;
} else if (a == 2) {
	var c: int{size: 32, signed: true, nullable: false} = 3;
}
print((c).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: c is not defined outside else if block")
	}
}

func TestIfNullConditionWithElse(t *testing.T) {
	input := `
var a: int{size: 32, signed: true, nullable: true} = null;
if (a == null) {
	print((1).toString());
} else {
	print((2).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestIfNullLiteralConditionWithElse(t *testing.T) {
	input := `
if (null) {
	print((1).toString());
} else {
	print((2).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestIfNullLiteralConditionNoElse(t *testing.T) {
	input := `
if (null) {
	print((1).toString());
}
print((2).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestIfConditionExprError(t *testing.T) {
	input := `
if (undefined) {
	print((1).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable in condition")
	}
}

func TestBlockStmtErrorPropagation(t *testing.T) {
	input := `
if (true) {
	print((undefinedVar).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable inside block")
	}
}

func TestExecElseUnexpectedType(t *testing.T) {
	i := New()
	i.env.Define("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 1})
	stmt := &ast.IfStmt{
		Condition: &ast.BoolLit{Value: false, Untyped: true},
		Then:      &ast.BlockStmt{},
		Else:      &ast.PrintStmt{Expr: &ast.IntegerLit{Value: 1, Untyped: true}},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error for unexpected else type")
	}
}

func TestVariableShadowingInIfBlock(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
if (true) {
	var x: int{size: 32, signed: true, nullable: false} = 2;
	print((x).toString());
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2\n1" {
		t.Errorf("expected '2\\n1', got %q", output)
	}
}

func TestEmptyThenBlock(t *testing.T) {
	input := `
if (true) { }
print((1).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestEmptyThenBlockWithElse(t *testing.T) {
	input := `
if (false) { } else {
	print((1).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestEmptyElseBlock(t *testing.T) {
	input := `
if (true) {
	print((1).toString());
} else { }
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestMultipleElseIfChain(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 2;
if (x == 0) {
	print((0).toString());
} else if (x == 1) {
	print((1).toString());
} else if (x == 2) {
	print((2).toString());
} else if (x == 3) {
	print((3).toString());
} else {
	print((4).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestAllElseIfConditionsFalseNoElse(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 10;
if (x == 0) {
	print((0).toString());
} else if (x == 1) {
	print((1).toString());
} else if (x == 2) {
	print((2).toString());
}
print((3).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "3" {
		t.Errorf("expected '3', got %q", output)
	}
}

func TestRedeclareVarAfterIfBlock(t *testing.T) {
	input := `
var a: int{size: 32, signed: true, nullable: false} = 1;
if (true) {
	var b: int{size: 32, signed: true, nullable: false} = 2;
}
var b: int{size: 32, signed: true, nullable: false} = 3;
print((b).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "3" {
		t.Errorf("expected '3', got %q", output)
	}
}

func TestNullLiteralConditionWithElseIf(t *testing.T) {
	input := `
if (null) {
	print((1).toString());
} else if (true) {
	print((2).toString());
} else {
	print((3).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestComplexExpressionCondition(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
var y: bool{nullable: false} = true;
if (x + 1 == 2 && y) {
	print((1).toString());
} else {
	print((2).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestMultipleStatementsInThenBlock(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
if (true) {
	x = 1;
	x = x + 1;
	x = x * 3;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "6" {
		t.Errorf("expected '6', got %q", output)
	}
}

func TestElseIfConditionCanAccessOuterVarModifiedBeforehand(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
x = 5;
if (x == 0) {
	print((0).toString());
} else if (x == 5) {
	print((5).toString());
} else {
	print((99).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "5" {
		t.Errorf("expected '5', got %q", output)
	}
}

func TestIfElseIfFirstConditionTrueSkipsRest(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
if (x == 0) {
	print((0).toString());
} else if (x == 0) {
	print((99).toString());
} else if (x == 0) {
	print((99).toString());
} else {
	print((99).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0" {
		t.Errorf("expected '0', got %q", output)
	}
}

func TestIfWithBoolVarCondition(t *testing.T) {
	input := `
var x: bool{nullable: false} = true;
if (x) {
	print((1).toString());
} else {
	print((2).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestElseIfModifiesOuterVarBeforeNextCondition(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 1;
if (x == 0) {
	print((0).toString());
} else if (x == 1) {
	x = 2;
} else if (x == 2) {
	print((2).toString());
} else {
	print((3).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}

func TestForLoopCount(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (i = 0; i < 5; i = i + 1) {
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2\n3\n4" {
		t.Errorf("expected '0\\n1\\n2\\n3\\n4', got %q", output)
	}
}

func TestForLoopWithVarDeclInit(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 3; i = i + 1) {
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestForLoopVarScopedToLoop(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 1; i = i + 1) {
	print((i).toString());
}
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: i is scoped to for loop")
	}
}

func TestForLoopEmptyInitCondUpdate(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (;;) {
	print((i).toString());
	i = i + 1;
	if (i >= 3) {
		break;
	}
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestWhileLoop(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (i < 3) {
	print((i).toString());
	i = i + 1;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestBreakInLoop(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (true) {
	print((i).toString());
	if (i >= 2) {
		break;
	}
	i = i + 1;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestSkipInLoop(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (i < 5) {
	i = i + 1;
	if (i == 3) {
		skip;
	}
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1\n2\n4\n5" {
		t.Errorf("expected '1\\n2\\n4\\n5', got %q", output)
	}
}

func TestSkipInForLoop(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 1; i <= 5; i = i + 1) {
	if (i == 3) {
		skip;
	}
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1\n2\n4\n5" {
		t.Errorf("expected '1\\n2\\n4\\n5', got %q", output)
	}
}

func TestBreakOutsideLoop(t *testing.T) {
	input := "break;"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for break outside loop")
	}
}

func TestSkipOutsideLoop(t *testing.T) {
	input := "skip;"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for skip outside loop")
	}
}

func TestNestedLoops(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 2; i = i + 1) {
	for (var j: int{size: 32, signed: true, nullable: false} = 0; j < 2; j = j + 1) {
		print((i * 10 + j).toString());
	}
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n10\n11" {
		t.Errorf("expected '0\\n1\\n10\\n11', got %q", output)
	}
}

func TestForLoopConditionNotBool(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
for (; x; ) { }
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for non-bool for condition")
	}
}

func TestWhileLoopConditionNotBool(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 0;
while (x) { }
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for non-bool while condition")
	}
}

func TestForLoopBreakFromInnerIf(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 10; i = i + 1) {
	if (i == 3) {
		break;
	}
}
print((1).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestForLoopInitError(t *testing.T) {
	input := `
for (var x: int{size: 8, signed: true, nullable: false} = 1000; false; ) { }
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for overflow in for init")
	}
}

func TestForLoopBodyScopedVar(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 1; i = i + 1) {
	var x: int{size: 32, signed: true, nullable: false} = 42;
	print((x).toString());
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: x is scoped to for body")
	}
}

func TestBreakWithSkipInSameLoop(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (i < 10) {
	i = i + 1;
	if (i == 3) {
		skip;
	}
	if (i == 5) {
		break;
	}
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1\n2\n4" {
		t.Errorf("expected '1\\n2\\n4', got %q", output)
	}
}

func TestBreakOutsideLoopError(t *testing.T) {
	input := "break;"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Fatalf("expected error for break outside loop")
	}
	if err.Error() != "break outside loop" {
		t.Errorf("expected 'break outside loop', got %q", err.Error())
	}
}

func TestSkipOutsideLoopError(t *testing.T) {
	input := "skip;"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Fatalf("expected error for skip outside loop")
	}
	if err.Error() != "skip outside loop" {
		t.Errorf("expected 'skip outside loop', got %q", err.Error())
	}
}

func TestForLoopConditionError(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (; undefinedVar; ) { }
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined condition variable")
	}
}

func TestForLoopBodyError(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (; i < 1; i = i + 1) {
	var x: int{size: 32, signed: true, nullable: false} = y;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable in body")
	}
}

func TestForLoopUpdateError(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (; i < 1; i = undefinedVar) { }
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable in update")
	}
}

func TestWhileLoopConditionError(t *testing.T) {
	input := "while (undefinedVar) { }"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined condition variable")
	}
}

func TestWhileLoopBodyError(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (i < 1) {
	var x: int{size: 32, signed: true, nullable: false} = y;
	i = i + 1;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable in body")
	}
}

func TestLoopSignalUnknown(t *testing.T) {
	var s LoopSignal = 99
	msg := s.Error()
	if msg != "unknown loop control" {
		t.Errorf("expected 'unknown loop control', got %q", msg)
	}
}

func TestWhileBreakInsideIf(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (i < 10) {
	if (i == 3) {
		break;
	}
	i = i + 1;
}
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "3" {
		t.Errorf("expected '3', got %q", output)
	}
}

func TestWhileFalseCondition(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 5;
while (false) {
	x = 10;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "5" {
		t.Errorf("expected '5', got %q", output)
	}
}

func TestForLoopFalseCondition(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 42; false; i = i + 1) {
	print((i).toString());
}
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: i is scoped to for loop")
	}
}

func TestForLoopCompoundUpdate(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 3; i += 1) {
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestForLoopVarDeclWithoutInit(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false}; i < 3; i = i + 1) {
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestForLoopInitRefOuterVar(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 5;
for (var i: int{size: 32, signed: true, nullable: false} = x; i < x + 2; i = i + 1) {
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "5\n6" {
		t.Errorf("expected '5\\n6', got %q", output)
	}
}

func TestForLoopOnlyInitBreak(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (i = 0;;) {
	if (i >= 3) {
		break;
	}
	print((i).toString());
	i = i + 1;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestNestedLoopsBreakInner(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 3; i = i + 1) {
	for (var j: int{size: 32, signed: true, nullable: false} = 0; j < 3; j = j + 1) {
		if (j == 1) {
			break;
		}
		print((i * 10 + j).toString());
	}
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n10\n20" {
		t.Errorf("expected '0\\n10\\n20', got %q", output)
	}
}

func TestForLoopBodyVarNotVisibleInUpdate(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
for (; i < 1; i = x) {
	var x: int{size: 32, signed: true, nullable: false} = 42;
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: body-scoped var should not be visible in update")
	}
}

func TestWhileLoopBodyVarNotVisibleInCondition(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
while (i < 3) {
	var x: int{size: 32, signed: true, nullable: false} = 42;
	i = i + 1;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error: body-scoped var should not be visible after loop")
	}
}

func TestIncrementInt(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 5;
i++;
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "6" {
		t.Errorf("expected '6', got %q", output)
	}
}

func TestDecrementInt(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 5;
i--;
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "4" {
		t.Errorf("expected '4', got %q", output)
	}
}

func TestIncrementFloat(t *testing.T) {
	input := `
var f: float{size: 64, nullable: false} = 3.5;
f++;
print((f).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "4.5" {
		t.Errorf("expected '4.5', got %q", output)
	}
}

func TestDecrementFloat(t *testing.T) {
	input := `
var f: float{size: 64, nullable: false} = 3.5;
f--;
print((f).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2.5" {
		t.Errorf("expected '2.5', got %q", output)
	}
}

func TestIncrementInForLoop(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 0; i < 3; i++) {
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "0\n1\n2" {
		t.Errorf("expected '0\\n1\\n2', got %q", output)
	}
}

func TestDecrementInForLoop(t *testing.T) {
	input := `
for (var i: int{size: 32, signed: true, nullable: false} = 2; i >= 0; i--) {
	print((i).toString());
}
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2\n1\n0" {
		t.Errorf("expected '2\\n1\\n0', got %q", output)
	}
}

func TestIncrementUndefinedVar(t *testing.T) {
	input := "undefinedVar++;"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	err := interp.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable")
	}
}

func TestIncrementNullVar(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: true};
i++;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	err := interp.Run(program)
	if err == nil {
		t.Errorf("expected error for null increment")
	}
}

func TestIncrementBoolVar(t *testing.T) {
	input := `
var b: bool{nullable: false} = true;
b++;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	err := interp.Run(program)
	if err == nil {
		t.Errorf("expected error for bool increment")
	}
}

func TestIncrementOverflow(t *testing.T) {
	input := `
var i: int{size: 8, signed: true, nullable: false} = 127;
i++;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	err := interp.Run(program)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestIncrementFromBlock(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 5;
if (true) {
	i++;
}
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "6" {
		t.Errorf("expected '6', got %q", output)
	}
}

func TestDecrementFromBlock(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 5;
if (true) {
	i--;
}
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "4" {
		t.Errorf("expected '4', got %q", output)
	}
}

func TestIncrementZero(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
i++;
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestDecrementZero(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = 0;
i--;
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "-1" {
		t.Errorf("expected '-1', got %q", output)
	}
}

func TestIncrementNegative(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = -3;
i++;
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "-2" {
		t.Errorf("expected '-2', got %q", output)
	}
}

func TestDecrementNegative(t *testing.T) {
	input := `
var i: int{size: 32, signed: true, nullable: false} = -3;
i--;
print((i).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	output := captureOutput(func() {
		err := interp.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "-4" {
		t.Errorf("expected '-4', got %q", output)
	}
}

func TestDecrementOverflow(t *testing.T) {
	input := `
var i: int{size: 8, signed: true, nullable: false} = -128;
i--;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	interp := New()
	err := interp.Run(program)
	if err == nil {
		t.Errorf("expected overflow error")
	}
}

func TestArrayDeclAndPrint(t *testing.T) {
	input := `
var a: array{size: 3}<int> = [1, 2, 3];
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[1, 2, 3]" {
		t.Errorf("expected '[1, 2, 3]', got %q", output)
	}
}

func TestArrayDeclWithoutInit(t *testing.T) {
	input := `
var a: array{size: 3}<int>;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Array without init is zero-filled
	if output != "[0, 0, 0]" {
		t.Errorf("expected '[0, 0, 0]', got %q", output)
	}
}

func TestArraySizeMismatch(t *testing.T) {
	input := `
var a: array{size: 3}<int> = [1, 2];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected size mismatch error")
	}
}

func TestArrayIndexAccess(t *testing.T) {
	input := `
var a: array{size: 3}<int> = [10, 20, 30];
print((a[1]).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "20" {
		t.Errorf("expected '20', got %q", output)
	}
}

func TestArrayIndexOutOfBounds(t *testing.T) {
	input := `
var a: array{size: 3}<int> = [1, 2, 3];
print((a[5]).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected out of bounds error")
	}
}

func TestArrayIndexedAssign(t *testing.T) {
	input := `
var a: array{size: 3}<int> = [1, 2, 3];
a[1] = 99;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[1, 99, 3]" {
		t.Errorf("expected '[1, 99, 3]', got %q", output)
	}
}

func TestArrayIndexedAssignOutOfBounds(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2];
a[5] = 99;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected out of bounds error on assignment")
	}
}

func TestArrayLength(t *testing.T) {
	input := `
var a: array{size: 5}<int> = [1, 2, 3, 4, 5];
print((a.length).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "5" {
		t.Errorf("expected '5', got %q", output)
	}
}

func TestListDeclAndPrint(t *testing.T) {
	input := `
var a: list<int> = [1, 2, 3];
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[1, 2, 3]" {
		t.Errorf("expected '[1, 2, 3]', got %q", output)
	}
}

func TestListDeclEmpty(t *testing.T) {
	input := `
var a: list<int>;
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[]" {
		t.Errorf("expected '[]', got %q", output)
	}
}

func TestListAdd(t *testing.T) {
	input := `
var a: list<int>;
a.add(42);
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[42]" {
		t.Errorf("expected '[42]', got %q", output)
	}
}

func TestListAddAtIndex(t *testing.T) {
	input := `
var a: list<int> = [1, 3];
a.add(2, 1);
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[1, 2, 3]" {
		t.Errorf("expected '[1, 2, 3]', got %q", output)
	}
}

func TestListRemove(t *testing.T) {
	input := `
var a: list<int> = [1, 2, 3];
a.remove(1);
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[1, 3]" {
		t.Errorf("expected '[1, 3]', got %q", output)
	}
}

func TestListLength(t *testing.T) {
	input := `
var a: list<int> = [1, 2, 3, 4];
print((a.length).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "4" {
		t.Errorf("expected '4', got %q", output)
	}
}

func TestListWithMinMax(t *testing.T) {
	input := `
var a: list{min: 1, max: 5}<int> = [1, 2, 3];
print((a.length).toString());
a.add(4);
print((a.length).toString());
a.add(5);
print((a.length).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "3\n4\n5" {
		t.Errorf("expected '3\\n4\\n5', got %q", output)
	}
}

func TestListAddBeyondMax(t *testing.T) {
	input := `
var a: list{min: 1, max: 2}<int> = [1, 2];
a.add(3);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for exceeding max capacity")
	}
}

func TestListRemoveBelowMin(t *testing.T) {
	input := `
var a: list{min: 2, max: 5}<int> = [1, 2];
a.remove(0);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for going below min capacity")
	}
}

func TestListRemoveReturnsElement(t *testing.T) {
	input := `
var a: list<int> = [1, 2, 3];
var x: int{size: 64} = a.remove(1);
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestArrayLiteralExpr(t *testing.T) {
	input := `
print(([1, 2, 3]).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[1, 2, 3]" {
		t.Errorf("expected '[1, 2, 3]', got %q", output)
	}
}

func TestEmptyArrayLiteral(t *testing.T) {
	input := `
print(([]).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[]" {
		t.Errorf("expected '[]', got %q", output)
	}
}

func TestArrayTypeDesc(t *testing.T) {
	ft := ast.FloatType{Size: 32, Nullable: false}
	at := ast.ArrayType{ElemType: ft, Size: 5}
	result := typeDescForType(at)
	if result != "array{size: 5}<non-nullable 32-bit float>" {
		t.Errorf("expected 'array{size: 5}<non-nullable 32-bit float>', got %q", result)
	}
}

func TestListTypeDesc(t *testing.T) {
	result := typeDescForType(ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true, Nullable: false}})
	if result != "list<non-nullable 64-bit signed int>" {
		t.Errorf("expected 'list<non-nullable 64-bit signed int>', got %q", result)
	}
}

func TestListTypeDescWithMinMax(t *testing.T) {
	result := typeDescForType(ast.ListType{
		ElemType: ast.IntegerType{Size: 32, Signed: true, Nullable: false},
		HasMin:   true, MinSize: 1,
		HasMax: true, MaxSize: 10,
	})
	if result != "list{min: 1, max: 10}<non-nullable 32-bit signed int>" {
		t.Errorf("expected 'list{min: 1, max: 10}<non-nullable 32-bit signed int>', got %q", result)
	}
}

func TestArrayTypeDisplay(t *testing.T) {
	input := "print((array{size: 3}<int>).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "array{size: 3}<non-nullable 64-bit signed int>" {
		t.Errorf("expected 'array{size: 3}<non-nullable 64-bit signed int>', got %q", output)
	}
}

func TestListTypeDisplay(t *testing.T) {
	input := "print((list<int>).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "list<non-nullable 64-bit signed int>" {
		t.Errorf("expected 'list<non-nullable 64-bit signed int>', got %q", output)
	}
}

func TestArrayMemberNotFound(t *testing.T) {
	i := New()
	i.env.Define("a", Value{
		IsArray:   true,
		ArrayData: []Value{{Data: 1}},
	})
	expr := &ast.MemberAccess{
		Object: &ast.VarRef{Name: "a"},
		Member: "nonexistent",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for nonexistent array member")
	}
}

func TestTypeRefArrayEval(t *testing.T) {
	i := New()
	at := ast.ArrayType{ElemType: ast.IntegerType{Size: 64, Signed: true}, Size: 4}
	expr := &ast.TypeRef{Type: at, IsType: true}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsType || !val.IsArray {
		t.Errorf("expected IsType and IsArray")
	}
	if _, ok := val.Type.(ast.ArrayType); !ok {
		t.Errorf("expected ArrayType, got %T", val.Type)
	}
}

func TestTypeRefListEval(t *testing.T) {
	i := New()
	lt := ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true}}
	expr := &ast.TypeRef{Type: lt, IsType: true}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsType || !val.IsArray {
		t.Errorf("expected IsType and IsArray")
	}
	if _, ok := val.Type.(ast.ListType); !ok {
		t.Errorf("expected ListType, got %T", val.Type)
	}
}

func TestExprStmtMethodCall(t *testing.T) {
	input := `
var a: list<int> = [1, 2, 3];
a.add(4);
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "[1, 2, 3, 4]" {
		t.Errorf("expected '[1, 2, 3, 4]', got %q", output)
	}
}

func TestArrayTypeEvalTypeMember(t *testing.T) {
	i := New()
	at := ast.ArrayType{ElemType: ast.IntegerType{Size: 32, Signed: true}, Size: 5}
	td := Value{IsType: true, IsArray: true, Type: at}
	val, err := i.evalTypeMember(td, "size")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 5 {
		t.Errorf("expected 5, got %d", val.Data)
	}

	val2, err2 := i.evalTypeMember(td, "length")
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if val2.Data != 5 {
		t.Errorf("expected 5, got %d", val2.Data)
	}
}

func TestArrayTypeEvalMemberNotFound(t *testing.T) {
	i := New()
	at := ast.ArrayType{ElemType: ast.IntegerType{Size: 32, Signed: true}, Size: 3}
	td := Value{IsType: true, IsArray: true, Type: at}
	_, err := i.evalTypeMember(td, "bogus")
	if err == nil {
		t.Errorf("expected error for nonexistent array type member")
	}
}

func TestTypeOfOnArray(t *testing.T) {
	input := `
var a: array{size: 4}<int> = [1, 2, 3, 4];
print((typeof(a)).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "array{size: 4}<non-nullable 64-bit signed int>" {
		t.Errorf("expected 'array{size: 4}<non-nullable 64-bit signed int>', got %q", output)
	}
}

func TestTypeOfOperandError(t *testing.T) {
	input := "print((typeof(undefinedVar)).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable")
	}
}

func TestAssignNonArrayToArrayVar(t *testing.T) {
	input := `
var a: array{size: 2}<int> = 42;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error assigning non-array to array variable")
	}
}

func TestAssignNonArrayToListVar(t *testing.T) {
	input := `
var a: list<int> = 42;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error assigning non-array to list variable")
	}
}

func TestListMinInitializer(t *testing.T) {
	input := `
var a: list{min: 3, max: 10}<int> = [1, 2];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for list below min size")
	}
}

func TestListMaxInitializer(t *testing.T) {
	input := `
var a: list{min: 1, max: 3}<int> = [1, 2, 3, 4];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for list above max size")
	}
}

func TestListDeclRequiresInitWithMin(t *testing.T) {
	input := `
var a: list{min: 1, max: 5}<int>;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for list with min but no initializer")
	}
}

// ---- Coverage edge cases for interpreter ----

func TestEvalTypeMemberListTypeLength(t *testing.T) {
	i := New()
	lt := ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true}}
	td := Value{IsType: true, IsArray: true, Type: lt}
	_, err := i.evalTypeMember(td, "length")
	if err == nil {
		t.Errorf("expected error for ListType length member")
	}
}

func TestEvalTypeMemberListTypeSize(t *testing.T) {
	i := New()
	lt := ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true}}
	td := Value{IsType: true, IsArray: true, Type: lt}
	_, err := i.evalTypeMember(td, "size")
	if err == nil {
		t.Errorf("expected error for ListType size member")
	}
}

func TestEvalTypeMemberFloat(t *testing.T) {
	i := New()
	ft := ast.FloatType{Size: 64}
	td := Value{IsType: true, IsFloat: true, FType: ft}
	val, err := i.evalTypeMember(td, "precision")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 15 {
		t.Errorf("expected precision 15, got %d", val.Data)
	}
}

func TestEvalTypeMemberBool(t *testing.T) {
	i := New()
	bt := ast.BoolType{}
	td := Value{IsType: true, IsBool: true, BType: bt}
	val, err := i.evalTypeMember(td, "size")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 8 {
		t.Errorf("expected size 8, got %d", val.Data)
	}
}

func TestExecArrayDeclInitExprEvalError(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [undefined_ref];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined var in initializer")
	}
}

func TestExecListDeclInitExprEvalError(t *testing.T) {
	input := `
var a: list<int> = [undefined_ref];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined var in initializer")
	}
}

func TestExecListDeclAssignNonList(t *testing.T) {
	input := `
var a: list<int> = 42;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error assigning non-list to list variable")
	}
}

func TestExecListDeclMinBoundErrorOnInit(t *testing.T) {
	input := `
var a: list{min: 3}<int> = [1];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for not enough elements")
	}
}

func TestExecListDeclMaxBoundErrorOnInit(t *testing.T) {
	input := `
var a: list{max: 2}<int> = [1, 2, 3];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for too many elements")
	}
}

func TestExecuteAssignmentOnArrayVar(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2];
a = [3, 4];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err != nil {
		t.Errorf("unexpected error for reassigning array variable: %v", err)
	}
	val, ok := i.env.Get("a")
	if !ok {
		t.Errorf("variable 'a' not found")
	}
	if !val.IsArray {
		t.Errorf("expected 'a' to be an array")
	}
	if len(val.ArrayData) != 2 {
		t.Errorf("expected 2 elements, got %d", len(val.ArrayData))
	}
	if val.ArrayData[0].Data != 3 || val.ArrayData[1].Data != 4 {
		t.Errorf("expected [3, 4], got %v", val)
	}
}

func TestExecuteIndexedAssignUndefinedVar(t *testing.T) {
	input := `
unknown[0] = 1;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for undefined variable")
	}
}

func TestExecuteIndexedAssignNonArray(t *testing.T) {
	input := `
var x: int = 42;
x[0] = 1;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for indexing non-array")
	}
}

func TestExecuteIndexedAssignWrongOp(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2];
a[0] += 1;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for += on indexed assign")
	}
}

func TestExecuteIndexedAssignNullIndex(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}})
	err := i.ExecuteStmt(&ast.Assignment{
		Name:  "a",
		Index: &ast.NullLit{},
		Op:    "=",
		Expr:  &ast.IntegerLit{Value: 42, Untyped: true},
	})
	if err == nil {
		t.Errorf("expected error for null index")
	}
}

func TestExecuteIndexedAssignOutOfBounds(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}})
	err := i.ExecuteStmt(&ast.Assignment{
		Name:  "a",
		Index: &ast.IntegerLit{Value: 5, Untyped: true},
		Op:    "=",
		Expr:  &ast.IntegerLit{Value: 42, Untyped: true},
	})
	if err == nil {
		t.Errorf("expected error for out of bounds index")
	}
}

func TestExecuteIndexedAssignExprError(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}})
	err := i.ExecuteStmt(&ast.Assignment{
		Name:  "a",
		Index: &ast.IntegerLit{Value: 0, Untyped: true},
		Op:    "=",
		Expr:  &ast.VarRef{Name: "undefined"},
	})
	if err == nil {
		t.Errorf("expected error for undefined var in rhs")
	}
}

func TestEvalExprArrayLitError(t *testing.T) {
	i := New()
	_, err := i.evalExpr(&ast.ArrayLit{
		Elements: []ast.Expr{&ast.VarRef{Name: "undefined"}},
	})
	if err == nil {
		t.Errorf("expected error for undefined var in array literal")
	}
}

func TestEvalExprIndexExprNonArray(t *testing.T) {
	i := New()
	i.env.Define("x", Value{Data: 42, Untyped: true})
	_, err := i.evalExpr(&ast.IndexExpr{
		Object: &ast.VarRef{Name: "x"},
		Index:  &ast.IntegerLit{Value: 0, Untyped: true},
	})
	if err == nil {
		t.Errorf("expected error for indexing non-array")
	}
}

func TestEvalExprIndexExprNullIndex(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}})
	_, err := i.evalExpr(&ast.IndexExpr{
		Object: &ast.VarRef{Name: "a"},
		Index:  &ast.NullLit{},
	})
	if err == nil {
		t.Errorf("expected error for null index")
	}
}

func TestEvalExprIndexExprOutOfBounds(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}})
	_, err := i.evalExpr(&ast.IndexExpr{
		Object: &ast.VarRef{Name: "a"},
		Index:  &ast.IntegerLit{Value: 5, Untyped: true},
	})
	if err == nil {
		t.Errorf("expected error for out of bounds index")
	}
}

func TestEvalExprIndexExprEvalObjectError(t *testing.T) {
	i := New()
	_, err := i.evalExpr(&ast.IndexExpr{
		Object: &ast.VarRef{Name: "undefined"},
		Index:  &ast.IntegerLit{Value: 0, Untyped: true},
	})
	if err == nil {
		t.Errorf("expected error for undefined var")
	}
}

func TestEvalExprIndexExprEvalIndexError(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}})
	_, err := i.evalExpr(&ast.IndexExpr{
		Object: &ast.VarRef{Name: "a"},
		Index:  &ast.VarRef{Name: "undefined"},
	})
	if err == nil {
		t.Errorf("expected error for undefined var in index")
	}
}

func TestEvalArrayMemberAddWrongArgs(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "add", Args: []ast.Expr{}})
	if err == nil {
		t.Errorf("expected error for add with no args")
	}
	_, err = i.evalArrayMember(&obj, &ast.MemberAccess{Member: "add", Args: []ast.Expr{
		&ast.IntegerLit{Value: 1, Untyped: true},
		&ast.IntegerLit{Value: 2, Untyped: true},
		&ast.IntegerLit{Value: 3, Untyped: true},
	}})
	if err == nil {
		t.Errorf("expected error for add with >2 args")
	}
}

func TestEvalArrayMemberAddEvalValueError(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "add", Args: []ast.Expr{
		&ast.VarRef{Name: "undefined"},
	}})
	if err == nil {
		t.Errorf("expected error for undefined var in add value")
	}
}

func TestEvalArrayMemberAddEvalIndexError(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "add", Args: []ast.Expr{
		&ast.IntegerLit{Value: 2, Untyped: true},
		&ast.VarRef{Name: "undefined"},
	}})
	if err == nil {
		t.Errorf("expected error for undefined var in add index")
	}
}

func TestEvalArrayMemberAddNullIndex(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "add", Args: []ast.Expr{
		&ast.IntegerLit{Value: 2, Untyped: true},
		&ast.NullLit{},
	}})
	if err == nil {
		t.Errorf("expected error for null index in add")
	}
}

func TestEvalArrayMemberAddOutOfBounds(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "add", Args: []ast.Expr{
		&ast.IntegerLit{Value: 2, Untyped: true},
		&ast.IntegerLit{Value: 5, Untyped: true},
	}})
	if err == nil {
		t.Errorf("expected error for out of bounds index in add")
	}
}

func TestEvalArrayMemberRemoveWrongArgs(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "remove", Args: []ast.Expr{}})
	if err == nil {
		t.Errorf("expected error for remove with no args")
	}
}

func TestEvalArrayMemberRemoveEvalIndexError(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "remove", Args: []ast.Expr{
		&ast.VarRef{Name: "undefined"},
	}})
	if err == nil {
		t.Errorf("expected error for undefined var in remove index")
	}
}

func TestEvalArrayMemberRemoveNullIndex(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "remove", Args: []ast.Expr{
		&ast.NullLit{},
	}})
	if err == nil {
		t.Errorf("expected error for null index in remove")
	}
}

func TestEvalArrayMemberRemoveOutOfBounds(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "remove", Args: []ast.Expr{
		&ast.IntegerLit{Value: 5, Untyped: true},
	}})
	if err == nil {
		t.Errorf("expected error for out of bounds index in remove")
	}
}

func TestArraySizeMismatchTooMany(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2, 3];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for size mismatch (too many)")
	}
}

func TestAssignOpOnArrayVar(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2];
a += 1;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for += on array variable")
	}
}

func TestEvalTypeMemberDefaultNotFound(t *testing.T) {
	i := New()
	it := ast.IntegerType{Size: 64, Signed: true}
	td := Value{IsType: true, IType: it}
	_, err := i.evalTypeMember(td, "nonexistent")
	if err == nil {
		t.Errorf("expected error for nonexistent type member")
	}
}

func TestEvalTypeMemberFloatNotFound(t *testing.T) {
	i := New()
	ft := ast.FloatType{Size: 64}
	td := Value{IsType: true, IsFloat: true, FType: ft}
	_, err := i.evalTypeMember(td, "nonexistent")
	if err == nil {
		t.Errorf("expected error for nonexistent float type member")
	}
}

func TestEvalTypeMemberBoolNotFound(t *testing.T) {
	i := New()
	bt := ast.BoolType{}
	td := Value{IsType: true, IsBool: true, BType: bt}
	_, err := i.evalTypeMember(td, "nonexistent")
	if err == nil {
		t.Errorf("expected error for nonexistent bool type member")
	}
}

func TestEvalMemberAccessNonArrayNonType(t *testing.T) {
	i := New()
	i.env.Define("x", Value{Data: 42, Untyped: true})
	_, err := i.evalExpr(&ast.MemberAccess{
		Object: &ast.VarRef{Name: "x"},
		Member: "foo",
	})
	if err == nil {
		t.Errorf("expected error for accessing member on int")
	}
}

func TestEvalArrayMemberListAddBeyondMax(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}, {Data: 2, Untyped: true}}, Type: ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true}, HasMax: true, MaxSize: 2}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "add", Args: []ast.Expr{
		&ast.IntegerLit{Value: 3, Untyped: true},
	}})
	if err == nil {
		t.Errorf("expected error for adding beyond max")
	}
}

func TestEvalArrayMemberListRemoveBelowMin(t *testing.T) {
	i := New()
	obj := Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}, Type: ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true}, HasMin: true, MinSize: 1}}
	_, err := i.evalArrayMember(&obj, &ast.MemberAccess{Member: "remove", Args: []ast.Expr{
		&ast.IntegerLit{Value: 0, Untyped: true},
	}})
	if err == nil {
		t.Errorf("expected error for removing below min")
	}
}

func TestExecuteIndexedAssignIndexEvalError(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1, Untyped: true}}})
	err := i.ExecuteStmt(&ast.Assignment{
		Name:  "a",
		Index: &ast.VarRef{Name: "undefined"},
		Op:    "=",
		Expr:  &ast.IntegerLit{Value: 42, Untyped: true},
	})
	if err == nil {
		t.Errorf("expected error for undefined var in index")
	}
}

func TestArraySizeInference(t *testing.T) {
	input := `
var a: array<int> = [1, 2, 3];
print((a.length).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "3" {
		t.Errorf("expected '3', got %q", output)
	}
}

func TestArraySizeInferencePrint(t *testing.T) {
	input := `
var a: array<int> = [1, 2, 3];
print((a).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "[1, 2, 3]" {
		t.Errorf("expected '[1, 2, 3]', got %q", output)
	}
}

func TestArrayBracketSyntaxRuntime(t *testing.T) {
	input := `
var a: array[3]<int> = [10, 20, 30];
print((a[1]).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "20" {
		t.Errorf("expected '20', got %q", output)
	}
}

func TestArrayElemTypeMember(t *testing.T) {
	input := "print((array{size: 5}<int>.elem_type).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "non-nullable 64-bit signed int" {
		t.Errorf("expected 'non-nullable 64-bit signed int', got %q", output)
	}
}

func TestArrayElemTypeMemberFloat(t *testing.T) {
	input := "print((array{size: 3}<float>.elem_type).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "non-nullable 64-bit float" {
		t.Errorf("expected 'non-nullable 64-bit float', got %q", output)
	}
}

func TestArrayElemTypeChainedMember(t *testing.T) {
	input := "print((array{size: 5}<int>.elem_type.size).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "64" {
		t.Errorf("expected '64', got %q", output)
	}
}

func TestArrayElemTypeMemberBool(t *testing.T) {
	input := "print((array{size: 3}<bool>.elem_type).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "non-nullable bool" {
		t.Errorf("expected 'non-nullable bool', got %q", output)
	}
}

func TestArrayElemTypeNestedArray(t *testing.T) {
	input := "print((array{size: 3}<array{size: 2}<int>>.elem_type).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "array{size: 2}<non-nullable 64-bit signed int>" {
		t.Errorf("expected 'array{size: 2}<...>', got %q", output)
	}
}

func TestArrayElemTypeNestedList(t *testing.T) {
	input := "print((array{size: 3}<list<int>>.elem_type).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "list<non-nullable 64-bit signed int>" {
		t.Errorf("expected 'list<...>', got %q", output)
	}
}

func TestArrayElemTypeOnListType(t *testing.T) {
	i := New()
	lt := ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true}}
	td := Value{IsType: true, IsArray: true, Type: lt}
	_, err := i.evalTypeMember(td, "elem_type")
	if err == nil {
		t.Errorf("expected error for elem_type on ListType")
	}
}

func TestArrayBracketSyntaxTypeDisplay(t *testing.T) {
	input := "print((array[4]<int>).toString());"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "array{size: 4}<non-nullable 64-bit signed int>" {
		t.Errorf("expected 'array{size: 4}<...>', got %q", output)
	}
}

func TestStringDeclAndPrint(t *testing.T) {
	input := `
var s: string = "hello";
print((s).toString());
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
	if output != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}
}

func TestStringLitPrint(t *testing.T) {
	input := `print("hello world");`
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
	if output != "hello world" {
		t.Errorf("expected 'hello world', got %q", output)
	}
}

func TestStringFixedSize(t *testing.T) {
	input := `
var s: string{size: 5} = "hello";
print((s).toString());
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
	if output != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}
}

func TestStringFixedSizeWrongLength(t *testing.T) {
	input := `
var s: string{size: 5} = "hello!";
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for wrong string length")
	}
}

func TestStringFixedSizeInitWithoutValue(t *testing.T) {
	input := `
var s: string{size: 5};
print((s).toString());
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
	if len(output) != 5 {
		t.Errorf("expected 5 null chars, got %d chars", len(output))
	}
}

func TestStringConcatOperator(t *testing.T) {
	input := `
var a: string = "hello";
var b: string = " world";
print((a + b).toString());
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
	if output != "hello world" {
		t.Errorf("expected 'hello world', got %q", output)
	}
}

func TestStringConcatMethod(t *testing.T) {
	input := `
var a: string = "hello";
print((a.concat(" world")).toString());
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
	if output != "hello world" {
		t.Errorf("expected 'hello world', got %q", output)
	}
}

func TestStringIndex(t *testing.T) {
	input := `
var s: string = "hello";
print((s[0]).toString());
print((s[4]).toString());
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
	if output != "h\no" {
		t.Errorf("expected 'h\\no', got %q", output)
	}
}

func TestStringIndexOutOfBounds(t *testing.T) {
	input := `
var s: string = "hi";
print((s[5]).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected out of bounds error")
	}
}

func TestStringIndexAssign(t *testing.T) {
	input := `
var s: string = "hello";
s[0] = "H";
print((s).toString());
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
	if output != "Hello" {
		t.Errorf("expected 'Hello', got %q", output)
	}
}

func TestStringLength(t *testing.T) {
	input := `
var s: string = "hello";
print((s.length).toString());
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
	if output != "5" {
		t.Errorf("expected '5', got %q", output)
	}
}

func TestStringEquality(t *testing.T) {
	input := `
var a: string = "hello";
var b: string = "hello";
var c: string = "world";
print((a == b).toString());
print((a == c).toString());
print((a != c).toString());
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
	if output != "true\nfalse\ntrue" {
		t.Errorf("expected 'true\\nfalse\\ntrue', got %q", output)
	}
}

func TestStringAddRemove(t *testing.T) {
	input := `
var s: string = "he";
s.add("l");
s.add("l");
s.add("o");
print((s).toString());
s.remove(4);
print((s).toString());
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
	if output != "hello\nhell" {
		t.Errorf("expected 'hello\\nhell', got %q", output)
	}
}

func TestStringAddToFixedFails(t *testing.T) {
	input := `
var s: string{size: 5} = "hello";
s.add("!");
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for add to fixed-size string")
	}
}

func TestStringRemoveFromFixedFails(t *testing.T) {
	input := `
var s: string{size: 5} = "hello";
s.remove(0);
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for remove from fixed-size string")
	}
}

func TestStringTypeRefPrint(t *testing.T) {
	input := `print((string).toString());`
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
	if output != "string" {
		t.Errorf("expected 'string', got %q", output)
	}
}

func TestStringTypeRefWithSize(t *testing.T) {
	input := `print((string{size: 10}).toString());`
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
	if output != "string{size: 10}" {
		t.Errorf("expected 'string{size: 10}', got %q", output)
	}
}

func TestStringEscapeSequences(t *testing.T) {
	input := "print((\"hello\\nworld\").toString());"
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
	if output != "hello\nworld" {
		t.Errorf("expected 'hello\\nworld', got %q", output)
	}
}

func TestStringAssignment(t *testing.T) {
	input := `
var s: string = "hello";
s = "world";
print((s).toString());
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
	if output != "world" {
		t.Errorf("expected 'world', got %q", output)
	}
}

func TestStringAddAtIndex(t *testing.T) {
	input := `
var s: string = "hllo";
s.add("e", 1);
print((s).toString());
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
	if output != "hello" {
		t.Errorf("expected 'hello', got %q", output)
	}
}

func TestStringDynamicBounds(t *testing.T) {
	input := `
var s: string{min: 1, max: 5} = "hi";
print((s.length).toString());
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
	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestStringUTF8(t *testing.T) {
	input := `
var s: string = "héllo";
print((s.length).toString());
print((s[1]).toString());
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
	if output != "5\né" {
		t.Errorf("expected '5\\né', got %q", output)
	}
}

// ---- String coverage tests ----

func TestTypeofStringLiteral(t *testing.T) {
	i := New()
	expr := &ast.TypeOfExpr{Expr: &ast.StringLit{Value: "hello", Untyped: true}}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsString {
		t.Errorf("expected IsString to be true")
	}
}

func TestStringTypeMemberAccess(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.StringType{}},
		Member: "size",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 0 {
		t.Errorf("expected 0, got %d", val.Data)
	}
}

func TestStringTypeMemberUnknown(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Type: ast.StringType{}},
		Member: "foo",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for unknown string type member")
	}
}

func TestValueStringForStringInArray(t *testing.T) {
	v := Value{
		IsArray:   true,
		ArrayData: []Value{{IsString: true, StringData: "hello"}, {IsString: true, StringData: "world"}},
	}
	result := v.String()
	if result != "[hello, world]" {
		t.Errorf("expected '[hello, world]', got %q", result)
	}
}

func TestTypeDescForTypeStringSize(t *testing.T) {
	v := Value{Type: ast.StringType{Size: 5}, IsType: true, IsString: true}
	result := typeDescForVal(v)
	if !strings.Contains(result, "string{size: 5}") {
		t.Errorf("expected 'string{size: 5}', got %q", result)
	}
}

func TestTypeDescForTypeStringHasMin(t *testing.T) {
	v := Value{Type: ast.StringType{HasMin: true, MinSize: 3}, IsType: true, IsString: true}
	result := typeDescForVal(v)
	if !strings.Contains(result, "string{min: 3}") {
		t.Errorf("expected 'string{min: 3}', got %q", result)
	}
}

func TestTypeDescForTypeStringHasMax(t *testing.T) {
	v := Value{Type: ast.StringType{HasMax: true, MaxSize: 10}, IsType: true, IsString: true}
	result := typeDescForVal(v)
	if !strings.Contains(result, "string{max: 10}") {
		t.Errorf("expected 'string{max: 10}', got %q", result)
	}
}

func TestTypeDescForTypeStringHasMinMax(t *testing.T) {
	v := Value{Type: ast.StringType{HasMin: true, MinSize: 3, HasMax: true, MaxSize: 10}, IsType: true, IsString: true}
	result := typeDescForVal(v)
	if !strings.Contains(result, "string{min: 3, max: 10}") {
		t.Errorf("expected 'string{min: 3, max: 10}', got %q", result)
	}
}

func TestEvalTypedStringLit(t *testing.T) {
	i := New()
	expr := &ast.StringLit{Value: "hello", SType: ast.StringType{Size: 5}, Untyped: false}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsString {
		t.Errorf("expected IsString to be true")
	}
	if val.Untyped {
		t.Errorf("expected Untyped to be false")
	}
	if val.SType.Size != 5 {
		t.Errorf("expected Size 5, got %d", val.SType.Size)
	}
}

func TestExecStringDeclDefaultEmpty(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "s",
		Type:  ast.StringType{},
		SType: ast.StringType{},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := i.env.Get("s")
	if !ok {
		t.Fatal("expected variable s to exist")
	}
	if !val.IsString {
		t.Errorf("expected IsString to be true")
	}
}

func TestExecStringDeclFixedSizeNoInit(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "s",
		Type:  ast.StringType{Size: 5},
		SType: ast.StringType{Size: 5},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := i.env.Get("s")
	if !ok {
		t.Fatal("expected variable s to exist")
	}
	if !val.IsString {
		t.Errorf("expected IsString to be true")
	}
}

func TestExecStringDeclWrongTypeError(t *testing.T) {
	i := New()
	err := i.execStringDecl(&ast.VarDecl{
		Name:  "s",
		SType: ast.StringType{},
		Expr:  &ast.IntegerLit{Value: 5, Untyped: true},
	}, ast.StringType{})
	if err == nil {
		t.Errorf("expected error assigning int to string")
	}
}

func TestExecStringDeclExprEvalError(t *testing.T) {
	i := New()
	err := i.execStringDecl(&ast.VarDecl{
		Name:  "s",
		SType: ast.StringType{},
		Expr:  &ast.PrintStmt{},
	}, ast.StringType{})
	if err == nil {
		t.Errorf("expected eval error")
	}
}

func TestExecStringDeclSizeMismatchError(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "s",
		Type:  ast.StringType{Size: 5},
		SType: ast.StringType{Size: 5},
		Expr:  &ast.StringLit{Value: "hello!", Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected size mismatch error")
	}
}

func TestExecStringDeclHasMinError(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "s",
		Type:  ast.StringType{HasMin: true, MinSize: 3},
		SType: ast.StringType{HasMin: true, MinSize: 3},
		Expr:  &ast.StringLit{Value: "ab", Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected min size error")
	}
}

func TestExecStringDeclHasMaxError(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "s",
		Type:  ast.StringType{HasMax: true, MaxSize: 2},
		SType: ast.StringType{HasMax: true, MaxSize: 2},
		Expr:  &ast.StringLit{Value: "abc", Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected max size error")
	}
}

func TestExecStringDeclMinNoInitError(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "s",
		Type:  ast.StringType{HasMin: true, MinSize: 3},
		SType: ast.StringType{HasMin: true, MinSize: 3},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error for missing initializer with min constraint")
	}
}

func TestExecStringDeclSuccess(t *testing.T) {
	i := New()
	stmt := &ast.VarDecl{
		Name:  "s",
		Type:  ast.StringType{Size: 5},
		SType: ast.StringType{Size: 5},
		Expr:  &ast.StringLit{Value: "hello", Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := i.env.Get("s")
	if !ok {
		t.Fatal("expected variable s to exist")
	}
	if val.StringData != "hello" {
		t.Errorf("expected 'hello', got %q", val.StringData)
	}
}

func TestExecuteAssignmentStringOpNotEqualError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name: "s",
		Op:   "+=",
		Expr: &ast.StringLit{Value: "world", Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for operator on string")
	}
}

func TestExecuteAssignmentStringWrongTypeError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello", Type: ast.StringType{}, SType: ast.StringType{}})
	i.env = env
	stmt := &ast.Assignment{
		Name: "s",
		Op:   "=",
		Expr: &ast.IntegerLit{Value: 5, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error assigning int to string")
	}
}

func TestExecuteAssignmentStringSizeMismatchError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{
		IsString: true, StringData: "hello",
		Type: ast.StringType{Size: 5}, SType: ast.StringType{Size: 5},
	})
	i.env = env
	stmt := &ast.Assignment{
		Name: "s",
		Op:   "=",
		Expr: &ast.StringLit{Value: "hello!", Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected size mismatch error")
	}
}

func TestExecuteAssignmentStringHasMinError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{
		IsString: true, StringData: "hello",
		Type: ast.StringType{HasMin: true, MinSize: 3}, SType: ast.StringType{HasMin: true, MinSize: 3},
	})
	i.env = env
	stmt := &ast.Assignment{
		Name: "s",
		Op:   "=",
		Expr: &ast.StringLit{Value: "ab", Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected min size error")
	}
}

func TestExecuteAssignmentStringHasMaxError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{
		IsString: true, StringData: "hi",
		Type: ast.StringType{HasMax: true, MaxSize: 2}, SType: ast.StringType{HasMax: true, MaxSize: 2},
	})
	i.env = env
	stmt := &ast.Assignment{
		Name: "s",
		Op:   "=",
		Expr: &ast.StringLit{Value: "abc", Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected max size error")
	}
}

func TestExecuteAssignmentStringSuccess(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello", Type: ast.StringType{}, SType: ast.StringType{}})
	i.env = env
	stmt := &ast.Assignment{
		Name: "s",
		Op:   "=",
		Expr: &ast.StringLit{Value: "world", Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("s")
	if val.StringData != "world" {
		t.Errorf("expected 'world', got %q", val.StringData)
	}
}

func TestExecuteAssignmentStringEvalExprError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello", Type: ast.StringType{}, SType: ast.StringType{}})
	i.env = env
	stmt := &ast.Assignment{
		Name: "s",
		Op:   "=",
		Expr: &ast.PrintStmt{},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected eval error")
	}
}

func TestExecuteIndexedAssignStringOpNotEqualError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name:  "s",
		Op:    "+=",
		Index: &ast.IntegerLit{Value: 0},
		Expr:  &ast.StringLit{Value: "A", Untyped: true},
	}
	err := i.executeIndexedAssign(stmt)
	if err == nil {
		t.Errorf("expected error for += on string index")
	}
}

func TestExecuteIndexedAssignStringIndexEvalError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name:  "s",
		Op:    "=",
		Index: &ast.PrintStmt{},
		Expr:  &ast.StringLit{Value: "A", Untyped: true},
	}
	err := i.executeIndexedAssign(stmt)
	if err == nil {
		t.Errorf("expected index eval error")
	}
}

func TestExecuteIndexedAssignStringNullIndexError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name:  "s",
		Op:    "=",
		Index: &ast.NullLit{},
		Expr:  &ast.StringLit{Value: "A", Untyped: true},
	}
	err := i.executeIndexedAssign(stmt)
	if err == nil {
		t.Errorf("expected null index error")
	}
}

func TestExecuteIndexedAssignStringOutOfBoundsError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name:  "s",
		Op:    "=",
		Index: &ast.IntegerLit{Value: 10},
		Expr:  &ast.StringLit{Value: "A", Untyped: true},
	}
	err := i.executeIndexedAssign(stmt)
	if err == nil {
		t.Errorf("expected out of bounds error")
	}
}

func TestExecuteIndexedAssignStringRightValEvalError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name:  "s",
		Op:    "=",
		Index: &ast.IntegerLit{Value: 0},
		Expr:  &ast.PrintStmt{},
	}
	err := i.executeIndexedAssign(stmt)
	if err == nil {
		t.Errorf("expected right value eval error")
	}
}

func TestExecuteIndexedAssignStringNotSingleCharError(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name:  "s",
		Op:    "=",
		Index: &ast.IntegerLit{Value: 0},
		Expr:  &ast.StringLit{Value: "AB", Untyped: true},
	}
	err := i.executeIndexedAssign(stmt)
	if err == nil {
		t.Errorf("expected error for multi-char assignment")
	}
}

func TestExecuteIndexedAssignStringSuccess(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	stmt := &ast.Assignment{
		Name:  "s",
		Op:    "=",
		Index: &ast.IntegerLit{Value: 0},
		Expr:  &ast.StringLit{Value: "H", Untyped: true},
	}
	err := i.executeIndexedAssign(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("s")
	if val.StringData != "Hello" {
		t.Errorf("expected 'Hello', got %q", val.StringData)
	}
}

func TestEvalStringMemberConcatArgCountError(t *testing.T) {
	i := New()
	// string.concat() with no args
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "concat",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for concat with no args")
	}
}

func TestEvalStringMemberConcatNonStringError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "concat",
		Args:   []ast.Expr{&ast.IntegerLit{Value: 5, Untyped: true}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for concat with non-string")
	}
}

func TestEvalStringMemberConcatSuccess(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "concat",
		Args:   []ast.Expr{&ast.StringLit{Value: " world", Untyped: true}},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.StringData != "hello world" {
		t.Errorf("expected 'hello world', got %q", val.StringData)
	}
}

func TestEvalStringMemberAddArgCountError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "add",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for add with no args")
	}
}

func TestEvalStringMemberAddArgEvalError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "add",
		Args:   []ast.Expr{&ast.PrintStmt{}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for add with bad arg")
	}
}

func TestEvalStringMemberAddNotSingleCharError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "AB", Untyped: true}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for add with multi-char")
	}
}

func TestEvalStringMemberAddSuccess(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Null {
		t.Errorf("expected null return from add")
	}
}

func TestEvalStringMemberAddWithIndexSuccess(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}, &ast.IntegerLit{Value: 0}},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Null {
		t.Errorf("expected null return from add")
	}
}

func TestEvalStringMemberAddSecondArgEvalError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}, &ast.PrintStmt{}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for add with bad second arg")
	}
}

func TestEvalStringMemberAddNullIndexError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}, &ast.NullLit{}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for null index")
	}
}

func TestEvalStringMemberAddMaxCapacityError(t *testing.T) {
	i := New()
	// Use a string with HasMax constraint to trigger max capacity check
	sval := Value{IsString: true, StringData: "hello", Type: ast.StringType{HasMax: true, MaxSize: 5}}
	expr := &ast.MemberAccess{
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}},
	}
	_, err := i.evalStringMember(&sval, expr)
	if err == nil {
		t.Errorf("expected max capacity error")
	}
}

func TestEvalStringMemberAddBoundsError(t *testing.T) {
	i := New()
	sval := Value{IsString: true, StringData: "hello"}
	expr := &ast.MemberAccess{
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}, &ast.IntegerLit{Value: 10}},
	}
	_, err := i.evalStringMember(&sval, expr)
	if err == nil {
		t.Errorf("expected bounds error")
	}
}

func TestEvalStringMemberRemoveArgCountError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "remove",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for remove with no args")
	}
}

func TestEvalStringMemberRemoveArgEvalError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "remove",
		Args:   []ast.Expr{&ast.PrintStmt{}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for remove with bad arg")
	}
}

func TestEvalStringMemberRemoveNullIndexError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "remove",
		Args:   []ast.Expr{&ast.NullLit{}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for remove null index")
	}
}

func TestEvalStringMemberRemoveBoundsError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "remove",
		Args:   []ast.Expr{&ast.IntegerLit{Value: 10}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected bounds error")
	}
}

func TestEvalStringMemberRemoveMinCapacityError(t *testing.T) {
	i := New()
	sval := Value{IsString: true, StringData: "hello", Type: ast.StringType{HasMin: true, MinSize: 5}}
	expr := &ast.MemberAccess{
		Member: "remove",
		Args:   []ast.Expr{&ast.IntegerLit{Value: 0}},
	}
	_, err := i.evalStringMember(&sval, expr)
	if err == nil {
		t.Errorf("expected min capacity error")
	}
}

func TestEvalStringMemberRemoveSuccess(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "remove",
		Args:   []ast.Expr{&ast.IntegerLit{Value: 0}},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.StringData != "h" {
		t.Errorf("expected removed char 'h', got %q", val.StringData)
	}
}

func TestEvalStringMemberAddFixedSizeError(t *testing.T) {
	i := New()
	sval := Value{IsString: true, StringData: "hello", Type: ast.StringType{Size: 5}}
	expr := &ast.MemberAccess{
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}},
	}
	_, err := i.evalStringMember(&sval, expr)
	if err == nil {
		t.Errorf("expected error for add on fixed-size string")
	}
}

func TestEvalStringMemberRemoveFixedSizeError(t *testing.T) {
	i := New()
	sval := Value{IsString: true, StringData: "hello", Type: ast.StringType{Size: 5}}
	expr := &ast.MemberAccess{
		Member: "remove",
		Args:   []ast.Expr{&ast.IntegerLit{Value: 0}},
	}
	_, err := i.evalStringMember(&sval, expr)
	if err == nil {
		t.Errorf("expected error for remove on fixed-size string")
	}
}

func TestEvalStringMemberUnknown(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "unknown",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for unknown member")
	}
}

func TestEvalStringMemberConcatEvalError(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.StringLit{Value: "hello", Untyped: true},
		Member: "concat",
		Args:   []ast.Expr{&ast.PrintStmt{}},
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for concat with bad arg")
	}
}

func TestEvalBinaryStringConcatTypeMismatchError(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.StringLit{Value: "hello", Untyped: true},
		Op:    "+",
		Right: &ast.IntegerLit{Value: 5, Untyped: true},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected type mismatch error for concat")
	}
}

func TestEvalStringMemberAddSuccessVarRef(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	expr := &ast.MemberAccess{
		Object: &ast.VarRef{Name: "s"},
		Member: "add",
		Args:   []ast.Expr{&ast.StringLit{Value: "!", Untyped: true}},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.Null {
		t.Errorf("expected null return")
	}
	result, _ := i.env.Get("s")
	if result.StringData != "hello!" {
		t.Errorf("expected 'hello!', got %q", result.StringData)
	}
}

func TestEvalStringMemberRemoveSuccessVarRef(t *testing.T) {
	i := New()
	env := NewEnv()
	env.Define("s", Value{IsString: true, StringData: "hello"})
	i.env = env
	expr := &ast.MemberAccess{
		Object: &ast.VarRef{Name: "s"},
		Member: "remove",
		Args:   []ast.Expr{&ast.IntegerLit{Value: 0}},
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.StringData != "h" {
		t.Errorf("expected 'h', got %q", val.StringData)
	}
	result, _ := i.env.Get("s")
	if result.StringData != "ello" {
		t.Errorf("expected 'ello', got %q", result.StringData)
	}
}

func TestPrintRequiresString(t *testing.T) {
	input := "print(42);"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for print with non-string, got none")
	}
	if err.Error() != "print requires a string argument, got untyped int literal" {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestToStringAsProperty(t *testing.T) {
	input := "print((42).toString);"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for toString as property, got none")
	}
	if err.Error() != `value of type untyped int literal has no attribute "toString"` {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestToStringWithArgs(t *testing.T) {
	input := "print((42).toString(42));"
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for toString with args, got none")
	}
	if err.Error() != "toString takes no arguments" {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestVarRedeclareSameScope(t *testing.T) {
	input := `
var x: int = 1;
var x: bool = true;
print(x.toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestArrayCompoundAssign(t *testing.T) {
	input := `
var a: array<int> = [];
a += 1;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err == nil {
			t.Errorf("expected error, got none")
		}
	})
	_ = output
}

func TestArrayInvalidMember(t *testing.T) {
	input := `
var a: array<int> = [];
a.foo;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for invalid member, got none")
	}
}

func TestListCompoundAssign(t *testing.T) {
	input := `
var a: list<int> = [];
a += 1;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestListInvalidMember(t *testing.T) {
	input := `
var a: list<int> = [];
a.foo;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for invalid member, got none")
	}
}

func TestArrayNoSpaceBeforeAssignRuntime(t *testing.T) {
	input := `
var a: array<int>=[1, 2, 3];
print(a.toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "[1, 2, 3]" {
		t.Errorf("expected '[1, 2, 3]', got %q", output)
	}
}

func TestListNoSpaceBeforeAssignRuntime(t *testing.T) {
	input := `
var a: list<int>=[1, 2, 3];
print(a.toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "[1, 2, 3]" {
		t.Errorf("expected '[1, 2, 3]', got %q", output)
	}
}

func TestShadowingElseBlock(t *testing.T) {
	input := `
var x: int = 1;
if (false) { } else {
	var x: int = 2;
	print((x).toString());
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "2\n1" {
		t.Errorf("expected '2\\n1', got %q", output)
	}
}

func TestShadowingElseIfChain(t *testing.T) {
	input := `
var x: int = 1;
if (false) { } else if (true) {
	var x: int = 2;
	print((x).toString());
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "2\n1" {
		t.Errorf("expected '2\\n1', got %q", output)
	}
}

func TestShadowingForBody(t *testing.T) {
	input := `
var x: int = 1;
for (var i: int = 0; i < 2; i = i + 1) {
	var x: int = 99;
	print((x).toString());
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "99\n99\n1" {
		t.Errorf("expected '99\\n99\\n1', got %q", output)
	}
}

func TestShadowingWhileBody(t *testing.T) {
	input := `
var x: int = 1;
var done: bool = false;
while (!done) {
	var x: int = 99;
	print((x).toString());
	done = true;
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "99\n1" {
		t.Errorf("expected '99\\n1', got %q", output)
	}
}

func TestShadowingDifferentTypes(t *testing.T) {
	input := `
var x: int = 42;
if (true) {
	var x: bool = true;
	print((x).toString());
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true\n42" {
		t.Errorf("expected 'true\\n42', got %q", output)
	}
}

func TestShadowingDeeplyNested(t *testing.T) {
	input := `
var x: int = 1;
if (true) {
	var x: int = 2;
	if (true) {
		var x: int = 3;
		print((x).toString());
	}
	print((x).toString());
}
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "3\n2\n1" {
		t.Errorf("expected '3\\n2\\n1', got %q", output)
	}
}

func TestShadowingForInitScope(t *testing.T) {
	input := `
var x: int = 1;
for (var x: int = 99; x < 1; x = x + 1) { }
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "1" {
		t.Errorf("expected '1', got %q", output)
	}
}

func TestRedeclareSameScope(t *testing.T) {
	input := `
var x: int = 1;
var x: bool = true;
print((x).toString());
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestRefDeclPrimitiveShared(t *testing.T) {
	input := `var b: int = 42; ref a = b; a = 10; print((b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "10" {
		t.Errorf("expected '10', got %q", output)
	}
}

func TestRefDeclPrimitiveIs(t *testing.T) {
	input := `var b: int = 42; ref a = b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestVarDeclPrimitiveIsIndependent(t *testing.T) {
	input := `var b: int = 42; var a: int = b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "false" {
		t.Errorf("expected 'false', got %q", output)
	}
}

func TestRefDeclArrayShared(t *testing.T) {
	input := `var b: array<int> = [1, 2, 3]; ref a = b; a.add(4); print((b.length).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "4" {
		t.Errorf("expected '4', got %q", output)
	}
}

func TestRefDeclArrayIs(t *testing.T) {
	input := `var b: array<int> = [1, 2, 3]; ref a = b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestCopyExprArraySameValues(t *testing.T) {
	input := `var b: array<int> = [1, 2, 3]; var a: array<int> = copy b; print((a[1]).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "2" {
		t.Errorf("expected '2', got %q", output)
	}
}

func TestCopyExprArrayNewOuter(t *testing.T) {
	input := `var b: array<array{size: 2}<int>> = [[1, 2], [3, 4]]; var a: array<array{size: 2}<int>> = copy b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "false" {
		t.Errorf("expected 'false' (copy creates new outer array), got %q", output)
	}
}

func TestCopyExprPrimitive(t *testing.T) {
	input := `var b: int = 42; var a: int = copy b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "false" {
		t.Errorf("expected 'false' (copy of primitive creates independent value), got %q", output)
	}
}

func TestIsOperatorSameVar(t *testing.T) {
	input := `var a: int = 42; print((a is a).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestIsOperatorDifferentVars(t *testing.T) {
	input := `var a: int = 42; var b: int = 42; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "false" {
		t.Errorf("expected 'false', got %q", output)
	}
}

func TestAssignRefRedirect(t *testing.T) {
	input := `var a: int = 5; var b: int = 10; a = ref b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true' after redirect, got %q", output)
	}
}

func TestAssignRefRedirectMutation(t *testing.T) {
	input := `var a: int = 5; var b: int = 10; a = ref b; a = 20; print((b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "20" {
		t.Errorf("expected '20' (redirected a shares with b), got %q", output)
	}
}

func TestRefDeclArrayIndexedAssignShared(t *testing.T) {
	input := `var b: array<int> = [1, 2, 3]; ref a = b; a[1] = 99; print((b[1]).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "99" {
		t.Errorf("expected '99' (ref array indexed assign affects original), got %q", output)
	}
}

func TestRefDeclIncDecShared(t *testing.T) {
	input := `var b: int = 5; ref a = b; a++; print((b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "6" {
		t.Errorf("expected '6' (inc/dec on ref affects original), got %q", output)
	}
}

func TestRefDeclTyped(t *testing.T) {
	input := `var b: int = 42; ref a: int{size: 32, signed: true, nullable: false} = b; a = 10; print((b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "10" {
		t.Errorf("expected '10', got %q", output)
	}
}

func TestRefParseErrorLiteral(t *testing.T) {
	input := `ref a = 5;`
	l := lexer.New(input)
	p := parser.New(l)
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Errorf("expected parse error for ref with literal")
	}
}

func TestRefDeclStringShared(t *testing.T) {
	input := `var b: string = "hello"; ref a = b; a.add("w"); print((b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "hellow" {
		t.Errorf("expected 'hellow' (ref string mutation affects original), got %q", output)
	}
}

func TestRefDeclGetPtrOuterEnv(t *testing.T) {
	input := `var b: int = 42; if (true) { ref a = b; a = 10; print((b).toString()); }`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "10" {
		t.Errorf("expected '10', got %q", output)
	}
}

func TestRefDeclFloatType(t *testing.T) {
	input := `var b: float = 3.14; ref a: float = b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestRefDeclFloatTypeError(t *testing.T) {
	input := `var b: int = 42; ref a: float = b;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected type error for float ref with int variable")
	}
}

func TestRefDeclBoolType(t *testing.T) {
	input := `var b: bool = true; ref a: bool = b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestRefDeclBoolTypeError(t *testing.T) {
	input := `var b: int = 42; ref a: bool = b;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected type error for bool ref with int variable")
	}
}

func TestRefDeclStringType(t *testing.T) {
	input := `var b: string = "hello"; ref a: string = b; print((a is b).toString());`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	output := captureOutput(func() {
		err := i.Run(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if output != "true" {
		t.Errorf("expected 'true', got %q", output)
	}
}

func TestRefDeclStringTypeError(t *testing.T) {
	input := `var b: int = 42; ref a: string = b;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected type error for string ref with int variable")
	}
}

func TestRefDeclIntegerTypeError(t *testing.T) {
	input := `var b: float = 3.14; ref a: int{size: 32, signed: true, nullable: false} = b;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected type error for int ref with float variable")
	}
}

func TestRefDeclNonVarRefError(t *testing.T) {
	i := New()
	stmt := &ast.RefDecl{
		Name: "a",
		Expr: &ast.IntegerLit{Value: 42, Untyped: true},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected error for ref with non-variable expression")
	}
}

func TestRefDeclArrayTypeViaAST(t *testing.T) {
	i := New()
	i.env.Define("b", Value{
		IsArray:   true,
		ArrayData: []Value{{Data: 1, IType: ast.IntegerType{Size: 64, Signed: true}}},
	})
	stmt := &ast.RefDecl{
		Name: "a",
		Type: ast.ArrayType{ElemType: ast.IntegerType{Size: 64, Signed: true}},
		Expr: &ast.VarRef{Name: "b"},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRefDeclArrayTypeErrorViaAST(t *testing.T) {
	i := New()
	i.env.Define("b", Value{Data: 42, IType: ast.IntegerType{Size: 64, Signed: true}})
	stmt := &ast.RefDecl{
		Name: "a",
		Type: ast.ArrayType{ElemType: ast.IntegerType{Size: 64, Signed: true}},
		Expr: &ast.VarRef{Name: "b"},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type error for array ref with int variable")
	}
}

func TestRefDeclListTypeViaAST(t *testing.T) {
	i := New()
	i.env.Define("b", Value{
		IsArray:   true,
		ArrayData: []Value{{Data: 1, IType: ast.IntegerType{Size: 64, Signed: true}}},
	})
	stmt := &ast.RefDecl{
		Name: "a",
		Type: ast.ListType{},
		Expr: &ast.VarRef{Name: "b"},
	}
	err := i.executeStmt(stmt)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRefDeclListTypeErrorViaAST(t *testing.T) {
	i := New()
	i.env.Define("b", Value{Data: 42, IType: ast.IntegerType{Size: 64, Signed: true}})
	stmt := &ast.RefDecl{
		Name: "a",
		Type: ast.ListType{},
		Expr: &ast.VarRef{Name: "b"},
	}
	err := i.executeStmt(stmt)
	if err == nil {
		t.Errorf("expected type error for list ref with int variable")
	}
}

func TestExecuteAssignmentRefIndexedError(t *testing.T) {
	i := New()
	stmt := &ast.Assignment{
		Name:  "x",
		Index: &ast.IntegerLit{Value: 0, Untyped: true},
		Op:    "=",
		Expr:  &ast.VarRef{Name: "y"},
		IsRef: true,
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for ref with indexed assignment")
	}
}

func TestExecuteAssignmentRefNonVarRefError(t *testing.T) {
	i := New()
	i.env.Define("x", Value{Data: 10, IType: ast.IntegerType{Size: 64, Signed: true}})
	stmt := &ast.Assignment{
		Name:  "x",
		Op:    "=",
		Expr:  &ast.IntegerLit{Value: 42, Untyped: true},
		IsRef: true,
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for ref assignment with non-variable expression")
	}
}

func TestCopyExprEvalError(t *testing.T) {
	input := `print(copy nonexistent);`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for copy with undefined variable")
	}
}

func TestIsExprLeftEvalError(t *testing.T) {
	input := `var y: int = 5; print(nonexistent is y);`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for is with undefined left operand")
	}
}

func TestIsExprRightEvalError(t *testing.T) {
	input := `var x: int = 5; print(x is nonexistent);`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for is with undefined right operand")
	}
}

func TestRefDeclUndefinedVar(t *testing.T) {
	input := `ref a = nonexistent;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for ref to undefined variable")
	}
}

func TestExecuteAssignmentRefUndefinedTarget(t *testing.T) {
	input := `x = ref nonexistent;`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for ref assignment with undefined target variable")
	}
}

func TestEvalBinaryModulo(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 17, Untyped: true},
		Op:    "%",
		Right: &ast.IntegerLit{Value: 5, Untyped: true},
	}
	val, err := i.evalBinary(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.Data != 2 {
		t.Errorf("expected 2, got %d", val.Data)
	}
}

func TestEvalBinaryModuloFloatError(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.FloatLit{Value: 10.5, Untyped: true},
		Op:    "%",
		Right: &ast.FloatLit{Value: 3.0, Untyped: true},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected error for modulo with float operands")
	}
}

func TestEvalBinaryModuloByZero(t *testing.T) {
	i := New()
	expr := &ast.BinaryExpr{
		Left:  &ast.IntegerLit{Value: 10, Untyped: true},
		Op:    "%",
		Right: &ast.IntegerLit{Value: 0, Untyped: true},
	}
	_, err := i.evalBinary(expr)
	if err == nil {
		t.Errorf("expected error for modulo by zero")
	}
}

func TestExecuteAssignmentModEq(t *testing.T) {
	i := New()
	i.env.Define("x", Value{Data: 17, IType: ast.IntegerType{Size: 64, Signed: true}})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%=",
		Expr: &ast.IntegerLit{Value: 5, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, _ := i.env.Get("x")
	if val.Data != 2 {
		t.Errorf("expected 2, got %d", val.Data)
	}
}

func TestExecuteAssignmentModByZero(t *testing.T) {
	i := New()
	i.env.Define("x", Value{Data: 10, IType: ast.IntegerType{Size: 64, Signed: true}})
	stmt := &ast.Assignment{
		Name: "x",
		Op:   "%=",
		Expr: &ast.IntegerLit{Value: 0, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for modulo by zero")
	}
}

func TestExecuteAssignmentArrayNonArrayError(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2];
a = 42;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for assigning non-array to array variable")
	}
}

func TestExecuteAssignmentArraySizeMismatch(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2];
a = [3, 4, 5];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for array size mismatch")
	}
}

func TestExecuteAssignmentListMinError(t *testing.T) {
	input := `
var a: list{min: 2}<int> = [1, 2];
a = [3];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for list min constraint violation")
	}
}

func TestExecuteAssignmentListMaxError(t *testing.T) {
	input := `
var a: list{max: 3}<int> = [1, 2];
a = [3, 4, 5, 6];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for list max constraint violation")
	}
}

func TestExecuteAssignmentListValidReassign(t *testing.T) {
	input := `
var a: list<int> = [1, 2];
a = [3, 4, 5];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := i.env.Get("a")
	if !ok {
		t.Fatalf("variable 'a' not found")
	}
	if !val.IsArray {
		t.Errorf("expected 'a' to be an array")
	}
	if len(val.ArrayData) != 3 {
		t.Errorf("expected 3 elements, got %d", len(val.ArrayData))
	}
}

func TestExecuteAssignmentFloatModEqError(t *testing.T) {
	input := `
var x: float{nullable: false} = 10.5;
x %= 3.0;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for modulo on float")
	}
}

func TestExecuteAssignmentArrayCompoundError(t *testing.T) {
	i := New()
	i.env.Define("a", Value{IsArray: true, ArrayData: []Value{{Data: 1}, {Data: 2}}, Type: ast.ArrayType{Size: 2, ElemType: ast.IntegerType{Size: 64, Signed: true}}})
	stmt := &ast.Assignment{
		Name: "a",
		Op:   "%=",
		Expr: &ast.IntegerLit{Value: 2, Untyped: true},
	}
	err := i.executeAssignment(stmt)
	if err == nil {
		t.Errorf("expected error for compound assign on array")
	}
}

func TestExecuteAssignmentArrayReassignEvalError(t *testing.T) {
	input := `
var a: array{size: 2}<int> = [1, 2];
a = nonexistent;
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err == nil {
		t.Errorf("expected error for reassigning array from undefined variable")
	}
}

func TestExecuteAssignmentArraySizeInference(t *testing.T) {
	input := `
var a: array{size: auto}<int>;
a = [1, 2, 3];
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	i := New()
	err := i.Run(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, ok := i.env.Get("a")
	if !ok {
		t.Fatalf("variable 'a' not found")
	}
	if len(val.ArrayData) != 3 {
		t.Errorf("expected 3 elements, got %d", len(val.ArrayData))
	}
}
