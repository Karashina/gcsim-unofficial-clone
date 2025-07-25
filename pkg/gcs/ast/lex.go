package ast

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/genshinsim/gcsim/pkg/shortcut"
)

const eof = -1

type stateFn func(*Lexer) stateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	input        string     // the string being scanned
	pos          Pos        // current position in the input
	start        Pos        // start position of this item
	width        Pos        // width of last rune read from input
	items        chan Token // channel of scanned items
	line         int        // 1+number of newlines seen
	startLine    int        // start line of this item
	parenDepth   int
	sqParenDepth int
	braceDepth   int
}

// next returns the next rune in the input.
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

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit passes an item back to the client.
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

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	// l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Token{
		Typ:  ItemError,
		Pos:  l.start,
		Val:  fmt.Sprintf(format, args...),
		Line: l.startLine,
	}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *Lexer) NextItem() Token {
	return <-l.items
}

// lex creates a new scanner for the input string.
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

// run runs the state machine for the lexer.
func (l *Lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// lexText scans until an opening action delimiter, "{{".
func lexText(l *Lexer) stateFn {
	// Either number, quoted string, or identifier.
	// Spaces separate arguments; runs of spaces turn into itemSpace.
	// Pipe symbols separate and are emitted.
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
		// //check if next item is a number or not; if number lexNumber
		// //otherwise it's a + sign
		// n := l.next()
		// if isNumeric(n) {
		// 	//back up twice
		// 	l.backup()
		// 	l.backup()
		// 	return lexNumber
		// }
		// //otherwise it's a plus sign
		// l.backup()
		l.emit(ItemPlus)
	case r == '/':
		// check if next is another / or not; if / then lexComment
		n := l.next()
		if n == '/' {
			l.ignore()
			return lexComment
		}
		l.backup()
		l.emit(ItemForwardSlash)
	case r == '.':
		// special look-ahead for ".field" so we don't break l.backup().
		if l.pos < Pos(len(l.input)) {
			r := l.input[l.pos]
			if r < '0' || '9' < r {
				return lexField
			}
		}
		fallthrough // '.' can start a number.
	case ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case r == '-':
		// if next item is a number then lex number
		// n := l.next()
		// if isNumeric(n) {
		// 	// backup twice
		// 	l.backup()
		// 	l.backup()
		// 	return lexNumber
		// }
		// // other wise it's a - sign
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
			// absorb
		}
	}
	// l.emit(itemComment)
	return lexText
}

// lexField scans a field: .Alphanumeric.
// The . has been scanned.
func lexField(l *Lexer) stateFn {
	if l.atTerminator() { // Nothing interesting follows -> "." or "$".
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

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *Lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
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
	// Optional leading sign.
	l.accept("+-")

	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}

	l.emit(ItemNumber)

	return lexText
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) || r == '%'
}

// is Numeric reports whether r is a digit
// func isNumeric(r rune) bool {
// 	return unicode.IsDigit(r)
// }

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier. Breaks .X.Y into two pieces. Also catches cases
// like "$x+2" not being acceptable without a space, in case we decide one
// day to implement arithmetic.
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
