// Code generated by "pipeline"; DO NOT EDIT.
package gorou

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
	3: {"travel"},
	7: {"hold", "travel", "weakspot"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Gorou, ValidateParamKeys)
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
	attack = [][]float64{
		attack_1,
		attack_2,
		attack_3,
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
		0.37754,
		0.40827,
		0.439,
		0.4829,
		0.51363,
		0.54875,
		0.59704,
		0.64533,
		0.69362,
		0.7463,
		0.79898,
		0.85166,
		0.90434,
		0.95702,
		1.0097,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.37152,
		0.40176,
		0.432,
		0.4752,
		0.50544,
		0.54,
		0.58752,
		0.63504,
		0.68256,
		0.7344,
		0.78624,
		0.83808,
		0.88992,
		0.94176,
		0.9936,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.4945,
		0.53475,
		0.575,
		0.6325,
		0.67275,
		0.71875,
		0.782,
		0.84525,
		0.9085,
		0.9775,
		1.0465,
		1.1155,
		1.1845,
		1.2535,
		1.3225,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		0.58996,
		0.63798,
		0.686,
		0.7546,
		0.80262,
		0.8575,
		0.93296,
		1.00842,
		1.08388,
		1.1662,
		1.24852,
		1.33084,
		1.41316,
		1.49548,
		1.5778,
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
		1.072,
		1.1524,
		1.2328,
		1.34,
		1.4204,
		1.5008,
		1.608,
		1.7152,
		1.8224,
		1.9296,
		2.0368,
		2.144,
		2.278,
		2.412,
		2.546,
	}
	// skill: skillDefBonus = [1]
	skillDefBonus = []float64{
		206.16,
		221.622,
		237.084,
		257.7,
		273.162,
		288.624,
		309.24,
		329.856,
		350.472,
		371.088,
		391.704,
		412.32,
		438.09,
		463.86,
		489.63,
	}
	// burst: burst = [0]
	burst = []float64{
		0.98216,
		1.055822,
		1.129484,
		1.2277,
		1.301362,
		1.375024,
		1.47324,
		1.571456,
		1.669672,
		1.767888,
		1.866104,
		1.96432,
		2.08709,
		2.20986,
		2.33263,
	}
	// burst: burstTick = [1]
	burstTick = []float64{
		0.613,
		0.658975,
		0.70495,
		0.76625,
		0.812225,
		0.8582,
		0.9195,
		0.9808,
		1.0421,
		1.1034,
		1.1647,
		1.226,
		1.302625,
		1.37925,
		1.455875,
	}
)
