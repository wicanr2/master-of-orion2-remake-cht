package main

import (
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/herodata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// herodatamercs.go:把 HERODATA.LBX 的真英雄轉成傭兵候選池。
//
// ============ 技能欄位是每個技能 2 bit,不是 1 bit ============
//
// 這裡原本這樣讀:
//
//	skillCommonResearcherBit = 1 << 6 // SKILL_RESEARCHER
//	skillCommonTraderBit     = 1 << 9 // SKILL_TRADER
//
// 而原版是(openorion2 `Leader::hasSkill`,gamestate.cpp:631-662):
//
//	return (skills >> (2 * skillnum)) & 0x3;
//
// **每個技能佔 2 bit**,值就是技能階(0 無 / 1 一般 / 2 進階)。所以 SKILL_RESEARCHER
// 在 bit 12-13、SKILL_TRADER 在 bit 18-19。舊的讀法拿 bit 6 當「有科學家技能」——
// 那其實是 SKILL_FAMOUS(skillnum 3)的低位;bit 9 則是 SKILL_MEGAWEALTH(skillnum 4)
// 的高位。**兩個標籤都貼錯人**,而且畫面上完全看不出來:名字是真的、等級是真的,
// 只有技能是錯的。
//
// 解碼交給 `gamedata.LeaderSkillTier`(它早就照 `hasSkill` 寫對了,只是這裡沒用它)。
//
// ============ 順帶修掉的三件事 ============
//
//  1. **Tier 寫死 1。** 進階技能(tier 2)的 +50% 加成因此一次都沒發生過。
//     現在逐技能讀真實的階。
//  2. **一位英雄只給一項技能。** 原版一位英雄可以同時有好幾項(所以才要 2 bit × N 個欄位)。
//     現在整份 `Skills` 都帶進 `shell.Leader`。
//  3. **艦艇軍官一律標成「指揮官」。** 那是類別通稱,但「指揮官」同時是 SKILL_COMMANDO
//     的譯名,而 `commandoLeaderTier` 就是掃這個字串——於是**每一位**雇來的艦艇軍官
//     都拿到了 Commando 的地面戰加成。通稱改成「艦長」,並且效果一律改用技能 id 判斷。

// loadHerodataMercs 從玩家自備的 HERODATA.LBX 解析原版真英雄,轉成傭兵候選池(shell.Leader),
// 依等級升冪排序(開局先遇到低階、便宜、雇得起的傭兵,對齊攻略「開局只有最低階領袖」)。
// 任何一步失敗回 nil(呼叫端退回內建策展名單,音訊/資料缺失絕不擋遊戲)。
func loadHerodataMercs(b *sceneBuilder, res *assets.Resolver) []shell.Leader {
	arch, err := res.OpenLBX("herodata.lbx")
	if err != nil {
		return nil
	}
	raw, err := arch.Asset(0)
	if err != nil {
		return nil
	}
	heroes, err := herodata.Parse(raw)
	if err != nil {
		return nil
	}
	out := make([]shell.Leader, 0, len(heroes))
	for _, h := range heroes {
		if h.Name == "" {
			continue
		}
		lvl := int(h.Level)
		if lvl < 1 {
			lvl = 1
		}
		if lvl > 5 {
			lvl = 5
		}
		skills := mercSkills(h)
		out = append(out, shell.Leader{
			Name:   h.Name,
			Skill:  mercSkillLabel(b, h, skills),
			Level:  lvl,
			Ship:   h.Ship(),
			Tier:   mercDisplayTier(skills),
			Skills: skills,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Level < out[j].Level })
	return out
}

// mercLeaderType 回傳這位英雄的領袖類型(gamedata.LeaderTypeCaptain / LeaderTypeAdmin)。
//
// `herodata.Leader.Type` 直接就是原版的欄位值(0=艦艇軍官),與 openorion2 的
// LEADER_TYPE_CAPTAIN/LEADER_TYPE_ADMIN 同一套編碼。
func mercLeaderType(h herodata.Leader) int {
	if h.Ship() {
		return gamedata.LeaderTypeCaptain
	}
	return gamedata.LeaderTypeAdmin
}

// mercSkills 解出這位英雄**全部**的技能與各自的階。
//
// 列舉順序照原版技能欄(專屬技能在前、通用技能在後,見 gamedata.LeaderSkillIDsFor);
// 解碼照 `Leader::hasSkill`(每技能 2 bit,並依領袖類型決定 specialSkills 讀不讀)。
func mercSkills(h herodata.Leader) []shell.LeaderSkill {
	lt := mercLeaderType(h)
	var out []shell.LeaderSkill
	for _, id := range gamedata.LeaderSkillIDsFor(lt) {
		tier := gamedata.LeaderSkillTier(id, lt, h.CommonSkills, h.SpecialSkills)
		if tier > 0 {
			out = append(out, shell.LeaderSkill{ID: id, Tier: tier})
		}
	}
	return out
}

// mercDisplayTier 回傳要放進 `shell.Leader.Tier` 的階。
//
// 那一欄是**舊路徑**用的(只有一個 `Skill` 標籤時代表那項技能的階);既然這裡已經填了
// 完整的 `Skills`,它只剩顯示意義,取第一項(= 標籤那一項)的階即可。
func mercDisplayTier(skills []shell.LeaderSkill) int {
	if len(skills) == 0 {
		return 0
	}
	return skills[0].Tier
}

// mercSkillLabel 決定畫面上顯示哪一個技能名。
//
// 取 `Skills` 的第一項——列舉順序已經照原版技能欄由上而下,所以取到的就是原版那一欄
// 最上面那個。一項技能都沒有時退回類別通稱。
//
// ⚠ 這個字串**只負責顯示**。技能效果一律走 `shell.Leader.Skills` 裡的 id,
// 不要再拿標籤去比對(見檔頭第 3 點)。
func mercSkillLabel(b *sceneBuilder, h herodata.Leader, skills []shell.LeaderSkill) string {
	if len(skills) > 0 {
		if n, ok := gamedata.LeaderSkillName(skills[0].ID); ok {
			return b.tr(n.ZH, n.EN)
		}
	}
	if h.Ship() {
		return b.tr("艦長", "Ship Officer")
	}
	return b.tr("行政官", "Administrator")
}
