// Code generated by "pipeline"; DO NOT EDIT.
package kachina

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
	1: {"hold"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Kachina, ValidateParamKeys)
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
		{attack_4},
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.494,
		0.534014,
		0.574522,
		0.631826,
		0.67184,
		0.717782,
		0.781014,
		0.844246,
		0.907478,
		0.976638,
		1.045304,
		1.114464,
		1.18313,
		1.18313,
		1.18313,
	}
	// attack: attack_2 = [1 2]
	attack_2 = [][]float64{
		{
			0.276,
			0.298356,
			0.320988,
			0.353004,
			0.37536,
			0.401028,
			0.436356,
			0.471684,
			0.507012,
			0.545652,
			0.584016,
			0.622656,
			0.66102,
			0.66102,
			0.66102,
		},
		{
			0.306,
			0.330786,
			0.355878,
			0.391374,
			0.41616,
			0.444618,
			0.483786,
			0.522954,
			0.562122,
			0.604962,
			0.647496,
			0.690336,
			0.73287,
			0.73287,
			0.73287,
		},
	}
	// attack: attack_3 = [3]
	attack_3 = []float64{
		0.704,
		0.761024,
		0.818752,
		0.900416,
		0.95744,
		1.022912,
		1.113024,
		1.203136,
		1.293248,
		1.391808,
		1.489664,
		1.588224,
		1.68608,
		1.68608,
		1.68608,
	}
	// attack: attack_4 = [4]
	attack_4 = []float64{
		0.774,
		0.836694,
		0.900162,
		0.989946,
		1.05264,
		1.124622,
		1.223694,
		1.322766,
		1.421838,
		1.530198,
		1.637784,
		1.746144,
		1.85373,
		1.85373,
		1.85373,
	}
	// attack: charge = [5]
	charge = []float64{
		1.127,
		1.218287,
		1.310701,
		1.441433,
		1.53272,
		1.637531,
		1.781787,
		1.926043,
		2.070299,
		2.228079,
		2.384732,
		2.542512,
		2.699165,
		2.699165,
		2.699165,
	}
	// skill: skillRide = [0]
	skillRide = []float64{
		0.878,
		0.94385,
		1.0097,
		1.0975,
		1.16335,
		1.2292,
		1.317,
		1.4048,
		1.4926,
		1.5804,
		1.6682,
		1.756,
		1.86575,
		1.86575,
		1.86575,
	}
	// skill: skill = [1]
	skillIndependent = []float64{
		0.638,
		0.68585,
		0.7337,
		0.7975,
		0.84535,
		0.8932,
		0.957,
		1.0208,
		1.0846,
		1.1484,
		1.2122,
		1.276,
		1.35575,
		1.35575,
		1.35575,
	}
	// burst: burst = [0]
	burst = []float64{
		3.806,
		4.137,
		4.425,
		4.810,
		5.099,
		5.387,
		5.772,
		6.157,
		6.542,
		6.926,
		7.311,
		7.696,
		8.177,
		8.177,
		8.177,
	}
)