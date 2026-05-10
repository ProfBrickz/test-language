package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"lang-interpreter/ast"
	"lang-interpreter/lexer"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
	warnings  []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	for p.curToken.Type != lexer.TOK_EOF {
		stmt := p.parseStmt()
		if stmt != nil {
			program.Stmts = append(program.Stmts, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) ParseSingleStmt() (ast.Stmt, []string, []string) {
	if p.curToken.Type == lexer.TOK_EOF {
		return nil, nil, nil
	}
	stmt := p.parseStmt()
	errors := p.errors
	warnings := p.warnings
	p.errors = nil
	p.warnings = nil
	p.nextToken()
	return stmt, errors, warnings
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.curToken.Type {
	case lexer.TOK_VAR:
		return p.parseVarDecl()
	case lexer.TOK_PRINT:
		return p.parsePrint()
	case lexer.TOK_IDENT:
		return p.parseAssignment()
	default:
		p.addError("unexpected token: %s", p.curToken.Literal)
		return nil
	}
}

func (p *Parser) parseVarDecl() *ast.VarDecl {
	p.nextToken()
	if p.curToken.Type != lexer.TOK_IDENT {
		p.addError("expected variable name, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal

	p.nextToken()
	if p.curToken.Type != lexer.TOK_COLON {
		p.addError("expected ':', got %s", p.curToken.Type)
		return nil
	}

	p.nextToken()

	var iType ast.IntegerType
	var fType ast.FloatType
	var bType ast.BoolType
	isFloat := false
	isBool := false

	if p.curToken.Type == lexer.TOK_BOOL {
		bType = p.parseBoolType()
		isBool = true
	} else if p.curToken.Type == lexer.TOK_FLOAT {
		fType = p.parseFloatType()
		isFloat = true
	} else {
		iType = p.parseIntegerType()
	}

	var expr ast.Expr
	if p.curToken.Type == lexer.TOK_ASSIGN {
		p.nextToken()
		expr = p.parseExpr()
	}

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.VarDecl{Name: name, IType: iType, FType: fType, BType: bType, Expr: expr, IsFloat: isFloat, IsBool: isBool}
}

func (p *Parser) parseFloatType() ast.FloatType {
	fType := ast.FloatType{Size: 64, Nullable: true}

	p.nextToken()
	if p.curToken.Type == lexer.TOK_LBRACE {
		for {
			p.nextToken()
			if p.curToken.Type == lexer.TOK_RBRACE {
				p.nextToken()
				break
			}
			if p.curToken.Type != lexer.TOK_SIZE && p.curToken.Type != lexer.TOK_NULLABLE {
				p.addError("expected 'size' or 'nullable', got %s", p.curToken.Type)
				break
			}
			key := p.curToken.Literal

			p.nextToken()
			if p.curToken.Type != lexer.TOK_COLON {
				p.addError("expected ':', got %s", p.curToken.Type)
				break
			}

			p.nextToken()
			switch key {
			case "size":
				if p.curToken.Type != lexer.TOK_INT_LIT {
					p.addError("expected integer literal for size, got %s", p.curToken.Type)
					break
				}
				size, err := strconv.Atoi(p.curToken.Literal)
				if err != nil || (size != 16 && size != 32 && size != 64) {
					p.addError("invalid size: %s, must be 16, 32, or 64", p.curToken.Literal)
					break
				}
				fType.Size = size
			case "nullable":
				if p.curToken.Type != lexer.TOK_TRUE && p.curToken.Type != lexer.TOK_FALSE {
					p.addError("expected 'true' or 'false', got %s", p.curToken.Type)
					break
				}
				fType.Nullable = p.curToken.Type == lexer.TOK_TRUE
			}

			p.nextToken()
			if p.curToken.Type == lexer.TOK_COMMA {
				continue
			}
			if p.curToken.Type == lexer.TOK_RBRACE {
				p.nextToken()
				break
			}
			p.addError("expected ',' or '}', got %s", p.curToken.Type)
			break
		}
	}

	return fType
}

func (p *Parser) parseBoolType() ast.BoolType {
	bType := ast.BoolType{Nullable: true}

	p.nextToken()
	if p.curToken.Type == lexer.TOK_LBRACE {
		for {
			p.nextToken()
			if p.curToken.Type == lexer.TOK_RBRACE {
				p.nextToken()
				break
			}
			if p.curToken.Type != lexer.TOK_NULLABLE {
				p.addError("expected 'nullable', got %s", p.curToken.Type)
				break
			}
			key := p.curToken.Literal

			p.nextToken()
			if p.curToken.Type != lexer.TOK_COLON {
				p.addError("expected ':', got %s", p.curToken.Type)
				break
			}

			p.nextToken()
			switch key {
			case "nullable":
				if p.curToken.Type != lexer.TOK_TRUE && p.curToken.Type != lexer.TOK_FALSE {
					p.addError("expected 'true' or 'false', got %s", p.curToken.Type)
					break
				}
				bType.Nullable = p.curToken.Type == lexer.TOK_TRUE
			}

			p.nextToken()
			if p.curToken.Type == lexer.TOK_COMMA {
				continue
			}
			if p.curToken.Type == lexer.TOK_RBRACE {
				p.nextToken()
				break
			}
			p.addError("expected ',' or '}', got %s", p.curToken.Type)
			break
		}
	}

	return bType
}

func (p *Parser) parseIntegerType() ast.IntegerType {
	iType := ast.IntegerType{Size: 64, Signed: true, Nullable: true}

	if p.curToken.Type != lexer.TOK_INT {
		p.addError("expected 'int', got %s", p.curToken.Type)
		return iType
	}

	p.nextToken()
	if p.curToken.Type == lexer.TOK_LBRACE {
		for {
			p.nextToken()
			if p.curToken.Type == lexer.TOK_RBRACE {
				p.nextToken()
				break
			}
			if p.curToken.Type != lexer.TOK_SIZE && p.curToken.Type != lexer.TOK_SIGNED && p.curToken.Type != lexer.TOK_NULLABLE {
				p.addError("expected 'size', 'signed', or 'nullable', got %s", p.curToken.Type)
				break
			}
			key := p.curToken.Literal

			p.nextToken()
			if p.curToken.Type != lexer.TOK_COLON {
				p.addError("expected ':', got %s", p.curToken.Type)
				break
			}

			p.nextToken()
			switch key {
			case "size":
				if p.curToken.Type != lexer.TOK_INT_LIT {
					p.addError("expected integer literal for size, got %s", p.curToken.Type)
					break
				}
				size, err := strconv.Atoi(p.curToken.Literal)
				if err != nil || (size != 8 && size != 16 && size != 32 && size != 64) {
					p.addError("invalid size: %s, must be 8, 16, 32, or 64", p.curToken.Literal)
					break
				}
				iType.Size = size
			case "signed":
				if p.curToken.Type != lexer.TOK_TRUE && p.curToken.Type != lexer.TOK_FALSE {
					p.addError("expected 'true' or 'false', got %s", p.curToken.Type)
					break
				}
				iType.Signed = p.curToken.Type == lexer.TOK_TRUE
			case "nullable":
				if p.curToken.Type != lexer.TOK_TRUE && p.curToken.Type != lexer.TOK_FALSE {
					p.addError("expected 'true' or 'false', got %s", p.curToken.Type)
					break
				}
				iType.Nullable = p.curToken.Type == lexer.TOK_TRUE
			}

			p.nextToken()
			if p.curToken.Type == lexer.TOK_COMMA {
				continue
			}
			if p.curToken.Type == lexer.TOK_RBRACE {
				p.nextToken()
				break
			}
			p.addError("expected ',' or '}', got %s", p.curToken.Type)
			break
		}
	}

	return iType
}

func (p *Parser) parseAssignment() *ast.Assignment {
	name := p.curToken.Literal

	p.nextToken()
	if !isAssignOp(p.curToken.Type) {
		p.addError("expected assignment operator, got %s", p.curToken.Type)
		return nil
	}
	op := p.curToken.Literal

	p.nextToken()
	expr := p.parseExpr()

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.Assignment{Name: name, Op: op, Expr: expr}
}

func isAssignOp(typ lexer.TokenType) bool {
	return typ == lexer.TOK_ASSIGN || typ == lexer.TOK_PLUS_EQ ||
		typ == lexer.TOK_MINUS_EQ || typ == lexer.TOK_STAR_EQ ||
		typ == lexer.TOK_SLASH_EQ
}

func (p *Parser) parsePrint() *ast.PrintStmt {
	p.nextToken()
	if p.curToken.Type != lexer.TOK_LPAREN {
		p.addError("expected '(', got %s", p.curToken.Type)
		return nil
	}

	p.nextToken()
	expr := p.parseExpr()

	if p.curToken.Type != lexer.TOK_RPAREN {
		p.addError("expected ')', got %s", p.curToken.Type)
		return nil
	}

	p.nextToken()
	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.PrintStmt{Expr: expr}
}

func (p *Parser) parseExpr() ast.Expr {
	return p.parseOr()
}

func (p *Parser) parseOr() ast.Expr {
	left := p.parseAnd()
	for p.curToken.Type == lexer.TOK_OR {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parseAnd()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expr {
	left := p.parseEquality()
	for p.curToken.Type == lexer.TOK_AND {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parseEquality()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseEquality() ast.Expr {
	left := p.parseComparison()
	for p.curToken.Type == lexer.TOK_EQ || p.curToken.Type == lexer.TOK_NOT_EQ {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parseComparison()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() ast.Expr {
	left := p.parseAddSub()
	for p.curToken.Type == lexer.TOK_LT || p.curToken.Type == lexer.TOK_GT ||
		p.curToken.Type == lexer.TOK_LTE || p.curToken.Type == lexer.TOK_GTE {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parseAddSub()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAddSub() ast.Expr {
	left := p.parseMulDiv()
	for p.curToken.Type == lexer.TOK_PLUS || p.curToken.Type == lexer.TOK_MINUS {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parseMulDiv()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseMulDiv() ast.Expr {
	left := p.parseUnary()
	for p.curToken.Type == lexer.TOK_STAR || p.curToken.Type == lexer.TOK_SLASH {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parseUnary()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expr {
	if p.curToken.Type == lexer.TOK_MINUS {
		p.nextToken()
		expr := p.parseMember()
		if lit, ok := expr.(*ast.IntegerLit); ok {
			lit.Value = -lit.Value
			return lit
		}
		if lit, ok := expr.(*ast.FloatLit); ok {
			lit.Value = -lit.Value
			return lit
		}
		return &ast.BinaryExpr{
			Left:  &ast.IntegerLit{Value: 0, Untyped: true},
			Op:    "-",
			Right: expr,
		}
	}
	if p.curToken.Type == lexer.TOK_NOT {
		p.nextToken()
		expr := p.parseUnary()
		return &ast.UnaryExpr{Op: "!", Right: expr}
	}
	return p.parseMember()
}

func (p *Parser) parseMember() ast.Expr {
	left := p.parsePrimary()
	for p.curToken.Type == lexer.TOK_DOT {
		p.nextToken()
		if !isMemberToken(p.curToken.Type) {
			p.addError("expected member name after '.'")
			break
		}
		name := p.curToken.Literal
		p.nextToken()
		left = &ast.MemberAccess{Object: left, Member: name}
	}
	return left
}

func isMemberToken(typ lexer.TokenType) bool {
	return typ == lexer.TOK_IDENT || typ == lexer.TOK_SIZE ||
		typ == lexer.TOK_SIGNED || typ == lexer.TOK_NULLABLE
}

func (p *Parser) parsePrimary() ast.Expr {
	switch p.curToken.Type {
	case lexer.TOK_INT_LIT:
		lit := p.curToken.Literal
		clean := strings.ReplaceAll(lit, "_", "")
		hasPrefix := strings.HasPrefix(lit, "0x") || strings.HasPrefix(lit, "0X") ||
			strings.HasPrefix(lit, "0b") || strings.HasPrefix(lit, "0B") ||
			strings.HasPrefix(lit, "0o") || strings.HasPrefix(lit, "0O")
		var val int64
		var err error
		if hasPrefix {
			val, err = strconv.ParseInt(clean, 0, 64)
		} else {
			val, err = strconv.ParseInt(clean, 10, 64)
			if len(clean) > 1 && clean[0] == '0' {
				p.addWarning("leading zeros in decimal literal: %s", lit)
			}
		}
		if err != nil {
			if strings.HasPrefix(lit, "0o") || strings.HasPrefix(lit, "0O") {
				p.addError("%s is an invalid octal literal", lit)
			} else if strings.HasPrefix(lit, "0b") || strings.HasPrefix(lit, "0B") {
				p.addError("%s is an invalid binary literal", lit)
			} else if strings.HasPrefix(lit, "0x") || strings.HasPrefix(lit, "0X") {
				p.addError("%s is an invalid hexadecimal literal", lit)
			} else {
				p.addError("%s is an invalid int literal", lit)
			}
			return nil
		}
		p.nextToken()
		return &ast.IntegerLit{Value: val, Untyped: true}
	case lexer.TOK_FLOAT_LIT:
		lit := p.curToken.Literal
		clean := strings.ReplaceAll(lit, "_", "")
		var val float64
		var err error
		if clean == "NaN" {
			val = math.NaN()
		} else if clean == "infinity" {
			val = math.Inf(1)
		} else if strings.HasPrefix(clean, "0x") || strings.HasPrefix(clean, "0X") {
			val = p.parsePrefixedFloat(clean[2:], 16)
		} else if strings.HasPrefix(clean, "0b") || strings.HasPrefix(clean, "0B") {
			val = p.parsePrefixedFloat(clean[2:], 2)
		} else if strings.HasPrefix(clean, "0o") || strings.HasPrefix(clean, "0O") {
			val = p.parsePrefixedFloat(clean[2:], 8)
		} else {
			val, err = strconv.ParseFloat(clean, 64)
			if err != nil {
				p.addError("invalid float literal: %s", lit)
				return nil
			}
		}
		p.nextToken()
		return &ast.FloatLit{Value: val, Untyped: true}
	case lexer.TOK_IDENT:
		tok := p.curToken
		p.nextToken()
		return &ast.VarRef{Name: tok.Literal}
	case lexer.TOK_LPAREN:
		p.nextToken()
		expr := p.parseExpr()
		if p.curToken.Type != lexer.TOK_RPAREN {
			p.addError("expected ')', got %s", p.curToken.Type)
			return nil
		}
		p.nextToken()
		return expr
	case lexer.TOK_TRUE:
		p.nextToken()
		return &ast.BoolLit{Value: true, Untyped: true}
	case lexer.TOK_FALSE:
		p.nextToken()
		return &ast.BoolLit{Value: false, Untyped: true}
	case lexer.TOK_NULL:
		p.nextToken()
		return &ast.NullLit{}
	case lexer.TOK_INT:
		iType := p.parseIntegerType()
		return &ast.TypeRef{Kind: "int", IType: iType}
	case lexer.TOK_FLOAT:
		fType := p.parseFloatType()
		return &ast.TypeRef{Kind: "float", FType: fType}
	case lexer.TOK_BOOL:
		bType := p.parseBoolType()
		return &ast.TypeRef{Kind: "bool", BType: bType}
	default:
		p.addError("unexpected token in expression: %s", p.curToken.Literal)
		return nil
	}
}

func (p *Parser) parsePrefixedFloat(s string, base int) float64 {
	parts := strings.SplitN(s, ".", 2)
	var val float64
	if parts[0] != "" {
		intVal, err := strconv.ParseInt(parts[0], base, 64)
		if err != nil {
			p.addError("invalid float literal: %s", p.curToken.Literal)
			return 0
		}
		val = float64(intVal)
	}
	if len(parts) > 1 && parts[1] != "" {
		fracVal, err := strconv.ParseInt(parts[1], base, 64)
		if err != nil {
			p.addError("invalid float literal: %s", p.curToken.Literal)
			return 0
		}
		val += float64(fracVal) / math.Pow(float64(base), float64(len(parts[1])))
	}
	return val
}

func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf("line %d: %s", p.curToken.Line, fmt.Sprintf(format, args...))
	p.errors = append(p.errors, msg)
}

func (p *Parser) addWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf("line %d: %s", p.curToken.Line, fmt.Sprintf(format, args...))
	p.warnings = append(p.warnings, msg)
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) Warnings() []string {
	return p.warnings
}
