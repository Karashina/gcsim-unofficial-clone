package eval

import (
	"fmt"
	"log"
	"testing"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/parser"
)

func TestNormalizeStrObjPassesNonStrval(t *testing.T) {
	obj := &number{ival: 42}
	result := normalizeStrObj(obj)
	if result != obj {
		t.Error("expected normalizeStrObj to return the same object for non-strval")
	}
}

func TestNormalizeStrObjUnknownNamePassthrough(t *testing.T) {
	obj := &strval{str: "unknownname"}
	result := normalizeStrObj(obj)
	sv, ok := result.(*strval)
	if !ok {
		t.Fatal("expected strval back")
	}
	if sv.str != "unknownname" {
		t.Errorf("expected \"unknownname\", got %q", sv.str)
	}
}

func TestNormalizeStrObjCanonicalNameUnchanged(t *testing.T) {
	// ショートカットマッピングのない正規名はそのまま返される
	obj := &strval{str: "somefullcanonicalname"}
	result := normalizeStrObj(obj)
	sv, ok := result.(*strval)
	if !ok {
		t.Fatal("expected strval back")
	}
	if sv.str != "somefullcanonicalname" {
		t.Errorf("expected \"somefullcanonicalname\", got %q", sv.str)
	}
}

// runGCSLExprはGCSL式を実行し結果のObjを返す。
func runGCSLExpr(t *testing.T, prog string) Obj {
	t.Helper()
	p := parser.New(prog)
	_, gcsl, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	eval, _ := NewEvaluator(gcsl, nil)
	eval.Log = log.Default()
	resultChan := make(chan Obj)
	go func() {
		res, err := eval.Run()
		if err != nil {
			fmt.Printf("eval error: %v\n", err)
		}
		resultChan <- res
	}()
	for {
		eval.Continue()
		a, err := eval.NextAction()
		if a == nil {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	return <-resultChan
}

func TestGCSLStringEqualTrue(t *testing.T) {
	result := runGCSLExpr(t, `return "hello" == "hello";`)
	rv, ok := result.(*retval)
	if !ok {
		t.Fatal("expected retval")
	}
	num, ok := rv.res.(*number)
	if !ok {
		t.Fatalf("expected number result, got %v", rv.res.Typ())
	}
	if num.ival != 1 {
		t.Error("expected \"hello\" == \"hello\" to be truthy (1)")
	}
}

func TestGCSLStringEqualFalse(t *testing.T) {
	result := runGCSLExpr(t, `return "hello" == "world";`)
	rv, ok := result.(*retval)
	if !ok {
		t.Fatal("expected retval")
	}
	num, ok := rv.res.(*number)
	if !ok {
		t.Fatalf("expected number result, got %v", rv.res.Typ())
	}
	if num.ival != 0 {
		t.Error("expected \"hello\" == \"world\" to be falsy (0)")
	}
}

func TestGCSLStringNotEqual(t *testing.T) {
	result := runGCSLExpr(t, `return "abc" != "xyz";`)
	rv, ok := result.(*retval)
	if !ok {
		t.Fatal("expected retval")
	}
	num, ok := rv.res.(*number)
	if !ok {
		t.Fatalf("expected number result, got %v", rv.res.Typ())
	}
	if num.ival != 1 {
		t.Error("expected \"abc\" != \"xyz\" to be truthy (1)")
	}
}

func TestGCSLStringInIfCondition(t *testing.T) {
	prog := `
	let x = "test";
	if x == "test" {
		return 1;
	}
	return 0;
	`
	result := runGCSLExpr(t, prog)
	rv, ok := result.(*retval)
	if !ok {
		t.Fatal("expected retval")
	}
	num, ok := rv.res.(*number)
	if !ok {
		t.Fatalf("expected number result, got %v", rv.res.Typ())
	}
	if num.ival != 1 {
		t.Error("expected string comparison in if to take the true branch")
	}
}
