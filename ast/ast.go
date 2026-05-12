package ast

type Node interface{}

type Expr interface {
	Node
}

type Stmt interface {
	Node
}

type Type interface {
	Kind() string
}

func (IntegerType) Kind() string { return "int" }
func (FloatType) Kind() string   { return "float" }
func (BoolType) Kind() string    { return "bool" }
func (ArrayType) Kind() string   { return "array" }
func (ListType) Kind() string    { return "list" }
func (StringType) Kind() string  { return "string" }

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

type ArrayType struct {
	ElemType Type
	Size     int
}

type ListType struct {
	ElemType Type
	HasMin   bool
	MinSize  int
	HasMax   bool
	MaxSize  int
}

type StringType struct {
	Size    int
	HasMin  bool
	MinSize int
	HasMax  bool
	MaxSize int
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

type StringLit struct {
	Value   string
	SType   StringType
	Untyped bool
}

type NullLit struct{}

type VarRef struct {
	Name string
}

type UnaryExpr struct {
	Op    string
	Right Expr
}

type TypeOfExpr struct {
	Expr Expr
}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

type IndexExpr struct {
	Object Expr
	Index  Expr
}

type ArrayLit struct {
	Elements []Expr
}

type MemberAccess struct {
	Object Expr
	Member string
	Args   []Expr
}

type TypeRef struct {
	Type   Type
	IsType bool
}

type VarDecl struct {
	Name     string
	Type     Type
	IType    IntegerType
	FType    FloatType
	BType    BoolType
	SType    StringType
	Expr     Expr
	IsFloat  bool
	IsBool   bool
	IsString bool
}

type Assignment struct {
	Name  string
	Index Expr
	Op    string
	Expr  Expr
}

type PrintStmt struct {
	Expr Expr
}

type ExprStmt struct {
	Expr Expr
}

type BlockStmt struct {
	Stmts []Stmt
}

type IfStmt struct {
	Condition Expr
	Then      *BlockStmt
	Else      Stmt
}

type ForStmt struct {
	Init      Stmt
	Condition Expr
	Update    Stmt
	Body      *BlockStmt
}

type WhileStmt struct {
	Condition Expr
	Body      *BlockStmt
}

type BreakStmt struct{}

type SkipStmt struct{}

type IncDecStmt struct {
	Name string
	Op   string
}

type Program struct {
	Stmts []Stmt
}
