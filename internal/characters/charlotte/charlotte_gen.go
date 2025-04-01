// Code generated by "pipeline"; DO NOT EDIT.
package charlotte

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
	1: {"hold"},
	5: {"collision"},
	6: {"collision"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Charlotte, ValidateParamKeys)
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
	// attack: arkhe = [8]
	arkhe = []float64{
		0.11168,
		0.120056,
		0.128432,
		0.1396,
		0.147976,
		0.156352,
		0.16752,
		0.178688,
		0.189856,
		0.201024,
		0.212192,
		0.22336,
		0.23732,
		0.25128,
		0.26524,
	}
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.498456,
		0.53584,
		0.573224,
		0.62307,
		0.660454,
		0.697838,
		0.747684,
		0.79753,
		0.847375,
		0.897221,
		0.947066,
		0.996912,
		1.059219,
		1.121526,
		1.183833,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.433752,
		0.466283,
		0.498815,
		0.54219,
		0.574721,
		0.607253,
		0.650628,
		0.694003,
		0.737378,
		0.780754,
		0.824129,
		0.867504,
		0.921723,
		0.975942,
		1.030161,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.646008,
		0.694459,
		0.742909,
		0.80751,
		0.855961,
		0.904411,
		0.969012,
		1.033613,
		1.098214,
		1.162814,
		1.227415,
		1.292016,
		1.372767,
		1.453518,
		1.534269,
	}
	// attack: charge = [3]
	charge = []float64{
		1.00512,
		1.080504,
		1.155888,
		1.2564,
		1.331784,
		1.407168,
		1.50768,
		1.608192,
		1.708704,
		1.809216,
		1.909728,
		2.01024,
		2.13588,
		2.26152,
		2.38716,
	}
	// attack: collision = [5]
	collision = []float64{
		0.568288,
		0.614544,
		0.6608,
		0.72688,
		0.773136,
		0.826,
		0.898688,
		0.971376,
		1.044064,
		1.12336,
		1.202656,
		1.281952,
		1.361248,
		1.440544,
		1.51984,
	}
	// attack: highPlunge = [7]
	highPlunge = []float64{
		1.419344,
		1.534872,
		1.6504,
		1.81544,
		1.930968,
		2.063,
		2.244544,
		2.426088,
		2.607632,
		2.80568,
		3.003728,
		3.201776,
		3.399824,
		3.597872,
		3.79592,
	}
	// attack: lowPlunge = [6]
	lowPlunge = []float64{
		1.136335,
		1.228828,
		1.32132,
		1.453452,
		1.545944,
		1.65165,
		1.796995,
		1.94234,
		2.087686,
		2.246244,
		2.404802,
		2.563361,
		2.721919,
		2.880478,
		3.039036,
	}
	// skill: skillHold = [1]
	skillHold = []float64{
		1.392,
		1.4964,
		1.6008,
		1.74,
		1.8444,
		1.9488,
		2.088,
		2.2272,
		2.3664,
		2.5056,
		2.6448,
		2.784,
		2.958,
		3.132,
		3.306,
	}
	// skill: skillHoldMark = [5]
	skillHoldMark = []float64{
		0.406,
		0.43645,
		0.4669,
		0.5075,
		0.53795,
		0.5684,
		0.609,
		0.6496,
		0.6902,
		0.7308,
		0.7714,
		0.812,
		0.86275,
		0.9135,
		0.96425,
	}
	// skill: skillPress = [0]
	skillPress = []float64{
		0.672,
		0.7224,
		0.7728,
		0.84,
		0.8904,
		0.9408,
		1.008,
		1.0752,
		1.1424,
		1.2096,
		1.2768,
		1.344,
		1.428,
		1.512,
		1.596,
	}
	// skill: skillPressMark = [2]
	skillPressMark = []float64{
		0.392,
		0.4214,
		0.4508,
		0.49,
		0.5194,
		0.5488,
		0.588,
		0.6272,
		0.6664,
		0.7056,
		0.7448,
		0.784,
		0.833,
		0.882,
		0.931,
	}
	// burst: burst = [2]
	burst = []float64{
		0.77616,
		0.834372,
		0.892584,
		0.9702,
		1.028412,
		1.086624,
		1.16424,
		1.241856,
		1.319472,
		1.397088,
		1.474704,
		1.55232,
		1.64934,
		1.74636,
		1.84338,
	}
	// burst: burstDot = [5]
	burstDot = []float64{
		0.06468,
		0.069531,
		0.074382,
		0.08085,
		0.085701,
		0.090552,
		0.09702,
		0.103488,
		0.109956,
		0.116424,
		0.122892,
		0.12936,
		0.137445,
		0.14553,
		0.153615,
	}
	// burst: burstDotHealFlat = [4]
	burstDotHealFlat = []float64{
		57.447098,
		63.192608,
		69.41691,
		76.12,
		83.30189,
		90.96256,
		99.102036,
		107.7203,
		116.81735,
		126.393196,
		136.44783,
		146.98126,
		157.9935,
		169.48451,
		181.45432,
	}
	// burst: burstDotHealPer = [3]
	burstDotHealPer = []float64{
		0.09216,
		0.099072,
		0.105984,
		0.1152,
		0.122112,
		0.129024,
		0.13824,
		0.147456,
		0.156672,
		0.165888,
		0.175104,
		0.18432,
		0.19584,
		0.20736,
		0.21888,
	}
	// burst: burstInitialHealFlat = [1]
	burstInitialHealFlat = []float64{
		1608.4863,
		1769.3573,
		1943.6342,
		2131.317,
		2332.4058,
		2546.9004,
		2774.801,
		3016.1074,
		3270.8198,
		3538.9382,
		3820.4624,
		4115.3926,
		4423.7285,
		4745.4707,
		5080.6187,
	}
	// burst: burstInitialHealPer = [0]
	burstInitialHealPer = []float64{
		2.565734,
		2.758164,
		2.950595,
		3.207168,
		3.399598,
		3.592028,
		3.848602,
		4.105175,
		4.361748,
		4.618322,
		4.874895,
		5.131469,
		5.452186,
		5.772902,
		6.093619,
	}
)
