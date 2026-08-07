package shell

// audience.go:AI 主動請求會談(原版 `Humans_Requesting_Diplomacy_` @ 0xFA795 +
// `Draw_Diplomacy_Request_Lights_` @ 0x83D06)。
//
// remake 的外交先前**只有玩家主動**:玩家點進外交畫面提議,AI 只會回應。原版不是這樣——
// AI 會來敲門(宣戰、提議和平、提議結盟),星圖上方就亮起一排會談請求燈。
// 那一層畫不出來,卡的是「誰在請求」這個狀態根本不存在。
//
// ============ 一、表示法是真值 ============
//
// `Humans_Requesting_Diplomacy_` 整支只有 `mov al, byte_1AB054; retn` ——
// **一個位元遮罩,每位對手一個 bit**。remake 用 `AIOpponent.WantsAudience` 表達同一件事
// (逐對手一個旗標,語意相同、不必自己維護 bit 位置)。
//
// ============ 二、⚠ 觸發條件不是照抄的 ============
//
// 原版設那個 bit 的地方在 `sub_F5A9F` —— 一支約 30 路跳表的 AI 行動分派函式,
// 觸發散在各個 case 裡。追出完整條件的成本很高而收穫有限,**所以沒照抄**。
//
// remake 改接在既有的 AI 模型上:**態勢改變時來敲門**。理由是 `ai.DecideStance` 的五級裡
// 有三級本身就是「要跟你講話」的語意 —— 宣戰、提議貿易、提議結盟。態勢從 A 變成 B
// 正是原版 AI 會主動聯絡玩家的時機(宣戰要通知、提議要開口)。
//
// **這是接在既有模型上的推導,不是新編的門檻值** —— 沒有引入任何新的數字。

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"

// 會談來意的代碼。
//
// ⚠ 用**代碼**不用中文:規則層不該吐顯示字串,顯示文字(以及英文版)是 UI 的事。
// 既有的 `stanceNames` 是中文,那是先前留下的;新欄位不再擴散這個作法。
const (
	AudienceReasonWar      = "war"      // 宣戰
	AudienceReasonTrade    = "trade"    // 提議貿易
	AudienceReasonAlliance = "alliance" // 提議結盟
)

// audienceReasonForStance 回傳某個態勢是否構成「來敲門」,以及敲門的來意代碼。
//
// 中立/敵視不敲門:前者沒事,後者是態度不是提案 —— 原版的敵視 AI 也不會特地來說「我討厭你」。
func audienceReasonForStance(stance string) (string, bool) {
	switch stance {
	case stanceNames[ai.StanceWar]:
		return AudienceReasonWar, true
	case stanceNames[ai.StanceProposeTrade]:
		return AudienceReasonTrade, true
	case stanceNames[ai.StanceProposeAlliance]:
		return AudienceReasonAlliance, true
	}
	return "", false
}

// noteStanceChange 在 AI 態勢改變時決定要不要請求會談。
//
// prev 是變化前的態勢(空字串 = 開局尚未設定,不觸發 —— 開局第一次算出態勢不算「改變」,
// 否則每個新對局一開始就會有一整排燈亮著)。
func (a *AIOpponent) noteStanceChange(prev string) {
	if prev == "" || prev == a.StanceName {
		return
	}
	if reason, ok := audienceReasonForStance(a.StanceName); ok {
		a.WantsAudience = true
		a.AudienceReason = reason
	}
}

// AudienceRequests 回傳目前正在請求會談的對手索引(依 AIPlayers 順序)。
//
// 原版的燈是**由右往左**排(x = 506 − n×寬),所以先請求的在最右邊;這裡回傳的順序即排列順序,
// 由呼叫端決定方向。
func (s *GameSession) AudienceRequests() []int {
	var out []int
	for i := range s.AIPlayers {
		if s.AIPlayers[i].WantsAudience {
			out = append(out, i)
		}
	}
	return out
}

// ClearAudienceRequest 清掉某位對手的請求(玩家進了外交畫面就算談過了)。
func (s *GameSession) ClearAudienceRequest(idx int) {
	if idx < 0 || idx >= len(s.AIPlayers) {
		return
	}
	s.AIPlayers[idx].WantsAudience = false
	s.AIPlayers[idx].AudienceReason = ""
}

// ClearAudienceRequestByName 依對手名稱清掉請求(外交畫面拿到的是名字不是索引)。
func (s *GameSession) ClearAudienceRequestByName(name string) {
	for i := range s.AIPlayers {
		if s.AIPlayers[i].Name == name {
			s.ClearAudienceRequest(i)
			return
		}
	}
}

// AudienceReasonFor 回傳某位對手的來意(沒有請求回空字串)。
func (s *GameSession) AudienceReasonFor(idx int) string {
	if idx < 0 || idx >= len(s.AIPlayers) || !s.AIPlayers[idx].WantsAudience {
		return ""
	}
	return s.AIPlayers[idx].AudienceReason
}
