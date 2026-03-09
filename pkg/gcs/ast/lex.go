package ast

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/shortcut"
)

const eof = -1

type stateFn func(*Lexer) stateFn

// Lexer はスキャナの状態を保持する。
type Lexer struct {
	input        string     // スキャン対象の文字列
	pos          Pos        // 入力内の現在位置
	start        Pos        // このアイテムの開始位置
	width        Pos        // 最後に読んだルーンの幅
	items        chan Token // スキャン済みアイテムのチャネル
	line         int        // 検出した改行数+1
	startLine    int        // このアイテムの開始行
	parenDepth   int
	sqParenDepth int
	braceDepth   int
}

// next は入力内の次のルーンを返す。
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek は入力内の次のルーンを返すが、消費しない。
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup はルーンを1つ戻す。next の呼び出しごとに1回だけ呼べる。
func (l *Lexer) backup() {
	l.pos -= l.width
	// 改行カウントを修正する。
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit はアイテムをクライアントに返す。
func (l *Lexer) emit(t TokenType) {
	l.items <- Token{
		Typ:  t,
		Pos:  l.start,
		Val:  l.input[l.start:l.pos],
		Line: l.startLine,
	}
	l.start = l.pos
	l.startLine = l.line
}

// ignore はこの時点までの保留中の入力をスキップする。
func (l *Lexer) ignore() {
	// l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept は次のルーンが有効な文字セットに含まれていれば消費する。
func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun は有効な文字セットからルーンの連続を消費する。
func (l *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf はエラートークンを返し、次の状態として nil ポインタを返すことで
// スキャンを終了し、l.nextItem を終了させる。
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Token{
		Typ:  ItemError,
		Pos:  l.start,
		Val:  fmt.Sprintf(format, args...),
		Line: l.startLine,
	}
	return nil
}

// NextItem は入力から次のアイテムを返す。
// パーサーから呼ばれ、字句解析ゴルーチンからは呼ばれない。
func (l *Lexer) NextItem() Token {
	return <-l.items
}

// NewLexer は入力文字列に対する新しいスキャナを作成する。
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:     input,
		items:     make(chan Token),
		line:      1,
		startLine: 1,
	}
	go l.run()
	return l
}

// run はレキサーのステートマシンを実行する。
func (l *Lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// lexText は開始アクション区切り文字 "{{" までスキャンする。
func lexText(l *Lexer) stateFn {
	// 数値、クォート文字列、または識別子のいずれか。
	// スペースは引数を区切り、連続するスペースは itemSpace になる。
	// パイプ記号は区切りとして発行される。
	// n := l.peek()
	// log.Printf("lexText next is %c\n", n)
	switch r := l.next(); {
	case r == eof:
		l.emit(ItemEOF)
		return nil
	case r == ';':
		l.emit(ItemTerminateLine)
	case r == ':':
		l.emit(ItemColon)
	case isSpace(r):
		l.ignore()
	case r == '#':
		l.ignore()
		return lexComment
	case r == '=':
		n := l.next()
		if n == '=' {
			l.emit(OpEqual)
		} else {
			l.backup()
			l.emit(ItemAssign)
		}
	case r == ',':
		l.emit(ItemComma)
	case r == '*':
		l.emit(ItemAsterisk)
	case r == '+':
		// //次のアイテムが数値かどうか確認し、数値なら lexNumber
		// //そうでなければ + 記号
		// n := l.next()
		// if isNumeric(n) {
		// 	//2つ戻る
		// 	l.backup()
		// 	l.backup()
		// 	return lexNumber
		// }
		// //そうでなければプラス記号
		// l.backup()
		l.emit(ItemPlus)
	case r == '/':
		// 次が / かどうか確認し、/ なら lexComment
		n := l.next()
		if n == '/' {
			l.ignore()
			return lexComment
		}
		l.backup()
		l.emit(ItemForwardSlash)
	case r == '.':
		// ".field" のための特別な先読み。l.backup() を壊さないようにする。
		if l.pos < Pos(len(l.input)) {
			r := l.input[l.pos]
			if r < '0' || '9' < r {
				return lexField
			}
		}
		fallthrough // '.' は数値の開始になりうる。
	case ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case r == '-':
		// 次のアイテムが数値なら数値を解析
		// n := l.next()
		// if isNumeric(n) {
		// 	// 2つ戻る
		// 	l.backup()
		// 	l.backup()
		// 	return lexNumber
		// }
		// // そうでなければ - 記号
		// l.backup()
		l.emit(ItemMinus)
	case r == '>':
		if n := l.next(); n == '=' {
			l.emit(OpGreaterThanOrEqual)
		} else {
			l.backup()
			l.emit(OpGreaterThan)
		}
	case r == '<':
		switch n := l.next(); n {
		case '=':
			l.emit(OpLessThanOrEqual)
		case '>':
			l.emit(OpNotEqual)
		default:
			l.backup()
			l.emit(OpLessThan)
		}
	case r == '|':
		if n := l.next(); n == '|' {
			l.emit(LogicOr)
		} else {
			return l.errorf("unrecognized character in action: %#U", r)
		}
	case r == '!':
		if n := l.next(); n == '=' {
			l.emit(OpNotEqual)
		} else {
			l.backup()
			l.emit(LogicNot)
		}
	case r == '"':
		return lexQuote
	case r == '&':
		if n := l.next(); n == '&' {
			l.emit(LogicAnd)
		} else {
			return l.errorf("unrecognized character in action: %#U", r)
		}
	case r == '(':
		l.emit(ItemLeftParen)
		l.parenDepth++
	case r == ')':
		l.emit(ItemRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren %#U", r)
		}
	case r == '[':
		l.emit(ItemLeftSquareParen)
		l.sqParenDepth++
	case r == ']':
		l.emit(ItemRightSquareParen)
		l.sqParenDepth--
		if l.sqParenDepth < 0 {
			return l.errorf("unexpected right sq paren %#U", r)
		}
	case r == '{':
		l.emit(ItemLeftBrace)
		l.braceDepth++
	case r == '}':
		l.emit(ItemRightBrace)
		l.braceDepth--
		if l.braceDepth < 0 {
			return l.errorf("unexpected right brace %#U", r)
		}
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
	return lexText
}

func lexComment(l *Lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case eof, '\n':
			l.backup()
			break Loop
		default:
			// 吸収する
		}
	}
	// l.emit(itemComment)
	return lexText
}

// lexField はフィールドをスキャンする: .Alphanumeric。
// . は既にスキャン済み。
func lexField(l *Lexer) stateFn {
	if l.atTerminator() { // 興味深いものが続かない -> "." または "$"。
		l.emit(ItemDot)
		return lexText
	}
	var r rune
	for {
		r = l.next()
		if !isAlphaNumeric(r) {
			l.backup()
			break
		}
	}
	if !l.atTerminator() {
		return l.errorf("bad character %#U", r)
	}
	l.emit(ItemField)
	return lexText
}

func lexQuote(l *Lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}
	l.emit(ItemString)
	return lexText
}

// lexIdentifier は英数字をスキャンする。
func lexIdentifier(l *Lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// 吸収する。
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}
			switch {
			case key[word] > ItemKeyword:
				l.emit(key[word])
			case word[0] == '.':
				l.emit(ItemField)
			case word == TrueVal, word == FalseVal:
				l.emit(ItemBool)
			default:
				l.emit(checkIdentifier(word))
			}
			break Loop
		}
	}
	return lexText
}

func checkIdentifier(word string) TokenType {
	if _, ok := StatKeys[word]; ok {
		return ItemStatKey
	}
	if _, ok := EleKeys[word]; ok {
		return ItemElementKey
	}
	if _, ok := shortcut.CharNameToKey[word]; ok {
		return ItemCharacterKey
	}
	if _, ok := actionKeys[word]; ok {
		return ItemActionKey
	}
	return ItemIdentifier
}

func lexNumber(l *Lexer) stateFn {
	// オプションの先頭符号。
	l.accept("+-")

	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}

	l.emit(ItemNumber)

	return lexText
}

// isSpace は r が空白文字かどうかを報告する。
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// isAlphaNumeric は r がアルファベット、数字、またはアンダースコアかどうかを報告する。
func isAlphaNumeric(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) || r == '%'
}

// isNumeric は r が数字かどうかを報告する
// func isNumeric(r rune) bool {
// 	return unicode.IsDigit(r)
// }

// atTerminator は入力が識別子の後に続く有効な終端文字かどうかを報告する。
// .X.Y を2つに分割する。また "$x+2" のようにスペースなしでは
// 受け入れられないケースも捕捉する（将来算術演算を実装する場合に備えて）。
func (l *Lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) {
		return true
	}
	switch r {
	case eof, '.', ',', '|', ':', ')', '(', '+', '=', '>', '<', '&', '!', ';', '[', ']', '{', '}':
		return true
	}
	return false
}
