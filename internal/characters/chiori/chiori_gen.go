// Code generated by "pipeline"; DO NOT EDIT.
package chiori

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
	validation.RegisterCharParamValidationFunc(keys.Chiori, ValidateParamKeys)
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
		attack_3,
		{attack_4},
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.494104,
		0.534322,
		0.57454,
		0.631994,
		0.672212,
		0.718175,
		0.781374,
		0.844574,
		0.907773,
		0.976718,
		1.045663,
		1.114608,
		1.183552,
		1.252497,
		1.321442,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.468339,
		0.506459,
		0.54458,
		0.599038,
		0.637159,
		0.680725,
		0.740629,
		0.800533,
		0.860436,
		0.925786,
		0.991136,
		1.056485,
		1.121835,
		1.187184,
		1.252534,
	}
	// attack: attack_3 = [2 3]
	attack_3 = [][]float64{
		{
			0.304165,
			0.328922,
			0.35368,
			0.389048,
			0.413806,
			0.4421,
			0.481005,
			0.51991,
			0.558814,
			0.601256,
			0.643698,
			0.686139,
			0.728581,
			0.771022,
			0.813464,
		},
		{
			0.304165,
			0.328922,
			0.35368,
			0.389048,
			0.413806,
			0.4421,
			0.481005,
			0.51991,
			0.558814,
			0.601256,
			0.643698,
			0.686139,
			0.728581,
			0.771022,
			0.813464,
		},
	}
	// attack: attack_4 = [4]
	attack_4 = []float64{
		0.751227,
		0.812374,
		0.87352,
		0.960872,
		1.022018,
		1.0919,
		1.187987,
		1.284074,
		1.380162,
		1.484984,
		1.589806,
		1.694629,
		1.799451,
		1.904274,
		2.009096,
	}
	// attack: charge = [5 6]
	charge = [][]float64{
		{
			0.54309,
			0.587295,
			0.6315,
			0.69465,
			0.738855,
			0.789375,
			0.85884,
			0.928305,
			0.99777,
			1.07355,
			1.14933,
			1.22511,
			1.30089,
			1.37667,
			1.45245,
		},
		{
			0.54309,
			0.587295,
			0.6315,
			0.69465,
			0.738855,
			0.789375,
			0.85884,
			0.928305,
			0.99777,
			1.07355,
			1.14933,
			1.22511,
			1.30089,
			1.37667,
			1.45245,
		},
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
	// attack: plunge = [8]
	plunge = []float64{
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
	// skill: thrustAtkScaling = [4]
	thrustAtkScaling = []float64{
		1.4928,
		1.60476,
		1.71672,
		1.866,
		1.97796,
		2.08992,
		2.2392,
		2.38848,
		2.53776,
		2.68704,
		2.83632,
		2.9856,
		3.1722,
		3.3588,
		3.5454,
	}
	// skill: thrustDefScaling = [5]
	thrustDefScaling = []float64{
		1.866,
		2.00595,
		2.1459,
		2.3325,
		2.47245,
		2.6124,
		2.799,
		2.9856,
		3.1722,
		3.3588,
		3.5454,
		3.732,
		3.96525,
		4.1985,
		4.43175,
	}
	// skill: turretAtkScaling = [0]
	turretAtkScaling = []float64{
		0.8208,
		0.88236,
		0.94392,
		1.026,
		1.08756,
		1.14912,
		1.2312,
		1.31328,
		1.39536,
		1.47744,
		1.55952,
		1.6416,
		1.7442,
		1.8468,
		1.9494,
	}
	// skill: turretDefScaling = [1]
	turretDefScaling = []float64{
		1.026,
		1.10295,
		1.1799,
		1.2825,
		1.35945,
		1.4364,
		1.539,
		1.6416,
		1.7442,
		1.8468,
		1.9494,
		2.052,
		2.18025,
		2.3085,
		2.43675,
	}
	// skill: turretDuration = [2]
	turretDuration = []float64{
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
		17,
	}
	// skill: turretInterval = [3]
	turretInterval = []float64{
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
		3.6,
	}
	// burst: burstAtkScaling = [0]
	burstAtkScaling = []float64{
		2.5632,
		2.75544,
		2.94768,
		3.204,
		3.39624,
		3.58848,
		3.8448,
		4.10112,
		4.35744,
		4.61376,
		4.87008,
		5.1264,
		5.4468,
		5.7672,
		6.0876,
	}
	// burst: burstCD = [2]
	burstCD = []float64{
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
	}
	// burst: burstDefScaling = [1]
	burstDefScaling = []float64{
		3.204,
		3.4443,
		3.6846,
		4.005,
		4.2453,
		4.4856,
		4.806,
		5.1264,
		5.4468,
		5.7672,
		6.0876,
		6.408,
		6.8085,
		7.209,
		7.6095,
	}
)
