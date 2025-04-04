// Code generated by "pipeline"; DO NOT EDIT.
package nilou

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
	1: {"travel"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Nilou, ValidateParamKeys)
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
	auto = [][]float64{
		auto_1,
		auto_2,
		auto_3,
	}
)

var (
	// attack: auto_1 = [0]
	auto_1 = []float64{
		0.503074,
		0.544022,
		0.58497,
		0.643467,
		0.684415,
		0.731213,
		0.795559,
		0.859906,
		0.924253,
		0.994449,
		1.064645,
		1.134842,
		1.205038,
		1.275235,
		1.345431,
	}
	// attack: auto_2 = [1]
	auto_2 = []float64{
		0.45439,
		0.491375,
		0.52836,
		0.581196,
		0.618181,
		0.66045,
		0.71857,
		0.776689,
		0.834809,
		0.898212,
		0.961615,
		1.025018,
		1.088422,
		1.151825,
		1.215228,
	}
	// attack: auto_3 = [2]
	auto_3 = []float64{
		0.70354,
		0.760805,
		0.81807,
		0.899877,
		0.957142,
		1.022588,
		1.112575,
		1.202563,
		1.292551,
		1.390719,
		1.488887,
		1.587056,
		1.685224,
		1.783393,
		1.881561,
	}
	// attack: charge = [3 4]
	charge = [][]float64{
		{
			0.50224,
			0.54312,
			0.584,
			0.6424,
			0.68328,
			0.73,
			0.79424,
			0.85848,
			0.92272,
			0.9928,
			1.06288,
			1.13296,
			1.20304,
			1.27312,
			1.3432,
		},
		{
			0.54438,
			0.58869,
			0.633,
			0.6963,
			0.74061,
			0.79125,
			0.86088,
			0.93051,
			1.00014,
			1.0761,
			1.15206,
			1.22802,
			1.30398,
			1.37994,
			1.4559,
		},
	}
	// skill: skill = [0]
	skill = []float64{
		0.033389,
		0.035893,
		0.038397,
		0.041736,
		0.04424,
		0.046744,
		0.050083,
		0.053422,
		0.056761,
		0.0601,
		0.063439,
		0.066778,
		0.070951,
		0.075125,
		0.079298,
	}
	// skill: swordDance = [5 6 3]
	swordDance = [][]float64{
		{
			0.045525,
			0.048939,
			0.052354,
			0.056906,
			0.06032,
			0.063735,
			0.068287,
			0.07284,
			0.077392,
			0.081945,
			0.086497,
			0.09105,
			0.09674,
			0.102431,
			0.108121,
		},
		{
			0.051445,
			0.055303,
			0.059162,
			0.064306,
			0.068164,
			0.072023,
			0.077167,
			0.082312,
			0.087456,
			0.092601,
			0.097745,
			0.10289,
			0.10932,
			0.115751,
			0.122181,
		},
		{
			0.071688,
			0.077065,
			0.082441,
			0.08961,
			0.094987,
			0.100363,
			0.107532,
			0.114701,
			0.12187,
			0.129038,
			0.136207,
			0.143376,
			0.152337,
			0.161298,
			0.170259,
		},
	}
	// skill: whirlingSteps = [1 2 4]
	whirlingSteps = [][]float64{
		{
			0.032619,
			0.035066,
			0.037512,
			0.040774,
			0.04322,
			0.045667,
			0.048929,
			0.052191,
			0.055453,
			0.058715,
			0.061976,
			0.065238,
			0.069316,
			0.073393,
			0.077471,
		},
		{
			0.039605,
			0.042575,
			0.045546,
			0.049506,
			0.052476,
			0.055447,
			0.059407,
			0.063368,
			0.067328,
			0.071289,
			0.075249,
			0.07921,
			0.08416,
			0.089111,
			0.094061,
		},
		{
			0.050616,
			0.054412,
			0.058208,
			0.06327,
			0.067066,
			0.070862,
			0.075924,
			0.080986,
			0.086047,
			0.091109,
			0.09617,
			0.101232,
			0.107559,
			0.113886,
			0.120213,
		},
	}
	// burst: burst = [0]
	burst = []float64{
		0.18432,
		0.198144,
		0.211968,
		0.2304,
		0.244224,
		0.258048,
		0.27648,
		0.294912,
		0.313344,
		0.331776,
		0.350208,
		0.36864,
		0.39168,
		0.41472,
		0.43776,
	}
	// burst: burstAeon = [1]
	burstAeon = []float64{
		0.22528,
		0.242176,
		0.259072,
		0.2816,
		0.298496,
		0.315392,
		0.33792,
		0.360448,
		0.382976,
		0.405504,
		0.428032,
		0.45056,
		0.47872,
		0.50688,
		0.53504,
	}
)
