package constant

import (
	"testing"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
)

func TestStringEqualTrue(t *testing.T) {
	left := Make("hello")
	right := Make("hello")
	op := ast.Token{Typ: ast.OpEqual}
	result, err := BinaryOp(op, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if !ToBool(result) {
		t.Error("expected \"hello\" == \"hello\" to be true")
	}
}

func TestStringEqualFalse(t *testing.T) {
	left := Make("hello")
	right := Make("world")
	op := ast.Token{Typ: ast.OpEqual}
	result, err := BinaryOp(op, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if ToBool(result) {
		t.Error("expected \"hello\" == \"world\" to be false")
	}
}

func TestStringNotEqualTrue(t *testing.T) {
	left := Make("hello")
	right := Make("world")
	op := ast.Token{Typ: ast.OpNotEqual}
	result, err := BinaryOp(op, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if !ToBool(result) {
		t.Error("expected \"hello\" != \"world\" to be true")
	}
}

func TestStringNotEqualFalse(t *testing.T) {
	left := Make("hello")
	right := Make("hello")
	op := ast.Token{Typ: ast.OpNotEqual}
	result, err := BinaryOp(op, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if ToBool(result) {
		t.Error("expected \"hello\" != \"hello\" to be false")
	}
}

func TestStringVsNumberError(t *testing.T) {
	left := Make("hello")
	right := Make(42)
	op := ast.Token{Typ: ast.OpEqual}
	_, err := BinaryOp(op, left, right)
	if err == nil {
		t.Error("expected error when comparing string with number")
	}
}

func TestStringUnsupportedOpError(t *testing.T) {
	left := Make("hello")
	right := Make("world")
	op := ast.Token{Typ: ast.OpGreaterThan}
	_, err := BinaryOp(op, left, right)
	if err == nil {
		t.Error("expected error for unsupported string operator >")
	}
}
