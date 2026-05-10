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

type BoolType struct {
	Nullable bool
}

type FloatType struct {
	Size     int
	Nullable bool
}

type IntegerLit struct {
	Value   int64
	IType   IntegerType
	Untyped bool
}

type FloatLit struct {
	Value   float64
	FType   FloatType
	Untyped bool
}

type BoolLit struct {
	Value   bool
	BType   BoolType
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

type MemberAccess struct {
	Object Expr
	Member string
}

type TypeRef struct {
	Kind   string // "int", "float", "bool"
	IType  IntegerType
	FType  FloatType
	BType  BoolType
	IsType bool
}

type VarDecl struct {
	Name    string
	IType   IntegerType
	FType   FloatType
	BType   BoolType
	Expr    Expr
	IsFloat bool
	IsBool  bool
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
