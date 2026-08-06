package gamedata

// 原版隨機事件表(36 種)。
//
// 來源與驗證(2026-08-06):
//   - 事件總數與好壞旗標:反組譯 `_event_good_array` @ 0x180E84(36 bytes,1=好事)
//   - 事件內容:原版 EVENTMSG.LBX 從資產 8 起、**每個事件 4 條訊息變體**,
//     (152-8)/4 = 36,與 `_event_good_array` 的長度完全吻合
//   - 逐項交叉檢查:礦產發現=好、礦產枯竭=壞、人口暴增=好、瘟疫=壞、
//     蟲洞=好、太空怪獸=壞……36 項全部自洽,兩個獨立來源互證
//   - 事件子系統的函式清單見原版 module 15(71 個符號,含 Event_Check_Plague_ /
//     Event_Check_Comet_ / Get_Event_Mineral_Depletion_Colony_ 等)
//
// 原版事件以 GNN(Galactic News Network)新聞快報的形式播報,每則事件有 4 條訊息
// 對應不同階段或不同說法(如彗星:預警 → 部分攔截 → 完全攔截 → 撞擊)。
//
// ⚠ remake 目前只實作得了其中一部分——有些事件需要 remake 還沒有的子系統
// (太空怪獸、超新星、時空異象、曲速漏斗)。`Implemented` 欄誠實標記哪些已接,
// 沒接的**照樣列在表裡**:這張表是權威記錄,不是「remake 現況」的鏡子。

// RandomEvent 是一種原版隨機事件。
type RandomEvent struct {
	ID   int    // 原版事件 id(0..35),同 _event_good_array 的索引
	Name string // 中文事件名(remake 自訂,原版無事件名字串,只有訊息)
	Good bool   // 原版 _event_good_array:是否為好事
	// MsgBase 是該事件在 EVENTMSG.LBX 的第一條訊息資產索引;該事件佔用
	// MsgBase..MsgBase+3 四個索引(部分為空字串,代表原版沒用到那個變體)。
	MsgBase int
	// Implemented 標記 remake 是否已能忠實觸發並結算這個事件。
	// false = 資料已備、機制待建(見檔頭說明),不是漏列。
	Implemented bool
	// Needs 說明 false 時缺什麼子系統(供 worklist 排序);已實作的留空。
	Needs string
	// Broadcast 標記這是「狀態播報」而非隨機抽取的事件——原版 GNN 也播帝國滅亡、
	// 擊敗安塔蘭、排行榜這類新聞,但它們由遊戲狀態觸發,不進隨機事件池。
	Broadcast bool
}

// RandomEvents 是原版 36 種事件的完整表。
var RandomEvents = []RandomEvent{
	{0, "古代遺骸科技", true, 8, true, "", false},
	{1, "氣候改善", true, 12, true, "", false},
	{2, "彗星來襲", false, 16, false, "彗星倒數與艦隊攔截判定", false},
	{3, "電腦病毒", false, 20, true, "", false},
	{4, "外交暗殺", false, 24, true, "", false},
	{5, "外交聯姻", true, 28, true, "", false},
	{6, "富商捐獻", true, 32, true, "", false},
	{7, "地震", false, 36, true, "", false},
	{8, "艦船爆炸", false, 40, true, "", false},
	{9, "超空間亂流", false, 44, false, "全銀河禁止航行的持續狀態", false},
	{10, "工業意外", false, 48, true, "", false},
	{11, "礦產枯竭", false, 52, true, "", false},
	{12, "礦產發現", true, 56, true, "", false},
	{13, "艦船叛變", false, 60, true, "", false},
	{14, "海盜活動", false, 64, false, "星系層級的持續海盜狀態與清剿", false},
	{15, "海盜劫掠", false, 68, true, "", false},
	{16, "瘟疫", false, 72, true, "", false},
	{17, "人口暴增", true, 76, true, "", false},
	{18, "秘密實驗", true, 80, true, "", false},
	{19, "太空變形蟲", false, 84, false, "太空怪獸實體與戰鬥", false},
	{20, "太空水晶", false, 88, false, "太空怪獸實體與戰鬥", false},
	{21, "太空巨龍", false, 92, false, "太空怪獸實體與戰鬥", false},
	{22, "太空鰻", false, 96, false, "太空怪獸實體與戰鬥", false},
	{23, "太空九頭蛇", false, 100, false, "太空怪獸實體與戰鬥", false},
	{24, "超新星", false, 104, false, "恆星狀態變化與系統毀滅", false},
	{25, "時空異象", false, 108, false, "星系凍結(禁建造/禁接觸)的持續狀態", false},
	{26, "超空間獸", false, 112, false, "航行中隨機損失艦船的持續狀態", false},
	{27, "曲速漏斗", false, 116, false, "艦隊受困與脫困判定", false},
	{28, "蟲洞", true, 120, false, "艦隊瞬間移動", false},
	{29, "帝國滅亡", true, 124, true, "", true},
	{30, "帝國壯大", true, 128, true, "", true},
	{31, "排行榜播報", true, 132, false, "GNN 定期排名播報(非隨機事件)", true},
	{32, "發現獵戶座", true, 136, false, "獵戶座星系與守護者", true},
	{33, "擊敗安塔蘭", true, 140, true, "", true},
	{34, "帝國投降", true, 144, true, "", true},
	{35, "叛軍同化", true, 148, false, "叛亂殖民地機制", true},
}

// RandomEventByID 依原版事件 id 取事件定義;id 越界回 nil。
func RandomEventByID(id int) *RandomEvent {
	for i := range RandomEvents {
		if RandomEvents[i].ID == id {
			return &RandomEvents[i]
		}
	}
	return nil
}

// ImplementedRandomEvents 回傳可進隨機事件池的事件:remake 已實作、且不是狀態播報類。
// 隨機事件的抽樣只從這裡挑,不會挑到還沒有機制的事件,也不會把「帝國滅亡」這類
// 由遊戲狀態決定的新聞當成隨機事件抽出來。
func ImplementedRandomEvents() []RandomEvent {
	out := make([]RandomEvent, 0, len(RandomEvents))
	for _, e := range RandomEvents {
		if e.Implemented && !e.Broadcast {
			out = append(out, e)
		}
	}
	return out
}

// EventMessageAssetIDs 回傳事件在 EVENTMSG.LBX 的四個訊息資產索引。
// 供未來直接讀原版訊息(中文化時對照原文用);remake 目前用自己的中文文案。
func EventMessageAssetIDs(id int) [4]int {
	e := RandomEventByID(id)
	if e == nil {
		return [4]int{}
	}
	return [4]int{e.MsgBase, e.MsgBase + 1, e.MsgBase + 2, e.MsgBase + 3}
}
