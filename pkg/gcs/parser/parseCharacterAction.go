package parser

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/validation"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/shortcut"
)

// parseAction はキャラクターアクションを含むノード、またはアクションリストを含むノードブロックを返す
func (p *Parser) parseAction() (ast.Stmt, error) {
	char, err := p.consume(ast.ItemCharacterKey)
	if err != nil {
		// 既にチェック済みなので、ここに到達するはずがない
		return nil, fmt.Errorf("ln%v: expecting character key, got %v", char.Line, char.Val)
	}
	charKey := shortcut.CharNameToKey[char.Val]

	// 次に複数のアクションキーが来るはず
	var actions []*ast.CallExpr
	if n := p.peek(); n.Typ != ast.ItemActionKey {
		return nil, fmt.Errorf("ln%v: expecting actions for character %v, got %v", n.Line, char.Val, n.Val)
	}

	// 全アクションは + フラグより前に来る必要がある
Loop:
	for {
		switch n := p.next(); n.Typ {
		case ast.ItemTerminateLine:
			// ここで停止
			break Loop
		case ast.ItemActionKey:
			actionKey := action.StringToAction(n.Val)
			expr := &ast.CallExpr{
				Pos: char.Pos,
				Fun: &ast.Ident{
					Pos:   n.Pos,
					Value: "execute_action",
				},
				Args: make([]ast.Expr, 0),
			}
			expr.Args = append(expr.Args,
				// キャラクター
				&ast.NumberLit{
					Pos:      char.Pos,
					IntVal:   int64(charKey),
					FloatVal: float64(charKey),
				},
				// アクション
				&ast.NumberLit{
					Pos:      n.Pos,
					IntVal:   int64(actionKey),
					FloatVal: float64(actionKey),
				},
			)
			// パラメータをチェック → 繰り返し
			param, err := p.acceptOptionalParamReturnMap()
			if err != nil {
				return nil, err
			}
			if param == nil {
				param = &ast.MapExpr{Pos: n.Pos}
			}
			// パラメータを検証
			// TODO: "コンパイル"ステップがまだないため非効率だが仕方ない
			m := param.(*ast.MapExpr).Fields
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			err = validation.ValidateCharParamKeys(charKey, actionKey, keys)
			if err != nil {
				return nil, fmt.Errorf("ln%v: character %v: %w", n.Line, charKey, err)
			}
			expr.Args = append(expr.Args, param)

			// オプションの : と数値
			count, err := p.acceptOptionalRepeaterReturnCount()
			if err != nil {
				return nil, err
			}
			// 配列に追加
			for i := 0; i < count; i++ {
				//TODO: 繰り返しアクションは全て同じマップにアクセスする
				// アビリティ実装はマップの変更を避けるべき
				actions = append(actions, expr)
			}

			n = p.next()
			if n.Typ != ast.ItemComma {
				p.backup()
				break Loop
			}
		default:
			//TODO: 無効なキーのエラーを修正する
			return nil, fmt.Errorf("ln%v: expecting actions for character %v, got %v", n.Line, char.Val, n.Val)
		}
	}
	// オプションフラグをチェック

	// 文を構築
	b := ast.NewBlockStmt(char.Pos)
	for _, v := range actions {
		b.Append(v)
	}
	return b, nil
}

func (p *Parser) acceptOptionalParamReturnMap() (ast.Expr, error) {
	// パラメータをチェック
	n := p.peek()
	if n.Typ != ast.ItemLeftSquareParen {
		return nil, nil
	}

	return p.parseMap()
}

func (p *Parser) acceptOptionalParamReturnOnlyIntMap() (map[string]int, error) {
	r := make(map[string]int)

	result, err := p.acceptOptionalParamReturnMap()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return r, nil
	}

	for k, v := range result.(*ast.MapExpr).Fields {
		switch v.(type) {
		case *ast.NumberLit:
			// スキップ
		default:
			return nil, fmt.Errorf("expected number in the map, got %v", v.String())
		}
		r[k] = int(v.(*ast.NumberLit).IntVal)
	}
	return r, nil
}

func (p *Parser) acceptOptionalRepeaterReturnCount() (int, error) {
	count := 1
	n := p.next()
	if n.Typ != ast.ItemColon {
		p.backup()
		return count, nil
	}
	// 次は数値であるべき
	n = p.next()
	if n.Typ != ast.ItemNumber {
		return count, fmt.Errorf("ln%v: expected a number after : but got %v", n.Line, n)
	}
	// 数値を解析
	count, err := itemNumberToInt(n)
	return count, err
}
