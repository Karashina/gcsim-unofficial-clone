// Code generated by "pipeline"; DO NOT EDIT.
package candace

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
	1: {"hold", "perfect"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Candace, ValidateParamKeys)
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
		attack_3,
		{attack_4},
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.60802,
		0.65751,
		0.707,
		0.7777,
		0.82719,
		0.88375,
		0.96152,
		1.03929,
		1.11706,
		1.2019,
		1.28674,
		1.37158,
		1.45642,
		1.54126,
		1.6261,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.61146,
		0.66123,
		0.711,
		0.7821,
		0.83187,
		0.88875,
		0.96696,
		1.04517,
		1.12338,
		1.2087,
		1.29402,
		1.37934,
		1.46466,
		1.54998,
		1.6353,
	}
	// attack: attack_3 = [2 3]
	attack_3 = [][]float64{
		{
			0.354879,
			0.383765,
			0.41265,
			0.453915,
			0.4828,
			0.515813,
			0.561204,
			0.606596,
			0.651987,
			0.701505,
			0.751023,
			0.800541,
			0.850059,
			0.899577,
			0.949095,
		},
		{
			0.433741,
			0.469046,
			0.50435,
			0.554785,
			0.59009,
			0.630438,
			0.685916,
			0.741395,
			0.796873,
			0.857395,
			0.917917,
			0.978439,
			1.038961,
			1.099483,
			1.160005,
		},
	}
	// attack: attack_4 = [4]
	attack_4 = []float64{
		0.94944,
		1.02672,
		1.104,
		1.2144,
		1.29168,
		1.38,
		1.50144,
		1.62288,
		1.74432,
		1.8768,
		2.00928,
		2.14176,
		2.27424,
		2.40672,
		2.5392,
	}
	// attack: charge = [5]
	charge = []float64{
		1.24184,
		1.34292,
		1.444,
		1.5884,
		1.68948,
		1.805,
		1.96384,
		2.12268,
		2.28152,
		2.4548,
		2.62808,
		2.80136,
		2.97464,
		3.14792,
		3.3212,
	}
	// skill: skillDmg = [0 3]
	skillDmg = [][]float64{
		{
			0.12,
			0.129,
			0.138,
			0.15,
			0.159,
			0.168,
			0.18,
			0.192,
			0.204,
			0.216,
			0.228,
			0.24,
			0.255,
			0.27,
			0.285,
		},
		{
			0.1904,
			0.20468,
			0.21896,
			0.238,
			0.25228,
			0.26656,
			0.2856,
			0.30464,
			0.32368,
			0.34272,
			0.36176,
			0.3808,
			0.4046,
			0.4284,
			0.4522,
		},
	}
	// skill: skillShieldFlat = [1]
	skillShieldFlat = []float64{
		1155.5629,
		1271.1353,
		1396.3386,
		1531.173,
		1675.6384,
		1829.7349,
		1993.4624,
		2166.8208,
		2349.8105,
		2542.4312,
		2744.6826,
		2956.5652,
		3178.0789,
		3409.2236,
		3649.9995,
	}
	// skill: skillShieldPct = [0]
	skillShieldPct = []float64{
		0.12,
		0.129,
		0.138,
		0.15,
		0.159,
		0.168,
		0.18,
		0.192,
		0.204,
		0.216,
		0.228,
		0.24,
		0.255,
		0.27,
		0.285,
	}
	// burst: burstDmg = [0]
	burstDmg = []float64{
		0.066104,
		0.071062,
		0.07602,
		0.08263,
		0.087588,
		0.092546,
		0.099156,
		0.105766,
		0.112377,
		0.118987,
		0.125598,
		0.132208,
		0.140471,
		0.148734,
		0.156997,
	}
	// burst: burstWaveDmg = [0]
	burstWaveDmg = []float64{
		0.066104,
		0.071062,
		0.07602,
		0.08263,
		0.087588,
		0.092546,
		0.099156,
		0.105766,
		0.112377,
		0.118987,
		0.125598,
		0.132208,
		0.140471,
		0.148734,
		0.156997,
	}
)
