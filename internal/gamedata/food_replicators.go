package gamedata

// food_replicators.go:**食物複製機**——把工業產能換成食物,救的是饑荒。
//
// 手冊 p.85 一整句就是完整規格:
//
//	Having this facility in a colony allows you to convert industrial production into
//	food on a **two-for-one basis** at a cost of **1 BC per food**, **as needed**.
//
// 三個限定詞各對應一條規則,少一條就不是原版:
//
//	two-for-one  → 2 產能換 1 食物
//	1 BC per food→ 換出來的每一單位食物再花 1 BC(從國庫,不是從產能)
//	as needed    → **只補足缺口**,不會換出盈餘去賣錢
//
// 最後那條最容易漏。少了它,一個有複製機的殖民地會把全部產能換成食物、
// 再靠「餘糧出售」換回 BC——那是一台印鈔機,而且原版沒有。
//
// ============ 維護費 10 BC 是它的平衡設計 ============
//
// | 來源 | 維護費 |
// |---|---|
// | 手冊 p.85 | 10 BC |
// | remake 建築表(原版執行檔 `off_17EB3D + 12`) | **10** |
//
// 10 BC 在建築表裡是**最貴的一棟**(第二貴是 5)。這不是隨手訂的:一棟能無視農業
// 產出把饑荒填平的建築,代價就是每回合固定燒錢。兩個獨立來源給同一個數字。
//
// ============ 誠實留白 ============
//
// 手冊沒說「國庫不夠付 1 BC/食物 時會怎樣」。這裡**不編規則**:換算照做、成本照報,
// 由帝國層跟其他維護費一起結算(國庫可以被壓成負數,那是既有行為)。
// 硬加一條「付不起就不換」會是憑空發明的規則,而且會讓饑荒在破產時突然惡化。

const (
	// FoodReplicatorProductionPerFood 是換 1 單位食物要花幾點產能(手冊 two-for-one)。
	FoodReplicatorProductionPerFood = 2
	// FoodReplicatorBCPerFood 是換 1 單位食物要花幾 BC(手冊 1 BC per food)。
	FoodReplicatorBCPerFood = 1
	// FoodReplicatorIndustryHalfPerHalfFood 是精確半單位帳本中，換出半單位食物
	// 要花的半產能。1 個完整產能 = 2 個半產能；手冊的 2:1 因此是 2 半產能
	// 換 1 半食物，也就是 2 完整產能換 1 完整食物。
	FoodReplicatorIndustryHalfPerHalfFood = 2
	// FoodReplicatorBCHalfPerHalfFood 是半食物的半 BC 成本。兩個半食物合成
	// 一個完整食物，剛好回到手冊的 1 BC per food。
	FoodReplicatorBCHalfPerHalfFood = 1
)

// FoodReplicatorConvertHalf 是 FoodReplicatorConvert 的精確版。
//
// deficitHalf 與 netIndustryHalf 都以「半單位」傳入。Cybernetic 的奇數人口
// 可能留下半食物缺口；若這裡先除以 2，就會把那半單位錯誤捨掉。半單位成本
// 是對手冊整數規格的最小延伸，並由帝國層累積半 BC，避免每回合無聲遺失。
// 這條半 BC 行為屬**強推論**：手冊只明寫完整食物 1 BC，原版靜態資料未直接
// 告知破碎單位的付款時機；實作選擇與存檔半單位帳本及 1 BC/food 保持一致。
func FoodReplicatorConvertHalf(deficitHalf, netIndustryHalf int) (foodHalf, productionSpentHalf int) {
	if deficitHalf <= 0 || netIndustryHalf <= 0 {
		return 0, 0
	}
	affordable := netIndustryHalf / FoodReplicatorIndustryHalfPerHalfFood
	foodHalf = deficitHalf
	if foodHalf > affordable {
		foodHalf = affordable
	}
	return foodHalf, foodHalf * FoodReplicatorIndustryHalfPerHalfFood
}

// FoodReplicatorConvert 算出這個殖民地這回合會把多少產能換成食物。
//
// deficit 是食物缺口(正數;盈餘或剛好時傳 0 或負數,回 0)。
// netIndustry 是扣完污染之後可用的產能。
//
// 回傳:換出幾單位食物、花掉幾點產能。BC 成本由呼叫端用 FoodReplicatorBCPerFood 乘。
//
// 「as needed」在這裡就是 `min(缺口, 產能換得起的量)`——不會超過缺口,也不會超過產能。
func FoodReplicatorConvert(deficit, netIndustry int) (food, productionSpent int) {
	if deficit <= 0 || netIndustry <= 0 {
		return 0, 0
	}
	// 舊 API 的契約是完整食物／完整產能；不能把奇數產能透過半單位
	// 轉換折回來，否則既有呼叫端會看到花 7 產能換 3 食物而破壞 2:1
	// 的整數相容規則。需要半單位時由 FoodReplicatorConvertHalf 直接處理。
	affordable := netIndustry / FoodReplicatorProductionPerFood
	food = deficit
	if food > affordable {
		food = affordable
	}
	return food, food * FoodReplicatorProductionPerFood
}
