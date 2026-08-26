package gamedata

import (
	_ "embed"
	"encoding/json"
)

// leader_skill_names.go:27 個領袖技能的**名字**與**列舉順序**。
//
// ============ 為什麼需要這張表 ============
//
// remake 一直用**中文標籤字串**當技能的識別鍵(`shell.leaderSkillIDByName`、
// `commandoLeaderTier` 比對 `l.Skill == "指揮官"`、`FleetHasNavigator` 比對 "領航員")。
// 那條路有兩個洞:
//
//  1. 標籤是**翻譯過的**——英文模式下 `Skill` 存的是 "Scientist",查表查不到,
//     **所有領袖加成當場全部失效**,而畫面上看不出任何異狀。
//  2. 同一個標籤被兩種語意共用——「指揮官」既是 SKILL_COMMANDO 的譯名,
//     又是 HERODATA 艦艇軍官的**通稱**,於是每一位雇來的艦艇軍官都拿到了
//     Commando 的地面戰加成。
//
// 這張表把「技能是什麼」從「畫面上怎麼寫」分開:id 是鍵,名字只是顯示。
//
// ============ 名字來源 ============
//
// 英文名逐字取自 GAME_MANUAL.pdf p.135-137 的三段技能表
// (General Abilities / Command Abilities / Administration Abilities);
// id 取自 openorion2 `gamestate.h` 的 `LeaderSkills` enum(已在 enums.go 生成)。
// 中文名沿用 remake 既有譯法,新增的照同一套命名慣例。
//
// ⚠ 手冊 General Abilities 那段還列了一條 **Tech Knowledge**
// (「Reveals the secret of at least one unknown technology when hired」),
// 但它**不在 LeaderSkills enum 裡**——原版把它存成 HERODATA 記錄的 `techs[3]` 欄位
// (見 internal/herodata/herodata.go 的 parseRecord),是一份具名科技清單而不是技能位元。
// 不要因為手冊把它排在一起就往這張表塞一個不存在的 id。

// LeaderSkillNames 是一項技能的顯示名(中/英)。
type LeaderSkillNames struct {
	Key string `json:"key"`
	ZH  string `json:"zh"`
	EN  string `json:"en"`
}

type leaderSkillTextEntry struct {
	ID int `json:"id"`
	LeaderSkillNames
}

//go:embed leader_skills.json
var leaderSkillTextJSON []byte

// leaderSkillNames 的玩家可見內容由 leader_skills.json 載入；Go 只保留原版數字 id。
// JSON 是隨程式嵌入的資料檔，因此封包不需要額外散落一份 runtime 相依檔。
var leaderSkillNames = loadLeaderSkillNames()

func loadLeaderSkillNames() map[LeaderSkills]LeaderSkillNames {
	var entries []leaderSkillTextEntry
	if err := json.Unmarshal(leaderSkillTextJSON, &entries); err != nil {
		panic("invalid embedded leader_skills.json: " + err.Error())
	}
	out := make(map[LeaderSkills]LeaderSkillNames, len(entries))
	keys := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		id := LeaderSkills(entry.ID)
		if entry.Key == "" || entry.ZH == "" || entry.EN == "" {
			panic("leader_skills.json contains an incomplete entry")
		}
		if _, duplicate := out[id]; duplicate {
			panic("leader_skills.json contains a duplicate id")
		}
		if _, duplicate := keys[entry.Key]; duplicate {
			panic("leader_skills.json contains a duplicate key")
		}
		out[id] = entry.LeaderSkillNames
		keys[entry.Key] = struct{}{}
	}
	return out
}

// LeaderSkillName 依技能 id 回傳顯示名。id 不在表上時回 (零值, false)。
func LeaderSkillName(skillID int) (LeaderSkillNames, bool) {
	n, ok := leaderSkillNames[LeaderSkills(skillID)]
	return n, ok
}

// LeaderSkillIDByZH 依中文標籤反查技能 id。
//
// 這是**舊路徑的相容層**:demo 領袖與既有測試只有一個中文 `Skill` 標籤,沒有 id。
// 新的程式碼一律直接帶 id(見檔頭)。
func LeaderSkillIDByZH(name string) (int, bool) {
	for id, n := range leaderSkillNames {
		if n.ZH == name {
			return int(id), true
		}
	}
	return 0, false
}

// LeaderSkillIDsFor 回傳某種領袖**可能擁有**的技能 id,順序照原版顯示欄的列舉順序。
//
// openorion2 `LeaderSkillsWidget::update`(officer.cpp:134-155)先跑專屬技能
// (艦艇軍官跑 CAPTAIN、殖民地領袖跑 ADMIN),再跑通用技能——所以專屬技能排在前面。
// remake 的 `Leader.Skill` 只放得下一個顯示標籤,挑「第一個」時照這個順序,
// 挑出來的就是原版那一欄最上面那個。
//
// leaderType 用 LeaderTypeCaptain / LeaderTypeAdmin。
func LeaderSkillIDsFor(leaderType int) []int {
	var base, count int
	if leaderType == LeaderTypeCaptain {
		base, count = 0x10, len(baseSkillValues[1])
	} else {
		base, count = 0x20, len(baseSkillValues[2])
	}
	out := make([]int, 0, count+len(baseSkillValues[0]))
	for i := 0; i < count; i++ {
		out = append(out, base+i)
	}
	for i := 0; i < len(baseSkillValues[0]); i++ {
		out = append(out, i)
	}
	return out
}
