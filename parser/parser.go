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
		if p.peekToken.Type == lexer.TOK_PLUS_PLUS || p.peekToken.Type == lexer.TOK_MINUS_MINUS {
			return p.parseIncDec()
		}
		if p.peekToken.Type == lexer.TOK_DOT {
			expr := p.parsePostfix()
			if p.curToken.Type != lexer.TOK_SEMICOLON {
				p.addError("expected ';'")
				return nil
			}
			return &ast.ExprStmt{Expr: expr}
		}
		if p.peekToken.Type == lexer.TOK_LBRACKET {
			return p.parseIndexedAssign()
		}
		return p.parseAssignment()
	case lexer.TOK_IF:
		return p.parseIf()
	case lexer.TOK_FOR:
		return p.parseFor()
	case lexer.TOK_WHILE:
		return p.parseWhile()
	case lexer.TOK_BREAK:
		return p.parseBreak()
	case lexer.TOK_SKIP:
		return p.parseSkip()
	case lexer.TOK_TYPEOF:
		expr := p.parseExpr()
		if p.curToken.Type != lexer.TOK_SEMICOLON {
			p.addError("expected ';'")
			return nil
		}
		return &ast.ExprStmt{Expr: expr}
	default:
		p.addError("unexpected token: %s", p.curToken.Literal)
		return nil
	}
}

func (p *Parser) parseType() ast.Type {
	switch p.curToken.Type {
	case lexer.TOK_INT:
		return p.parseIntegerType()
	case lexer.TOK_FLOAT:
		return p.parseFloatType()
	case lexer.TOK_BOOL:
		return p.parseBoolType()
	case lexer.TOK_ARRAY:
		return p.parseArrayType()
	case lexer.TOK_LIST:
		return p.parseListType()
	default:
		p.addError("expected type, got %s", p.curToken.Type)
		return ast.IntegerType{Size: 64, Signed: true, Nullable: true}
	}
}

func (p *Parser) parseTypeParams(validKeys map[string]bool) map[string]interface{} {
	params := make(map[string]interface{})
	if p.curToken.Type != lexer.TOK_LBRACE {
		return params
	}
	for {
		p.nextToken()
		if p.curToken.Type == lexer.TOK_RBRACE {
			p.nextToken()
			break
		}
		if p.curToken.Type != lexer.TOK_SIZE && p.curToken.Type != lexer.TOK_SIGNED &&
			p.curToken.Type != lexer.TOK_NULLABLE && p.curToken.Type != lexer.TOK_MIN &&
			p.curToken.Type != lexer.TOK_MAX {
			p.addError("expected type parameter, got %s", p.curToken.Type)
			break
		}
		key := p.curToken.Literal
		if !validKeys[key] {
			p.addError("unexpected parameter: %s", key)
			break
		}

		p.nextToken()
		if p.curToken.Type != lexer.TOK_COLON {
			p.addError("expected ':', got %s", p.curToken.Type)
			break
		}

		p.nextToken()
		switch key {
		case "size", "min", "max":
			if p.curToken.Type == lexer.TOK_AUTO {
				params[key] = "auto"
			} else if p.curToken.Type == lexer.TOK_INT_LIT {
				val, err := strconv.Atoi(p.curToken.Literal)
				if err != nil {
					p.addError("invalid integer: %s", p.curToken.Literal)
					break
				}
				params[key] = val
			} else {
				p.addError("expected integer or 'auto', got %s", p.curToken.Type)
				break
			}
		case "signed", "nullable":
			if p.curToken.Type != lexer.TOK_TRUE && p.curToken.Type != lexer.TOK_FALSE {
				p.addError("expected 'true' or 'false', got %s", p.curToken.Type)
				break
			}
			params[key] = p.curToken.Type == lexer.TOK_TRUE
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
	return params
}

func forceTypeNotNullable(t ast.Type) ast.Type {
	switch tt := t.(type) {
	case ast.IntegerType:
		tt.Nullable = false
		return tt
	case ast.FloatType:
		tt.Nullable = false
		return tt
	case ast.BoolType:
		tt.Nullable = false
		return tt
	}
	return t
}

func (p *Parser) parseArrayType() ast.ArrayType {
	p.nextToken()

	params := p.parseTypeParams(map[string]bool{"size": true})

	size := 0
	if v, ok := params["size"]; ok {
		if s, ok := v.(int); ok {
			size = s
		}
	}

	// Support bracket syntax: array[5]<int>
	if p.curToken.Type == lexer.TOK_LBRACKET && len(params) == 0 {
		p.nextToken()
		if p.curToken.Type == lexer.TOK_INT_LIT {
			val, err := strconv.Atoi(p.curToken.Literal)
			if err != nil {
				p.addError("invalid array size: %s", p.curToken.Literal)
			} else {
				size = val
			}
		} else {
			p.addError("expected integer array size, got %s", p.curToken.Type)
		}
		p.nextToken()
		if p.curToken.Type != lexer.TOK_RBRACKET {
			p.addError("expected ']'")
		}
		p.nextToken()
	}

	if p.curToken.Type != lexer.TOK_LT {
		p.addError("expected '<' for array element type, got %s", p.curToken.Type)
		return ast.ArrayType{ElemType: ast.IntegerType{Size: 64, Signed: true, Nullable: true}, Size: size}
	}
	p.nextToken()

	elemType := p.parseType()
	elemType = forceTypeNotNullable(elemType)

	if p.curToken.Type != lexer.TOK_GT {
		p.addError("expected '>' after array element type, got %s", p.curToken.Type)
		return ast.ArrayType{ElemType: elemType, Size: size}
	}
	p.nextToken()

	return ast.ArrayType{ElemType: elemType, Size: size}
}

func (p *Parser) parseListType() ast.ListType {
	p.nextToken()

	params := p.parseTypeParams(map[string]bool{"min": true, "max": true})

	lt := ast.ListType{ElemType: ast.IntegerType{Size: 64, Signed: true, Nullable: true}}

	if v, ok := params["min"]; ok {
		if s, ok := v.(int); ok {
			lt.HasMin = true
			lt.MinSize = s
		}
	}
	if v, ok := params["max"]; ok {
		if s, ok := v.(int); ok {
			lt.HasMax = true
			lt.MaxSize = s
		}
	}

	if p.curToken.Type != lexer.TOK_LT {
		p.addError("expected '<' for list element type, got %s", p.curToken.Type)
		return lt
	}
	p.nextToken()

	elemType := p.parseType()
	elemType = forceTypeNotNullable(elemType)

	if p.curToken.Type != lexer.TOK_GT {
		p.addError("expected '>' after list element type, got %s", p.curToken.Type)
		return ast.ListType{ElemType: elemType, HasMin: lt.HasMin, MinSize: lt.MinSize, HasMax: lt.HasMax, MaxSize: lt.MaxSize}
	}
	p.nextToken()

	lt.ElemType = elemType
	return lt
}

func (p *Parser) parseFloatType() ast.FloatType {
	fType := ast.FloatType{Size: 64, Nullable: true}

	p.nextToken()
	params := p.parseTypeParams(map[string]bool{"size": true, "nullable": true})

	if v, ok := params["size"]; ok {
		if s, ok := v.(int); ok && (s == 16 || s == 32 || s == 64) {
			fType.Size = s
		} else if ok {
			p.addError("invalid size: %d, must be 16, 32, or 64", v)
		}
	}
	if v, ok := params["nullable"]; ok {
		if b, ok := v.(bool); ok {
			fType.Nullable = b
		}
	}

	return fType
}

func (p *Parser) parseBoolType() ast.BoolType {
	bType := ast.BoolType{Nullable: true}

	p.nextToken()
	params := p.parseTypeParams(map[string]bool{"nullable": true})

	if v, ok := params["nullable"]; ok {
		if b, ok := v.(bool); ok {
			bType.Nullable = b
		}
	}

	return bType
}

func (p *Parser) parseIntegerType() ast.IntegerType {
	iType := ast.IntegerType{Size: 64, Signed: true, Nullable: true}

	p.nextToken()
	params := p.parseTypeParams(map[string]bool{"size": true, "signed": true, "nullable": true})

	if v, ok := params["size"]; ok {
		if s, ok := v.(int); ok && (s == 8 || s == 16 || s == 32 || s == 64) {
			iType.Size = s
		} else if ok {
			p.addError("invalid size: %d, must be 8, 16, 32, or 64", v)
		}
	}
	if v, ok := params["signed"]; ok {
		if b, ok := v.(bool); ok {
			iType.Signed = b
		}
	}
	if v, ok := params["nullable"]; ok {
		if b, ok := v.(bool); ok {
			iType.Nullable = b
		}
	}

	return iType
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

	t := p.parseType()

	var iType ast.IntegerType
	var fType ast.FloatType
	var bType ast.BoolType
	isFloat := false
	isBool := false

	switch typ := t.(type) {
	case ast.IntegerType:
		iType = typ
	case ast.FloatType:
		fType = typ
		isFloat = true
	case ast.BoolType:
		bType = typ
		isBool = true
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

	return &ast.VarDecl{Name: name, Type: t, IType: iType, FType: fType, BType: bType, Expr: expr, IsFloat: isFloat, IsBool: isBool}
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

func (p *Parser) parseIndexedAssign() *ast.Assignment {
	name := p.curToken.Literal
	p.nextToken()
	p.nextToken()
	index := p.parseExpr()
	if p.curToken.Type != lexer.TOK_RBRACKET {
		p.addError("expected ']'")
		return nil
	}
	p.nextToken()
	if !isAssignOp(p.curToken.Type) {
		p.addError("expected assignment operator, got %s", p.curToken.Type)
		return nil
	}
	op := p.curToken.Literal
	p.nextToken()
	expr := p.parseExpr()
	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';'")
		return nil
	}
	return &ast.Assignment{Name: name, Index: index, Op: op, Expr: expr}
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

func (p *Parser) parseBlock() *ast.BlockStmt {
	block := &ast.BlockStmt{}

	if p.curToken.Type == lexer.TOK_LBRACE {
		p.nextToken()

		for p.curToken.Type != lexer.TOK_RBRACE && p.curToken.Type != lexer.TOK_EOF {
			stmt := p.parseStmt()
			if stmt != nil {
				block.Stmts = append(block.Stmts, stmt)
			}
			p.nextToken()
		}

		if p.curToken.Type == lexer.TOK_EOF {
			p.addError("expected '}'")
			return nil
		}

		return block
	}

	stmt := p.parseStmt()
	if stmt != nil {
		block.Stmts = append(block.Stmts, stmt)
	}
	return block
}

func (p *Parser) parseIf() *ast.IfStmt {
	p.nextToken()

	if p.curToken.Type != lexer.TOK_LPAREN {
		p.addError("expected '(', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	condition := p.parseExpr()

	if p.curToken.Type != lexer.TOK_RPAREN {
		p.addError("expected ')', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	thenBlock := p.parseBlock()
	if thenBlock == nil {
		return nil
	}

	var elseStmt ast.Stmt
	if p.peekToken.Type == lexer.TOK_ELSE {
		p.nextToken()
		p.nextToken()

		if p.curToken.Type == lexer.TOK_IF {
			elseStmt = p.parseIf()
		} else {
			elseStmt = p.parseBlock()
		}
	}

	return &ast.IfStmt{Condition: condition, Then: thenBlock, Else: elseStmt}
}

func (p *Parser) parseFor() *ast.ForStmt {
	p.nextToken()

	if p.curToken.Type != lexer.TOK_LPAREN {
		p.addError("expected '(', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	var init ast.Stmt
	if p.curToken.Type != lexer.TOK_SEMICOLON {
		if p.curToken.Type == lexer.TOK_VAR {
			init = p.parseVarDecl()
		} else if p.curToken.Type == lexer.TOK_IDENT {
			init = p.parseAssignment()
		} else {
			p.addError("expected variable declaration or assignment in for init, got %s", p.curToken.Type)
			return nil
		}
	}
	p.nextToken()

	var cond ast.Expr
	if p.curToken.Type != lexer.TOK_SEMICOLON {
		cond = p.parseExpr()
	}
	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';' after for condition")
		return nil
	}
	p.nextToken()

	var update ast.Stmt
	if p.curToken.Type != lexer.TOK_RPAREN {
		if p.curToken.Type != lexer.TOK_IDENT {
			p.addError("expected identifier in for update, got %s", p.curToken.Type)
			return nil
		}
		name := p.curToken.Literal
		p.nextToken()

		if p.curToken.Type == lexer.TOK_PLUS_PLUS || p.curToken.Type == lexer.TOK_MINUS_MINUS {
			op := p.curToken.Literal
			p.nextToken()
			update = &ast.IncDecStmt{Name: name, Op: op}
		} else {
			var op string
			switch p.curToken.Type {
			case lexer.TOK_ASSIGN:
				op = "="
			case lexer.TOK_PLUS_EQ:
				op = "+="
			case lexer.TOK_MINUS_EQ:
				op = "-="
			case lexer.TOK_STAR_EQ:
				op = "*="
			case lexer.TOK_SLASH_EQ:
				op = "/="
			default:
				p.addError("expected assignment operator or '++'/'--' in for update, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken()

			expr := p.parseExpr()
			update = &ast.Assignment{Name: name, Op: op, Expr: expr}
		}
	}

	if p.curToken.Type != lexer.TOK_RPAREN {
		p.addError("expected ')' after for update, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.ForStmt{Init: init, Condition: cond, Update: update, Body: body}
}

func (p *Parser) parseWhile() *ast.WhileStmt {
	p.nextToken()

	if p.curToken.Type != lexer.TOK_LPAREN {
		p.addError("expected '(', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	cond := p.parseExpr()

	if p.curToken.Type != lexer.TOK_RPAREN {
		p.addError("expected ')', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.WhileStmt{Condition: cond, Body: body}
}

func (p *Parser) parseBreak() *ast.BreakStmt {
	p.nextToken()

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.BreakStmt{}
}

func (p *Parser) parseSkip() *ast.SkipStmt {
	p.nextToken()

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.SkipStmt{}
}

func (p *Parser) parseIncDec() *ast.IncDecStmt {
	ident := p.curToken.Literal
	p.nextToken()
	op := p.curToken.Literal
	p.nextToken()

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.IncDecStmt{Name: ident, Op: op}
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
		if _, ok := left.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean || operator")
		}
		if _, ok := right.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean || operator")
		}
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
		if _, ok := left.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean && operator")
		}
		if _, ok := right.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean && operator")
		}
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
		if _, ok := left.(*ast.NullLit); ok {
			p.addWarning("using null literal with comparison operator %s", op)
		}
		if _, ok := right.(*ast.NullLit); ok {
			p.addWarning("using null literal with comparison operator %s", op)
		}
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
		if p.peekToken.Type == lexer.TOK_MINUS {
			p.addError("unary minus cannot be applied to a minus expression")
			return nil
		}
		p.nextToken()
		expr := p.parsePostfix()
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
		if _, ok := expr.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean ! operator")
		}
		return &ast.UnaryExpr{Op: "!", Right: expr}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Expr {
	left := p.parsePrimary()
	for {
		switch p.curToken.Type {
		case lexer.TOK_LBRACKET:
			p.nextToken()
			index := p.parseExpr()
			if p.curToken.Type != lexer.TOK_RBRACKET {
				p.addError("expected ']'")
				return left
			}
			p.nextToken()
			left = &ast.IndexExpr{Object: left, Index: index}
		case lexer.TOK_DOT:
			p.nextToken()
			if !isMemberToken(p.curToken.Type) {
				p.addError("expected member name after '.'")
				break
			}
			name := p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == lexer.TOK_LPAREN {
				p.nextToken()
				var args []ast.Expr
				if p.curToken.Type != lexer.TOK_RPAREN {
					args = append(args, p.parseExpr())
					for p.curToken.Type == lexer.TOK_COMMA {
						p.nextToken()
						args = append(args, p.parseExpr())
					}
				}
				if p.curToken.Type != lexer.TOK_RPAREN {
					p.addError("expected ')' after arguments")
					break
				}
				p.nextToken()
				left = &ast.MemberAccess{Object: left, Member: name, Args: args}
			} else {
				left = &ast.MemberAccess{Object: left, Member: name}
			}
		default:
			return left
		}
	}
}

func isMemberToken(typ lexer.TokenType) bool {
	return typ == lexer.TOK_IDENT || typ == lexer.TOK_SIZE ||
		typ == lexer.TOK_SIGNED || typ == lexer.TOK_NULLABLE ||
		typ == lexer.TOK_MIN || typ == lexer.TOK_MAX
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
	case lexer.TOK_LBRACKET:
		return p.parseArrayLit()
	case lexer.TOK_INT:
		t := p.parseIntegerType()
		return &ast.TypeRef{Type: t, IsType: true}
	case lexer.TOK_FLOAT:
		t := p.parseFloatType()
		return &ast.TypeRef{Type: t, IsType: true}
	case lexer.TOK_BOOL:
		t := p.parseBoolType()
		return &ast.TypeRef{Type: t, IsType: true}
	case lexer.TOK_ARRAY:
		t := p.parseArrayType()
		return &ast.TypeRef{Type: t, IsType: true}
	case lexer.TOK_LIST:
		t := p.parseListType()
		return &ast.TypeRef{Type: t, IsType: true}
	case lexer.TOK_TYPEOF:
		if p.peekToken.Type != lexer.TOK_LPAREN {
			p.addError("expected '(' after 'typeof'")
			return nil
		}
		p.nextToken()
		p.nextToken()
		expr := p.parseExpr()
		if p.curToken.Type != lexer.TOK_RPAREN {
			p.addError("expected ')' after typeof operand")
			return nil
		}
		p.nextToken()
		return &ast.TypeOfExpr{Expr: expr}
	default:
		p.addError("unexpected token in expression: %s", p.curToken.Literal)
		return nil
	}
}

func (p *Parser) parseArrayLit() ast.Expr {
	p.nextToken()
	var elements []ast.Expr
	if p.curToken.Type != lexer.TOK_RBRACKET {
		elements = append(elements, p.parseExpr())
		for p.curToken.Type == lexer.TOK_COMMA {
			p.nextToken()
			elements = append(elements, p.parseExpr())
		}
	}
	if p.curToken.Type != lexer.TOK_RBRACKET {
		p.addError("expected ']'")
		return &ast.ArrayLit{Elements: elements}
	}
	p.nextToken()
	return &ast.ArrayLit{Elements: elements}
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
