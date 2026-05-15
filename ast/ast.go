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
	Line    int
}

type FloatLit struct {
	Value   float64
	FType   FloatType
	Untyped bool
	Line    int
}

type BoolLit struct {
	Value   bool
	BType   BoolType
	Untyped bool
	Line    int
}

type StringLit struct {
	Value   string
	SType   StringType
	Untyped bool
	Line    int
}

type NullLit struct {
	Line int
}

type VarRef struct {
	Name string
	Line int
}

type UnaryExpr struct {
	Op    string
	Right Expr
	Line  int
}

type TypeOfExpr struct {
	Expr Expr
	Line int
}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
	Line  int
}

type IndexExpr struct {
	Object Expr
	Index  Expr
	Line   int
}

type ArrayLit struct {
	Elements []Expr
	Line     int
}

type MemberAccess struct {
	Object Expr
	Member string
	Args   []Expr
	Line   int
}

type TypeRef struct {
	Type   Type
	IsType bool
	Line   int
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
	Line     int
}

type Assignment struct {
	Name  string
	Index Expr
	Op    string
	Expr  Expr
	IsRef bool
	Line  int
}

type RefDecl struct {
	Name string
	Type Type
	Expr Expr
	Line int
}

type CopyExpr struct {
	Right Expr
	Line  int
}

type RefExpr struct {
	Right Expr
	Line  int
}

type IsExpr struct {
	Left  Expr
	Right Expr
	Line  int
}

type PrintStmt struct {
	Expr Expr
	Line int
}

type ExprStmt struct {
	Expr Expr
	Line int
}

type BlockStmt struct {
	Stmts []Stmt
	Line  int
}

type IfStmt struct {
	Condition Expr
	Then      *BlockStmt
	Else      Stmt
	Line      int
}

type ForStmt struct {
	Init      Stmt
	Condition Expr
	Update    Stmt
	Body      *BlockStmt
	Line      int
}

type ForInStmt struct {
	VarName string
	Iter    Expr
	Body    *BlockStmt
	Line    int
}

type ForAtStmt struct {
	VarName string
	Iter    Expr
	Body    *BlockStmt
	Line    int
}

type ForOfStmt struct {
	VarName1 string
	VarName2 string
	Iter     Expr
	Body     *BlockStmt
	Line     int
}

type WhileStmt struct {
	Condition Expr
	Body      *BlockStmt
	Line      int
}

type BreakStmt struct {
	Line int
}

type SkipStmt struct {
	Line int
}

type IncDecStmt struct {
	Name string
	Op   string
	Line int
}

type FuncDecl struct {
	Name       string
	Parameters []Param
	ReturnType Type
	Body       *BlockStmt
	Line       int
}

type Param struct {
	Name string
	Type Type
}

type ReturnStmt struct {
	Value Expr
	Line  int
}

type CallExpr struct {
	Function Expr
	Args     []Expr
	Line     int
}

type Program struct {
	Stmts []Stmt
}
