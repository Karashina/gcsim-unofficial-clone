package simulator

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/simulation"
)

func GenerateCharacterDetails(cfg *info.ActionList) ([]*model.Character, error) {
	cpy := cfg.Copy()

	c, err := simulation.NewCore(CryptoRandSeed(), false, cpy)
	if err != nil {
		return nil, err
	}
	// 新しいシミュレーションを作成して実行
	//TODO: ここでnilでも問題ないはず
	sim, err := simulation.New(cpy, nil, c)
	if err != nil {
		return nil, err
	}

	return sim.CharacterDetails(), nil
}
