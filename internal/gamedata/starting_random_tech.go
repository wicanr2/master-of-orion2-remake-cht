package gamedata

// starting_random_tech.go:**先進級開局的 19 個隨機主題**——照抄挑選的**結構**,不照抄權重。
//
// ============ 這個缺口從第 80 項就開著 ============
//
// `Init_Player_Tech_` @ 0x5E55F 的主迴圈跑 `var_18` 次(1 / 6 / 25):
// 前 6 次取固定表 `word_18111C`,**第 7 次起改由 `sub_FD335` 隨機挑**。
// 所以先進級是「六個固定 + 十九個隨機」——remake 先前只發那六個,少了 19 個。
//
// ============ 為什麼不照抄 ============
//
// `sub_FD335` 的評分是 `weight × horizon ÷ turns`,而 `weight` 來自
// `sub_FC845`——**985 行**的逐科技估值器,吃成本表、種族旗標、性格、政體與一個
// 估值函式。第 88 與 91 項各判斷過一次「一次讀就照抄風險太高」,這次量到了行數,
// 結論不變。
//
// ============ 但結構是讀得出來的 ============
//
// 剩下那三分之二完全可讀,而且是選擇的骨架:
//
//	var_1C = [player+0xAC]              ; 每回合研究點,0 → 1
//	ecx    = 15                          ; 「幾回合內研究得完」的視野
//	loop:
//	  for 每個候選 tech:
//	    turns = 主題成本 ÷ var_1C         ; 0 → 1
//	    if turns > ecx: 這個候選這輪不算
//	    score = weight × ecx ÷ turns
//	  if 一個都沒中: ecx = ecx × 3 ÷ 2,再來一輪
//	sub_FE96F(score 陣列) = **加權隨機**挑一個
//
// 三件事因此是確定的:
//
//  1. **只從「現在可研究」的主題挑**(`[player+0xC4+topic] == 2`),不是全表亂挑
//  2. **偏好便宜的**:score 與 turns 成反比,而 turns 正比於主題成本
//  3. **視野會放寬**:15 → 22 → 33 …,直到至少有一個候選在視野內
//
// 把 `weight` 一律當 1 之後,`score = ecx ÷ turns`——**選擇仍然由成本主導**,
// 只是失去「這個科技對這個種族/性格有多有用」那一層。這比只發六個接近原版得多,
// 而且沒有任何一個數字是編的。
//
// ============ 誠實留白 ============
//
//   - **`weight` 一律 1。** 少的是 `sub_FC845` 那 985 行的估值。AI 因此不會偏好
//     「對它有用」的科技,只會偏好便宜的。
//   - `sub_FD335` 尾巴還有一段依 `[player+0x28]`(值 1/2/4/5)做的二次過濾,
//     那個欄位**沒查到寫入端、沒有名字**(第 90 項也遇到同一個欄位),**不照抄**。
//   - 每回合研究點在「開局發科技」那一刻其實還沒意義(還沒有回合跑過)。
//     呼叫端傳母星的初始研究產出,那是最接近原版當下狀態的值。

// StartingRandomHorizonInitial 是「幾回合內研究得完」的初始視野(原版 `mov ecx, 0Fh`)。
const StartingRandomHorizonInitial = 15

// startingRandomHorizonGrow 把視野放寬一輪(原版 `lea eax,[ecx+ecx*2] / cdq / sub / sar 1` = ×3÷2)。
func startingRandomHorizonGrow(h int) int {
	n := h * 3 / 2
	if n <= h {
		return h + 1 // 防呆:視野必須真的變大,否則外層迴圈不會結束
	}
	return n
}

// StartingRandomTopicScores 依原版的結構算出每個候選主題的權重。
//
// available 是「現在可研究」的主題(原版的狀態 2);researchPerTurn 是每回合研究點。
// 回傳與 available 等長的分數陣列,分數 0 = 這一輪不列入。
//
// ⚠ 每個候選的基礎 weight 一律 **1**(見檔頭誠實留白)。
func StartingRandomTopicScores(available []ResearchTopic, researchPerTurn int) []int {
	if len(available) == 0 {
		return nil
	}
	if researchPerTurn < 1 {
		researchPerTurn = 1 // 原版:var_1C == 0 → 1
	}
	scores := make([]int, len(available))
	for horizon := StartingRandomHorizonInitial; ; horizon = startingRandomHorizonGrow(horizon) {
		any := false
		for i, t := range available {
			cost := 0
			if int(t) >= 0 && int(t) < len(OrigTopicCost) {
				cost = OrigTopicCost[int(t)]
			}
			turns := cost / researchPerTurn
			if turns < 1 {
				turns = 1 // 原版:算出 0 → 1
			}
			if turns > horizon {
				scores[i] = 0
				continue
			}
			scores[i] = horizon / turns // weight = 1(見檔頭)
			if scores[i] > 0 {
				any = true
			}
		}
		if any {
			return scores
		}
		// 視野放到比最貴的主題還大之後仍然沒有候選,代表 available 全是成本 0 之類的
		// 異常資料——不要無限放寬。
		if horizon > 1<<20 {
			return scores
		}
	}
}

// StartingRandomTopicPick 從 available 裡加權隨機挑一個主題(原版 `sub_FE96F`)。
//
// roll 傳 `rng.Intn`,由呼叫端提供——開局發科技必須是決定性的(見 shell/determinism.go)。
// available 為空時回傳 (0, false)。
func StartingRandomTopicPick(available []ResearchTopic, researchPerTurn int, roll func(n int) int) (ResearchTopic, bool) {
	scores := StartingRandomTopicScores(available, researchPerTurn)
	total := 0
	for _, s := range scores {
		total += s
	}
	if total <= 0 {
		return 0, false
	}
	r := roll(total)
	for i, s := range scores {
		if r < s {
			return available[i], true
		}
		r -= s
	}
	return available[len(available)-1], true
}

// ============ 粒度:原版挑的是「科技應用」,不是「主題」 ============
//
// 2026-08-08(第 108 項)發現的一個比「weight = 1」嚴重得多的落差。
//
// `Choose_Tech_Application_` @ 0xFD335 的迴圈**跑的是 212 個 tech-item**,不是 83 個主題:
//
//	mov  ecx, 0Dh                    ; 13 = tech-item 記錄的 stride
//	lea  eax, [ebp+var_6C0]
//	mov  ebx, 350h                   ; 0x350 = 848 = 212 × 4,逐 tech-item 的分數陣列
//	call memset_
//	...
//	cmp  byte ptr [eax+117h], 1      ; [player + 0x117 + techIdx] —— 逐 **tech-item** 的狀態
//	movsx ebx, word_17E07F[ecx]      ; 這個 tech-item 屬於哪個主題
//	cmp  byte ptr [ebx+eax+0C4h], 2  ; [player + 0xC4 + topic] —— 那個主題可不可以研究
//	movsx edx, di                    ; ★ 傳給估值函式的是 **tech-item 索引**
//	call sub_FC845
//
// 兩個索引空間並存:`+0xC4` 是逐主題、`+0x117` 是逐 tech-item。函式名也直說了
// ——**Choose Tech _Application_**。
//
// 而 remake 的 `applyStartingRandomTech` 先前是 `ps.CompletedTopics[t] = true`
// ——發**整個主題**。一個主題底下有 2–3 個抉擇,完成主題而不做抉擇時
// `componentUnlockedFor` 會把那些**全部**解鎖(見該函式)。
//
// 所以先進級開局在 remake 拿到的東西是原版的兩到三倍。這不是權重誤差,是**粒度錯誤**
// ——而且權重再怎麼校準也修不了它。

// StartingRandomApplicationPick 從一個主題的抉擇裡挑一項(原版 `Choose_Tech_Application_`
// 的粒度)。回傳 (科技, true);`ResearchAll` 主題回 (0, false)。
//
// `ResearchAll` 的主題手冊明說研究完就三項全拿(見 techtree.go),**沒有抉擇可挑**,
// 所以那類主題回 false,呼叫端維持「整個主題都解鎖」的既有行為。
//
// 權重用 `TechCategoryWeight`(原版估值函式階段 B 給 `ecx` 的初始值,是一手值)。
// ⚠ 階段 C–N 的十幾段加成沒有套——那些依賴的玩家欄位語意還沒查出來,
// 見 `orig_tech_value_tables.go` 檔頭。所以這是「照原版的粒度 + 原版的起點權重」,
// **不是**原版估值的完整重現。
//
// roll 傳 rng.Intn(回 [0,n))。
func StartingRandomApplicationPick(topic ResearchTopic, roll func(n int) int) (Technology, bool) {
	i := int(topic)
	if i < 0 || i >= len(researchChoices) {
		return 0, false
	}
	rc := researchChoices[i]
	if rc.ResearchAll || len(rc.Choices) == 0 {
		return 0, false
	}
	if len(rc.Choices) == 1 {
		return rc.Choices[0], true
	}
	weights := make([]int, len(rc.Choices))
	total := 0
	for k, tech := range rc.Choices {
		w := TechCategoryWeight(tech)
		if w < 1 {
			w = 1 // category 查不到時不讓它變成「永遠選不到」
		}
		weights[k] = w
		total += w
	}
	r := roll(total)
	for k, w := range weights {
		if r < w {
			return rc.Choices[k], true
		}
		r -= w
	}
	return rc.Choices[len(rc.Choices)-1], true
}
