// Code generated by "pipeline"; DO NOT EDIT.
package citlali

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
	2: {"travel"},
	3: {"travel"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Citlali, ValidateParamKeys)
}

func ValidateParamKeys(a action.Action, keys []string) error {
	valid, ok := paramKeysValidation[a]
	if !ok {
		return nil
	}
	for _, v := range keys {
		if !slices.Contains(valid, v) {
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
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.51396,
		0.552507,
		0.591054,
		0.64245,
		0.680997,
		0.719544,
		0.77094,
		0.822336,
		0.873732,
		0.925128,
		0.976524,
		1.02792,
		1.092165,
		1.15641,
		1.220655,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.446256,
		0.479725,
		0.513194,
		0.55782,
		0.591289,
		0.624758,
		0.669384,
		0.71401,
		0.758635,
		0.803261,
		0.847886,
		0.892512,
		0.948294,
		1.004076,
		1.059858,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.700344,
		0.75287,
		0.805396,
		0.87543,
		0.927956,
		0.980482,
		1.050516,
		1.12055,
		1.190585,
		1.260619,
		1.330654,
		1.400688,
		1.488231,
		1.575774,
		1.663317,
	}
	// attack: charge = [3]
	charge = []float64{
		1.4288,
		1.53596,
		1.64312,
		1.786,
		1.89316,
		2.00032,
		2.1432,
		2.28608,
		2.42896,
		2.57184,
		2.71472,
		2.8576,
		3.0362,
		3.2148,
		3.3934,
	}
	// skill: bite = [0]
	bite = []float64{
		0.0868,
		0.09331,
		0.09982,
		0.1085,
		0.11501,
		0.12152,
		0.1302,
		0.13888,
		0.14756,
		0.15624,
		0.16492,
		0.1736,
		0.18445,
		0.1953,
		0.20615,
	}
	// skill: momentumBonus = [1]
	momentumBonus = []float64{
		0.0434,
		0.046655,
		0.04991,
		0.05425,
		0.057505,
		0.06076,
		0.0651,
		0.06944,
		0.07378,
		0.07812,
		0.08246,
		0.0868,
		0.092225,
		0.09765,
		0.103075,
	}
	// skill: surgingBite = [2]
	surgingBite = []float64{
		0.217,
		0.233275,
		0.24955,
		0.27125,
		0.287525,
		0.3038,
		0.3255,
		0.3472,
		0.3689,
		0.3906,
		0.4123,
		0.434,
		0.461125,
		0.48825,
		0.515375,
	}
	// burst: burst = [0]
	burst = []float64{
		0.584392,
		0.628221,
		0.672051,
		0.73049,
		0.774319,
		0.818149,
		0.876588,
		0.935027,
		0.993466,
		1.051906,
		1.110345,
		1.168784,
		1.241833,
		1.314882,
		1.387931,
	}
)
