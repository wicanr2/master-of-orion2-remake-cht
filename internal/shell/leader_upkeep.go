package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// leader_upkeep.go:領袖的**每回合維護費**。
//
// ============ 這條規則整支寫好了,只是沒有人扣過錢 ============
//
// `gamedata.LeaderMaintenanceCost` 從 openorion2 的 `GameState::leaderMaintenanceCost`
// 移植過來、有單元測試、註解連 `LEADER_ID_LOKNAR` 那個不移植的特例都交代了
// ——**零個生產端呼叫**。一局跑 300 回合的覆蓋率量下來是 0.0%。
//
// 所以 remake 的領袖:雇用時付一次錢(`MercHireCost` → `engine.HireLeader`),
// 之後**永久免費**。
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
// ============ 誠實留白 ============
//
//   - **付不出來時只是扣到 0,領袖不會離職。** 原版錢不夠會怎樣沒有查到規則
//     (手冊只說要付,沒說不付的後果),所以這裡不自己發明一個懲罰。
//     國庫本來就可能為負(見 session.go 的既有處理),這裡沿用「夾在 0」的既有慣例。
//   - **AI 對手不付。** AI 沒有領袖模型(`AIOpponent` 沒有 Leaders 欄位),
//     不是這裡漏掉,是那一整層還不存在。

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

// advanceLeaderUpkeep 從國庫扣掉這回合的領袖維護費,回傳實際扣掉的金額。
//
// 國庫已經是負的就不再往下扣(比照 session.go 對 bcLoss 的既有處理:那裡的註解寫明
// 「BC 為負時若只判斷 bcLoss > BC 會把損失夾成負值,反而變成加錢」)。
func (s *GameSession) advanceLeaderUpkeep() int {
	cost := s.LeaderUpkeepTotal()
	if cost <= 0 {
		return 0
	}
	if s.Player.BC <= 0 {
		return 0
	}
	if cost > s.Player.BC {
		cost = s.Player.BC
	}
	s.Player.BC -= cost
	return cost
}
