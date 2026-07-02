# 字型選擇(kick-off)

> 決定繁體中文顯示用什麼字型。本文列判準與候選,最終需以實機渲染截圖驗證後定案。

## 1. 判準(第一性原理)

1. **授權可散布**:若要打包進發行版,必須是可自由散布的開源授權(OFL / GPL / Apache)。私有商用字型不可。
2. **繁體中文涵蓋**:至少涵蓋常用字 + 遊戲用到的科技/種族/星系術語。缺字會靜默不畫(MOO1 §速查:atlas 缺字 → 中文消失)。
3. **美術風格對齊**:MOO2 是 1996 點陣美術(320×200 upscale)。**像素字型**視覺最契合;向量字(如 Noto)在低解析底圖上會顯得「太新」,但在高解析合成層下可接受且更清楚。
4. **小字可讀**:UI 密集(側欄、清單),字型在小尺寸下要能辨識,尤其筆劃多的繁體字。
5. **ebiten 整合**:`text/v2` + `opentype`(TTF/OTF)或自製點陣 atlas。TTF 最省事;點陣像素字型也多以 TTF 散布。

## 2. 候選

| 字型 | 類型 | 授權 | 繁中 | 備註 |
|---|---|---|---|---|
| **Cubic 11(方舟像素字體)** | 11px 像素 | OFL | ✅ 繁中為主 | 專為繁中設計的像素字,retro 契合度最高;11px 小字需在合成層放大 |
| **Fusion Pixel Font(縫合像素)** | 8/10/12px 像素 | OFL | ✅ | 多尺寸,可依 UI 密度選;涵蓋廣 |
| **Zpix(最像素)** | 12px 像素 | OFL | ✅(含繁) | 常見於獨立遊戲 |
| **Noto Sans CJK TC / 思源黑體** | 向量 | OFL | ✅ 完整 | 涵蓋最全、最清楚;非像素風,檔案大(可 subset) |
| **GNU Unifont** | 16px 點陣 | GPL+font-exc | ✅ | 涵蓋全但美觀差,當保底 fallback |

## 3. 建議(依 mom 實戰修正)

> ⚠️ 先前初稿把像素字型(Cubic 11/Fusion Pixel)列為主字型。經 `08-mom-ebiten-cht-playbook.md` §4 的實戰教訓修正如下。

- **主字型:Noto Sans CJK TC**(OFL,可散布)—— mom 專案**已驗證可行**的選擇。涵蓋最全,`golang.org/x/image/sfnt` 解析穩定。
- **[HARD] 定案前先驗 Go 能否解析**:mom 原想用 AR PL UMing(點陣風),但其 `.ttc` 是 **CFF/舊式,Go sfnt 解析失敗**。→ 任何字型(含下面的像素字型)在定案前,必須先確認 `opentype.Parse` / `ParseCollection` 在 Go 讀得動(glyf 輪廓 TTF 較穩,CFF/.ttc 有風險)。
- **像素風仍是選項,但要先驗**:Cubic 11 / Fusion Pixel 若以 glyf TTF 散布且 Go 解析得動,可作為對齊 MOO2 美術的美術選項;但需注意兩點:①小 logical 字級下筆劃多的繁體字在點陣字可能不可讀;② 我們用 supersample(見 `08` §2),點陣字放大會保留方塊感(retro 可接受,但要截圖確認)。**先做 Noto 打通流程,像素字型當 A/B 比較。**
- **英文**:可續用原版 .lbx 點陣英文字(維持原味),或統一用主字型英數字 —— 截圖比較後定。

> 缺字防護:主字型 Noto TC 涵蓋近全;若之後採像素主字型,務必保留 Noto 向量 fallback(缺字靜默消失是 MOO1 踩過的雷,對齊 CLAUDE.md 完整性優先)。

## 4. 待辦

- [ ] 先用 Noto Sans CJK TC 在 ebiten(supersample 4×)打通渲染流程(對齊 `08` §2)。
- [ ] 驗證候選字型 Go `opentype` 解析可行性(Noto 已知可、像素字型需驗;CFF/.ttc 有風險)。
- [ ] 取得 Cubic 11 / Fusion Pixel(若解析可行),與 Noto 渲染同段遊戲文字(科技名 + 描述 + 按鈕)截圖 A/B 比較。
- [ ] 決定 supersample 倍率與字級門檻(依呼叫端字高,對應 `08` §2 基線 0.82)。
- [ ] 窮舉遊戲全文字集(LBX 字串 + hardcode + UI 字串)做缺字掃描,`pyftsubset`(docker)產子集。
- [ ] 定案後字型子集 go:embed 內嵌 + README 標明授權。
