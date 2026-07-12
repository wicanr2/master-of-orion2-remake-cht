package main

import (
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/herodata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// 通用技能 bit(openorion2 gamestate.h LeaderSkills enum,COMMON_SKILLS_TYPE=0 起算):
// ASSASSIN=0…RESEARCHER=6…TRADER=9。commonSkills bitmask 對應位設起即擁有該技能。
const (
	skillCommonResearcherBit = 1 << 6 // SKILL_RESEARCHER
	skillCommonTraderBit     = 1 << 9 // SKILL_TRADER
)

// loadHerodataMercs 從玩家自備的 HERODATA.LBX 解析原版真英雄,轉成傭兵候選池(shell.Leader),
// 依等級升冪排序(開局先遇到低階、便宜、雇得起的傭兵,對齊攻略「開局只有最低階領袖」)。
// 任何一步失敗回 nil(呼叫端退回內建策展名單,音訊/資料缺失絕不擋遊戲)。
func loadHerodataMercs(res *assets.Resolver) []shell.Leader {
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
		out = append(out, shell.Leader{Name: h.Name, Skill: mercSkillLabel(h), Level: lvl, Ship: h.Ship(), Tier: 1})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Level < out[j].Level })
	return out
}

// mercSkillLabel 把 HERODATA 英雄的技能映射到 remake 的中文技能標籤。只有 remake 已建模的技能
// (科學家→研究加成 / 貿易家→收入加成,見 shell.leaderSkillIDByName)會產生實際殖民地效果;
// 其餘依類別給通用標籤(艦艇軍官→指揮官、殖民地→行政官),無 modeled 加成——這與 remake「只接
// 少數技能」的既有限制一致(誠實,不臆造效果)。
func mercSkillLabel(h herodata.Leader) string {
	switch {
	case h.CommonSkills&skillCommonResearcherBit != 0:
		return "科學家"
	case h.CommonSkills&skillCommonTraderBit != 0:
		return "貿易家"
	case h.Ship():
		return "指揮官"
	default:
		return "行政官"
	}
}
