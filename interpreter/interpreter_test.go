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
var x: int{size: 32, signed: true, nullable: false} = 10;
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
var x: int{size: 32, signed: true, nullable: false} = 10;
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
var x: int{size: 32, signed: true, nullable: false} = 10;
var y: int{size: 32, signed: true, nullable: false} = 20;
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
var x: int{size: 32, signed: true, nullable: true} = null;
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
var x: int{size: 8, signed: true, nullable: false} = 1000;
print(x);
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
		{Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 10, IsFloat: false}, "32-bit signed int(10)"},
		{Value{IType: ast.IntegerType{Size: 16, Signed: false}, Data: 5, IsFloat: false}, "16-bit unsigned int(5)"},
		{Value{FType: ast.FloatType{Size: 32}, FData: 3.14, IsFloat: true}, "32-bit float(3.14)"},
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
		{ast.IntegerType{Size: 32, Signed: true, Nullable: false}, ast.FloatType{}, false, "32-bit signed int"},
		{ast.IntegerType{Size: 16, Signed: false, Nullable: true}, ast.FloatType{}, false, "nullable 16-bit unsigned int"},
		{ast.IntegerType{Size: 64, Signed: true, Nullable: true}, ast.FloatType{}, false, "nullable 64-bit signed int"},
		{ast.IntegerType{}, ast.FloatType{Size: 32}, true, "32-bit float"},
		{ast.IntegerType{}, ast.FloatType{Size: 64}, true, "64-bit float"},
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
var x: int{size: 8, signed: true, nullable: false} = 10;
var y: int{size: 32, signed: true, nullable: false} = x;
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
var x: int{size: 8, signed: true, nullable: false} = 10;
x = 200;
print(x);
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
var x: int{size: 32, signed: true, nullable: false} = 10;
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

func TestExecuteAssignmentWithUntypedExpr(t *testing.T) {
	input := `
var x: int{size: 32, signed: true, nullable: false} = 10;
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
var x: int{size: 32, signed: true, nullable: false} = 10;
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
var x: int{size: 32, signed: true, nullable: false} = 10;
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
var x: int{size: 32, signed: true, nullable: false} = 10;
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
var x: int{size: 32, signed: true, nullable: true} = 10;
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

func TestFloatVarDecl(t *testing.T) {
	input := `
var a: float{size: 32} = 3.14;
print(a);
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
print(a);
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
print(a);
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
print(b);
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
print(b);
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
print(b);
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
print(a);
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
print(a + b);
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
print(a);
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
	input := "print(.1);"
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
	input := "print(-.5);"
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
	input := "print(.15 + .1);"
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
	input := "print(.1e2);"
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
print(a + b);
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
print(a + b);
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
	env.Set("x", val)

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
	i.env.Set("x", Value{IType: ast.IntegerType{Size: 32, Signed: true, Nullable: true}, Data: 10})
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
		Op:    "%", // modulo not supported
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
	i.env.Set("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 10})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 5.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 2.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 5.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 2.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 5.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 2.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 10.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 16}, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 16}, FData: 60000.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 16}, FData: 60000.0, IsFloat: true})
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
	i.env.Set("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, IsFloat: false})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 16}, FData: 100.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.5, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 16}, FData: 1.0, IsFloat: true})
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 16}, FData: 1.0, IsFloat: true})
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
print(a);
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
print(x);
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
print(x);
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
print(x);
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
print(x);
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
print(x);
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
print(x);
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
print(x);
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
print(x);
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
print(a);
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
print(a);
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
print(a);
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
print(a);
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
	i.env.Set("x", Value{FType: ast.FloatType{Size: 32}, FData: 1.0, IsFloat: true})
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
print(x);`
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
print(x);`
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
print(x);`
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
print(x);`
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
print(x);`
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
print(x);`
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
print(x);`
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
	input := `print(true);
print(false);`
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
		{Value{BType: ast.BoolType{}, BData: true, IsBool: true}, "bool(true)"},
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
print(x);`
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
print(b);`
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
print(b);`
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
	i.env.Set("x", Value{BType: ast.BoolType{Nullable: false}, BData: true, IsBool: true})
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
	i.env.Set("x", Value{BType: ast.BoolType{Nullable: false}, BData: true, IsBool: true})
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
	input := "print(int);"
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
	if output != "64-bit signed int" {
		t.Errorf("expected '64-bit signed int', got %q", output)
	}
}

func TestPrintTypeRefFloat(t *testing.T) {
	input := "print(float);"
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
	if output != "64-bit float" {
		t.Errorf("expected '64-bit float', got %q", output)
	}
}

func TestPrintTypeRefBool(t *testing.T) {
	input := "print(bool);"
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
	if output != "bool" {
		t.Errorf("expected 'bool', got %q", output)
	}
}

func TestIntMinMaxDefault(t *testing.T) {
	input := `
print(int.min);
print(int.max);
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
print(int{size: 8}.min);
print(int{size: 8}.max);
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
print(int{size: 8, signed: false}.min);
print(int{size: 8, signed: false}.max);
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
print(float.min);
print(float.max);
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
print(float.min_subnormal);
print(float.min_normal);
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
print(float.precision);
print(float.min_exponent);
print(float.max_exponent);
print(float.size);
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
print(float{size: 32}.min);
print(float{size: 32}.max);
print(float{size: 32}.min_subnormal);
print(float{size: 32}.min_normal);
print(float{size: 32}.precision);
print(float{size: 32}.min_exponent);
print(float{size: 32}.max_exponent);
print(float{size: 32}.size);
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

func TestVarDotType(t *testing.T) {
	input := `
var a: int{size: 8} = 42;
print(a.type);
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
	if output != "8-bit signed int" {
		t.Errorf("expected '8-bit signed int', got %q", output)
	}
}

func TestVarDotTypeDotMin(t *testing.T) {
	input := `
var a: int{size: 8} = 42;
print(a.type.min);
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

func TestVarFloatDotType(t *testing.T) {
	input := `
var a: float{size: 32} = 3.14;
print(a.type);
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
	if output != "32-bit float" {
		t.Errorf("expected '32-bit float', got %q", output)
	}
}

func TestVarFloatDotTypeDotMax(t *testing.T) {
	input := `
var a: float{size: 32} = 3.14;
print(a.type.max);
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
	input := "print(-int{size: 8}.min);"
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 16, Signed: true}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 16, Signed: true}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 32, Signed: true}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 32, Signed: true}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 16, Signed: false}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 16, Signed: false}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 32, Signed: false}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 64, Signed: false}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 16}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 32}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 32}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 8, Signed: true}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 32, Signed: false}},
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
	expr := &ast.TypeRef{Kind: "bool", BType: ast.BoolType{Nullable: false}}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsType || !val.IsBool {
		t.Errorf("expected IsType and IsBool")
	}
}

func TestBoolTypeMemberType(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Kind: "bool", BType: ast.BoolType{Nullable: false}},
		Member: "type",
	}
	val, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !val.IsType || !val.IsBool {
		t.Errorf("expected IsType and IsBool")
	}
}

func TestBoolSize(t *testing.T) {
	i := New()
	expr := &ast.MemberAccess{
		Object: &ast.TypeRef{Kind: "bool", BType: ast.BoolType{Nullable: false}},
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
	input := "print(bool.size);"
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
	input := "print(bool);"
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
	if output != "bool" {
		t.Errorf("expected 'bool', got %q", output)
	}
}

func TestValueDotTypeOnTypedVar(t *testing.T) {
	input := `
var a: int{size: 32} = 100;
print(a.type.min);
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 64}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 64}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 64}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 64}},
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
		Object: &ast.TypeRef{Kind: "float", FType: ast.FloatType{Size: 64}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 64, Signed: false}},
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
	input := "print(int{size: 64, signed: false}.max);"
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
	input := "print(int{size: 64, signed: false}.min);"
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
	if result != "32-bit signed int" {
		t.Errorf("expected '32-bit signed int', got %q", result)
	}

	v2 := Value{IsType: true, IsFloat: true, FType: ast.FloatType{Size: 64, Nullable: false}}
	result2 := v2.String()
	if result2 != "64-bit float" {
		t.Errorf("expected '64-bit float', got %q", result2)
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
		Object: &ast.TypeRef{Kind: "bool", BType: ast.BoolType{Nullable: false}},
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
		Object: &ast.TypeRef{Kind: "int", IType: ast.IntegerType{Size: 64, Signed: true}},
		Member: "nonexistent",
	}
	_, err := i.evalExpr(expr)
	if err == nil {
		t.Errorf("expected error for nonexistent type member")
	}
}

func TestValueMemberError(t *testing.T) {
	i := New()
	i.env.Set("x", Value{Untyped: true, Data: 42})
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
	input := "print(1 == 1);"
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
	input := "print(2 > 1);"
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
	input := "print(1.5 < 2.5);"
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
	input := "print(true && false);"
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
	input := "print(true || false);"
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
	input := "print(!true);"
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
	input := "print(!!true);"
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
	input := "print(true || false && false);"
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
	input := "print(1 + 2 < 3);"
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
	input := "print(1 != 2);"
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
	i.env.Set("a", Value{IType: ast.IntegerType{Size: 64, Signed: true}, Data: 5})
	i.env.Set("b", Value{FType: ast.FloatType{Size: 64}, FData: 5.0, IsFloat: true})
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
	i.env.Set("a", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 5})
	i.env.Set("b", Value{FType: ast.FloatType{Size: 64}, FData: 5.0, IsFloat: true})
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
	i.env.Set("a", Value{IType: ast.IntegerType{Size: 64, Signed: true}, Data: 5})
	i.env.Set("b", Value{FType: ast.FloatType{Size: 64}, FData: 10.0, IsFloat: true})
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
	i.env.Set("a", Value{Untyped: true, BData: true, IsBool: true})
	i.env.Set("b", Value{Untyped: true, Data: 1})
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
print(a == b);
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
		{"print(null == null);", "true"},
		{"print(null != null);", "false"},
		{"print(null == 1);", "false"},
		{"print(null != 1);", "true"},
		{"print(1 == null);", "false"},
		{"print(1 != null);", "true"},
		{"print(null == 1.0);", "false"},
		{"print(null != 1.0);", "true"},
		{"print(null < 1);", "false"},
		{"print(null > 1);", "false"},
		{"print(null <= 1);", "false"},
		{"print(null >= 1);", "false"},
		{"print(1 < null);", "false"},
		{"print(1 > null);", "false"},
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
		{"print(null && true);", "false"},
		{"print(null && false);", "false"},
		{"print(true && null);", "false"},
		{"print(false && null);", "false"},
		{"print(null || true);", "true"},
		{"print(null || false);", "false"},
		{"print(true || null);", "true"},
		{"print(false || null);", "false"},
		{"print(!null);", "true"},
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
		{"print(NaN == NaN);", "false"},
		{"print(NaN != NaN);", "true"},
		{"print(NaN < 0);", "false"},
		{"print(NaN > 0);", "false"},
		{"print(NaN <= 0);", "false"},
		{"print(NaN >= 0);", "false"},
		{"print(0 < NaN);", "false"},
		{"print(0 > NaN);", "false"},
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
		{"print(infinity == infinity);", "true"},
		{"print(-infinity < 0);", "true"},
		{"print(-infinity > 0);", "false"},
		{"print(infinity > 1000);", "true"},
		{"print(-infinity < -1000);", "true"},
		{"print(infinity == 1.0);", "false"},
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
print(1.0 / 0.0);
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
		{"print(!!true);", "true"},
		{"print(!!false);", "false"},
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
		"print(true < false);",
		"print(true > false);",
		"print(true <= false);",
		"print(true >= false);",
		"print(1 < true);",
		"print(true < 1);",
		"print(1.0 < true);",
		"print(true < 1.0);",
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
		{"print(false && true);", "false"},
		{"print(true || false);", "true"},
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
print(x);
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
print(x);
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
print(x);
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
print(x);
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
	print(0);
} else if (x == 1) {
	print(1);
} else {
	print(2);
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
	print(1);
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
print(x);
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
print(b);
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
	print(a);
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
print(x);
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
print(b);
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
print(c);
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
	print(1);
} else {
	print(2);
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
	print(1);
} else {
	print(2);
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
	print(1);
}
print(2);
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
	print(1);
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
	print(undefinedVar);
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
	i.env.Set("x", Value{IType: ast.IntegerType{Size: 32, Signed: true}, Data: 1})
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
	print(x);
}
print(x);
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
print(1);
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
	print(1);
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
	print(1);
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
	print(0);
} else if (x == 1) {
	print(1);
} else if (x == 2) {
	print(2);
} else if (x == 3) {
	print(3);
} else {
	print(4);
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
	print(0);
} else if (x == 1) {
	print(1);
} else if (x == 2) {
	print(2);
}
print(3);
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
print(b);
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
	print(1);
} else if (true) {
	print(2);
} else {
	print(3);
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
	print(1);
} else {
	print(2);
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
print(x);
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
	print(0);
} else if (x == 5) {
	print(5);
} else {
	print(99);
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
	print(0);
} else if (x == 0) {
	print(99);
} else if (x == 0) {
	print(99);
} else {
	print(99);
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
	print(1);
} else {
	print(2);
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
	print(0);
} else if (x == 1) {
	x = 2;
} else if (x == 2) {
	print(2);
} else {
	print(3);
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
	print(i);
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
	print(i);
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
	print(i);
}
print(i);
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
	print(i);
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
	print(i);
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
	print(i);
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
	print(i);
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
	print(i);
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
		print(i * 10 + j);
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
print(1);
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
	print(x);
}
print(x);
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
	print(i);
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
print(i);
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
print(x);
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
	print(i);
}
print(i);
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
	print(i);
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
	print(i);
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
	print(i);
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
	print(i);
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
		print(i * 10 + j);
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
print(x);
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
print(i);
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
print(i);
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
print(f);
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
print(f);
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
	print(i);
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
	print(i);
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
print(i);
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
print(i);
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
print(i);
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
print(i);
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
print(i);
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
print(i);
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
