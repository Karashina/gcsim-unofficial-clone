// Code generated by "pipeline"; DO NOT EDIT.
package arlecchino

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
	4: {"early_cancel"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Arlecchino, ValidateParamKeys)
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
		{attack_6},
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.475004,
		0.513667,
		0.55233,
		0.607563,
		0.646226,
		0.690412,
		0.751169,
		0.811925,
		0.872681,
		0.938961,
		1.005241,
		1.07152,
		1.1378,
		1.204079,
		1.270359,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.521057,
		0.563468,
		0.60588,
		0.666468,
		0.70888,
		0.75735,
		0.823997,
		0.890644,
		0.95729,
		1.029996,
		1.102702,
		1.175407,
		1.248113,
		1.320818,
		1.393524,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.653858,
		0.707079,
		0.7603,
		0.83633,
		0.889551,
		0.950375,
		1.034008,
		1.117641,
		1.201274,
		1.29251,
		1.383746,
		1.474982,
		1.566218,
		1.657454,
		1.74869,
	}
	// attack: attack_4 = [3 3]
	attack_4 = [][]float64{
		{
			0.371451,
			0.401686,
			0.43192,
			0.475112,
			0.505346,
			0.5399,
			0.587411,
			0.634922,
			0.682434,
			0.734264,
			0.786094,
			0.837925,
			0.889755,
			0.941586,
			0.993416,
		},
		{
			0.371451,
			0.401686,
			0.43192,
			0.475112,
			0.505346,
			0.5399,
			0.587411,
			0.634922,
			0.682434,
			0.734264,
			0.786094,
			0.837925,
			0.889755,
			0.941586,
			0.993416,
		},
	}
	// attack: attack_5 = [4]
	attack_5 = []float64{
		0.699816,
		0.756778,
		0.81374,
		0.895114,
		0.952076,
		1.017175,
		1.106686,
		1.196198,
		1.285709,
		1.383358,
		1.481007,
		1.578656,
		1.676304,
		1.773953,
		1.871602,
	}
	// attack: attack_6 = [5]
	attack_6 = []float64{
		0.853782,
		0.923276,
		0.99277,
		1.092047,
		1.161541,
		1.240962,
		1.350167,
		1.459372,
		1.568577,
		1.687709,
		1.806841,
		1.925974,
		2.045106,
		2.164239,
		2.283371,
	}
	// attack: charge = [6]
	charge = []float64{
		0.90816,
		0.98208,
		1.056,
		1.1616,
		1.23552,
		1.32,
		1.43616,
		1.55232,
		1.66848,
		1.7952,
		1.92192,
		2.04864,
		2.17536,
		2.30208,
		2.4288,
	}
	// attack: collision = [8]
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
	// attack: highPlunge = [10]
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
	// attack: lowPlunge = [9]
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
	// attack: masque = [11]
	masque = []float64{
		1.204,
		1.302,
		1.4,
		1.54,
		1.638,
		1.75,
		1.904,
		2.058,
		2.212,
		2.38,
		2.548,
		2.716,
		2.884,
		3.052,
		3.22,
	}
	// skill: skillFinal = [1]
	skillFinal = []float64{
		1.3356,
		1.43577,
		1.53594,
		1.6695,
		1.76967,
		1.86984,
		2.0034,
		2.13696,
		2.27052,
		2.40408,
		2.53764,
		2.6712,
		2.83815,
		3.0051,
		3.17205,
	}
	// skill: skillSigil = [2]
	skillSigil = []float64{
		0.318,
		0.34185,
		0.3657,
		0.3975,
		0.42135,
		0.4452,
		0.477,
		0.5088,
		0.5406,
		0.5724,
		0.6042,
		0.636,
		0.67575,
		0.7155,
		0.75525,
	}
	// skill: skillSpike = [0]
	skillSpike = []float64{
		0.1484,
		0.15953,
		0.17066,
		0.1855,
		0.19663,
		0.20776,
		0.2226,
		0.23744,
		0.25228,
		0.26712,
		0.28196,
		0.2968,
		0.31535,
		0.3339,
		0.35245,
	}
	// burst: burst = [0]
	burst = []float64{
		3.704,
		3.9818,
		4.2596,
		4.63,
		4.9078,
		5.1856,
		5.556,
		5.9264,
		6.2968,
		6.6672,
		7.0376,
		7.408,
		7.871,
		8.334,
		8.797,
	}
)
