// Code generated by "pipeline"; DO NOT EDIT.
package emilie

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
	1: {"recast"},
	3: {"travel"},
	6: {"collision"},
	7: {"collision"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Emilie, ValidateParamKeys)
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
	attack = [][][]float64{
		{attack_1},
		{attack_2},
		{attack_3},
		{attack_4},
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.486,
		0.525366,
		0.565218,
		0.621594,
		0.66096,
		0.706158,
		0.768366,
		0.830574,
		0.892782,
		0.960822,
		1.028376,
		1.096416,
		1.16397,
		1.16397,
		1.16397,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.449,
		0.485369,
		0.522187,
		0.574271,
		0.61064,
		0.652397,
		0.709869,
		0.767341,
		0.824813,
		0.887673,
		0.950084,
		1.012944,
		1.075355,
		1.075355,
		1.075355,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.593,
		0.641033,
		0.689659,
		0.758447,
		0.80648,
		0.861629,
		0.937533,
		1.013437,
		1.089341,
		1.172361,
		1.254788,
		1.337808,
		1.420235,
		1.420235,
		1.420235,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		0.751,
		0.811831,
		0.873413,
		0.960529,
		1.02136,
		1.091203,
		1.187331,
		1.283459,
		1.379587,
		1.484727,
		1.589116,
		1.694256,
		1.798645,
		1.798645,
		1.798645,
	}
	// attack: charge = [4]
	charge = []float64{
		0.913,
		0.986953,
		1.061819,
		1.167727,
		1.24168,
		1.326589,
		1.443453,
		1.560317,
		1.677181,
		1.805001,
		1.931908,
		2.059728,
		2.186635,
		2.186635,
		2.186635,
	}
	// attack: collision = [5]
	collision = []float64{
		0.639324,
		0.691362,
		0.7434,
		0.81774,
		0.869778,
		0.92925,
		1.011024,
		1.092798,
		1.174572,
		1.26378,
		1.352988,
		1.442196,
		1.531404,
		1.620612,
		1.70982,
	}
	// attack: highPlunge = [7]
	highPlunge = []float64{
		1.596762,
		1.726731,
		1.8567,
		2.04237,
		2.172339,
		2.320875,
		2.525112,
		2.729349,
		2.933586,
		3.15639,
		3.379194,
		3.601998,
		3.824802,
		4.047606,
		4.27041,
	}
	// attack: lowPlunge = [6]
	lowPlunge = []float64{
		1.278377,
		1.382431,
		1.486485,
		1.635134,
		1.739187,
		1.858106,
		2.02162,
		2.185133,
		2.348646,
		2.527025,
		2.705403,
		2.883781,
		3.062159,
		3.240537,
		3.418915,
	}
	// skill: skill = [0]
	skill = []float64{
		0.471,
		0.506325,
		0.54165,
		0.58875,
		0.624075,
		0.6594,
		0.7065,
		0.7536,
		0.8007,
		0.8478,
		0.8949,
		0.942,
		1.000875,
		1.000875,
		1.000875,
	}
	// skill: skillTickL1 = [1]
	skillTickL1 = []float64{
		0.396,
		0.4257,
		0.4554,
		0.495,
		0.5247,
		0.5544,
		0.594,
		0.6336,
		0.6732,
		0.7128,
		0.7524,
		0.792,
		0.8415,
		0.8415,
		0.8415,
	}
	// skill: skillTickL2 = [2]
	skillTickL2 = []float64{
		0.84,
		0.903,
		0.966,
		1.05,
		1.113,
		1.176,
		1.26,
		1.344,
		1.428,
		1.512,
		1.596,
		1.68,
		1.785,
		1.785,
		1.785,
	}
	// skill: arkhe = [3]
	arkhe = []float64{
		0.385,
		0.413875,
		0.44275,
		0.48125,
		0.510125,
		0.539,
		0.5775,
		0.616,
		0.6545,
		0.693,
		0.7315,
		0.77,
		0.818125,
		0.818125,
		0.818125,
	}
	// burst: burst = [0]
	burst = []float64{
		2.172,
		2.3349,
		2.4978,
		2.715,
		2.8779,
		3.0408,
		3.258,
		3.4752,
		3.6924,
		3.9096,
		4.1268,
		4.344,
		4.6155,
		4.6155,
		4.6155,
	}
)
