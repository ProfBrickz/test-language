package parser

import (
	"fmt"
	"strconv"

	"lang-interpreter/ast"
	"lang-interpreter/lexer"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.Token
	peekToken lexer.Token
	errors    []string
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

func (p *Parser) ParseSingleStmt() (ast.Stmt, []string) {
	if p.curToken.Type == lexer.TOK_EOF {
		return nil, nil
	}
	stmt := p.parseStmt()
	errors := p.errors
	p.errors = nil
	p.nextToken()
	return stmt, errors
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
	iType := p.parseIntegerType()

	var expr ast.Expr
	if p.curToken.Type == lexer.TOK_ASSIGN {
		p.nextToken()
		expr = p.parseExpr()
	}

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.VarDecl{Name: name, IType: iType, Expr: expr}
}

func (p *Parser) parseIntegerType() ast.IntegerType {
	iType := ast.IntegerType{Size: 64, Signed: true, Nullable: true}

	if p.curToken.Type != lexer.TOK_INTEGER {
		p.addError("expected 'integer', got %s", p.curToken.Type)
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
	return p.parseAddSub()
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
	left := p.parsePrimary()
	for p.curToken.Type == lexer.TOK_STAR || p.curToken.Type == lexer.TOK_SLASH {
		op := p.curToken.Literal
		p.nextToken()
		right := p.parsePrimary()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parsePrimary() ast.Expr {
	switch p.curToken.Type {
	case lexer.TOK_INT_LIT:
		val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err != nil {
			p.addError("invalid integer literal: %s", p.curToken.Literal)
			return nil
		}
		p.nextToken()
		return &ast.IntegerLit{Value: val, Untyped: true}
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
	case lexer.TOK_NULL:
		p.nextToken()
		return &ast.NullLit{}
	default:
		p.addError("unexpected token in expression: %s", p.curToken.Literal)
		return nil
	}
}

func (p *Parser) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf("line %d: %s", p.curToken.Line, fmt.Sprintf(format, args...))
	p.errors = append(p.errors, msg)
}

func (p *Parser) Errors() []string {
	return p.errors
}
