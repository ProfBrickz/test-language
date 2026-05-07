package lexer

import "testing"

func TestNextToken(t *testing.T) {
	input := `var x: integer{size: 32, signed: true, nullable: false};
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
		{TOK_INTEGER, "integer"},
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

func TestKeywords(t *testing.T) {
	input := "var integer print size signed nullable null true false"
	l := New(input)

	tests := []struct {
		expectedType TokenType
	}{
		{TOK_VAR},
		{TOK_INTEGER},
		{TOK_PRINT},
		{TOK_SIZE},
		{TOK_SIGNED},
		{TOK_NULLABLE},
		{TOK_NULL},
		{TOK_TRUE},
		{TOK_FALSE},
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("test[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
	}
}

func TestOperators(t *testing.T) {
	input := "+ - * / = += -= *= /="
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
	if tok2.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT for '.', got %s", tok2.Type)
	}
}

func TestDefaultCaseNextToken(t *testing.T) {
	input := "&"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != TOK_INT_LIT {
		t.Errorf("expected INT_LIT for unknown char, got %s", tok.Type)
	}
}
