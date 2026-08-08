package gamedata

// starting_tech.go:**開局送的研究主題**——依 NEW GAME 的 TECH LEVEL 決定送幾個、送哪些。
//
// ============ 這張表先前是拿不到的 ============
//
// `shell.TechLevels` 的註解一直寫著:
//
//	開局已知科技領域數:Pre-warp 2 個、其餘 6 個。要補得先查出那 6 個是哪些
//	(手冊只說預設第一個是 field #29),沒有一手表之前不臆造。
//
// 2026-08-07 從 `Init_Player_Tech_` @ 0x5E55F 挖到了那張表。
//
// **送幾個**由 `byte_199CB5`(NEW GAME 的 TECH LEVEL,來自 `word_1A1360`)決定:
//
//	al = byte_199CB5
//	al == 0 → var_18 = 1
//	al == 1 → var_18 = 6
//	al == 2 → var_18 = 0x19 = 25
//
// **送哪些**在 `word_18111C` @ 0x18111C,六個 word:
//
//	dw 1Dh, 37h, 16h, 39h, 1Ch, 17h   →  29, 55, 22, 57, 28, 23
//
// 主迴圈 `esi` 從 0 跑到 `var_18`:前 6 次取這張固定表,第 7 次起改由 `sub_FD335` 隨機挑
// ——所以 25 級是「六個固定 + 十九個隨機」,不是二十五個固定。
//
// ============ 三方互證 ============
//
// | 來源 | 說法 |
// |---|---|
// | 手冊 | 「預設的第一個是 field #29」(remake 先前就記下了這句) |
// | 反組譯 | `word_18111C[0] = 0x1D = 29` |
// | remake 的主題編號 | 29 = `TOPIC_ENGINEERING` |
//
// 第二個 55 = `TOPIC_NUCLEAR_FISSION` 同樣對得上:手冊說 Average「已具備星際航行所需的
// 全部科技」,而 remake 早就把 `FTLTopic` 定成 `TOPIC_NUCLEAR_FISSION`(核融合引擎在那一層)。
// **兩條獨立的線指到同一個編號**,不是巧合。
//
// ⚠ 反組譯裡 `byte_199CB5 >= 3` 那條路徑**沒有設 `var_18`**(`enter` 不清堆疊,值是殘留的)。
// 那不是這裡要還原的東西:NEW GAME 只給三級(選項數 3、`NEWGAME.LBX` 資產 20–22 三張圖),
// 所以 3 這個值根本進不來。記在這裡只是說明「為什麼不做第四級」。

// StartingTopicOrder 是開局依序送出的六個研究主題(原版 `word_18111C`,順序照原版)。
//
// 順序有意義:TECH LEVEL 越低就從這張表**取前面幾個**,不是隨機挑。
var StartingTopicOrder = [6]ResearchTopic{
	TOPIC_ENGINEERING,     // 29 = 0x1D
	TOPIC_NUCLEAR_FISSION, // 55 = 0x37(FTL 所在的主題)
	TOPIC_CHEMISTRY,       // 22 = 0x16
	TOPIC_PHYSICS,         // 57 = 0x39
	TOPIC_ELECTRONICS,     // 28 = 0x1C
	TOPIC_COLD_FUSION,     // 23 = 0x17
}

// StartingTopicCount 回傳某個 TECH LEVEL 開局要送幾個主題(原版的 `var_18`)。
//
// 超出 0..2 一律回「一般」那一級的 6:NEW GAME 只給三級,而原版對第四級是讀到殘留值
// ——照抄一個未初始化的堆疊值不叫還原。
func StartingTopicCount(techLevel int) int {
	switch techLevel {
	case 0:
		return 1
	case 1:
		return 6
	case 2:
		return 25
	}
	return 6
}

// StartingTopics 回傳某個 TECH LEVEL 開局送的主題清單。
//
// ⚠ 只回**固定表**的部分。原版超過 6 個之後是 `sub_FD335` 逐個隨機挑。
// 那一段 2026-08-07(第 44 項)接上了,在 `starting_random_tech.go`
// ——這支仍然只回固定表,隨機那一半由 `shell.applyStartingRandomTech` 另外發。
// 兩者的數量關係由 `StartingTopicRandomExtras` 表示。
func StartingTopics(techLevel int) []ResearchTopic {
	n := StartingTopicCount(techLevel)
	if n > len(StartingTopicOrder) {
		n = len(StartingTopicOrder)
	}
	if n < 0 {
		n = 0
	}
	out := make([]ResearchTopic, n)
	copy(out, StartingTopicOrder[:n])
	return out
}

// StartingTopicRandomExtras 回傳原版在固定表之外**還會隨機送幾個**。
//
// 這個函式原本的理由是「把缺口變成一個看得見的數字」:先進開局照原版應該再隨機拿 19 個,
// 而 remake 當時沒送。
//
// **2026-08-07(第 44 項)那 19 個接上了**,所以它現在的角色變成「還要再發幾個」
// ——`shell.applyStartingRandomTech` 直接用這個數字當迴圈次數。
func StartingTopicRandomExtras(techLevel int) int {
	n := StartingTopicCount(techLevel) - len(StartingTopicOrder)
	if n < 0 {
		return 0
	}
	return n
}
