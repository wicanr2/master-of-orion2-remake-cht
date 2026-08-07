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
