package ast

import "fmt"

// Token はスキャナから返されるトークンまたはテキスト文字列を表す。
type Token struct {
	Typ  TokenType // このアイテムの型。
	Pos  Pos       // 入力文字列内でのこのアイテムの開始位置（バイト単位）。
	Val  string    // このアイテムの値。
	Line int       // このアイテムの開始行番号。
}

func (t Token) String() string {
	switch {
	case t.Typ == ItemEOF:
		return "EOF"
	case t.Typ == ItemError:
		return t.Val
	case t.Typ == ItemTerminateLine:
		return ";"
	case t.Typ > ItemTerminateLine && t.Typ < ItemKeyword:
		return t.Val
	case t.Typ > ItemKeyword:
		return fmt.Sprintf("<%s>", t.Val)
		// case len(i.val) > 10:
		// 	return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", t.Val)
}

// TokenType はトークンの種類を識別する。
type TokenType int

const (
	ItemError TokenType = iota // エラー発生; 値はエラーテキスト

	ItemEOF
	ItemTerminateLine    // 行末を示す \n
	ItemAssign           // 代入を導入する等号 ('=')
	ItemComma            // 識別子リストを区切るカンマ (',')
	ItemLeftParen        // '('
	ItemRightParen       // ')'
	ItemLeftSquareParen  // '['
	ItemRightSquareParen // ']'
	ItemLeftBrace        // '{'
	ItemRightBrace       // '}'
	ItemColon            // ':'
	ItemPlus             // '+'
	ItemMinus            // '-'
	ItemAsterisk         // '*'
	ItemForwardSlash     // '/'
	// 以下は論理演算子
	ItemLogicOP // 論理演算の境界としてのみ使用
	LogicNot    // !
	LogicAnd    // && キーワード
	LogicOr     // || キーワード
	// 以下は比較演算子
	itemCompareOp        // 比較演算子の境界としてのみ使用
	OpEqual              // == キーワード
	OpNotEqual           // != キーワード
	OpGreaterThan        // > キーワード
	OpGreaterThanOrEqual // >= キーワード
	OpLessThan           // < キーワード
	OpLessThanOrEqual    // <= キーワード
	ItemDot              // カーソル '.'
	// アイテム型
	ItemTypes
	ItemField      // '.' で始まる英数字識別子
	ItemIdentifier // '.' で始まらない英数字識別子
	ItemNumber     // 単純な数値
	ItemBool       // ブーリアン
	ItemString     // 引用符を含む文字列
	// キーワードは以下の後に続く。
	ItemKeyword        // キーワードの境界としてのみ使用
	KeywordLet         // let
	KeywordWhile       // while
	KeywordIf          // if
	KeywordElse        // else
	KeywordFn          // fn
	KeywordSwitch      // switch
	KeywordCase        // ケース
	KeywordDefault     // デフォルト
	KeywordBreak       // break
	KeywordContinue    // continue
	KeywordFallthrough // fallthrough
	KeywordReturn      // return
	KeywordFor         // for
	// これ以降のキーワードは原神固有（汎用スクリプトキーワードではない）
	// キャラクター関連の特殊キーワード
	KeywordOptions           // options
	KeywordAdd               // 追加
	KeywordChar              // キャラクター
	KeywordStats             // stats
	KeywordWeapon            // weapon
	KeywordSet               // set
	KeywordLvl               // lvl
	KeywordRefine            // refine
	KeywordCons              // 命ノ星座
	KeywordTalent            // talent
	KeywordCount             // count
	KeywordParams            // params
	KeywordLabel             // label
	KeywordUntil             // until
	KeywordActive            // アクティブ
	KeywordTarget            // ターゲット
	KeywordResist            // resist
	KeywordEnergy            // energy
	KeywordParticleThreshold // particle_threshold
	KeywordParticleDropCount // particle_drop_count
	KeywordParticleElement   // particle_element
	KeywordHurt              // hurt

	// gcsim 固有のキーワードはこれ以降に続く
	ItemKeys
	ItemStatKey      // stats: def%, def, etc..
	ItemElementKey   // elements: pyro, hydro, etc..
	ItemCharacterKey // characters: albedo, amber, etc..
	ItemActionKey    // actions: skill, burst, attack, charge, etc...
)

type Precedence int

const (
	_ Precedence = iota
	Lowest
	LogicalOr
	LogicalAnd // TODO: && と || 用に別の優先度を作るべき？
	Equals
	LessOrGreater
	Sum
	Product
	Prefix
	Call
)

var precedences = map[TokenType]Precedence{
	LogicOr:              LogicalOr,
	LogicAnd:             LogicalAnd,
	OpEqual:              Equals,
	OpNotEqual:           Equals,
	OpLessThan:           LessOrGreater,
	OpGreaterThan:        LessOrGreater,
	OpLessThanOrEqual:    LessOrGreater,
	OpGreaterThanOrEqual: LessOrGreater,
	ItemPlus:             Sum,
	ItemMinus:            Sum,
	ItemForwardSlash:     Product,
	ItemAsterisk:         Product,
	ItemLeftParen:        Call,
}

func (t Token) Precedence() Precedence {
	if p, ok := precedences[t.Typ]; ok {
		return p
	}
	return Lowest
}
