// Code generated by "pipeline"; DO NOT EDIT.
package kirara

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
	1: {"short_hold", "hold"},
	2: {"hits", "mine_delay"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Kirara, ValidateParamKeys)
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
		0.47902,
		0.51801,
		0.557,
		0.6127,
		0.65169,
		0.69625,
		0.75752,
		0.81879,
		0.88006,
		0.9469,
		1.01374,
		1.08058,
		1.14742,
		1.21426,
		1.2811,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.46354,
		0.50127,
		0.539,
		0.5929,
		0.63063,
		0.67375,
		0.73304,
		0.79233,
		0.85162,
		0.9163,
		0.98098,
		1.04566,
		1.11034,
		1.17502,
		1.2397,
	}
	// attack: attack_3 = [2 3]
	attack_3 = [][]float64{
		{
			0.254216,
			0.274908,
			0.2956,
			0.32516,
			0.345852,
			0.3695,
			0.402016,
			0.434532,
			0.467048,
			0.50252,
			0.537992,
			0.573464,
			0.608936,
			0.644408,
			0.67988,
		},
		{
			0.381324,
			0.412362,
			0.4434,
			0.48774,
			0.518778,
			0.55425,
			0.603024,
			0.651798,
			0.700572,
			0.75378,
			0.806988,
			0.860196,
			0.913404,
			0.966612,
			1.01982,
		},
	}
	// attack: attack_4 = [4]
	attack_4 = []float64{
		0.73272,
		0.79236,
		0.852,
		0.9372,
		0.99684,
		1.065,
		1.15872,
		1.25244,
		1.34616,
		1.4484,
		1.55064,
		1.65288,
		1.75512,
		1.85736,
		1.9596,
	}
	// attack: charge = [5 6 6]
	charge = [][]float64{
		{
			0.223772,
			0.241986,
			0.2602,
			0.28622,
			0.304434,
			0.32525,
			0.353872,
			0.382494,
			0.411116,
			0.44234,
			0.473564,
			0.504788,
			0.536012,
			0.567236,
			0.59846,
		},
		{
			0.447544,
			0.483972,
			0.5204,
			0.57244,
			0.608868,
			0.6505,
			0.707744,
			0.764988,
			0.822232,
			0.88468,
			0.947128,
			1.009576,
			1.072024,
			1.134472,
			1.19692,
		},
		{
			0.447544,
			0.483972,
			0.5204,
			0.57244,
			0.608868,
			0.6505,
			0.707744,
			0.764988,
			0.822232,
			0.88468,
			0.947128,
			1.009576,
			1.072024,
			1.134472,
			1.19692,
		},
	}
	// skill: catDmg = [6]
	catDmg = []float64{
		0.336,
		0.3612,
		0.3864,
		0.42,
		0.4452,
		0.4704,
		0.504,
		0.5376,
		0.5712,
		0.6048,
		0.6384,
		0.672,
		0.714,
		0.756,
		0.798,
	}
	// skill: flipclawDmg = [8]
	flipclawDmg = []float64{
		1.44,
		1.548,
		1.656,
		1.8,
		1.908,
		2.016,
		2.16,
		2.304,
		2.448,
		2.592,
		2.736,
		2.88,
		3.06,
		3.24,
		3.42,
	}
	// skill: maxShieldFlat = [4]
	maxShieldFlat = []float64{
		1541.0796,
		1695.2089,
		1862.1824,
		2042,
		2234.6616,
		2440.1675,
		2658.5176,
		2889.7117,
		3133.7498,
		3390.632,
		3660.3584,
		3942.929,
		4238.3438,
		4546.6025,
		4867.705,
	}
	// skill: maxShieldPP = [3]
	maxShieldPP = []float64{
		0.16,
		0.172,
		0.184,
		0.2,
		0.212,
		0.224,
		0.24,
		0.256,
		0.272,
		0.288,
		0.304,
		0.32,
		0.34,
		0.36,
		0.38,
	}
	// skill: shieldFlat = [2]
	shieldFlat = []float64{
		962.2313,
		1058.4679,
		1162.7241,
		1275,
		1395.2957,
		1523.611,
		1659.946,
		1804.3008,
		1956.6753,
		2117.0693,
		2285.4834,
		2461.917,
		2646.3704,
		2838.8433,
		3039.336,
	}
	// skill: shieldPP = [1]
	shieldPP = []float64{
		0.1,
		0.1075,
		0.115,
		0.125,
		0.1325,
		0.14,
		0.15,
		0.16,
		0.17,
		0.18,
		0.19,
		0.2,
		0.2125,
		0.225,
		0.2375,
	}
	// skill: skillPress = [0]
	skillPress = []float64{
		1.04,
		1.118,
		1.196,
		1.3,
		1.378,
		1.456,
		1.56,
		1.664,
		1.768,
		1.872,
		1.976,
		2.08,
		2.21,
		2.34,
		2.47,
	}
	// burst: burst = [0]
	burst = []float64{
		5.7024,
		6.13008,
		6.55776,
		7.128,
		7.55568,
		7.98336,
		8.5536,
		9.12384,
		9.69408,
		10.26432,
		10.83456,
		11.4048,
		12.1176,
		12.8304,
		13.5432,
	}
	// burst: cardamom = [1]
	cardamom = []float64{
		0.3564,
		0.38313,
		0.40986,
		0.4455,
		0.47223,
		0.49896,
		0.5346,
		0.57024,
		0.60588,
		0.64152,
		0.67716,
		0.7128,
		0.75735,
		0.8019,
		0.84645,
	}
)
