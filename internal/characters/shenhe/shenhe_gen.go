// Code generated by "pipeline"; DO NOT EDIT.
package shenhe

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
	validation.RegisterCharParamValidationFunc(keys.Shenhe, ValidateParamKeys)
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
	attack = [][][]float64{
		{attack_1},
		{attack_2},
		{attack_3},
		attack_4,
		{attack_5},
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.43258,
		0.46779,
		0.503,
		0.5533,
		0.58851,
		0.62875,
		0.68408,
		0.73941,
		0.79474,
		0.8551,
		0.91546,
		0.97582,
		1.03618,
		1.09654,
		1.1569,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.40248,
		0.43524,
		0.468,
		0.5148,
		0.54756,
		0.585,
		0.63648,
		0.68796,
		0.73944,
		0.7956,
		0.85176,
		0.90792,
		0.96408,
		1.02024,
		1.0764,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.5332,
		0.5766,
		0.62,
		0.682,
		0.7254,
		0.775,
		0.8432,
		0.9114,
		0.9796,
		1.054,
		1.1284,
		1.2028,
		1.2772,
		1.3516,
		1.426,
	}
	// attack: attack_4 = [3 3]
	attack_4 = [][]float64{
		{
			0.26316,
			0.28458,
			0.306,
			0.3366,
			0.35802,
			0.3825,
			0.41616,
			0.44982,
			0.48348,
			0.5202,
			0.55692,
			0.59364,
			0.63036,
			0.66708,
			0.7038,
		},
		{
			0.26316,
			0.28458,
			0.306,
			0.3366,
			0.35802,
			0.3825,
			0.41616,
			0.44982,
			0.48348,
			0.5202,
			0.55692,
			0.59364,
			0.63036,
			0.66708,
			0.7038,
		},
	}
	// attack: attack_5 = [5]
	attack_5 = []float64{
		0.65618,
		0.70959,
		0.763,
		0.8393,
		0.89271,
		0.95375,
		1.03768,
		1.12161,
		1.20554,
		1.2971,
		1.38866,
		1.48022,
		1.57178,
		1.66334,
		1.7549,
	}
	// attack: charged = [6]
	charged = []float64{
		1.106734,
		1.196817,
		1.2869,
		1.41559,
		1.505673,
		1.608625,
		1.750184,
		1.891743,
		2.033302,
		2.18773,
		2.342158,
		2.496586,
		2.651014,
		2.805442,
		2.95987,
	}
	// skill: skillHold = [1]
	skillHold = []float64{
		1.888,
		2.0296,
		2.1712,
		2.36,
		2.5016,
		2.6432,
		2.832,
		3.0208,
		3.2096,
		3.3984,
		3.5872,
		3.776,
		4.012,
		4.248,
		4.484,
	}
	// skill: skillPress = [0]
	skillPress = []float64{
		1.392,
		1.4964,
		1.6008,
		1.74,
		1.8444,
		1.9488,
		2.088,
		2.2272,
		2.3664,
		2.5056,
		2.6448,
		2.784,
		2.958,
		3.132,
		3.306,
	}
	// skill: skillpp = [2]
	skillpp = []float64{
		0.45656,
		0.490802,
		0.525044,
		0.5707,
		0.604942,
		0.639184,
		0.68484,
		0.730496,
		0.776152,
		0.821808,
		0.867464,
		0.91312,
		0.97019,
		1.02726,
		1.08433,
	}
	// burst: burst = [0]
	burst = []float64{
		1.008,
		1.0836,
		1.1592,
		1.26,
		1.3356,
		1.4112,
		1.512,
		1.6128,
		1.7136,
		1.8144,
		1.9152,
		2.016,
		2.142,
		2.268,
		2.394,
	}
	// burst: burstdot = [2]
	burstdot = []float64{
		0.3312,
		0.35604,
		0.38088,
		0.414,
		0.43884,
		0.46368,
		0.4968,
		0.52992,
		0.56304,
		0.59616,
		0.62928,
		0.6624,
		0.7038,
		0.7452,
		0.7866,
	}
	// burst: burstrespp = [1]
	burstrespp = []float64{
		0.06,
		0.07,
		0.08,
		0.09,
		0.1,
		0.11,
		0.12,
		0.13,
		0.14,
		0.15,
		0.15,
		0.15,
		0.15,
		0.15,
		0.15,
	}
)
