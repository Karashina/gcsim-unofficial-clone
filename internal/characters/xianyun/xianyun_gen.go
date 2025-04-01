// Code generated by "pipeline"; DO NOT EDIT.
package xianyun

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
	4: {"travel"},
	5: {"collision"},
	6: {"collision"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Xianyun, ValidateParamKeys)
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
		0.403024,
		0.433251,
		0.463478,
		0.50378,
		0.534007,
		0.564234,
		0.604536,
		0.644838,
		0.685141,
		0.725443,
		0.765746,
		0.806048,
		0.856426,
		0.906804,
		0.957182,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.388552,
		0.417693,
		0.446835,
		0.48569,
		0.514831,
		0.543973,
		0.582828,
		0.621683,
		0.660538,
		0.699394,
		0.738249,
		0.777104,
		0.825673,
		0.874242,
		0.922811,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.488776,
		0.525434,
		0.562092,
		0.61097,
		0.647628,
		0.684286,
		0.733164,
		0.782042,
		0.830919,
		0.879797,
		0.928674,
		0.977552,
		1.038649,
		1.099746,
		1.160843,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		0.649168,
		0.697856,
		0.746543,
		0.81146,
		0.860148,
		0.908835,
		0.973752,
		1.038669,
		1.103586,
		1.168502,
		1.233419,
		1.298336,
		1.379482,
		1.460628,
		1.541774,
	}
	// attack: charged = [4]
	charged = []float64{
		1.2312,
		1.32354,
		1.41588,
		1.539,
		1.63134,
		1.72368,
		1.8468,
		1.96992,
		2.09304,
		2.21616,
		2.33928,
		2.4624,
		2.6163,
		2.7702,
		2.9241,
	}
	// attack: collision = [6]
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
	// attack: highPlunge = [8]
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
	// attack: lowPlunge = [7]
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
	// skill: leap = [1 2 3]
	leap = [][]float64{
		{
			1.16,
			1.247,
			1.334,
			1.45,
			1.537,
			1.624,
			1.74,
			1.856,
			1.972,
			2.088,
			2.204,
			2.32,
			2.465,
			2.61,
			2.755,
		},
		{
			1.48,
			1.591,
			1.702,
			1.85,
			1.961,
			2.072,
			2.22,
			2.368,
			2.516,
			2.664,
			2.812,
			2.96,
			3.145,
			3.33,
			3.515,
		},
		{
			3.376,
			3.6292,
			3.8824,
			4.22,
			4.4732,
			4.7264,
			5.064,
			5.4016,
			5.7392,
			6.0768,
			6.4144,
			6.752,
			7.174,
			7.596,
			8.018,
		},
	}
	// skill: skillPress = [0]
	skillPress = []float64{
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
	}
	// burst: burst = [0]
	burst = []float64{
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
	// burst: burstDot = [1]
	burstDot = []float64{
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
	// burst: healDotFlat = [4]
	healDotFlat = []float64{
		269.6314,
		296.59833,
		325.81244,
		357.2738,
		390.98242,
		426.9383,
		465.1414,
		505.5917,
		548.2893,
		593.23413,
		640.42615,
		689.8655,
		741.552,
		795.4858,
		851.6668,
	}
	// burst: healDotP = [5]
	healDotP = []float64{
		0.43008,
		0.462336,
		0.494592,
		0.5376,
		0.569856,
		0.602112,
		0.64512,
		0.688128,
		0.731136,
		0.774144,
		0.817152,
		0.86016,
		0.91392,
		0.96768,
		1.02144,
	}
	// burst: healInstantFlat = [2]
	healInstantFlat = []float64{
		577.7816,
		635.5678,
		698.1695,
		765.58673,
		837.8195,
		914.86774,
		996.7315,
		1083.4108,
		1174.9056,
		1271.216,
		1372.3418,
		1478.2831,
		1589.04,
		1704.6124,
		1825.0002,
	}
	// burst: healInstantP = [3]
	healInstantP = []float64{
		0.9216,
		0.99072,
		1.05984,
		1.152,
		1.22112,
		1.29024,
		1.3824,
		1.47456,
		1.56672,
		1.65888,
		1.75104,
		1.8432,
		1.9584,
		2.0736,
		2.1888,
	}
)
