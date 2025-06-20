package eval

import (
	"fmt"
	"log"
	"testing"

	"github.com/genshinsim/gcsim/pkg/gcs/parser"
)

func TestFib(t *testing.T) {
	prog := `
	fn fib(a number) number {
		if a <= 1 {
			return a;
		}
		return fib(a - 1) + fib(a - 2);
	}
	let y = fib(9);
	print(y);
	return y;
	`
	p := parser.New(prog)
	_, gcsl, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("program:")
	fmt.Println(gcsl.String())
	eval, _ := NewEvaluator(gcsl, nil)
	eval.Log = log.Default()
	resultChan := make(chan Obj)
	go func() {
		res, err := eval.Run()
		fmt.Printf("done with result: %v, err: %v\n", res, err)
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
	result := <-resultChan
	if result.Typ() != typRet {
		t.Errorf("expecting type to return ret, got %v", result.Typ())
	}
	if eval.Err() != nil {
		t.Error(eval.Err())
	}
	// should get 34
	res := result.(*retval)
	val, ok := res.res.(*number)
	if !ok {
		t.Errorf("expecting number for return, got %v", res.res.Typ())
		t.FailNow()
	}
	if val.ival != 34 {
		t.Errorf("expecting answer to be 34, got %v", val.ival)
	}
}

func TestFunctional(t *testing.T) {
	prog := `
	fn g(a number) fn() number {
		return fn() number {
			return a + 1;
		};
	}
	let x = g(1)();
	print(x);
	return x;
	`
	p := parser.New(prog)
	_, gcsl, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("program:")
	fmt.Println(gcsl.String())
	eval, _ := NewEvaluator(gcsl, nil)
	eval.Log = log.Default()
	resultChan := make(chan Obj)
	go func() {
		res, err := eval.Run()
		fmt.Printf("done with result: %v, err: %v\n", res, err)
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
	result := <-resultChan
	if result.Typ() != typRet {
		t.Errorf("expecting type to return ret, got %v", result.Typ())
	}
	if eval.Err() != nil {
		t.Error(eval.Err())
	}
	// should get 2
	res := result.(*retval)
	val, ok := res.res.(*number)
	if !ok {
		t.Errorf("expecting number for return, got %v", res.res.Typ())
		t.FailNow()
	}
	if val.ival != 2 {
		t.Errorf("expecting answer to be 2, got %v", val.ival)
	}
}

func TestAnonFunc(t *testing.T) {
	prog := `
	let x = fn(a) { return a + 1; } (2) + 2;
	print(x);
	return x;
	`
	p := parser.New(prog)
	_, gcsl, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	eval, _ := NewEvaluator(gcsl, nil)
	eval.Log = log.Default()
	resultChan := make(chan Obj)
	go func() {
		res, err := eval.Run()
		fmt.Printf("done with result: %v, err: %v\n", res, err)
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
	result := <-resultChan
	if result.Typ() != typRet {
		t.Errorf("expecting type to return ret, got %v", result.Typ())
	}
	if eval.Err() != nil {
		t.Error(eval.Err())
	}
	// should get 5
	res := result.(*retval)
	val, ok := res.res.(*number)
	if !ok {
		t.Errorf("expecting number for return, got %v", res.res.Typ())
		t.FailNow()
	}
	if val.ival != 5 {
		t.Errorf("expecting answer to be 5, got %v", val.ival)
	}
}

func TestStringFunc(t *testing.T) {
	prog := `
	let x string = fn(a) string { return "hello world"; } (2);
	print(x);
	return x;
	`
	p := parser.New(prog)
	_, gcsl, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	eval, _ := NewEvaluator(gcsl, nil)
	eval.Log = log.Default()
	resultChan := make(chan Obj)
	go func() {
		res, err := eval.Run()
		fmt.Printf("done with result: %v, err: %v\n", res, err)
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
	result := <-resultChan
	if result.Typ() != typRet {
		t.Errorf("expecting type to return ret, got %v", result.Typ())
	}
	if eval.Err() != nil {
		t.Error(eval.Err())
	}
	// should get 5
	res := result.(*retval)
	val, ok := res.res.(*strval)
	if !ok {
		t.Errorf("expecting number for return, got %v", res.res.Typ())
		t.FailNow()
	}
	if val.str != "hello world" {
		t.Errorf("expecting result to be hello world, got %v", val.str)
	}
}

func TestNestedActions(t *testing.T) {
	prog := `
	fn do() {
		print("hi");
	}
	do();
	`
	p := parser.New(prog)
	_, gcsl, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("program:")
	fmt.Println(gcsl.String())
	eval, _ := NewEvaluator(gcsl, nil)
	eval.Log = log.Default()
	resultChan := make(chan Obj)
	go func() {
		res, err := eval.Run()
		fmt.Printf("done with result: %v, err: %v\n", res, err)
		if err != nil {
			log.Fatalf("test run failed with error: %v", err)
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
	result := <-resultChan
	// by default, functions return num
	if result.Typ() != typNum {
		t.Errorf("expecting type to return num, got %v", result.Typ())
	}
	if eval.Err() != nil {
		t.Error(eval.Err())
	}
}

func TestIsEven(t *testing.T) {
	prog := `is_even(1);`
	p := parser.New(prog)
	_, gcsl, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("program:")
	fmt.Println(gcsl.String())
	eval, _ := NewEvaluator(gcsl, nil)
	eval.Log = log.Default()
	resultChan := make(chan Obj)
	go func() {
		res, err := eval.Run()
		fmt.Printf("done with result: %v, err: %v\n", res, err)
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
	result := <-resultChan
	if result.Typ() != typNum {
		t.Errorf("expecting type to return num, got %v", result.Typ())
	}
	if eval.Err() != nil {
		t.Error(eval.Err())
	}
	val := result.(*number)
	if val.ival != 0 {
		t.Errorf("expecting result to be 0, got %v", val.ival)
	}
}
