// Code generated by "pipeline"; DO NOT EDIT.
package lanyan

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
	validation.RegisterCharParamValidationFunc(keys.Lanyan, ValidateParamKeys)
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
		attack_3,
		{attack_4},
	}
)

var (
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.4144,
		0.44548,
		0.47656,
		0.518,
		0.54908,
		0.58016,
		0.6216,
		0.66304,
		0.70448,
		0.74592,
		0.78736,
		0.8288,
		0.8806,
		0.9324,
		0.9842,
	}
	// attack: attack_2 = [1 2]
	attack_2 = [][]float64{
		{
			0.20412,
			0.219429,
			0.234738,
			0.25515,
			0.270459,
			0.285768,
			0.30618,
			0.326592,
			0.347004,
			0.367416,
			0.387828,
			0.40824,
			0.433755,
			0.45927,
			0.484785,
		},
		{
			0.24948,
			0.268191,
			0.286902,
			0.31185,
			0.330561,
			0.349272,
			0.37422,
			0.399168,
			0.424116,
			0.449064,
			0.474012,
			0.49896,
			0.530145,
			0.56133,
			0.592515,
		},
	}
	// attack: attack_3 = [3 4]
	attack_3 = [][]float64{
		{
			0.2692,
			0.28939,
			0.30958,
			0.3365,
			0.35669,
			0.37688,
			0.4038,
			0.43072,
			0.45764,
			0.48456,
			0.51148,
			0.5384,
			0.57205,
			0.6057,
			0.63935,
		},
		{
			0.2692,
			0.28939,
			0.30958,
			0.3365,
			0.35669,
			0.37688,
			0.4038,
			0.43072,
			0.45764,
			0.48456,
			0.51148,
			0.5384,
			0.57205,
			0.6057,
			0.63935,
		},
	}
	// attack: attack_4 = [5]
	attack_4 = []float64{
		0.6456,
		0.69402,
		0.74244,
		0.807,
		0.85542,
		0.90384,
		0.9684,
		1.03296,
		1.09752,
		1.16208,
		1.22664,
		1.2912,
		1.3719,
		1.4526,
		1.5333,
	}
	// attack: charge = [6]
	charge = []float64{
		0.3784,
		0.40678,
		0.43516,
		0.473,
		0.50138,
		0.52976,
		0.5676,
		0.60544,
		0.64328,
		0.68112,
		0.71896,
		0.7568,
		0.8041,
		0.8514,
		0.8987,
	}
	// skill: ring = [0]
	ring = []float64{
		0.96256,
		1.034752,
		1.106944,
		1.2032,
		1.275392,
		1.347584,
		1.44384,
		1.540096,
		1.636352,
		1.732608,
		1.828864,
		1.92512,
		2.04544,
		2.16576,
		2.28608,
	}
	// skill: shieldAmt = [1]
	shieldAmt = []float64{
		2.7648,
		2.97216,
		3.17952,
		3.456,
		3.66336,
		3.87072,
		4.1472,
		4.42368,
		4.70016,
		4.97664,
		5.25312,
		5.5296,
		5.8752,
		6.2208,
		6.5664,
	}
	// skill: shieldFlat = [2]
	shieldFlat = []float64{
		1155.5629,
		1271.1353,
		1396.3386,
		1531.173,
		1675.6384,
		1829.7349,
		1993.4624,
		2166.8208,
		2349.8105,
		2542.4312,
		2744.6826,
		2956.5652,
		3178.0789,
		3409.2236,
		3649.9995,
	}
	// burst: burst = [0]
	burst = []float64{
		2.41064,
		2.591438,
		2.772236,
		3.0133,
		3.194098,
		3.374896,
		3.61596,
		3.857024,
		4.098088,
		4.339152,
		4.580216,
		4.82128,
		5.12261,
		5.42394,
		5.72527,
	}
)
