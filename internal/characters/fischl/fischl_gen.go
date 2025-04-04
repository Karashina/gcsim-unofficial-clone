// Code generated by "pipeline"; DO NOT EDIT.
package fischl

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
	7: {"hold", "travel", "weakspot"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Fischl, ValidateParamKeys)
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
		auto_4,
		auto_5,
	}
)

var (
	// attack: aim = [5]
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
	// attack: auto_1 = [0]
	auto_1 = []float64{
		0.44118,
		0.47709,
		0.513,
		0.5643,
		0.60021,
		0.64125,
		0.69768,
		0.75411,
		0.81054,
		0.8721,
		0.93366,
		0.99522,
		1.05678,
		1.11834,
		1.1799,
	}
	// attack: auto_2 = [1]
	auto_2 = []float64{
		0.46784,
		0.50592,
		0.544,
		0.5984,
		0.63648,
		0.68,
		0.73984,
		0.79968,
		0.85952,
		0.9248,
		0.99008,
		1.05536,
		1.12064,
		1.18592,
		1.2512,
	}
	// attack: auto_3 = [2]
	auto_3 = []float64{
		0.58136,
		0.62868,
		0.676,
		0.7436,
		0.79092,
		0.845,
		0.91936,
		0.99372,
		1.06808,
		1.1492,
		1.23032,
		1.31144,
		1.39256,
		1.47368,
		1.5548,
	}
	// attack: auto_4 = [3]
	auto_4 = []float64{
		0.57706,
		0.62403,
		0.671,
		0.7381,
		0.78507,
		0.83875,
		0.91256,
		0.98637,
		1.06018,
		1.1407,
		1.22122,
		1.30174,
		1.38226,
		1.46278,
		1.5433,
	}
	// attack: auto_5 = [4]
	auto_5 = []float64{
		0.72068,
		0.77934,
		0.838,
		0.9218,
		0.98046,
		1.0475,
		1.13968,
		1.23186,
		1.32404,
		1.4246,
		1.52516,
		1.62572,
		1.72628,
		1.82684,
		1.9274,
	}
	// attack: fullaim = [6]
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
	// skill: birdAtk = [0]
	birdAtk = []float64{
		0.888,
		0.9546,
		1.0212,
		1.11,
		1.1766,
		1.2432,
		1.332,
		1.4208,
		1.5096,
		1.5984,
		1.6872,
		1.776,
		1.887,
		1.998,
		2.109,
	}
	// skill: birdSum = [1]
	birdSum = []float64{
		1.1544,
		1.24098,
		1.32756,
		1.443,
		1.52958,
		1.61616,
		1.7316,
		1.84704,
		1.96248,
		2.07792,
		2.19336,
		2.3088,
		2.4531,
		2.5974,
		2.7417,
	}
	// burst: burst = [0]
	burst = []float64{
		2.08,
		2.236,
		2.392,
		2.6,
		2.756,
		2.912,
		3.12,
		3.328,
		3.536,
		3.744,
		3.952,
		4.16,
		4.42,
		4.68,
		4.94,
	}
)
