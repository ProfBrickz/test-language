package ast

import "testing"

func TestTypeKind(t *testing.T) {
	tests := []struct {
		typ  Type
		want string
	}{
		{IntegerType{}, "int"},
		{FloatType{}, "float"},
		{BoolType{}, "bool"},
		{ArrayType{}, "array"},
		{ListType{}, "list"},
		{StringType{}, "string"},
		{UnionType{}, "union"},
	}

	for _, tt := range tests {
		got := tt.typ.Kind()
		if got != tt.want {
			t.Errorf("(%T).Kind() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}
