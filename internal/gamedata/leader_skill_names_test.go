package gamedata

import "testing"

// 27 個技能一個不多一個不少——enum 是原版的,表是給人看的,兩邊對不上就是漏譯或憑空多譯。
func TestEveryLeaderSkillHasAName(t *testing.T) {
	all := []LeaderSkills{
		SKILL_ASSASSIN, SKILL_COMMANDO, SKILL_DIPLOMAT, SKILL_FAMOUS, SKILL_MEGAWEALTH,
		SKILL_OPERATIONS, SKILL_RESEARCHER, SKILL_SPYMASTER, SKILL_TELEPATH, SKILL_TRADER,
		SKILL_ENGINEER, SKILL_FIGHTER_PILOT, SKILL_GALACTIC_LORE, SKILL_HELMSMAN,
		SKILL_NAVIGATOR, SKILL_ORDNANCE, SKILL_SECURITY, SKILL_WEAPONRY,
		SKILL_ENVIRONMENTALIST, SKILL_FARMING_LEADER, SKILL_FINANCIAL_LEADER,
		SKILL_INSTRUCTOR, SKILL_LABOR_LEADER, SKILL_MEDICINE, SKILL_SCIENCE_LEADER,
		SKILL_SPIRITUAL_LEADER, SKILL_TACTICS,
	}
	if len(all) != len(leaderSkillNames) {
		t.Fatalf("enum 有 %d 個技能,名字表有 %d 個", len(all), len(leaderSkillNames))
	}
	seenZH := map[string]LeaderSkills{}
	for _, id := range all {
		n, ok := LeaderSkillName(int(id))
		if !ok {
			t.Errorf("技能 %#x 沒有名字", int(id))
			continue
		}
		if n.ZH == "" || n.EN == "" {
			t.Errorf("技能 %#x 名字不完整:%+v", int(id), n)
		}
		if prev, dup := seenZH[n.ZH]; dup {
			t.Errorf("中文名 %q 同時給了 %#x 與 %#x——標籤重複就沒辦法當識別鍵",
				n.ZH, int(prev), int(id))
		}
		seenZH[n.ZH] = id
	}
}

// 「指揮官」是 SKILL_COMMANDO 的譯名,不是「艦艇軍官」的通稱。
//
// 這一條釘的是一個真的踩過的坑:HERODATA 的艦艇軍官一律被標成「指揮官」,
// 而 commandoLeaderTier 就是掃這個字串——於是**每一位**雇來的艦艇軍官都拿到了
// Commando 的地面戰加成。標籤與技能必須是一對一的。
func TestCommandoLabelBelongsToCommandoOnly(t *testing.T) {
	id, ok := LeaderSkillIDByZH("指揮官")
	if !ok {
		t.Fatal("「指揮官」查不到 id")
	}
	if id != int(SKILL_COMMANDO) {
		t.Errorf("「指揮官」應是 SKILL_COMMANDO(%#x),得到 %#x", int(SKILL_COMMANDO), id)
	}
}

// 中文標籤反查是舊路徑的相容層,round-trip 要對得回去。
func TestLeaderSkillIDByZHRoundTrips(t *testing.T) {
	for id, n := range leaderSkillNames {
		got, ok := LeaderSkillIDByZH(n.ZH)
		if !ok || got != int(id) {
			t.Errorf("%q 反查得到 (%#x,%v),預期 %#x", n.ZH, got, ok, int(id))
		}
	}
	if _, ok := LeaderSkillIDByZH("不存在的技能"); ok {
		t.Error("查不到的標籤應回 false")
	}
}

// 列舉順序:專屬技能在前、通用技能在後(openorion2 LeaderSkillsWidget::update)。
func TestLeaderSkillIDsForPutsSpecialSkillsFirst(t *testing.T) {
	cap := LeaderSkillIDsFor(LeaderTypeCaptain)
	if len(cap) != 8+10 {
		t.Fatalf("艦艇軍官應有 8 專屬 + 10 通用 = 18 個候選,得到 %d", len(cap))
	}
	if cap[0] != int(SKILL_ENGINEER) {
		t.Errorf("艦艇軍官第一個候選應是 SKILL_ENGINEER(%#x),得到 %#x", int(SKILL_ENGINEER), cap[0])
	}
	if cap[8] != int(SKILL_ASSASSIN) {
		t.Errorf("第 9 個(通用段開頭)應是 SKILL_ASSASSIN(0),得到 %#x", cap[8])
	}

	adm := LeaderSkillIDsFor(LeaderTypeAdmin)
	if len(adm) != 9+10 {
		t.Fatalf("殖民地領袖應有 9 專屬 + 10 通用 = 19 個候選,得到 %d", len(adm))
	}
	if adm[0] != int(SKILL_ENVIRONMENTALIST) {
		t.Errorf("殖民地領袖第一個候選應是 SKILL_ENVIRONMENTALIST(%#x),得到 %#x",
			int(SKILL_ENVIRONMENTALIST), adm[0])
	}
	// 兩種領袖的專屬段不可以互相出現——admin 技能不會長在艦艇軍官身上。
	for _, id := range cap {
		if id&0x30 == 0x20 {
			t.Errorf("艦艇軍官的候選裡出現 admin 技能 %#x", id)
		}
	}
	for _, id := range adm {
		if id&0x30 == 0x10 {
			t.Errorf("殖民地領袖的候選裡出現 captain 技能 %#x", id)
		}
	}
}

// 每個候選 id 都查得到名字——列舉出來卻沒名字就等於畫面上是空白。
func TestEveryEnumeratedSkillIsNamed(t *testing.T) {
	for _, lt := range []int{LeaderTypeCaptain, LeaderTypeAdmin} {
		for _, id := range LeaderSkillIDsFor(lt) {
			if _, ok := LeaderSkillName(id); !ok {
				t.Errorf("leaderType %d 的候選 %#x 沒有名字", lt, id)
			}
		}
	}
}
