# 原版 MOO2 新遊戲流程 + 種族選擇畫面(對原版盤點)

> 目的:忠實還原原版「按 New Game 之後」的完整畫面序列,取代目前 remake 的「單一設定畫面 + 種族循環框 → 程序生成假星系」。
> 日期:2026-07-10。oracle:`moo2_patch1.5/GAME_MANUAL.pdf`(gameplay 手冊,188 頁,權威)+ 原版 LBX 資產渲染。**openorion2 的 `newGame()` 是 STUB(未實作),不能當流程參考**(證實 openorion2 只是渲染殼)。
> 驗收原則:畫面版面/座標最終要對原版截圖像素對齊(rulebook 64/65),本檔區分「手冊已證流程」「資產尺寸已量」「讀圖已確認」與「待確認」。

## 一、原版新遊戲畫面序列(手冊已證,附頁碼)

| # | 畫面 | 玩家操作 | 手冊頁 |
|---|---|---|---|
| 1 | **Universal Menu**(主選單)| 點 New Game | p.8 |
| 2 | **Galactic Setup**(星系設定)| 8 項決策:Difficulty(5級)/Galaxy Size(5級 20–71星)/Galaxy Age(3級)/Number of Players(2–8)/Starting Civilization(4級)+ 3 開關(Tactical Combat、Random Events、Antaran Attacks);Accept 進下一步、Cancel 回主選單 | p.9–14 |
| 3 | **Race Selection**(種族選擇,**獨立畫面**)| 13 預設種族 + Custom;滑鼠移到種族上顯示肖像 + 特殊能力 | p.14–18 |
| 4 | **Race Customization**(自訂點數,**選 Custom 才有**)| 先選一族肖像當外觀 → 3 欄 11 主題加/扣分,起始 10 Picks / 200% Score,可 Clear 歸零;可改種族名 | p.18–23 |
| 5 | **命名 + 選旗色** | 接受建議名或自訂輸入,選旗幟顏色 | p.14 |
| 6 | 進正式遊戲 | 母星/起始殖民地/起始科技由**步驟 2 的 Starting Civilization** 決定(非種族畫面):Pre-warp 只有母星無 FTL、Average 有母星+基本艦隊含殖民船、Advanced 已有多科技與跨星系艦隊 | p.13 |

**三個關鍵確認(手冊原文)**:
1. 種族選擇是**獨立畫面**:「Once you finish setting up the galaxy, the Race Selection screen appears」(p.14)。
2. **Custom 點數畫面確實獨立存在**:Race Customization,3 欄 11 主題,10 Picks/200%(p.18–23),政府型態(Feudal/Dictatorship/Democracy/Unification…)與特殊能力各有具體數值。
3. 選完種族→進遊戲前只有**命名 + 旗色**兩步(p.14);起始殖民地/科技來自 Starting Civilization 設定,不屬種族畫面。

## 二、13 預設種族(手冊 p.15–18,字母序)

Alkari · Bulrathi · Darloks · Elerians · Gnolams · Humans · Klackons · Meklars · Mrrshan · Psilons · Sakkra · Silicoids · Trilarians(+ Custom)= 14 項。

## 三、RACESEL.LBX 資產地圖(138 資產;尺寸已量,肖像已讀圖確認)

| asset | 尺寸 | 幀 | 內嵌調色盤 | 判定 |
|---|---|---|---|---|
| 0 | 300×333 | 1 | 無 | 肖像框/局部背景(待讀圖)|
| **1–14** | 123×45 | 2 | 無 | **13 族 + Custom 的名稱按鈕**(2 幀=一般/高亮),數量 14 對上種族數 |
| **15–28** | 290×322 | 1 | **有** | **14 張種族全身肖像**——★asset 15/16/20 已讀圖確認為原版種族肖像(15=有翼鳥龍族/Alkari 樣、16=貓爬蟲族)。順序推定同字母序,逐 index 對應待確認 |
| 29 | 209×39 | 1 | 無 | 標題/標籤 |
| 30 | 315×160 | 1 | 無 | 種族描述文字底板(待讀圖)|
| 31 | 111×31 | 2 | 無 | 按鈕(Accept/Last Race?)|
| 32 | 534×341 | 1 | 有 | 大面板;渲染近全透(key-color),疑為疊在框內的內容層,**非滿版底圖**(待讀圖)|
| 33 | 219×29 | 1 | 無 | 標籤 |
| 34–137 | ~84×94(8 尺寸循環)| 1 | 無 | 104 = 8×13,每族 8 張小頭像/圖示變體(用途待讀圖)|

**注**:RACESEL 內沒有 640×480 滿版圖;最大 534×341。滿版底框可能沿用他處或 NEWGAME 資產,待讀圖確認畫面如何組合。

## 四、NEWGAME.LBX 資產地圖(30 資產;對照現有設定畫面)

| asset | 尺寸 | 判定 |
|---|---|---|
| 0 | 94×33(2幀)| 按鈕 |
| 1–22 | 65×65 | 5 組可循環圖示(Difficulty/Size/Age/#Players/StartCiv)各級別圖;分組待讀圖 |
| 23–27 | ~144×32 / 95×22(2幀)| 3 開關 + Accept/Cancel |
| **28** | **640×480** | **設定畫面滿版底圖**(現有 `newGameSetup()` 已用此,座標已 PIL 量測)|
| 29 | 185×42 | 標題文字圖 |

## 五、現有 remake 差距

現況 `cmd/moo2/interactive.go` 的 `newGameSetup()`:
- 只有**單一 NEWGAME#28 設定畫面**,種族是設定畫面內一個**循環框**(非獨立畫面);
- ACCEPT 後直接 `RegenGalaxy`(程序生成假星系)+ `ApplyRace`(套自編加成),**跳過**種族選擇畫面、Custom 點數、命名/旗色、真實母星初始。

## 六、還原計畫(分步,逐畫面對原版對齊)

1. ✅ **獨立種族選擇畫面**(task 7,2026-07-10 完成合成版):`cmd/moo2/raceselect.go`——RACEOPT#0 螢幕框 + 14 族中文名清單 + 選定顯示真肖像(RACESEL 15–28,字母序,15=Alkari/20=Humans 讀圖確認)+ 描述 + 取消/接受。設定畫面 Accept 改導向此畫面。**已 headless 渲染確認名稱↔肖像正確**。⚠ 版面為合成近似,**尚未對原版截圖像素對齊**(待取得原版 Race Selection 截圖後校正,接優先3)。用 RACESEL 自身的名稱按鈕圖(1–14)取代文字、描述板(30)亦待接。
2. ✅ **種族自訂點數畫面**(2026-07-10 完成 v1):`cmd/moo2/customrace.go`——RACEOPT#0 框 + 生產類循環(人口成長/農業/工業/研究/商業/艦艇攻擊/政府)+ 11 特殊能力開關 + 即時「剩餘點數/10」+ 互斥 + 超支不可接受。點數為官方 patch 1.5 真值(見 `custom-race-picks.md`)。種族選擇選「自訂種族」→ 此畫面 → 接受聚合數值加成開局。⚠ 生產/成長/戰鬥類數值有實際套用;**政府型態/特殊能力深層效果尚未模擬**(只計點數);版面待像素對齊。
3. ✅ **命名 + 旗色畫面**(2026-07-10 完成):`cmd/moo2/nameflag.go`——RACEOPT#0 框 + 可鍵盤編輯帝國名(預填「<族>帝國」)+ 8 旗色;開始遊戲存 `session.PlayerName/FlagColor`。**新遊戲流程鏈補完**:主選單→設定→種族選擇→[自訂點數]→命名旗色→進遊戲。
4. **真實起始狀態**(⚠ 待 oracle):依 Starting Civilization 設定真實母星/殖民地/科技,取代 `RegenGalaxy` 亂數(對照手冊 p.13 四級 civ)。**手冊無精確起始數值,需 DOSBox 開局參考存檔解析,否則會自編近似——暫緩以守不臆造鐵律**。
5. 每步驟以 xvfb 截圖 + 原版對照,熱區逐一量測(接 HANDOFF 優先3)。

## 待辦(讀圖確認項)

- [ ] 讀圖確認 asset 0/30/32 用途、34–137 小圖用途、名稱按鈕(1–14)版面座標。
- [ ] 逐 index 對應肖像 15–28 ↔ 13 族字母序 + Custom。
- [ ] NEWGAME 1–22 的 5 組循環圖示分組。
- [ ] 取得原版 Race Selection 畫面截圖(DOSBox 或他版)做像素對齊基準。
