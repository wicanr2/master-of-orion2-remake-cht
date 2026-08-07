package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/herodata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

// commonBits 把「第 n 個通用技能是第 t 階」寫成技能位元。
// 原版每個技能佔 **2 bit**(Leader::hasSkill:`(skills >> (2*skillnum)) & 0x3`)。
func commonBits(pairs ...[2]int) uint32 {
	var v uint32
	for _, p := range pairs {
		v |= uint32(p[1]) << uint(2*p[0])
	}
	return v
}

// 這一支釘的就是第 103 項修掉的那個錯:技能位元是 2 bit 一組,不是 1 bit。
//
// 舊程式用 `1 << 6` 當「有科學家技能」。bit 6 其實是 **skillnum 3(SKILL_FAMOUS)** 的低位,
// 所以真正的科學家會被漏掉、名人會被當成科學家——而畫面上名字與等級都是對的,
// 只有技能悄悄貼錯人。
func TestSkillBitsAreTwoBitsPerSkill(t *testing.T) {
	researcher := herodata.Leader{
		Name: "科學家", Type: gamedata.LeaderTypeAdmin,
		CommonSkills: commonBits([2]int{int(gamedata.SKILL_RESEARCHER), 1}),
	}
	skills := mercSkills(researcher)
	if len(skills) != 1 || skills[0].ID != int(gamedata.SKILL_RESEARCHER) || skills[0].Tier != 1 {
		t.Fatalf("SKILL_RESEARCHER 應在 bit 12-13,解出 %+v", skills)
	}

	// 舊的錯誤遮罩 `1 << 6` 指到的是 SKILL_FAMOUS,不是科學家。
	famous := herodata.Leader{
		Name: "名人", Type: gamedata.LeaderTypeAdmin, CommonSkills: 1 << 6,
	}
	got := mercSkills(famous)
	if len(got) != 1 || got[0].ID != int(gamedata.SKILL_FAMOUS) {
		t.Fatalf("bit 6 應是 SKILL_FAMOUS(%#x),解出 %+v", int(gamedata.SKILL_FAMOUS), got)
	}
	for _, sk := range got {
		if sk.ID == int(gamedata.SKILL_RESEARCHER) {
			t.Error("bit 6 不該解成科學家——那正是舊遮罩的錯法")
		}
	}
}

// 階讀得出來:2 = 進階(+50%)。Tier 寫死 1 的時候這一格永遠是 1。
func TestAdvancedTierIsDecoded(t *testing.T) {
	h := herodata.Leader{
		Name: "老手", Type: gamedata.LeaderTypeAdmin,
		CommonSkills: commonBits([2]int{int(gamedata.SKILL_TRADER), 2}),
	}
	skills := mercSkills(h)
	if len(skills) != 1 || skills[0].Tier != 2 {
		t.Fatalf("應解出進階階(2),得到 %+v", skills)
	}
	if got := mercDisplayTier(skills); got != 2 {
		t.Errorf("顯示用的 Tier 應跟著是 2,得到 %d", got)
	}
	// 正對照:沒有技能時顯示階是 0,不是預設的 1。
	if got := mercDisplayTier(nil); got != 0 {
		t.Errorf("無技能的領袖 Tier 應為 0,得到 %d", got)
	}
}

// specialSkills 要看領袖類型才知道該當 captain 還是 admin 技能解讀——
// 同一組位元在兩種領袖身上是**不同的技能**。
func TestSpecialSkillsAreGatedByLeaderType(t *testing.T) {
	bits := uint32(1) // skillnum 0,階 1

	captain := herodata.Leader{Name: "艦長", Type: gamedata.LeaderTypeCaptain, SpecialSkills: bits}
	if s := mercSkills(captain); len(s) != 1 || s[0].ID != int(gamedata.SKILL_ENGINEER) {
		t.Errorf("艦艇軍官的專屬 skillnum 0 應是 SKILL_ENGINEER,得到 %+v", s)
	}

	admin := herodata.Leader{Name: "行政官", Type: gamedata.LeaderTypeAdmin, SpecialSkills: bits}
	if s := mercSkills(admin); len(s) != 1 || s[0].ID != int(gamedata.SKILL_ENVIRONMENTALIST) {
		t.Errorf("殖民地領袖的專屬 skillnum 0 應是 SKILL_ENVIRONMENTALIST,得到 %+v", s)
	}
}

// 一位英雄可以同時有好幾項技能,而且專屬技能排在通用技能前面(原版技能欄的順序)。
func TestMercSkillsListsAllSkillsSpecialFirst(t *testing.T) {
	h := herodata.Leader{
		Name: "多才", Type: gamedata.LeaderTypeAdmin,
		SpecialSkills: 1 << (2 * 1), // admin skillnum 1 = SKILL_FARMING_LEADER
		CommonSkills:  commonBits([2]int{int(gamedata.SKILL_RESEARCHER), 1}),
	}
	skills := mercSkills(h)
	if len(skills) != 2 {
		t.Fatalf("應解出 2 項技能,得到 %+v", skills)
	}
	if skills[0].ID != int(gamedata.SKILL_FARMING_LEADER) {
		t.Errorf("專屬技能應排在前面,第一項是 %#x", skills[0].ID)
	}
	if skills[1].ID != int(gamedata.SKILL_RESEARCHER) {
		t.Errorf("通用技能應排在後面,第二項是 %#x", skills[1].ID)
	}
}

// 顯示標籤會被翻譯,所以它**不能**當識別鍵——這支確認兩種語言下標籤不同、但技能 id 一樣。
func TestLabelIsTranslatedButSkillIDIsNot(t *testing.T) {
	h := herodata.Leader{
		Name: "Von Neumann", Type: gamedata.LeaderTypeAdmin,
		CommonSkills: commonBits([2]int{int(gamedata.SKILL_RESEARCHER), 1}),
	}
	skills := mercSkills(h)

	en := mercSkillLabel(&sceneBuilder{}, h, skills) // Lang 零值 = 英文
	zh := mercSkillLabel(&sceneBuilder{lang: i18n.Traditional}, h, skills)
	if en != "Researcher" {
		t.Errorf("英文標籤應是 Researcher,得到 %q", en)
	}
	if zh != "科學家" {
		t.Errorf("中文標籤應是 科學家,得到 %q", zh)
	}
	if len(skills) != 1 || skills[0].ID != int(gamedata.SKILL_RESEARCHER) {
		t.Errorf("兩種語言下技能 id 都應該是 SKILL_RESEARCHER,得到 %+v", skills)
	}
}

// 沒有任何技能的英雄退回類別通稱——而通稱**不可以**是某個真技能的譯名。
//
// 舊版艦艇軍官的通稱是「指揮官」,那正是 SKILL_COMMANDO 的譯名,於是
// `commandoLeaderTier` 掃字串時把每一位艦艇軍官都算成 Commando。
func TestGenericLabelDoesNotCollideWithARealSkillName(t *testing.T) {
	for _, tc := range []struct {
		typ  uint8
		want string
	}{
		{gamedata.LeaderTypeCaptain, "艦長"},
		{gamedata.LeaderTypeAdmin, "行政官"},
	} {
		h := herodata.Leader{Name: "無技能", Type: tc.typ}
		got := mercSkillLabel(&sceneBuilder{lang: i18n.Traditional}, h, mercSkills(h))
		if got != tc.want {
			t.Errorf("類型 %d 的通稱應是 %q,得到 %q", tc.typ, tc.want, got)
		}
		if id, ok := gamedata.LeaderSkillIDByZH(got); ok {
			t.Errorf("通稱 %q 撞到真技能 %#x 的譯名——效果判斷會把所有這類領袖都算進去",
				got, id)
		}
	}
}
