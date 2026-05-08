package lexer

import (
	"fmt"
	"unicode"
)

type TokenType string

const (
	TOK_BOOL      TokenType = "BOOL"
	TOK_INTEGER   TokenType = "INTEGER"
	TOK_FLOAT     TokenType = "FLOAT"
	TOK_VAR       TokenType = "VAR"
	TOK_PRINT     TokenType = "PRINT"
	TOK_IDENT     TokenType = "IDENT"
	TOK_INT_LIT   TokenType = "INT_LIT"
	TOK_FLOAT_LIT TokenType = "FLOAT_LIT"
	TOK_SIZE      TokenType = "SIZE"
	TOK_SIGNED    TokenType = "SIGNED"
	TOK_NULL      TokenType = "NULL"
	TOK_NULLABLE  TokenType = "NULLABLE"
	TOK_TRUE      TokenType = "TRUE"
	TOK_FALSE     TokenType = "FALSE"

	TOK_COLON     TokenType = ":"
	TOK_SEMICOLON TokenType = ";"
	TOK_COMMA     TokenType = ","
	TOK_ASSIGN    TokenType = "="
	TOK_PLUS_EQ   TokenType = "+="
	TOK_MINUS_EQ  TokenType = "-="
	TOK_STAR_EQ   TokenType = "*="
	TOK_SLASH_EQ  TokenType = "/="
	TOK_PLUS      TokenType = "+"
	TOK_MINUS     TokenType = "-"
	TOK_STAR      TokenType = "*"
	TOK_SLASH     TokenType = "/"

	TOK_LBRACE TokenType = "{"
	TOK_RBRACE TokenType = "}"
	TOK_LPAREN TokenType = "("
	TOK_RPAREN TokenType = ")"

	TOK_EOF TokenType = "EOF"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

type Lexer struct {
	input string
	pos   int
	line  int
}

func New(input string) *Lexer {
	return &Lexer{input: input, line: 1}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{TOK_EOF, "", l.line}
	}

	ch := l.input[l.pos]

	if unicode.IsDigit(rune(ch)) {
		return l.readNumber()
	}

	if unicode.IsLetter(rune(ch)) || ch == '_' {
		return l.readIdentifier()
	}

	switch ch {
	case ':':
		return l.makeToken(TOK_COLON, 1)
	case ';':
		return l.makeToken(TOK_SEMICOLON, 1)
	case ',':
		return l.makeToken(TOK_COMMA, 1)
	case '{':
		return l.makeToken(TOK_LBRACE, 1)
	case '}':
		return l.makeToken(TOK_RBRACE, 1)
	case '(':
		return l.makeToken(TOK_LPAREN, 1)
	case ')':
		return l.makeToken(TOK_RPAREN, 1)
	case '+':
		if l.peekNext() == '=' {
			return l.makeToken(TOK_PLUS_EQ, 2)
		}
		return l.makeToken(TOK_PLUS, 1)
	case '-':
		if l.peekNext() == '=' {
			return l.makeToken(TOK_MINUS_EQ, 2)
		}
		return l.makeToken(TOK_MINUS, 1)
	case '*':
		if l.peekNext() == '=' {
			return l.makeToken(TOK_STAR_EQ, 2)
		}
		return l.makeToken(TOK_STAR, 1)
	case '/':
		if l.peekNext() == '=' {
			return l.makeToken(TOK_SLASH_EQ, 2)
		}
		return l.makeToken(TOK_SLASH, 1)
	case '=':
		return l.makeToken(TOK_ASSIGN, 1)
	default:
		tok := Token{TOK_INT_LIT, string(ch), l.line}
		l.pos++
		return tok
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\n' {
			l.line++
			l.pos++
			continue
		}
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.pos++
			continue
		}
		if ch == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.pos++
			}
			continue
		}
		break
	}
}

func (l *Lexer) readNumber() Token {
	start := l.pos

	if l.input[l.pos] == '0' && l.pos+1 < len(l.input) {
		next := l.input[l.pos+1]
		if next == 'x' || next == 'X' {
			return l.readHexNumber(start)
		}
		if next == 'b' || next == 'B' {
			return l.readBinNumber(start)
		}
		if next == 'o' || next == 'O' {
			return l.readOctNumber(start)
		}
	}

	// Decimal integer or float with optional underscores and scientific notation
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '_' || unicode.IsDigit(rune(ch)) {
			l.pos++
		} else {
			break
		}
	}

	hasDecimal := false
	hasExponent := false

	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		if l.pos+1 < len(l.input) && unicode.IsDigit(rune(l.input[l.pos+1])) {
			hasDecimal = true
			l.pos++
			for l.pos < len(l.input) {
				ch := l.input[l.pos]
				if ch == '_' || unicode.IsDigit(rune(ch)) {
					l.pos++
				} else {
					break
				}
			}
		}
	}

	if l.pos < len(l.input) && (l.input[l.pos] == 'e' || l.input[l.pos] == 'E') {
		expStart := l.pos
		l.pos++
		if l.pos < len(l.input) && (l.input[l.pos] == '+' || l.input[l.pos] == '-') {
			l.pos++
		}
		digitCount := 0
		for l.pos < len(l.input) {
			ch := l.input[l.pos]
			if ch == '_' || unicode.IsDigit(rune(ch)) {
				digitCount++
				l.pos++
			} else {
				break
			}
		}
		if digitCount > 0 {
			hasExponent = true
		} else {
			l.pos = expStart
		}
	}

	lit := l.input[start:l.pos]
	if hasDecimal || hasExponent {
		return Token{TOK_FLOAT_LIT, lit, l.line}
	}
	return Token{TOK_INT_LIT, lit, l.line}
}

func (l *Lexer) readHexNumber(start int) Token {
	l.pos += 2
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '_' || unicode.IsDigit(rune(ch)) || unicode.IsLetter(rune(ch)) {
			l.pos++
		} else {
			break
		}
	}
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.input) {
			ch := l.input[l.pos]
			if ch == '_' || unicode.IsDigit(rune(ch)) || unicode.IsLetter(rune(ch)) {
				l.pos++
			} else {
				break
			}
		}
		return Token{TOK_FLOAT_LIT, l.input[start:l.pos], l.line}
	}
	return Token{TOK_INT_LIT, l.input[start:l.pos], l.line}
}

func (l *Lexer) readBinNumber(start int) Token {
	l.pos += 2
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '_' || unicode.IsDigit(rune(ch)) || unicode.IsLetter(rune(ch)) {
			l.pos++
		} else {
			break
		}
	}
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.input) {
			ch := l.input[l.pos]
			if ch == '_' || unicode.IsDigit(rune(ch)) || unicode.IsLetter(rune(ch)) {
				l.pos++
			} else {
				break
			}
		}
		return Token{TOK_FLOAT_LIT, l.input[start:l.pos], l.line}
	}
	return Token{TOK_INT_LIT, l.input[start:l.pos], l.line}
}

func (l *Lexer) readOctNumber(start int) Token {
	l.pos += 2
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '_' || unicode.IsDigit(rune(ch)) || unicode.IsLetter(rune(ch)) {
			l.pos++
		} else {
			break
		}
	}
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.input) {
			ch := l.input[l.pos]
			if ch == '_' || unicode.IsDigit(rune(ch)) || unicode.IsLetter(rune(ch)) {
				l.pos++
			} else {
				break
			}
		}
		return Token{TOK_FLOAT_LIT, l.input[start:l.pos], l.line}
	}
	return Token{TOK_INT_LIT, l.input[start:l.pos], l.line}
}

func (l *Lexer) readIdentifier() Token {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsLetter(rune(l.input[l.pos])) || unicode.IsDigit(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
		l.pos++
	}
	word := l.input[start:l.pos]
	switch word {
	case "var":
		return Token{TOK_VAR, word, l.line}
	case "bool":
		return Token{TOK_BOOL, word, l.line}
	case "integer":
		return Token{TOK_INTEGER, word, l.line}
	case "float":
		return Token{TOK_FLOAT, word, l.line}
	case "print":
		return Token{TOK_PRINT, word, l.line}
	case "size":
		return Token{TOK_SIZE, word, l.line}
	case "signed":
		return Token{TOK_SIGNED, word, l.line}
	case "nullable":
		return Token{TOK_NULLABLE, word, l.line}
	case "null":
		return Token{TOK_NULL, word, l.line}
	case "true":
		return Token{TOK_TRUE, word, l.line}
	case "false":
		return Token{TOK_FALSE, word, l.line}
	default:
		return Token{TOK_IDENT, word, l.line}
	}
}

func (l *Lexer) peekNext() byte {
	if l.pos+1 < len(l.input) {
		return l.input[l.pos+1]
	}
	return 0
}

func (l *Lexer) makeToken(typ TokenType, length int) Token {
	lit := l.input[l.pos : l.pos+length]
	l.pos += length
	return Token{typ, lit, l.line}
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Line: %d}", t.Type, t.Literal, t.Line)
}
