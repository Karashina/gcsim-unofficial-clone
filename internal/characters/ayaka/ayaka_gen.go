// Code generated by "pipeline"; DO NOT EDIT.
package ayaka

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
	5: {"collision"},
	6: {"collision"},
	8: {"f"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Ayaka, ValidateParamKeys)
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
		0.457253,
		0.494472,
		0.53169,
		0.584859,
		0.622077,
		0.664613,
		0.723098,
		0.781584,
		0.84007,
		0.903873,
		0.967676,
		1.031479,
		1.095281,
		1.159084,
		1.222887,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.486846,
		0.526473,
		0.5661,
		0.62271,
		0.662337,
		0.707625,
		0.769896,
		0.832167,
		0.894438,
		0.96237,
		1.030302,
		1.098234,
		1.166166,
		1.234098,
		1.30203,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.626218,
		0.677189,
		0.72816,
		0.800976,
		0.851947,
		0.9102,
		0.990298,
		1.070395,
		1.150493,
		1.237872,
		1.325251,
		1.41263,
		1.50001,
		1.587389,
		1.674768,
	}
	// attack: attack_4 = [3 3 3]
	attack_4 = [][]float64{
		{
			0.226464,
			0.244897,
			0.26333,
			0.289663,
			0.308096,
			0.329163,
			0.358129,
			0.387095,
			0.416061,
			0.447661,
			0.479261,
			0.51086,
			0.54246,
			0.574059,
			0.605659,
		},
		{
			0.226464,
			0.244897,
			0.26333,
			0.289663,
			0.308096,
			0.329163,
			0.358129,
			0.387095,
			0.416061,
			0.447661,
			0.479261,
			0.51086,
			0.54246,
			0.574059,
			0.605659,
		},
		{
			0.226464,
			0.244897,
			0.26333,
			0.289663,
			0.308096,
			0.329163,
			0.358129,
			0.387095,
			0.416061,
			0.447661,
			0.479261,
			0.51086,
			0.54246,
			0.574059,
			0.605659,
		},
	}
	// attack: attack_5 = [6]
	attack_5 = []float64{
		0.781817,
		0.845454,
		0.90909,
		0.999999,
		1.063635,
		1.136363,
		1.236362,
		1.336362,
		1.436362,
		1.545453,
		1.654544,
		1.763635,
		1.872725,
		1.981816,
		2.090907,
	}
	// attack: ca = [7]
	ca = []float64{
		0.55126,
		0.59613,
		0.641,
		0.7051,
		0.74997,
		0.80125,
		0.87176,
		0.94227,
		1.01278,
		1.0897,
		1.16662,
		1.24354,
		1.32046,
		1.39738,
		1.4743,
	}
	// attack: collision = [9]
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
	// attack: highPlunge = [11]
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
	// attack: lowPlunge = [10]
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
		2.392,
		2.5714,
		2.7508,
		2.99,
		3.1694,
		3.3488,
		3.588,
		3.8272,
		4.0664,
		4.3056,
		4.5448,
		4.784,
		5.083,
		5.382,
		5.681,
	}
	// burst: burstBloom = [1]
	burstBloom = []float64{
		1.6845,
		1.810837,
		1.937175,
		2.105625,
		2.231962,
		2.3583,
		2.52675,
		2.6952,
		2.86365,
		3.0321,
		3.20055,
		3.369,
		3.579562,
		3.790125,
		4.000687,
	}
	// burst: burstCut = [0]
	burstCut = []float64{
		1.123,
		1.207225,
		1.29145,
		1.40375,
		1.487975,
		1.5722,
		1.6845,
		1.7968,
		1.9091,
		2.0214,
		2.1337,
		2.246,
		2.386375,
		2.52675,
		2.667125,
	}
)
