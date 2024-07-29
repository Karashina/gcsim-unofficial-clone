// Code generated by "pipeline"; DO NOT EDIT.
package yelan

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
	1: {"marked"},
	3: {"travel"},
	7: {"hold", "travel", "weakspot"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Yelan, ValidateParamKeys)
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
		attack_4,
	}
)

var (
	// attack: aim = [4]
	aim = []float64{
		0.4386,
		0.4743,
		0.51,
		0.561,
		0.5967,
		0.6375,
		0.6936,
		0.7497,
		0.8058,
		0.867,
		0.9282,
		0.9894,
		1.0506,
		1.1118,
		1.173,
	}
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.40678,
		0.43989,
		0.473,
		0.5203,
		0.55341,
		0.59125,
		0.64328,
		0.69531,
		0.74734,
		0.8041,
		0.86086,
		0.91762,
		0.97438,
		1.03114,
		1.0879,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.39044,
		0.42222,
		0.454,
		0.4994,
		0.53118,
		0.5675,
		0.61744,
		0.66738,
		0.71732,
		0.7718,
		0.82628,
		0.88076,
		0.93524,
		0.98972,
		1.0442,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.516,
		0.558,
		0.6,
		0.66,
		0.702,
		0.75,
		0.816,
		0.882,
		0.948,
		1.02,
		1.092,
		1.164,
		1.236,
		1.308,
		1.38,
	}
	// attack: attack_4 = [3 3]
	attack_4 = [][]float64{
		{
			0.32508,
			0.35154,
			0.378,
			0.4158,
			0.44226,
			0.4725,
			0.51408,
			0.55566,
			0.59724,
			0.6426,
			0.68796,
			0.73332,
			0.77868,
			0.82404,
			0.8694,
		},
		{
			0.32508,
			0.35154,
			0.378,
			0.4158,
			0.44226,
			0.4725,
			0.51408,
			0.55566,
			0.59724,
			0.6426,
			0.68796,
			0.73332,
			0.77868,
			0.82404,
			0.8694,
		},
	}
	// attack: barb = [6]
	barb = []float64{
		0.11576,
		0.124442,
		0.133124,
		0.1447,
		0.153382,
		0.162064,
		0.17364,
		0.185216,
		0.196792,
		0.208368,
		0.219944,
		0.23152,
		0.24599,
		0.26046,
		0.27493,
	}
	// attack: fullaim = [5]
	fullaim = []float64{
		1.24,
		1.333,
		1.426,
		1.55,
		1.643,
		1.736,
		1.86,
		1.984,
		2.108,
		2.232,
		2.356,
		2.48,
		2.635,
		2.79,
		2.945,
	}
	// skill: skill = [0]
	skill = []float64{
		0.226136,
		0.243096,
		0.260056,
		0.28267,
		0.29963,
		0.31659,
		0.339204,
		0.361818,
		0.384431,
		0.407045,
		0.429658,
		0.452272,
		0.480539,
		0.508806,
		0.537073,
	}
	// burst: burst = [0]
	burst = []float64{
		0.07308,
		0.078561,
		0.084042,
		0.09135,
		0.096831,
		0.102312,
		0.10962,
		0.116928,
		0.124236,
		0.131544,
		0.138852,
		0.14616,
		0.155295,
		0.16443,
		0.173565,
	}
	// burst: burstDice = [1]
	burstDice = []float64{
		0.04872,
		0.052374,
		0.056028,
		0.0609,
		0.064554,
		0.068208,
		0.07308,
		0.077952,
		0.082824,
		0.087696,
		0.092568,
		0.09744,
		0.10353,
		0.10962,
		0.11571,
	}
)
