package eval

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
)

type Eval struct {
	Core *core.Core
	AST  ast.Node
	Log  *log.Logger

	next chan bool         // 続行前にこのチャネルを待機する
	work chan *action.Eval // このチャネルに作業を送信する
	// 最初に発生したエラーで non-nil に設定される
	// Run() が既にエラーで終了している可能性があるため必要
	err error

	isTerminated bool
}

type Env struct {
	parent *Env
	varMap map[string]*Obj
}

func NewEvaluator(ast ast.Node, c *core.Core) (*Eval, error) {
	e := &Eval{
		AST:  ast,
		Core: c,
		next: make(chan bool),
		work: make(chan *action.Eval),
	}
	return e, nil
}

func NewEnv(parent *Env) *Env {
	return &Env{
		parent: parent,
		varMap: make(map[string]*Obj),
	}
}

//nolint:gocritic // *Obj に対して非ポインタ型は意味をなさない
func (e *Env) v(s string) (*Obj, error) {
	v, ok := e.varMap[s]
	if ok {
		return v, nil
	}
	if e.parent != nil {
		return e.parent.v(s)
	}
	return nil, fmt.Errorf("variable %v does not exist", s)
}

// eval に即座に終了を指示する
func (e *Eval) Exit() error {
	// 残っている作業があれば排出する
	select {
	case <-e.work:
	default:
	}
	if e.isTerminated {
		return e.err
	}
	// これ以上送信や続行ができないようにする
	e.isTerminated = true
	close(e.next)
	close(e.work)
	return e.err
}

func (e *Eval) Continue() {
	if e.isTerminated {
		return
	}
	e.next <- true
}

// NextAction は eval に次のアクションを返すよう要求する。アクションがなければ nil, nil を返す
func (e *Eval) NextAction() (*action.Eval, error) {
	next, ok := <-e.work
	if !ok {
		return nil, nil
	}
	return next, nil
}

func (e *Eval) Start() {
	//TODO: ここでパニックをキャッチすることを検討する
	e.Run()
}

func (e *Eval) Err() error {
	return e.err
}

// Run は提供された AST を実行する。原神固有のアクションは
// NextAction() 経由で利用可能となる
// TODO: すべての関数が実際にエラーを返すようにして defer を除去する
//
//nolint:nonamedreturns,nakedret // not possible to perform the res, err modification without named return
func (e *Eval) Run() (res Obj, err error) {
	defer func() {
		// この defer は e.err が正しく設定されることを保証する。これは最初の defer でなければならない
		// defer は後入れ先出しで呼ばれるため、パニック処理より前に定義する必要がある
		e.err = err
	}()
	//TODO: 将来的にはこれを除去したい
	defer func() {
		// パニックが発生した場合は回復する。それ以外の場合 err は nil に設定する。
		if pErr := recover(); pErr != nil {
			err = fmt.Errorf("panic occured: %v", pErr)
		}
	}()
	// 唯一の送信者であるため work を必ず閉じる
	defer e.Exit()
	if e.Log == nil {
		e.Log = log.New(io.Discard, "", log.LstdFlags)
	}
	// ErrTerminate が破棄されるようにする
	defer func() {
		if errors.Is(err, ErrTerminated) {
			err = nil
		}
	}()

	global := NewEnv(nil)
	e.initSysFuncs(global)

	// 開始シグナルを受け取ったら実行を開始する
	err = e.waitForNext()
	if err != nil {
		return
	}

	// これは Action に到達するまで実行される
	// その後アクションを resp チャネルに渡す
	// そして Next を待ってから再度実行する
	res, err = e.evalNode(e.AST, global)
	return
}

func (e *Eval) waitForNext() error {
	_, ok := <-e.next
	if !ok {
		return ErrTerminated // これ以上の作業なし、シャットダウン
	}
	return nil
}

func (e *Eval) sendWork(w *action.Eval) {
	e.work <- w
}

var ErrTerminated = errors.New("eval terminated")

type Obj interface {
	Inspect() string
	Typ() ObjTyp
}

type ObjTyp int

const (
	typNull ObjTyp = iota
	typNum
	typStr
	typFun
	typBif // built-in function
	typMap
	typRet
	typCtr
	// typTerminate
)

var typStrings = map[ObjTyp]string{
	typNull: "null",
	typNum:  "number",
	typStr:  "string",
	typFun:  "function",
	typBif:  "builtin_function",
	typMap:  "map",
	typRet:  "return",
	typCtr:  "control",
}

func (o ObjTyp) String() string {
	if name, ok := typStrings[o]; ok {
		return name
	}
	return "unknown"
}

// 各種 Obj 型
type (
	null   struct{}
	number struct {
		ival    int64
		fval    float64
		isFloat bool
	}

	strval struct {
		str string
	}

	funcval struct {
		Args      []*ast.Ident
		Body      *ast.BlockStmt
		Signature *ast.FuncType
		Env       *Env
	}

	systemFunc func(*ast.CallExpr, *Env) (Obj, error)

	bfuncval struct {
		Body systemFunc
		Env  *Env
	}

	mapval struct {
		fields map[string]Obj
	}

	retval struct {
		res Obj
	}

	ctrl struct {
		typ ast.CtrlTyp
	}
)

// null 型
func (n *null) Inspect() string { return "null" }
func (n *null) Typ() ObjTyp     { return typNull }

// terminate 型
// func (n *terminate) Inspect() string { return "terminate" }
// func (n *terminate) Typ() ObjTyp     { return typTerminate }

// number 型
func (n *number) Inspect() string {
	if n.isFloat {
		return strconv.FormatFloat(n.fval, 'f', -1, 64)
	}
	return strconv.FormatInt(n.ival, 10)
}
func (n *number) Typ() ObjTyp { return typNum }

// strval 型
func (s *strval) Inspect() string { return s.str }
func (s *strval) Typ() ObjTyp     { return typStr }

// funcval 型
func (f *funcval) Inspect() string { return "function" }
func (f *funcval) Typ() ObjTyp     { return typFun }

// bfuncval 型
func (b *bfuncval) Inspect() string { return "built-in function" }
func (b *bfuncval) Typ() ObjTyp     { return typBif }

// retval 型
func (r *retval) Inspect() string {
	return r.res.Inspect()
}
func (r *retval) Typ() ObjTyp { return typRet }

// mapval 型
func (m *mapval) Inspect() string {
	str := "["
	done := false
	for k, v := range m.fields {
		if done {
			str += ", "
		}
		done = true

		str += k + " = " + v.Inspect()
	}
	str += "]"
	return str
}
func (m *mapval) Typ() ObjTyp { return typMap }

// ctrl 型
func (c *ctrl) Inspect() string {
	switch c.typ {
	case ast.CtrlContinue:
		return "continue"
	case ast.CtrlBreak:
		return "break"
	case ast.CtrlFallthrough:
		return "fallthrough"
	}
	return "invalid"
}
func (c *ctrl) Typ() ObjTyp { return typCtr }

func (e *Eval) evalNode(n ast.Node, env *Env) (Obj, error) {
	switch v := n.(type) {
	case ast.Expr:
		return e.evalExpr(v, env)
	case ast.Stmt:
		return e.evalStmt(v, env)
	default:
		return &null{}, nil
	}
}
