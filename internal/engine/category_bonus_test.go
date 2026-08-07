package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func catBonusBase() ColonyState {
	return ColonyState{
		Population: 9, PopMax: 12, Farmers: 3, Workers: 3, Scientists: 3,
		FoodPerFarmer: 4, IndustryPerWorker: 4, ResearchPerScientist: 4,
		PlanetSize: 2, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
}

// 三個分項百分比各自**只動自己那一項**——士氣是三項一起動,這三個不是。
func TestCategoryBonusesAffectOnlyTheirOwnOutput(t *testing.T) {
	off := RunColonyTurn(catBonusBase())

	food := catBonusBase()
	food.FoodBonusPercent = 100 // 好驗的整數倍
	gotFood := RunColonyTurn(food)
	if gotFood.Food <= off.Food {
		t.Errorf("農業官應提高食物:%d → %d", off.Food, gotFood.Food)
	}
	if gotFood.GrossIndustry != off.GrossIndustry || gotFood.Research != off.Research {
		t.Errorf("農業官不該動工業/研究:工業 %d→%d、研究 %d→%d",
			off.GrossIndustry, gotFood.GrossIndustry, off.Research, gotFood.Research)
	}

	ind := catBonusBase()
	ind.IndustryBonusPercent = 100
	gotInd := RunColonyTurn(ind)
	if gotInd.GrossIndustry <= off.GrossIndustry {
		t.Errorf("勞工官應提高工業:%d → %d", off.GrossIndustry, gotInd.GrossIndustry)
	}
	if gotInd.Food != off.Food || gotInd.Research != off.Research {
		t.Errorf("勞工官不該動食物/研究:食物 %d→%d、研究 %d→%d",
			off.Food, gotInd.Food, off.Research, gotInd.Research)
	}

	res := catBonusBase()
	res.ResearchBonusPercent = 100
	gotRes := RunColonyTurn(res)
	if gotRes.Research <= off.Research {
		t.Errorf("科學官應提高研究:%d → %d", off.Research, gotRes.Research)
	}
	if gotRes.Food != off.Food || gotRes.GrossIndustry != off.GrossIndustry {
		t.Errorf("科學官不該動食物/工業:食物 %d→%d、工業 %d→%d",
			off.Food, gotRes.Food, off.GrossIndustry, gotRes.GrossIndustry)
	}
}

// 正對照:士氣**是**三項一起動——少了這條,「分項百分比其實什麼都沒接上」也會通過上面那支。
func TestMoraleStillAffectsAllThreeOutputs(t *testing.T) {
	off := RunColonyTurn(catBonusBase())
	m := catBonusBase()
	m.MoralePercent = 100
	got := RunColonyTurn(m)
	if got.Food <= off.Food || got.GrossIndustry <= off.GrossIndustry || got.Research <= off.Research {
		t.Errorf("士氣應三項一起提高:食物 %d→%d、工業 %d→%d、研究 %d→%d",
			off.Food, got.Food, off.GrossIndustry, got.GrossIndustry, off.Research, got.Research)
	}
}

// 固定加成不吃百分比(與士氣/重力的既有處理一致)。
func TestCategoryBonusesDoNotScaleFlatValues(t *testing.T) {
	base := catBonusBase()
	base.Farmers = 0 // 只剩固定加成
	base.FlatFood = 10
	off := RunColonyTurn(base)
	on := base
	on.FoodBonusPercent = 100
	got := RunColonyTurn(on)
	if got.Food != off.Food {
		t.Errorf("農夫為 0 時食物全來自 FlatFood,不該被百分比放大:%d → %d", off.Food, got.Food)
	}
}
