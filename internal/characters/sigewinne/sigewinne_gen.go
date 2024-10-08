// Code generated by "pipeline"; DO NOT EDIT.
package sigewinne

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
	0: {"hold"},
	1: {"hold", "travel", "weakspot"},
}

func init() {
	base = &model.AvatarData{}
	err := prototext.Unmarshal(pbData, base)
	if err != nil {
		panic(err)
	}
	validation.RegisterCharParamValidationFunc(keys.Sigewinne, ValidateParamKeys)
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
		{attack_2},
		{attack_3},
	}
)

var (
	// attack: aim = [3]
	aim = []float64{
		0.439,
		0.474559,
		0.510557,
		0.561481,
		0.59704,
		0.637867,
		0.694059,
		0.750251,
		0.806443,
		0.867903,
		0.928924,
		0.990384,
		1.051405,
		1.051405,
		1.051405,
	}
	// attack: attack_1 = [0]
	attack_1 = []float64{
		0.526,
		0.568606,
		0.611738,
		0.672754,
		0.71536,
		0.764278,
		0.831606,
		0.898934,
		0.966262,
		1.039902,
		1.113016,
		1.186656,
		1.25977,
		1.25977,
		1.25977,
	}
	// attack: attack_2 = [1]
	attack_2 = []float64{
		0.511,
		0.552391,
		0.594293,
		0.653569,
		0.69496,
		0.742483,
		0.807891,
		0.873299,
		0.938707,
		1.010247,
		1.081276,
		1.152816,
		1.223845,
		1.223845,
		1.223845,
	}
	// attack: attack_3 = [2]
	attack_3 = []float64{
		0.266,
		0.287546,
		0.309358,
		0.340214,
		0.36176,
		0.386498,
		0.420546,
		0.454594,
		0.488642,
		0.525882,
		0.562856,
		0.600096,
		0.63707,
		0.63707,
		0.63707,
	}
	// attack: Mini-Stration Bubble DMG = [4]
	ministration = []float64{
		0.228,
		0.2451,
		0.2622,
		0.285,
		0.3021,
		0.3192,
		0.342,
		0.3648,
		0.3876,
		0.4104,
		0.434112,
		0.46512,
		0.496128,
		0.496128,
		0.496128,
	}
	// attack: fullaim = [5]
	fullaim = []float64{
		1.141,
		1.226575,
		1.31215,
		1.42625,
		1.511825,
		1.5974,
		1.7115,
		1.8256,
		1.9397,
		2.0538,
		2.1679,
		2.282,
		2.424625,
		2.424625,
		2.424625,
	}
	// skill: skill = [0]
	skill = []float64{
		0.0228,
		0.02451,
		0.02622,
		0.0285,
		0.03021,
		0.03192,
		0.0342,
		0.03648,
		0.03876,
		0.04104,
		0.04332,
		0.0456,
		0.04845,
		0.04845,
		0.04845,
	}
	// skill: skill heal = [1]
	skillheal = []float64{
		0.028,
		0.0301,
		0.0322,
		0.035,
		0.0371,
		0.0392,
		0.042,
		0.0448,
		0.0476,
		0.0504,
		0.0532,
		0.056,
		0.0595,
		0.0595,
		0.0595,
	}
	// skill: skill heal = [2]
	skillbonus = []float64{
		269,
		296,
		325,
		357,
		390,
		426,
		465,
		505,
		548,
		593,
		640,
		689,
		741,
		741,
		741,
	}
	// skill: skill heal = [3]
	skillaligned = []float64{
		0.0068,
		0.00731,
		0.00782,
		0.0085,
		0.00901,
		0.00952,
		0.0102,
		0.01088,
		0.01156,
		0.01224,
		0.01292,
		0.0136,
		0.01445,
		0.01445,
		0.01445,
	}
	// burst: burst = [0]
	burst = []float64{
		0.118,
		0.12685,
		0.1357,
		0.1475,
		0.15635,
		0.1652,
		0.177,
		0.1888,
		0.2006,
		0.2124,
		0.2242,
		0.236,
		0.25075,
		0.25075,
		0.25075,
	}
)
