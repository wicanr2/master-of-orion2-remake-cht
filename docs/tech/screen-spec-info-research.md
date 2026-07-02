# 資訊/研究畫面渲染規格

分析對象:`InfoView`(`openorion2/src/info.cpp`/`info.h`,含 5 個分頁 Widget)與研究相關畫面
(`openorion2/src/tech.cpp`/`tech.h`,含 `ResearchSelectWindow`/`ResearchListWindow`/`TechListWidget`)。
目標:比照百科檢視器(`docs/tech/cjk-screen-rendering.md`)的 CJK 樣板(Registry + `uifont.Wrap` + 自繪面板),
盤點這兩類畫面的文字元素、座標、TSV 來源與現有中文譯文,供下一輪實作對照。

**方法論說明**:`misc.tsv` 是英文文字為 key 的顯示層覆蓋表,不含原始 `str_id`。為了把
`BILL_INFO_*`/`BILL2_TECHGROUP_*`/`BILL_RESEARCH_*` 這些數字常數準確對回英文原文,本文用專案既有的
`internal/lbx` 套件寫了一支一次性 dump 工具(比照 `cmd/lbxstrings`,但改用引擎真實的 `str_id×6` 語言交錯
stride,見 `openorion2/src/lbx.cpp:253-292` 的 `loadFile`),對 scratchpad 內玩家自備的
`BILLTEXT.LBX`/`BILLTEX2.LBX` 做 `str_id → 英文原文` 的實測(非猜測排序)。dump 工具只在本次分析暫用,
未留在 repo 內。以下座標/常數/英文原文全部來自實際原始碼與這份 dump 結果交叉核對。

## 全域缺口總覽(先講重點)

openorion2 對這兩類畫面的實作程度落差很大,先列出來,細節見各節:

| 子畫面 | 完成度 | 說明 |
|---|---|---|
| `RaceInfoWidget`(種族統計) | 完整 | 唯一內容完整的 `InfoView` 分頁 |
| `TechReviewWidget`(科技總覽) | 完整 | 清單 + 說明 + 插圖都有接線 |
| `HistoryGraphWidget`(歷史曲線圖) | **只有標題** | 只畫標題文字 + 兩個空框,無曲線圖繪製邏輯 |
| `TurnSummaryWidget`(回合摘要) | **只有標題** | 只畫標題文字 + 一個空框,無回合報告內容 |
| `DocsWidget`(參考資料) | **只有標題** | 只畫標題文字 + 四個空框,未接百科內容 |
| `ResearchSelectWindow`(研究領域選單) | 完整 | 8 大領域按鈕 + 各領域下一項可選科技 |
| `ResearchListWindow`(單領域科技清單) | 完整 | 分頁清單 + 標題 |
| `MessageBoxWindow(tech,cost)`(科技詳情彈窗) | 完整 | 右鍵/點選科技彈出的說明+研究成本 |

`BILLTEXT.LBX` 實際還存有 `str_id 10-13`(`Categories`/`How to?`/`Category:`/`How to:`,疑似給
`DocsWidget` 用)、`str_id 26-60`(`MESSAGE SUMMARY AS OF SD: `/`Net Income: `/`NO ACTIVE SPIES`/
外交關係詞 `FEUD`…`HARMONY`,疑似給 `TurnSummaryWidget` 用)、`str_id 64`(`BASIC`)、`str_id 73`
(`OTHER`)——這些字串**原始遊戲資料裡真的存在**,但目前 openorion2 完全沒有任何 call site 引用(已用
`grep -rn "TXT_MISC_BILLTEXT"` 對全部 `.cpp` 逐一核對,見下方各節),所以**尚未翻進 `misc.tsv`**
(靜態溯源:不翻引擎不顯示的死字串)。等這三個分頁補實作時,是很明確的下一步字串來源。

## 共用元件與座標基準

- `InfoView`/`TechReviewWidget`/`RaceInfoWidget` 等分頁 Widget 的 `redraw(x, y, curtick)` 收到的
  `(x, y)` 是面板原點,實際畫面座標 = 面板原點 + 分頁 Widget 自身偏移(`_panels[i]` 建構時吃
  `(206, 0, SCREEN_WIDTH-206, SCREEN_HEIGHT)`,見 `info.cpp:1050-1063`)。本文表格內的座標一律是
  Widget `redraw()` 內部相對於自己 `(x, y)` 的偏移值(即程式碼字面座標),需要換算成螢幕絕對座標時
  再加 `206`(x 方向的分頁區起點)。
- 標題文字固定用 `FONTSIZE_TITLE`(`gfx.h:35`,值 5)+ `TITLE_COLOR_INFO`(`gfx.h:44`,值 2),
  `centerText(x+208, y+31, ..., OUTLINE_NONE, 3)`,五個分頁標題共用同一組座標/字型/顏色
  (`info.cpp:519-521,751-753,856-858,962-965,991-993`)。

## 1. `HistoryGraphWidget`(歷史曲線圖)

`info.cpp:507-525`。

| 元素 | 英文/來源常數 | 座標(相對面板原點) | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 標題 | `misctext(BILLTEXT, BILL_INFO_TITLE_HISTORY)` = `BILL_INFO_TITLE_HISTORY=5`(`lang.h:23`)→ `"HISTORY GRAPH"` | `centerText(x+208, y+31)`,`info.cpp:519-521` | `misc.tsv` | 歷史曲線圖 |
| 空框 1(無文字) | — | `drawInfoBox(x+14, y+60, 390, 63)`,`info.cpp:523` | — | — |
| 空框 2(無文字) | — | `drawInfoBox(x+14, y+132, 390, 282)`,`info.cpp:524` | — | — |

**缺口**:曲線圖本體(玩家歷史勢力值走勢)完全沒有繪製邏輯,`_game`/`_activePlayer` 成員在
`redraw()` 裡未被使用。中文化這個分頁目前只需處理標題,曲線圖是獨立的未來工作項。

## 2. `TechReviewWidget`(科技總覽)

`info.cpp:527-772`。

| 元素 | 英文/來源常數 | 座標 | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 標題 | `BILL_INFO_TITLE_TECH=6` → `"TECH REVIEW"` | `centerText(x+208, y+31)`,`info.cpp:751-753` | `misc.tsv` | 科技總覽 |
| 上/下翻頁按鈕 | 純圖片精靈(`ASSET_INFO_UP_BUTTON`/`ASSET_INFO_DOWN_BUTTON`,`info.cpp:571-586`) | `(204,65)`/`(204,395)` | — | 無文字,不需翻譯 |
| 4 個分類頁籤按鈕 | 圖片精靈 `ASSET_INFO_ACHIEVEMENTS_BUTTON+i`(`info.cpp:593-599`) | 依序排列 x 累加,y=0 | **待查證** | 是否烘焙文字未逐一截圖確認,建議實測畫面截圖再判斷是否需要中文替換圖 |
| 科技分組清單(`TechListWidget`,26 種分組) | `misctext(BILLTEX2, review_groups[i][j].title_id)`,`info.cpp:633-636`;`title_id` 為 `BILL2_TECHGROUP_*`(`lang.h:41-66`) | `TechListWidget` 建於 `(15,62,203,350)`(`info.cpp:601-602`);組標題實際繪製於 `fitText(x+2, y, ...)`(`tech.cpp:654-655`) | `misc.tsv`(BILLTEX2 26 條,見下表) | 見下表 |
| 科技項目名稱 | `techname(TNAME_TECH_NONE+tech_id)`(`info.cpp:624-625`) | `fitText(x+12, ypos, ...)`(`tech.cpp:668-669`) | `tech.tsv`(非 misc) | 已完成 |
| 選中科技說明標題 | `gameLang->help(tech_id)->title`(`info.cpp:687-688`) | `TextLayout::appendText(...,0,0,173,ALIGN_CENTER)`,面板於 `(x+227,y+60,180,45)` | `help.tsv`(非 misc,復用百科檢視器譯文) | 已完成 |
| 選中科技說明本文 | `gameLang->help(tech_id)->text`(`info.cpp:691`) | `appendText(...,0,196,173)`,面板 `(x+227,y+264,180,150)` | `help.tsv` | 已完成 |
| 科技插圖 | `GuiSprite(TECHPIC_ARCHIVE, tech_id,...)`(`info.cpp:693`) | `(x+227,y+115)` | — | 圖片,不需翻譯 |

### `BILL2_TECHGROUP_*` 26 條分組標題(`misc.tsv` BILLTEX2 段;`str_id` 由本輪 dump 實測取得)

| `str_id` | 常數(`lang.h`) | 英文原文 | `misc.tsv` 中文 |
|---|---|---|---|
| 0 | `BILL2_TECHGROUP_MISC` | Miscellaneous | 雜項 |
| 1 | `BILL2_TECHGROUP_CONSTRUCTION` | New Construction Types | 新建造類型 |
| 2 | `BILL2_TECHGROUP_SPIES` | Spies | 間諜 |
| 3 | `BILL2_TECHGROUP_COLONY` | Colony | 殖民 |
| 4 | `BILL2_TECHGROUP_GROUND_COMBAT` | Ground Combat | 地面戰鬥 |
| 5 | `BILL2_TECHGROUP_SHIP_EQUIP` | Ship Equipment | 艦艇裝備 |
| 6 | `BILL2_TECHGROUP_FOOD` | Food | 食物 |
| 7 | `BILL2_TECHGROUP_POLLUTION` | Pollution | 污染 |
| 8 | `BILL2_TECHGROUP_MORALE` | Morale | 士氣 |
| 9 | `BILL2_TECHGROUP_PRODUCTION` | Production | 生產 |
| 10 | `BILL2_TECHGROUP_COLONY_DEFENSE` | Defense | 防禦 |
| 11 | `BILL2_TECHGROUP_RESEARCH` | Research | 研究 |
| 12 | `BILL2_TECHGROUP_MONEY` | Money | 金錢 |
| 13 | `BILL2_TECHGROUP_BEAMS` | Beams | 光束武器 |
| 14 | `BILL2_TECHGROUP_MISSILES` | Missiles/Torpedoes | 飛彈/魚雷 |
| 15 | `BILL2_TECHGROUP_BOMBS` | Bombs/Biological | 炸彈/生化 |
| 16 | `BILL2_TECHGROUP_FIGHTERS` | Fighters | 戰機 |
| 17 | `BILL2_TECHGROUP_SPECIAL` | Special | 特殊 |
| 18 | `BILL2_TECHGROUP_SHIELDS` | Shields | 護盾 |
| 19 | `BILL2_TECHGROUP_ARMOR` | Armor | 裝甲 |
| 20 | `BILL2_TECHGROUP_DRIVES` | Drives | 引擎 |
| 21 | `BILL2_TECHGROUP_COMPUTERS` | Computers | 電腦 |
| 22 | `BILL2_TECHGROUP_FUELS` | Fuels/Range | 燃料/航程 |
| 23 | `BILL2_TECHGROUP_SHIP_DEFENSE` | Defense | 防禦(與 id10 同英文原文,不同科技分類,per-source 查詢不衝突) |
| 24 | `BILL2_TECHGROUP_SCANNERS` | Scanners | 掃描器 |
| 25 | `BILL2_TECHGROUP_OFFENSE` | Offense | 攻擊 |

> 註:`id10`(殖民防禦)與 `id23`(艦艇防禦)原文都是 `"Defense"`,`misc.tsv` 用英文字當 key 只會存一列
> `Defense→防禦`,兩處 `misctext()` call 都查到同一列——語意上兩個分組都叫「防禦」在遊戲原文裡本就同名,
> 非翻譯層的 bug。

## 3. `RaceInfoWidget`(種族統計,唯一完整分頁)

`info.cpp:774-938`。

| 元素 | 英文/來源常數 | 座標公式(`i`=0..3 為 2×2 方格索引) | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 標題 | `BILL_INFO_TITLE_RACE=7` → `"RACE STATISTICS"` | `centerText(x+208,y+31)`,`info.cpp:856-858` | `misc.tsv` | 種族統計 |
| 方格外框 | — | `xpos=x+(i%2?213:17)`,`ypos=y+(i/2?243:60)`,`drawInfoBox(xpos,ypos,188,171)`,`info.cpp:860-865` | — | — |
| 無接觸提示 | `BILL_INFO_NO_CONTACT=15` → `"NO CONTACT"` | `raceFnt->centerText(xpos+93,ypos+4)`,`info.cpp:870-873` | `misc.tsv` | 尚未接觸 |
| 種族名(大寫) | `pptr->race`(來自種族資料,非 misctext) | `raceFnt->centerText(xpos+93,ypos+4)`,`info.cpp:877-881` | `races.tsv`(非 misc) | 已完成;**注意**:`buf.toUpper()`(`info.cpp:878`)對 CJK 字串是 no-op 但仍會跑,移植時需確認 Go 端 `strings.ToUpper` 對中文不產生副作用(應無影響,純英文邏輯) |
| 滅亡標記 | `BILL_INFO_ELIMINATED=14` → `"( E L I M I N A T E D )"` | `itemFnt->centerText(xpos+93,ypos+17, OUTLINE_FULL)`,`info.cpp:884-888` | `misc.tsv` | (已滅亡) |
| 政體 | `estrings(ESTR_GOVERNMENT_NAMES + traits[TRAIT_GOVERNMENT])`,模板 `"^ %s"` | `itemFnt->renderText(xpos+3,ypos)`,`info.cpp:892-896` | `estrings.tsv`(非 misc) | 已完成(政體制,如「聯邦制」) |
| 31 項種族特性(`TRAITS_COUNT=31`,`gamestate.h:54`) | `raceInfo(j)`,模板 `"^ %s"` + 數值後綴(見下) | 逐行 `renderText(xpos+3,ypos)`,`ypos` 累加 `itemFnt->height()+1`,`info.cpp:899-927` | `raceinfo.tsv`(非 misc) | 已完成 |
| 貧礦母星特例 | `raceInfo(TRAIT_POOR_HOMEWORLD)` | 承上一行 `ypos`,`info.cpp:929-934` | `raceinfo.tsv` | 已完成 |

**動態數值組裝(種族特性行)**(`info.cpp:906-922`):

```
buf = "^ " + raceInfo(j)                          // 特性名稱
if j == TRAIT_FARMING || j == TRAIT_MONEY:
    fmt = (traits[j] 為奇數) ? "%+3.1f" : "%+2.0f"  // 半值特性,可能有 .5
    buf += fmt(traits[j] / 2.0)
    if j == TRAIT_MONEY: buf += " BC"               // 單位不翻譯(BC=遊戲貨幣代碼)
elif j < TRAIT_LOW_G:
    buf += "%+d"(traits[j])                          // 一般數值特性帶正負號
# j >= TRAIT_LOW_G 的特性(重力/棲息地/種族天賦等)只有名稱,無數值後綴
```

`" BC"` 是貨幣代碼字面量,不經過 TSV,建議中文化維持原樣(和 `RP`〔Research Points〕、`EP`〔Experience
Points〕同類,MOO2 原文慣例保留英文縮寫)。

## 4. `TurnSummaryWidget`(回合摘要)

`info.cpp:940-968`。

| 元素 | 英文/來源常數 | 座標 | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 標題 | `BILL_INFO_TITLE_TURN_SUMMARY=8` → `"TURN SUMMARY"` | `centerText(x+208,y+31)`,`info.cpp:962-965` | `misc.tsv` | 回合摘要 |
| 空框(無文字) | — | `drawInfoBox(x+12,y+60,395,355)`,`info.cpp:967` | — | — |

**缺口**:整個回合報告(訊息摘要、外交關係列表)未實作。`BILLTEXT.LBX` 內 `str_id 26-60` 明顯是給這個
畫面用的資料(本輪 dump 實測):

| `str_id` | 英文原文 | 推測用途 |
|---|---|---|
| 26 | `MESSAGE SUMMARY AS OF SD: \x88` | 標題模板(`\x88`=stardate 佔位符) |
| 27 | `Net Income: \x82 BC` | 淨收入模板(`\x82`=數值佔位符) |
| 28 | `NO ACTIVE SPIES` | 無派駐間諜提示 |
| 29-30 | `NONE`(×2) | 通用「無」 |
| 31 | `ALLIANCES:` | 同盟列表標題 |
| 32 | `WARS:` | 戰爭列表標題 |
| 33-49 | `FEUD`/`HATE`/`DISCORD`/…/`HARMONY` | 17 級外交關係形容詞(由差到好) |
| 50 | `(IGNORED)` | 忽略標記 |
| 57-58 | `RESEARCH TREATY:`/`TRADE TREATY:` | 條約標籤 |
| 59-60 | `GIVING \x82% TRIBUTE`/`RECEIVING \x82% TRIBUTE` | 進貢關係模板 |

這些字串**目前沒有任何 call site**(已用 `grep -rn "TXT_MISC_BILLTEXT" openorion2/src/*.cpp` 核對,
命中僅 `info.cpp`/`tech.cpp`/`guimisc.cpp` 既有 8 處,不含此範圍),尚未進 `misc.tsv`。等
`TurnSummaryWidget` 補實作時,這張表可以直接當翻譯清單使用,不用重新 dump。

## 5. `DocsWidget`(參考資料)

`info.cpp:970-999`。

| 元素 | 英文/來源常數 | 座標 | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 標題 | `BILL_INFO_TITLE_REFERENCE=9` → `"REFERENCE"` | `centerText(x+208,y+31)`,`info.cpp:991-993` | `misc.tsv` | 參考資料 |
| 空框 ×4(無文字) | — | `(x+12,y+60,197,27)`/`(x+12,y+93,197,321)`/`(x+212,y+60,196,27)`/`(x+212,y+93,196,321)`,`info.cpp:995-998` | — | — |

**缺口**:這個分頁完全沒有接上任何百科內容(左右各一個「分類欄」+「內容欄」的框架,推測是原版
「依分類瀏覽百科」介面,而非本專案已完成的單條百科檢視器)。`BILLTEXT.LBX` 的 `str_id 10-13`
(本輪 dump 實測)疑似正是這個分頁的介面文字:

| `str_id` | 英文原文 | 推測用途 |
|---|---|---|
| 10 | `Categories` | 分類欄標題 |
| 11 | `How to?` | 操作說明分類(頁籤?) |
| 12 | `Category:` | 目前選取分類標籤 |
| 13 | `How to:` | 操作說明標籤 |

同樣**尚無 call site**,未進 `misc.tsv`。此分頁補實作時,內容顯示邏輯應直接復用
`docs/tech/cjk-screen-rendering.md` 記錄的百科檢視器全鏈路(`help.tsv` + `Font.Wrap`),差異只在於
外框改成「左側分類清單 + 右側內容」的雙欄版面,而非單條全螢幕。

## 6. `TechListWidget`(科技總覽/研究清單共用元件)

`tech.cpp:307-688`,同時被 `TechReviewWidget`(§2)與 `ResearchListWindow`(§8)使用,是兩張畫面
真正共用的清單渲染邏輯。

| 元素 | 內容來源 | 座標 | TSV 來源 |
|---|---|---|---|
| 分組標題 | 呼叫端傳入的 `title`(見 §2/§8) | `fitText(x+2, y, maxw-4, _titleFont, color, title, OUTLINE_NONE, 2)`,`tech.cpp:654-655` | 依呼叫端 |
| 分組內科技項目 | 呼叫端傳入的 `TechListItem[]`(科技名稱字串) | `fitText(x+12, ypos, maxw-14, _itemFont, color, name, OUTLINE_NONE, 2)`,`tech.cpp:668-669` | 依呼叫端(`tech.tsv`) |

`fitText` 會依可用寬度自動縮字距(呼叫端傳入 `spacing=2` 作為上限),這點與 `uifont.WrapText` 的
CJK 逐字斷行策略不同——`fitText` 是**單行縮排擠壓**,不是換行。中文字通常比原始拉丁字寬,若直接沿用
這條擠壓邏輯可能造成中文標題被過度壓縮到難以辨識,移植時建議改用真正的換行或縮小字級,而非壓縮字距。

## 7. `ResearchSelectWindow`(研究:8 大領域選單,彈窗)

`tech.cpp:920-1020`,由 `GuiWindow` 彈出(非 `InfoView` 分頁,是點選「選擇研究方向」時跳出的獨立視窗)。

| 元素 | 英文/來源常數 | 座標 | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 8 個領域按鈕 | 圖片精靈 `ASSET_TECHSEL_AREA_BUTTONS+off+i`(`tech.cpp:956-965`) | `x = i%2 ? 248:21`,`y = ypos[i/2]-img->height()`,`ypos[]={45,149,254,361}` | — | **烘焙點陣圖文字,非 TSV 可翻**(見下方警示) |
| 各領域「下一項可選科技」摘要(`ResearchSelectWidget`) | 領域標題:`techname(TNAME_RTOPIC_STARTING_TECH+_topic)` 或超科技 `estrings(ESTR_RTOPIC_HYPER)`(`tech.cpp:836-840`) | `titlefnt->renderText(x+2,y+3,...)`,`tech.cpp:846` | `tech.tsv`/`estrings.tsv`(非 misc) | 已完成 |
| 可選科技清單(每項最多 4 選 1) | `techname(TNAME_TECH_NONE+_choices[i])`;純研究則 `misctext(BILLTEXT, BILL_PURE_RESEARCH)` | `fnt->renderText(x+12,y,...)`,`tech.cpp:914` | `tech.tsv` / `misc.tsv`(純研究="Pure research"→純研究) | 已完成 |
| 研究成本標籤 | `LabelWidget` 純數字模板 `"%u RP"`(`tech.cpp:1002`) | `_costLabels[i]`,`(i%2?368:141, ypos[i/2]-13, 86, 14)` | — | 數字模板,`RP` 不譯 |

**按鈕中文化警示(對應專案 CLAUDE.md「按鈕的中文化一定要參考先前的中文化經驗」)**:
`ASSET_TECHSEL_AREA_BUTTONS` 這組是 MOO2 原版**烘焙點陣圖按鈕**(領域名稱畫在點陣圖裡,不是文字層),
和 `master-of-orion`(SDL2 版)/`master-of-magic` 遇過的「按鈕貼圖含英文字」是同一類陷阱。這組按鈕
**不能靠 `misc.tsv` 翻譯**——`misctext()` 全域都沒有引用這 8 個按鈕的文字(它們是圖,不是
`misctext()` 呼叫),必須比照先前專案的作法:換成中文重繪的按鈕圖,或改成「透明/單色底圖 + 疊加
`uifont` 動態渲染的中文文字層」。`InfoView` 五個分頁切換按鈕(`ASSET_INFO_HISTORY_BUTTON+i`,
`info.cpp:1037-1042`)以及 `TechReviewWidget` 的 4 個分類頁籤(`ASSET_INFO_ACHIEVEMENTS_BUTTON+i`,
`info.cpp:594-599`)同樣是圖片精靈,尚未逐一截圖確認是否烘焙文字,但因為和研究領域按鈕来自同一套
UI 慣例,應預設當作「疑似烘焙文字」處理,實作前先截圖確認。

## 8. `ResearchListWindow`(研究:單領域科技樹清單,彈窗)

`tech.cpp:1022-1239`。

| 元素 | 英文/來源常數 | 座標 | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 標題(領域名) | `misctext(BILLTEXT, BILL_RESEARCH_TOPIC_BIOLOGY + _area)`;`BILL_RESEARCH_TOPIC_BIOLOGY=65`(`lang.h:39`),`_area` 為 `ResearchArea` enum(`gamestate.h:229-238`,0=BIOLOGY..7=SOCIOLOGY) | `fnt->renderText(x, _y+16, ...)`,`tech.cpp:1233-1234` | `misc.tsv` | 見下表 |
| 標題尾碼「清單」 | `misctext(BILLTEXT, BILL_RESEARCH_TOPIC_LIST)`;`BILL_RESEARCH_TOPIC_LIST=63` → `" LIST"` | 緊接標題後 `renderText`,`x` 取前一次呼叫回傳值,`tech.cpp:1229,1235-1236` | `misc.tsv` | 清單(前導空白) |
| 科技分組清單 | 見 §6 `TechListWidget`;分組標題 = `techname(TNAME_RTOPIC_STARTING_TECH+topic)`(`tech.cpp:1130-1132`) | `TechListWidget` 建於 `(14,40,228,354)`,`tech.cpp:1052-1053` | `tech.tsv`(非 misc) | 已完成 |
| 超科技分組 | `estrings(ESTR_RTOPIC_HYPER)`(`tech.cpp:1162-1163`) | 同上清單末尾 | `estrings.tsv` | 已完成 |
| 上/下翻頁、返回按鈕 | 圖片精靈(`ASSET_TECHLIST_UP/DOWN/RETURN_BUTTON`,`tech.cpp:1062-1076`) | `(247,46)`/`(248,367)`/`(187,403)` | — | 待截圖確認是否烘焙文字 |

### `BILL_RESEARCH_TOPIC_BIOLOGY+area` 8 條領域標題(`str_id 65-72`,本輪 dump 實測,與 `ResearchArea`
enum 順序完全對齊)

| `_area` | 常數 | 英文原文 | `misc.tsv` 中文 |
|---|---|---|---|
| 0 | `RESEARCH_BIOLOGY` | BIOLOGY | 生物學 |
| 1 | `RESEARCH_POWER` | POWER | 能源 |
| 2 | `RESEARCH_PHYSICS` | PHYSICS | 物理學 |
| 3 | `RESEARCH_CONSTRUCTION` | CONSTRUCTION | 建造學 |
| 4 | `RESEARCH_FIELDS` | FORCE FIELDS | 力場學 |
| 5 | `RESEARCH_CHEMISTRY` | CHEMISTRY | 化學 |
| 6 | `RESEARCH_COMPUTERS` | COMPUTERS | 電腦學 |
| 7 | `RESEARCH_SOCIOLOGY` | SOCIOLOGY | 社會學 |

`str_id 64`(`"BASIC"`→基礎)實測存在於 `BILLTEXT.LBX`,但 `_area` 範圍固定 0-7 對到 `str_id 65-72`,
**用不到 64**,目前無任何 call site 引用——同屬「資料存在但引擎未接線」的死字串,不建議投機翻譯到
`misc.tsv`(已翻,但屬於超出目前接線範圍的項目,标記存查即可)。

## 9. `MessageBoxWindow(tech, cost)`(科技詳情彈窗,含研究成本)

`guimisc.cpp:110-131`,由 `ResearchSelectWidget::handleMouseUp`(右鍵,`tech.cpp:805-807`)與
`ResearchListWindow::showTechHelp`(點選檢視,`tech.cpp:1180-1195`)觸發,是研究相關畫面裡「查看某科技
詳情」的共用彈窗。

| 元素 | 英文/來源常數 | 動態組裝 | TSV 來源 | 對應中文 |
|---|---|---|---|---|
| 科技標題 | `gameLang->help(tech)->title`(`guimisc.cpp:117-119`) | — | `help.tsv` | 已完成 |
| 科技描述本文 | `gameLang->help(tech)->text`(`guimisc.cpp:120-122`) | — | `help.tsv` | 已完成 |
| 研究成本行 | `misctext(BILLTEXT, BILL_RESEARCH_COST)`;`BILL_RESEARCH_COST=61` → `"Research cost:"` | `buf.printf("%s%u RP", str, cost)`(`guimisc.cpp:123-124`),即「標籤 + 數字 + `" RP"`」字面相接、**無空白分隔**(原文本身無尾空白,``研究成本:150 RP`` 這種黏在一起的排法是原版設計,非 bug) | `misc.tsv` | 研究成本: |

## 動態數值組裝總表(`printf` 模板 + 佔位符)

| 畫面 | 模板 | 佔位符 | 位置 |
|---|---|---|---|
| `InfoView` 星曆日期 | `"%u.%u"` | `stardate/10`, `stardate%10` | `info.cpp:1117` |
| `InfoView` BC 收入 | `"%u%s"` | 收入數字 + `misc` 字串 `" BC INCOME"` | `info.cpp:1130` |
| `InfoView` 維護費百分比 | `"%u%%"` | 百分比數字(另外一次 `renderText` 畫分類名,非同一模板) | `info.cpp:1138` |
| `RaceInfoWidget` 特性行 | `"^ %s"` + `"%+3.1f"`/`"%+2.0f"`/`"%+d"` | 特性名稱 + 數值(規則見 §3) | `info.cpp:894-921` |
| `ResearchSelectWindow` 研究成本標籤 | `"%u RP"` | 成本數字 | `tech.cpp:1002` |
| `MessageBoxWindow` 研究成本行 | `"%s%u RP"` | `misc` 字串「Research cost:」+ 成本數字 | `guimisc.cpp:124` |
| `ResearchListWindow` 標題 | 兩次獨立 `renderText`(非單一模板) | 領域名 + `" LIST"` | `tech.cpp:1233-1236` |

以上模板都是「英文字面 + 數字/單位」直接字串相接,**不是** `fmt.Sprintf` 風格的具名佔位符
(`%s` 一律代表某段已翻譯好的中文短字串)。中文化時只要 `misc.tsv` 對應行翻對,`printf` 組裝邏輯本身
不需更動;唯一要注意的是中文語序——例如「研究成本:150 RP」中文語序和英文一致(標籤在前數字在後),
不需要額外處理位置互換。

## 統計

- 分析畫面:2 大類 9 個渲染單元(`InfoView` 5 分頁 + `ResearchSelectWindow`/`ResearchListWindow`/
  `TechListWidget`/`MessageBoxWindow(tech,cost)` 4 個研究相關畫面)。
- 抓到文字元素(已接線、實際會被 `misctext()`/`techname()`/`estrings()`/`raceInfo()`/`help()` 呼叫繪製者):
  約 40 處渲染呼叫點(不含科技/特性等清單型「每項一次」的迴圈渲染,那些各自對應既有 `tech.tsv`/
  `raceinfo.tsv`/`estrings.tsv`/`help.tsv`)。
- 能對到既有 `misc.tsv` 譯文:27 個 `misctext()` 呼叫點,涵蓋 5 個標題 + `ELIMINATED`/`NO CONTACT` +
  收入/維護費 7 類標籤 + 研究成本/純研究/清單尾碼 + 8 大研究領域標題 + 26 條科技分組標題,合計對應
  `misc.tsv` 中 **27 個唯一英文 key**(BILLTEXT 19 個 + BILLTEX2 全部 26 個減去與 BILLTEXT 無重複部分;
  精確唯一英文字串數見各節表格)。
- 額外發現但**尚未進 `misc.tsv`**、屬於未來分頁實作缺口的字串:`str_id 10-13`(參考資料介面,4 條)、
  `str_id 26-60`(回合摘要介面,約 30 條,含 17 級外交關係形容詞)、`str_id 64`(`BASIC`,已翻但未接線)、
  `str_id 73`(`OTHER`,已翻但未接線)。

## 用 CJK 樣板實作此畫面的步驟建議

沿用 `docs/tech/cjk-screen-rendering.md` 的三塊基建(`i18n.Registry` + `uifont.Wrap`/`fitText` 等價物
+ 自繪面板),建議順序:

1. **先做 `RaceInfoWidget` 與 `TechReviewWidget`**——這兩個是唯一內容完整的分頁,英文/座標/TSV
   對照已在 §2/§3 列全,不需要額外補資料,可以直接照抄百科檢視器的 `Registry.Source("misc")` +
   `Registry.Source("tech")`/`Registry.Source("help")`/`Registry.Source("estrings")`/
   `Registry.Source("raceinfo")` 多來源查詢模式(呼應 `i18n-catalog-architecture.md` 的
   per-source catalog 架構,這兩個畫面剛好是「單畫面多來源」的典型案例)。
2. **`TechListWidget` 換行策略需重新設計**,不要照搬原版 `fitText` 的縮字距擠壓(§6 已說明中文字寬
   問題),改用 `uifont.WrapText` 或依可用寬度縮小字級,並對超長中文科技分組標題(如「新建造類型」
   對比原文 "New Construction Types")做寬度實測。
3. **研究彈窗(`ResearchSelectWindow`)的 8 個領域按鈕是烘焙點陣圖**(§7 警示),移植前必須先截圖
   原版素材確認,若確認烘焙文字,採用「素材去文字化 + `uifont` 疊加中文」的既有解法(比照
   `master-of-orion`/`master-of-magic` 先前專案經驗),不要嘗試用 `misc.tsv` 翻譯不存在的字串 key。
4. **`InfoView` 五個分頁切換按鈕、`TechReviewWidget` 4 個分類頁籤、`ResearchListWindow` 的翻頁/返回
   按鈕**,建議實作前先用 `cmd/lbxdump`(或既有的 headless 截圖 harness)截圖比對,一次性判定
   是否烘焙文字,避免逐一按鈕臨場判斷造成不一致。
5. **`HistoryGraphWidget`/`TurnSummaryWidget`/`DocsWidget` 三個空畫面**,若要補實作內容,§4/§5 已把
   `BILLTEXT.LBX` 裡對應的候選字串(`str_id 10-13`、`26-60`)整理成表,可以直接排進下一輪翻譯
   worklist,不需要重新 dump LBX。`DocsWidget` 的內容顯示邏輯建議直接復用百科檢視器已驗證的
   `help.tsv` + `Font.Wrap` 全鏈路,只需另外處理雙欄(分類清單+內容)版面。
6. 沿用百科檢視器的 headless 截圖驗證流程(`-shot out.png -frames N`),每完成一個分頁就截圖比對
   座標與換行是否與本文表格一致。
