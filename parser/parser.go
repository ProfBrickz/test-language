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
			p.nextToken()
		} else if len(p.errors) > 0 {
			p.synchronize()
		} else {
			p.nextToken()
		}
	}
	return program
}

func (p *Parser) synchronize() {
	p.nextToken()
	for p.curToken.Type != lexer.TOK_EOF {
		if p.curToken.Type == lexer.TOK_SEMICOLON {
			p.nextToken()
			return
		}
		switch p.curToken.Type {
		case lexer.TOK_VAR, lexer.TOK_REF, lexer.TOK_PRINT, lexer.TOK_FUNCTION,
			lexer.TOK_RETURN, lexer.TOK_IDENT, lexer.TOK_IF, lexer.TOK_FOR,
			lexer.TOK_WHILE, lexer.TOK_SWITCH, lexer.TOK_BREAK, lexer.TOK_SKIP,
			lexer.TOK_TYPEOF, lexer.TOK_RBRACE:
			return
		}
		p.nextToken()
	}
}

func (p *Parser) synchronizeInBlock() {
	p.nextToken()
	for p.curToken.Type != lexer.TOK_EOF && p.curToken.Type != lexer.TOK_RBRACE {
		if p.curToken.Type == lexer.TOK_SEMICOLON {
			p.nextToken()
			return
		}
		switch p.curToken.Type {
		case lexer.TOK_VAR, lexer.TOK_REF, lexer.TOK_PRINT, lexer.TOK_FUNCTION,
			lexer.TOK_RETURN, lexer.TOK_IDENT, lexer.TOK_IF, lexer.TOK_FOR,
			lexer.TOK_WHILE, lexer.TOK_SWITCH, lexer.TOK_BREAK, lexer.TOK_SKIP,
			lexer.TOK_TYPEOF:
			return
		}
		p.nextToken()
	}
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
	case lexer.TOK_REF:
		return p.parseRefDecl()
	case lexer.TOK_PRINT:
		return p.parsePrint()
	case lexer.TOK_FUNCTION:
		return p.parseFuncDecl()
	case lexer.TOK_RETURN:
		return p.parseReturn()
	case lexer.TOK_IDENT:
		if p.peekToken.Type == lexer.TOK_PLUS_PLUS || p.peekToken.Type == lexer.TOK_MINUS_MINUS {
			return p.parseIncDec()
		}
		if p.peekToken.Type == lexer.TOK_DOT {
			stmtLine := p.curToken.Line
			expr := p.parsePostfix()
			if p.curToken.Type != lexer.TOK_SEMICOLON {
				p.addError("expected ';'")
				return nil
			}
			return &ast.ExprStmt{Expr: expr, Line: stmtLine}
		}
		if p.peekToken.Type == lexer.TOK_LPAREN {
			stmtLine := p.curToken.Line
			expr := p.parsePostfix()
			if p.curToken.Type != lexer.TOK_SEMICOLON {
				p.addError("expected ';'")
				return nil
			}
			return &ast.ExprStmt{Expr: expr, Line: stmtLine}
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
	case lexer.TOK_SWITCH:
		return p.parseSwitch()
	case lexer.TOK_CASE:
		p.addError("case outside switch")
		return nil
	case lexer.TOK_DEFAULT:
		p.addError("default outside switch")
		return nil
	case lexer.TOK_BREAK:
		return p.parseBreak()
	case lexer.TOK_SKIP:
		return p.parseSkip()
	case lexer.TOK_TYPEOF:
		stmtLine := p.curToken.Line
		expr := p.parseExpr()
		if p.curToken.Type != lexer.TOK_SEMICOLON {
			p.addError("expected ';'")
			return nil
		}
		return &ast.ExprStmt{Expr: expr, Line: stmtLine}
	default:
		p.addError("unexpected token: %s", p.curToken.Literal)
		return nil
	}
}

func (p *Parser) parseType() ast.Type {
	types := []ast.Type{p.parseAtomType()}
	for p.curToken.Type == lexer.TOK_PIPE {
		p.nextToken()
		types = append(types, p.parseAtomType())
	}
	if len(types) == 1 {
		return types[0]
	}
	return ast.UnionType{Types: types}
}

func (p *Parser) parseAtomType() ast.Type {
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
	case lexer.TOK_STRING:
		return p.parseStringType()
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
	if p.peekToken.Type != lexer.TOK_RBRACE && !validKeys[p.peekToken.Literal] {
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

	if p.curToken.Type == lexer.TOK_GTE {
		p.curToken = lexer.Token{Type: lexer.TOK_ASSIGN, Literal: "=", Line: p.curToken.Line}
		return ast.ArrayType{ElemType: elemType, Size: size}
	}
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

	if p.curToken.Type == lexer.TOK_GTE {
		p.curToken = lexer.Token{Type: lexer.TOK_ASSIGN, Literal: "=", Line: p.curToken.Line}
		lt.ElemType = elemType
		return lt
	}
	if p.curToken.Type != lexer.TOK_GT {
		p.addError("expected '>' after list element type, got %s", p.curToken.Type)
		return ast.ListType{ElemType: elemType, HasMin: lt.HasMin, MinSize: lt.MinSize, HasMax: lt.HasMax, MaxSize: lt.MaxSize}
	}
	p.nextToken()
	lt.ElemType = elemType
	return lt
}

func (p *Parser) parseStringType() ast.StringType {
	p.nextToken()

	params := p.parseTypeParams(map[string]bool{"size": true, "min": true, "max": true})

	st := ast.StringType{}

	if v, ok := params["size"]; ok {
		if s, ok := v.(int); ok {
			st.Size = s
		}
	}
	if v, ok := params["min"]; ok {
		if s, ok := v.(int); ok {
			st.HasMin = true
			st.MinSize = s
		}
	}
	if v, ok := params["max"]; ok {
		if s, ok := v.(int); ok {
			st.HasMax = true
			st.MaxSize = s
		}
	}

	return st
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
	line := p.curToken.Line
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
	var sType ast.StringType
	var uType ast.UnionType
	isFloat := false
	isBool := false
	isString := false
	isUnion := false

	switch typ := t.(type) {
	case ast.IntegerType:
		iType = typ
	case ast.FloatType:
		fType = typ
		isFloat = true
	case ast.BoolType:
		bType = typ
		isBool = true
	case ast.StringType:
		sType = typ
		isString = true
	case ast.UnionType:
		uType = typ
		isUnion = true
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

	return &ast.VarDecl{Name: name, Type: t, IType: iType, FType: fType, BType: bType, SType: sType, UnionType: uType, Expr: expr, IsFloat: isFloat, IsBool: isBool, IsString: isString, IsUnion: isUnion, Line: line}
}

func (p *Parser) parseRefDecl() *ast.RefDecl {
	line := p.curToken.Line
	p.nextToken()
	if p.curToken.Type != lexer.TOK_IDENT {
		p.addError("expected variable name, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken()

	var declType ast.Type
	if p.curToken.Type == lexer.TOK_COLON {
		p.nextToken()
		declType = p.parseType()
	}

	if p.curToken.Type != lexer.TOK_ASSIGN {
		p.addError("expected '=', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	expr := p.parseExpr()

	if _, ok := expr.(*ast.VarRef); !ok {
		p.addError("ref declaration requires a variable reference on the right side")
		return nil
	}

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.RefDecl{Name: name, Type: declType, Expr: expr, Line: line}
}

func (p *Parser) parseAssignment() *ast.Assignment {
	name := p.curToken.Literal
	line := p.curToken.Line

	p.nextToken()
	if !isAssignOp(p.curToken.Type) {
		p.addError("expected assignment operator, got %s", p.curToken.Type)
		return nil
	}
	op := p.curToken.Literal

	p.nextToken()

	isRef := false
	var expr ast.Expr
	if p.curToken.Type == lexer.TOK_REF && op == "=" {
		isRef = true
		p.nextToken()
		if p.curToken.Type != lexer.TOK_IDENT {
			p.addError("expected variable name after 'ref' in assignment")
			return nil
		}
		tok := p.curToken
		p.nextToken()
		expr = &ast.VarRef{Name: tok.Literal}
	} else {
		expr = p.parseExpr()
	}

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.Assignment{Name: name, Op: op, Expr: expr, IsRef: isRef, Line: line}
}

func (p *Parser) parseIndexedAssign() *ast.Assignment {
	name := p.curToken.Literal
	line := p.curToken.Line
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
	return &ast.Assignment{Name: name, Index: index, Op: op, Expr: expr, Line: line}
}

func isAssignOp(typ lexer.TokenType) bool {
	return typ == lexer.TOK_ASSIGN || typ == lexer.TOK_PLUS_EQ ||
		typ == lexer.TOK_MINUS_EQ || typ == lexer.TOK_STAR_EQ ||
		typ == lexer.TOK_SLASH_EQ || typ == lexer.TOK_MOD_EQ
}

func (p *Parser) parsePrint() *ast.PrintStmt {
	line := p.curToken.Line
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

	return &ast.PrintStmt{Expr: expr, Line: line}
}

func (p *Parser) parseBlock() *ast.BlockStmt {
	block := &ast.BlockStmt{Line: p.curToken.Line}

	if p.curToken.Type == lexer.TOK_LBRACE {
		p.nextToken()

		for p.curToken.Type != lexer.TOK_RBRACE && p.curToken.Type != lexer.TOK_EOF {
			stmt := p.parseStmt()
			if stmt != nil {
				block.Stmts = append(block.Stmts, stmt)
				p.nextToken()
			} else if len(p.errors) > 0 {
				p.synchronizeInBlock()
			} else {
				p.nextToken()
			}
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

func (p *Parser) parseFuncDecl() ast.Stmt {
	line := p.curToken.Line
	p.nextToken()
	if p.curToken.Type != lexer.TOK_IDENT {
		p.addError("expected function name, got %s", p.curToken.Type)
		return nil
	}
	name := p.curToken.Literal
	p.nextToken()

	if p.curToken.Type != lexer.TOK_LPAREN {
		p.addError("expected '(' after function name, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	params := make([]ast.Param, 0)
	if p.curToken.Type != lexer.TOK_RPAREN {
		params = append(params, p.parseParam())
		for p.curToken.Type == lexer.TOK_COMMA {
			p.nextToken()
			params = append(params, p.parseParam())
		}
	}
	if p.curToken.Type != lexer.TOK_RPAREN {
		p.addError("expected ')' after parameters, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	var returnType ast.Type
	if p.curToken.Type == lexer.TOK_COLON {
		p.nextToken()
		returnType = p.parseType()
	}

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.FuncDecl{Name: name, Parameters: params, ReturnType: returnType, Body: body, Line: line}
}

func (p *Parser) parseParam() ast.Param {
	if p.curToken.Type != lexer.TOK_IDENT {
		p.addError("expected parameter name, got %s", p.curToken.Type)
		return ast.Param{}
	}
	name := p.curToken.Literal
	p.nextToken()
	if p.curToken.Type != lexer.TOK_COLON {
		p.addError("expected ':' after parameter name, got %s", p.curToken.Type)
		return ast.Param{}
	}
	p.nextToken()
	typ := p.parseType()
	return ast.Param{Name: name, Type: typ}
}

func (p *Parser) parseReturn() ast.Stmt {
	line := p.curToken.Line
	p.nextToken()

	var value ast.Expr
	if p.curToken.Type != lexer.TOK_SEMICOLON {
		value = p.parseExpr()
	}

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';' after return, got %s", p.curToken.Type)
		return nil
	}

	return &ast.ReturnStmt{Value: value, Line: line}
}

func (p *Parser) parseIf() *ast.IfStmt {
	line := p.curToken.Line
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

	return &ast.IfStmt{Condition: condition, Then: thenBlock, Else: elseStmt, Line: line}
}

func (p *Parser) parseFor() ast.Stmt {
	line := p.curToken.Line
	p.nextToken()

	if p.curToken.Type != lexer.TOK_LPAREN {
		p.addError("expected '(', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	// for-of: for ( var <ident>, <ident> of <expr> ) <body>
	// for-in/for-at: for ( var <ident> in/at <expr> ) <body>
	if p.curToken.Type == lexer.TOK_VAR && p.peekToken.Type == lexer.TOK_IDENT {
		p.nextToken() // consume var
		name1 := p.curToken.Literal
		varLine := p.curToken.Line
		p.nextToken() // consume ident

		// for (var k, v of ...)
		if p.curToken.Type == lexer.TOK_COMMA {
			p.nextToken() // consume comma
			if p.curToken.Type != lexer.TOK_IDENT {
				p.addError("expected variable name after ','")
				return nil
			}
			name2 := p.curToken.Literal
			p.nextToken() // consume second ident

			if p.curToken.Type != lexer.TOK_OF {
				p.addError("expected 'of', got %s", p.curToken.Type)
				return nil
			}
			p.nextToken() // consume of
			iter := p.parseExpr()
			if p.curToken.Type != lexer.TOK_RPAREN {
				p.addError("expected ')'")
				return nil
			}
			p.nextToken()
			body := p.parseBlock()
			if body == nil {
				return nil
			}
			return &ast.ForOfStmt{VarName1: name1, VarName2: name2, Iter: iter, Body: body, Line: line}
		}

		if p.curToken.Type == lexer.TOK_IN || p.curToken.Type == lexer.TOK_AT {
			isIn := p.curToken.Type == lexer.TOK_IN
			p.nextToken()
			iter := p.parseExpr()
			if p.curToken.Type != lexer.TOK_RPAREN {
				p.addError("expected ')'")
				return nil
			}
			p.nextToken()
			body := p.parseBlock()
			if body == nil {
				return nil
			}
			if isIn {
				return &ast.ForInStmt{VarName: name1, Iter: iter, Body: body, Line: line}
			}
			return &ast.ForAtStmt{VarName: name1, Iter: iter, Body: body, Line: line}
		}

		// Not for-in/at. Must be C-style var decl.
		if p.curToken.Type != lexer.TOK_COLON {
			p.addError("expected ':', 'in', or 'at' after variable name, got %s", p.curToken.Type)
			return nil
		}
		p.nextToken()
		t := p.parseType()

		var iType ast.IntegerType
		var fType ast.FloatType
		var bType ast.BoolType
		var sType ast.StringType
		var uType ast.UnionType
		isFloat := false
		isBool := false
		isString := false
		isUnion := false
		switch typ := t.(type) {
		case ast.IntegerType:
			iType = typ
		case ast.FloatType:
			fType = typ
			isFloat = true
		case ast.BoolType:
			bType = typ
			isBool = true
		case ast.StringType:
			sType = typ
			isString = true
		case ast.UnionType:
			uType = typ
			isUnion = true
		}

		var expr ast.Expr
		if p.curToken.Type == lexer.TOK_ASSIGN {
			p.nextToken()
			expr = p.parseExpr()
		}

		if p.curToken.Type != lexer.TOK_SEMICOLON {
			p.addError("expected ';'")
			return nil
		}

		init := &ast.VarDecl{Name: name1, Type: t, IType: iType, FType: fType, BType: bType, SType: sType, UnionType: uType, Expr: expr, IsFloat: isFloat, IsBool: isBool, IsString: isString, IsUnion: isUnion, Line: varLine}
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
			updateName := p.curToken.Literal
			updateLine := p.curToken.Line
			p.nextToken()

			if p.curToken.Type == lexer.TOK_PLUS_PLUS || p.curToken.Type == lexer.TOK_MINUS_MINUS {
				op := p.curToken.Literal
				p.nextToken()
				update = &ast.IncDecStmt{Name: updateName, Op: op, Line: updateLine}
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
				case lexer.TOK_MOD_EQ:
					op = "%="
				default:
					p.addError("expected assignment operator or '++'/'--' in for update, got %s", p.curToken.Type)
					return nil
				}
				p.nextToken()
				updExpr := p.parseExpr()
				update = &ast.Assignment{Name: updateName, Op: op, Expr: updExpr, Line: updateLine}
			}
		}

		if p.curToken.Type != lexer.TOK_RPAREN {
			p.addError("expected ')', got %s", p.curToken.Type)
			return nil
		}
		p.nextToken()

		body := p.parseBlock()
		if body == nil {
			return nil
		}

		return &ast.ForStmt{Init: init, Condition: cond, Update: update, Body: body, Line: line}
	}

	// Original C-style for logic for non-var init
	var init ast.Stmt
	if p.curToken.Type != lexer.TOK_SEMICOLON {
		if p.curToken.Type == lexer.TOK_VAR {
			init = p.parseVarDecl()
		} else if p.curToken.Type == lexer.TOK_REF {
			init = p.parseRefDecl()
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
		updateLine := p.curToken.Line
		p.nextToken()

		if p.curToken.Type == lexer.TOK_PLUS_PLUS || p.curToken.Type == lexer.TOK_MINUS_MINUS {
			op := p.curToken.Literal
			p.nextToken()
			update = &ast.IncDecStmt{Name: name, Op: op, Line: updateLine}
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
			case lexer.TOK_MOD_EQ:
				op = "%="
			default:
				p.addError("expected assignment operator or '++'/'--' in for update, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken()

			expr := p.parseExpr()
			update = &ast.Assignment{Name: name, Op: op, Expr: expr, Line: updateLine}
		}
	}

	if p.curToken.Type != lexer.TOK_RPAREN {
		p.addError("expected ')', got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	body := p.parseBlock()
	if body == nil {
		return nil
	}

	return &ast.ForStmt{Init: init, Condition: cond, Update: update, Body: body, Line: line}
}

func (p *Parser) parseWhile() *ast.WhileStmt {
	line := p.curToken.Line
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

	return &ast.WhileStmt{Condition: cond, Body: body, Line: line}
}

func (p *Parser) parseSwitch() *ast.SwitchStmt {
	line := p.curToken.Line
	p.nextToken()

	if p.curToken.Type != lexer.TOK_LPAREN {
		p.addError("expected '(' after switch, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	value := p.parseExpr()

	if p.curToken.Type != lexer.TOK_RPAREN {
		p.addError("expected ')' after switch expression, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	if p.curToken.Type != lexer.TOK_LBRACE {
		p.addError("expected '{' after switch expression, got %s", p.curToken.Type)
		return nil
	}
	p.nextToken()

	cases := make([]ast.CaseClause, 0)

	for p.curToken.Type != lexer.TOK_RBRACE && p.curToken.Type != lexer.TOK_EOF {
		if p.curToken.Type == lexer.TOK_CASE {
			caseLine := p.curToken.Line
			p.nextToken()

			if p.curToken.Type != lexer.TOK_LPAREN {
				p.addError("expected '(' after case, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken()

			var op string
			var caseExpr ast.Expr

			switch p.curToken.Type {
			case lexer.TOK_EQ:
				op = "=="
				p.nextToken()
				caseExpr = p.parseExpr()
			case lexer.TOK_NOT_EQ:
				op = "!="
				p.nextToken()
				caseExpr = p.parseExpr()
			case lexer.TOK_LT:
				op = "<"
				p.nextToken()
				caseExpr = p.parseExpr()
			case lexer.TOK_GT:
				op = ">"
				p.nextToken()
				caseExpr = p.parseExpr()
			case lexer.TOK_LTE:
				op = "<="
				p.nextToken()
				caseExpr = p.parseExpr()
			case lexer.TOK_GTE:
				op = ">="
				p.nextToken()
				caseExpr = p.parseExpr()
			default:
				op = ""
				caseExpr = p.parseExpr()
			}

			if p.curToken.Type != lexer.TOK_RPAREN {
				p.addError("expected ')' after case expression, got %s", p.curToken.Type)
				return nil
			}
			p.nextToken()

			body := p.parseBlock()
			if body == nil {
				return nil
			}

			cases = append(cases, ast.CaseClause{
				Op:    op,
				Value: caseExpr,
				Body:  body,
				Line:  caseLine,
			})
		} else if p.curToken.Type == lexer.TOK_DEFAULT {
			caseLine := p.curToken.Line
			p.nextToken()

			body := p.parseBlock()
			if body == nil {
				return nil
			}

			cases = append(cases, ast.CaseClause{
				Default: true,
				Body:    body,
				Line:    caseLine,
			})
		} else {
			p.addError("expected case or default in switch, got %s", p.curToken.Type)
			return nil
		}

		p.nextToken()
	}

	if p.curToken.Type == lexer.TOK_EOF {
		p.addError("expected '}' to close switch")
		return nil
	}

	return &ast.SwitchStmt{Value: value, Cases: cases, Line: line}
}

func (p *Parser) parseBreak() *ast.BreakStmt {
	p.nextToken()

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.BreakStmt{Line: p.curToken.Line}
}
func (p *Parser) parseSkip() *ast.SkipStmt {
	p.nextToken()

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.SkipStmt{Line: p.curToken.Line}
}

func (p *Parser) parseIncDec() *ast.IncDecStmt {
	ident := p.curToken.Literal
	identLine := p.curToken.Line
	p.nextToken()
	op := p.curToken.Literal
	p.nextToken()

	if p.curToken.Type != lexer.TOK_SEMICOLON {
		p.addError("expected ';', got %s", p.curToken.Type)
		return nil
	}

	return &ast.IncDecStmt{Name: ident, Op: op, Line: identLine}
}

func (p *Parser) parseExpr() ast.Expr {
	return p.parseOr()
}

func (p *Parser) parseOr() ast.Expr {
	left := p.parseAnd()
	for p.curToken.Type == lexer.TOK_OR {
		op := p.curToken.Literal
		opLine := p.curToken.Line
		p.nextToken()
		right := p.parseAnd()
		if _, ok := left.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean || operator")
		}
		if _, ok := right.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean || operator")
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Line: opLine}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expr {
	left := p.parseEquality()
	for p.curToken.Type == lexer.TOK_AND {
		op := p.curToken.Literal
		opLine := p.curToken.Line
		p.nextToken()
		right := p.parseEquality()
		if _, ok := left.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean && operator")
		}
		if _, ok := right.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean && operator")
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Line: opLine}
	}
	return left
}

func (p *Parser) parseEquality() ast.Expr {
	left := p.parseComparison()
	for p.curToken.Type == lexer.TOK_EQ || p.curToken.Type == lexer.TOK_NOT_EQ ||
		p.curToken.Type == lexer.TOK_IS {
		if p.curToken.Type == lexer.TOK_IS {
			isLine := p.curToken.Line
			p.nextToken()
			var right ast.Expr
			if isTypeToken(p.curToken.Type) {
				t := p.parseType()
				right = &ast.TypeRef{Type: t, IsType: true, Line: isLine}
			} else {
				right = p.parseComparison()
			}
			left = &ast.IsExpr{Left: left, Right: right, Line: isLine}
		} else {
			op := p.curToken.Literal
			opLine := p.curToken.Line
			p.nextToken()
			right := p.parseComparison()
			left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Line: opLine}
		}
	}
	return left
}

func isTypeToken(tok lexer.TokenType) bool {
	return tok == lexer.TOK_INT || tok == lexer.TOK_FLOAT || tok == lexer.TOK_BOOL ||
		tok == lexer.TOK_STRING || tok == lexer.TOK_ARRAY || tok == lexer.TOK_LIST
}

func (p *Parser) parseComparison() ast.Expr {
	left := p.parseAddSub()
	for p.curToken.Type == lexer.TOK_LT || p.curToken.Type == lexer.TOK_GT ||
		p.curToken.Type == lexer.TOK_LTE || p.curToken.Type == lexer.TOK_GTE {
		op := p.curToken.Literal
		opLine := p.curToken.Line
		p.nextToken()
		right := p.parseAddSub()
		if _, ok := left.(*ast.NullLit); ok {
			p.addWarning("using null literal with comparison operator %s", op)
		}
		if _, ok := right.(*ast.NullLit); ok {
			p.addWarning("using null literal with comparison operator %s", op)
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Line: opLine}
	}
	return left
}

func (p *Parser) parseAddSub() ast.Expr {
	left := p.parseMulDiv()
	for p.curToken.Type == lexer.TOK_PLUS || p.curToken.Type == lexer.TOK_MINUS {
		op := p.curToken.Literal
		opLine := p.curToken.Line
		p.nextToken()
		right := p.parseMulDiv()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Line: opLine}
	}
	return left
}

func (p *Parser) parseMulDiv() ast.Expr {
	left := p.parseUnary()
	for p.curToken.Type == lexer.TOK_STAR || p.curToken.Type == lexer.TOK_SLASH || p.curToken.Type == lexer.TOK_MODULO {
		op := p.curToken.Literal
		opLine := p.curToken.Line
		p.nextToken()
		right := p.parseUnary()
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right, Line: opLine}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expr {
	if p.curToken.Type == lexer.TOK_MINUS {
		if p.peekToken.Type == lexer.TOK_MINUS {
			p.addError("unary minus cannot be applied to a minus expression")
			return nil
		}
		opLine := p.curToken.Line
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
			Left:  &ast.IntegerLit{Value: 0, Untyped: true, Line: opLine},
			Op:    "-",
			Right: expr,
			Line:  opLine,
		}
	}
	if p.curToken.Type == lexer.TOK_NOT {
		opLine := p.curToken.Line
		p.nextToken()
		expr := p.parseUnary()
		if _, ok := expr.(*ast.NullLit); ok {
			p.addWarning("using null literal with boolean ! operator")
		}
		return &ast.UnaryExpr{Op: "!", Right: expr, Line: opLine}
	}
	if p.curToken.Type == lexer.TOK_COPY {
		opLine := p.curToken.Line
		p.nextToken()
		expr := p.parseUnary()
		return &ast.CopyExpr{Right: expr, Line: opLine}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Expr {
	left := p.parsePrimary()
	for {
		switch p.curToken.Type {
		case lexer.TOK_LBRACKET:
			opLine := p.curToken.Line
			p.nextToken()
			index := p.parseExpr()
			if p.curToken.Type != lexer.TOK_RBRACKET {
				p.addError("expected ']'")
				return left
			}
			p.nextToken()
			left = &ast.IndexExpr{Object: left, Index: index, Line: opLine}
		case lexer.TOK_DOT:
			opLine := p.curToken.Line
			p.nextToken()
			if !isMemberToken(p.curToken.Type) {
				p.addError("expected member name after '.'")
				break
			}
			name := p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == lexer.TOK_LPAREN {
				p.nextToken()
				args := make([]ast.Expr, 0)
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
				left = &ast.MemberAccess{Object: left, Member: name, Args: args, Line: opLine}
			} else {
				left = &ast.MemberAccess{Object: left, Member: name, Line: opLine}
			}
		case lexer.TOK_LPAREN:
			opLine := p.curToken.Line
			p.nextToken()
			args := make([]ast.Expr, 0)
			if p.curToken.Type != lexer.TOK_RPAREN {
				args = append(args, p.parseExpr())
				for p.curToken.Type == lexer.TOK_COMMA {
					p.nextToken()
					args = append(args, p.parseExpr())
				}
			}
			if p.curToken.Type != lexer.TOK_RPAREN {
				p.addError("expected ')' after arguments")
				return left
			}
			p.nextToken()
			left = &ast.CallExpr{Function: left, Args: args, Line: opLine}
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
		litLine := p.curToken.Line
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
		return &ast.IntegerLit{Value: val, Untyped: true, Line: litLine}
	case lexer.TOK_FLOAT_LIT:
		lit := p.curToken.Literal
		litLine := p.curToken.Line
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
		return &ast.FloatLit{Value: val, Untyped: true, Line: litLine}
	case lexer.TOK_IDENT:
		tok := p.curToken
		p.nextToken()
		return &ast.VarRef{Name: tok.Literal, Line: tok.Line}
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
		pLine := p.curToken.Line
		p.nextToken()
		return &ast.BoolLit{Value: true, Untyped: true, Line: pLine}
	case lexer.TOK_FALSE:
		pLine := p.curToken.Line
		p.nextToken()
		return &ast.BoolLit{Value: false, Untyped: true, Line: pLine}
	case lexer.TOK_NULL:
		pLine := p.curToken.Line
		p.nextToken()
		return &ast.NullLit{Line: pLine}
	case lexer.TOK_STRING_LIT:
		lit := p.curToken.Literal
		litLine := p.curToken.Line
		p.nextToken()
		return &ast.StringLit{Value: lit, Untyped: true, Line: litLine}
	case lexer.TOK_LBRACKET:
		return p.parseArrayLit()
	case lexer.TOK_INT:
		refLine := p.curToken.Line
		t := p.parseIntegerType()
		return &ast.TypeRef{Type: t, IsType: true, Line: refLine}
	case lexer.TOK_FLOAT:
		refLine := p.curToken.Line
		t := p.parseFloatType()
		return &ast.TypeRef{Type: t, IsType: true, Line: refLine}
	case lexer.TOK_BOOL:
		refLine := p.curToken.Line
		t := p.parseBoolType()
		return &ast.TypeRef{Type: t, IsType: true, Line: refLine}
	case lexer.TOK_ARRAY:
		refLine := p.curToken.Line
		t := p.parseArrayType()
		return &ast.TypeRef{Type: t, IsType: true, Line: refLine}
	case lexer.TOK_LIST:
		refLine := p.curToken.Line
		t := p.parseListType()
		return &ast.TypeRef{Type: t, IsType: true, Line: refLine}
	case lexer.TOK_STRING:
		refLine := p.curToken.Line
		t := p.parseStringType()
		return &ast.TypeRef{Type: t, IsType: true, Line: refLine}
	case lexer.TOK_TYPEOF:
		if p.peekToken.Type != lexer.TOK_LPAREN {
			p.addError("expected '(' after 'typeof'")
			return nil
		}
		typeofLine := p.curToken.Line
		p.nextToken()
		p.nextToken()
		expr := p.parseExpr()
		if p.curToken.Type != lexer.TOK_RPAREN {
			p.addError("expected ')' after typeof operand")
			return nil
		}
		p.nextToken()
		return &ast.TypeOfExpr{Expr: expr, Line: typeofLine}
	default:
		p.addError("unexpected token in expression: %s", p.curToken.Literal)
		return nil
	}
}

func (p *Parser) parseArrayLit() ast.Expr {
	arrLine := p.curToken.Line
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
		return &ast.ArrayLit{Elements: elements, Line: arrLine}
	}
	p.nextToken()
	return &ast.ArrayLit{Elements: elements, Line: arrLine}
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
