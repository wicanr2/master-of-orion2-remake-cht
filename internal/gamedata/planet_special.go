package gamedata

import "math/rand"

// 行星特殊物產(12 種)。
//
// 四個來源互相印證:
//   - 種類與編碼:openorion2 `enum SpecialType`(0..11)
//   - 出現機率:反組譯 `_planet_special_weighted_chance` @ 0x17D832(12 bytes,**總和恰為 100**)
//   - 定性效果:GAME_MANUAL.pdf 逐條原文(見各常數註解)
//   - **數量級效果**:反組譯 `Do_System_Discoveries_At_Star_` @ 0xE9927 與
//     `Make_New_Colony_Or_Outpost_` @ 0xE5EB3 的實際指令(手冊只說「加進國庫」不給金額,
//     這兩個函式給了)
//
// 交叉驗證:手冊說寶石礦 +10 BC/回合、金礦 +5 BC/回合(兩倍),而反編出來的 AI 估值加分
// 正好也是兩倍(2560 vs 1280)——兩個獨立來源對同一組相對關係給出一致的答案。
//
// 讀組語時怎麼確定「這個位元組真的是 special」:`Make_New_Colony_Or_Outpost_` 裡的行星指標
// stride 是 0x11 = 17 bytes,正好是 openorion2 `struct Planet` 的大小,而它比較的
// `[planet+0x0F]` 對上該結構的 `special` 欄位;同一個函式還寫 `[colony+0x0A]`(= population)、
// `[colony+0x0C]`(= colonists[])、`[colony+0xE2]`(= climate),三個偏移全部對得上,
// 且符號表裡另有一個獨立函式就叫 `Planet_Has_Splinter_Colony_`,內容正是 `[planet+0x0F] == 7`。
//
// remake 先前完全沒有這個系統:每顆行星都一樣,沒有「這顆星有寶石礦,搶下來」這種決策。

// PlanetSpecial 是行星特殊物產代碼(對齊 openorion2 SpecialType)。
type PlanetSpecial int

const (
	NoSpecial        PlanetSpecial = 0
	BadSpecial1      PlanetSpecial = 1  // 手冊未描述效果,誠實留白
	SpaceDebris      PlanetSpecial = 2  // 手冊:探索時一次性入袋(金額未給)
	PirateCache      PlanetSpecial = 3  // 手冊:探索時一次性入袋(金額未給)
	GoldDeposits     PlanetSpecial = 4  // 手冊:殖民後 +5 BC/回合
	GemDeposits      PlanetSpecial = 5  // 手冊:殖民後 +10 BC/回合
	Natives          PlanetSpecial = 6  // 手冊:原住民併入人口,只當農夫,食物 +2
	SplinterColony   PlanetSpecial = 7  // 手冊:殖民時直接有 3 人口
	LostHero         PlanetSpecial = 8  // 手冊:困在此地的傭兵領袖,救出後免費加入
	BadSpecial2      PlanetSpecial = 9  // 手冊未描述效果;原版出現機率也是 0
	AncientArtifacts PlanetSpecial = 10 // 手冊:科學家每人研究 5 而非 3
	OrionSpecial     PlanetSpecial = 11 // 獵戶座(原版出現機率 0,由星系生成另行指定)
)

// planetSpecialWeights 是各特殊物產的出現權重(原版 `_planet_special_weighted_chance`
// @ 0x17D832)。索引 = PlanetSpecial,**總和恰為 100**——這是「12 欄真的對應 SpecialType 0..11」
// 的獨立佐證。
var planetSpecialWeights = [12]int{64, 5, 3, 2, 3, 2, 9, 4, 3, 0, 5, 0}

// planetSpecialNames 是各特殊物產的中文顯示名。
var planetSpecialNames = [12]string{
	"", "異常", "太空殘骸", "海盜藏寶", "金礦", "寶石礦",
	"原住民", "失散殖民地", "受困英雄", "異常", "遠古文物", "獵戶座",
}

// PlanetSpecialName 回傳中文顯示名;無特殊物產回空字串。
func PlanetSpecialName(s PlanetSpecial) string {
	if s < 0 || int(s) >= len(planetSpecialNames) {
		return ""
	}
	return planetSpecialNames[s]
}

// PlanetSpecialWeight 回傳該特殊物產的出現權重(%)。
func PlanetSpecialWeight(s PlanetSpecial) int {
	if s < 0 || int(s) >= len(planetSpecialWeights) {
		return 0
	}
	return planetSpecialWeights[s]
}

// RollPlanetSpecial 依原版權重骰一個特殊物產(64% 是「無」)。
func RollPlanetSpecial(r *rand.Rand) PlanetSpecial {
	w := make([]int, len(planetSpecialWeights))
	copy(w, planetSpecialWeights[:])
	return PlanetSpecial(WeightedChoice(r, w))
}

// --- 效果(手冊定性 + 反組譯定量)---

// SpecialIncomePerTurn 回傳該特殊物產給殖民地的每回合 BC 加成。
// 手冊原文:「Gems: … Any colony established on that planet generates +10 BC every turn.」
// 「Gold: … generates +5 BC every turn.」
func SpecialIncomePerTurn(s PlanetSpecial) int {
	switch s {
	case GemDeposits:
		return 10
	case GoldDeposits:
		return 5
	}
	return 0
}

// SpecialFoodPerFarmerBonus 回傳該特殊物產給每農夫的食物加成。
// 手冊原文(Natives):「They work only as farmers (at a +2 food production advantage)」。
func SpecialFoodPerFarmerBonus(s PlanetSpecial) int {
	if s == Natives {
		return 2
	}
	return 0
}

// SpecialResearchPerScientist 回傳該特殊物產下每科學家的研究產出;0 = 不覆蓋、用一般值。
// 手冊原文(Artifacts):「Each unit of population assigned to science produces 5 research
// points instead of the usual 3」。
func SpecialResearchPerScientist(s PlanetSpecial) int {
	if s == AncientArtifacts {
		return 5
	}
	return 0
}

// NativePopulationUnits 是原住民行星被殖民時額外加入的人口單位數。
//
// 原版 `Make_New_Colony_Or_Outpost_` @ 0xE5EB3:special == 6 時,除了殖民船帶來的那 1 個
// 人口單位(colony+0x0C)之外,再從 colony+0x10 迴圈寫到 colony+0x1C(stride 4 = sizeof
// Colonist),**正好 3 個**,然後 `[colony+0x0A] = 4`(population)、`[planet+0x0F] = 0`
// (special 被消耗掉)。
//
// 這 3 個人口單位的欄位也逐位元對得上 openorion2 `Colonist::load`
// (race = bits0-3、loyalty = bits4-6、job = bits7-8):race 被寫成 **9**,而 openorion2 的
// `MAX_RACES = MAX_PLAYERS+2`,註解寫明「player races + androids + natives」——8 是機器人、
// **9 就是原住民**;job 位元被清成 0 = 農夫,正好對上手冊「They work only as farmers」。
const NativePopulationUnits = 3

// SpecialExtraPopulationOnColonize 回傳殖民這顆行星時額外加入的人口單位數(原住民 3,其餘 0)。
func SpecialExtraPopulationOnColonize(s PlanetSpecial) int {
	if s == Natives {
		return NativePopulationUnits
	}
	return 0
}

// SpecialConsumedOnColonize 回傳殖民後這個特殊物產是否消失。
// 原版只在原住民分支寫 `[planet+0x0F] = 0`——金礦/寶石礦/遠古文物的效果是持續性的,不消耗。
func SpecialConsumedOnColonize(s PlanetSpecial) bool {
	return s == Natives
}

// --- 抵達星系時的一次性發現(原版 `Do_System_Discoveries_At_Star_` @ 0xE9927)---
//
// 原版把「太空殘骸 / 海盜藏寶 / 失散殖民地 / 受困英雄 / 遠古文物」做成**抵達該星系時觸發一次**
// 的發現事件(不是殖民時),由該函式依 `Star.special` 分派:
//
//	2  → 國庫 += 0x32(50)         ; add dword [player+0x32], 32h
//	3  → 國庫 += 0x64(100)        ; add dword [player+0x32], 64h
//	7  → 就地生出一個殖民地,人口 = min(Colony_Race_Pop_Limit_(), 3)
//	8  → 只設訊息碼(領袖在訊息流程裡加入)
//	10 → 送一項「目前可研究」的科技,亂數挑一項
//
// 分派完之後 `Star.special` 被覆寫成「訊息碼 = 玩家編號 + 50/60/70/80/90」,所以同一個
// 星系不會重複觸發。**行星層的 `Planet.special` 沒有被清掉**——遠古文物的每科學家 5 研究、
// 金礦/寶石礦的每回合收入都是持續效果,靠的就是行星那份沒被動到。

// SpecialDiscoveryBC 回傳抵達該星系時一次性入袋的 BC(太空殘骸 50、海盜藏寶 100)。
// 手冊只說「is added to your treasury」不給金額,這兩個數字來自上述反組譯的兩條 add 指令。
func SpecialDiscoveryBC(s PlanetSpecial) int {
	switch s {
	case SpaceDebris:
		return 50
	case PirateCache:
		return 100
	}
	return 0
}

// SplinterColonyMaxPopulation 是失散殖民地被發現時的人口上限。
// 原版:`call Colony_Race_Pop_Limit_` → `[colony+0x0A] = al`,接著 `cmp al, 3 / jbe / mov
// byte [colony+0x0A], 3`——**取該行星人口上限與 3 的較小值**。手冊只給了 3 這個上限
// (「The colony is small (3 population)」),沒提小行星會更少。
const SplinterColonyMaxPopulation = 3

// SpecialFoundsSplinterColony 回傳抵達時是否會直接生出一個殖民地。
func SpecialFoundsSplinterColony(s PlanetSpecial) bool { return s == SplinterColony }

// SpecialGrantsFreeLeader 回傳抵達時是否會獲得一名免費領袖。
// 手冊原文(Hero):「A mercenary leader has been marooned on one of the planets…
// In gratitude for the rescue…」;原版在這個分支只設訊息碼,實際入列在訊息流程裡。
func SpecialGrantsFreeLeader(s PlanetSpecial) bool {
	return s == LostHero
}

// SpecialGrantsFreeTech 回傳抵達時是否會白送一項科技。
//
// 原版遠古文物分支的完整動作:先看 `[star+0x70]` 是否已用過,沒用過就掃過 0xCC(204)個研究
// 主題記錄,挑出「狀態 == 2」(openorion2 `RSTATE_READY`,即現在就能研究)且通過另一個條件
// 檢查的主題,用蓄水池抽樣選一項送出(`Random_(k) == 1`)。
//
// 送幾項:`Random_(4) / 4 + 1`。原版的 `Random_(n)` **回 1..n**,不是 0..n-1
// (0x1247A0:LCG 取樣、拒絕超界值,最後 `div bucket` 再 `inc eax`),所以 `Random_(4)`
// ∈ {1,2,3,4},整數除 4 得 0(roll 1-3)或 1(roll 4)→ **1 項,25% 機率 2 項**。
//
// ⚠ 訂正:本檔第一版把它記成「恆為 1 項、原版寫壞了」,那是把 Random_ 當成 C 的
// `rand()%n` 語意的誤讀。訂正的方法是回頭讀 Random_ 本身,不是從呼叫端繼續推。
// 同一個誤讀連帶影響另外兩處,都已一併訂正(見 docs/re/01-gap-report.md「Random_ 訂正」):
// 蓄水池抽樣 `Random_(k)==1` 其實是**正確的** 1/k(k=1 時必中,不是「第一個永遠不會被選」);
// 失散殖民地在不可耕行星上的職務 `Random_(2) & 3` 是**工人或科學家**(1 或 2),不是農夫或工人。
func SpecialGrantsFreeTech(s PlanetSpecial) bool {
	return s == AncientArtifacts
}

// ArtifactFreeTechCount 依原版公式回傳遠古文物送幾項科技(roll = Random_(4) 的結果,1..4)。
func ArtifactFreeTechCount(roll int) int {
	if roll < 1 {
		roll = 1
	}
	if roll > 4 {
		roll = 4
	}
	return roll/4 + 1
}

// SpecialIsSystemDiscovery 回傳該特殊物產是否屬於「抵達星系時一次性觸發」那一類。
func SpecialIsSystemDiscovery(s PlanetSpecial) bool {
	switch s {
	case SpaceDebris, PirateCache, SplinterColony, LostHero, AncientArtifacts:
		return true
	}
	return false
}
