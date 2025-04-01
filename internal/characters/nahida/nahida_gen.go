// Code generated by "pipeline"; DO NOT EDIT.
package nahida

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
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Nahida, ValidateParamKeys)
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
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.403048,
		0.433277,
		0.463505,
		0.50381,
		0.534039,
		0.564267,
		0.604572,
		0.644877,
		0.685182,
		0.725486,
		0.765791,
		0.806096,
		0.856477,
		0.906858,
		0.957239,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.369744,
		0.397475,
		0.425206,
		0.46218,
		0.489911,
		0.517642,
		0.554616,
		0.59159,
		0.628565,
		0.665539,
		0.702514,
		0.739488,
		0.785706,
		0.831924,
		0.878142,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.458744,
		0.49315,
		0.527556,
		0.57343,
		0.607836,
		0.642242,
		0.688116,
		0.73399,
		0.779865,
		0.825739,
		0.871614,
		0.917488,
		0.974831,
		1.032174,
		1.089517,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		0.584064,
		0.627869,
		0.671674,
		0.73008,
		0.773885,
		0.81769,
		0.876096,
		0.934502,
		0.992909,
		1.051315,
		1.109722,
		1.168128,
		1.241136,
		1.314144,
		1.387152,
	}
	// attack: charge = [4]
	charge = []float64{
		1.32,
		1.419,
		1.518,
		1.65,
		1.749,
		1.848,
		1.98,
		2.112,
		2.244,
		2.376,
		2.508,
		2.64,
		2.805,
		2.97,
		3.135,
	}
	// skill: skillHold = [1]
	skillHold = []float64{
		1.304,
		1.4018,
		1.4996,
		1.63,
		1.7278,
		1.8256,
		1.956,
		2.0864,
		2.2168,
		2.3472,
		2.4776,
		2.608,
		2.771,
		2.934,
		3.097,
	}
	// skill: skillPress = [0]
	skillPress = []float64{
		0.984,
		1.0578,
		1.1316,
		1.23,
		1.3038,
		1.3776,
		1.476,
		1.5744,
		1.6728,
		1.7712,
		1.8696,
		1.968,
		2.091,
		2.214,
		2.337,
	}
	// skill: triKarmaAtk = [2]
	triKarmaAtk = []float64{
		1.032,
		1.1094,
		1.1868,
		1.29,
		1.3674,
		1.4448,
		1.548,
		1.6512,
		1.7544,
		1.8576,
		1.9608,
		2.064,
		2.193,
		2.322,
		2.451,
	}
	// skill: triKarmaEM = [3]
	triKarmaEM = []float64{
		2.064,
		2.2188,
		2.3736,
		2.58,
		2.7348,
		2.8896,
		3.096,
		3.3024,
		3.5088,
		3.7152,
		3.9216,
		4.128,
		4.386,
		4.644,
		4.902,
	}
	// burst: burstTriKarmaCDReduction = [2 3]
	burstTriKarmaCDReduction = [][]float64{
		{
			0.248,
			0.2666,
			0.2852,
			0.31,
			0.3286,
			0.3472,
			0.372,
			0.3968,
			0.4216,
			0.4464,
			0.4712,
			0.496,
			0.527,
			0.558,
			0.589,
		},
		{
			0.372,
			0.3999,
			0.4278,
			0.465,
			0.4929,
			0.5208,
			0.558,
			0.5952,
			0.6324,
			0.6696,
			0.7068,
			0.744,
			0.7905,
			0.837,
			0.8835,
		},
	}
	// burst: burstTriKarmaDmgBonus = [0 1]
	burstTriKarmaDmgBonus = [][]float64{
		{
			0.1488,
			0.15996,
			0.17112,
			0.186,
			0.19716,
			0.20832,
			0.2232,
			0.23808,
			0.25296,
			0.26784,
			0.28272,
			0.2976,
			0.3162,
			0.3348,
			0.3534,
		},
		{
			0.2232,
			0.23994,
			0.25668,
			0.279,
			0.29574,
			0.31248,
			0.3348,
			0.35712,
			0.37944,
			0.40176,
			0.42408,
			0.4464,
			0.4743,
			0.5022,
			0.5301,
		},
	}
	// burst: burstTriKarmaDurationExtend = [4 5]
	burstTriKarmaDurationExtend = [][]float64{
		{
			3.344,
			3.5948,
			3.8456,
			4.18,
			4.4308,
			4.6816,
			5.016,
			5.3504,
			5.6848,
			6.0192,
			6.3536,
			6.688,
			7.106,
			7.524,
			7.942,
		},
		{
			5.016,
			5.3922,
			5.7684,
			6.27,
			6.6462,
			7.0224,
			7.524,
			8.0256,
			8.5272,
			9.0288,
			9.5304,
			10.032,
			10.659,
			11.286,
			11.913,
		},
	}
)
