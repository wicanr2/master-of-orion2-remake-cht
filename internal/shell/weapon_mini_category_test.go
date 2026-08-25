package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestDesignUsesSeparateWeaponAndSpecialMiniCategories(t *testing.T) {
	torpedo := componentIndexByName(WeaponOptions, "質子魚雷")
	achilles := componentIndexByName(SpecialOptions, "阿基里斯瞄準器")
	used := shipDesignSpaceUsedWithMiniLevel("巡洋艦", torpedo, 0, 0, achilles, nil,
		gamedata.ARC_FWD, 2, 2, 2)
	// 質子魚雷 base 30 與阿基里斯巡洋艦 base 25 都走 category1 level2=70%。
	// 魚雷固定全向再 +50%：21→31；特殊裝置為 18，合計 49。
	if used != 49 {
		t.Fatalf("分類微型化後佔格=%d want 49", used)
	}
}

func TestFixedSpecialDeviceDoesNotShrinkButCostDoes(t *testing.T) {
	noWeapon := 0
	augmented := componentIndexByName(SpecialOptions, "增強引擎")
	space := shipDesignSpaceUsedWithMiniLevel("巡洋艦", noWeapon, 0, 0, augmented, nil,
		gamedata.ARC_FWD, 5, 5, 255)
	if space != gamedata.SpecialDeviceSpace(gamedata.SPEC_AUGMENTED_ENGINES, gamedata.SHIP_CRUISER) {
		t.Fatalf("增強引擎 category2 不得縮小，得到 %d", space)
	}
	base := designCostWithMiniLevel("巡洋艦", noWeapon, 0, 0, augmented, nil,
		gamedata.ARC_FWD, 0, 0, 255)
	mini := designCostWithMiniLevel("巡洋艦", noWeapon, 0, 0, augmented, nil,
		gamedata.ARC_FWD, 0, 5, 255)
	if mini >= base {
		t.Fatalf("category2 只固定佔格，成本仍應微型化: base=%d mini=%d", base, mini)
	}
}
