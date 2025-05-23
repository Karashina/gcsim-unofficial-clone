// Code generated by "pipeline"; DO NOT EDIT.
package emilie

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
	2: {"travel"},
	5: {"collision"},
	6: {"collision"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Emilie, ValidateParamKeys)
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
	skillLumidouce = [][]float64{
		skillLumidouce_1,
		skillLumidouce_2,
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.485608,
		0.525134,
		0.56466,
		0.621126,
		0.660652,
		0.705825,
		0.767938,
		0.83005,
		0.892163,
		0.959922,
		1.027681,
		1.09544,
		1.1632,
		1.230959,
		1.298718,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.448954,
		0.485497,
		0.52204,
		0.574244,
		0.610787,
		0.65255,
		0.709974,
		0.767399,
		0.824823,
		0.887468,
		0.950113,
		1.012758,
		1.075402,
		1.138047,
		1.200692,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.593004,
		0.641272,
		0.68954,
		0.758494,
		0.806762,
		0.861925,
		0.937774,
		1.013624,
		1.089473,
		1.172218,
		1.254963,
		1.337708,
		1.420452,
		1.503197,
		1.585942,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		0.751029,
		0.81216,
		0.87329,
		0.960619,
		1.021749,
		1.091613,
		1.187674,
		1.283736,
		1.379798,
		1.484593,
		1.589388,
		1.694183,
		1.798977,
		1.903772,
		2.008567,
	}
	// attack: charge = [4]
	charge = []float64{
		0.91332,
		0.98766,
		1.062,
		1.1682,
		1.24254,
		1.3275,
		1.44432,
		1.56114,
		1.67796,
		1.8054,
		1.93284,
		2.06028,
		2.18772,
		2.31516,
		2.4426,
	}
	// attack: collision = [6]
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
	// attack: highPlunge = [8]
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
	// attack: lowPlunge = [7]
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
	// skill: skillArke = [4]
	skillArke = []float64{
		0.3852,
		0.41409,
		0.44298,
		0.4815,
		0.51039,
		0.53928,
		0.5778,
		0.61632,
		0.65484,
		0.69336,
		0.73188,
		0.7704,
		0.81855,
		0.8667,
		0.91485,
	}
	// skill: skillArkeCD = [5]
	skillArkeCD = []float64{
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
		10,
	}
	// skill: skillCD = [6]
	skillCD = []float64{
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
		14,
	}
	// skill: skillDMG = [0]
	skillDMG = []float64{
		0.4708,
		0.50611,
		0.54142,
		0.5885,
		0.62381,
		0.65912,
		0.7062,
		0.75328,
		0.80036,
		0.84744,
		0.89452,
		0.9416,
		1.00045,
		1.0593,
		1.11815,
	}
	// skill: skillDuration = [3]
	skillDuration = []float64{
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
		22,
	}
	// skill: skillLumidouce_1 = [1]
	skillLumidouce_1 = []float64{
		0.396,
		0.4257,
		0.4554,
		0.495,
		0.5247,
		0.5544,
		0.594,
		0.6336,
		0.6732,
		0.7128,
		0.7524,
		0.792,
		0.8415,
		0.891,
		0.9405,
	}
	// skill: skillLumidouce_2 = [2]
	skillLumidouce_2 = []float64{
		0.84,
		0.903,
		0.966,
		1.05,
		1.113,
		1.176,
		1.26,
		1.344,
		1.428,
		1.512,
		1.596,
		1.68,
		1.785,
		1.89,
		1.995,
	}
	// burst: burstCD = [2]
	burstCD = []float64{
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
		13.5,
	}
	// burst: burstDMG = [0]
	burstDMG = []float64{
		2.172,
		2.3349,
		2.4978,
		2.715,
		2.8779,
		3.0408,
		3.258,
		3.4752,
		3.6924,
		3.9096,
		4.1268,
		4.344,
		4.6155,
		4.887,
		5.1585,
	}
	// burst: burstDuration = [1]
	burstDuration = []float64{
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
		2.8,
	}
)
