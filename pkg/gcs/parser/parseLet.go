package parser

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
)

// let ident = expr; の形式を期待
func (p *Parser) parseLet() (ast.Stmt, error) {
	n := p.next()

	ident, err := p.consume(ast.ItemIdentifier)
	if err != nil {
		// 次のトークンが識別子でない
		return nil, fmt.Errorf("ln%v: expecting identifier after let, got %v", ident.Line, ident.Val)
	}

	stmt := &ast.LetStmt{
		Pos:   n.Pos,
		Ident: ident,
	}

	// オプションの型情報; 存在しない場合は数値型と仮定
	if l := p.peek(); l.Typ != ast.ItemAssign {
		stmt.Type, err = p.parseTyping()
		if err != nil {
			return nil, err
		}
	}
	if stmt.Type == nil {
		stmt.Type = &ast.NumberType{Pos: n.Pos}
	}

	a, err := p.consume(ast.ItemAssign)
	if err != nil {
		// 次のトークンが識別子でない
		return nil, fmt.Errorf("ln%v: expecting = after identifier in let statement, got %v", a.Line, a.Val)
	}

	stmt.Val, err = p.parseExpr(ast.Lowest)

	return stmt, err
}
