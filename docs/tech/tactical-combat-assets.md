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

## 三、重建規格(交實作子代理)

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
