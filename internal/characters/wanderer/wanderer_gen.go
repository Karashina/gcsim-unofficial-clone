// Code generated by "pipeline"; DO NOT EDIT.
package wanderer

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
	6: {"collision"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Wanderer, ValidateParamKeys)
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
	attack = [][][]float64{
		{attack_1},
		{attack_2},
		attack_3,
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.68714,
		0.74307,
		0.799,
		0.8789,
		0.93483,
		0.99875,
		1.08664,
		1.17453,
		1.26242,
		1.3583,
		1.45418,
		1.55006,
		1.64594,
		1.74182,
		1.8377,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.65016,
		0.70308,
		0.756,
		0.8316,
		0.88452,
		0.945,
		1.02816,
		1.11132,
		1.19448,
		1.2852,
		1.37592,
		1.46664,
		1.55736,
		1.64808,
		1.7388,
	}
	// attack: attack_3 = [2 2]
	attack_3 = [][]float64{
		{
			0.47644,
			0.51522,
			0.554,
			0.6094,
			0.64818,
			0.6925,
			0.75344,
			0.81438,
			0.87532,
			0.9418,
			1.00828,
			1.07476,
			1.14124,
			1.20772,
			1.2742,
		},
		{
			0.47644,
			0.51522,
			0.554,
			0.6094,
			0.64818,
			0.6925,
			0.75344,
			0.81438,
			0.87532,
			0.9418,
			1.00828,
			1.07476,
			1.14124,
			1.20772,
			1.2742,
		},
	}
	// attack: charge = [4]
	charge = []float64{
		1.3208,
		1.41986,
		1.51892,
		1.651,
		1.75006,
		1.84912,
		1.9812,
		2.11328,
		2.24536,
		2.37744,
		2.50952,
		2.6416,
		2.8067,
		2.9718,
		3.1369,
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
	// attack: plunge = [6]
	plunge = []float64{
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
	// skill: skill = [0]
	skill = []float64{
		0.952,
		1.0234,
		1.0948,
		1.19,
		1.2614,
		1.3328,
		1.428,
		1.5232,
		1.6184,
		1.7136,
		1.8088,
		1.904,
		2.023,
		2.142,
		2.261,
	}
	// skill: skillCABonus = [2]
	skillCABonus = []float64{
		1.26386,
		1.27966,
		1.29546,
		1.316,
		1.3318,
		1.3476,
		1.36814,
		1.38868,
		1.40922,
		1.42976,
		1.4503,
		1.47084,
		1.49138,
		1.51192,
		1.53246,
	}
	// skill: skillNABonus = [1]
	skillNABonus = []float64{
		1.329825,
		1.349575,
		1.369325,
		1.395,
		1.41475,
		1.4345,
		1.460175,
		1.48585,
		1.511525,
		1.5372,
		1.562875,
		1.58855,
		1.614225,
		1.6399,
		1.665575,
	}
	// burst: burst = [0]
	burst = []float64{
		1.472,
		1.5824,
		1.6928,
		1.84,
		1.9504,
		2.0608,
		2.208,
		2.3552,
		2.5024,
		2.6496,
		2.7968,
		2.944,
		3.128,
		3.312,
		3.496,
	}
)
