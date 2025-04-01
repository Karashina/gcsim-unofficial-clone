// Code generated by "pipeline"; DO NOT EDIT.
package razor

import (
	_ "embed"

	"fmt"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/gcs/validation"
	"github.com/genshinsim/gcsim/pkg/model"
	"google.golang.org/protobuf/encoding/prototext"
	"slices"
)

//go:embed data_gen.textproto
var pbData []byte
var base *model.AvatarData
var paramKeysValidation = map[action.Action][]string{
	1: {"hold"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Razor, ValidateParamKeys)
}

func ValidateParamKeys(a action.Action, keys []string) error {
	valid, ok := paramKeysValidation[a]
	if !ok {
		return nil
	}
	for _, v := range keys {
		if !slices.Contains(valid, v) {
			if v == "movement" {
				return nil
			}
			return fmt.Errorf("key %v is invalid for action %v", v, a.String())
		}
	}
	return nil
}

func (x *char) Data() *model.AvatarData {
	return base
}

var (
	attack = [][]float64{
		attack_1,
		attack_2,
		attack_3,
		attack_4,
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.9592,
		1.0246,
		1.09,
		1.1772,
		1.2426,
		1.3189,
		1.417,
		1.5151,
		1.6132,
		1.7113,
		1.8094,
		1.9075,
		2.0056,
		2.1037,
		2.2018,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.82632,
		0.88266,
		0.939,
		1.01412,
		1.07046,
		1.13619,
		1.2207,
		1.30521,
		1.38972,
		1.47423,
		1.55874,
		1.64325,
		1.72776,
		1.81227,
		1.89678,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		1.03312,
		1.10356,
		1.174,
		1.26792,
		1.33836,
		1.42054,
		1.5262,
		1.63186,
		1.73752,
		1.84318,
		1.94884,
		2.0545,
		2.16016,
		2.26582,
		2.37148,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		1.36048,
		1.45324,
		1.546,
		1.66968,
		1.76244,
		1.87066,
		2.0098,
		2.14894,
		2.28808,
		2.42722,
		2.56636,
		2.7055,
		2.84464,
		2.98378,
		3.12292,
	}
	// skill: skillHold = [1]
	skillHold = []float64{
		2.952,
		3.1734,
		3.3948,
		3.69,
		3.9114,
		4.1328,
		4.428,
		4.7232,
		5.0184,
		5.3136,
		5.6088,
		5.904,
		6.273,
		6.642,
		7.011,
	}
	// skill: skillPress = [0]
	skillPress = []float64{
		1.992,
		2.1414,
		2.2908,
		2.49,
		2.6394,
		2.7888,
		2.988,
		3.1872,
		3.3864,
		3.5856,
		3.7848,
		3.984,
		4.233,
		4.482,
		4.731,
	}
	// burst: burstATKSpeed = [2]
	burstATKSpeed = []float64{
		0.26,
		0.28,
		0.3,
		0.32,
		0.34,
		0.36,
		0.37,
		0.38,
		0.39,
		0.4,
		0.4,
		0.4,
		0.4,
		0.4,
		0.4,
	}
	// burst: burstDmg = [0]
	burstDmg = []float64{
		1.6,
		1.72,
		1.84,
		2,
		2.12,
		2.24,
		2.4,
		2.56,
		2.72,
		2.88,
		3.04,
		3.2,
		3.4,
		3.6,
		3.8,
	}
	// burst: wolfDmg = [1]
	wolfDmg = []float64{
		0.24,
		0.258,
		0.276,
		0.3,
		0.318,
		0.336,
		0.36,
		0.384,
		0.408,
		0.432,
		0.456,
		0.48,
		0.51,
		0.54,
		0.57,
	}
)
