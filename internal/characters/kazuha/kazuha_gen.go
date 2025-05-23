// Code generated by "pipeline"; DO NOT EDIT.
package kazuha

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
	1: {"hold", "glide_cancel"},
	5: {"collision"},
	6: {"collision"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Kazuha, ValidateParamKeys)
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
		attack_5,
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.44978,
		0.48639,
		0.523,
		0.5753,
		0.61191,
		0.65375,
		0.71128,
		0.76881,
		0.82634,
		0.8891,
		0.961013,
		1.045582,
		1.130151,
		1.21472,
		1.306977,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.45236,
		0.48918,
		0.526,
		0.5786,
		0.61542,
		0.6575,
		0.71536,
		0.77322,
		0.83108,
		0.8942,
		0.966525,
		1.051579,
		1.136633,
		1.221688,
		1.314474,
	}
	// attack: attack_3 = [2 3]
	attack_3 = [][]float64{
		{
			0.258,
			0.279,
			0.3,
			0.33,
			0.351,
			0.375,
			0.408,
			0.441,
			0.474,
			0.51,
			0.55125,
			0.59976,
			0.64827,
			0.69678,
			0.7497,
		},
		{
			0.3096,
			0.3348,
			0.36,
			0.396,
			0.4212,
			0.45,
			0.4896,
			0.5292,
			0.5688,
			0.612,
			0.6615,
			0.719712,
			0.777924,
			0.836136,
			0.89964,
		},
	}
	// attack: attack_4 = [4]
	attack_4 = []float64{
		0.60716,
		0.65658,
		0.706,
		0.7766,
		0.82602,
		0.8825,
		0.96016,
		1.03782,
		1.11548,
		1.2002,
		1.297275,
		1.411435,
		1.525595,
		1.639756,
		1.764294,
	}
	// attack: attack_5 = [5 5 5]
	attack_5 = [][]float64{
		{
			0.2537,
			0.27435,
			0.295,
			0.3245,
			0.34515,
			0.36875,
			0.4012,
			0.43365,
			0.4661,
			0.5015,
			0.542063,
			0.589764,
			0.637465,
			0.685167,
			0.737205,
		},
		{
			0.2537,
			0.27435,
			0.295,
			0.3245,
			0.34515,
			0.36875,
			0.4012,
			0.43365,
			0.4661,
			0.5015,
			0.542063,
			0.589764,
			0.637465,
			0.685167,
			0.737205,
		},
		{
			0.2537,
			0.27435,
			0.295,
			0.3245,
			0.34515,
			0.36875,
			0.4012,
			0.43365,
			0.4661,
			0.5015,
			0.542063,
			0.589764,
			0.637465,
			0.685167,
			0.737205,
		},
	}
	// attack: charge = [6 7]
	charge = [][]float64{
		{
			0.43,
			0.465,
			0.5,
			0.55,
			0.585,
			0.625,
			0.68,
			0.735,
			0.79,
			0.85,
			0.91875,
			0.9996,
			1.08045,
			1.1613,
			1.2495,
		},
		{
			0.74648,
			0.80724,
			0.868,
			0.9548,
			1.01556,
			1.085,
			1.18048,
			1.27596,
			1.37144,
			1.4756,
			1.59495,
			1.735306,
			1.875661,
			2.016017,
			2.169132,
		},
	}
	// attack: collision = [9]
	collision = []float64{
		0.818335,
		0.884943,
		0.951552,
		1.046707,
		1.113316,
		1.18944,
		1.294111,
		1.398781,
		1.503452,
		1.617638,
		1.731825,
		1.846011,
		1.960197,
		2.074383,
		2.18857,
	}
	// attack: highPlunge = [11]
	highPlunge = []float64{
		2.043855,
		2.210216,
		2.376576,
		2.614234,
		2.780594,
		2.97072,
		3.232143,
		3.493567,
		3.75499,
		4.040179,
		4.325368,
		4.610557,
		4.895747,
		5.180936,
		5.466125,
	}
	// attack: lowPlunge = [10]
	lowPlunge = []float64{
		1.636323,
		1.769512,
		1.902701,
		2.092971,
		2.22616,
		2.378376,
		2.587673,
		2.79697,
		3.006267,
		3.234591,
		3.462915,
		3.69124,
		3.919564,
		4.147888,
		4.376212,
	}
	// skill: skill = [0]
	skill = []float64{
		1.92,
		2.064,
		2.208,
		2.4,
		2.544,
		2.688,
		2.88,
		3.072,
		3.264,
		3.456,
		3.648,
		3.84,
		4.08,
		4.32,
		4.56,
	}
	// skill: skillHold = [2]
	skillHold = []float64{
		2.608,
		2.8036,
		2.9992,
		3.26,
		3.4556,
		3.6512,
		3.912,
		4.1728,
		4.4336,
		4.6944,
		4.9552,
		5.216,
		5.542,
		5.868,
		6.194,
	}
	// burst: burstDot = [1]
	burstDot = []float64{
		1.2,
		1.29,
		1.38,
		1.5,
		1.59,
		1.68,
		1.8,
		1.92,
		2.04,
		2.16,
		2.28,
		2.4,
		2.55,
		2.7,
		2.85,
	}
	// burst: burstEleDot = [2]
	burstEleDot = []float64{
		0.36,
		0.387,
		0.414,
		0.45,
		0.477,
		0.504,
		0.54,
		0.576,
		0.612,
		0.648,
		0.684,
		0.72,
		0.765,
		0.81,
		0.855,
	}
	// burst: burstSlash = [0]
	burstSlash = []float64{
		2.624,
		2.8208,
		3.0176,
		3.28,
		3.4768,
		3.6736,
		3.936,
		4.1984,
		4.4608,
		4.7232,
		4.9856,
		5.248,
		5.576,
		5.904,
		6.232,
	}
)
