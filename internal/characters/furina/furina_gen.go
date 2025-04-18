// Code generated by "pipeline"; DO NOT EDIT.
package furina

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
	5: {"collision"},
	6: {"collision"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Furina, ValidateParamKeys)
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
	// attack: arkhe = [9]
	arkhe = []float64{
		0.0946,
		0.1023,
		0.11,
		0.121,
		0.1287,
		0.1375,
		0.1496,
		0.1617,
		0.1738,
		0.187,
		0.2002,
		0.2134,
		0.2266,
		0.2398,
		0.253,
	}
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.483862,
		0.523246,
		0.56263,
		0.618893,
		0.658277,
		0.703287,
		0.765177,
		0.827066,
		0.888955,
		0.956471,
		1.023987,
		1.091502,
		1.159018,
		1.226533,
		1.294049,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.437293,
		0.472886,
		0.50848,
		0.559328,
		0.594922,
		0.6356,
		0.691533,
		0.747466,
		0.803398,
		0.864416,
		0.925434,
		0.986451,
		1.047469,
		1.108486,
		1.169504,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.5512,
		0.596065,
		0.64093,
		0.705023,
		0.749888,
		0.801162,
		0.871665,
		0.942167,
		1.012669,
		1.089581,
		1.166493,
		1.243404,
		1.320316,
		1.397227,
		1.474139,
	}
	// attack: attack_4 = [3]
	attack_4 = []float64{
		0.732978,
		0.792639,
		0.8523,
		0.93753,
		0.997191,
		1.065375,
		1.159128,
		1.252881,
		1.346634,
		1.44891,
		1.551186,
		1.653462,
		1.755738,
		1.858014,
		1.96029,
	}
	// attack: charge = [4]
	charge = []float64{
		0.74218,
		0.80259,
		0.863,
		0.9493,
		1.00971,
		1.07875,
		1.17368,
		1.26861,
		1.36354,
		1.4671,
		1.57066,
		1.67422,
		1.77778,
		1.88134,
		1.9849,
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
	// skill: skillChevalmarin = [3]
	skillChevalmarin = []float64{
		0.03232,
		0.034744,
		0.037168,
		0.0404,
		0.042824,
		0.045248,
		0.04848,
		0.051712,
		0.054944,
		0.058176,
		0.061408,
		0.06464,
		0.06868,
		0.07272,
		0.07676,
	}
	// skill: skillCrabaletta = [4]
	skillCrabaletta = []float64{
		0.08288,
		0.089096,
		0.095312,
		0.1036,
		0.109816,
		0.116032,
		0.12432,
		0.132608,
		0.140896,
		0.149184,
		0.157472,
		0.16576,
		0.17612,
		0.18648,
		0.19684,
	}
	// skill: skillOusiaBubble = [0]
	skillOusiaBubble = []float64{
		0.07864,
		0.084538,
		0.090436,
		0.0983,
		0.104198,
		0.110096,
		0.11796,
		0.125824,
		0.133688,
		0.141552,
		0.149416,
		0.15728,
		0.16711,
		0.17694,
		0.18677,
	}
	// skill: skillSingerHealFlat = [9]
	skillSingerHealFlat = []float64{
		462.2253,
		508.45425,
		558.53564,
		612.4694,
		670.2556,
		731.8942,
		797.3852,
		866.72864,
		939.9245,
		1016.9728,
		1097.8734,
		1182.6265,
		1271.232,
		1363.69,
		1460.0002,
	}
	// skill: skillSingerHealScale = [8]
	skillSingerHealScale = []float64{
		0.048,
		0.0516,
		0.0552,
		0.06,
		0.0636,
		0.0672,
		0.072,
		0.0768,
		0.0816,
		0.0864,
		0.0912,
		0.096,
		0.102,
		0.108,
		0.114,
	}
	// skill: skillUsher = [2]
	skillUsher = []float64{
		0.0596,
		0.06407,
		0.06854,
		0.0745,
		0.07897,
		0.08344,
		0.0894,
		0.09536,
		0.10132,
		0.10728,
		0.11324,
		0.1192,
		0.12665,
		0.1341,
		0.14155,
	}
	// burst: burstDMG = [0]
	burstDMG = []float64{
		0.114064,
		0.122619,
		0.131174,
		0.14258,
		0.151135,
		0.15969,
		0.171096,
		0.182502,
		0.193909,
		0.205315,
		0.216722,
		0.228128,
		0.242386,
		0.256644,
		0.270902,
	}
	// burst: burstFanfareDMGRatio = [4]
	burstFanfareDMGRatio = []float64{
		0.0007,
		0.0009,
		0.0011,
		0.0013,
		0.0015,
		0.0017,
		0.0019,
		0.0021,
		0.0023,
		0.0025,
		0.0027,
		0.0029,
		0.0031,
		0.0033,
		0.0035,
	}
	// burst: burstFanfareHBRatio = [5]
	burstFanfareHBRatio = []float64{
		0.0001,
		0.0002,
		0.0003,
		0.0004,
		0.0005,
		0.0006,
		0.0007,
		0.0008,
		0.0009,
		0.001,
		0.0011,
		0.0012,
		0.0013,
		0.0014,
		0.0015,
	}
)
