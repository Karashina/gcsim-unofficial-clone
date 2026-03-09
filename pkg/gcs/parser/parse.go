package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/shortcut"
)

// Parse は ActionList とパース失敗時のエラーを返す
func (p *Parser) Parse() (*info.ActionList, ast.Node, error) {
	var err error
	for state := parseRows; state != nil; {
		state, err = state(p)
		if err != nil {
			return nil, nil, err
		}
	}

	// 整合性チェック
	if len(p.charOrder) > 4 {
		p.res.Errors = append(p.res.Errors, fmt.Errorf("config contains a total of %v characters; cannot exceed 4", len(p.charOrder)))
	}

	if p.res.InitialChar == keys.NoChar {
		p.res.Errors = append(p.res.Errors, errors.New("config does not contain active char"))
	}

	initialCharFound := false
	for _, v := range p.charOrder {
		p.res.Characters = append(p.res.Characters, *p.chars[v])
		// アクティブキャラがチームに含まれているか確認
		if v == p.res.InitialChar {
			initialCharFound = true
		}
		// セット数をチェック
		count := 0
		for _, c := range p.chars[v].Sets {
			count += c
		}
		if count > 5 {
			p.res.Errors = append(p.res.Errors, fmt.Errorf("character %v has more than 5 total set items", v.String()))
		}
	}

	if !initialCharFound && p.res.InitialChar != 0 {
		p.res.Errors = append(p.res.Errors, fmt.Errorf("active char %v not found in team", p.res.InitialChar))
	}

	if len(p.res.Targets) == 0 {
		p.res.Errors = append(p.res.Errors, errors.New("config does not contain any targets"))
	}

	// デフォルト値を設定。座標はデフォルト 0,0 のまま
	for i := range p.res.Targets {
		if p.res.Targets[i].Pos.R == 0 {
			p.res.Targets[i].Pos.R = 1
		}
	}

	// ダメージモードの場合、全ターゲットにHPが設定されているか確認
	if p.res.Settings.DamageMode {
		for i := range p.res.Targets {
			if p.res.Targets[i].HP == 0 {
				p.res.Errors = append(p.res.Errors, fmt.Errorf("damage mode is activated; target #%v does not have hp set", i+1))
			}
		}
	}

	// エラーメッセージを構築
	p.res.ErrorMsgs = make([]string, 0, len(p.res.Errors))
	for _, v := range p.res.Errors {
		p.res.ErrorMsgs = append(p.res.ErrorMsgs, v.Error())
	}

	return p.res, p.prog, nil
}

func parseRows(p *Parser) (parseFn, error) {
	switch n := p.peek(); n.Typ {
	case ast.ItemCharacterKey:
		p.next()
		// キャラクターのステータス等かアクションかを確認
		if p.peek().Typ != ast.ItemActionKey {
			// ActionStmt ではない
			// キャラクターとセットキーを設定
			key, ok := shortcut.CharNameToKey[n.Val]
			if !ok {
				// ここに到達することはない
				return nil, fmt.Errorf("ln%v: unexpected error; invalid char key %v", n.Line, n.Val)
			}
			if _, ok := p.chars[key]; !ok {
				p.newChar(key)
			}
			p.currentCharKey = key
			return parseCharacter, nil
		}
		p.backup()
		// アクションアイテムをパース
		// return parseProgram, nil
		node, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		p.prog.Append(node)
		return parseRows, nil
	case ast.KeywordActive:
		p.next()
		// 次はキャラクター、その後行末であるべき
		char, err := p.consume(ast.ItemCharacterKey)
		if err != nil {
			return nil, fmt.Errorf("ln%v: setting active char: invalid char %v", char.Line, char.Val)
		}
		p.res.InitialChar = shortcut.CharNameToKey[char.Val]
		n, err := p.consume(ast.ItemTerminateLine)
		if err != nil {
			return nil, fmt.Errorf("ln%v: expecting ; after active <char>, got %v", n.Line, n.Val)
		}
		return parseRows, nil
	case ast.KeywordTarget:
		p.next()
		return parseTarget, nil
	case ast.KeywordEnergy:
		p.next()
		return parseEnergy, nil
	case ast.KeywordHurt:
		p.next()
		return parseHurt, nil
	case ast.KeywordOptions:
		p.next()
		return parseOptions, nil
	case ast.ItemEOF:
		return nil, nil
	case ast.ItemActionKey:
		return nil, fmt.Errorf("ln%v: unexpected line starts with an action: %v", n.Line, n.Val)
	default: // default should be look for gcsl
		node, err := p.parseStatement()
		p.prog.Append(node)
		if err != nil {
			return nil, err
		}
		return parseRows, nil
	}
}

func (p *Parser) parseStatement() (ast.Node, error) {
	// セミコロンで終了する文と終了しない文がある
	hasSemi := true
	stmtType := ""
	var node ast.Node
	var err error
	switch n := p.peek(); n.Typ {
	case ast.KeywordBreak:
		fallthrough
	case ast.KeywordFallthrough:
		fallthrough
	case ast.KeywordContinue:
		stmtType = "continue"
		node, err = p.parseCtrl()
	case ast.KeywordLet:
		stmtType = "let"
		node, err = p.parseLet()
	case ast.ItemCharacterKey:
		stmtType = "char action"
		node, err = p.parseAction()
	case ast.KeywordReturn:
		stmtType = "return"
		node, err = p.parseReturn()
	case ast.KeywordIf:
		node, err = p.parseIf()
		hasSemi = false
	case ast.KeywordSwitch:
		node, err = p.parseSwitch()
		hasSemi = false
	case ast.KeywordFn:
		// let x = で始まらない関数宣言をパースする
		// 機能的には let 文と同じ
		node, err = p.parseFnStmt()
		hasSemi = false
	case ast.KeywordWhile:
		node, err = p.parseWhile()
		hasSemi = false
	case ast.KeywordFor:
		node, err = p.parseFor()
		hasSemi = false
	case ast.ItemLeftBrace:
		node, err = p.parseBlock()
		hasSemi = false
	case ast.ItemIdentifier:
		p.next()
		// 後に = があるか確認
		if x := p.peek(); x.Typ == ast.ItemAssign {
			p.backup()
			node, err = p.parseAssign()
			break
		}
		// 代入がなければ式として扱う
		p.backup()
		fallthrough
	default:
		node, err = p.parseExpr(ast.Lowest)
	}
	// パースエラーが発生したか確認
	if err != nil {
		return node, err
	}
	// セミコロンを確認
	if hasSemi {
		n, err := p.consume(ast.ItemTerminateLine)
		if err != nil {
			return nil, fmt.Errorf("ln%v: expecting ; at end of %v statement, got %v", n.Line, stmtType, n.Val)
		}
	}
	return node, nil
}

// ident = expr の形式を期待
func (p *Parser) parseAssign() (ast.Stmt, error) {
	ident, err := p.consume(ast.ItemIdentifier)
	if err != nil {
		// 次のトークンが識別子ではない
		return nil, fmt.Errorf("ln%v: expecting identifier in assign statement, got %v", ident.Line, ident.Val)
	}

	a, err := p.consume(ast.ItemAssign)
	if err != nil {
		// 次のトークンが識別子ではない
		return nil, fmt.Errorf("ln%v: expecting = after identifier in assign statement, got %v", a.Line, a.Val)
	}

	expr, err := p.parseExpr(ast.Lowest)

	if err != nil {
		return nil, err
	}

	stmt := &ast.AssignStmt{
		Pos:   ident.Pos,
		Ident: ident,
		Val:   expr,
	}

	return stmt, nil
}

func (p *Parser) parseIf() (ast.Stmt, error) {
	n := p.next()

	stmt := &ast.IfStmt{
		Pos: n.Pos,
	}

	var err error

	stmt.Condition, err = p.parseExpr(ast.Lowest)
	if err != nil {
		return nil, err
	}

	// 次に { を期待
	if n := p.peek(); n.Typ != ast.ItemLeftBrace {
		return nil, fmt.Errorf("ln%v: expecting { after if, got %v", n.Line, n.Val)
	}

	stmt.IfBlock, err = p.parseBlock() // ブロックをパース
	if err != nil {
		return nil, err
	}

	// else がなければ終了
	if n := p.peek(); n.Typ != ast.KeywordElse {
		return stmt, nil
	}

	// else キーワードをスキップ
	p.next()

	// 次の文を期待（if またはブロックのいずれか）
	block, err := p.parseStatement()
	switch block.(type) {
	case *ast.IfStmt, *ast.BlockStmt:
	default:
		stmt.ElseBlock = nil
		return stmt, fmt.Errorf("ln%v: expecting either if or normal block after else", n.Line)
	}

	stmt.ElseBlock = block.(ast.Stmt)

	return stmt, err
}

func (p *Parser) parseSwitch() (ast.Stmt, error) {
	// switch 式 { } をパース
	n, err := p.consume(ast.KeywordSwitch)
	if err != nil {
		panic("unreachable")
	}

	stmt := &ast.SwitchStmt{
		Pos: n.Pos,
	}

	// 条件は省略可能。次が itemLeftBrace なら条件を 1 に設定
	if n := p.peek(); n.Typ != ast.ItemLeftBrace {
		stmt.Condition, err = p.parseExpr(ast.Lowest)
		if err != nil {
			return nil, err
		}
	} else {
		stmt.Condition = nil
	}

	if n := p.next(); n.Typ != ast.ItemLeftBrace {
		return nil, fmt.Errorf("ln%v: expecting { after switch, got %v", n.Line, n.Val)
	}

	// } が出るまで case を探す
	for n := p.next(); n.Typ != ast.ItemRightBrace; n = p.next() {
		var err error
		// case 式: ブロック を期待
		switch n.Typ {
		case ast.KeywordCase:
			cs := &ast.CaseStmt{
				Pos: n.Pos,
			}
			cs.Condition, err = p.parseExpr(ast.Lowest)
			if err != nil {
				return nil, err
			}
			// コロンの後、次の case まで読み取る
			if n := p.peek(); n.Typ != ast.ItemColon {
				return nil, fmt.Errorf("ln%v: expecting : after case, got %v", n.Line, n.Val)
			}
			cs.Body, err = p.parseCaseBody()
			if err != nil {
				return nil, err
			}
			stmt.Cases = append(stmt.Cases, cs)
		case ast.KeywordDefault:
			// コロンの後、次の case まで読み取る
			if p.peek().Typ != ast.ItemColon {
				return nil, fmt.Errorf("ln%v: expecting : after default, got %v", n.Line, n.Val)
			}
			stmt.Default, err = p.parseCaseBody()
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("ln%v: expecting case or default token, got %v", n.Line, n.Val)
		}
	}

	return stmt, nil
}

func (p *Parser) parseCaseBody() (*ast.BlockStmt, error) {
	n := p.next() // : から開始
	block := ast.NewBlockStmt(n.Pos)
	var node ast.Node
	var err error
	// } に到達するまで1行ずつパース
	for {
		// 不正な行がないことを確認
		switch n := p.peek(); n.Typ {
		case ast.ItemCharacterKey:
			if !p.peekValidCharAction() {
				n = p.next()
				return nil, fmt.Errorf("ln%v: expecting action after character token, got %v", n.Line, n.Val)
			}
		case ast.KeywordDefault:
			fallthrough
		case ast.KeywordCase:
			fallthrough
		case ast.ItemRightBrace:
			return block, nil
		case ast.ItemEOF:
			return nil, fmt.Errorf("reached end of file without closing }")
		}
		// ここで文をパース
		node, err = p.parseStatement()
		if err != nil {
			return nil, err
		}
		block.Append(node)
	}
}

// while { } をパース
func (p *Parser) parseWhile() (ast.Stmt, error) {
	n := p.next()

	stmt := &ast.WhileStmt{
		Pos: n.Pos,
	}

	var err error

	stmt.Condition, err = p.parseExpr(ast.Lowest)
	if err != nil {
		return nil, err
	}

	// 次に { を期待
	if n := p.peek(); n.Typ != ast.ItemLeftBrace {
		return nil, fmt.Errorf("ln%v: expecting { after while, got %v", n.Line, n.Val)
	}

	stmt.WhileBlock, err = p.parseBlock() // ブロックをパース

	return stmt, err
}

// for <初期化 ;> <条件> <; 後処理> { <本体> }
// for { <本体> }
func (p *Parser) existVarDecl() bool {
	switch n := p.peek(); n.Typ {
	case ast.KeywordLet:
		return true
	case ast.ItemIdentifier:
		p.next()
		b := p.peek().Typ == ast.ItemAssign
		p.backup()
		return b
	}
	return false
}

func (p *Parser) parseFor() (ast.Stmt, error) {
	n := p.next()

	stmt := &ast.ForStmt{
		Pos: n.Pos,
	}

	var err error

	if n := p.peek(); n.Typ == ast.ItemLeftBrace {
		stmt.Body, err = p.parseBlock() // ブロックをパース
		return stmt, err
	}

	// 初期化
	if p.existVarDecl() {
		if n := p.peek(); n.Typ == ast.KeywordLet {
			stmt.Init, err = p.parseLet()
		} else {
			stmt.Init, err = p.parseAssign()
		}
		if err != nil {
			return nil, err
		}

		if n := p.peek(); n.Typ != ast.ItemTerminateLine {
			return nil, fmt.Errorf("ln%v: expecting ; after statement, got %v", n.Line, n.Val)
		}
		p.next() // ; をスキップ
	}

	// 条件
	stmt.Cond, err = p.parseExpr(ast.Lowest)
	if err != nil {
		return nil, err
	}

	// 後処理
	if n := p.peek(); n.Typ == ast.ItemTerminateLine {
		p.next() // ; をスキップ
		if n := p.peek(); n.Typ != ast.ItemLeftBrace {
			stmt.Post, err = p.parseAssign()
			if err != nil {
				return nil, err
			}
		}
	}

	// 次に { を期待
	if n := p.peek(); n.Typ != ast.ItemLeftBrace {
		return nil, fmt.Errorf("ln%v: expecting { after for, got %v", n.Line, n.Val)
	}

	stmt.Body, err = p.parseBlock() // ブロックをパース

	return stmt, err
}

func (p *Parser) parseReturn() (ast.Stmt, error) {
	n := p.next() // return
	stmt := &ast.ReturnStmt{
		Pos: n.Pos,
	}
	var err error
	stmt.Val, err = p.parseExpr(ast.Lowest)
	return stmt, err
}

func (p *Parser) parseCtrl() (ast.Stmt, error) {
	n := p.next()
	stmt := &ast.CtrlStmt{
		Pos: n.Pos,
	}
	switch n.Typ {
	case ast.KeywordBreak:
		stmt.Typ = ast.CtrlBreak
	case ast.KeywordContinue:
		stmt.Typ = ast.CtrlContinue
	case ast.KeywordFallthrough:
		stmt.Typ = ast.CtrlFallthrough
	default:
		return nil, fmt.Errorf("ln%v: expecting ctrl token, got %v", n.Line, n.Val)
	}
	return stmt, nil
}

func (p *Parser) parseCall(fun ast.Expr) (ast.Expr, error) {
	// (引数) を期待
	n, err := p.consume(ast.ItemLeftParen)
	if err != nil {
		return nil, fmt.Errorf("expecting ( after ident, got %v", fun.String())
	}
	expr := &ast.CallExpr{
		Pos: n.Pos,
		Fun: fun,
	}
	expr.Args, err = p.parseCallArgs()

	return expr, err
}

func (p *Parser) parseCallArgs() ([]ast.Expr, error) {
	var args []ast.Expr

	if p.peek().Typ == ast.ItemRightParen {
		// 右括弧を消費
		p.next()
		return args, nil
	}

	// 次は式であるべき
	exp, err := p.parseExpr(ast.Lowest)
	if err != nil {
		return args, err
	}
	args = append(args, exp)

	for p.peek().Typ == ast.ItemComma {
		p.next() // skip the comma
		exp, err = p.parseExpr(ast.Lowest)
		if err != nil {
			return args, err
		}
		args = append(args, exp)
	}

	if n := p.next(); n.Typ != ast.ItemRightParen {
		p.backup()
		return nil, fmt.Errorf("ln%v: expecting ) at end of function call, got: %v", n.Line, n.Pos)
	}

	return args, nil
}

// 現在のトークンが "character" の前提で、有効なキャラクターアクションか確認
func (p *Parser) peekValidCharAction() bool {
	p.next()
	// キャラクターのステータス等かアクションかを確認
	if p.peek().Typ != ast.ItemActionKey {
		p.backup()
		// ActionStmt ではない
		return false
	}
	p.backup()
	return true
}

// parseBlock は BlockStmt を含むノードを返す
func (p *Parser) parseBlock() (*ast.BlockStmt, error) {
	// {} で囲まれているべき
	n, err := p.consume(ast.ItemLeftBrace)
	if err != nil {
		return nil, fmt.Errorf("ln%v: expecting {, got %v", n.Line, n.Val)
	}
	block := ast.NewBlockStmt(n.Pos)
	var node ast.Node
	// } に到達するまで1行ずつパース
	for {
		// 不正な行がないことを確認
		switch n := p.peek(); n.Typ {
		case ast.ItemCharacterKey:
			if !p.peekValidCharAction() {
				n = p.next()
				return nil, fmt.Errorf("ln%v: expecting action after character token, got %v", n.Line, n.Val)
			}
		case ast.ItemRightBrace:
			p.next() // 中括弧を消費
			return block, nil
		case ast.ItemEOF:
			return nil, fmt.Errorf("reached end of file without closing }")
		}
		// ここで文をパース
		node, err = p.parseStatement()
		if err != nil {
			return nil, err
		}
		block.Append(node)
	}
}
func (p *Parser) parseExpr(pre ast.Precedence) (ast.Expr, error) {
	t := p.next()
	prefix := p.prefixParseFns[t.Typ]
	if prefix == nil {
		return nil, fmt.Errorf("ln%v: no prefix parse function for %v", t.Line, t.Val)
	}
	p.backup()
	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}

	for n := p.peek(); n.Typ != ast.ItemTerminateLine && pre < n.Precedence(); n = p.peek() {
		infix := p.infixParseFns[n.Typ]
		if infix == nil {
			break
		}

		leftExp, err = infix(leftExp)
		if err != nil {
			return nil, err
		}
	}

	if p.constantFolding {
		leftExp, err = foldConstants(leftExp)
		if err != nil {
			return nil, err
		}
	}

	return leftExp, nil
}

// 次は識別子
func (p *Parser) parseIdent() (ast.Expr, error) {
	n := p.next()
	return &ast.Ident{Pos: n.Pos, Value: n.Val}, nil
}

func (p *Parser) parseField() (ast.Expr, error) {
	// 次はフィールド。フィールドが続く限りパースし続ける
	// その後すべてを連結する
	n := p.next()
	fields := make([]string, 0, 5)
	for ; n.Typ == ast.ItemField; n = p.next() {
		fields = append(fields, strings.Trim(n.Val, "."))
	}
	// ここで1つ多く消費しているため戻す
	p.backup()
	return &ast.Field{Pos: n.Pos, Value: fields}, nil
}

func (p *Parser) parseString() (ast.Expr, error) {
	n := p.next()
	return &ast.StringLit{Pos: n.Pos, Value: n.Val}, nil
}

func (p *Parser) parseNumber() (ast.Expr, error) {
	// 文字列、整数、浮動小数点数、またはブール値
	n := p.next()
	num := &ast.NumberLit{Pos: n.Pos}
	// 整数としてパースを試み、失敗したら浮動小数点数を試す
	iv, err := strconv.ParseInt(n.Val, 10, 64)
	if err == nil {
		num.IntVal = iv
		num.FloatVal = float64(iv)
	} else {
		fv, err := strconv.ParseFloat(n.Val, 64)
		if err != nil {
			return nil, fmt.Errorf("ln%v: cannot parse %v to number", n.Line, n.Val)
		}
		num.IsFloat = true
		num.FloatVal = fv
	}
	return num, nil
}

func (p *Parser) parseBool() (ast.Expr, error) {
	// ブール値は数値 (true = 1, false = 0)
	n := p.next()
	num := &ast.NumberLit{Pos: n.Pos}
	switch n.Val {
	case ast.TrueVal:
		num.IntVal = 1
		num.FloatVal = 1
	case ast.FalseVal:
		num.IntVal = 0
		num.FloatVal = 0
	default:
		return nil, fmt.Errorf("ln%v: expecting boolean, got %v", n.Line, n.Val)
	}
	return num, nil
}

func (p *Parser) parseUnaryExpr() (ast.Expr, error) {
	n := p.next()
	switch n.Typ {
	case ast.LogicNot:
	case ast.ItemMinus:
	default:
		return nil, fmt.Errorf("ln%v: unrecognized unary operator %v", n.Line, n.Val)
	}
	var err error
	expr := &ast.UnaryExpr{
		Pos: n.Pos,
		Op:  n,
	}
	expr.Right, err = p.parseExpr(ast.Prefix)
	return expr, err
}

func (p *Parser) parseBinaryExpr(left ast.Expr) (ast.Expr, error) {
	n := p.next()
	expr := &ast.BinaryExpr{
		Pos:  n.Pos,
		Op:   n,
		Left: left,
	}
	pr := n.Precedence()
	var err error
	expr.Right, err = p.parseExpr(pr)
	return expr, err
}

func (p *Parser) parseParen() (ast.Expr, error) {
	// 括弧をスキップ
	p.next()

	exp, err := p.parseExpr(ast.Lowest)
	if err != nil {
		return nil, err
	}

	if n := p.peek(); n.Typ != ast.ItemRightParen {
		return nil, nil
	}
	p.next() // 右括弧を消費

	return exp, nil
}

func (p *Parser) parseMap() (ast.Expr, error) {
	// 括弧をスキップ
	n := p.next()
	expr := &ast.MapExpr{Pos: n.Pos}

	if p.peek().Typ == ast.ItemRightSquareParen { // 空のマップ
		p.next()
		return expr, nil
	}

	expr.Fields = make(map[string]ast.Expr)
	// 閉じ角括弧まで繰り返す
	for {
		// ident = int の形式を期待
		i, err := p.consume(ast.ItemIdentifier)
		if err != nil {
			return nil, fmt.Errorf("ln%v: expecting identifier in map expression, got %v", i.Line, i.Val)
		}

		a, err := p.consume(ast.ItemAssign)
		if err != nil {
			return nil, fmt.Errorf("ln%v: expecting = after identifier in map expression, got %v", a.Line, a.Val)
		}

		e, err := p.parseExpr(ast.Lowest)
		if err != nil {
			return nil, err
		}
		expr.Fields[i.Val] = e

		// ] なら返す。, なら続行。それ以外はエラー
		n := p.next()
		switch n.Typ {
		case ast.ItemRightSquareParen:
			return expr, nil
		case ast.ItemComma:
			// 何もしない、続行
		default:
			return nil, fmt.Errorf("ln%v: <action param> bad token %v", n.Line, n)
		}
	}
}
