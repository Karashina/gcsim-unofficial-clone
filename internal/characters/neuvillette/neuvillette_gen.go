// Code generated by "pipeline"; DO NOT EDIT.
package neuvillette

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
	4: {"short", "ticks"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Neuvillette, ValidateParamKeys)
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
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.545768,
		0.586701,
		0.627633,
		0.68221,
		0.723143,
		0.764075,
		0.818652,
		0.873229,
		0.927806,
		0.982382,
		1.036959,
		1.091536,
		1.159757,
		1.227978,
		1.296199,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.462456,
		0.49714,
		0.531824,
		0.57807,
		0.612754,
		0.647438,
		0.693684,
		0.73993,
		0.786175,
		0.832421,
		0.878666,
		0.924912,
		0.982719,
		1.040526,
		1.098333,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.723376,
		0.777629,
		0.831882,
		0.90422,
		0.958473,
		1.012726,
		1.085064,
		1.157402,
		1.229739,
		1.302077,
		1.374414,
		1.446752,
		1.537174,
		1.627596,
		1.718018,
	}
	// attack: charge = [3]
	charge = []float64{
		1.368,
		1.4706,
		1.5732,
		1.71,
		1.8126,
		1.9152,
		2.052,
		2.1888,
		2.3256,
		2.4624,
		2.5992,
		2.736,
		2.907,
		3.078,
		3.249,
	}
	// attack: chargeJudgement = [4]
	chargeJudgement = []float64{
		0.073186,
		0.079143,
		0.0851,
		0.09361,
		0.099567,
		0.106375,
		0.115736,
		0.125097,
		0.134458,
		0.14467,
		0.154882,
		0.165094,
		0.175306,
		0.185518,
		0.19573,
	}
	// skill: skill = [0]
	skill = []float64{
		0.12864,
		0.138288,
		0.147936,
		0.1608,
		0.170448,
		0.180096,
		0.19296,
		0.205824,
		0.218688,
		0.231552,
		0.244416,
		0.25728,
		0.27336,
		0.28944,
		0.30552,
	}
	// skill: thorn = [1]
	thorn = []float64{
		0.208,
		0.2236,
		0.2392,
		0.26,
		0.2756,
		0.2912,
		0.312,
		0.3328,
		0.3536,
		0.3744,
		0.3952,
		0.416,
		0.442,
		0.468,
		0.494,
	}
	// burst: burst = [0]
	burst = []float64{
		0.222578,
		0.239272,
		0.255965,
		0.278223,
		0.294916,
		0.31161,
		0.333868,
		0.356125,
		0.378383,
		0.400641,
		0.422899,
		0.445157,
		0.472979,
		0.500801,
		0.528624,
	}
	// burst: burstWaterfall = [1]
	burstWaterfall = []float64{
		0.091055,
		0.097884,
		0.104713,
		0.113818,
		0.120647,
		0.127477,
		0.136582,
		0.145688,
		0.154793,
		0.163898,
		0.173004,
		0.182109,
		0.193491,
		0.204873,
		0.216255,
	}
)
