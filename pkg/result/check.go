package result

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

// このリストにキャラクターを追加すると、ビューアーに「不完全な警告」が表示される
var incompleteCharacters = []keys.Char{
	keys.TestCharDoNotUse,
}

func IsCharacterComplete(char keys.Char) bool {
	for _, v := range incompleteCharacters {
		if v == char {
			return false
		}
	}
	return true
}
