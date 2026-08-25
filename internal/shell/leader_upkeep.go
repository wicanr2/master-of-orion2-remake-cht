package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// leader_upkeep.go:領袖的**每回合維護費**。
//
// ============ 原版維護費分項 ============
//
// IDA Pro 已由 sub_94A9D @ 0x94A9D 證實基本式為雇用價 / 100 向上取整、最低 1，
// Megawealth／免役條件回傳 0；Compute_Player_Maintenance_ @ 0xE20A0 將各領袖結果加總。
// prepPlayerDerived 會把本函式總額送入同一次帝國國庫結算。
//
// 手冊講「遇難領袖」事件時把這條規則講得很明白(GAME_MANUAL.pdf,Marooned Leader):
//
//	In gratitude for the rescue, this leader joins your empire for **no hiring cost**.
//	You are still expected to **pay maintenance**, however.
//
// 那句話之所以要特別寫,正是因為「免雇用費」與「免維護費」是兩件事。而 remake 的
// `grantMaroonedLeader`(discovery.go)給的領袖,先前兩樣都免。
//
// ============ 費用怎麼算 ============
//
// 每位 `ceil(hireCost/100)`,下限 1;有 Megawealth 技能(任一階)者免費。
// `hireCost` 用與 `MercHireCost` **同一條公式**——不另立一套,否則同一位領袖會出現
// 「雇用時算貴、維護時算便宜」這種對不起來的狀況。
//
// 財務危機中的資產處分由 Player_Maintenance_ @ 0xEE0B0 另行處理；本檔只計算軍官分項。

// leaderUpkeepCost 回傳一位領袖的每回合維護費。
func leaderUpkeepCost(ld Leader) int {
	exp := leaderDisplayLevelToExpLevel(ld.Level)
	hire := gamedata.LeaderHireCost(5, exp, 0) // 基準與 MercHireCost 相同
	return gamedata.LeaderMaintenanceCost(hire, leaderHasMegawealth(ld))
}

// leaderHasMegawealth 回報這位領袖有沒有 Megawealth(任一階即免維護費)。
//
// 走 `leaderSkills` 那條既有路徑(真英雄看 Skills 位元、demo 領袖退回 Skill 標籤反查),
// 不另外比對字串——技能標籤會被翻譯,拿它當識別鍵在英文模式下會查不到(見 Leader.Skills 註解)。
func leaderHasMegawealth(ld Leader) bool {
	for _, sk := range leaderSkills(ld) {
		if sk.ID == int(gamedata.SKILL_MEGAWEALTH) {
			return true
		}
	}
	return false
}

// LeaderUpkeepTotal 回傳玩家這回合要付的領袖維護費總額(供 UI/測試檢視)。
func (s *GameSession) LeaderUpkeepTotal() int {
	total := 0
	for _, ld := range s.Leaders {
		total += leaderUpkeepCost(ld)
	}
	return total
}
