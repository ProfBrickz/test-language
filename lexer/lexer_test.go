package lexer

import "testing"

func TestNextToken(t *testing.T) {
	input := `var x: int{size: 32, signed: true, nullable: false};
print(x);
x += 5;
null;`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOK_VAR, "var"},
		{TOK_IDENT, "x"},
		{TOK_COLON, ":"},
		{TOK_INT, "int"},
		{TOK_LBRACE, "{"},
		{TOK_SIZE, "size"},
		{TOK_COLON, ":"},
		{TOK_INT_LIT, "32"},
		{TOK_COMMA, ","},
		{TOK_SIGNED, "signed"},
		{TOK_COLON, ":"},
		{TOK_TRUE, "true"},
		{TOK_COMMA, ","},
		{TOK_NULLABLE, "nullable"},
		{TOK_COLON, ":"},
		{TOK_FALSE, "false"},
		{TOK_RBRACE, "}"},
		{TOK_SEMICOLON, ";"},
		{TOK_PRINT, "print"},
		{TOK_LPAREN, "("},
		{TOK_IDENT, "x"},
		{TOK_RPAREN, ")"},
		{TOK_SEMICOLON, ";"},
		{TOK_IDENT, "x"},
		{TOK_PLUS_EQ, "+="},
		{TOK_INT_LIT, "5"},
		{TOK_SEMICOLON, ";"},
		{TOK_NULL, "null"},
		{TOK_SEMICOLON, ";"},
		{TOK_EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("test[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestIntegerLiterals(t *testing.T) {
	input := "123 456 789"
	l := New(input)

	expected := []int64{123, 456, 789}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != TOK_INT_LIT {
			t.Errorf("test[%d] - expected INT_LIT, got %s", i, tok.Type)
		}
		if tok.Literal != string(rune('0'+exp%10)) && tok.Literal != "123" && tok.Literal != "456" && tok.Literal != "789" {
			t.Errorf("test[%d] - unexpected literal: %s", i, tok.Literal)
		}
	}
}

func TestIdentifiers(t *testing.T) {
	input := "foo _bar var123"
	l := New(input)

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOK_IDENT, "foo"},
		{TOK_IDENT, "_bar"},
		{TOK_IDENT, "var123"},
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("test[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestBoolKeyword(t *testing.T) {
	input := "bool"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_BOOL {
		t.Errorf("expected BOOL, got %s", tok.Type)
	}
}

func TestKeywords(t *testing.T) {
	input := "var int print size signed nullable null true false bool if else for while break skip array list auto min max typeof"
	l := New(input)

	tests := []struct {
		expectedType TokenType
	}{
		{TOK_VAR},
		{TOK_INT},
		{TOK_PRINT},
		{TOK_SIZE},
		{TOK_SIGNED},
		{TOK_NULLABLE},
		{TOK_NULL},
		{TOK_TRUE},
		{TOK_FALSE},
		{TOK_BOOL},
		{TOK_IF},
		{TOK_ELSE},
		{TOK_FOR},
		{TOK_WHILE},
		{TOK_BREAK},
		{TOK_SKIP},
		{TOK_ARRAY},
		{TOK_LIST},
		{TOK_AUTO},
		{TOK_MIN},
		{TOK_MAX},
		{TOK_TYPEOF},
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("test[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
	}
}

func TestOperators(t *testing.T) {
	input := "+ - * / = += -= *= /= ++ -- [ ]"
	l := New(input)

	tests := []struct {
		expectedType TokenType
	}{
		{TOK_PLUS},
		{TOK_MINUS},
		{TOK_STAR},
		{TOK_SLASH},
		{TOK_ASSIGN},
		{TOK_PLUS_EQ},
		{TOK_MINUS_EQ},
		{TOK_STAR_EQ},
		{TOK_SLASH_EQ},
		{TOK_PLUS_PLUS},
		{TOK_MINUS_MINUS},
		{TOK_LBRACKET},
		{TOK_RBRACKET},
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("test[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
	}
}

func TestLineTracking(t *testing.T) {
	input := "var x;\nprint(x);\n"
	l := New(input)

	// var x;
	tok := l.NextToken()
	if tok.Line != 1 {
		t.Errorf("expected line 1, got %d", tok.Line)
	}

	// consume until semicolon
	for tok.Type != TOK_SEMICOLON {
		tok = l.NextToken()
	}

	// print(x); should be on line 2
	tok = l.NextToken() // print
	if tok.Line != 2 {
		t.Errorf("expected line 2, got %d", tok.Line)
	}
}

func TestComments(t *testing.T) {
	input := "var x; // this is a comment\nprint(x);"
	l := New(input)

	// var
	tok := l.NextToken()
	if tok.Type != TOK_VAR {
		t.Errorf("expected VAR, got %s", tok.Type)
	}

	// consume until print
	for tok.Type != TOK_PRINT {
		tok = l.NextToken()
	}

	if tok.Type != TOK_PRINT {
		t.Errorf("expected PRINT, got %s", tok.Type)
	}
}

func TestFloatKeyword(t *testing.T) {
	input := "float"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_FLOAT {
		t.Errorf("expected FLOAT, got %s", tok.Type)
	}
}

func TestFloatLit(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"3.14", TOK_FLOAT_LIT, "3.14"},
		{"1.5", TOK_FLOAT_LIT, "1.5"},
		{"0.0", TOK_FLOAT_LIT, "0.0"},
		{"123.456", TOK_FLOAT_LIT, "123.456"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("expected %s, got %s", tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("expected literal %q, got %q", tt.expectedLit, tok.Literal)
		}
	}
}

func TestTokenString(t *testing.T) {
	tok := Token{Type: TOK_FLOAT_LIT, Literal: "3.14", Line: 1}
	str := tok.String()
	if str == "" {
		t.Errorf("expected non-empty string")
	}
}

func TestPeekNext(t *testing.T) {
	l := New("a")
	l.NextToken() // consume 'a'
	next := l.peekNext()
	if next != 0 {
		t.Errorf("expected 0 at end of input, got %d", next)
	}

	l2 := New("ab")
	// Don't consume, just peek
	next2 := l2.peekNext()
	if next2 != 'b' {
		t.Errorf("expected 'b' (98), got %d", next2)
	}
}

func TestReadNumberDotWithoutDigit(t *testing.T) {
	// Input with a dot but no following digit - should return int, not float
	input := "123. abc"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT, got %s", tok.Type)
	}
	if tok.Literal != "123" {
		t.Errorf("expected '123', got %q", tok.Literal)
	}

	// Next token should be dot
	tok2 := l.NextToken()
	if tok2.Type != TOK_DOT {
		t.Errorf("expected DOT for '.', got %s", tok2.Type)
	}
}

func TestDefaultCaseNextToken(t *testing.T) {
	input := "@"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT for unknown char, got %s", tok.Type)
	}
	if tok.Literal != "@" {
		t.Errorf("expected literal '@', got %q", tok.Literal)
	}
}

func TestScientificNotation(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"1e20", TOK_FLOAT_LIT, "1e20"},
		{"1e+20", TOK_FLOAT_LIT, "1e+20"},
		{"1e-20", TOK_FLOAT_LIT, "1e-20"},
		{"1.5e3", TOK_FLOAT_LIT, "1.5e3"},
		{"1.5e+3", TOK_FLOAT_LIT, "1.5e+3"},
		{"1.5e-3", TOK_FLOAT_LIT, "1.5e-3"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestUnderscoreInNumbers(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"100_000", TOK_INT_LIT, "100_000"},
		{"1000_1000", TOK_INT_LIT, "1000_1000"},
		{"1_0_0__0_0_0", TOK_INT_LIT, "1_0_0__0_0_0"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestUnderscoreInFloats(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"1_000.5", TOK_FLOAT_LIT, "1_000.5"},
		{"1_000.5e1_0", TOK_FLOAT_LIT, "1_000.5e1_0"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestBinaryLiteral(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"0b0101_0101", TOK_INT_LIT, "0b0101_0101"},
		{"0b1010", TOK_INT_LIT, "0b1010"},
		{"0B1010", TOK_INT_LIT, "0B1010"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestOctalLiteral(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"0o012_345_67", TOK_INT_LIT, "0o012_345_67"},
		{"0o777", TOK_INT_LIT, "0o777"},
		{"0O777", TOK_INT_LIT, "0O777"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestHexLiteral(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"0xffee_d2a5", TOK_INT_LIT, "0xffee_d2a5"},
		{"0xFF", TOK_INT_LIT, "0xFF"},
		{"0XFF", TOK_INT_LIT, "0XFF"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestNumberSequenceWithEOF(t *testing.T) {
	input := "1e20"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_FLOAT_LIT {
		t.Errorf("expected FLOAT_LIT, got %s", tok.Type)
	}
	if tok.Literal != "1e20" {
		t.Errorf("expected '1e20', got %q", tok.Literal)
	}

	tok2 := l.NextToken()
	if tok2.Type != TOK_EOF {
		t.Errorf("expected EOF, got %s", tok2.Type)
	}
}

func TestHexFloatLexer(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"0xf.f", TOK_FLOAT_LIT, "0xf.f"},
		{"0x.1", TOK_FLOAT_LIT, "0x.1"},
		{"0xabc.def", TOK_FLOAT_LIT, "0xabc.def"},
		{"0xf.f+", TOK_FLOAT_LIT, "0xf.f"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestBinFloatLexer(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"0b1.01", TOK_FLOAT_LIT, "0b1.01"},
		{"0b0.1", TOK_FLOAT_LIT, "0b0.1"},
		{"0b1.01+", TOK_FLOAT_LIT, "0b1.01"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestOctFloatLexer(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"0o7.7", TOK_FLOAT_LIT, "0o7.7"},
		{"0o0.4", TOK_FLOAT_LIT, "0o0.4"},
		{"0o7.7+", TOK_FLOAT_LIT, "0o7.7"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestExponentResetOnNoDigits(t *testing.T) {
	l := New("1e")
	tok := l.NextToken()
	if tok.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT, got %s", tok.Type)
	}
	if tok.Literal != "1" {
		t.Errorf("expected '1', got %q", tok.Literal)
	}

	tok2 := l.NextToken()
	if tok2.Type != TOK_IDENT {
		t.Errorf("expected IDENT, got %s", tok2.Type)
	}
	if tok2.Literal != "e" {
		t.Errorf("expected 'e', got %q", tok2.Literal)
	}
}

func TestExponentLoopBreakOnNonDigit(t *testing.T) {
	l := New("1e;")
	tok := l.NextToken()
	if tok.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT, got %s", tok.Type)
	}
	if tok.Literal != "1" {
		t.Errorf("expected '1', got %q", tok.Literal)
	}
}

func TestNaNFloatLiteral(t *testing.T) {
	l := New("NaN")
	tok := l.NextToken()
	if tok.Type != TOK_FLOAT_LIT {
		t.Errorf("expected FLOAT_LIT, got %s", tok.Type)
	}
	if tok.Literal != "NaN" {
		t.Errorf("expected literal %q, got %q", "NaN", tok.Literal)
	}
}

func TestInfinityFloatLiteral(t *testing.T) {
	l := New("infinity")
	tok := l.NextToken()
	if tok.Type != TOK_FLOAT_LIT {
		t.Errorf("expected FLOAT_LIT, got %s", tok.Type)
	}
	if tok.Literal != "infinity" {
		t.Errorf("expected literal %q, got %q", "infinity", tok.Literal)
	}
}

func TestDotToken(t *testing.T) {
	l := New(".")
	tok := l.NextToken()
	if tok.Type != TOK_DOT {
		t.Errorf("expected DOT, got %s", tok.Type)
	}
	if tok.Literal != "." {
		t.Errorf("expected literal '.', got %q", tok.Literal)
	}
}

func TestDotLeadingFloatSemicolon(t *testing.T) {
	// .1e5; should lex as float ".1e5" then semicolon (exponent digits followed by non-digit hits break)
	l := New(".1e5;")
	tok1 := l.NextToken()
	if tok1.Type != TOK_FLOAT_LIT {
		t.Errorf("expected FLOAT_LIT, got %s", tok1.Type)
	}
	if tok1.Literal != ".1e5" {
		t.Errorf("expected literal %q, got %q", ".1e5", tok1.Literal)
	}
	tok2 := l.NextToken()
	if tok2.Type != TOK_SEMICOLON {
		t.Errorf("expected SEMICOLON, got %s", tok2.Type)
	}
}

func TestDotLeadingFloatBacktrack(t *testing.T) {
	// .1e should lex as float ".1" then ident "e" (exponent with no digits backtracks)
	l := New(".1e")
	tok1 := l.NextToken()
	if tok1.Type != TOK_FLOAT_LIT {
		t.Errorf("expected FLOAT_LIT, got %s", tok1.Type)
	}
	if tok1.Literal != ".1" {
		t.Errorf("expected literal %q, got %q", ".1", tok1.Literal)
	}
	tok2 := l.NextToken()
	if tok2.Type != TOK_IDENT {
		t.Errorf("expected IDENT, got %s", tok2.Type)
	}
	if tok2.Literal != "e" {
		t.Errorf("expected literal %q, got %q", "e", tok2.Literal)
	}
}

func TestDotLeadingFloat(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{".1", TOK_FLOAT_LIT, ".1"},
		{".123", TOK_FLOAT_LIT, ".123"},
		{".1e5", TOK_FLOAT_LIT, ".1e5"},
		{".1e-5", TOK_FLOAT_LIT, ".1e-5"},
		{".5_000", TOK_FLOAT_LIT, ".5_000"},
		{".5_000e1_0", TOK_FLOAT_LIT, ".5_000e1_0"},
		{".1e+5", TOK_FLOAT_LIT, ".1e+5"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestConditionTokens(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"==", TOK_EQ, "=="},
		{"!=", TOK_NOT_EQ, "!="},
		{"<", TOK_LT, "<"},
		{">", TOK_GT, ">"},
		{"<=", TOK_LTE, "<="},
		{">=", TOK_GTE, ">="},
		{"&&", TOK_AND, "&&"},
		{"||", TOK_OR, "||"},
		{"!", TOK_NOT, "!"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestConditionTokensInSequence(t *testing.T) {
	input := "== != < > <= >= && || !"
	l := New(input)
	tests := []TokenType{TOK_EQ, TOK_NOT_EQ, TOK_LT, TOK_GT, TOK_LTE, TOK_GTE, TOK_AND, TOK_OR, TOK_NOT}
	for _, expected := range tests {
		tok := l.NextToken()
		if tok.Type != expected {
			t.Errorf("expected %s, got %s", expected, tok.Type)
		}
	}
}

func TestSingleAmpersandAsIntLit(t *testing.T) {
	l := New("&")
	tok := l.NextToken()
	if tok.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT, got %s", tok.Type)
	}
}

func TestSinglePipeAsIntLit(t *testing.T) {
	l := New("|")
	tok := l.NextToken()
	if tok.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT, got %s", tok.Type)
	}
}

func TestPrefixOnlyLiterals(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"0x", TOK_INT_LIT, "0x"},
		{"0X", TOK_INT_LIT, "0X"},
		{"0b", TOK_INT_LIT, "0b"},
		{"0B", TOK_INT_LIT, "0B"},
		{"0o", TOK_INT_LIT, "0o"},
		{"0O", TOK_INT_LIT, "0O"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestStringKeyword(t *testing.T) {
	input := "string"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING {
		t.Errorf("expected STRING, got %s", tok.Type)
	}
}

func TestStringLiteral(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{`"hello"`, TOK_STRING_LIT, "hello"},
		{`""`, TOK_STRING_LIT, ""},
		{`"a b c"`, TOK_STRING_LIT, "a b c"},
		{`"hello\nworld"`, TOK_STRING_LIT, "hello\nworld"},
		{`"tab\there"`, TOK_STRING_LIT, "tab\there"},
		{`"quote\"here"`, TOK_STRING_LIT, `quote"here`},
		{`"back\\slash"`, TOK_STRING_LIT, "back\\slash"},
		{`"null\0char"`, TOK_STRING_LIT, "null\x00char"},
		{`"hex\x41"`, TOK_STRING_LIT, "hexA"},
		{`"unicode\u0041"`, TOK_STRING_LIT, "unicodeA"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestStringLiteralInExpression(t *testing.T) {
	input := `print("hello");`
	l := New(input)
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOK_PRINT, "print"},
		{TOK_LPAREN, "("},
		{TOK_STRING_LIT, "hello"},
		{TOK_RPAREN, ")"},
		{TOK_SEMICOLON, ";"},
		{TOK_EOF, ""},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("test[%d] - expected %s, got %s", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("test[%d] - expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestUnterminatedString(t *testing.T) {
	input := `"hello`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "hello" {
		t.Errorf("expected 'hello', got %q", tok.Literal)
	}
}

func TestStringEscapeBackslashAtEnd(t *testing.T) {
	input := `"hello\`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "hello" {
		t.Errorf("expected 'hello', got %q", tok.Literal)
	}
}

func TestStringEscapeSingleQuote(t *testing.T) {
	tests := []struct {
		input       string
		expectedLit string
	}{
		{`"\'"`, "'"},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != TOK_STRING_LIT {
			t.Errorf("input %q: expected STRING_LIT, got %s", tt.input, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestStringEscapeCarriageReturn(t *testing.T) {
	input := `"hello\rworld"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "hello\rworld" {
		t.Errorf("expected 'hello\\rworld', got %q", tok.Literal)
	}
}

func TestStringEscapeHexAtEnd(t *testing.T) {
	input := `"abc\x`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "abcx" {
		t.Errorf("expected 'abcx', got %q", tok.Literal)
	}
}

func TestStringEscapeHexInvalid(t *testing.T) {
	input := `"abc\xGG"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "abcxG" {
		t.Errorf("expected 'abcxG', got %q", tok.Literal)
	}
}

func TestStringEscape8DigitUnicode(t *testing.T) {
	input := `"\U00000041"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "A" {
		t.Errorf("expected 'A', got %q", tok.Literal)
	}
}

func TestStringEscapeUnrecognized(t *testing.T) {
	input := `"hello\qworld"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "hello\\qworld" {
		t.Errorf("expected 'hello\\\\qworld', got %q", tok.Literal)
	}
}

func TestStringEscapeHexLowerCase(t *testing.T) {
	input := `"\x4a"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "J" {
		t.Errorf("expected 'J', got %q", tok.Literal)
	}
}

func TestStringEscapeUnicodeWithLowerHex(t *testing.T) {
	input := `"\u0061"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "a" {
		t.Errorf("expected 'a', got %q", tok.Literal)
	}
}

func TestReadHexDigitsEOF(t *testing.T) {
	input := `"\uAB`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "\u00ab" {
		t.Errorf("expected '\u00ab', got %q", tok.Literal)
	}
}

func TestReadHexDigitsInvalid(t *testing.T) {
	input := `"\uABGG"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "\u00abGG" {
		t.Errorf("expected '\\u00abGG', got %q", tok.Literal)
	}
}

func TestHexValLowercase(t *testing.T) {
	input := `"\xaf"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "\xaf" {
		t.Errorf("expected '\\xaf', got %q", tok.Literal)
	}
}

func TestHexValUppercase(t *testing.T) {
	input := `"\xAF"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "\xaf" {
		t.Errorf("expected '\\xaf', got %q", tok.Literal)
	}
}

func TestHexValDefault(t *testing.T) {
	input := `"\xG0"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "x0" {
		t.Errorf("expected 'x0', got %q", tok.Literal)
	}
}

func TestStringEscapeUnicodeWithUpperHexDigits(t *testing.T) {
	input := `"\u00AF"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_STRING_LIT {
		t.Errorf("expected STRING_LIT, got %s", tok.Type)
	}
	if tok.Literal != "\u00af" {
		t.Errorf("expected '\\u00af', got %q", tok.Literal)
	}
}

func TestCaseVariantFloatKeywords(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"nan", TOK_IDENT, "nan"},
		{"NAN", TOK_IDENT, "NAN"},
		{"Nan", TOK_IDENT, "Nan"},
		{"inf", TOK_IDENT, "inf"},
		{"INF", TOK_IDENT, "INF"},
		{"Inf", TOK_IDENT, "Inf"},
		{"Infinity", TOK_IDENT, "Infinity"},
		{"INFINITY", TOK_IDENT, "INFINITY"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: expected %s, got %s", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestRefCopyIsKeywords(t *testing.T) {
	input := "ref copy is"
	l := New(input)
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOK_REF, "ref"},
		{TOK_COPY, "copy"},
		{TOK_IS, "is"},
	}
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("test[%d] - expected %s, got %s", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("test[%d] - expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
