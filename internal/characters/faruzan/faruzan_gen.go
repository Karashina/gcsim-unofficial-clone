// Code generated by "pipeline"; DO NOT EDIT.
package faruzan

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
	validation.RegisterCharParamValidationFunc(keys.Faruzan, ValidateParamKeys)
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
		0.447295,
		0.483702,
		0.52011,
		0.572121,
		0.608529,
		0.650137,
		0.70735,
		0.764562,
		0.821774,
		0.884187,
		0.9466,
		1.009013,
		1.071427,
		1.13384,
		1.196253,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.421864,
		0.456202,
		0.49054,
		0.539594,
		0.573932,
		0.613175,
		0.667134,
		0.721094,
		0.775053,
		0.833918,
		0.892783,
		0.951648,
		1.010512,
		1.069377,
		1.128242,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.531635,
		0.574907,
		0.61818,
		0.679998,
		0.723271,
		0.772725,
		0.840725,
		0.908725,
		0.976724,
		1.050906,
		1.125088,
		1.199269,
		1.273451,
		1.347632,
		1.421814,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		0.706206,
		0.763688,
		0.82117,
		0.903287,
		0.960769,
		1.026463,
		1.116791,
		1.20712,
		1.297449,
		1.395989,
		1.494529,
		1.59307,
		1.69161,
		1.790151,
		1.888691,
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
		1.488,
		1.5996,
		1.7112,
		1.86,
		1.9716,
		2.0832,
		2.232,
		2.3808,
		2.5296,
		2.6784,
		2.8272,
		2.976,
		3.162,
		3.348,
		3.534,
	}
	// skill: vortexDmg = [1]
	vortexDmg = []float64{
		1.08,
		1.161,
		1.242,
		1.35,
		1.431,
		1.512,
		1.62,
		1.728,
		1.836,
		1.944,
		2.052,
		2.16,
		2.295,
		2.43,
		2.565,
	}
	// burst: burst = [0]
	burst = []float64{
		3.776,
		4.0592,
		4.3424,
		4.72,
		5.0032,
		5.2864,
		5.664,
		6.0416,
		6.4192,
		6.7968,
		7.1744,
		7.552,
		8.024,
		8.496,
		8.968,
	}
	// burst: burstBuff = [1]
	burstBuff = []float64{
		0.18,
		0.1935,
		0.207,
		0.225,
		0.2385,
		0.252,
		0.27,
		0.288,
		0.306,
		0.324,
		0.342,
		0.36,
		0.3825,
		0.405,
		0.4275,
	}
)
