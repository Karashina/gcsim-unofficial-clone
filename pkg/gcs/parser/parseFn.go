package parser

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
)

func (p *Parser) parseFnStmt() (ast.Stmt, error) {
	// fn ident(...引数){ ブロック }
	n := p.next()
	if n.Typ != ast.KeywordFn {
		return nil, fmt.Errorf("ln %v: expecting fn, got %v", n.Line, n.Val)
	}
	n = p.next()
	if n.Typ != ast.ItemIdentifier {
		return nil, fmt.Errorf("ln %v: expecting identifier after fn, got %v", n.Line, n.Val)
	}
	// 関数本体が必要
	lit, err := p.parseFn()
	if err != nil {
		return nil, err
	}
	return &ast.FnStmt{
		Pos:   n.Pos,
		Ident: n,
		Func:  lit,
	}, nil
}

func (p *Parser) parseFnExpr() (ast.Expr, error) {
	// fn (...識別子) { ブロック }
	// fn を消費
	n := p.next()
	if n.Typ != ast.KeywordFn {
		return nil, fmt.Errorf("ln %v: expecting fn, got %v", n.Line, n.Val)
	}
	// 関数本体が必要
	lit, err := p.parseFn()
	if err != nil {
		return nil, err
	}
	return &ast.FuncExpr{
		Pos:  n.Pos,
		Func: lit,
	}, nil
}

func (p *Parser) parseFn() (*ast.FuncLit, error) {
	// (...識別子){ ブロック }
	var err error

	// n は左括弧であることを期待
	n := p.peek()
	if n.Typ != ast.ItemLeftParen {
		return nil, fmt.Errorf("ln%v: expecting ( after identifier, got %v", n.Line, n.Val)
	}

	lit := &ast.FuncLit{
		Pos: n.Pos,
		Signature: &ast.FuncType{
			Pos: n.Pos,
		},
	}

	// 引数をパース
	lit.Args, lit.Signature.ArgsType, err = p.parseFnArgs()
	if err != nil {
		return nil, err
	}

	// 引数が重複していないかチェック
	chk := make(map[string]bool)
	for _, v := range lit.Args {
		if _, ok := chk[v.Value]; ok {
			return nil, fmt.Errorf("ln%v: fn contains duplicated param name %v", n.Line, v.Value)
		}
		chk[v.Value] = true
	}

	// 次が左ブレースでない場合、型情報を期待
	if l := p.peek(); l.Typ != ast.ItemLeftBrace {
		lit.Signature.ResultType, err = p.parseTyping()
		if err != nil {
			return nil, err
		}
	}
	//TODO: nilの場合、互換性のため数値型と仮定する
	if lit.Signature.ResultType == nil {
		//TODO: ここの位置情報は正しくない…開き括弧の位置ではないはず
		//TODO: パーサーに現在位置を追加して修正すべき
		lit.Signature.ResultType = &ast.NumberType{Pos: n.Pos}
	}

	lit.Body, err = p.parseBlock()
	if err != nil {
		return nil, err
	}

	return lit, nil
}

func (p *Parser) parseFnArgs() ([]*ast.Ident, []ast.ExprType, error) {
	// ( を消費
	var args []*ast.Ident
	var argsType []ast.ExprType
	p.next()
	for n := p.next(); n.Typ != ast.ItemRightParen; n = p.next() {
		a := &ast.Ident{}
		// 識別子かカンマを期待
		if n.Typ != ast.ItemIdentifier {
			return nil, nil, fmt.Errorf("ln%v: expecting identifier in param list, got %v", n.Line, n.Val)
		}
		a.Pos = n.Pos
		a.Value = n.Val

		args = append(args, a)

		// オプションの型情報をチェック
		// 存在しない場合は数値型と仮定
		typ, err := p.parseOptionalType()
		if err != nil {
			return nil, nil, err
		}
		//TODO: nilの場合、互換性のため数値型と仮定する
		if typ == nil {
			typ = &ast.NumberType{Pos: n.Pos}
		}

		argsType = append(argsType, typ)

		// 次のトークンがカンマの場合、その後に別の識別子があるはず
		// そうでなければエラー
		if l := p.peek(); l.Typ == ast.ItemComma {
			p.next() // consume the comma
			if l = p.peek(); l.Typ != ast.ItemIdentifier {
				return nil, nil, fmt.Errorf("ln%v: expecting another identifier after comma in param list, got %v", n.Line, n.Val)
			}
		}
	}
	return args, argsType, nil
}
