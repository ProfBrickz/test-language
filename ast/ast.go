package ast

type Node interface {
}

type Expr interface {
	Node
}

type Stmt interface {
	Node
}

type IntegerType struct {
	Size     int
	Signed   bool
	Nullable bool
}

type IntegerLit struct {
	Value   int64
	IType   IntegerType
	Untyped bool
}

type NullLit struct {
}

type VarRef struct {
	Name string
}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

type VarDecl struct {
	Name  string
	IType IntegerType
	Expr  Expr
}

type Assignment struct {
	Name string
	Op   string
	Expr Expr
}

type PrintStmt struct {
	Expr Expr
}

type Program struct {
	Stmts []Stmt
}
