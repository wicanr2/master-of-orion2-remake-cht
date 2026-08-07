package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// ground_types.go:地面戰的部隊類型索引。
//
// ⚠ **2026-08-07 訂正:先前這裡把陸戰隊排在 0、戰車營排在 1,是反的。**
// `Compute_Ground_Combat_Info_` @ 0xEC3CE 的四個 case 顯示類型 0 才是最強的那種
// (+10 攻擊、+1 耐受),而手冊說裝甲是 tank battalions、陸戰隊是基準。
// 先前兩種填同一個攻擊力,所以順序不影響結果;現在接上逐類型的差之後就會差很多。
//
// 真正的定義在 `gamedata.GroundTypeArmor/Marines/Militia`,這裡只是轉個名字給本套件用。
const (
	groundTypeTanks   = gamedata.GroundTypeArmor   // 0:裝甲 / 戰車營
	groundTypeMarines = gamedata.GroundTypeMarines // 1:陸戰隊
)
