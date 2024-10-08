// Code generated by "pipeline"; DO NOT EDIT.
package xilonen

import (
	_ "embed"

	"fmt"
	"slices"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/gcs/validation"
	"github.com/genshinsim/gcsim/pkg/model"
	"google.golang.org/protobuf/encoding/prototext"
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
	validation.RegisterCharParamValidationFunc(keys.Xilonen, ValidateParamKeys)
}

func ValidateParamKeys(a action.Action, keys []string) error {
	valid, ok := paramKeysValidation[a]
	if !ok {
		return nil
	}
	for _, v := range keys {
		if !slices.Contains(valid, v) {
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
		attack_2,
		{attack_3},
	}
	bladeroller = [][]float64{
		bladeroller_1,
		bladeroller_2,
		bladeroller_3,
		bladeroller_4,
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.518,
		0.559958,
		0.602434,
		0.662522,
		0.70448,
		0.752654,
		0.818958,
		0.885262,
		0.951566,
		1.024086,
		1.096088,
		1.168608,
		1.24061,
		1.24061,
		1.24061,		
	}
	// attack: attack_2 = [1 2]
	attack_2 = [][]float64{
		{
			0.0274,
			0.0296194,
			0.0318662,
			0.0350446,
			0.037264,
			0.0398122,
			0.0433194,
			0.0468266,
			0.0503338,
			0.0541698,
			0.0579784,
			0.0618144,
			0.065623,
			0.065623,
			0.065623,			
		},
		{
			0.0274,
			0.0296194,
			0.0318662,
			0.0350446,
			0.037264,
			0.0398122,
			0.0433194,
			0.0468266,
			0.0503338,
			0.0541698,
			0.0579784,
			0.0618144,
			0.065623,
			0.065623,
			0.065623,			
		},
	}
	// attack: attack_3 = [3]
	attack_3 = []float64{
		0.0274,
		0.0296194,
		0.0318662,
		0.0350446,
		0.037264,
		0.0398122,
		0.0433194,
		0.0468266,
		0.0503338,
		0.0541698,
		0.0579784,
		0.0618144,
		0.065623,
		0.065623,
		0.065623,			
	}
	// attack: charge = [4]
	charge = []float64{
		0.913,
		0.986953,
		1.061819,
		1.167727,
		1.24168,
		1.326589,
		1.443453,
		1.560317,
		1.677181,
		1.805001,
		1.931908,
		2.059728,
		2.186635,
		2.186635,
		2.186635,		
	}
	// attack: bladeroller_1 = [5]
	bladeroller_1 = []float64{
		0.56,
		0.60536,
		0.65128,
		0.71624,
		0.7616,
		0.81368,
		0.88536,
		0.95704,
		1.02872,
		1.10712,
		1.18496,
		1.26336,
		1.3412,
		1.3412,
		1.3412,		
	}
	// attack: bladeroller_2 = [6]
	bladeroller_2 = []float64{
		0.55,
		0.59455,
		0.63965,
		0.70345,
		0.748,
		0.79915,
		0.86955,
		0.93995,
		1.01035,
		1.08735,
		1.1638,
		1.2408,
		1.31725,
		1.31725,
		1.31725,		
	}
	// attack: bladeroller_3 = [7]
	bladeroller_3 = []float64{
		0.658,
		0.711298,
		0.765254,
		0.841582,
		0.89488,
		0.956074,
		1.040298,
		1.124522,
		1.208746,
		1.300866,
		1.392328,
		1.484448,
		1.57591,
		1.57591,
		1.57591,		
	}
	// attack: bladeroller_4 = [8]
	bladeroller_4 = []float64{
		0.86,
		0.92966,
		1.00018,
		1.09994,
		1.1696,
		1.24958,
		1.35966,
		1.46974,
		1.57982,
		1.70022,
		1.81976,
		1.94016,
		2.0597,
		2.0597,
		2.0597,		
	}
	// attack: lowPlunge = [9]
	lowPlunge = []float64{
		1.278,
		1.381518,
		1.486314,
		1.634562,
		1.73808,
		1.856934,
		2.020518,
		2.184102,
		2.347686,
		2.526606,
		2.704248,
		2.883168,
		3.06081,
		3.06081,
		3.06081,		
	}
	// attack: lowPlunge = [10]
	highPlunge = []float64{
		1.597,
		1.726357,
		1.857311,
		2.042563,
		2.17192,
		2.320441,
		2.524857,
		2.729273,
		2.933689,
		3.157269,
		3.379252,
		3.602832,
		3.824815,
		3.824815,
		3.824815,		
	}
	// attack: lowPlunge = [11]
	collision = []float64{
		0.639,
		0.690759,
		0.743157,
		0.817281,
		0.86904,
		0.928467,
		1.010259,
		1.092051,
		1.173843,
		1.263303,
		1.352124,
		1.441584,
		1.530405,
		1.530405,
		1.530405,		
	}
	// skill: skill = [0]
	skill = []float64{
		1.792,
		1.9264,
		2.0608,
		2.24,
		2.3744,
		2.5088,
		2.688,
		2.8672,
		3.0464,
		3.2256,
		3.4048,
		3.584,
		3.808,
		3.808,
		3.808,		
	}
	// skill: skillRes = [1]
	skillRes = []float64{
		0.09,
		0.12,
		0.15,
		0.18,
		0.21,
		0.24,
		0.27,
		0.30,
		0.33,
		0.36,
		0.39,
		0.42,
		0.45,
		0.45,
		0.45,
	}
	// burst: burst = [0]
	burst = []float64{
		2.813,
		3.023975,
		3.23495,
		3.51625,
		3.727225,
		3.9382,
		4.2195,
		4.5008,
		4.7821,
		5.0634,
		5.3447,
		5.626,
		5.977625,
		5.977625,
		5.977625,		
	}
	// burst: burstheal = [1]
	burstheal = []float64{
		1.04,
		1.118,
		1.196,
		1.3,
		1.378,
		1.456,
		1.56,
		1.664,
		1.768,
		1.872,
		1.976,
		2.08,
		2.21,
		2.21,
		2.21,			
	}
	// burst: burstconst = [2]
	bursthealconst = []float64{
		501,
		551,
		605,
		664,
		726,
		793,
		864,
		939,
		1018,
		1102,
		1189,
		1281,
		1377,
		1377,
		1377,	
	}
)
