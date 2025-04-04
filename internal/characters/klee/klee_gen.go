// Code generated by "pipeline"; DO NOT EDIT.
package klee

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
	1: {"bounce", "mine", "mine_delay", "release"},
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
	validation.RegisterCharParamValidationFunc(keys.Klee, ValidateParamKeys)
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
		0.7216,
		0.77572,
		0.82984,
		0.902,
		0.95612,
		1.01024,
		1.0824,
		1.15456,
		1.22672,
		1.29888,
		1.373926,
		1.472064,
		1.570202,
		1.668339,
		1.766477,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.624,
		0.6708,
		0.7176,
		0.78,
		0.8268,
		0.8736,
		0.936,
		0.9984,
		1.0608,
		1.1232,
		1.188096,
		1.27296,
		1.357824,
		1.442688,
		1.527552,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.8992,
		0.96664,
		1.03408,
		1.124,
		1.19144,
		1.25888,
		1.3488,
		1.43872,
		1.52864,
		1.61856,
		1.712077,
		1.834368,
		1.956659,
		2.07895,
		2.201242,
	}
	// attack: charge = [3]
	charge = []float64{
		1.5736,
		1.69162,
		1.80964,
		1.967,
		2.08502,
		2.20304,
		2.3604,
		2.51776,
		2.67512,
		2.83248,
		2.996134,
		3.210144,
		3.424154,
		3.638163,
		3.852173,
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
	// skill: jumpy = [0]
	jumpy = []float64{
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
	// skill: mine = [3]
	mine = []float64{
		0.328,
		0.3526,
		0.3772,
		0.41,
		0.4346,
		0.4592,
		0.492,
		0.5248,
		0.5576,
		0.5904,
		0.6232,
		0.656,
		0.697,
		0.738,
		0.779,
	}
	// burst: burst = [0]
	burst = []float64{
		0.4264,
		0.45838,
		0.49036,
		0.533,
		0.56498,
		0.59696,
		0.6396,
		0.68224,
		0.72488,
		0.76752,
		0.81016,
		0.8528,
		0.9061,
		0.9594,
		1.0127,
	}
)
