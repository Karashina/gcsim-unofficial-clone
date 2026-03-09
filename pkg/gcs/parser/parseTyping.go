package parser

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
)

func (p *Parser) parseOptionalType() (ast.ExprType, error) {
	// 型情報が存在するのはfnシグネチャまたはlet文の後のみ
	// 次のトークンが識別子またはfnであれば型情報と仮定して安全
	// どちらでもなければnilを返す
	n := p.peek()
	switch n.Typ {
	case ast.ItemIdentifier:
	case ast.KeywordFn:
	default:
		return nil, nil
	}
	return p.parseTyping()
}

func (p *Parser) parseTyping() (ast.ExprType, error) {
	// 次は以下のいずれかの値を持つ識別子であること、そうでなければエラー
	// - 数値
	// - 文字列
	// - fn(...) : ...
	n := p.peek()
	switch n.Typ {
	case ast.ItemIdentifier:
		return p.parseBasicType()
	case ast.KeywordFn:
		return p.parseFnType()
	default:
		return nil, fmt.Errorf("ln%v: error parsing type info, unexpected value after :, got %v", n.Line, n.Val)
	}
}

func (p *Parser) parseBasicType() (ast.ExprType, error) {
	n := p.next()
	if n.Typ != ast.ItemIdentifier {
		return nil, fmt.Errorf("ln%v: error parsing basic type, expecting identifier, got %v", n.Line, n.Val)
	}
	switch n.Val {
	case "string":
		return &ast.StringType{Pos: n.Pos}, nil
	case "number":
		return &ast.NumberType{Pos: n.Pos}, nil
	case "map":
		return &ast.MapType{Pos: n.Pos}, nil
	default:
		return nil, fmt.Errorf("ln%v: unexpected basic type parsing type info; got %v", n.Line, n.Val)
	}
}

func (p *Parser) parseFnType() (ast.ExprType, error) {
	// 以下のような形式を期待: fn(number) : number
	// 以下も有効:
	//   fn(fn(number):number, number) : fn(number)
	// 再帰呼び出しが多くなる...
	var err error
	n := p.next()
	if n.Typ != ast.KeywordFn {
		return nil, fmt.Errorf("ln%v: error parsing fn type, expecting fn, got %v", n.Line, n.Val)
	}
	res := &ast.FuncType{
		Pos: n.Pos,
	}
	// 以下の順序を期待する:
	// - (
	// - カンマ区切りの型
	// - )
	// - オプションの戻り値
	n = p.next()
	if n.Typ != ast.ItemLeftParen {
		return nil, fmt.Errorf("ln%v: expecting ( after fn parsing typing, got %v", n.Line, n.Val)
	}
	done := false
	// 次が右括弧ならこの関数は引数なし
	if l := p.peek(); l.Typ == ast.ItemRightParen {
		// トークンを消費
		p.next()
		done = true
	}
	for !done {
		// 最初のトークンまたはトークン群が型情報であることを期待する
		typ, err := p.parseTyping()
		if err != nil {
			return nil, err
		}
		res.ArgsType = append(res.ArgsType, typ)

		// 次のトークンは ) で終了を示すか、カンマでパースの継続を意味する
		n = p.next()
		switch n.Typ {
		case ast.ItemRightParen:
			// 次が ) なら終了
			done = true
		case ast.ItemComma:
			// カンマは継続を意味する
		default:
			// 予期しないトークン
			return nil, fmt.Errorf("ln%v: unexpected token parsing fn type: %v", n.Line, n.Val)
		}
	}
	// オプションの戻り値型を確認
	res.ResultType, err = p.parseOptionalType()
	if err != nil {
		return nil, err
	}
	//TODO: これは互換性のためのみ存在する; 削除すべき?
	if res.ResultType == nil {
		res.ResultType = &ast.NumberType{Pos: n.Pos}
	}
	return res, nil
}
