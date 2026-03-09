package ast

import (
	"fmt"
	"testing"
)

func TestFields(t *testing.T) {
	input := `if .status.field > 0 { print("hi") };`

	l := NewLexer(input)
	for n := l.NextItem(); n.Typ != ItemEOF; n = l.NextItem() {
		fmt.Println(n)
	}
}
func TestBasicToken(t *testing.T) {
	input := `
	let y = fn(x) {
		return x + 1;
	}
	let x = 5;
	label A:
	while {
		#comment
		x = y(x);
		if x > 10 {
			break A;
		}
		//コメント
		switch x {
		case 1:
			fallthrough;
		case 2:
			fallthrough;
		case 3:
			break A;
		}
	}
	
	for x = 0; x < 5; x = x + 1 {
		let i = y(x);
	}

	-1
	1
	-
	-a
	`

	expected := []Token{
		// 関数
		{Typ: KeywordLet, Val: "let"},
		{Typ: ItemIdentifier, Val: "y"},
		{Typ: ItemAssign, Val: "="},
		{Typ: KeywordFn, Val: "fn"},
		{Typ: ItemLeftParen, Val: "("},
		{Typ: ItemIdentifier, Val: "x"},
		// {typ: typeNum, Val: "num"},
		{Typ: ItemRightParen, Val: ")"},
		// {typ: typeNum, Val: "num"}
		{Typ: ItemLeftBrace, Val: "{"},
		{Typ: KeywordReturn, Val: "return"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemPlus, Val: "+"},
		{Typ: ItemNumber, Val: "1"},
		{Typ: ItemTerminateLine, Val: ";"},
		{Typ: ItemRightBrace, Val: "}"},
		// 変数
		{Typ: KeywordLet, Val: "let"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemAssign, Val: "="},
		{Typ: ItemNumber, Val: "5"},
		{Typ: ItemTerminateLine, Val: ";"},
		// ラベル
		{Typ: KeywordLabel, Val: "label"},
		{Typ: ItemIdentifier, Val: "A"},
		{Typ: ItemColon, Val: ":"},
		// while ループ
		{Typ: KeywordWhile, Val: "while"},
		{Typ: ItemLeftBrace, Val: "{"},
		// コメント
		// {typ: itemComment, Val: "comment"},
		// 関数呼び出し
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemAssign, Val: "="},
		{Typ: ItemIdentifier, Val: "y"},
		{Typ: ItemLeftParen, Val: "("},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemRightParen, Val: ")"},
		{Typ: ItemTerminateLine, Val: ";"},
		// if 文
		{Typ: KeywordIf, Val: "if"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: OpGreaterThan, Val: ">"},
		{Typ: ItemNumber, Val: "10"},
		{Typ: ItemLeftBrace, Val: "{"},
		// break 文
		{Typ: KeywordBreak, Val: "break"},
		{Typ: ItemIdentifier, Val: "A"},
		{Typ: ItemTerminateLine, Val: ";"},
		// if 終了
		{Typ: ItemRightBrace, Val: "}"},
		// コメント
		// {typ: itemComment, Val: "comment"},
		// switch 文
		{Typ: KeywordSwitch, Val: "switch"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemLeftBrace, Val: "{"},
		// ケース
		{Typ: KeywordCase, Val: "case"},
		{Typ: ItemNumber, Val: "1"},
		{Typ: ItemColon, Val: ":"},
		{Typ: KeywordFallthrough, Val: "fallthrough"},
		{Typ: ItemTerminateLine, Val: ";"},
		// ケース
		{Typ: KeywordCase, Val: "case"},
		{Typ: ItemNumber, Val: "2"},
		{Typ: ItemColon, Val: ":"},
		{Typ: KeywordFallthrough, Val: "fallthrough"},
		{Typ: ItemTerminateLine, Val: ";"},
		// ケース
		{Typ: KeywordCase, Val: "case"},
		{Typ: ItemNumber, Val: "3"},
		{Typ: ItemColon, Val: ":"},
		{Typ: KeywordBreak, Val: "break"},
		{Typ: ItemIdentifier, Val: "A"},
		{Typ: ItemTerminateLine, Val: ";"},
		// switch 終了
		{Typ: ItemRightBrace, Val: "}"},
		// while 終了
		{Typ: ItemRightBrace, Val: "}"},
		// for ループ
		{Typ: KeywordFor, Val: "for"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemAssign, Val: "="},
		{Typ: ItemNumber, Val: "0"},
		{Typ: ItemTerminateLine, Val: ";"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: OpLessThan, Val: "<"},
		{Typ: ItemNumber, Val: "5"},
		{Typ: ItemTerminateLine, Val: ";"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemAssign, Val: "="},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemPlus, Val: "+"},
		{Typ: ItemNumber, Val: "1"},
		{Typ: ItemLeftBrace, Val: "{"},
		// 本体
		{Typ: KeywordLet, Val: "let"},
		{Typ: ItemIdentifier, Val: "i"},
		{Typ: ItemAssign, Val: "="},
		{Typ: ItemIdentifier, Val: "y"},
		{Typ: ItemLeftParen, Val: "("},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: ItemRightParen, Val: ")"},
		{Typ: ItemTerminateLine, Val: ";"},
		// for 終了
		{Typ: ItemRightBrace, Val: "}"},
		// その他のテスト
		{Typ: ItemMinus, Val: "-"},
		{Typ: ItemNumber, Val: "1"},
		{Typ: ItemNumber, Val: "1"},
		{Typ: ItemMinus, Val: "-"},
		{Typ: ItemMinus, Val: "-"},
		{Typ: ItemIdentifier, Val: "a"},
	}

	l := NewLexer(input)
	i := 0
	for n := l.NextItem(); n.Typ != ItemEOF; n = l.NextItem() {
		if expected[i].Typ != n.Typ && expected[i].Val != n.Val {
			t.Errorf("expected %v got %v", expected[i], n)
		}
		if i < len(expected)-1 {
			i++
		}
	}
}

func TestElseSpace(t *testing.T) {
	input := `
	if x > 1{}else{}
	`

	expected := []Token{
		// 関数
		{Typ: KeywordIf, Val: "if"},
		{Typ: ItemIdentifier, Val: "x"},
		{Typ: OpGreaterThan, Val: ">"},
		{Typ: ItemNumber, Val: "1"},
		{Typ: ItemLeftBrace, Val: "{"},
		{Typ: ItemRightBrace, Val: "}"},
		{Typ: KeywordElse, Val: "else"},
		{Typ: ItemLeftBrace, Val: "{"},
		{Typ: ItemRightBrace, Val: "}"},
	}

	l := NewLexer(input)
	i := 0
	for n := l.NextItem(); n.Typ != ItemEOF; n = l.NextItem() {
		if expected[i].Typ != n.Typ && expected[i].Val != n.Val {
			t.Errorf("expected %v got %v", expected[i], n)
			t.FailNow()
		}
		if i < len(expected)-1 {
			i++
		}
	}
}
