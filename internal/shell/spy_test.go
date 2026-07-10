package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// clonePlayerState 深拷貝 engine.PlayerState 的三個 map 欄位(其餘為值型別,結構賦值已足夠),
// 供間諜測試在多次試驗(不同 rng 種子)間重跑同一個起始狀態,不被上一次試驗的 applyTechTheft
// 汙染。
func clonePlayerState(ps engine.PlayerState) engine.PlayerState {
	out := ps
	if ps.CompletedTopics != nil {
		out.CompletedTopics = make(map[gamedata.ResearchTopic]bool, len(ps.CompletedTopics))
		for k, v := range ps.CompletedTopics {
			out.CompletedTopics[k] = v
		}
	}
	if ps.ChosenTech != nil {
		out.ChosenTech = make(map[gamedata.ResearchTopic]gamedata.Technology, len(ps.ChosenTech))
		for k, v := range ps.ChosenTech {
			out.ChosenTech[k] = v
		}
	}
	if ps.ExplicitChoice != nil {
		out.ExplicitChoice = make(map[gamedata.ResearchTopic]bool, len(ps.ExplicitChoice))
		for k, v := range ps.ExplicitChoice {
			out.ExplicitChoice[k] = v
		}
	}
	return out
}

// --- spyStealOptions / psKnowsTech / applyTechTheft(純函式,不涉及 rng) ---

func TestSpyStealOptions_FindsTechDefenderHasAttackerLacks(t *testing.T) {
	attacker := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true},
	}
	defender := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_STARTING_TECH:         true,
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: true,
		},
		ChosenTech: map[gamedata.ResearchTopic]gamedata.Technology{
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: gamedata.TECH_AUTOMATED_FACTORIES,
		},
	}
	opts := spyStealOptions(attacker, defender)
	if len(opts) != 1 {
		t.Fatalf("spyStealOptions 長度 = %d,預期 1(僅 TOPIC_ADVANCED_CONSTRUCTION 一項可偷)", len(opts))
	}
	if opts[0].Topic != gamedata.TOPIC_ADVANCED_CONSTRUCTION || opts[0].Tech != gamedata.TECH_AUTOMATED_FACTORIES {
		t.Errorf("opts[0] = %+v,預期 {TOPIC_ADVANCED_CONSTRUCTION TECH_AUTOMATED_FACTORIES}", opts[0])
	}
}

// TestSpyStealOptions_EmptyWhenNoAdvantage 驗證雙方已知科技相同時(defender 沒有 attacker
// 不知道的科技),無可偷項目——對應「沒有可偷的就無效」的硬邊界。
func TestSpyStealOptions_EmptyWhenNoAdvantage(t *testing.T) {
	shared := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_STARTING_TECH:         true,
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: true,
		},
		ChosenTech: map[gamedata.ResearchTopic]gamedata.Technology{
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: gamedata.TECH_AUTOMATED_FACTORIES,
		},
	}
	if opts := spyStealOptions(shared, shared); len(opts) != 0 {
		t.Errorf("雙方科技相同時 spyStealOptions 長度 = %d,預期 0", len(opts))
	}
}

// TestSpyStealOptions_NilDefenderTopics 驗證 defender.CompletedTopics 為 nil(尚未研究任何
// 東西)時安全回傳空,不 panic。
func TestSpyStealOptions_NilDefenderTopics(t *testing.T) {
	attacker := engine.PlayerState{}
	defender := engine.PlayerState{}
	if opts := spyStealOptions(attacker, defender); opts != nil {
		t.Errorf("nil CompletedTopics 應回傳 nil/空切片,got %+v", opts)
	}
}

func TestApplyTechTheft_OnlyGrantsStolenTechNotSiblings(t *testing.T) {
	attacker := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true},
	}
	applyTechTheft(&attacker, spyStealOption{
		Topic: gamedata.TOPIC_ADVANCED_CONSTRUCTION, Tech: gamedata.TECH_AUTOMATED_FACTORIES,
	})
	if !attacker.CompletedTopics[gamedata.TOPIC_ADVANCED_CONSTRUCTION] {
		t.Fatal("偷竊後 CompletedTopics 應標記該主題已完成")
	}
	if !psKnowsTech(attacker, gamedata.TOPIC_ADVANCED_CONSTRUCTION, gamedata.TECH_AUTOMATED_FACTORIES) {
		t.Error("偷到的科技應通過 psKnowsTech 判定")
	}
	// TOPIC_ADVANCED_CONSTRUCTION 的其餘選項(TECH_HEAVY_ARMOR)不應因為偷了同主題的另一項
	// 就跟著解鎖——偷竊比照「明確抉擇」語意,只解鎖偷到的那一項(componentUnlockedFor 規則)。
	if psKnowsTech(attacker, gamedata.TOPIC_ADVANCED_CONSTRUCTION, gamedata.TECH_HEAVY_ARMOR) {
		t.Error("偷竊不應連帶解鎖同主題的其餘選項(TECH_HEAVY_ARMOR)")
	}
}

// --- resolveSpyVsSpy(SpyVsSpy 對抗判定) ---

func TestResolveSpyVsSpy(t *testing.T) {
	cases := []struct {
		name             string
		ab, db           int
		hide             bool
		wantAttackerDead bool
		wantDefenderDead bool
	}{
		// 基準情況(ab=0,db=0):defenderB=SpyVsSpyDefenderBonus(0)=20,attackerB=0,
		// net=-20,兩門檻(+80/-80)都不到,雙方均安全。
		{"baseline_no_kill", 0, 0, false, false, false},
		// attacker 遠強(ab=200):attackerB=200,defenderB=20,net=180>=80,defender 被擊殺。
		{"attacker_strong_kills_defender", 200, 0, false, false, true},
		// defender 遠強(db=200):attackerB=0,defenderB=220,net=-220<=-80,attacker 被擊殺。
		{"defender_strong_kills_attacker", 0, 200, false, true, false},
		// HIDE 指令給 attacker +20(手冊:「the attacker gets +20 if he has chosen HIDE」),
		// ab=65 平常不夠(65-20=45,未達 80),但 hide 後 attackerB=85,net=65,仍未達 80——
		// 換成 ab=65 且提高 db 差距不變,驗證 hide 確實有加成但門檻仍需自行跨過。
		{"hide_bonus_not_enough_alone", 65, 0, true, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := resolveSpyVsSpy(c.ab, c.db, c.hide)
			if out.AttackerKilled != c.wantAttackerDead {
				t.Errorf("AttackerKilled = %v,預期 %v", out.AttackerKilled, c.wantAttackerDead)
			}
			if out.DefenderKilled != c.wantDefenderDead {
				t.Errorf("DefenderKilled = %v,預期 %v", out.DefenderKilled, c.wantDefenderDead)
			}
		})
	}
}

// TestResolveSpyVsSpy_HideAddsBonus 直接驗證 HIDE 指令讓 attacker 淨值多 20(用剛好卡在門檻
// 兩側的 ab 值,只有加了 HIDE 才會跨過 +80 門檻擊殺 defender)。
func TestResolveSpyVsSpy_HideAddsBonus(t *testing.T) {
	// ab=79,db=0:attackerB=79,defenderB=20,net=59,不足 80。
	if out := resolveSpyVsSpy(79, 0, false); out.DefenderKilled {
		t.Fatal("不 HIDE 時 net=59 不應達到擊殺門檻")
	}
	// 同樣 ab=79 但 HIDE:attackerB=99,defenderB=20,net=79,仍不足 80(驗證邊界未跨過)。
	if out := resolveSpyVsSpy(79, 0, true); out.DefenderKilled {
		t.Fatal("net=79 仍不足 80,不應擊殺")
	}
	// ab=80 + HIDE:attackerB=100,defenderB=20,net=80,達到門檻,defender 被擊殺。
	if out := resolveSpyVsSpy(80, 0, true); !out.DefenderKilled {
		t.Fatal("net=80(HIDE +20 加成後)應達到擊殺門檻,defender 應被擊殺")
	}
}

// --- 效果門檻公式組成:防禦方 bonus 應降低成功率 ---

// TestDefenderBonusLowersSuccessChance 驗證 SpyEffectiveThreshold + SpyRollChance 這組
// spyStealAttempt 實際使用的公式組合,在 defender bonus(db)升高時成功機率 p 會下降。
//
// 現行 remake 的 spyDefenderBonus() 因為沒有 Agent 追蹤模型而固定回 0(見該函式註解、
// TODO)——這裡直接餵公式不同的 db 值,驗證「防禦方 bonus 降低成功率」這條手冊規則在公式
// 組成上是正確的,以便將來接上 Agent 訓練系統(spyDefenderBonus 改吃真實 agent 數)時,
// 這條數學關係已經先被驗證過。
func TestDefenderBonusLowersSuccessChance(t *testing.T) {
	const ab = 15 // 固定攻擊方加成(比照 8 名間諜:SpySlotBonus(8)=15)
	pNoDefense := gamedata.SpyRollChance(gamedata.SpyEffectiveThreshold(gamedata.SpyThresholdSteal, 0, ab))
	pWithDefense := gamedata.SpyRollChance(gamedata.SpyEffectiveThreshold(gamedata.SpyThresholdSteal, 30, ab))
	if !(pWithDefense < pNoDefense) {
		t.Errorf("db=30 的成功率(%v)應低於 db=0(%v)", pWithDefense, pNoDefense)
	}
}

// --- TrainSpy ---

func TestTrainSpy_DeductsBCAndIncrementsCount(t *testing.T) {
	s := NewDemoSession()
	before := s.Player.BC
	if ok := s.TrainSpy(0); !ok {
		t.Fatal("TrainSpy(0) 應成功(BC 充足、目標索引合法)")
	}
	if s.Player.BC != before-spyTrainCostBC {
		t.Errorf("訓練後 BC = %d,預期 %d", s.Player.BC, before-spyTrainCostBC)
	}
	if len(s.PlayerSpies) == 0 || s.PlayerSpies[0] != 1 {
		t.Errorf("PlayerSpies[0] = %v,預期 1", s.PlayerSpies)
	}
}

func TestTrainSpy_InsufficientBCFails(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = spyTrainCostBC - 1
	if ok := s.TrainSpy(0); ok {
		t.Fatal("BC 不足時 TrainSpy 應回傳 false")
	}
	if len(s.PlayerSpies) != 0 && s.PlayerSpies[0] != 0 {
		t.Errorf("BC 不足時不應增加間諜數,got %v", s.PlayerSpies)
	}
}

func TestTrainSpy_InvalidTargetIndex(t *testing.T) {
	s := NewDemoSession()
	before := s.Player.BC
	if ok := s.TrainSpy(len(s.AIPlayers)); ok {
		t.Fatal("越界的 targetIdx 應回傳 false")
	}
	if s.Player.BC != before {
		t.Errorf("越界目標不應扣款,BC = %d,預期 %d", s.Player.BC, before)
	}
}

// --- spyStealAttempt(含 rng 的整合行為;固定種子驗證) ---

// TestSpyStealAttempt_SuccessAppliesTheft 用固定 rng 種子搜尋一個「本回合擲骰成功」的種子
// (公式本身在 gamedata/spy_test.go 已驗證,這裡驗證 shell 層把成功結果正確套用到
// attacker PlayerState),斷言:成功時訊息非空、偷到的科技確實寫入 attacker 的
// CompletedTopics/ChosenTech。種子搜尋上限 2000,找不到視為公式/機率配置有誤而非種子不巧。
func TestSpyStealAttempt_SuccessAppliesTheft(t *testing.T) {
	attackerBase := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true},
	}
	defender := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{
			gamedata.TOPIC_STARTING_TECH:         true,
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: true,
		},
		ChosenTech: map[gamedata.ResearchTopic]gamedata.Technology{
			gamedata.TOPIC_ADVANCED_CONSTRUCTION: gamedata.TECH_AUTOMATED_FACTORIES,
		},
	}
	const spies = 10 // SpySlotBonus(10)=15,E=80-15=65,p≈(99-65)*(98-65)/2/9900+0.01≈0.067(~6.7%)

	found := false
	for seed := int64(1); seed <= 2000 && !found; seed++ {
		attacker := clonePlayerState(attackerBase)
		rng := rand.New(rand.NewSource(seed))
		msgs, _ := spyStealAttempt(rng, &attacker, defender, spies, "我方", "AI")
		if len(msgs) == 0 {
			continue
		}
		if !attacker.CompletedTopics[gamedata.TOPIC_ADVANCED_CONSTRUCTION] {
			continue // 這則訊息是 SpyVsSpy 擊殺訊息,非偷竊成功(理論上此設定下不會發生,防禦性檢查)
		}
		found = true
		if attacker.ChosenTech[gamedata.TOPIC_ADVANCED_CONSTRUCTION] != gamedata.TECH_AUTOMATED_FACTORIES {
			t.Errorf("偷到的科技 = %v,預期 TECH_AUTOMATED_FACTORIES",
				attacker.ChosenTech[gamedata.TOPIC_ADVANCED_CONSTRUCTION])
		}
		if !attacker.ExplicitChoice[gamedata.TOPIC_ADVANCED_CONSTRUCTION] {
			t.Error("偷竊應標記 ExplicitChoice=true(只解鎖偷到的那一項)")
		}
	}
	if !found {
		t.Fatal("2000 個種子內都沒能重現一次諜報成功,可能是機率公式/wiring 有誤")
	}
}

// TestSpyStealAttempt_NoOptionsIsHarmless 驗證「攻守雙方已知科技相同、無可偷項目」時,即使
// 擲骰判定為成功,也只會記一則「得手但無可偷」訊息,不 panic、不誤改任何科技狀態。
func TestSpyStealAttempt_NoOptionsIsHarmless(t *testing.T) {
	shared := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_STARTING_TECH: true},
	}
	const spies = 63 // 拉到滿(SpySlotBonus(63)=41)提高成功機率,增加測到「成功但無可偷」分支的機會
	for seed := int64(1); seed <= 200; seed++ {
		attacker := clonePlayerState(shared)
		defender := clonePlayerState(shared)
		rng := rand.New(rand.NewSource(seed))
		spyStealAttempt(rng, &attacker, defender, spies, "我方", "AI")
		// 無論訊息是否非空,attacker 的已知科技集不應改變(shared 只有 TOPIC_STARTING_TECH,
		// 沒有可偷項目)。
		if len(attacker.CompletedTopics) != len(shared.CompletedTopics) {
			t.Fatalf("seed=%d:攻守雙方科技相同時,attacker 的 CompletedTopics 不應被改動,got %+v",
				seed, attacker.CompletedTopics)
		}
	}
}

// --- advanceEspionage(GameSession 整合:訓練 → 結算 → 維護費) ---

// TestAdvanceEspionage_Isolated 直接呼叫 advanceEspionage(不跑完整 EndTurn 經濟結算),隔離
// 驗證維護費扣款金額精確等於 spyCount * spyMaintenancePerSpyBC。
func TestAdvanceEspionage_Isolated(t *testing.T) {
	s := NewDemoSession()
	s.PlayerSpies = []int{2}
	before := s.Player.BC
	s.advanceEspionage()
	wantMaintenance := 2 * spyMaintenancePerSpyBC
	if s.Player.BC != before-wantMaintenance {
		t.Errorf("扣款後 BC = %d,預期 %d(維護費 %d)", s.Player.BC, before-wantMaintenance, wantMaintenance)
	}
}

// TestAdvanceEspionage_ZeroSpiesNoOp 驗證預設 0 間諜時 advanceEspionage 完全不影響 BC/科技
// 狀態、LastEspionage 維持空——維持既有(未使用間諜系統的)對局行為不變。
func TestAdvanceEspionage_ZeroSpiesNoOp(t *testing.T) {
	s := NewDemoSession()
	before := s.Player.BC
	s.advanceEspionage()
	if s.Player.BC != before {
		t.Errorf("0 間諜時 BC 不應變動,got %d,預期 %d", s.Player.BC, before)
	}
	if len(s.LastEspionage) != 0 {
		t.Errorf("0 間諜時 LastEspionage 應為空,got %v", s.LastEspionage)
	}
}

// TestAdvanceEspionage_SpyCountNeverNegative 驗證正常遊戲流程下(spyAttackerBonus 上限 41,
// 到不了 SpyVsSpy ±80 擊殺門檻,見 resolveSpyVsSpy 檔頭說明)多回合結算後 PlayerSpies 不會
// 變成負數;真正的擊殺門檻邏輯已在 TestResolveSpyVsSpy 用構造值覆蓋。
func TestAdvanceEspionage_SpyCountNeverNegative(t *testing.T) {
	s := NewDemoSession()
	s.PlayerSpies = []int{1}
	for i := 0; i < 50; i++ {
		s.advanceEspionage()
		if s.PlayerSpies[0] < 0 {
			t.Fatalf("第 %d 輪後 PlayerSpies[0] = %d,不應為負", i, s.PlayerSpies[0])
		}
	}
}
