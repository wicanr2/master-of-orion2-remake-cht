package gamedata

// starlane.go:星圖上的距離與速度(秒差距)。
//
// 這一檔存在的理由:手冊有三條規則都以「秒差距/回合」表述,而 remake 先前的星圖移動是
// `ETA = ceil(正規化距離 × 8)` —— 一個沒有速度概念的固定換算,那三條規則因此**無處可掛**:
//
//	星雲    「Ships traveling through a nebula are reduced in speed to 1 parsec per turn.」
//	黑洞    「No ship can safely pass within 2 parsecs of a black hole
//	         (unless the ship contains an officer with the Navigator skill).」
//	領袖    Navigator:「increases the speed of the fleet by 1 or 2 parsecs per turn.」
//	建築    Warp Field Interdictor:「radius of 3 full parsecs … slows all enemy ships
//	         approaching the system to a speed of 1 parsec per turn.」
//
// ============ 一、1 秒差距 = 30 個遊戲座標單位(真值)============
//
// `Parsecs_Between_Points_` @ 0xEBE79 整支就是:
//
//	d² = dx² + dy²
//	回傳最小的 p 使得 p² × 900 ≥ d²      ; 900 = 30²
//
// 也就是 **`秒差距 = ceil(距離 / 30)`**,而且結果是**整數**(原版全程用整數秒差距)。
// 那個 `900` 是立即數,不是推的。
//
// ============ 二、四種銀河大小的尺寸(真值)============
//
// `sub_1693B6` 的 4 路跳表(@ 0x8CF1B)逐檔寫死寬高與 SizeFactor:
//
//	檔位 0:506 × 400   單位 = 16.9 × 13.3 秒差距,SizeFactor 10
//	檔位 1:759 × 600   單位 = 25.3 × 20.0 秒差距,SizeFactor 15
//	檔位 2:1012 × 800  單位 = 33.7 × 26.7 秒差距,SizeFactor 20
//	檔位 3:1518 × 1200 單位 = 50.6 × 40.0 秒差距,SizeFactor 30
//
// 三重交叉驗證都對上:寬恆為 `SizeFactor × 50.6`、高恆為 `SizeFactor × 40`;
// 而原版存檔 SAVE10.GAM 讀出來就是 `759 × 600`、`SizeFactor = 15`(檔位 1)。
//
// ============ 三、星數 → 檔位(真值)============
//
// `Galaxy_Size_From_N_Stars_` @ 0x798D2 的門檻是 **20 / 36 / 54 / 72**。
//
// ============ 四、引擎速度(手冊逐條)============
//
// 手冊的元件說明一條一條寫死,而且都補了同一句
// 「This drive is added to all your ships … as soon as you complete your research」——
// **引擎是全帝國自動升級,不是單艦掛載的元件**。所以速度只看「已研究到的最高階引擎」。

// ParsecUnits 是 1 秒差距的遊戲座標單位數(`Parsecs_Between_Points_` 裡的 900 = 30²)。
const ParsecUnits = 30

// GalaxyDim 是一種銀河大小的尺寸(遊戲座標單位)與 SizeFactor。
type GalaxyDim struct {
	Width, Height int
	SizeFactor    int
}

// GalaxyDims 是四種銀河大小的尺寸,逐檔抄自 `sub_1693B6` 的跳表。
var GalaxyDims = [4]GalaxyDim{
	{506, 400, 10},
	{759, 600, 15},
	{1012, 800, 20},
	{1518, 1200, 30},
}

// GalaxyStarCounts 是四檔的星數,取自 `Galaxy_Size_From_N_Stars_` 的門檻(20/36/54/72)。
var GalaxyStarCounts = [4]int{20, 36, 54, 72}

// GalaxySizeFromStars 是 `Galaxy_Size_From_N_Stars_` @ 0x798D2:星數 → 銀河大小檔位。
//
// ⚠ 原版對「超過 72 星」是回傳存檔裡記的那個檔位(而不是 3);remake 沒有那個欄位,
// 夾在 3。差別只在超出原版最大銀河的情況。
func GalaxySizeFromStars(n int) int {
	switch {
	case n <= GalaxyStarCounts[0]:
		return 0
	case n <= GalaxyStarCounts[1]:
		return 1
	case n <= GalaxyStarCounts[2]:
		return 2
	}
	return 3
}

// GalaxyParsecSpan 回傳某檔位銀河的寬高(秒差距,浮點——只有兩點之間的距離才取整)。
func GalaxyParsecSpan(sizeClass int) (float64, float64) {
	if sizeClass < 0 {
		sizeClass = 0
	}
	if sizeClass >= len(GalaxyDims) {
		sizeClass = len(GalaxyDims) - 1
	}
	d := GalaxyDims[sizeClass]
	return float64(d.Width) / ParsecUnits, float64(d.Height) / ParsecUnits
}

// ParsecsBetween 是 `Parsecs_Between_Points_`:兩點的距離換成**整數**秒差距(無條件進位)。
//
// 傳進來的 dx/dy 已經是秒差距(浮點);原版是拿遊戲座標單位算的,除以 30 之後等價。
func ParsecsBetween(dxParsec, dyParsec float64) int {
	d2 := dxParsec*dxParsec + dyParsec*dyParsec
	p := 0
	for float64(p)*float64(p) < d2 {
		p++
	}
	return p
}

// ---- 引擎速度 ----

// DriveSpeeds 是各級 FTL 引擎的每回合秒差距,逐條抄自手冊的元件說明。
//
// 索引即引擎階(0 = 尚未研究任何 FTL)。**曲速前開局沒有 FTL**,那由呼叫端另外判斷
// (見 `GameSession.FleetHasFTL`),不是這裡的 0。
var DriveSpeeds = []int{0, 2, 3, 4, 5, 6, 7}

// DriveTech 是一級引擎:所屬研究主題 + 該主題內要選中的科技。
type DriveTech struct {
	Topic ResearchTopic
	Tech  Technology
}

// DriveTechOrder 是引擎由低到高,對應 `DriveSpeeds` 的 1..6。
//
// 每一級都是「主題內三選一」(核融裂變那一題是 ResearchAll,拿到主題就有引擎),
// 所以判定要走與 `groundEquipTechOwned` 相同的規則,不能只看主題完成。
var DriveTechOrder = []DriveTech{
	{TOPIC_NUCLEAR_FISSION, TECH_NUCLEAR_DRIVE},              // 2 秒差距/回合(手冊:FTL 裡最慢的)
	{TOPIC_ADVANCED_FUSION, TECH_FUSION_DRIVE},               // 3
	{TOPIC_ION_FISSION, TECH_ION_DRIVE},                      // 4
	{TOPIC_ANTIMATTER_FISSION, TECH_ANTIMATTER_DRIVE},        // 5
	{TOPIC_HYPER_DIMENSIONAL_FISSION, TECH_HYPER_DRIVE},      // 6
	{TOPIC_INTERPHASED_FISSION, TECH_INTERPHASED_DRIVE},      // 7
}

// NebulaSpeed 是星雲中的航速(手冊:「reduced in speed to 1 parsec per turn」)。
const NebulaSpeed = 1

// ---- 兩種星門(都是 Achievement 科技,效果只在**自己的殖民地之間**)----

// JumpGateSpeedBonus 是躍遷門的加速(手冊 Jump Gate:「increases the speed of your ships
// traveling between two of your colony systems by 3 parsecs a turn」)。
const JumpGateSpeedBonus = 3

// StarGateETA 是星際之門的航程(手冊 Star Gate:「allows instantaneous (1 turn) travel
// between any two of your systems」)。
const StarGateETA = 1

// TechRef 是「某研究主題內的某個科技」。
type TechRef struct {
	Topic ResearchTopic
	Tech  Technology
}

// JumpGateTech / StarGateTech 是兩種星門的科技位置(techtree.go 的多選一)。
var (
	JumpGateTech = TechRef{TOPIC_SUBSPACE_PHYSICS, TECH_JUMP_GATE}
	StarGateTech = TechRef{TOPIC_TEMPORAL_PHYSICS, TECH_STAR_GATE}
)

// InterdictorSpeed 是被 Warp Field Interdictor 場籠罩的敵艦航速(手冊同一句話的另一半)。
const InterdictorSpeed = 1

// InterdictorRadiusParsecs 是 Warp Field Interdictor 的作用半徑(手冊:「a radius of 3 full parsecs」)。
const InterdictorRadiusParsecs = 3

// BlackHoleAvoidParsecs 是黑洞的禁行半徑(手冊:「No ship can safely pass within 2 parsecs
// of a black hole (unless the ship contains an officer with the Navigator skill)」)。
const BlackHoleAvoidParsecs = 2

// FleetSpeedForDrive 回傳某引擎階的航速;超出範圍夾在兩端。
func FleetSpeedForDrive(tier int) int {
	if tier < 0 {
		tier = 0
	}
	if tier >= len(DriveSpeeds) {
		tier = len(DriveSpeeds) - 1
	}
	return DriveSpeeds[tier]
}

// DriveTierFromTechs 回傳已研究到的最高引擎階(0 = 一階都沒有)。
//
// owned(topic, tech) 回傳該科技是否已擁有。手冊說引擎會自動裝到全帝國的船上,所以只看最高階。
func DriveTierFromTechs(owned func(ResearchTopic, Technology) bool) int {
	tier := 0
	if owned == nil {
		return tier
	}
	for i, d := range DriveTechOrder {
		if owned(d.Topic, d.Tech) {
			tier = i + 1
		}
	}
	return tier
}
