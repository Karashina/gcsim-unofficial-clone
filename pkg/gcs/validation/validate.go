package validation

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
)

func ValidateCharParamKeys(c keys.Char, a action.Action, keys []string) error {
	f, ok := charValidParamKeys[c]
	if !ok {
		// バリデーション関数が登録されていなければすべてOK
		return nil
	}
	return f(a, keys)
}
