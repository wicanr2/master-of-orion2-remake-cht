# 格子戰術戰鬥畫面:原版資產清單與重建規格

> 日期:2026-07-10。狀態:**資產結構已 dump 驗證**;調色盤來源為假設待實作時渲染核對。
> 目的:把現行自繪 tacticalScreen(星空底+格線+token)換成原版 COMBAT/STARBG/CMBTSHP 美術。
> 方法:`cmd/lbxinfo` dump 結構;openorion2 **無 combat 渲染器**(只渲染 galaxy/主畫面),故 combat 調色盤無權威參考,需渲染實測。

## 一、資產清單(lbxinfo 實測)

### STARBG.LBX(戰場背景,54 資產)
| idx | 尺寸 | 用途 |
|---|---|---|
| 0–5 | 640×480 | **6 張全螢幕星空戰場背景**(無內嵌調色盤,需借) |
| 6+ | 各種 | 星球/星雲(部分含 32/256 色內嵌 palette) |

openorion2 galaxy 用 `STARBG#3` 當背景,palette = `_gui->palette()`(全域 GUI 調色盤)。戰鬥背景推測同樣借共用調色盤。

### COMBAT.LBX(戰鬥 UI chrome,90 資產)
| idx | 尺寸 | flags | 用途 |
|---|---|---|---|
| 0 | 640×129 | 0x0000 | **底部控制列**(指令按鈕面板) |
| 5 | 341×406 | 0x0000 | 側面資訊面板 |
| 9 | 170×180 | 0x0000 | 面板 |
| 11 | 1×1 | 0x1000 | **256 色調色盤源**(COMBAT 場景 palette;供 STARBG/CMBTSHP 借用之候選) |
| 1,2,3,6,7,8,10 | 小 | 0x0400 FILLBG | 按鈕(2 幀 up/down) |
| 12–22+ | 85×80 | 0x0100 NOCOMPRESS | UI 圖示/按鈕 |

### CMBTSHP.LBX(艦艇 sprite,360 資產)
- 全部 **59×60、20 幀、flags 0x0000**。20 幀 = **旋轉方向/朝向**(戰鬥中艦艇轉向)。
- 360 = 多艦型 × 尺寸 × 種族。無內嵌調色盤,借戰鬥場景 palette。
- 另有 CMBTFGTR(戰機)、CMBTMISL(飛彈)、CMBTPLNT(戰鬥星球,含 32 色內嵌 palette)。

## 二、調色盤來源(已渲染核實 2026-07-10)

**結論:`COMBAT.LBX#11`(1×1、flag 0x1000、256 色)是 STARBG#0–5 與 CMBTSHP 的正確調色盤。**

核實方法(rulebook 64 拓樸比對):
- STARBG 是**稀疏 RLE**(僅 ~6.6% 像素有寫入),其餘未寫入=**透明**,原版設計是**疊在純黑太空背景上**。
- ⚠ 陷阱:直接看渲染 PNG 會把透明區當白色 → 誤判「全畫面雜點」。三個不相關候選色盤得到**同形狀**雜點就是鐵證(色盤只改顏色不改透明遮罩)。**必須疊黑底**(`alpha_composite` over black)才看得出真圖。
- 疊黑底後比色彩連貫:COMBAT#11 → STARBG 五變體(#0,1,2,4,5)乾淨藍白星點+星雲、CMBTSHP(#0,#20)乾淨灰白艦體+紅色引擎/武器重點;BUFFER0#0(galaxy 全域 GUI 色盤)星空可但**艦艇錯成洋紅**;MAINMENU#1 全錯。
- ∴ 借對色盤是 **per-screen**(COMBAT.LBX 自帶戰鬥專屬色盤),非隨便借全域色盤。同 DIPLOMAT 專職色盤持有資產慣例。

工具:`cmd/moo2` 已加 `-pallbx <file>` 旗標支援跨-LBX 借色盤:
`-lbx STARBG.LBX -asset 0 -pallbx COMBAT.LBX -palasset 11 -shot out.png`
實作時:先鋪**純黑**底,再貼 STARBG(透明處透出黑),艦艇/UI 同借 COMBAT#11。

## 三、重建規格與進度

### Phase 1 — ✅ 已完成(2026-07-10,實作子代理 + 主代理核實)

`tacticalScreen`(interactive.go)視覺層換原版美術,**戰鬥數學/流程/RNG 零改動**:
- `loadCombatBG`:STARBG#0 借 COMBAT#11 → 黑底星空背景(未寫入處透明疊黑)。
- `loadCombatBar`:COMBAT#0 借 COMBAT#11 → 底部控制列(渲染出原版真 UI:WEAPONS/SPECIALS/AUTO/SCAN/BOARD/RETREAT/WAIT/DONE/OPTIONS 按鈕 + 迷你星圖)。
- `loadCombatShip`:CMBTSHP#0 frame0 借 COMBAT#11,keyColor → 59×60 艦艇 sprite(佔位:全艦共用)。
- draw():黑底 → STARBG → 淡格線 → 艦艇 sprite(敵方水平翻轉)→ 控制列;三者任一載入失敗都 fallback 回原自繪。
- 主代理核實:控制列/艦艇/背景渲染圖皆色彩連貫無雜色;diff 確認戰鬥邏輯未動。

### Phase 1 端到端驗證(2026-07-10,`-gamegallery` 截圖廊)
補了 `cmd/moo2` 的 `-gamegallery <dir>` headless 導覽模式:主選單→種族→命名→星系→殖民地→研究→**外交→戰鬥**,各畫面存一張圖。實測 8 張全渲染:
- 外交:原版議事廳 + **薩克拉使節** + 全繁中對談選項 ✓
- 戰鬥:STARBG 星空 + COMBAT 原版控制列 + 中文艦名/提示 ✓
- ⚠ **CPU 教訓**:截圖廊初版沒有終止保護 + 存圖用 `tick==目標`(精確相等),ebiten Update/Draw 解耦會跳幀漏存 → **永不終止的 render loop 空轉燒 CPU**(兩個容器各卡 17–20 分)。已修:存圖改 `tick>=目標`、Update 超過末 tick+3 硬性終止;跑時容器內外雙 `timeout`。**教訓:headless GUI 迴圈必設硬性終止 + timeout**(rulebook 35)。

### Phase 2 進行中(2026-07-10)
- ✅ **控制列按鈕中文化**:7 個實際控制按鈕(AUTO→自動/SCAN→掃描/BOARD→登船/RETREAT→撤退/WAIT→等待/DONE→完成/OPTIONS→選項)已疊深色底+中文蓋掉烘進英文(`drawBarLabelsCHT` + `barButtonsCHT` 座標表,於實際截圖逐像素量測)。log 移到控制列上方星空避免壓按鈕。截圖驗證對齊乾淨。
- ⏳ 未使用清單面板的 WEAPONS/SPECIALS 欄位標頭刻意略過(remake 未顯示武器清單);若之後接上武器列表再中文化。
- ⏳ 艦艇 sprite 仍全共用 CMBTSHP#30(可見佔位);per 艦型/尺寸完整對照待做。

### Phase 1 遺留(後續)
- ⚠ **艦艇 sprite 太暗看不見**:CMBTSHP#0 是小型深色戰機(疊灰底可見輪廓,疊黑星空幾乎隱形)。sprite 載入/縮放/翻轉管線已通(debug 確認 t.ship 非 nil、59×60、20 幀),但 #0 不適合當佔位。Phase 2 需挑「較大/較亮」或按實際艦型選 sprite,並考慮加選取高亮/描邊讓艦艇在星空上可讀。現況艦艇仍主要靠標籤方框+艦名辨識(功能正常)。
- keyColor 修正:loadCombatShip 從強制 true 改為 `im.KeyColor()`(CMBTSHP flags=0→false);強制 true 會把 index-0 艦體判透明。

### 原始規格(Phase 1 依此,保留供追溯)

現行 `tacticalScreen`(interactive.go)為自繪:8×6 方格 + token + HP 條,**戰鬥數學已接 gamedata 真公式(保留)**。改動:

1. **背景**:載 STARBG#0(或依星域選 0–5),cross-LBX 借 COMBAT#11 palette,640×480 鋪底,取代 `dst.Fill` 星空。
2. **控制列**:COMBAT#0(640×129)貼底部(y≈351),取代自繪 log 列;文字/按鈕疊其上。
3. **艦艇**:CMBTSHP sprite(59×60)取代彩色 token;先用 frame0(朝向後續再依移動向量選 20 幀之一)。艦型→asset index 對應待查(先固定一款示範)。
4. **保留**:命中/傷害/過盾/過甲真公式(`combat_formula.go`、`ResolveShot`)、回合制流程、RNG 種子。

**驗證**:headless `-game` 導覽到戰鬥截圖,對照原版戰鬥畫面構圖(背景+底列+艦艇)。palette 借對則艦艇/背景色彩連貫,借錯全畫面雜點(同 DIPLOMAT 配對律教訓,見 `diplomat-lbx-layout.md`)。

## 四、待查(RE 後續)
- CMBTSHP 艦型/尺寸/種族 → asset index 的對應表(360 個)。
- 20 幀朝向的角度對應(哪一幀對哪個航向)。
- STARBG 6 張背景是否依星域/星球類型選用。

## 追記(2026-07-11):CMBTSHP vs ships.lbx —— #12 艦型對照為何 RE-gated

派工前對 openorion2 做柵欄原則核實,釐清一個關鍵歧義,記錄以免未來重踩:

- **openorion2 完全不使用 CMBTSHP.LBX**(全 `src/*.cpp` grep 零命中)。它的艦艇 sprite 來自 **`ships.lbx`**(`ships.cpp` `SHIPSPRITE_ARCHIVE`),索引公式 `ships.cpp:112-114`:`ships.lbx[玩家色 × (MAX_SHIP_SPRITES+1=50) + ship.picture]`,配玩家色調色盤;`ship.picture`(0-48,`MAX_SHIP_SPRITES=49`)存在存檔 ShipDesign(remake `internal/save` 已解析 `Picture uint8`)。特殊 sprite:ANTARAN=0/GUARDIAN=7/MONSTER=8/MINIMONSTER=20。
- **但 `ships.lbx` 的 `_shipimg` 在 openorion2 是否用於戰術戰鬥畫面,無法從碼確認**;而 **CMBTSHP.LBX 資產 0 有 20 幀**(旋轉動畫),與「戰術戰鬥中艦艇會朝不同方向」的需求吻合——強烈暗示 **CMBTSHP 才是 DOS 原版戰術戰鬥的艦 sprite**,`ships.lbx` 可能是星系/艦隊/設計畫面的靜態小圖。故 remake 現行選 CMBTSHP 疊戰場**很可能是對的**,不應貿然改成 ships.lbx。
- **結論:#12(艦級 → CMBTSHP 資產索引 完整對照)確實 RE-gated。** 對照關係**不在 openorion2 參考碼**(它不碰 CMBTSHP),CMBTSHP 360 資產的分組語意(哪些資產對應哪個艦體大小/種族/朝向)須靠 ① DOSBox 原版戰鬥截圖當 oracle 逐一比對,或 ② 逆推 CMBTSHP.LBX 內部結構(幀數/尺寸分組)+ 原版艦體美術交叉驗證。非本輪「接已備 gamedata 模型」式的乾淨自驅任務。現行「全艦共用 CMBTSHP#30 佔位」維持,待 DOSBox oracle 就緒再逐級對照。
