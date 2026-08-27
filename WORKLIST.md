# WORKLIST — 銀河霸主2 go/ebiten 重製 + 繁中化

> ⚠ **勾選狀態會過期,以程式碼為準**(rule 63)。這份清單橫跨數十輪,底下 Phase 0–4 有不少
> `[ ]` 其實早就做完了(2026-08-07 已修掉一批可核實的)。要判斷「某項做了沒」,
> **先 grep 程式碼,別信這裡的方框**。各輪的實際成果以 `docs/re/01-gap-report.md`
> 與 `git log` 為準。

> ⚠ **百分比只作有日期、可重算的工程儀表板。** 依 2026-08-25 使用者決策，README 分開呈現
> remake 功能、原版玩法對齊與發行驗證三項比例；不合成單一「還原度」，也不把測試綠升格為
> 原版逐幀／逐位元一致。現況仍以下方活表與 [`docs/re/parity-matrix.tsv`](docs/re/parity-matrix.tsv)
> 為準，定性邊界見 [`docs/HONEST-STATUS.md`](docs/HONEST-STATUS.md)。
> 下方大量 `[x]` 中,gameplay 子系統類仍須以「對原版實測」重評數值對齊度(rulebook 65),不因測試綠自評還原度。
>
> **本週收尾快照（2026-08-27）**：依本段唯一活表重新核算，40 項中 26 項完成、
> 7 項進行中、7 項待辦；完成計 1、進行中計 0.5，remake 功能完成度為 **74%**。
> 原版玩法對齊度 **49%** 與發行驗證完成度 **33%** 的分母及公式見 `README.md`；
> 本週新增的文案外部化切片仍屬既有進行中母項，因此沒有造成分數跨級。
>
> 音樂、可操作的新遊戲流程、主要畫面與 gameplay 子系統都有 remake 實作，但「有實作」不等於
> 原版忠實完成。2026-08-12 IDA Pro 重新稽核在議會票數、議會排程與戰略轟炸找到直接反例；
> 議會人口票數、召開排程與戰略轟炸三外圈／結果尺度已於 2026-08-24 修正；
> 並確認外交／AI 等核心系統仍使用設計性重建。忠實度現況以
> [`docs/re/parity-re-audit-20260812.md`](docs/re/parity-re-audit-20260812.md) 與下方活表為準。
>
> 可勾選工作清單,對應 `PLAN.md` 階段。允許擴充(CLAUDE.md)。完整性優先:不預先砍項;卡關記錄方法換路,不寫「暫緩/低投報」。
> 圖例:`[ ]` 待辦 `[~]` 進行中 `[x]` 完成(⚠ 多為自製系統的完成,非原版對齊)。

## ★ 剩餘工作（2026-08-27 依 parity 矩陣、source 與 README 畫面重新核對）

> **這一段是唯一該看的剩餘工作表。**
>
> **維護方式**(這一段本身也會過期,所以規則寫在這裡):
> - 底下各 Phase 的每一個未完成方框,判定寫在**它自己那一行**——不在別處另寫一份摘要,
>   那正是斷言殘留的來源。**這裡也不複述「還剩幾項」**,那個數字每做一項就錯一次。
> - **做完一項就當場把那一行改對**,過期的斷言直接刪掉、只留正確結論;不劃刪節線也不留註記,
>   因為留著就會被讀到。要查某項當初怎麼錯的,看 `docs/re/01-gap-report.md` 的編號項
>   (那裡是工程日誌,錯誤的推導過程本身有價值)。
> - **「方向已改」與「漏做」要分開**:前者直接移除整條(留著會讓人以為架構有缺),
>   後者留在表上並寫清楚擋在什麼後面。
>
> ⚠ 本檔只有下面的「2026-08-24 盤點結論」是目前待辦來源；其後的附錄與 A–AO/B* 表是
> 證據／歷史紀錄，不要從其中的舊 `[ ]` 或「仍缺」敘述重新開工。

### 2026-08-27 盤點結論（目前唯一待辦來源）

> 範圍遵循使用者 2026-08-12 的新決策：重新以原版玩法忠實度為最高優先，採 IDA Pro
> 追查原版玩家機制，再對回 Go／Ebitengine 的實際消費端；仍不深挖與遊戲玩法無關的內部功能。
> `[ ]` 表示原版機制尚未形成可回查的垂直證據鏈，不代表目前基本 4X 對局不可玩。
> 原版證據留白與 remake 功能缺口分開記，避免把「可玩」誤寫成「已忠實重製」。
> Windows API／Win95 平台內部行為已依 2026-08-25 使用者決策設為 RE 停止線：只追回玩家可見的
> 輸入、輸出、錯誤與時序契約；平台內部呼叫以有標註的現代近似完成，不為 API 內部實作另開缺口。

#### 原版忠實度重新稽核（最高優先）

- [x] **2026-08-28 三個長跑回歸已分類並修正**：議會第二屆確實準時召開，但原版搖擺票重擲後
  流會，舊測試錯把「再次當選」當排程契約；AI 0 是 Creative，舊研究測試錯把合法的
  `ExplicitChoice` 空集合當成未研究；擴張測試則錯假設新殖民地短期資產必高於只造艦的單母星。
  改為直接驗證第二屆排程、非 Creative application 與新增殖民地 `ColonyOutput` 後，另發現真缺口：
  職務 RE 漏掉 `sub_D6AD4 → sub_D6A00` 追加農夫，導致 AI 人口 8→1。現已補該下游切片，
  三項定向長跑通過；`player+0x38` 精確食物運輸容量仍依 AI 職務規格列為 DRAFT，不冒稱完成。

- [~] **玩家可見文案外部化**：2026-08-26 已把既有 `assets/i18n/*.json` 與內嵌副本統一轉為
  有序 JSON 並移除 `go:embed` 副本，載入器保留 per-source、先出現者優先及原版單位元控制標記契約。剩餘工作是逐畫面
  移除 `tr(中文, 英文)` 與直接繪製的硬編文案，改以穩定鍵值查詢。多人資訊面板已完成第一個
  垂直切片：IDA 重新確認七個 MULTIGM 狀態共用 loader／draw 與按鈕 helper；`netinfo.go`
  現只保存 `netinfo.*` 語意鍵，中英文標題、狀態與按鈕均由 `assets/i18n/ui.json` 提供，並有
  靜態防回歸與雙語 catalog 測試。自訂種族亦完成第二個切片：正版 `RACESTUF.LBX` asset 0
  已推翻泛用 `None/Poor/Good/Great`，現依原版類別專屬數值字串顯示；規則分支只比較 typed ID，
  `customrace.go` 不再保存中英文玩家文案，並以 `textSafeRect` 約束標題、類別、選項、特殊能力、
  成本與按鈕。命名／旗色頁亦完成第三個切片：IDA 推翻 `Player_Name_ @ 0xE5E09` 是輸入畫面及
  `Request_Banner_Color_ @ 0xFBEE1` 是單機旗色選單的舊符號解釋；旗色玩法資料現只保留穩定鍵、
  索引與 RGB，標題、提示、按鈕、預設名及八色名稱皆由 `ui.json` 提供。原版兩畫面、真旗幟圖與
  `RACEOPT.LBX#0` 錯誤背景仍是獨立忠實度缺口，尚未完成。INFO 五個子畫面也已完成第四個切片：
  標題、動態模板、欄名、態勢／關係、回合摘要及 Reference 目錄均由 `ui.json` 供應，英文模式不再
  直接輸出 `shell` 保存的中文狀態，Reference 亦不再暴露專案內部路徑。原版已知 `BILLTEXT.LBX`
  標題與 remake 補充目錄已在 JSON 註記及 RE 稽核中分級；完整 Reference 內容與 Turn Summary
  動態組裝仍是原版證據留白，不把現行補充升格為精確還原。殖民地建造佇列已完成第五個切片：
  六顆按鈕、狀態訊息、格式模板、Housing／Trade Goods 顯示名皆由 `ui.json` 供應；直接字型繪製
  已改成雙軸安全框，舊 `y=470` 越界訊息移到佇列上方兩行區。IDA 另證實 `func_names.txt` 把
  `Draw_Build_Queue_Popup_` 錯放在 5-byte thunk `0xB3E75`，完整繪製函式實為
  `sub_B3CF7 @ 0xB3CF7..0xB3E75`；`0xB08CA`／`0xB094C` 則是兩個相鄰且都被呼叫的函式，
  後者精確語意維持未知。殖民地主畫面已完成第六個切片：按鈕、轉場、行星前綴、同化／人口、
  產出、建築與職業文案均改用 `ui.json`，所有文字走雙軸安全框；職業列依 IDA 的
  x=`310..510`、y=`62+30i` 保持原位並補足實際字墨高度，按鈕移除重複 inset。IDA 同時證實
  完整 `Colony_Screen_` caller 鏈位於 `sub_C058A @ 0xC058A..0xC0965`，而
  `func_names.txt` 的 `0xC0965` 是無 caller 的相鄰小函式。35 張中文畫廊已重跑；腳本先解除
  新局研究 application gate，正常第一回合存為回合摘要。原版一般事件至少第 50 回合才可能出現，
  事件圖改用外部 JSON 雙語固定戰報，只作版面驗證，不再把研究頁或第一回合誤標成事件。
  目視抽查事件、回合摘要、殖民地主畫面與最終得分，檔名與內容一致且文字未越框。
  事件／勘查快報亦完成第七個切片：固定台標、好壞／發現標記、標題格式、按鈕與轉場改由
  `ui.json` 供應，正文及按鈕均走雙軸安全框，且依原版接受整張 640×480 點擊。IDA 證實
  `sub_2031D` 啟動、`sub_20538` 載入 `EVENTS.LBX#0/#1`、`sub_20460` 繪製，事件插圖由
  相鄰 `sub_203CB` 取 `eventID+2`；`func_names.txt` 把四個名稱依序錯放到相鄰函式。
  正版 `EVENTS.LBX` 已交叉驗證為 38 個資產，現優先播放 `#1` 的 31 幀累積動畫並於
  `(320,14)` 疊 `#2..37` 的 36 張事件圖；缺檔、非法 ID 或勘查報告安全退回自繪面板。
  正版資產與缺資產 fallback 兩套中文畫廊皆為 35/35，已目視抽查 `05_event.png` 的插圖、
  台標、正文與按鈕安全框。動畫每 3 tick 一幀仍明標 remake timing approximation。
  最終得分／名人堂已完成第八個切片：九個分項改由規則層提供穩定 `TextKey`，勝敗標題、原因、
  摘要、提示與轉場均由 `ui.json` 供應，`hiscore.go` 不再內嵌玩家文案或直接呼叫字型繪製。
  IDA 證實 `sub_9EB42 @ 0x9EB42..0x9EC32` 載入 `SCORE.LBX#0..14`、
  `sub_9E27A @ 0x9E27A..0x9E3AD` 以 `#0` 畫 640×480 背景，並由 `sub_9D7EA`
  建立整張畫面輸入欄；舊外部符號把 loader 誤當完整畫面函式。remake 現驗證 `#0` 尺寸、
  frame 與內嵌調色盤後使用原版機械框，缺檔才退回 `TURNSUM.LBX` 自繪面板；繁中只覆蓋烘字
  文字牌，英文保留原圖。正版與缺 `SCORE.LBX` 兩套畫廊皆為 35/35，已目視抽查九列、總分及
  全畫面點擊提示均在內容框內；`#1..14` race icon 的 writer／consumer 尚未閉合，未任意接入。
  地面戰戰報已完成第九個切片：標題、攻守統計、三種戰果、按鈕、錯誤與轉場均由 `ui.json`
  供應，固定畫廊殖民地名亦不再藏於 `interactive.go`；各列、標題、戰果與按鈕改走雙軸安全框。
  IDA 同時推翻 `func_names.txt` 的 `Colony_Combat_Screen_ @ 0xB771D` 舊解釋：該函式只掃描
  每側 40 筆、stride `0x19` 的單位記錄；完整設定／生命週期是
  `sub_B7491 @ 0xB7491..0xB771D`，逐幀繪製與戰鬥推進是
  `sub_B7289 @ 0xB7289..0xB7491`。原版即時戰鬥完成後自動返回，remake 一次性解算後的
  定格與繼續按鈕維持明標近似。正版與缺 `COLGCBT.LBX` 兩套畫廊均為 35/35；人工抽查另發現
  單位洋紅色 ramp 未替換；IDA 進一步證實 `sub_B8EFB` 依 raw color 對
  `0xC0..0xC7`／`0xE8..0xEB` 做兩張 index 表替換，現已逐值串到 typed
  `AttackerColor`／`DefenderColor`，`#21` 面板框則保留原色。palette provider 的原版精確來源仍未知。
  軌道轟炸戰報已完成第十個切片：標題、四列戰果、按鈕、錯誤與轉場均由 `ui.json` 供應，
  生物武器戰果取代一般人口列而不再留下會產生第五列的過期註解；標題、四列與按鈕全部改走
  雙軸安全框。IDA 證實 `sub_B4D02` 是結果外層入口、`sub_B4800` 是逐幀繪製／推進回呼，
  `sub_B4606` 的炸彈記錄 stride 為 `0x0F` 且未啟用炸彈只在 `Random_(5)==1` 時落下。
  同時推翻 `sub_B8EFB` 可證明轟炸背景精確換色的舊註解：其直接 caller 全在地面戰 loader；
  現行 `COLONY.LBX#8` 紫色三階映射已改標 remake 視覺近似，並由 typed `DefenderColor` 使用
  被轟炸方色彩，不再錯用攻方旗色。`COLONY.LBX#8` 原版直接 loader／palette 鏈仍是非阻塞留白。
  載入／儲存遊戲彈窗已完成第十一個切片：槽位、錯誤、按鈕與轉場文案均由 `ui.json` 供應，
  名稱、星曆、時間、訊息與兩顆按鈕改走雙軸安全框。IDA 證實 `sub_7E154` 儲存流程與
  `sub_804B7` 主選單讀檔入口都共用 `sub_7F206` 繪製回呼；`sub_803D9` 依序載入
  `GAME.LBX#20..26`，槽位 stride `0x25`、列距 `0x1F` 且固定十列。舊 gap report 同時寫
  「已完成」與「只有載入」的矛盾敘述已清除。英文存檔模式現在會以 JSON 的 `SAVE` 覆蓋
  `#21` 的 `LOAD` 烘字，缺資產時兩顆按鈕亦有自繪 fallback，不再消失。英文畫廊另實際抓到
  未指定外部字型時 `sceneBuilder.fnt=nil` 會在高解析錄→放 panic；`loadFont` 現改用內建點陣字
  作英文動態文案 fallback，原版烘字是否露出仍由各畫面控制，不再由全域 nil 字型決定。
  正版繁中、缺 `GAME.LBX` fallback 與正版英文畫廊均完成 35 張；已目視抽查 `18_loadgame.png`，
  三欄文字未侵入圖示，按鈕中心一致，過長日期依規格顯示省略號而不越框。
  指揮點數摘要已完成第十二個切片：標題、六個欄位與關閉提示全部改由
  `ui.json` 供應，欄名／數值分欄、懲罰列與關閉列均經雙軸安全框繪製。IDA 證實
  `sub_8BAB9` 直接呼叫的是 `sub_E2644 @ 0xE2644..0xE2671`，再由它呼叫
  `sub_E2000`；舊文件所稱 `sub_DDF24` 獨立泛用視窗函式並不存在，raw `0xDDF24`
  是 `sub_DDEFB` 內的尾端 call site。同輪修正畫面把每點 10 BC 超額懲罰誤顯示為
  1 BC 的錯誤，現改讀 `gamedata.IncomeCommandOverflowCostPerPoint`，與引擎消費端共用契約。
  正版繁中、正版英文與缺 `GAME.LBX` fallback 畫廊均完成 35 張；已目視抽查
  `27_commandpoints.png`，三者標題、分欄數值與關閉提示都在面板內。
  對局內遊戲選單已完成第十三個切片：六顆按鈕、音量標籤、遷移連線狀態、無存檔
  訊息與轉場名均改由 `ui.json` 供應；按鈕、設定列與訊息列經雙軸安全框繪製。
  英文正版資產保留按鈕烘字，繁中擦底疊字；缺 `GAME.LBX` 時則兩種語言都有自繪
  視窗、按鈕、音量標籤與可用熱區。IDA 證實 `_Game_Popup_ @ 0x8012F` 四路分流、
  `Do_Main_Game_Popup_ @ 0x7DD41`、`_Draw_Main_Game_Popup_ @ 0x7F701` 及
  `Load_Game_Popup_Pictures_ @ 0x7EA5C` 的 `GAME.LBX#0..7` 連續載入鏈。
  正版繁中、正版英文與缺 `GAME.LBX` fallback 畫廊均完成 35 張；已目視抽查
  `19_gamemenu.png`，六顆按鈕的文字中心一致，兩個音量標籤可見且未越出中央面板。
  格子戰術控制列已完成第十四個切片：七顆按鈕與 SCAN／BOARD／AUTO／撤退／登艦戰報等
  玩家訊息均由 `ui.json` 供應，`tacticalbar.go` 不再呼叫 `tr`；按鈕表只保留座標、動作識別字
  與文案鍵。52×16 文字板、54×18 熱區與中心點由程式測試綁在同一份座標，雙語標籤亦經字型
  尺寸檢查。IDA 證實 `sub_2F4EE` 與 `sub_34921` 是多 caller 的戰鬥控制 helper，但外部符號所稱
  `Combat_Screen_ @ 0x478A2` 並非 IDA 函式；相鄰 `sub_478A3` 只是 stride `0x169` 記錄旗標
  predicate。`dword_1A1244` 只有直接 writer，間接字串消費端未閉合，因此原版精確按鈕 widget ID
  與 SCAN／BOARD 字串維持未知。另修正過期的 SETTINGS 提示：13 列設定頁已完成，尚缺的是戰鬥中
  保存狀態並返回同一戰局的轉場。
  研究應用選擇已完成第十五個切片：標題、目前 field 模板與兩個轉場名稱均由 `ui.json` 供應，
  科技及 topic 名稱維持 `tech.json` 單一來源；`researchchoice.go` 不再呼叫 `tr` 或直接繪製字型。
  標題、field 摘要及 application row 全部改走雙軸安全框，row 文字中心與 hover／點擊框由同一個
  `rowRect` 推導。IDA Pro 9.4 本輪唯讀重跑仍確認 `sub_10DC12 @ 0x10DC12` 是選擇函式、
  `sub_E4410 @ 0xE4410` 是突破消費端；field 與 application 在研究開始前寫入。原版把選擇整合於
  `TECHSEL.LBX#0`，remake 的獨立 application 面板仍明標介面轉接近似；精確 row widget 與將
  adapter 合併回八領域畫面是獨立視覺忠實度缺口，不以本輪文案測試升格。
  殖民地 REFIT 已完成第十六個切片：標題、副標題、候選列、預覽、報廢警告、按鈕與所有
  玩家可見錯誤均由 `ui.json` 供應；規則層改回傳穩定 typed `RefitError`，不再保存中文錯誤句子，
  未知錯誤亦有外部文案 fallback。所有欄位改走雙軸安全框，實際 bitmap font 測試另抓到 18px
  標題字墨高 32px、會與舊副標題重疊，現保持標題中心不變並重分配 32+16px header。
  IDA Pro 9.4 本輪證實兩組 REFIT popup／draw／loader 函式邊界，也推翻外部符號把
  `sub_D108B @ 0xD108B..0xD10D2` 與 `sub_D10EE @ 0xD10EE..0xD2754` 都稱為
  `Refit_Cost_` 的混用；後者會呼叫前者，兩者必須分開引用。remake 的單頁自動最佳模板仍是
  已揭露的設計庫替代 UX，不因文案與版面完成而升格為原版兩階段 popup parity。
  共用文字輸入彈窗已完成第十七個切片：`inputBox` API 改收 `ui.json` title key，不再接受已翻譯
  自由句子；對局名稱、直接主機位址、ACCEPT、Enter／Esc 提示與主機預設名均已外部化，三個現有
  caller 皆改用穩定鍵。標題、輸入內容＋游標、98×28 按鈕與底部提示各有雙軸安全框；實際 glyph
  量測修正舊程式把字級當字高、以及提示中心回侵按鈕的問題。IDA Pro 9.4 本輪確認
  `sub_91B89 @ 0x91B89..0x91BB4`、`sub_91BB4 @ 0x91BB4..0x91BD4`、
  `sub_91BD4 @ 0x91BD4..0x91F14`、`sub_91F14 @ 0x91F14..0x9222A` 與
  `sub_F5777 @ 0xF5777..0xF5883` 的獨立邊界及 caller 鏈。現代 IME、Enter／Esc 與 30-frame
  caret 維持明標平台近似，不為 Win95 掃描碼 API 內部另開 RE。Docker + Xvfb 中文畫廊已重跑
  35/35；首次截圖證明按鈕下方提示仍壓到 `INBOX` 美術邊框，現移到輸入欄與按鈕間的 20px 空帶，
  第二次 `34_inputbox.png` 目視確認標題、內容、提示與 ACCEPT 均未互相侵入或超出彈窗。
  領袖技能顯示資料已完成第十八個切片：27 個技能的原版數字 ID 與 2-bit 階級規則留在 Go，
  穩定鍵、中英文名稱改由嵌入式 `internal/gamedata/leader_skills.json` 載入；缺欄、重複 ID 或
  非法 JSON 會直接失敗，不再讓規則表同時承擔玩家文案。無技能領袖的艦長／行政官通稱亦移到
  `ui.json`，避免再次與 Commando 技能標籤碰撞。IDA Pro 9.4 本輪以既有 `.i64` 可寫副本重跑
  隨機領袖招募稽核，確認招募與技能效果仍由 typed ID／2-bit tier 驅動；英文技能名稱由手冊
  p.135–137 錨定，繁中只標為 remake 顯示資料。舊中文存檔的標籤反查仍由同一份 JSON 提供，
  未因外部化而失去加成。本批證據與規格見
  [`docs/re/leader-skill-display-text-audit-20260827.md`](docs/re/leader-skill-display-text-audit-20260827.md) 與
  [`docs/tech/leader-skill-external-text-spec.md`](docs/tech/leader-skill-external-text-spec.md)。
  艦隊名冊選取標記已完成第十九個切片：實際畫廊抓到舊 `✔` 在 runtime 字型路徑成為缺字方框，
  現改由 `ui.json` 提供同寬 ASCII `[x]／[ ]`，Go 只保存兩個語意鍵；正版中文畫廊重跑
  35/35，`07_fleet.png` 三艘預設選取標記可辨識且未侵入艦名、艦級或損傷欄。證據與規格併入
  [`docs/re/auto-select-ships-setting-audit-20260827.md`](docs/re/auto-select-ships-setting-audit-20260827.md) 與
  [`docs/tech/auto-select-ships-setting-spec.md`](docs/tech/auto-select-ships-setting-spec.md)。
  熱座逐席交接已完成第二十個切片：標題、玩家／席位模板、隱私說明、結算提示、接手按鈕及
  兩個轉場標籤全部由 `ui.json` 供應，`hotseat.go` 不再內嵌中英文玩家句子或直接呼叫字型繪製。
  360×230 自繪 adapter 的六個文字區皆有雙軸安全框，按鈕文字與 110×30 熱區共用中心；
  最長雙語模板的 bitmap glyph 量測與中英文畫廊各 35/35 已通過，兩種 `24_hotseat.png` 均
  目視確認未越框；英文操作句另經實圖修正，避免把 TAKE OVER 拆成兩行。
  IDA 同輪推翻舊註解把 `sub_628E2 @ 0x628E2..0x62BB7` 直接當成逐回合 privacy gate 的斷言：
  它確為熱座互動流程並讀取四個文字 ID，但現有證據不足以證明 remake 視窗尺寸、錨點或 TAKE
  OVER 文案是原版精確畫面，因此明標必要的同機隱私轉接設計。證據與規格見
  [`docs/re/hotseat-handoff-ui-audit-20260827.md`](docs/re/hotseat-handoff-ui-audit-20260827.md) 與
  [`docs/tech/hotseat-handoff-external-text-spec.md`](docs/tech/hotseat-handoff-external-text-spec.md)。
  集結點／遷移流程已完成第二十一個切片：規則層改回傳 typed `RelocateRefusal` 與怪獸確認
  布林值，不再保存中文句子；拒絕、提示、結果、按鈕、轉場、怪獸名與清單分隔符全部由
  `ui.json` 供應。IDA Pro 9.4 確認 `sub_75035 @ 0x75035..0x75180` 的黑洞、探索、怪獸及
  殖民地檢查，以及文字 ID `0x83..0x88` 的方向差異；精確原文未匯出，故 JSON 只標等義
  介面轉接。艦隊列表兩個 140×18 入口改用獨立安全框；畫廊另抓到英文王座廳提示會壓到
  下方入口，現一併外部化並放入不相交的 288×18 安全框。中英文畫廊各 35/35 通過，
  `07_fleet.png` 與 `29_confirm.png` 目視確認無重疊、越框或英文單字硬切。證據與規格見
  [`docs/re/relocation-player-text-audit-20260827.md`](docs/re/relocation-player-text-audit-20260827.md) 與
  [`docs/tech/relocation-external-text-spec.md`](docs/tech/relocation-external-text-spec.md)。
  安塔蘭王座廳已完成第二十二個切片：`internal/shell` 的中文阻擋句改為 typed
  `AntaranAssaultBlockReason`，標題、情境、戰力、勝算、按鈕、阻擋原因與轉場全部由
  `ui.json` 供應；七個文字區均改走雙軸安全框，兩顆按鈕文字框與 190×44 熱區共用中心。
  IDA Pro 9.4 證實原版 `sub_14AAC @ 0x14AAC..0x14BFD` 外層、`sub_14BFD` 繪製、
  `sub_14C83` 的 `ANTAROOM.LBX` 載入及 `sub_14D7C` 的 `ANTARMSG.LBX` 訊息選擇；
  外部符號名稱整組錯位，現以 raw 位址保留勘誤。原版滿版訊息與整張畫面輸入已證實，remake
  戰力比較及兩顆按鈕仍明標介面 adapter。實際畫廊先抓到英文按鈕被省略為
  `LAUNCH FINAL ASSA…`，現縮為完整 `FINAL ASSAULT` 並以測試禁止按鈕裁切；中英文畫廊各
  35/35，兩種 `08_antaranroom.png` 目視確認未越框。舊 gap report 誤稱只取最終定格亦已清除；
  現行 55 幀會依序播放，僅每 3 tick 一幀仍是 timing approximation。證據與規格見
  [`docs/re/antaran-room-player-text-audit-20260827.md`](docs/re/antaran-room-player-text-audit-20260827.md) 與
  [`docs/tech/antaran-room-external-text-spec.md`](docs/tech/antaran-room-external-text-spec.md)。
  種族選擇已完成第二十三個切片：Go 表格只保留十四個穩定種族鍵、肖像索引與規則索引；標題、
  族名、形容詞、單行摘要、取消按鈕、預設帝國名格式與轉場均由 `ui.json` 供應。規則層原先未被
  遊戲消費的 `Desc`／`EnDesc` 已移除，外交索引相容查找改由外部雙語名稱與 typed 索引完成。
  IDA Pro 9.4 證實 `sub_5BC74 @ 0x5BC74..0x5BD97` loader、`sub_5BD97` draw 與
  `sub_5C510 @ 0x5C510..0x5CF37` 外層，十四個欄位座標為 x=`0x15F+0x7E*(i/7)`、
  y=`0x5A+0x30*(i%7)`；外部兩份符號表在這三個位址互相衝突，故文件保留 raw 位址。
  中英文畫廊各 35/35；`02_raceselect.png` 目視確認中文所有文字在框內且按鈕置中，英文保留
  原版烘字，選中項與新增摘要／取消轉接面板也未裁切。證據與規格見
  [`docs/re/race-selection-player-text-audit-20260827.md`](docs/re/race-selection-player-text-audit-20260827.md) 與
  [`docs/tech/race-selection-external-text-spec.md`](docs/tech/race-selection-external-text-spec.md)。
  多人主設定已完成第二十四個切片：十顆按鈕只保留資產、座標、動作與穩定文案鍵；標題、
  按鈕、熱座席位格式、TCP／熱座說明、錯誤訊息與轉場均由 `ui.json` 供應。英文正版資產保留
  烘字，繁中與缺資產 fallback 從同一鍵繪製；標題、按鈕及底部說明／錯誤兩列均改走雙軸
  安全框。IDA Pro 9.4 證實 `sub_F42CA @ 0xF42CA..0xF44BB` loader、
  `sub_F009A @ 0xF009A..0xF03F2` widget builder、`sub_F4D99` 主畫面與
  `sub_F5691 @ 0xF5691..0xF5777` 模式 writer；欄位偏移 `0x3B／0x10D`、列偏移
  `0x5B／0x7A／0x9B／0xBB` 及模式 byte `0x199F3A` 均可回查原始 bytes。
  目標測試先抓到缺資產時 TEN 錯套一般按鈕 154px 寬度，現依 `MULTIGM.LBX#256` 已證實尺寸
  改為 253×30，CANCEL fallback 同步釘住 129×25。中英文畫廊各 35/35，兩種
  `23_multiplayer.png` 目視確認無烘字外洩、裁切、越框或按鈕偏心。TCP 與席位循環仍明標
  現代轉接，不升格為原版 IPX parity；證據與規格見
  [`docs/re/multiplayer-setup-player-text-audit-20260827.md`](docs/re/multiplayer-setup-player-text-audit-20260827.md) 與
  [`docs/tech/multiplayer-setup-external-text-spec.md`](docs/tech/multiplayer-setup-external-text-spec.md)。
  區網對局選擇已完成第二十五個切片：標題、空清單兩列、直接位址、取消、兩種連線錯誤及
  三個畫廊示範對局名均由 `ui.json` 供應，`choosemultinetgame.go` 不再內嵌玩家文案或呼叫
  `tr`。十列各自拆成名稱／位址／人數三個不相交的雙軸安全框；標題、空清單、按鈕與訊息列
  亦改走安全框。英文有正版 `MULTIGM.LBX#41` 時保留 `JOIN NETWORK GAME SETUP` 烘字，
  繁中與缺 panel fallback 才重繪。IDA Pro 9.4 證實 `sub_F0C8E @ 0xF0C8E..0xF0E17`
  外層、`sub_F40D3 @ 0xF40D3..0xF41AD` loader、
  `sub_EFF87 @ 0xEFF87..0xF009A` field builder 與
  `sub_F1AF4 @ 0xF1AF4..0xF1CE6` draw；面板、十列、按鈕與選中脈動的立即數均可回查。
  `sub_F5777 @ 0xF5777..0xF5883` 另證實名稱上限 8 與逐一 `strcmp` 重名檢查；舊註解所稱
  remake 尚無名稱輸入框已刪除，現行主機流程早已用 `inputBox` 與 `netplay.GameNameMax=8` 接入。
  UDP discovery 與直接 TCP 位址仍明標 IPX 的現代轉接，不冒稱原版協定內部。中英文畫廊各
  35/35，兩種 `33_netgames.png` 目視確認無標題重繪錯誤、欄位重疊、裁切或按鈕偏心；證據與
  規格見
  [`docs/re/choose-multi-net-game-player-text-audit-20260827.md`](docs/re/choose-multi-net-game-player-text-audit-20260827.md) 與
  [`docs/tech/choose-multi-net-game-external-text-spec.md`](docs/tech/choose-multi-net-game-external-text-spec.md)。
  連線玩家名冊已完成第二十六個切片：標題、空席／主機／本機標記、列號模板、等待提示、
  位址、共同種子、網路錯誤、預設玩家名與畫廊假資料均改由 `ui.json` 供應；標題、列號、
  玩家名及兩行狀態分別套用雙軸安全框。舊文件所稱只能連本機的斷言已清除，現行
  「直接輸入位址」是正常後備入口。原版逐列種族／顏色指派仍是獨立玩法缺口，未因名冊
  文案與版面完成而升格；規格見
  [`docs/tech/network-roster-external-text-spec.md`](docs/tech/network-roster-external-text-spec.md)。
  `Net_Next_Turn` 畫廊 adapter 已完成第二十七個切片：標題、三種玩家狀態、本機後綴、
  回合／指紋格式、分岔警告、兩種聊天前綴、游標及畫廊訊息全由 `ui.json` 供應；
  `internal/netplay` 只接受呼叫端外部格式，不再保存可見括號字串。IDA Pro 9.4 重跑確認
  `sub_FC470 @ 0xFC470..0xFC6A5`、`sub_F3E42 @ 0xF3E42..0xF3FC6`、
  `sub_EFCEA @ 0xEFCEA..0xEFE7A`、`sub_F1075 @ 0xF1075..0xF166E`、
  `sub_F31BB @ 0xF31BB..0xF33AE` 與 `sub_F55A4 @ 0xF55A4..0xF5681`；輸入列
  `0xBB／0x11`、玩家列步距 `0x19`、GNN speaker `8` 與聊天記錄 `+0x47C` 可回查，
  第一列 y 仍維持近似。runtime 字型測試另抓到標題、指紋、分岔警告與輸入列舊框垂直不足，
  現已依實際 glyph ink 修正；中英文畫廊各 35/35，兩種 `30_netwait.png` 目視無越框。
  正式多人回合等待已完成第二十八個切片：IDA 補查證實 `sub_FC470` 分流
  `sub_FBFE2 @ 0xFBFE2..0xFC299` 與 `sub_FC2D2 @ 0xFC2D2..0xFC470`，客戶端直接呼叫
  `sub_F7E95 @ 0xF7E95..0xF83B7` 送單，主客戶端都使用 `sub_F1075` renderer。
  `networkWaitScreen` 現保留現代兩階段鎖步 update loop，但正式畫面共用 `netNextTurnScreen`
  的原版三塊面板、玩家列、聊天記錄與輸入列；等待時可輸入並送出 `KindChat`，renderer 不 poll
  session，避免吞掉鎖步封包。自製 556×396 替代面板與直接字型繪製已移除，固定轉場、後備名、
  協定錯誤及返回提示均由 `ui.json` 供應；雙 peer 測試已在同一正式等待階段同時傳聊天、
  `turn_done` 與 `turn_ready` 並完成第一回合。證據與規格見
  [`docs/re/net-next-turn-player-text-audit-20260827.md`](docs/re/net-next-turn-player-text-audit-20260827.md) 與
  [`docs/tech/net-next-turn-external-text-spec.md`](docs/tech/net-next-turn-external-text-spec.md)、
  [`docs/tech/network-turn-wait-player-path-spec.md`](docs/tech/network-turn-wait-player-path-spec.md)。
  熱座真人帝國選擇已完成第二十九個切片：標題、說明、選取數、列格式、勾選標記、按鈕、
  轉場與建立錯誤均由 `ui.json` 供應，內建種族名複用 `race.select.race.*.name`；所有文字區
  都改走雙軸安全框，列名及按鈕不再直接呼叫字型繪製。IDA Pro 9.4 同輪修正舊摘要：
  `sub_121F0 @ 0x121F0..0x12227` 不是只檢查 `player+0x28 == 100`，還要求
  `player+0x24 == 0` 才計入真人數。`+0x24` 精確欄名仍未知，保留 raw 偏移；現行選取清單
  是 `GameSession` 單玩家模型所需的接管 adapter，不冒稱原版版面。證據與規格見
  [`docs/re/hotseat-empire-selection-audit-20260827.md`](docs/re/hotseat-empire-selection-audit-20260827.md) 與
  [`docs/spec/hotseat-empire-selection.md`](docs/spec/hotseat-empire-selection.md)。
  共用是／否確認框已完成第三十個切片：繁中按鈕及缺資產英文後備由 `ui.json` 提供，
  `confirmbox.go` 不再內嵌固定按鈕文案或直接呼叫字型繪製；51×21 文字框與點擊熱區共用
  整數像素中心。缺 `CONFIRM.LBX#1/#2` 時會畫出可見底板、邊框與雙語標籤，不再留下可點但
  不可見的熱區。IDA Pro 9.4 複核 `sub_77658 @ 0x77658..0x77860` 的 34 個直接呼叫點、
  Y／N 快捷鍵、兩顆按鈕座標、`sub_778E4` hover frame，以及從字級 4 降到 1、文字高度
  `<=0x7E` 的原版停止條件；現行固定字級＋省略號維持明標近似。繁中與英文正版資料畫廊
  各 35/35，兩種 `29_confirm.png` 已目視確認正文與按鈕在框內，英文原版烘字未被覆蓋。
  證據與規格見
  [`docs/re/confirm-box-player-text-audit-20260827.md`](docs/re/confirm-box-player-text-audit-20260827.md) 與
  [`docs/spec/confirm-box-external-text.md`](docs/spec/confirm-box-external-text.md)。
  星圖外交會談請求燈已完成第三十一個切片：四種來意 glyph 改由 `ui.json` 提供，
  `audience.go` 不再呼叫 `tr` 或直接繪製字型；22×16 方塊的文字、面板與點擊熱區由同一矩形
  推導，實際點陣 glyph 高 16px，故只保留 2px 水平內縮。IDA Pro 9.4 同輪推翻舊版面斷言：
  `sub_83D06 @ 0x83D06..0x83DEA` 把第一盞燈的**左緣**放在 x=506，不是把右緣放在 506；
  後續才按已畫數乘動畫物件 0 寬度往左移。遮罩由 `sub_FA795` 唯一讀取
  `byte_1AB054`，逐帝國 race 索引、frame 遞增與 `+6` frame count 歸零均可回查原始 bytes。
  原版逐種族動畫資產仍未知；現行旗色方塊、來意色與 glyph 是明標 adapter，AI 觸發政策也
  維持既有近似，不因本輪版面與文案完成而升格。證據與規格見
  [`docs/re/diplomacy-request-lights-audit-20260827.md`](docs/re/diplomacy-request-lights-audit-20260827.md) 與
  [`docs/spec/diplomacy-request-light-external-text.md`](docs/spec/diplomacy-request-light-external-text.md)。
  格子戰術戰機介面已完成第三十二個切片：出擊、返航、場上摘要、登艦結果與三種 token glyph
  共九組固定文案全部由 `ui.json` 提供，`tacticalfighter.go` 不再呼叫 `tr` 或直接呼叫字型繪製。
  90×26 出擊鈕與 24×16 中隊 token 均以 `textSafeRect` 約束，按鈕文字框與命中框共用邊界。
  原版 `sub_3AC20`／`sub_3AD57`／`sub_3D2DF` 只支撐既有 runtime 玩法證據；標題列出擊鈕、
  單 token 與 glyph 是明標 remake adapter，不升格為原版逐架動畫。IDA 專用映像本輪因
  IDAPython 未設定及授權不可用而無法產生新匯出，已如實記為工具鏈阻塞；既有結論未升格。
  Docker＋Xvfb 繁中畫廊已完成 35/35，目視抽查 `16_tactical.png`，出擊戰報與 `◇4` token
  均在各自框內，未侵入相鄰艦艇名稱或控制列。
  證據與規格見
  [`docs/re/tactical-fighter-text-layout-audit-20260827.md`](docs/re/tactical-fighter-text-layout-audit-20260827.md) 與
  [`docs/spec/tactical-fighter-text-layout.md`](docs/spec/tactical-fighter-text-layout.md)。
  Smacker 過場已完成第三十三個切片：IDA Pro 9.4 證實 `sub_14DF7 @ 0x14DF7..0x15085`
  是 `Play_Cinematic_` 完整外層，逐幀迴圈同時讀 `Keyboard_Status_`／`Read_Key_` 與
  `Mouse_Button_`，且沒有任何固定文字列印。remake 現接受任意鍵或滑鼠跳過，並移除非原版的
  底部「點擊跳過」提示；片頭與最終得分轉場名稱改由 `ui.json` 供應。外部符號表的相鄰名稱
  錯位已保留 raw 位址說明，不覆蓋資料庫名稱。Smacker 真實資產音軌亦已依
  `docs/re/cutscene-audio-audit-20260827.md` 接入，DAC／PIT／MSS 逐週期內部依停止線不追回。
  Docker＋Xvfb 繁中畫廊完成 35/35；目視抽查 `21_intro.png` 與
  `22_ending.png`，影片只含原始 frame 與黑邊，沒有額外提示侵入畫面。證據與規格見
  [`docs/re/cutscene-player-path-audit-20260827.md`](docs/re/cutscene-player-path-audit-20260827.md) 與
  [`docs/spec/cutscene-player-path.md`](docs/spec/cutscene-player-path.md)。
  外交對談已完成第三十四個切片：原版 `DIPLOMAT.LBX` 房間／使節動畫及請求燈證據與
  remake 三欄操作轉接明確分級；13 個提議、五種動態解約、標題、使節模板、協議摘要與離開
  按鈕均改由 `ui.json` 供應。`diplomacyOption` 只保存穩定文案鍵與規則 action，按鈕文字框由
  點擊熱區推導並經雙語最長字串測試；同時修正五種協議皆可終止時第五顆按鈕從 x=640
  跑出畫布、解約列與最後一列提議重疊的舊版面。外交規則與三種餽贈亦不再直接組中文句子：
  shell 回傳 typed `DiplomacyResultCode` 及金額／科技／殖民地參數，UI 才以
  `diplomacy.response.*` 雙語 JSON 模板格式化，未知 code 安全 fallback。條約摘要也改由
  `TreatySummaryParts` 回傳正式狀態、進貢百分比、普通協議回合／值與特殊貿易種類；名稱、
  BC／RP 模板與分隔符只存在 `ui.json`，空狀態與未知種類皆有 typed fallback。原版逐句提議分頁、
  完整回應動畫與熱區仍維持未知，
  不因外部文案完成而升格。證據與規格見
  [`docs/re/diplomacy-audience-text-audit-20260827.md`](docs/re/diplomacy-audience-text-audit-20260827.md) 與
  [`docs/spec/diplomacy-audience-external-text.md`](docs/spec/diplomacy-audience-external-text.md)、
  [`docs/spec/diplomacy-result-external-text.md`](docs/spec/diplomacy-result-external-text.md)、
  [`docs/spec/treaty-summary-external-text.md`](docs/spec/treaty-summary-external-text.md)。
  星圖選星側欄已完成第三十五個切片：稅率、環境兩列、陸戰隊／艦員、怪獸／特殊物產前綴、
  ETA、拓殖／前哨／轟炸／入侵／心靈控制／派遣按鈕及本地成功提示均由 `ui.json` 供應，
  行星與怪獸名稱只作動態參數。面板高度由 132 修為 140，使 y=446..466 的心靈控制列完整
  落在框內；雙語最長字串、格式模板與面板 containment 均有測試。拓殖／前哨站亦完成第三十六個
  切片：`internal/shell` 的自由中文 `Reason string` 已改成 typed 拒絕碼，怪獸與天體類別以
  enum 參數傳到 UI；星圖與行星列表只從 `ui.json` 組合雙語結果，行星列表相鄰的選取、派遣與
  成功提示亦不再內嵌於 Go。轟炸／入侵／心靈控制／怪獸戰術已完成第三十七個切片：三種結果
  結構及 `StartMonsterCombat` 第三回傳值均改為 typed code，星圖只從 `ui.json` 取得雙語拒絕
  文案；無 UI 消費端的舊 `AttackMonster` 亦移除預先組好的中文 `Name`／`Message`，改留 typed
  怪獸種類與純數值戰果。這只封閉前置 gate 與文案分層，不重開傷亡公式、怪獸 blueprint 或
  即時動畫。其他尚未 typed 化的 shell 玩家訊息仍隨後續逐畫面盤點，不把本批誤稱全域完成。
  銀河議會已完成第三十八個切片：標題、成立／勝負狀態、候選人與逐帝國投票列、玩家三選一、
  接受／拒絕決議等固定句與格式模板均由 `ui.json` 供應，`council()` 不再呼叫 `tr`。三個
  150×40 投票文字框及兩個 400×26 決議文字框由實際熱區推導並共用中心；標題、摘要與投票列
  亦使用不相交的雙軸安全框。runtime 字型測試抓到舊 18px 決議字墨高 32px、超出 22px 內容框，
  現降為 12px 並以雙語最長模板驗證。議會選舉公式與 10 幀資產鏈不因文案切片重新升格；精確
  原版逐字與動畫停留時間仍維持未知。證據與規格見
  [`docs/re/council-screen-text-audit-20260827.md`](docs/re/council-screen-text-audit-20260827.md) 與
  [`docs/spec/council-screen-external-text.md`](docs/spec/council-screen-external-text.md)。
  艦隊戰果摘要已完成第三十九個切片：勝敗、敵方、開戰艦數、逐回合與總損失模板改由
  `ui.json` 供應，面板標題／關閉字則沿用 `misc.json` 的 overlay catalog；`battleResult()` 不再
  呼叫 `tr`。快速結算與安塔蘭終局原先保存的中文 `Log []string` 改為 typed
  `BattleRoundResult` 純數值，安塔蘭固定敵方名亦改由 `BattleEnemyKind` 交給 UI 翻譯。
  最多六列戰報、總損失及三列摘要各有不相交雙軸安全框，雙語長敵方名與最大測試數值均經
  runtime 字型 containment 驗證。此畫面仍是借用 `TURNSUM.LBX#0` 的 remake adapter，未升格為
  原版專用戰果畫面 parity。證據與規格見
  [`docs/re/battle-result-screen-text-audit-20260827.md`](docs/re/battle-result-screen-text-audit-20260827.md) 與
  [`docs/spec/battle-result-screen-external-text.md`](docs/spec/battle-result-screen-external-text.md)。
  行星列表已完成第四十個切片：固定的殖民地／前哨狀態與空清單提示由 `ui.json` 供應，
  `planets()` 不再呼叫 `tr`；無任何呼叫端、卻仍在 shell 組中文句子的 `OutpostTargetHint` 已移除。
  每個 50px 列的五欄主文字與星系／特殊物產／佔領狀態第二行改用由原欄界推導的雙軸安全框，
  不再只有行星名手動限寬。右側操作訊息框底緣 y=384，與 y=386 的殖民船按鈕保持 2px 間隔；
  空清單也有獨立框。雙語最長狀態、行星名與派遣訊息已通過 runtime 字型及列 containment 測試。
  排序／篩選與殖民／前哨玩法未在文案切片中改動。證據與規格見
  [`docs/re/planet-list-screen-text-audit-20260827.md`](docs/re/planet-list-screen-text-audit-20260827.md) 與
  [`docs/spec/planet-list-screen-external-text.md`](docs/spec/planet-list-screen-external-text.md)。
  新遊戲設定已完成第四十一個切片：固定 DIFFICULTY／GALAXY SIZE／PLAYERS／TECH LEVEL、
  三個開關與 ACCEPT／CANCEL 原已由 `menu.json` overlay catalog 提供；本輪再把星系大小＋星數、
  帝國數兩個動態模板及設定／主選單轉場移至 `ui.json`，`ngSettings` 到 `newGameSetup()` 的完整
  source slice 已無 `tr`。五個數值列持續使用原版 100×20 selector 熱區衍生的 `ngStripTextRect`，
  並以繁中／英文最長實際設定值驗證 12→11→10px 字高降級至少有一級能落入 16px 內容區；文字
  與熱區中心一致。原版值圖、兩版背景索引、選項數與開局規則均未改動。證據與規格見
  [`docs/re/newgame-setup-text-audit-20260827.md`](docs/re/newgame-setup-text-audit-20260827.md) 與
  [`docs/spec/newgame-setup-external-text.md`](docs/spec/newgame-setup-external-text.md)。
  艦隊作戰名冊已完成第四十二個切片：未知地點、艦隊標頭、航行後綴、拆分提示、結構損傷與
  四種轉場名稱均移至 `ui.json`，`fleet()` 已無 `tr`。艦艇名稱、艦級與損傷改用三個互不重疊的
  雙軸安全欄，標頭與拆分提示亦限於左側名冊框；雙語長星名、長艦名、三位數 ETA 與損傷均有
  runtime 字型 containment 測試。原版名冊超出可視容量時的捲動／分頁輸入仍屬未知，現有 source
  也沒有捲動狀態；本切片沒有猜造新操作，這個容量缺口須待原版輸入證據或使用者產品決策後再
  關閉。證據與規格見
  [`docs/re/fleet-operations-text-audit-20260827.md`](docs/re/fleet-operations-text-audit-20260827.md) 與
  [`docs/spec/fleet-operations-external-text.md`](docs/spec/fleet-operations-external-text.md)。
  領袖／軍官管理已完成第四十三個切片：雇用、任命、撤回、解雇等操作回饋，目標提示、名冊
  模板、任命狀態、空池提示、選取／候選標記與轉場名稱全部由 `ui.json` 供應；HERODATA 技能
  顯示名則直接選取 `leader_skills.json` 資料，`officer()` 與相鄰技能顯示 helper 均不再呼叫
  `tr`。頂部目標／回饋／雇用模式及四列姓名／技能／任命狀態已改用不重疊雙軸安全框；原本
  畫在 y=445、會壓住 y=441 底部按鈕列的「還有 N 位」額外提示已移除，保留原版已證實的
  上下捲動入口；HIRE 狀態由操作回饋呈現，不另畫會撞第一列任命狀態的常駐提示。雙語長領袖名、
  艦名、技能名與最大測試數值已通過 runtime 字型 containment。
  點名冊列立即任命仍是明標 remake adapter，未因文案切片升格為原版 `Check_Officer_Fields_`
  精確控制流。證據與規格見
  [`docs/re/officer-screen-text-layout-audit-20260827.md`](docs/re/officer-screen-text-layout-audit-20260827.md) 與
  [`docs/spec/officer-screen-external-text.md`](docs/spec/officer-screen-external-text.md)。
  殖民地總覽已完成第四十四個切片：殖民地列、建造進度／佇列／已建摘要、Empire Summary、
  Planetary Info、Production Info、饑荒／估算標記與三種轉場均由 `ui.json` 供應；原先內嵌 Go
  的 23 組氣候／重力／礦產／大小雙語顯示名及 unknown fallback 也移入 JSON，enum 只保留
  穩定 key 映射。九列五欄與三個下方面板皆改走雙軸安全框，`postDraw` 不再直接呼叫字型。
  runtime 量測證實 7px 繁中字墨仍高 16px：原建造＋已建兩行及 Planetary 六列必然重疊，故
  建造資訊合為單列（主要建造置前），Planetary 移除與上表重複的殖民地序號，三面板統一為
  五列×17px。雙語長環境名、建築名與六位數產出已通過 containment；懸停內容與 Empire
  Summary 仍是明標手冊＋remake adapter，未升格為原版逐字／逐值 parity。證據與規格見
  [`docs/re/colony-summary-text-layout-audit-20260827.md`](docs/re/colony-summary-text-layout-audit-20260827.md) 與
  [`docs/spec/colony-summary-external-text.md`](docs/spec/colony-summary-external-text.md)。
  艦艇設計已完成第四十五個切片：畫面轉場、建造回饋、艦體成本、元件屬性、射界、彈架、
  掛載控制、解鎖摘要、空間分解與改造晶片文案均由 `ui.json` 供應，Go 僅保留穩定 key 與
  動態數值。解鎖摘要／建造錯誤、空間表及改造區改成固定且互不侵入的雙軸安全框；runtime
  字形量測另抓出 15px 晶片列距小於 16px 中文字形的必然重疊，現已改為四條 16px 列，
  文字與點擊區共用邊界並在面板底界內結束。武器規則、成本、空間、射界、彈架與改造可用性
  未被此文案切片更動；右上原版資訊面板逐字內容與部分操作控制流仍維持明標 adapter，未升格
  為原版精確 parity。證據與規格見
  [`docs/re/ship-design-screen-text-layout-audit-20260827.md`](docs/re/ship-design-screen-text-layout-audit-20260827.md) 與
  [`docs/spec/ship-design-external-text.md`](docs/spec/ship-design-external-text.md)。
  主選單已完成第四十六個切片：左下語言列、規則版本列及相關轉場名稱均由 `ui.json` 供應，
  Go 只保留穩定 key、目前語言與 `1.3`／`1.5` 動態值。兩條 remake 擴充列改由各自
  220×22 點擊列推導雙軸安全框，文字與熱區共用中心；繁中、英文及刻意加長的版本值已通過
  runtime fallback 字型 containment；現行繁中與英文正版資料畫廊各 36/36，英文實圖另確認正常值
  完整顯示、不依賴省略號收尾。原版證據只支持 `MAINMENU.LBX#21` 的六顆既有按鈕，
  語言／版本列仍明標 remake requirement，不升格為原版控制。證據與規格見
  [`docs/re/main-menu-text-layout-audit-20260827.md`](docs/re/main-menu-text-layout-audit-20260827.md) 與
  [`docs/spec/main-menu-external-text.md`](docs/spec/main-menu-external-text.md)。
  RACES 種族／間諜畫面已完成第四十七個切片：帝國態勢、軍力／星數、AI 彼此關係、增派 Spy、
  三種任務、Agent 狀態／控制及轉場名稱均由 `ui.json` 供應；`shell.SpyMissionLabel` 與
  `AIRelationName` 已移除，規則層只留 typed 任務、關係分數及相容存檔狀態。runtime 稽核推翻
  五條 16px 字形可塞入約 73px 資訊區的舊假設，現改為四條互不重疊的 17px 節奏；整區外交
  hit／hover 保留。新增正常玩家路徑 `15a_races.png` 又證實舊 Agent 控制會遮住左欄第 4 個
  帝國槽，現移入右下 BONUSES 面板並在 y=418 外交按鈕列前結束。繁中與英文正版資料畫廊各
  36/36，兩張 RACES 畫面均已目視確認。原版三個任務槽左右語意與完整 callback 仍未知，現行
  增派／循環／隱匿控制維持明標 remake adapter。證據與規格見
  [`docs/re/races-spy-text-layout-audit-20260827.md`](docs/re/races-spy-text-layout-audit-20260827.md) 與
  [`docs/spec/races-spy-external-text.md`](docs/spec/races-spy-external-text.md)。
  格子戰術已完成第四十八個切片：畫面標題、開場操作提示、行動佇列、移動、目標拒絕、
  開火、回合、戰機附註、勝敗、戰後轉場，以及武器槽三態／右鍵明細均由 `ui.json` 供應；
  `newTacticalScreenForShips` 到戰術繪製終點的來源切片已無 `.tr(`，Go 僅保留語意鍵與具型別
  動態值。原版三態與單次待命恢復已有 IDA／手冊錨點；戰報逐字內容、右鍵彈窗外觀與循環
  方向仍維持明標 remake adapter。`cmd/moo2`＋`internal/shell` 全套測試通過，繁中與英文正版
  資料畫廊各 36/36，實圖抽查 `16_tactical.png` 的標題、戰報、艦名、武器列與控制列均在
  安全框內。證據與規格見
  [`docs/re/tactical-dynamic-battle-log-audit-20260827.md`](docs/re/tactical-dynamic-battle-log-audit-20260827.md) 與
  [`docs/spec/tactical-dynamic-text-catalog.md`](docs/spec/tactical-dynamic-text-catalog.md)。
  回合摘要已完成第四十九個切片：星曆、經濟、破產、研究、飢荒／叛亂、事件包裝、四種建造
  結果與轉場模板均由 `ui.json` 供應；規則層的中文 `LastBuilt []string` 已改為暫態 typed
  `BuildNotice`，一般建造名稱在 UI 依語言翻譯。地震刪殖民地不再把完工通知誤當平行陣列裁切，
  而是刪除該殖民地通知並重編後續索引。四條基礎列與 y=168..306 動態區都使用固定安全框；
  破產、飢荒、叛亂、研究、多項完工、事件、安塔蘭與突襲同回合時最多七列，超量末列依實際
  字型量測加省略號，不侵入 y=324 關閉按鈕。`cmd/moo2`＋`internal/shell` 全套測試通過，繁中與
  英文正版資料畫廊各 36/36，實圖抽查 `06_turnsummary.png` 的標題、四列與關閉鈕均在框內。
  原版完整逐欄組裝、排序及分頁仍未知；現行摘要維持明標 remake adapter。證據與規格見
  [`docs/re/turn-summary-dynamic-text-layout-audit-20260827.md`](docs/re/turn-summary-dynamic-text-layout-audit-20260827.md) 與
  [`docs/spec/turn-summary-external-text.md`](docs/spec/turn-summary-external-text.md)。
  研究領域畫面已完成第五十個切片：目前主題、RP 成本、超進階等級、完成狀態及星系轉場均由
  `ui.json` 供應，主題名稱仍由 `tech.json` 提供；Go 僅傳 typed 主題、等級與成本。八個動態列
  現由各自 hit region 推導單行雙軸安全框，不再只有寬度限制。原版 `sub_10DC12 @ 0x10DC12`
  已證實 field／application 寫回與 `TECHSEL.LBX#0` 資產鏈；現行「各領域一列摘要＋另頁選
  application」仍明標 remake adapter，未知原版 widget 精確幾何。`cmd/moo2`＋`internal/shell`
  全套測試通過；繁中與英文正版資料畫廊各 37/37，新增 `35_research.png`，實圖抽查八列均在
  領域框內。證據與規格見
  [`docs/re/research-area-dynamic-text-layout-audit-20260827.md`](docs/re/research-area-dynamic-text-layout-audit-20260827.md) 與
  [`docs/spec/research-area-external-text.md`](docs/spec/research-area-external-text.md)。
  星圖快捷鍵回饋已完成第五十一個切片：F10 成功、失敗與無槽位回饋，以及 F9 起點、目標與
  秒差距模板均由 `ui.json` 供應；Go 只傳 runtime 錯誤或距離數值。IDA 非破壞性匯出確認
  `sub_825A8 @ 0x825A8..0x82809` 的負值輸入碼 jump table，並重驗
  `sub_EBE79 @ 0xEBE79..0xEBEB7` 以 `900=30²` 及整數平方根向上取整計算秒差距。原版逐字
  提示、幾何與再按 F9 取消仍未知，現行回饋維持明標 remake adapter。游標提示會依星圖邊界
  翻邊並夾限，距離中點及快速存檔各使用固定雙軸安全框，不再由未截斷字寬擴張底板。
  `cmd/moo2`＋`internal/shell` 全套測試通過；繁中與英文正版資料畫廊各 37/37，實圖抽查
  `28_measure.png` 的距離框均在星圖內。證據與規格見
  [`docs/re/galaxy-hotkey-feedback-layout-audit-20260827.md`](docs/re/galaxy-hotkey-feedback-layout-audit-20260827.md) 與
  [`docs/spec/galaxy-hotkey-external-text.md`](docs/spec/galaxy-hotkey-external-text.md)。
  Runtime 顯示標籤已完成第五十二個切片：四個歷史指標、十五個新遊戲值、同星系天體數、
  玩家圖例、四種戰機、敵艦／熱座組字及三種 Unknown fallback 均由 `ui.json` 供應；Go 只保留
  typed enum／索引到穩定鍵的映射與動態參數。無效新遊戲索引現依目前語言顯示 Unknown，
  不再在繁中模式固定回英文。這些詞義分別沿用已審查的新遊戲、INFO、戰機與熱座 RE；
  歷史圖例、舊存檔未知值及名稱組字仍明標 remake adapter，不升格原版逐字 parity。
  漢字 literal 棘輪依實際輸出由 16 收緊至 13，剩餘六個規則查表鍵與七個 dev-only 檢視器標題；
  同時新增純英文來源契約，避免舊探針看不到的英文硬編文案回流。`cmd/moo2`＋`internal/shell`
  全套測試通過；繁中與英文正版資料畫廊各 37/37，實圖抽查新遊戲、行星列表、INFO 歷史、
  格子戰術及熱座交接均無 key 洩漏或越框。證據與規格見
  [`docs/re/runtime-label-catalog-audit-20260827.md`](docs/re/runtime-label-catalog-audit-20260827.md) 與
  [`docs/spec/runtime-label-catalog.md`](docs/spec/runtime-label-catalog.md)。
  元件顯示名稱已完成第五十三個切片：六個 `TECH_NONE` 佔位／抽象元件的中英文名稱與未知
  fallback 均由 `ui.json` 供應，Go 只保留既有規則鍵到語意鍵的路由。舊註解所稱五筆已按
  實際清單訂正為六筆；未知無科技元件在英文模式也不再退回中文規則鍵。有 `UnlockTech` 的
  英文名稱仍由原版執行檔科技名稱表推導，避免複製第二份元件表；繁中科技元件名稱仍沿用
  typed 元件資料，屬後續資料表外部化邊界。本輪 `cmd/moo2`＋`internal/shell` 測試通過，
  繁中與英文正版資料畫廊各 37/37；實圖抽查 `25_shipdesign.png` 的 `無護盾`／`無` 與
  `No Shield`／`None` 均正確且未越框。後續閉合科技元件的繁中顯示邊界：有 `UnlockTech` 的
  元件現一律以原版英文科技鍵查外部 `tech.json`，不再直接輸出 `Component.Name`；例如規則鍵
  「雷射」在畫面使用正式科技譯名「雷射砲」。新增測試逐項確認四張元件表都有外部繁中譯文，
  並禁止顯示函式回傳內部規則鍵；雙語畫廊重跑仍各 37/37，實圖未越框。證據與規格見
  [`docs/re/component-display-label-catalog-audit-20260827.md`](docs/re/component-display-label-catalog-audit-20260827.md) 與
  [`docs/spec/component-display-label-catalog.md`](docs/spec/component-display-label-catalog.md)。
  艦級顯示名稱已完成第五十四個切片：六個戰鬥艦體、四個支援艦及未知 fallback 的中英文文字
  均由 `ui.json` 提供；`Ship.Class`／`shipClassZH` 僅保留規則與存檔鍵，不再直接成為玩家輸出。
  艦艇設計標題、六個艦體名與 CLEAR／CANCEL／BUILD 亦移除 Go 內嵌文字；艦體熱區 action 從
  英文顯示名改成 `hull:<index>`，翻譯不再兼任控制識別。`cmd/moo2`＋`internal/shell` 測試通過，
  繁中與英文正版資料畫廊各 37/37；實圖抽查 `07_fleet.png` 與 `25_shipdesign.png` 的支援艦、
  六艦體、標題及按鈕均正確且未越框。證據與規格見
  [`docs/re/ship-class-display-catalog-audit-20260827.md`](docs/re/ship-class-display-catalog-audit-20260827.md) 與
  [`docs/spec/ship-class-display-catalog.md`](docs/spec/ship-class-display-catalog.md)。
  自訂種族控制文字已完成第五十五個切片：數值選項成本、特殊能力 on／off marker、能力列組字
  與正負點數格式全部由 `ui.json` 提供，`customrace.go` 只傳名稱、選取狀態與整數。RACESTUF
  只證實選項語料，未證實 `○／●` 與成本排版；文件已把這組視覺維持為 remake adapter，沒有
  升格原版 glyph。畫廊新增 `02b_customrace.png`，使用正式 `customRace()` 畫面建構流程；繁中
  與英文正版資料畫廊現各 38/38。實圖抽查兩種語言的標記、名稱、成本、標題與按鈕均在安全框
  內，英文長能力名依既有單行省略契約收尾。證據與規格見
  [`docs/re/custom-race-ui-text-audit-20260826.md`](docs/re/custom-race-ui-text-audit-20260826.md) 與
  [`docs/spec/custom-race-external-text.md`](docs/spec/custom-race-external-text.md)。
  行星特殊物產與勘查報告已完成第五十六個切片：十二個原版代碼只保留 typed enum 到
  `planet.special.*` 語意鍵的映射，中英文名稱、`★` 標記及八種勘查結果模板均由
  `ui.json` 提供。新 `SystemDiscovery` 只保存 BC、人口、領袖與科技主題等型別化結果，
  不再於規則層組合雙語句子；舊 Name／Message 欄位僅供既有存檔安全回退。遠古文物所贈科技
  改保存 `ResearchTopic`，顯示時逐項經 `tech.json` 翻譯，修正英文報告混入中文頓號的問題。
  聚焦測試已覆蓋特殊物產 catalog、typed 科技清單、舊存檔回退及來源碼防回歸；繁中與英文
  正版資料畫廊各重跑 38/38，抽查星圖、殖民地主畫面及行星列表沒有文案鍵外洩或新增越框。
  原版勘查報告逐字與 `★` 排版仍明標未知／remake adapter，不因外部化升格 parity。證據與規格見
  [`docs/re/planet-special-display-catalog-audit-20260827.md`](docs/re/planet-special-display-catalog-audit-20260827.md) 與
  [`docs/spec/planet-special-display-catalog.md`](docs/spec/planet-special-display-catalog.md)。
  安塔蘭戰略入侵通知已完成第五十七個切片：出發、抵達 AI 殖民星、未設防抵達及玩家守軍
  勝／敗不再由 `internal/shell` 組合雙語句子，規則層只保存通知種類、雙語星名、ETA、損失與
  是否擊退的 `AntaranNotice`。回合摘要與 INFO 摘要共用 `antaranNoticeText`，五種句型、警示
  符號及未知 fallback 全由 `ui.json` 提供；熱座席位與重要摘要 gate 也已改讀 typed notice。
  現有 IDA 證據只支持出兵、目標星與戰鬥資料，不支持任何通知逐字，因此句型仍明標 remake
  adapter。聚焦測試已覆蓋五個雙語分支、英文星名、熱座往返、重要摘要及來源碼防回歸。
  繁中與英文正版資料畫廊各重跑 38/38；抽查回合摘要與 INFO 畫面未見 key 外洩、重疊或越框，
  但畫廊沒有強制注入安塔蘭事件，五種實際通知內容仍由 typed 測試證明而非冒充視覺 oracle。
  證據與規格見
  [`docs/re/antaran-notice-text-audit-20260827.md`](docs/re/antaran-notice-text-audit-20260827.md) 與
  [`docs/spec/antaran-notice-external-text.md`](docs/spec/antaran-notice-external-text.md)。
  AI 殖民地突襲通知已完成第五十八個切片：`AIRaidReport` 只保留 AI／星系名、勝敗、人口、
  BC、建築與艦力損失；規則層已移除 `Message`／`MessageEN` 與額外 `LastRaid` 字串旗標。
  擊退、突破、建築摧毀與攻方折損主句／片段／分隔符均由 `ui.json` 提供，回合摘要與 INFO
  共用 `aiRaidNoticeText`；英文建築名經既有 building catalog 翻譯，不再直接插入中文規則名。
  重要摘要、熱座席位及舊測試也改以 `LastRaidReport != nil` 判定。原版證據支持殖民地估值與
  艦隊抵達，不支持現行發動／損失／通知句型，故仍明標 remake adapter。聚焦測試已覆蓋雙語
  勝敗、兩個可選片段、未知名稱 fallback、熱座往返與來源碼防回歸。
  繁中與英文正版資料畫廊各重跑 38/38；正常畫廊未強制觸發 AI 突襲，因此只證明 UI 流程沒有
  回歸，實際通知內容與建築翻譯由 typed 測試覆蓋，不冒充原版視覺對照。證據與規格見
  [`docs/re/ai-raid-notice-text-audit-20260827.md`](docs/re/ai-raid-notice-text-audit-20260827.md) 與
  [`docs/spec/ai-raid-notice-external-text.md`](docs/spec/ai-raid-notice-external-text.md)。
  銀河議會回合通知已完成第五十九個切片：規則層移除繁中 `LastCouncil` 成品字串，改以
  `CouncilNotice` 保存候選不足、等待真人投票、玩家當選、AI 當選待回應及無人過門檻五種
  typed 結果，以及屆次、候選索引、明確當選 slot、得票與總票。INFO 回合摘要現共用
  `councilNoticeText`；五種雙語模板、玩家稱呼與未知 fallback 均由 `ui.json` 提供，英文模式
  依 AI `RaceIndex` 顯示原版英文種族名，不再洩漏繁中規則字串。議會排程、`Vote_Check_`、
  2／3 門檻與接受／拒絕流程未改；原版證據只支持狀態與票數，通知逐字仍明標 remake adapter。
  聚焦測試已覆蓋五種雙語結果、英文名稱及來源碼防回歸；繁中與英文正版資料畫廊各重跑
  38/38。正常畫廊未強制召開議會，故只證明 UI 路由無回歸，通知內容由 typed 測試覆蓋。
  證據與規格見
  [`docs/re/council-notice-text-audit-20260827.md`](docs/re/council-notice-text-audit-20260827.md) 與
  [`docs/spec/council-notice-external-text.md`](docs/spec/council-notice-external-text.md)。
  帝國投降事件 34 已完成第六十個切片：既有 `EventReport.Target*`／`SecondaryTarget*` 雙 target
  record 現直接作 typed 顯示輸入，`queueEmpireSurrender` 不再建立中英文完整句子。事件畫面、
  回合摘要與 INFO 摘要共用 `eventReportMessageText`；投降模板由 `ui.json` 提供，AI 名稱依
  `RaceIndex` 顯示當前語言，玩家／熱座自訂名保留，非法 target 安全回退未知帝國。其餘尚未
  遷移事件與舊存檔仍可沿用 `Message`／`MessageEN`，未以全域刪欄破壞相容性。原版已證實
  pending→事件 34→延後資產接收順序；本輪未改自動投降近似 gate 或資產 consumer，通知逐字
  仍明標 remake adapter。聚焦測試覆蓋 AI→AI、AI→玩家、AI→熱座、非法 target、舊存檔 fallback、
  pending／延後轉移與來源碼防回歸；繁中與英文正版資料畫廊各重跑 38/38。正常畫廊未強制
  觸發投降，故只證明事件路由無回歸，事件 34 實際文案由 typed 測試覆蓋。證據與規格見
  [`docs/re/empire-surrender-notice-text-audit-20260827.md`](docs/re/empire-surrender-notice-text-audit-20260827.md) 與
  [`docs/spec/empire-surrender-notice-external-text.md`](docs/spec/empire-surrender-notice-external-text.md)。
  帝國滅亡事件 29 已完成第六十一個切片：`detectEmpireEliminationBroadcasts` 在 active 帝國失去
  最後殖民地時只排入原版事件 ID 與 typed target，不再建立中英文成品句子；既有可存檔
  `EmpireAlive` 去重狀態、清理順序與事件佇列未改。共用 `eventReportMessageText` 現以
  `event.status.empire_eliminated` JSON 模板顯示；AI 依 `RaceIndex` 翻譯，玩家／熱座保留自訂名，
  非法 target 與舊存檔均有安全 fallback。原版 `sub_E4EB3`→`sub_233AB` 的 active→inactive
  觸發與單帝國 record 已證實，但四種隨機原文尚未復原，單一摘要仍明標 remake adapter。
  聚焦測試覆蓋 AI、玩家、熱座、非法 target、舊存檔、成品欄為空及不重播；繁中與英文正版
  資料畫廊各重跑 38/38。正常畫廊未強制消滅帝國，故只證明事件路由無回歸，實際事件 29 文案
  由 typed 測試覆蓋。證據與規格見
  [`docs/re/empire-elimination-notice-text-audit-20260827.md`](docs/re/empire-elimination-notice-text-audit-20260827.md) 與
  [`docs/spec/empire-elimination-notice-external-text.md`](docs/spec/empire-elimination-notice-external-text.md)。
  本批證據與規格見
  [`docs/re/strategic-combat-result-text-audit-20260827.md`](docs/re/strategic-combat-result-text-audit-20260827.md) 與
  [`docs/spec/strategic-combat-result-external-text.md`](docs/spec/strategic-combat-result-external-text.md)；拓殖／前哨證據與規格見
  [`docs/re/colonization-outpost-result-audit-20260827.md`](docs/re/colonization-outpost-result-audit-20260827.md) 與
  [`docs/tech/colonization-outpost-result-text-spec.md`](docs/tech/colonization-outpost-result-text-spec.md)；星圖證據與規格見
  [`docs/re/galaxy-star-panel-text-audit-20260827.md`](docs/re/galaxy-star-panel-text-audit-20260827.md) 與
  [`docs/spec/galaxy-star-panel-external-text.md`](docs/spec/galaxy-star-panel-external-text.md)。
  其餘自繪畫面仍待逐批遷移；通用規格見
  [`docs/spec/external-player-text.md`](docs/spec/external-player-text.md)，本批證據與規格見
  [`docs/re/netinfo-text-contract-audit-20260826.md`](docs/re/netinfo-text-contract-audit-20260826.md) 與
  [`docs/spec/netinfo-external-text.md`](docs/spec/netinfo-external-text.md)、
  [`docs/re/custom-race-ui-text-audit-20260826.md`](docs/re/custom-race-ui-text-audit-20260826.md) 與
  [`docs/spec/custom-race-external-text.md`](docs/spec/custom-race-external-text.md)、
  [`docs/re/name-banner-flow-audit-20260826.md`](docs/re/name-banner-flow-audit-20260826.md) 與
  [`docs/tech/name-banner-external-text-spec.md`](docs/tech/name-banner-external-text-spec.md)、
  [`docs/re/info-subscreen-text-audit-20260826.md`](docs/re/info-subscreen-text-audit-20260826.md) 與
  [`docs/tech/info-subscreen-external-text-spec.md`](docs/tech/info-subscreen-external-text-spec.md)、
  [`docs/re/build-queue-ui-text-audit-20260826.md`](docs/re/build-queue-ui-text-audit-20260826.md) 與
  [`docs/tech/build-queue-external-text-spec.md`](docs/tech/build-queue-external-text-spec.md)、
  [`docs/re/colony-screen-ui-text-audit-20260826.md`](docs/re/colony-screen-ui-text-audit-20260826.md) 與
  [`docs/tech/colony-screen-external-text-spec.md`](docs/tech/colony-screen-external-text-spec.md)、
  [`docs/re/event-screen-ui-text-audit-20260826.md`](docs/re/event-screen-ui-text-audit-20260826.md) 與
  [`docs/tech/event-screen-external-text-spec.md`](docs/tech/event-screen-external-text-spec.md)、
  [`docs/re/hi-score-screen-ui-text-audit-20260826.md`](docs/re/hi-score-screen-ui-text-audit-20260826.md) 與
  [`docs/tech/hi-score-external-text-spec.md`](docs/tech/hi-score-external-text-spec.md)、
  [`docs/re/ground-combat-screen-ui-text-audit-20260826.md`](docs/re/ground-combat-screen-ui-text-audit-20260826.md) 與
  [`docs/tech/ground-combat-external-text-spec.md`](docs/tech/ground-combat-external-text-spec.md)、
  [`docs/re/colony-bombing-screen-ui-audit-20260826.md`](docs/re/colony-bombing-screen-ui-audit-20260826.md) 與
  [`docs/tech/colony-bombing-external-text-spec.md`](docs/tech/colony-bombing-external-text-spec.md)、
  [`docs/re/load-save-popup-ui-text-audit-20260826.md`](docs/re/load-save-popup-ui-text-audit-20260826.md) 與
  [`docs/tech/load-save-external-text-spec.md`](docs/tech/load-save-external-text-spec.md)、
  [`docs/re/command-points-screen-ui-text-audit-20260826.md`](docs/re/command-points-screen-ui-text-audit-20260826.md) 與
  [`docs/tech/command-points-external-text-spec.md`](docs/tech/command-points-external-text-spec.md)、
  [`docs/re/game-menu-popup-ui-text-audit-20260826.md`](docs/re/game-menu-popup-ui-text-audit-20260826.md) 與
  [`docs/tech/game-menu-external-text-spec.md`](docs/tech/game-menu-external-text-spec.md)、
  [`docs/re/tactical-control-bar-text-audit-20260826.md`](docs/re/tactical-control-bar-text-audit-20260826.md) 與
  [`docs/tech/tactical-control-bar-external-text-spec.md`](docs/tech/tactical-control-bar-external-text-spec.md)、
  [`docs/re/research-choice-ui-text-audit-20260826.md`](docs/re/research-choice-ui-text-audit-20260826.md) 與
  [`docs/tech/research-choice-external-text-spec.md`](docs/tech/research-choice-external-text-spec.md)、
  [`docs/re/refit-ui-text-audit-20260827.md`](docs/re/refit-ui-text-audit-20260827.md) 與
  [`docs/tech/refit-external-text-spec.md`](docs/tech/refit-external-text-spec.md)、
  [`docs/re/input-box-ui-text-audit-20260827.md`](docs/re/input-box-ui-text-audit-20260827.md) 與
  [`docs/tech/input-box-external-text-spec.md`](docs/tech/input-box-external-text-spec.md)。程式註解、測試文字與除錯日誌不列入玩家文案。

- [x] **原版對局內 SETTINGS 分頁**：2026-08-26 已完成原版 13 列畫面、資產、外部雙語文案、
  原版預設值、`.GAM` 匯入與 JSON 往返。IDA 證實
  `Do_Options_Game_Popup_ @ 0x7E00F..0x7E154` 與 `_Draw_Options_Game_Popup_ @ 0x7FA28..0x8011F`
  由 `_Game_Popup_ @ 0x8012F` 的四路 switch 直接進入，並已追回
  `sub_7FA28` 的 13 欄、文字 ID、17 px 列距、設定位元、`sub_7EFEF` 載入、
  `Update_Game_Settings_ @ 0x7F14C` 回寫及 `sub_127E1` 預設。remake 已接遷移線、自動存檔、
  結局動畫、回合摘要，以及 Auto Delete Trade Goods／Housing 的原版建造佇列確認／清除消費端。
  Ship Initiative 已由 IDA 證實開啟時合併排序、關閉時依陣營分批；remake 已接快速結算
  與格子戰術的敵我全域穩定排序、關閉時雙方分批、格子戰術 WAIT／DONE／AUTO，以及能量
  吸收器的一回合期限。格子戰術使用單場暫態 ID，戰損壓縮後不會把待行動項錯接到另一艘船；
  原版 seeking-missile 逐 tick 狀態維持非阻塞證據邊界。規格與證據見
  `docs/tech/ship-initiative-settings-spec.md`、`docs/re/game-menu-popup-ui-text-audit-20260826.md`。
  Auto Select Ships 已由 `byte_199BE1 @ 0x199BE1` 唯一玩法讀取端、`sub_70875 @ 0x70875`
  與逐艦選取 writer `sub_7229E @ 0x7229E` 閉合：開啟時進入或切換艦隊會選取目前艦隊全部艦艇，
  關閉時保持空集合；玩家全不選後重繪不會被強制選回，拆分後亦依新索引重建。
  Auto Select Colony 已由 `byte_199BE3 @ 0x199BE3`、星系索引重建
  `sub_123CE @ 0x123CE`／`sub_7862B @ 0x7862B`、三條點選消費端，以及手動巡覽清除 writer
  `sub_876DB @ 0x876DB`／`sub_87720 @ 0x87720` 閉合。remake 開啟時點我方殖民星會直接進入
  對應殖民地畫面；F5／F6 明確手動巡覽會關閉此設定。原版星系內多圖示堆疊在 remake
  映射為單一殖民地入口，明標介面轉接近似；證據與規格見
  `docs/re/auto-select-colony-setting-audit-20260827.md`、
  `docs/tech/auto-select-colony-setting-spec.md`。
  Enemy Moves 已由 `byte_199BDF @ 0x199BDF` 的完整直接 xref 稽核、原版 help 契約與兩個
  敵方在途資料 consumer 釐清證據邊界：原版設定 byte 沒有已閉合的直接玩法讀取端，因此
  remake 不臆造 AI 規則，只在開啟時把起點與目的地皆可見的在途 AI 艦隊以星圖航線及移動
  marker 顯示；關閉時不畫，也不洩漏霧區目的地。線色與 timing 明標視覺近似；typed 查詢、
  純幾何、狀態指紋與正常畫廊均有測試。證據與規格見
  `docs/re/enemy-moves-setting-audit-20260827.md`、`docs/tech/enemy-moves-setting-spec.md`。
  Expanding Help 已由 `byte_199BE0 @ 0x199BE0` 的多個消費端與共同 renderer
  `sub_83EFD @ 0x83EFD..0x84356` 閉合：原版開啟時以 10 步插值展開說明面板，關閉時立即顯示。
  remake 的 SETTINGS 每列已接右鍵情境說明，依目前暫存設定選擇十步展開或立即顯示；標題與
  本文全部位於 `assets/i18n/ui.json`，安全框換行與純幾何回歸測試已通過。這是正常玩家路徑的
  介面轉接近似，不宣稱已逐一映射原版所有 help hotspot；證據與規格見
  `docs/re/expanding-help-setting-audit-20260827.md`、`docs/tech/expanding-help-setting-spec.md`。
  Show GNN Report 已由 `byte_199BE5 @ 0x199BE5`、事件 switch `sub_21371 @ 0x21371`、
  文字組裝鏈 `sub_21B6D @ 0x21B6D` 與官方 help 契約閉合：開啟時以 GNN 畫面中斷報告，
  關閉時略過主播畫面但仍把特殊事件列入一般回合摘要。remake 已依此修正結算後路由；即使
  End Of Turn Summary 同時關閉，特殊事件也不會靜默消失。星系勘查是玩家自家回報，不受 GNN
  選項抑制；同回合事件則保留於摘要。證據、規格與組合測試見
  `docs/re/show-gnn-report-setting-audit-20260827.md`、`docs/tech/show-gnn-report-setting-spec.md`。
  Serious Summary 已由唯一玩法 consumer `sub_FE0EA @ 0xFE0EA..0xFE250` 閉合：它逐筆掃描
  18-byte 回合報告記錄，開啟時只有 serious raw type／subtype 能觸發整張摘要；不是在已開啟
  的摘要裡刪除一般行。remake 已接 typed gate：官方 help 明列的飢荒、叛亂、破產處分會觸發，
  安塔蘭與敵方突襲是明標強推論的威脅擴充；一般經濟、研究與建造完成不觸發。GNN 關閉後的
  特殊事件通知優先保留，從快報按繼續與直接結算共用同一 gate。飢荒／叛亂計數文案已外置於
  `ui.json`。證據與規格見 `docs/re/serious-turn-summary-setting-audit-20260827.md`、
  `docs/tech/serious-turn-summary-setting-spec.md`。
  End Of Turn Wait 已由 `byte_199BDD @ 0x199BDD`、連續回合入口 `sub_8AD82 @ 0x8AD82`、
  全畫面中止輸入 `sub_83411 @ 0x83411`、中斷 writer `sub_84E9D @ 0x84E9D` 與官方 help
  契約閉合。remake 在單人局關閉此設定時，會以固定 15 tick 的介面時序近似連續推進；任一滑鼠
  點擊、研究選擇、勝負、GNN／勘查快報或回合摘要都會停止。熱座與網路鎖步不啟用連續推進，
  避免越過玩家席位與同步閘門。證據、規格與回歸測試見
  `docs/re/end-of-turn-wait-setting-audit-20260827.md`、
  `docs/tech/end-of-turn-wait-setting-spec.md`。至此 13 列設定皆有正常玩家路徑消費端；
  DOS／Win95 平台 API 內部維持既定停止線。

- [x] **共用知識庫防錯閘門**：`~/.codex/knowledge-base/local/retro-remake-gameplay-parity-audit.md`
  已把本次錯判提煉成跨專案流程，涵蓋重新稽核觸發條件、具名符號限制、玩家機制證據矩陣、
  IDA 深抽樣、自洽測試與原版 oracle 分級，以及 README／發行聲明閘門；知識路由與 IDA Docker
  衛生條目已加入按需入口。MOO2 的公式與位址仍只留在本專案證據，不污染通用知識。
- [~] **建立玩家機制—IDA—Go 證據矩陣**：第二輪已由 14 列擴充為 31 列；一次性 IDA Pro 9.4 已確認原版 5,397 個函式；內建
  符號表有 4,201 個函式名稱，但名稱不等於資料流已研究。本輪高影響抽樣證實
  編譯器／runtime helper 已另建排除索引：DOS 版證實為 Watcom C/C++ 32-bit＋DOS/4GW，
  Win95 版證實為 Microsoft Visual C++ runtime 家族（PE linker 4.20）；stack probe、x87 初始化、
  C++／SEH 例外處理不列為玩法缺口。精確編譯器小版本仍為未知，且舊 `symbols_ea.tsv` 的 runtime
  區域錯位已記錄勘誤；後續矩陣只計玩家可見 consumer，詳見
  `docs/re/compiler-runtime-fingerprint-20260827.md`。
  `Calc_Council_Vote_ @ 0x15B90` 為 `ceil(population/10)`，Go 已於 2026-08-24 依 IDAPython
  caller／consumer 證據修正；
  `Check_For_Council_Meeting_ @ 0x168AF` 已有原版排程與門檻，Go 卻以手冊未寫為由自訂 8 回合；
  `Strategic_Bombardment_ @ 0x4257E` 固定走三輪原版快速戰鬥鏈；Go 已於 2026-08-24 修正
  原先的 5/10 全武器抽象模型，5/10 現只作用於炸彈攻擊當量。
  第二輪另確認 `Next_Turn_Calc_` 含 55 個直接 call，事件、間諜、艦隊航行、殖民地完整回合、
  研究突破、安塔蘭週期入侵與原版 AI 等此前未完整入矩陣；`Determine_Event_` 的一般排程已於
  2026-08-25 閉合並取代 30% 自訂值；同日已接全局帝國目標與熱座／AI 基本回寫，並由
  `sub_2230A`／`sub_206A2` 閉合富商捐獻的每 20 回合 100 BC 階梯，以及海盜劫掠的
  國庫至少 100 BC、30..50% 向下取整公式；目前玩家、熱座與 AI 已共用同一消費端。
  同日另由事件建立端 `0x229CC..0x229F7` 與消費端 `0x2091D..0x209B2` 閉合電腦病毒：
  研究進度至少 10 RP 才適用，損失為 `min(progress, Random(50)+50)`；玩家、熱座與 AI
  已共用同一可存檔亂數消費端。事件 0 古代科技亦由 `sub_58853` 的全銀河光束／護盾
  application 選擇、事件 record 兩槽與 `sub_E4204` 雙授予 consumer 閉合；玩家、熱座與 AI
  現在授予一至兩項明確科技並更新設計，不再自訂增加 RP。外交、殖民地與持續事件等 AI
  複合效果仍未完整；事件 1 氣候改善已由 `sub_2310C`／`sub_23D44` 的 reservoir＋200 次
  重試目標鏈與 `sub_206A2 @ 0x207AC..0x2083E` 消費端閉合：Toxic..Arid 可直接變 Terran，
  玩家、熱座與 AI 會同步 colony／planet 並重算食物與容量，不再誤走一般 Terraforming 階梯。
  事件 11／12 礦產亦已由 `sub_2325E`／`sub_232BB` 與 `sub_206A2` 閉合：枯竭只選
  Ultra Rich 並作 `4→3`，發現作 `min(4, old+2)`；玩家、熱座與 AI 會同步 planet／colony
  與每工人產能，且不再錯誤重算重力。枯竭已經 `sub_23DA0` 只在無 Capitol 候選間抽樣；
  發現維持 `sub_23D44` 的全殖民地候選，兩條不再共用錯誤的泛用 selector。
  事件 17 人口暴增亦已由 `sub_2230A`、`sub_206A2`、`sub_23509` 與 `sub_E1839` 閉合：
  不再一次性自訂 `+2` 人口，而是目標殖民地逐族成長加 100 百分點；前五個 active turn
  不結束，第六回合起每回合 1/20，age > 20 強制結束。玩家、熱座與 AI 均走可存檔持續 record。
  事件 16 瘟疫亦已由 `sub_23B64`、`sub_206A2`、`sub_234B8` 與 `sub_E1839` 閉合：不再
  一次性自訂扣兩人口，而是逐族成長扣 200 百分點；初始治療需求為建立時殖民地 RP 乘上
  `Random(8)+2×難度`，之後由實際研究產出逐回合清償。玩家、熱座與 AI 均接持續 record。
  事件 7 地震亦已由 `sub_2230A`、`sub_238A8`、`sub_DD2F2` 與 `sub_DCEBD` 閉合：不再
  固定自訂扣 1～2 人口並拆一棟，而是以
  `max(1, (人口+raw 建築數)×(Random(3)+Random(2))/10)` 產生傷害，再由一般建築、
  陸戰隊、戰車、建造進度與人口的共用戰略候選池分配。玩家、熱座與 AI 均回寫傷亡，
  最後人口死亡會移除殖民地與平行陣列；`sub_23DA0` 的無 Capitol 候選條件已接，packed
  colonist 死亡身分仍明標近似。
  事件 10 工業事故亦已由 `sub_231B4`、`sub_23833`、`sub_DD13E`、`sub_DD2F2` 與
  `sub_DCEBD` 閉合：環境耐受帝國直接免疫，只選 Barren..Gaia 殖民地；特殊命中數為
  `人口×(Random(3)+Random(3))/10` 且可為 0，逐次排除 Android 人口並在人口／陸戰隊／
  戰車間分配，最後仍固定結算 1 點一般戰略殖民地傷害。玩家、熱座與 AI 均已回寫人口群組、
  駐軍、建築效果、建造進度與殖民地平行陣列；氣候不再被錯誤降級。typed 群組不保存原版
  packed 順序，故只宣稱候選集合與分布對齊；事件目標已排除完成 Capitol 的殖民地。
  事件 2 彗星亦已由 `sub_2230A`、`sub_206A2`、`sub_23B28`、`sub_23780` 與
  `sub_DD2F2` 閉合：耐久為 `10×(Random(5)+10+難度)`、倒數為
  `Random(5)+10-難度`；目標星系所有停泊艦艇不分 owner 逐艘貢獻 `艦體級+1`，撞擊時以
  `max(1,(人口+raw 建築數)×(Random(3)+Random(3))/10)` 走共用戰略殖民地傷害鏈。
  玩家、熱座與 AI 均有可存檔持續 record、攔截及撞擊回寫，不再因缺 case 永遠抽不到。
  remake 以 ETA==0 對應 raw ship status 0；事件 14 海盜活動亦已由 `sub_2230A`、
  `sub_23BEC`、`sub_2448F`、`sub_206A2`、`sub_23B28` 與 `sub_242FC` 閉合：初始強度依
  同星系各帝國運輸船總數及五級難度換算，低於 5 不成立；每回合按剩餘強度百分比讓同星
  帝國各損失一艘運輸船，再由所有 owner 的停泊艦艇以 `艦體級+1` 清剿。玩家、熱座與 AI
  共用可存檔 record，並已補齊事件 2／16／17／25／14／24 的同星互斥。remake 以殖民地
  presence 重建 raw `star+0x38` bitset；事件 2／7／10／11／16 共用的 `sub_23DA0`
  無 Capitol 候選條件已接到玩家、熱座目前席位與 AI。
  事件 9 超空間亂流亦已由 `sub_2230A`、`sub_206A2`、`sub_233FA` 及航行／AI callers
  閉合：全銀河 record 建立時 age=0，第六次 consumer 起每回合 `1/20` 解除，age>20
  強制結束；玩家、熱座與 AI 的既有航程及新派遣均會凍結，跨維度種族依 raw
  `player+0x8BC` 分支免疫。record、雙語訊息與 JSON 往返已接；1.31／1.50 對彗星／怪獸
  候選的亂流衝突方向有官方 changelog 差異，仍待版本 profile 閉合，不冒稱共同規則。
  事件 18 秘密實驗亦已由 `sub_206A2 @ 0x20ECA..0x20F64`、`sub_E4410` 與
  `sub_21371 @ 0x2171D..0x21730` 閉合：事件保存目標帝國目前研究 field、立即走既有
  application／Creative／Hyper-Advanced 完成鏈，隨後清空研究進度與 field；玩家、熱座與
  AI 均接科技 callback 及艦型更新。舊 `80 + Turn` RP 加值已移除，無 field 時不再捏造科技。
  事件 8 艦船爆炸亦已由 `sub_23CED`、`sub_206A2 @ 0x20AD7..0x20B61`、
  `sub_941C6` 與 `sub_A163A` 閉合：目標帝國 active 艦艇以 reservoir sampling 等機率選一艘，
  只移除該艦；有指派軍官時軍官一併死亡。玩家、熱座與 AI 已接線，舊戰鬥引擎爆炸勢能、
  三艘鄰艦傷害與「保留最後一艘」代理均已移除。
  事件 13 艦艇叛變則由原始跳表與完整寫入端推翻舊 remake：1.31 建立端把只在事件
  4／5、8、27 賦值的未初始化區域變數寫入 record，consumer 與新聞 dispatcher 均無 case 13
  效果，事件模組亦沒有艦艇 owner 回寫 caller。現改為玩家、熱座與 AI 都只產生不具名播報，
  不再自創選艦、移交 AI 或狀態變更；1.50 二進位未取得，該版本仍標未知。
  事件 4／5 外交暗殺／聯姻亦已閉合：`sub_23B7D` 先在存活已接觸帝國間 reservoir 抽第二帝國，
  caller 再要求正式狀態 4／5；抽到和平對象時候選直接失敗而不重抽。consumer 以 raw
  `-100／+100` 呼叫 `Change_Relations_ @ 0x4E3B5`。2026-08-27 完整控制流勘誤證實：該函式
  讀取第二帝國對受害帝國的反向正式狀態，值 `>=4` 時在 `0x4E75C..0x4E764`
  直接返回；舊述「套政體／Charismatic 後夾到 -25」對對稱戰爭路徑不可達。Go 已移除
  自訂 `±12` 與這個假效果；玩家／AI 與 AI／AI 事件仍播報，但保留現有關係分數。
  熱座真人彼此外交矩陣與原版方向不對稱條約仍是資料模型缺口。
  事件 27 曲速漏斗亦已由 `sub_23CED`、`sub_2230A @ 0x22D2F`、
  `sub_206A2 @ 0x212BE`、`sub_21371 @ 0x21902` 與完整
  `Move_All_Ships_Toward_Stars_ @ 0xFFEEA` 閉合：1.31 保存均勻抽中的 active 艦艇索引，
  第五年齡後每回合 `1/20` 脫困且 `age>20` 強制解除，但事件 consumer 與艦隊移動函式
  都沒有停航回寫。Go 現以玩家／熱座／AI 報告型 persistent record 重製，不再把舊
  「艦隊受困」文案誤讀成需要自行凍結 ETA；1.50 二進位未取得，該版本仍標未知。
  證據與規格見
  [`docs/re/random-event-bc-effects-audit-20260825.md`](docs/re/random-event-bc-effects-audit-20260825.md) 與
  [`docs/spec/random-event-bc-effects.md`](docs/spec/random-event-bc-effects.md)、
  [`docs/re/random-event-computer-virus-audit-20260825.md`](docs/re/random-event-computer-virus-audit-20260825.md) 與
  [`docs/spec/random-event-computer-virus.md`](docs/spec/random-event-computer-virus.md)、
  [`docs/re/random-event-ancient-tech-audit-20260825.md`](docs/re/random-event-ancient-tech-audit-20260825.md) 與
  [`docs/spec/random-event-ancient-tech.md`](docs/spec/random-event-ancient-tech.md)、
  [`docs/re/random-event-minerals-audit-20260825.md`](docs/re/random-event-minerals-audit-20260825.md) 與
  [`docs/spec/random-event-minerals.md`](docs/spec/random-event-minerals.md)、
  [`docs/re/random-event-population-boom-audit-20260825.md`](docs/re/random-event-population-boom-audit-20260825.md) 與
  [`docs/spec/random-event-population-boom.md`](docs/spec/random-event-population-boom.md)、
  [`docs/re/random-event-plague-audit-20260825.md`](docs/re/random-event-plague-audit-20260825.md) 與
  [`docs/spec/random-event-plague.md`](docs/spec/random-event-plague.md)、
  [`docs/re/random-event-earthquake-audit-20260825.md`](docs/re/random-event-earthquake-audit-20260825.md) 與
  [`docs/spec/random-event-earthquake.md`](docs/spec/random-event-earthquake.md)、
  [`docs/re/random-event-industrial-accident-audit-20260825.md`](docs/re/random-event-industrial-accident-audit-20260825.md) 與
  [`docs/spec/random-event-industrial-accident.md`](docs/spec/random-event-industrial-accident.md)、
  [`docs/re/random-event-comet-audit-20260825.md`](docs/re/random-event-comet-audit-20260825.md) 與
  [`docs/spec/random-event-comet.md`](docs/spec/random-event-comet.md)、
  [`docs/re/random-event-pirate-activity-audit-20260825.md`](docs/re/random-event-pirate-activity-audit-20260825.md) 與
  [`docs/spec/random-event-pirate-activity.md`](docs/spec/random-event-pirate-activity.md)、
  [`docs/re/random-event-hyperspace-flux-audit-20260825.md`](docs/re/random-event-hyperspace-flux-audit-20260825.md) 與
  [`docs/spec/random-event-hyperspace-flux.md`](docs/spec/random-event-hyperspace-flux.md)、
  [`docs/re/random-event-secret-experiment-audit-20260825.md`](docs/re/random-event-secret-experiment-audit-20260825.md) 與
  [`docs/spec/random-event-secret-experiment.md`](docs/spec/random-event-secret-experiment.md)、
  [`docs/re/random-event-ship-explosion-audit-20260825.md`](docs/re/random-event-ship-explosion-audit-20260825.md) 與
  [`docs/spec/random-event-ship-explosion.md`](docs/spec/random-event-ship-explosion.md)、
  [`docs/re/random-event-ship-mutiny-audit-20260825.md`](docs/re/random-event-ship-mutiny-audit-20260825.md) 與
  [`docs/spec/random-event-ship-mutiny.md`](docs/spec/random-event-ship-mutiny.md)、
  [`docs/re/random-event-diplomatic-incidents-audit-20260825.md`](docs/re/random-event-diplomatic-incidents-audit-20260825.md) 與
  [`docs/spec/random-event-diplomatic-incidents.md`](docs/spec/random-event-diplomatic-incidents.md)、
  [`docs/re/random-event-warp-funnel-audit-20260825.md`](docs/re/random-event-warp-funnel-audit-20260825.md) 與
  [`docs/spec/random-event-warp-funnel.md`](docs/spec/random-event-warp-funnel.md)、
  [`docs/re/random-event-climate-audit-20260825.md`](docs/re/random-event-climate-audit-20260825.md) 與
  [`docs/spec/random-event-climate.md`](docs/spec/random-event-climate.md)；
  安塔蘭的舊 20／15 直接傷害腳本已於 2026-08-25 移除；目前剩餘 owner 8 精確戰鬥 record
  、殖民地固定防禦／戰後消費與中途座標限制另列於艦隊／安塔蘭條目。完整性審查見
  [`docs/re/re-completeness-review-20260813.md`](docs/re/re-completeness-review-20260813.md)，第一輪公式反例見
  [`docs/re/parity-re-audit-20260812.md`](docs/re/parity-re-audit-20260812.md)，
  可機械搜尋的逐系統列見 [`docs/re/parity-matrix.tsv`](docs/re/parity-matrix.tsv)。
- [x] **清除現況文件與 source 過期斷言**（2026-08-24；2026-08-27 再稽核）：
  `docs/HONEST-STATUS.md` 已濃縮為目前可證實／近似／未知邊界；重複且已漂移的玩法狀態表已移除，不再宣稱經濟、
  兩條戰鬥路徑或全畫面已完全 parity；同步修正客製種族、AI 研究、飛彈防禦、突擊艇
  與 AI 種族殖民的過期 source 註解。2026-08-27 另掃描 424 份專案 Markdown，刪除已由活表／
  parity 矩陣取代的 `HANDOFF`、完成度評估、剩餘工作路線圖、玩法狀態表與舊文件稽核報告；
  同步修正舊儀表板分數、8 月 12 日打包「最新」措辭、`.GAM` importer、敵方戰機命中／傷害及
  武器／食物舊值。歷史翻案只保留可回查的正確證據與 Git，不再保留錯誤句子副本。
- [x] **新遊戲規則與開局忠實化**：2026-08-24 已重審 RE 並建立
  [`docs/spec/newgame-opening-rules.md`](docs/spec/newgame-opening-rules.md)。開局建築已依原版優先表、
  `3/5/9` 上限與 `min(ceil(2/3 population), cap)` 接線；先進級已執行固定六次加
  十九次決定性選擇，並以科技應用粒度寫入 `ChosenTech`／`ExplicitChoice`；
  1.31／1.50 開局與存讀檔各有垂直測試。已移除 `0.3/0.6/1.0/1.5/2.2` 通用浮點
  倍率及其對敵艦代理與外交關係的不當縮放；難度仍由已證實的離散 helper 消費。
  2026-08-25 已由 IDA Pro 閉合 `sub_FD335` 的應用級單次抽選與 `sub_FC845` 人類／共用估值鏈，
  並接入玩家正常開局；同日再閉合 `sub_589D6` raw 6／4／7 加權初始化、
  `sub_FC845` AI category／種族分支及 `sub_FD335` 難度門檻，AI 開局不再使用類別權重近似。
  `sub_12983` 亦已證實先轉換種族特性 1..9，後呼叫 `sub_589D6`；特性 6 的
  20／40 比較是原始 runtime 分支，現行 25／50 表不會觸發，不再列為時序矛盾。
  正式 UI 原本在種族／政府之前發先進級科技的順序錯誤亦已修正：現在以一次性
  pending 門在最終種族完成後重建開局研究，真人 raw4 也送入共用估值後段；
  客製種族完整 31 格 runtime 特性已可經 JSON／熱座往返，內建種族政體不再全部誤用獨裁。
  原版 AI 經濟／建造／常態研究主題／艦隊／外交的逐子系統難度分支分別留在 AI 與經濟
  活表，不再重複阻擋開局條目，也不以新的通用倍率代替。新增證據見
  [`docs/re/starting-tech-application-audit-20260825.md`](docs/re/starting-tech-application-audit-20260825.md)、
  [`docs/re/ai-starting-tech-profile-audit-20260825.md`](docs/re/ai-starting-tech-profile-audit-20260825.md)，原開局審查見
  [`docs/re/human-starting-tech-profile-audit-20260825.md`](docs/re/human-starting-tech-profile-audit-20260825.md)，
  [`docs/re/newgame-opening-audit-20260824.md`](docs/re/newgame-opening-audit-20260824.md)。
- [~] **回合、殖民地、研究與帝國經濟忠實化**：由 `Next_Turn_Calc_ @ 0x136B3` 固定 call graph
  對回 `Apply_Colony_Pop_Growth_ @ 0xE2DCA`、`Do_Colony_Calculations_ @ 0xE2B31`、
  `Apply_All_Colony_Changes_ @ 0xE3FDC`、`Colony_Research_Production_ @ 0xDFF74`、
  `Check_For_Research_Breakthrough_ @ 0xE44E0`、`Compute_Player_Maintenance_ @ 0xE2000` 與
  `Player_Maintenance_ @ 0xEE0B0`。需閉合人口職務／成長、食物、工業、污染、士氣、建造、
  同化、研究進度／突破／溢出／Creative／Uncreative、建築／艦隊／間諜／領袖／指揮點維護、
  負國庫及回寫順序。2026-08-24 已依 patch 1.50 官方手冊閉合人口點數尺度：每 1,000 點
  兌換一人口、複製中心固定 +100 點，移除 `popGrowthThreshold=300` 調校值；999／1,000 邊界、
  快照餘數、30 回合成長與上限測試均已通過。仍待 IDA 證據：1.31 常數、AI 難度項、
  事件／職務重排與整體殖民地回寫順序。證據見
  [`docs/re/population-growth-scale-audit-20260824.md`](docs/re/population-growth-scale-audit-20260824.md)，
  規格見 [`docs/spec/population-growth-scale.md`](docs/spec/population-growth-scale.md)。
  2026-08-25 已由 IDA Pro 閉合 `Compute_Player_Maintenance_` 的六分項與加總，並證實
  `Player_Maintenance_` 是負國庫資產處分鏈而非另一套公式；間諜每名 1 BC、軍官維護費已
  改在單次帝國經濟結算扣除，不再於後段重複／夾額扣款。納貢維持既有跨帝國轉帳以避免重扣。
  同日再閉合三輪建築篩選、最低價值候選、半價退款、間諜裁撤與領袖解雇，已接入正常玩家／
  熱座回合、可逆建築效果與回合摘要。後續勘誤確認 `sub_EDDF7` 是強制拆船而非放棄殖民地；
  艦種、友軍星、mission `{0,3,7}`、呼叫順序與設計成本四分之一／零退款已接線。戰鬥艦最低
  效能候選目前用 remake 戰力近似；`player+0x59D`／`sub_ED908` 已定位為 AI 晚期科技鏈的
  四棟研究設施清理，只在重建完整 AI 經濟時再接，不列作玩家破產缺口。證據見
  [`docs/re/player-maintenance-audit-20260825.md`](docs/re/player-maintenance-audit-20260825.md)，規格見
  [`docs/spec/player-maintenance.md`](docs/spec/player-maintenance.md) 與
  [`docs/spec/player-bankruptcy.md`](docs/spec/player-bankruptcy.md)。
  2026-08-25 再由 `sub_DE280`／`sub_DE22C`／`sub_DDF2C` 閉合逐人口產出：raw bit 9 是
  WORKING、raw bit 10 是 PRISONER，後者按實際職務逐人扣 `5/20`，不是按殖民地總人口比例
  猜分。Go 已保存逐職務未同化人口，`.GAM` 可精確匯入，征服／改派／同化與 JSON 均可往返；
  舊 JSON 才回退比例法。原版同化選哪一筆 packed colonist 仍是可重播近似，逐 race 基礎產出、
  重力天賦及 loyalty 非產出消費端仍待後續切片。證據見
  [`docs/re/colonist-prisoner-production-audit-20260825.md`](docs/re/colonist-prisoner-production-audit-20260825.md)，規格見
  [`docs/spec/colonist-prisoner-production.md`](docs/spec/colonist-prisoner-production.md)。
  同日亦由 `Check_For_Research_Breakthrough_ @ 0xE44E0` 閉合研究突破：累積值必須嚴格超過
  成本，突破率為超額百分比 1..100，只有本回合正研究產出才擲 `Random(100)`；失敗保留全部
  累積值，成功清零而不結轉溢出。Go 已接玩家、熱座與 AI 的可存檔研究亂數流。
  `Colony_Research_Production_` 亦已閉合研究人口 dispatch 與四棟研究建築消費端：研究實驗室、
  行星超級電腦、銀河網路中心、自動實驗室只固定增加 5／10／15／30 RP，不會再額外修改
  每位科學家產出；舊手冊摘要造成的重複加成已修正，建造與破產拆除均有垂直測試。仍待
  玩家 `sub_10DC12`、AI `sub_DC288`、Creative／Uncreative 選項形成與 `sub_E4410` 授予時序
  已閉合並接線：application 在投入研究前選定，突破後才寫入已擁有科技。`sub_E4204` 的特殊
  application 分支亦已稽核：Battleoids 連帶授予 Armor Barracks，並以額外科技集合保留同主題
  原選擇；政府、星系遮罩、殖民地／玩家與曲速欄位則是原版 raw 快取重算，remake 沿用既有
  動態消費端。玩家、AI、熱座、偷竊／餽贈、遺跡與開局授予均走共用 callback。逐人口產出方面，
  2026-08-25 已由 `sub_DDF2C`／`sub_DDFD3` 證實重力碼 4／3／2 對應 100%／75%／50%，
  並接上玩家、客製種族、AI、熱座與 `.GAM` 的所有者 Low-G／Normal-G／High-G；舊 JSON
  未知安全回退 Normal-G，重力產生器仍優先歸零。2026-08-25 後續已補 typed population
  groups：確認 packed 低四位是當局 player slot 而非十三族 OrigIdx，`.GAM` 依 slot/job/prisoner
  聚合並注入各 player 的 `+0x8A1/+0x8A2/+0x8A3`、Low／High-G 與 Aquatic profile；引擎改為
  逐群計算，改派、征服、同化、一般成長與玩家事件／轟炸傷亡會同步維持群組，JSON 可往返。
  後續 IDA 已閉合 Android `+6/+3/+3`、Natives `+4/+0/+0` 與兩者重力免疫，並證實
  `colony+0xC8[slot]` 逐槽成長率、`+0xB4[slot]` 逐槽累積池及 1,000 點新增相同 slot 人口；
  玩家、AI、`.GAM` 與 JSON 均已接線。負成長亦已由 `sub_DEB4B`／`sub_DF546`／`sub_E1839`／
  `sub_E2DCA` 閉合：Cybernetic／Lithovore／Android／Natives 逐族半單位消耗、資源供應優先序、
  signed rate、最後一人保護、非農夫優先與 reservoir sampling 均接玩家／AI及可存檔亂數流；
  所有負池會先完成刪除，再以 owner 固定第一、其餘正常玩家槽 Fisher–Yates 洗牌的順序處理
  正成長，Android／Natives 不做自然正成長。AI 的 population player slot 現亦跨 JSON 保存；
  舊檔優先從殖民地 owner slot 復原，避免讀檔後 active slot 數與人口亂數序列分歧。
  typed groups 不保存原版 packed colonist 排列，故只宣稱候選集合與機率分布對齊；仍待未覆蓋的
  低頻人口 mutation path，不把完整混居 parity 宣告完成。證據見
  [`docs/re/owner-race-gravity-production-audit-20260825.md`](docs/re/owner-race-gravity-production-audit-20260825.md)，規格見
  [`docs/spec/owner-race-gravity-production.md`](docs/spec/owner-race-gravity-production.md)、
  [`docs/re/mixed-race-colonist-production-audit-20260825.md`](docs/re/mixed-race-colonist-production-audit-20260825.md) 與
  [`docs/spec/mixed-race-colonist-production.md`](docs/spec/mixed-race-colonist-production.md)、
  [`docs/spec/population-consumption-and-negative-growth.md`](docs/spec/population-consumption-and-negative-growth.md)。仍待其餘人口
  忠誠／士氣／領袖產出分項。事件產出已先閉合一條高影響 consumer：`sub_E2710` 會保留
  受事件殖民地自身 RP，卻不納入帝國研究聚合；超新星星系 RP 現只投入搶救，不再同時增加
  一般研究。舊 `sub_23DFE = Pick_Random_Colony_No_Capitol` 名稱已由原始指令推翻，訂正為
  event-colony filter；事件 raw type 2／6 完整 enum 仍留在事件系統整體 RE。證據見
  [`docs/re/research-breakthrough-audit-20260825.md`](docs/re/research-breakthrough-audit-20260825.md)、
  [`docs/re/colony-research-production-audit-20260825.md`](docs/re/colony-research-production-audit-20260825.md)、
  [`docs/re/research-application-selection-audit-20260825.md`](docs/re/research-application-selection-audit-20260825.md)、
  [`docs/re/research-application-callback-audit-20260825.md`](docs/re/research-application-callback-audit-20260825.md)、
  [`docs/re/event-colony-research-diversion-audit-20260825.md`](docs/re/event-colony-research-diversion-audit-20260825.md)，規格見
  [`docs/spec/research-breakthrough.md`](docs/spec/research-breakthrough.md) 與
  [`docs/spec/colony-research-production.md`](docs/spec/colony-research-production.md) 與
  [`docs/spec/research-application-selection.md`](docs/spec/research-application-selection.md)、
  [`docs/spec/research-application-callbacks.md`](docs/spec/research-application-callbacks.md) 與
  [`docs/spec/event-colony-research-diversion.md`](docs/spec/event-colony-research-diversion.md)。
  2026-08-25 又由 `sub_DE280` 閉合政體／士氣／領袖的共同逐人口加總：統一兩級只對
  食物與工業 `+50%／+100%` 並忽略一般士氣；封建／邦聯研究 `-50%／-25%`，民主／聯邦
  `+50%／+75%`。Go 已移除永久烘入每人口率的累乘作法，改由玩家與 AI 每回合依生效政體
  重建逐職務百分點；證據與規格見
  [`docs/re/colony-government-output-audit-20260825.md`](docs/re/colony-government-output-audit-20260825.md) 與
  [`docs/spec/colony-government-output.md`](docs/spec/colony-government-output.md)。AI 額外難度項仍留在 AI 經濟切片。
- [~] **外交、間諜、領袖與 AI 忠實化**：重建 `Diplomacy_Test_ @ 0x53146`、
  `Change_Relations_ @ 0x4E3B5`、`NPC_To_NPC_Treaty_Negotiations_ @ 0x2552D` 與原版間諜／領袖
  回合鏈；另依 `Resolve_Spies_ @ 0x10192B`、`Compute_Spy_Bonuses_ @ 0x100A83`、
  `Check_For_Officer_Level_ @ 0x92FDA`、`Decrement_Officer_ETA_ @ 0x934CF`、
  `Random_Officer_Check_ @ 0x97A66` 與 `Process_Trade_And_Research_Agreements_ @ 0x101E77`
  追回間諜成本／維護／Agent 傷亡、SABOTAGE 兩表上游、領袖經驗／ETA／招募／AI 任命、條約收益
  成長與終止時序。2026-08-25 已閉合普通貿易／研究協議逐回合鏈與玩家隨機領袖招募，並勘誤相鄰符號：
  `sub_101E77` 才是回合處理器，`sub_101EE3`／`sub_101F82` 分別建立貿易／研究協議。
  current 每回合增加 goal/5，goal%5 才擲 `Random(5)` 補一點；達 goal 封頂，goal 下降時立即
  下修。Go 已接 AI 高槽到低槽、再玩家高槽到低槽的方向順序、每方向 trade-before-research、
  雙方經濟 consumer 與可存檔獨立亂數流。證據與規格見
  [`docs/re/trade-research-agreement-turn-audit-20260825.md`](docs/re/trade-research-agreement-turn-audit-20260825.md) 與
  [`docs/spec/trade-research-agreement-turns.md`](docs/spec/trade-research-agreement-turns.md)。2026-08-25 另由
  `sub_100A3E/sub_100A83` 閉合 SABOTAGE 兩張帝國攻防表：五科技、諜報特性、心靈感應、
  最佳 Spy Master／Telepath 與八格政府表均已接回既有 AB／DB 分層，派駐人數仍由
  `sub_101483` 在消費端加入；證據與規格見
  [`docs/re/sabotage-score-upstream-audit-20260825.md`](docs/re/sabotage-score-upstream-audit-20260825.md) 與
  [`docs/spec/sabotage-score-upstream.md`](docs/spec/sabotage-score-upstream.md)。仍待
  完整一般外交關係演化與 AI 戰爭／協議決策。2026-08-27 已以 IDA Pro 9.4
  匯出 `Change_Relations_` 完整 406 條指令與 30 個直接 caller，並勘誤事件 4／5 在對稱戰爭
  下應早退。其餘 caller 仍需依 reason、方向條約、人類／AI 守門與 `+0x64F..+0x6BF`
  快取／抱怨欄位分類，不得用單一線性 `adjustRelation` 冒充已閉合。2026-08-27
  另由 `sub_4D78E` 閉合 `byte_180ED4` 14×14 種族關係目標表，並接回
  `Diplomacy_Growth_` 的 `Random(105)`／`Random(4)`／`Random(2)` 漂移與
  戰爭 -90 壓制；舊玩家↔AI 性格／軍力差及 AI↔AI 艦隊差關係漂移與其測試斷言已移除。
  AI↔AI 現依 `0x4E27C..0x4E2E3` 的鏡射順序保存高槽→低槽 raw current，並接回
  既有議會／政策 consumer、存檔與熱座矩陣壓縮。`NPC_To_NPC_Treaty_Negotiations_`
  也已閉合 ordered pair、難度頻率、八格政府表、raw 聲望／條約與協議記憶、第三方戰爭、
  互不侵犯／同盟／貿易／研究及納貢 mode 2；原有 `-25／+12／+25／+8` 自編政策門檻已移除。
  新 raw 矩陣均可存檔並隨熱座壓縮。2026-08-27 又閉合一般 AI↔AI 宣戰／停戰垂直切片：
  `sub_25DF1` reason 23 候選、`sub_51078` policy 4 與 -75..-99 關係 writer、`sub_5090C`
  的 `+0x717` 戰爭計時／`+0x72F` 冷卻，以及 `sub_2670A／sub_524FB` 的難度停戰門檻、
  relation +50 封頂 0 與 30 回合解除均已接入；矩陣可存檔並隨熱座壓縮。2026-08-28 已進一步
  證實 `+0x5EC` 是 `sub_D3D34` 每回合由逐艦八武器槽、改造、命中、船員／損傷及觀察者防禦
  重建的八欄方向矩陣，不是單一艦體總和；`.GAM` 垂直輸入已補齊：
  computer／size／armor／shield／base combat speed、五個 damaged-special bytes、分離損傷與 crew
  level 均無損進入 typed `Ship` 並可 JSON 往返；同輪另由 IDA 匯出 `sub_5EE27` 十一個 signed
  modifier words，逐艦純規則已實作八槽、戰機轉換、觀察者扣減、命中／彈藥、電腦與耐久修正。
  新造 AI 艦現也保存可證實的 raw weapon ID／mods、computer、size、armor、shield 與 base combat
  speed；未閉合改造失敗即關閉。IDA 另把 observer 的電腦武器扣減與引擎／種族防禦拆成兩張
  精確表。IDA 續追回 `sub_582BF／sub_58425` 的 hull／armor 表、強化船體／重型裝甲分軌與
  `sub_54E5B` 的電腦、掃描器、軍官、艦員、種族艦攻；typed producer 現逐艦產生 owner×observer
  方向矩陣，納貢與一般宣戰已共同消費。非零 `FleetStrength` 卻缺實艦 raw 資料的舊存檔才走
  明標 `exact=false` 的相容回退。
  證據與規格見
  [`docs/re/npc-power-matrix-audit-20260828.md`](docs/re/npc-power-matrix-audit-20260828.md) 與
  [`docs/spec/npc-power-matrix.md`](docs/spec/npc-power-matrix.md)。
  同輪已訂正 `+0x71F` 並非單純 treaty-break：`Change_Relations_` 的 `+0x64F／+0x65F` 保存 pending
  reason／幅度，`sub_252D5` 依政府表、協議與亂數形成 `+0x71F` 重複事件記憶。一般公式、
  ordered pair、雙向鏡射、談判第三方 +5、存檔與熱座壓縮均已接；納貢 reason 14 是首條可達
  writer。`+0x727` 另由全域 42 個站點確認為單向永久違約旗標；一般 AI 宣戰現先依 `sub_5138E`
  寫 target→actor，再讓 government 4 的事件門檻及 `word_18105C` 締約／納貢 cooldown 改走索引 6。
  其餘玩家回應 writer 與 reason 的 AI↔AI 可達 caller仍待閉合。自訂種族採中立 fallback；
  反向玩家條約、`+0x737` 鎖定 writer仍待閉合。2026-08-28 已依 `sub_25DF1` 的分支掃描順序
  接回特殊宣戰 reason 20、reason 68 與 reason 113 的食物赤字路徑；`sub_4DAB2` 證實
  `player+0xB0 < 0` 會累加 `+0x7EC`、否則歸零，該 signed-word streak 已由 AI 經濟輸出產生、
  保存並抵達正式宣戰 consumer。`+0x60E` 的另一條 reason 113 已完成
  `.GAM` raw 保留、JSON 往返與消費端，但 runtime producer 仍未知；其餘尚未閉合的是
  reason 22 的 `+0x6FF` 殖民破壞怨值 AI↔AI producer，以及
  `+0x6AF／+0x6BF` 記憶。特殊貿易與 ETA callback
  仍是可玩 remake 模型。證據與規格見
  [`docs/re/npc-war-ceasefire-audit-20260827.md`](docs/re/npc-war-ceasefire-audit-20260827.md) 與
  [`docs/spec/npc-war-ceasefire.md`](docs/spec/npc-war-ceasefire.md)、
  [`docs/re/npc-special-war-policy-audit-20260828.md`](docs/re/npc-special-war-policy-audit-20260828.md) 與
  [`docs/spec/npc-special-war-policy.md`](docs/spec/npc-special-war-policy.md)。
  `sub_233FA` 外層 gate 同日由既有事件 9 稽核交叉閉合：超空間亂流 active 時非跨維度 AI
  不執行宣戰候選，`player+0x8BC` 對應的跨維度種族可照常執行；正常與免疫分支均有測試。
  同日亦訂正 `+0x7EC` 的 `inc word` 為 16-bit 回繞，而非過去文件誤述的飽和。
  領袖招募現依 `sub_97A66/sub_9781D/sub_97B2D` 每回合擲骰，包含前五回合門檻、
  Charismatic／Repulsive、Famous 一般／進階加成、兩類四席與隨星曆開放的隨機候選前綴；
  亂數流可隨存檔續接。證據與規格見
  [`docs/re/random-officer-recruitment-audit-20260825.md`](docs/re/random-officer-recruitment-audit-20260825.md) 與
  [`docs/spec/random-officer-recruitment.md`](docs/spec/random-officer-recruitment.md)。
  AI 領袖任命也已依 `Do_AI_Leaders_ @ 0xD7439` 接上上一回合 offer、技能偏好 gate、
  嚴格 `treasury > hireCost+50`、30 回合拒絕 cooldown、動態艦艇／殖民地任命與 JSON 往返；
  舊 `0xD7662` 入口及依種族固定贈送 Commando 的代理已移除。raw 艦艇五欄與殖民地評分
  沒有一對一表示處仍明標近似。證據與規格見
  [`docs/re/ai-leader-assignment-audit-20260825.md`](docs/re/ai-leader-assignment-audit-20260825.md) 與
  [`docs/spec/ai-leader-assignment.md`](docs/spec/ai-leader-assignment.md)。
- [ ] **原版 AI 決策器忠實化**：`internal/ai.NewDecider(ModeOriginal, ...)` 目前仍回傳
  `RemakeDecider` 並以 `ok=false` 告知 fallback；依 `Compute_AI_Data_ @ 0xD3D34`、
  `Move_All_AI_ @ 0xDBB29`、`All_AI_Colonize_ @ 0xE67F6`、`All_AI_Tech_Select_ @ 0xDCA69`、
  `Search_For_Battles_ @ 0xE9D62` 重建原版 state machine、難度作弊、殖民地職務與建造、研究／
  應用選擇、艦隊目標與戰鬥決策；常態研究切片已接原版 application 級估值抽選，但不能用
  單一切片替代完整原版 AI parity。2026-08-26 已把官方五級 Generic AI
  bonuses 的 Growth／Food／Prod／Res／BC 與 Command Deficit `12/11/10/9/8` 接進 AI 每回合
  暫態殖民地與帝國結算；玩家不吃加成，
  負 quarter 採向下取整，且修正 `NewDemoSession` 性格難度為 1、session 卻留零值 Tutor 的雙真相。
  Spy Bonus `-2/-1/0/+1/+2` 亦已進 AI 攻守兩張能力表；數值已證實，攻守共同注入及逐殖民地
  quarter 捨入／士氣重力順序仍為強推論。Troops／Marines `-2/-1/0/+1/+2` 經盤點確認早已由
  `GroundDifficultyBonus` 接入 AI 殖民地防守與叛軍；Antaran Marines `2×difficulty-4` 的
  owner≥8 共用地面／登艦 block 已由 IDA caller 閉合，但目前沒有 owner≥8 typed 地面單位，
  不另造原版未證實的登陸事件。2026-08-26 另閉合 `All_AI_Tech_Select_ → sub_DC288 → sub_FD335`：
  新局 AI 保存 raw6／raw4／raw7 profile，常態回合以一次 application 級估值抽選同時決定
  field/application；舊存檔 profile 未知才回退 remake 啟發式。2026-08-26 另由
  `All_Colony_AI_ → Colony_AI_ → sub_D6E1D → sub_D10EE／Assign_Colony_Building_` 證實
  原版生產是逐殖民地產品，不是全帝國工業單一造艦池；Go 已保存每星 AI 殖民地產品與進度，
  先消費 typed 可建建築，沒有候選的產能才進造艦轉接層，且同一份產能不再重複消費。
  raw 1..48 可建 gate、難度濾門與加權抽選控制流已接；同日直接解碼 `0xCFF62` 的 47-case
  跳表，raw 4／7／12／34／36 已改用人口與 Honorable 性格的完整原版精確分數；另由
  `sub_DC288 @ 0xDC2E8` 的唯一直接寫入與四個 consumer，先閉合 raw 6／19／30／35 的
  late-tech 零分；其後再由 `player+0x117+TechnologyID`、`colony+0x136+BuildingID`、礦產與
  政府碼交叉閉合共用 priority gate，四棟研究設施現已使用完整正值／零值公式；raw 15
  Biospheres 亦已閉合為 priority gate 時 0、否則 `18+[Pacifist]`。raw 16 Food Replicators
  已由主要人口 player-slot 選擇、Lithovore 差異旗標與帝國食物差額兩個寫入端閉合，現依
  「主要人口與 owner 不得同為 Lithovore」、食物赤字與 Pacifist 使用完整公式；profile
  不完整時不冒稱 exact。2026-08-26 回查 cache+2 唯一寫入端已修正舊分支方向誤讀。
  raw 10 Cloning Center 亦已閉合 population-growth trait、結算前 1500 BC gate、signed word
  淨 BC／64 與 `sub_134C92` 整數平方根；計分不消耗候選抽選 RNG。raw 20 Holo Simulator
  與 raw 31 Pleasure Dome 也已依八政府、人口 2／3 邊界及 budget-factor gate 接入固定 10／16。
  raw 17 Gaia Transformation 已閉合為 `budgetFactor+[Pacifist]`，並補上 Terran 候選、
  一次性產品完成及 AI 殖民地／全局行星 Gaia 同步，不會誤記為常駐建築。
  raw 44 Terraforming 亦已依 Barren..Arid 內層跳表、Aquatic、Pacifist、priority gate 與
  budget factor 接入精確分數；只有具下一級氣候的殖民地可選，完工同步殖民地／行星且不留常駐旗標。
  raw 37 Soil Enrichment 已依修正後的 cache+2、每農夫食物半單位、帝國食物差額與 Pacifist
  閉合；適用氣候可正常選建，完工只增加每農夫食物且不留常駐旗標。
  raw 21 Hydroponic Farm、raw 43 Subterranean Farms 與 raw 46 Weather Controller 亦已由
  外層跳表、`colony+0xDD` 半單位快取及 `player+0xB0` signed word 消費端閉合：前兩者保留
  四段食物表與完整赤字幅度，後者保留赤字二段式；三者的 Pacifist、cache+2 與 priority gate
  差異均已接線。AI 可由正常唯一候選完成三棟建築，並分別寫回固定食物或每農夫食物效果。
  raw 5 Atmospheric Renewer、raw 13 Core Waste Dumps、raw 32 Pollution Processor 也已追回
  `Compute_AI_Data_` 7-byte cache 的唯一寫入鏈與 `sub_DEE1B` 清污成本回寫；現以主要人口
  Tolerant、精確 `PollutionCleanupCost` 的 5／10 邊界、整數平方根及 Pacifist 計分，且只有
  raw 13 受 priority gate 阻擋。三棟均可由正常候選完工並寫回污染旗標。
  raw 25 Planetary Gravity Generator 已依 owner 的 High-G 優先、Low-G／一般分支與行星
  `LOW_G=0／NORMAL_G=1／HEAVY_G=2` 閉合完整 `0／3／6+[Pacifist]` 表；不誤吃 priority
  gate。正常 AI 候選完工後的 `NormalizeGravity` 已由逐人口工業產出 consumer 驗證，不只測旗標。
  raw 29 Planetary Stock Exchange、raw 39 Spaceport 與 raw 33 Recyclotron 亦已沿
  `Compute_AI_Data_ cache+1 → sub_E0C1D → sub_E0A93 → sub_E0A18` 閉合主要人口容量；
  現依人口門檻、priority gate、Honorable／Pacifist、主要人口 Tolerant 與完整容量公式計分。
  三棟均可由正常唯一候選完工，並由該殖民地 BC 收入或不產生污染的人口產能 consumer 驗證。
  raw 38 Space Academy 亦已由 `sub_DEE1B` 的淨工業寫入與 `sub_E2710` 帝國加總端閉合
  `colony+0xE9`；現保留 17 工業／5 人口／budget gate、`netIndustry-15` unsigned 平方根及
  1000 上限。正常 AI 候選完工後，同星系 AI 實艦每回合會取得學院經驗；AI 匯總造艦無來源
  殖民地，故 AI 新艦起始等級仍是明示資料模型限制。
  raw 2 Armor Barracks 與 raw 22 Marine Barracks 已沿 `sub_D3A68`／`sub_D3BA0` 的
  近圈條約／無政策／戰爭／延伸航程四槽、`Compute_AI_Data_ cache+5` 的他國艦隊 ETA=9，
  以及 `sub_CFF02` 的戰爭帝國外族人口旗標閉合；`sub_10034D` 的燃料表亦證實
  Standard／Deuterium／Iridium／Urridium／Thorium 航程為 4／6／9／12／255。兩式已接
  Ruthless、人口／budget、政府與交叉兵營 gate；正常唯一候選完工後由 AI 陸戰隊／戰車營
  補充 consumer 驗證。舊存檔若缺 fuel application 或外交／population-slot 對映，仍誠實走 fallback。
  raw 23 Planetary Barrier Shield、raw 24 Planetary Flux Shield 與 raw 28 Planetary Radiation
  Shield 亦已沿同一星系壓力 context 閉合 priority／ETA9 gate、四槽係數、Ruthless、budget，
  以及 Radiation Shield 的 Radiated／Pacifist 加分。三棟可由正常唯一候選完工，會依序取代
  低階護盾、同步 AI colony／planet 的 Radiated→Barren，並由既有軌道轟炸 consumer 讀取
  每發 5／10／20 減傷；context 不完整時維持明示 fallback。
  raw 40 Star Base、raw 8 Battlestation 與 raw 41 Star Fortress 亦已由同一星系壓力 context、
  `sub_E2000` 的 `player+0x3A／+0x3C` 寫入鏈及 UI／AI consumer 閉合；typed 分數現保留
  兩組外交係數、`max(0,used+1-supply)` 指揮赤字、Ruthless 與 budget factor。正常候選完工
  會維持 Star Base→Battlestation→Star Fortress 單槽取代鏈，且已由指揮評等與掃描範圍
  consumer 抽測；衛星精確武裝／hull space 仍維持原有近似，不因本項升格。
  raw 26 Missile Base、raw 27 Ground Batteries、raw 42 Stellar Converter 與 raw 47 Fighter
  Garrison 亦已由共享 `0xD050B..0xD0549` 尾端閉合；四棟現依 ETA9、四槽星系壓力、
  Ruthless、priority gate 與 budget factor 精確計分。正常唯一候選均可完工並新增對應固定
  防禦反擊者；反擊者本身的 space／武器／戰機火力仍保留手冊＋近似證據等級。
  raw 45 Warp Field Interdictor 已由 `0xD05BD..0xD0614` 與星系 bit-mask writer
  `sub_E5296 @ 0xE5296..0xE53CD` 閉合：同星系已有己方干擾器時只保留 budget factor 一半，
  否則依 ETA9 與 treaty／no-policy／war／extended 四槽使用 `5／2／3／4／1` 係數，非零再加
  Ruthless。正常候選、完工與三秒差距航線降速 consumer 已抽測；`route.go` 不再內嵌中文建築名，
  改以 raw 45 對照資料表。
  raw 3 Artemis System Net 已由 `0xD054E..0xD05B8` 閉合；它依原版共用 raw 45 星系 bit，
  尚無 raw 45 時採 ETA9／treaty／no-policy／war／extended 的 `10／1／1／3／1` 係數，非零再加
  Ruthless 與 budget factor 一半。同星系已有 raw 45 時只留 budget factor 一半。正常候選、
  完工、玩家艦隊與 AI 艦隊抵達 consumer 均已抽測；兩條 consumer 都改以 raw 3 查資料表，
  不再保存內嵌中文建築名。至此一般建築表可映射的 raw 1..47 均有 exact 計分或已證實零分。
  raw 1 Alien Management Center、raw 11 Colony Base、raw 14 Dimensional Portal、raw 18
  與 raw 48 已證實直接走共同零分尾端；一般建築表中的 raw 1／14 已接 exact 零分，正常
  候選只剩兩者任一時不會被類別代理誤選。raw 9 Capitol 亦已閉合：remake 會保存每個帝國
  的指定 Capitol 行星、欄位存在狀態與失守後待重建狀態；非統一政體只在指定行星加入成本
  200 的 Capitol 候選，AI 精確分數為 100。攻陷時移除且不過戶 Capitol，舊擁有者改選其餘
  人口最高殖民地（同人口取較低 colony index），新擁有者沒有指定行星時接手該行星；完工後
  清除士氣懲罰。狀態已接玩家、AI、熱座與 JSON 往返，玩家可見名稱由
  `assets/i18n/tech.json` 提供，不內嵌於 Go。
  帝國配額、支援／戰鬥艦產品仍待閉合，故建造整體仍是部分完成。艦隊／外交及其餘 AI
  state machine 亦待閉合。2026-08-28 又以 IDA 證實 AI 職務分配是封鎖／未封鎖
  分流、逐 colonist 排序、邊際輸出與全帝國迭代；這直接反證現有逐殖民地
  `Decider.ColonyJobs` 足以承載 original mode。四個 comparator、job raw ID、
  `player+0xAA/+0xAC` producer、逐 race／prisoner 候選與帝國停止比較現已閉合；
  未封鎖殖民地會先依原版最低工人／半工業消耗配置，再以研究－工業邊際逐人平衡，
  並同步 `PopulationGroups`。`Compute_Blockades_` producer、`sub_D61E7` 封鎖分支與
  `sub_D6AD4 → sub_D6A00` 追加農夫 consumer 均已接入；因尚無 typed 運輸容量，追加農夫
  對 `player+0x38` 的殖民地選擇仍採失敗即關閉近似，不能據此勾掉完整 AI 決策器；見
  [`docs/re/ai-colony-jobs-audit-20260828.md`](docs/re/ai-colony-jobs-audit-20260828.md) 與
  [`docs/spec/ai-colony-jobs.md`](docs/spec/ai-colony-jobs.md)。其餘 AI 建造證據見
  [`docs/re/ai-difficulty-economy-audit-20260826.md`](docs/re/ai-difficulty-economy-audit-20260826.md) 與
  [`docs/spec/ai-difficulty-economy.md`](docs/spec/ai-difficulty-economy.md)、
  [`docs/re/ai-command-deficit-audit-20260826.md`](docs/re/ai-command-deficit-audit-20260826.md) 與
  [`docs/spec/ai-command-deficit.md`](docs/spec/ai-command-deficit.md)、
  [`docs/re/ai-spy-difficulty-audit-20260826.md`](docs/re/ai-spy-difficulty-audit-20260826.md) 與
  [`docs/spec/ai-spy-difficulty.md`](docs/spec/ai-spy-difficulty.md)、
  [`docs/re/antaran-marines-caller-audit-20260826.md`](docs/re/antaran-marines-caller-audit-20260826.md) 與
  [`docs/re/ai-normal-research-selection-audit-20260826.md`](docs/re/ai-normal-research-selection-audit-20260826.md)、
  [`docs/spec/ai-normal-research-selection.md`](docs/spec/ai-normal-research-selection.md)、
  [`docs/re/ai-colony-build-selection-audit-20260826.md`](docs/re/ai-colony-build-selection-audit-20260826.md) 與
  [`docs/spec/ai-colony-build-selection.md`](docs/spec/ai-colony-build-selection.md)。
  2026-08-28 另閉合 AI 稅率 consumer：`player+0x31` 的直接寫入只有真人稅率視窗
  `sub_CC198`，AI 回合沒有依國庫主動調稅 producer。新建 AI 現從 0% 起步並保持現值；匯入
  存檔的非零稅率也不被覆蓋。舊 10／30／50% `RemakeDecider` 門檻已退出正常原版路徑；見
  [`docs/re/ai-tax-rate-audit-20260828.md`](docs/re/ai-tax-rate-audit-20260828.md) 與
  [`docs/spec/ai-tax-rate.md`](docs/spec/ai-tax-rate.md)。
  同日亦訂正 AI 對真人戰爭 policy 的下游：`sub_51078` 已證實 human 戰爭依難度寫 raw 5／6，
  `aiForeignPolicyFor` 現直接保留正式 4／5／6 給目標估值，不再由中文態勢與關係猜回錯誤 policy。
  後續 IDA 勘誤證實 `sub_25DF1` 所有候選 target loop 均排除 human，不能把 AI↔AI 的方向
  bias／duration／cooldown 錯搬給真人。真正可見入口之一是 `sub_DB257 @ 0xDB257` 的 AI 艦隊
  接戰鏈：抵達玩家殖民星時現會先呼叫 typed human-war writer，依難度寫 raw 5／6、清協議、
  寫 -75..-99 關係並點亮宣戰會談；raw 6 也已接入每回合關係成長／戰爭 -90 drift 與存檔。
  後續全庫 raw displacement 稽核已找到正常 producer `sub_53EDB → sub_544A1`：現已接入
  `player+0x816` 的 20–39 回合 decision cooldown 與存檔，並移除固定 12 回合寬限、1.25 倍
  軍力門檻及誤用 losing-ground personality chance。後續又閉合 `player+0x88F` 接觸後逐回合
  遞增／250 上限／至少 10 回合 gate、`word_181080` 七欄 personality score、正負 score
  threshold、RNG 消耗順序與四類結果純規則。`sub_544A1` 的完整 directional incident memory、
  排名／科技趨勢，以及 `sub_4F93B` 的科技／殖民地候選 producer 仍待 typed 化；四種 action
  kind、候選 gate、RNG 順序與 BC／科技／直接要求／殖民地 payload 核心已成為純規則。
  Honorable AI 的玩家正式違約鏈亦已閉合：`break_formal` 永久寫 AI→玩家 `+0x727` 並存檔，
  base score 由 Honorable +20 改讀 Dishonored -10；普通貿易解約不誤寫。
  `sub_4F0DC` 的事件記憶下游亦已追回：它只把 reason 1..9 從 `+0x64F` 複製至
  `+0x6CF`；`sub_544A1` 再以 signed `+0x71F` 與共用表
  `word_180CF0=[1,2,3,3,4,5,2]` 計算負分及 `reason+70`。純規則與原始表測試已接；
  `sub_4F0DC` 的完整上游門檻及正常玩家事件 reason producer 仍待閉合，未以猜測接線。
  同一 score 的存活帝國人口優勢 `-10`，以及第 100 回合後雙方 40 回合人口成長差也已由
  `+0xA6／+0xB9B` raw 讀取閉合為純規則；其餘國力／科技項仍待拆解。
  `sub_500CF` 亦確認等同既有 `OriginalNPCPowerRatio`，ratio>=300 且政府 raw!=5 時的
  `-ratio/40` 與行動上限 150 已接純規則；尚缺玩家↔AI 逐艦方向 `+0x5EC` shell producer。
  目前願戰來源仍是明示的
  `DecideStance` fallback，不冒稱原版完成。見
  [`docs/re/ai-human-diplomacy-dispatch-audit-20260828.md`](docs/re/ai-human-diplomacy-dispatch-audit-20260828.md) 與
  [`docs/spec/ai-human-formal-war-policy.md`](docs/spec/ai-human-formal-war-policy.md)、
  [`docs/re/ai-human-fleet-target-producer-audit-20260828.md`](docs/re/ai-human-fleet-target-producer-audit-20260828.md) 與
  [`docs/spec/ai-human-fleet-target-producer.md`](docs/spec/ai-human-fleet-target-producer.md)。
- [ ] **艦隊、殖民、事件與安塔蘭忠實化**：重建 `Move_All_Ships_Toward_Stars_ @ 0xFFEEA`、
  `sub_E5EB3 @ 0xE5EB3` 殖民建立鏈、`Compute_Blockades_ @ 0xE5097`、`Compute_Contacts_ @ 0xEB192`、
  `Check_All_Rebellions_ @ 0xED44A`、`Determine_Event_ @ 0x2230A` 與
  `Antaran_Invasion_Check_ @ 0x63D92`；需包含逐回合座標、截擊、合併／拆分、星雲／黑洞／
  干擾器、殖民前置與起始狀態、封鎖／接觸、叛亂兵力／政府／士氣／易手、29 種隨機事件與
  7 種狀態播報的目標／持續 record／效果，以及安塔蘭難度／時間／艦隊／目標／實際戰鬥。現在 ETA 航行與
  部分事件複合效果仍有 remake 模型。一般事件排程已依
  `sub_2230A` 接入前五次保護、五級 delta 公式、`Random(512)`、五候選與最早日期；
  `sub_22D57` 的總人口來源、好／壞事件極值排除與差平方權重亦已由 IDA 閉合並寫成純規則測試。
  全局排程、Lucky 全槽掃描、人口權重目標、熱座非目前席位回寫，以及 AI 的 BC／RP
  一對一事件已接線；AI 殖民地／艦隊／外交與持續 record 等複合效果尚未閉合。
  2026-08-28 已由 `sub_E5097` 閉合封鎖資料形狀與主鏈時序：`star+0x2A` 是被封鎖
  player mask，八個 `+0x2B` byte 是逐受害者 blockader mask；每回合先清表，再由有效、
  已抵達艦隊與 owner policy raw 4..6 重建，owner>=8 封鎖同星所有殖民者。`.GAM`
  匯入現無損保存兩張 mask，不再壓成 bool。正常回合亦已由玩家／非目前熱座／AI 艦隊、
  多 owner 同星 occupied mask 與 owner>=8 怪物／安塔蘭分支重建；AI 在下一回合依
  `sub_D61E7` 的 food-industry 排序補農夫、其餘轉工人，無農業事件分支轉工人／科學家。
  封鎖 caller 亦已勘誤為 `-Random_(5)`／reason raw 7；AI 被真人封鎖時依
  `sub_4E3B5` 的關係、政體、Charismatic 與算術 `/4` 寫 `+0x6BF` typed 積怨，
  policy>=4 於 `0x4E75C` 早退且不改一般關係分數。尚缺目前資料模型沒有的真人對真人
  raw policy，以及沒有玩家可見 consumer 的真人側反向 `+0x6BF`；見
  [`docs/re/blockades-audit-20260828.md`](docs/re/blockades-audit-20260828.md) 與
  [`docs/spec/blockades.md`](docs/spec/blockades.md)。
  事件 19–23 怪獸入侵已由 `sub_2230A`、`sub_23BEC`、`sub_206A2`、`sub_A16BF` 與
  `sub_A1A23` 閉合最早回合／亂流排除、受害帝國有效殖民星 reservoir sampling、五個獨立
  loader 及 owner 8 航行入口。Go 已移除「第一顆無主星」與太空鰻借用變形蟲的舊代理，
  玩家、非目前熱座席位與 AI 均可在自身殖民星建立正確種類並隨 JSON 保存。owner 8 首次航行
  已由 `sub_A1762`、`sub_A1A23`、`sub_FF799`、`sub_EBE79`、`sub_EBB0C` 與 `sub_FFDDA`
  閉合：戰略速度固定 1 parsec/turn，ETA 為出生點到目的星的無條件進位星距；Go 現以可存檔
  `TransitETA` 推進，抵達前不盤據、不阻擋拓殖且不推進太空鰻 age。因 remake 不保存原版視窗
  捲動 globals，出生邊界仍為銀河矩形近似；途中逐座標截擊仍待閉合。同星不同種類已由
  `Search_For_Battles_` 的 owner/type side bit、`sub_E8194` side reservoir sampling 與
  `sub_E84A5` 逐項消費閉合資料形狀；`sub_E8029` 已訂正為 side 選定後的對手選擇：
  remake 現會列舉及顯示全部停泊群組，戰勝只移除本次第一個停泊群組，不再誤刪同目的星
  航行 record。原版 RNG 屬全銀河自動戰鬥排程順序；玩家單次攻擊採穩定切片順序，以維持
  鎖步重播及戰後 record 回寫一致，明標為介面轉接而非原版排程順序。
  2026-08-27 已訂正殖民入口：`sub_BB082 @ 0xBB082` 是畫面 helper，真正建立函式為
  `sub_E5EB3`。一般殖民人口 1、不可自然耕作／Lithovore／Cybernetic 從工人開始、其餘從農夫
  開始，Native 三人固定農夫，以及正常 caller 消耗殖民船均已接；前哨重用 raw cache 與完整
  callback 仍待。見 `docs/re/colonization-starting-job-audit-20260827.md`。
  五種事件怪物藍圖已閉合 raw type、艦體、引擎、電腦、special bytes、武器槽、戰速、結構與
  怪物專用裝甲；Go 的新生成怪物及太空鰻分裂已接雙血池。抵達後的殖民地戰鬥已由
  `Search_For_Battles_`、`sub_E8029`、`Do_1_Combat_`、
  `sub_E87D2`、`sub_4267B`、`Strategic_Bombardment_` 與 `sub_DD2F2` 閉合控制流：太空鰻
  不攻擊，其餘怪物先進固定防禦戰，存活者固定三輪轟炸、總傷害 `/40` 後走共用殖民地傷亡池；
  玩家、非目前熱座與 AI 均可回寫建築、駐軍、建造進度、人口及殖民地移除。武器 mods 的
  `raw 0x0002/0x0004` 已由 `sub_2B9E3`／`sub_39434` 閉合為 HV／PD，快速怪物反擊現逐槽、
  逐門消費原版武器表與 mount 數量，不再每回合只打一發推測傷害。ID 44 Plasma Flux 的
  目標尺寸倍率，以及 ID 45 Caustic Slime 的可堆疊 `+0x43`、四面逐回合傷害與每回合減 5
  已由 `sub_289C4`／`sub_ACF83`／`sub_4A5CE`／`sub_39985` 閉合並接入格子戰術；快速結算
  對黏液採四面同回合近似。Dragon `0x4000` 已由改造表、category 2 分支及執行期 +50
  傷害資料流閉合為 OVR；格子戰術先把基礎 300 加成至 450，再依 ID 40 每格扣 15，快速
  結算採無格距離的近距 450。Plasma Flux 舊稱六格已由 `sub_3A82F`／`sub_ADE18` 與
  CMBTSFX asset 6 勘誤為 96px（約 4.8 格）歐氏半徑；格子戰術現會聚合 mount 全部門數，
  對雙方鄰艦套距離平方衰減及尺寸分段擲值，快速結算則以中心距離作用全艦隊。戰機中隊亦已
  接原版整隊 50% 迴避與逐架 `25×衰減傷害/單架耐久` 傷亡；remake 沒有持久在途飛彈 record，
  故飛彈受 Flux 傷害仍是資料模型限制。反擊選目標與固定防禦部分數值仍明標近似。正常星系「攻擊怪物」現會建立
  原版藍圖的逐體 `CombatShip` 並進入格子戰術，戰後把存活怪物裝甲／結構加總回寫；結果另有
  `monster_combat_outcome` 鎖步指令，重播不再遺失怪物戰損。太空鰻分裂已由
  `sub_DB8D8 @ 0xDB8D8` 閉合：每艘停泊 type 13
  怪獸各自累加 `ship+0x61`，恰好 30 時歸零，同星建立 age 0 新生鰻，全銀河 active 上限 4。
  Go 以同星 `Count`、逐體 `EelAges` 與聚合結構接入正常 EndTurn、舊 JSON／多人快照及玩家回報；
  新生個體不會在建立回合再次增齡。證據與規格見
  [`docs/re/random-event-monsters-audit-20260825.md`](docs/re/random-event-monsters-audit-20260825.md) 與
  [`docs/spec/random-event-monsters.md`](docs/spec/random-event-monsters.md)，航行追加證據見
  [`docs/re/event-monster-route-audit-20260825.md`](docs/re/event-monster-route-audit-20260825.md) 與
  [`docs/spec/event-monster-route.md`](docs/spec/event-monster-route.md)；戰後鏈見
  [`docs/re/event-monster-colony-battle-audit-20260825.md`](docs/re/event-monster-colony-battle-audit-20260825.md) 與
  [`docs/spec/event-monster-colony-battle.md`](docs/spec/event-monster-colony-battle.md)；藍圖見
  [`docs/re/event-monster-blueprints-audit-20260825.md`](docs/re/event-monster-blueprints-audit-20260825.md) 與
  [`docs/spec/event-monster-blueprints.md`](docs/spec/event-monster-blueprints.md)；武器 runtime 見
  [`docs/re/event-monster-weapon-runtime-audit-20260825.md`](docs/re/event-monster-weapon-runtime-audit-20260825.md) 與
  [`docs/spec/event-monster-weapon-runtime.md`](docs/spec/event-monster-weapon-runtime.md)；特殊武器見
  [`docs/re/event-monster-special-runtime-audit-20260825.md`](docs/re/event-monster-special-runtime-audit-20260825.md) 與
  [`docs/spec/event-monster-special-runtime.md`](docs/spec/event-monster-special-runtime.md)，Flux 深入證據與戰機規格見
  [`docs/re/event-monster-plasma-flux-spread-audit-20260825.md`](docs/re/event-monster-plasma-flux-spread-audit-20260825.md)、
  [`docs/spec/event-monster-plasma-flux-spread.md`](docs/spec/event-monster-plasma-flux-spread.md) 與
  [`docs/spec/event-monster-plasma-flux-fighters.md`](docs/spec/event-monster-plasma-flux-fighters.md)；同星群組見
  [`docs/re/event-monster-same-star-groups-audit-20260825.md`](docs/re/event-monster-same-star-groups-audit-20260825.md) 與
  [`docs/spec/event-monster-same-star-groups.md`](docs/spec/event-monster-same-star-groups.md)；戰術入口見
  [`docs/re/event-monster-tactical-entry-audit-20260825.md`](docs/re/event-monster-tactical-entry-audit-20260825.md) 與
  [`docs/spec/event-monster-tactical-combat.md`](docs/spec/event-monster-tactical-combat.md)。
  事件 24 超新星亦已由 `sub_23A5F`、`sub_2230A`、`sub_23B64`、`sub_206A2`、`sub_E2A70`
  與 `sub_DCDAC` 閉合：目標改為全銀河星系 rejection sampling，倒數改為 1.31 的
  `Random(5)+10-difficulty`，初始需求嚴格為建立時全星系 RP×倒數；每回合全 owner 殖民地
  RP 只投入搶救，失敗時玩家、非目前熱座與 AI 同星殖民地及各自殖民行星均會移除／變
  Radiated。舊 `×(倒數+1)`、均勻 6–14、目前玩家限定與單一代表行星代理均已移除。
  超新星候選星已要求至少一座 active 且無 Capitol 的殖民地，成立後仍消費／影響同星全部
  active 殖民地；1.50 倒數版本差仍未知。證據與規格見
  [`docs/re/random-event-supernova-audit-20260825.md`](docs/re/random-event-supernova-audit-20260825.md) 與
  [`docs/spec/random-event-supernova.md`](docs/spec/random-event-supernova.md)。
  事件 25 時空異象已由 `Determine_Event_`、`sub_23BEC`、`sub_242FC`、`sub_206A2`、
  `sub_23DFE` 與 `sub_E2710` 閉合受害帝國殖民星 reservoir sampling、同星互斥、age 0..4
  不抽骰、age 5..20 的 `Random(20)==1` 與 age 21 強制解除。Go 已以 `Turns=-1` 對齊首個
  active age，且只在 age>4 消耗事件亂數；玩家、非目前熱座、AI 及同星多 owner 的食物、
  工業、研究與人口成長回合輸入均會凍結，原始殖民資料與事件 record 可存檔往返。
  食物／工業／建造／維護／人口移動的原版 raw 欄位尚未逐項靜態閉合，現依官方手冊與共用
  重算鏈實作；時空異象使用 `sub_23BEC` 而非 `sub_23DA0`，現有靜態證據不支持替它套用
  Capitol 排除條件；GNN state 同回合時序與 1.50 版本差仍未知。證據與規格見
  [`docs/re/random-event-stasis-audit-20260825.md`](docs/re/random-event-stasis-audit-20260825.md) 與
  [`docs/spec/random-event-stasis.md`](docs/spec/random-event-stasis.md)。
  事件 26 超空間獸已由 `Determine_Event_`、`sub_206A2`、`sub_100618`、`sub_23CED`、
  `sub_941C6` 與 `sub_8A4C4` 閉合：active 回合沒有舊 remake 自創的 20% 攻擊骰，而是從
  record 指定帝國所有航行艦 reservoir sampling 一艘，處理後再以 `sub_22D57` 重抽下回合
  帝國；age 5 起 `Random(20)==1`、age 21 強制解除，且解除判定先於刪艦。Go 已保存可存檔
  `TargetKind/TargetIndex`，接通玩家多艦隊、非目前熱座與 AI，移除固定目前艦隊／最弱艦代理。
  原版 0x81-byte ship 鏈首替換在 remake 的 `Fleet` 切片沒有同構中間狀態，現以玩家可見的
  均勻刪除結果重製並明標資料模型近似；1.50 差異仍未知。證據與規格見
  [`docs/re/random-event-warp-beast-audit-20260825.md`](docs/re/random-event-warp-beast-audit-20260825.md) 與
  [`docs/spec/random-event-warp-beast.md`](docs/spec/random-event-warp-beast.md)。
  事件 28 蟲洞已由 `Determine_Event_ @ 0x2230A`、`sub_100519 @ 0x100519` 與
  `sub_FFDDA @ 0xFFDDA` 閉合：一般好事件目標帝國的非跨維度航行船按船 reservoir sampling，
  抽船只用來定同位置／目的地群組，整支艦隊會在事件建立回合立即寫成抵達；record 保存實際
  移動艦數與目的星供新聞。Go 已移除「目前選取艦隊 ETA=1」代理，改為按艦數權重選玩家多
  艦隊並共用探索／水雷／發現抵達 consumer，非目前熱座與 AI 亦立即回寫位置；跨維度、空艦隊
  與非法目的地會誠實拒絕。原版逐船座標中間值在 remake `Fleet` 模型沒有同構欄位，1.50 差異
  仍未知。證據與規格見
  [`docs/re/random-event-wormhole-audit-20260825.md`](docs/re/random-event-wormhole-audit-20260825.md) 與
  [`docs/spec/random-event-wormhole.md`](docs/spec/random-event-wormhole.md)。
  狀態播報 29–35 亦已由七個 setter、新聞 dispatcher 與觸發 caller 閉合 record 契約：帝國滅亡
  依存活轉換、擴張採三階段殖民星門檻、排行榜限 elapsed > 50／議會前且每回合 1/40、Orion
  由未探索特殊星抵達觸發。Go 已接玩家／熱座／AI 目標、同回合新聞佇列、JSON／多人快照，
  並由實際安塔蘭勝利與殖民地叛亂結果建立 33／35；後兩者因 1.31 無直接 caller，明標為
  觸發近似。事件 34 的舊斷言亦已由更深 IDA 稽核訂正：`sub_E4D06` 只寫 pending surrender，
  `sub_E4DC9/sub_E4B5F` 才延後移交殖民地、前 83 項已知科技、國庫、貨運艦與解除任命的領袖，
  並刪除投降者艦艇、清空外交／協議及改派可表示間諜。Go 已依此接 AI→AI、AI→玩家／熱座、
  穩定 AI 索引、JSON／多人快照與事件 34；raw `+0x717` 及 `sub_27A3D` 完整接收者評分未定名，
  自動觸發保留為正式戰爭＋壓倒性國力的明示近似。證據與規格見
  [`docs/re/event-status-broadcasts-audit-20260825.md`](docs/re/event-status-broadcasts-audit-20260825.md) 與
  [`docs/spec/event-status-broadcasts.md`](docs/spec/event-status-broadcasts.md)、
  [`docs/re/empire-surrender-audit-20260825.md`](docs/re/empire-surrender-audit-20260825.md) 與
  [`docs/spec/empire-surrender.md`](docs/spec/empire-surrender.md)。
  安塔蘭週期入侵亦已由 `Antaran_Invasion_Check_ @ 0x63D92` 及其完整直接 helper 鏈閉合
  科技等級 `200/100/0` 延遲、每 25 回合加速資源、難度 `100/150/200%`、攻守資源池、
  `{2,5,12,30,75}` 成本、攻守艦上限、readiness 與 `Random(200)`、Lucky `/3` 帝國權重、
  殖民星均勻抽樣及最多五艘部署。Go 已改為全局一次、可存檔 pending ETA 與快速戰鬥，
  不再固定攻玩家母星或直接扣 BC／人口；owner 8 的逐座標航行、完整快速／戰術 battle record
  、殖民地固定防禦及戰後消費仍是明示近似。已抵達艦隊現在會按快速戰鬥實際倖存數回寫
  offensive／deployed pool，不再把抵達或出兵誤當成全數損失。證據與規格見
  [`docs/re/antaran-periodic-invasion-audit-20260825.md`](docs/re/antaran-periodic-invasion-audit-20260825.md) 與
  [`docs/spec/antaran-periodic-invasion.md`](docs/spec/antaran-periodic-invasion.md)。
- [~] **戰鬥、地面戰、轟炸與防禦忠實化**：分別追快速結算與格子戰術，完成原版艦艇／武器 runtime
  設計、戰機出擊與返航、`Fire_Fighter_Bomb_ @ 0x3AC20`、
  `Fire_Fighter_Beam_ @ 0x3AD57`、`Ground_Combat_Round_ @ 0xEC4FE`、
  `Strategic_Bombardment_ @ 0x4257E`、
  `Fighter_Garrison_Strength_ @ 0x5F64C`、`Get_Colony_Hits_ @ 0x42371` 的欄位與回寫鏈。2026-08-24
  已以 IDAPython 閉合 `Strategic_Bombardment_` 控制流：Go 外圈由錯誤的全武器 5／10 次改為原版
  固定三次，5／10 限縮為炸彈攻擊當量，接上 >30,000 提前停止，並由 `sub_4267B` caller 證實
  runtime 結果是 `/40` 後直接寫 record `+3`，已取代錯用手冊 UI `/100` 的結算；非炸彈兩版本相同、
  炸彈版本差、逐擊護盾、停止門檻與 `/40` 邊界均有測試。仍待逐武器數量與快速戰鬥 record；
  2026-08-27 已重新閉合 `Fighter_Garrison_Strength_`：舊版 10／6／4 中隊乘泛用戰機近似值已
  改為原版兩組最佳武器扣最佳裝甲、三檔 40/0、40/24、32/24 權重、`/2` 與 64000 上限，並接回
  殖民地反擊正常路徑；`byte_199CB4==1` 的固定 120 模式語意與逐架戰術生成仍分開標未知。證據與
  規格見 [`docs/re/fighter-garrison-strength-audit-20260827.md`](docs/re/fighter-garrison-strength-audit-20260827.md)
  及 [`docs/spec/fighter-garrison-strength.md`](docs/spec/fighter-garrison-strength.md)。
  `Get_Colony_Hits_` 已閉合人口／士兵／戰車／非軌道建築的耐久公式並接入顯示值；
  `sub_E87D2` → `sub_DD2F2 @ 0xDD2F2` → `sub_DCEBD @ 0xDCEBD` 的戰略殖民地內部
  傷亡鏈亦已閉合並實作：一般建築、陸戰隊、戰車、建造進度與人口共用隨機候選池，
  成本不足即停止，並保留最後人口 100 點與建造進度 tail-index 不對稱。仍待逐武器數量、
  快速戰鬥 record，以及 `sub_DCEBD` 排除的八類獨立防禦戰鬥者摧毀鏈。證據見
  [`docs/re/strategic-bombardment-audit-20260824.md`](docs/re/strategic-bombardment-audit-20260824.md)
  [`docs/re/colony-hits-audit-20260824.md`](docs/re/colony-hits-audit-20260824.md) 與
  [`docs/re/strategic-colony-casualties-audit-20260824.md`](docs/re/strategic-colony-casualties-audit-20260824.md)，規格見
  [`docs/spec/strategic-bombardment.md`](docs/spec/strategic-bombardment.md) 與
  [`docs/spec/colony-combat-hits.md`](docs/spec/colony-combat-hits.md) 與
  [`docs/spec/strategic-colony-casualties.md`](docs/spec/strategic-colony-casualties.md)。
- [ ] **艦船經驗、維修與艦艇設計忠實化**：依 `Do_All_Ships_XP_Check_ @ 0x14A27`、
  `Repair_Ships_At_Colonies_ @ 0x580F5`、`Weapon_Cost_ @ 0x6EC74`、`Weapon_Space_ @ 0x6EE8E`、
  `Auto_Design_Ship_ @ 0x616A5` 追回逐艦 XP、等級門檻、學院／領袖加成，以及科技小型化、
  arc／mods 疊加、彈藥架、特殊裝置、艦體成本／空間與 AI 自動設計。2026-08-24 已閉合
  `Repair_Ships_At_Colonies_` 與 `Repair_Ship_Full_ @ 0x581F3`：一般玩家僅
  `COMBAT_SHIP`、靜止、合法 Star 且 star `+0x38` owner bit 成立者靠港完整修復；Go 已排除
  殖民船／前哨船並保留殖民地／前哨站基地強推論。原版會恢復八武器槽並清除護盾、引擎、
  電腦、特殊裝置、裝甲與結構損傷；remake 只有聚合 `Ship.Damage`，逐系統模型仍是明確缺口。
  同日亦閉合 `Do_All_Ships_XP_Check_` 的每回合鏈：原版掃描 500 筆、接受 `Status < 5` 與
  `Owner < 8`（包含支援艦），每回合增加 1、同星系每座己方太空學院再加 1，並取玩家活動
  領袖中最強 Instructor 加成；結果上限 500。等級門檻為 50／150／500／1000，Warlord
  將顯示等級提升一階且最高仍為 5。Go 已依此加入每回合 500 上限並保留支援艦消費端；
  `sub_4B184 @ 0x4B184` 的戰後 XP 鏈亦已閉合：winner-side 且仍連到持久 Ship 的每艘艦
  取得 `max(1, floor(被摧毀敵艦 1-based 艦體級總和/2))`，零擊沉仍為 1，俘獲不計入，
  並直接加到 `Ship+0x72` 而不套每回合 500 上限。快速與格子戰術兩路徑均已接線；
  快速結算改依 combatant 原始索引移除真實陣亡參戰艦，並依手冊 p.119 明確移除遭交戰的
  支援艦；戰術 command 保存摧毀艦體級總和以維持重播一致。原版 battle record
  `+0x24／+0x4B` 的完整 enum 與 AI 逐艦 XP
  仍受現有資料模型限制。武器設計方面亦已由 IDA 閉合 `Weapon_Cost_`／`Weapon_Space_`：
  成本微型化為 100／75／55／40／30／25%，一般武器佔格為 100／80／65／50／35／25%；
  改造先套 HV／PD、再套其餘加總百分比，AF 已由錯誤固定 +50 訂正為 +50%。Go 已把後續
  主題鏈等級接到造艦 UI、成本、佔格與建造時的改造門檻。飛彈彈架亦已由 IDA 閉合
  2／5／10／15／20 發對應 10／20／30／35／40 的成本／佔格，並接到設計 UI、建造命令、
  舊存檔回退、快速結算與格子戰術的逐發消耗。`sub_6F11C` 的三種佔格 category 與
  `sub_6D048` 的八領域 Hyper 重複等級亦已閉合：一般武器、魚雷／特殊武器、固定不縮小
  機庫／裝置分別套正確階梯，特殊裝置使用自己的科技等級，Hyper 可重複研究、存檔並影響
  造艦與改裝。`Player_Research_Cost_ @ 0xE1E96` 與 `Promote_A_Hyper_Value_By_Field_ @ 0x10F919`
  已閉合後續 Hyper 成本為版本基礎值加已完成 level×10000；玩家、AI、顯示與舊存檔遷移均已接線。
  `Auto_Design_Ship_ @ 0x616A5` 的五個 caller、99-byte 六艦體設計庫、八類 switch、八武器槽／
  八特殊槽上限及各類 missile／fighter／beam／bomb helper 已由 IDA 閉合；Go 已保存 raw role、
  依科技建立合法單槽 loadout，並接到正常設計畫面的初始最佳裝備。runtime 與 JSON 快照已能
  保存原版八個武器記錄及特殊裝置 raw ID；`.GAM` 匯入不再只取第一槽，且保留合法的
  `ARC_360 = 16`。2026-08-24 已建立玩家側六艦體持久 `ShipBlueprint`：舊存檔依當前科技補齊，
  元件修改、JSON、熱座與 TCP 快照均可往返；設計畫面改為點艦體只切換／保存 blueprint，只有
  底部 BUILD 扣款造船，不再共用一份巡洋艦暫態選擇。同日再以 IDAPython 閉合
  `Update_Player_Ship_Designs_ @ 0x57112`：真人戰術分支遍歷六筆設計，只更新最佳裝甲、曲速引擎、
  FTL 速度與造價，保留玩家武器／特殊裝置；Go 已在研究完成、研究抉擇、間諜偷竊與遺跡科技
  入口同步六筆裝甲，曲速與造價沿用既有即時計算，並以深複本測試鎖住其餘欄位不變。原版
  `Update_Equipment_On_Ships_Being_Built_` 的 raw status 4／6 在 remake 沒有對等的未完工 ship record，
  目前以交付時讀取最新設計／科技對映，僅屬資料模型近似。2026-08-24 多槽建造與快速結算
  已形成第一條消費鏈：`BuildShipDesign` 深複製全部 `WeaponMounts`／`SpecialIDs` 到新艦，快速
  結算保持單艦一份 HP／護盾／狀態，依槽序與 `WorkingCount` 逐發射擊，飛彈彈藥亦逐槽扣除。
  格子戰術亦由 `StartCombat` 深複製多槽，僅在多槽／多門時進入逐槽派送，逐槽套名稱、類型、
  射界、傷害與彈藥；單槽路徑與亂數順序維持不變，匿蹤、儲能、`Fired`、點防禦與戰損仍是
  同一艘艦的狀態。同日完成多槽造價／空間與玩家編輯：艦體／自動設備／特殊只計一次，每槽
  依武器、typed mods、微型化、射界、彈架求單門值後乘 1..99 數量；UI 可選八槽、增刪槽、
  調數量並逐槽編輯武器／mods／射界／彈藥，顯示、合法性、扣款與完整 blueprint 重播命令共用
  同一計算。未知 raw 武器或只有未解 `RawMods` 的槽採失敗即關閉，不免費建造。35 張 Docker +
  Xvfb 畫廊已重跑並抽查 `25_shipdesign.png`，新增控制文字均在框內。2026-08-24 收尾再加入
  typed `ShipSpecialMount` 八槽：`.GAM` raw ID 已知者轉受控名稱、未知者保留並失敗即關閉；
  設計 UI 可切換／增刪特殊槽，成本、空間、JSON／多人快照與完整建造命令均保存。所有玩家玩法
  消費端改走 `shipHasSpecial`，快速與格子戰術可同時套用多項攻防／速度／連射／偵測／研究／登艦
  效果；多種機庫會各派一隊且快速結算各計一次。2026-08-25 已接格子戰術逐武器槽的可用／待命／關閉、狀態色、有界名稱／數量／彈藥回饋與右鍵明細；待命會略過下次同艦射擊後恢復，關閉持續保留。同日再以 `Load_Combat_Ship_ @ 0x4954A` 的八槽載入鏈閉合逐槽 PD：快速與格子戰術依槽序／工作數自動迎擊，紅色關閉仍會對飛彈／戰機開火，攔截餘數跨槽與回合保留；敵方 typed 飛彈還擊亦有消費端。證據見 `docs/re/tactical-weapon-status-audit-20260824.md` 與 `docs/re/tactical-point-defense-slots-audit-20260825.md`，規格見 `docs/spec/tactical-point-defense-slots.md`。仍待戰略模式與一般 AI blueprint／逐艦設計／建造；不可用殖民地
  2026-08-25 再以 IDA caller 資料流證實 AI 初始 hull 0..5／role 0..5，科技更新只重建
  hull 0..4；Go 已保存每個 AI 的六筆 blueprint、實艦與造艦進度，依 AI 自身科技設計，並接上
  快速／格子戰術、戰損、存檔與網路狀態。事件 13 的「叛變移交艦艇」已由 1.31 IDA 證據
  推翻並移除，不再列作 AI blueprint 消費端。`FleetStrength` 現為實艦衍生摘要；
  AI 對 AI 與 raid 的高階比例結算會回寫移除實艦，單一主力艦隊亦接每回合／太空學院／
  Instructor XP。仍待戰略模式、AI 多艦隊、原版精確 AI 生產評分與未解 raw mods，
  不以目前可重現的最高可負擔 hull 政策冒稱原版 exact。
  同日再由 `Crew_Boarding_Combat_Bonus_ @ 0x35CAD`、`Boarding_Action_Type_ @ 0x2C129`、
  `Get_Fleet_Commando_Bonus_ @ 0x35FDA` 與 `Get_Fleet_Security_Bonus_ @ 0x360AA` 閉合登艦加成鏈：
  攻守雙方各自消費艦員 Bo 與同 owner 有效參戰艦軍官的最高 Commando，守方另取最高 Security
  及保安站 +20；Go 的快速／格子戰術已各自帶入玩家或 AI 的種族、科技、High-G、Powered Armor
  與艦員資料，不再把玩家陸戰隊數值套給 AI。這一切片已對齊；其餘艦船設計待辦不受影響。
  證據見 [`docs/re/ship-repair-audit-20260824.md`](docs/re/ship-repair-audit-20260824.md)，規格見
  [`docs/spec/ship-colony-repair.md`](docs/spec/ship-colony-repair.md)；XP 證據與規格分別見
  [`docs/re/ship-crew-xp-audit-20260824.md`](docs/re/ship-crew-xp-audit-20260824.md) 與
  [`docs/spec/ship-crew-xp-turn.md`](docs/spec/ship-crew-xp-turn.md)；戰後 XP 見
  [`docs/re/ship-battle-crew-xp-audit-20260824.md`](docs/re/ship-battle-crew-xp-audit-20260824.md) 與
  [`docs/spec/ship-battle-crew-xp.md`](docs/spec/ship-battle-crew-xp.md)；武器設計見
  [`docs/re/weapon-cost-space-audit-20260824.md`](docs/re/weapon-cost-space-audit-20260824.md) 與
  [`docs/spec/weapon-cost-space-miniaturization.md`](docs/spec/weapon-cost-space-miniaturization.md)；飛彈彈架證據與規格見
  [`docs/re/missile-ammo-rack-audit-20260824.md`](docs/re/missile-ammo-rack-audit-20260824.md) 與
  [`docs/spec/missile-ammo-racks.md`](docs/spec/missile-ammo-racks.md)；分類／Hyper 證據與規格見
  [`docs/re/weapon-mini-categories-hyper-audit-20260824.md`](docs/re/weapon-mini-categories-hyper-audit-20260824.md) 與
  [`docs/spec/weapon-mini-categories-hyper.md`](docs/spec/weapon-mini-categories-hyper.md)；重複成本證據與規格見
  [`docs/re/hyper-repeated-cost-audit-20260824.md`](docs/re/hyper-repeated-cost-audit-20260824.md) 與
  [`docs/spec/hyper-repeated-cost.md`](docs/spec/hyper-repeated-cost.md)；自動設計證據與規格見
  [`docs/re/auto-design-ship-audit-20260824.md`](docs/re/auto-design-ship-audit-20260824.md)、
  [`docs/spec/auto-ship-design.md`](docs/spec/auto-ship-design.md) 與
  [`docs/spec/multi-slot-ship-blueprint.md`](docs/spec/multi-slot-ship-blueprint.md)、
  [`docs/spec/player-ship-design-library.md`](docs/spec/player-ship-design-library.md)、
  [`docs/re/update-player-ship-designs-audit-20260824.md`](docs/re/update-player-ship-designs-audit-20260824.md) 與
  [`docs/spec/update-player-ship-designs.md`](docs/spec/update-player-ship-designs.md)、
  [`docs/spec/multi-slot-build-and-quick-combat.md`](docs/spec/multi-slot-build-and-quick-combat.md)、
  [`docs/re/ai-ship-blueprint-build-audit-20260825.md`](docs/re/ai-ship-blueprint-build-audit-20260825.md) 與
  [`docs/spec/ai-ship-blueprints-and-build.md`](docs/spec/ai-ship-blueprints-and-build.md)。目前部分特殊
  元件成本、未解 raw mods、特殊裝置逐槽效果、戰略模式與 AI 精確生產評分仍是 remake 值、
  固定政策或尚無完整證據。
- [~] **議會與勝利條件忠實化**：`ceil(population/10)` 已依 `sub_15B90 @ 0x15B90` 原始指令、
  唯一 caller 與票表／總票數 consumer 修正。2026-08-24 再以 IDAPython 閉合
  `Check_For_Council_Meeting_ @ 0x168AF`：首次最早第 25 回合、後續固定間隔 25、至少 3 個
  存續帝國，已殖民星數採 `trunc(total/2)`；`sub_15239 @ 0x15239` 證實第一次召開後寫旗標。
  Go 已移除固定 8 回合與任意早期立即召開，加入 Turn 24/25、49/50、奇數星與快照往返測試。
  證據見 [`docs/re/council-schedule-audit-20260824.md`](docs/re/council-schedule-audit-20260824.md)，
  規格見 [`docs/spec/council-schedule.md`](docs/spec/council-schedule.md)。本輪另以
  `Council_Votes_ @ 0x15EBC`、`Vote_Check_ @ 0x16021`、`sub_161E4 @ 0x161E4` 與
  `sub_1633C @ 0x1633C` 閉合候選排序、AI 對兩候選人的獨立 1..200 檢定、雙通過／雙失敗棄權、
  真人三選一、含棄權票的 2/3 分母與上屆選票保存；固定關係門檻已移除，議會亂數與待投票狀態
  可存讀。正式政策、貿易／研究協定、Charismatic、Repulsive、Imperium 與上屆選票分數已接；
  `+0x617/+0x6D7/+0x7EE/+0x827`、`sub_78398` 與高難度真人縮放仍待外交資料模型閉合，不以
  `Relation` 重複猜填。證據見 [`docs/re/council-voting-audit-20260824.md`](docs/re/council-voting-audit-20260824.md)，
  規格見 [`docs/spec/council-voting.md`](docs/spec/council-voting.md)。
- [ ] **歷史、分數與客製種族效果忠實化**：依 `Record_History_ @ 0x10208A` 及最終分數消費端
  追回欄位、取樣頻率、正規化、勝利加成與存檔；依 `Convert_Custom_Race_Flags_ @ 0x5BC24`
  對 22 項選項逐能力追到原版經濟、外交、間諜、戰鬥、星圖、事件與 AI consumer。現在每項已有
  至少一個 remake 消費端，只能證明選項不是裝飾品，不能宣稱效果與原版逐值／全路徑一致。
  2026-08-25 已閉合八項最終分數 orchestrator、未使用 Picks／Evolutionary Mutation 倍率及
  `(raw×percent+50)/100`，並移除無證據負分 clamp；倍率已接客製種族、JSON 與熱座。
  同次 IDA 證實歷史圖是 Fleet／Technology／Population／Buildings 四項、350 格 ring 與動態
  divisor 重縮放，推翻舊 400 筆 Population／BC／Fleet 模型；2026-08-25 已實作四項共同除數、
  既有資料重縮放、350 筆上限、繁中／英文 UI、JSON 除數往返及不相容舊歷史清除。科技 raw
  `player+0x224` 的 remake 對應採已完成主題／Hyper 成本累積，證據等級仍為強推論；本項剩餘
  範圍是客製種族 22 項效果的原版逐值／全 consumer 閉合，不再重開歷史資料形狀。
  2026-08-25 已由 `sub_E3456 @ 0xE3456` 閉合同化 raw 尺度：八政體 rate 為
  `30/60/30/60/60/120/12/16`、門檻 240；Charismatic 進度加倍且優先於 Repulsive 減半，
  異族管理中心的 120 基礎值也受兩者修正。Go 已改存 0..239 raw 餘數，修正政體／建築切換、
  UI ETA 與舊 JSON 回合進度遷移；packed colonist 清除順序仍是可重播近似。
  證據與規格見 [`docs/re/history-score-audit-20260825.md`](docs/re/history-score-audit-20260825.md) 與
  [`docs/spec/final-score-multiplier.md`](docs/spec/final-score-multiplier.md) 與
  [`docs/spec/history-ring.md`](docs/spec/history-ring.md)。
  同化證據與規格見 [`docs/re/assimilation-race-traits-audit-20260825.md`](docs/re/assimilation-race-traits-audit-20260825.md) 與
  [`docs/spec/assimilation-race-traits.md`](docs/spec/assimilation-race-traits.md)。同日 Lucky 事件亦由
  `sub_245C4/sub_24511/sub_22D57/sub_2230A` 閉合：一般壞事件選中 Lucky 玩家時取消，不再把
  一般事件池錯誤過濾成全好；逐回合計數器、四種設定除數、`Random(1000)` 門檻、相對第 50 回合
  閘門、額外好事件池、JSON 與熱座往返均已接線。標準 remake 尚未分開暴露 Random Events／
  Antaran Attacks，執行期採兩者皆開的原版除數 8。一般事件固定 30% 已由
  `sub_2230A @ 0x2230A` 的前五次保護、五級 delta 公式、`Random(512)`、五次 0..28 候選與
  事件最早日期取代；排程狀態可 JSON 往返。全局目標不再依熱座席位數重複排程，非目前席位
  與 AI 基本 BC／RP 事件可寫回；其餘 AI 複合效果仍屬事件系統待閉合範圍。
  Lucky 證據與規格見 [`docs/re/lucky-event-audit-20260825.md`](docs/re/lucky-event-audit-20260825.md) 與
  [`docs/spec/lucky-events.md`](docs/spec/lucky-events.md)；一般排程見
  [`docs/re/random-event-schedule-audit-20260825.md`](docs/re/random-event-schedule-audit-20260825.md) 與
  [`docs/spec/random-event-schedule.md`](docs/spec/random-event-schedule.md)。全銀河目標的純權重亦已依
  `sub_22D57` 閉合；全局目標索引、Lucky 第一成功槽、熱座非目前席位及 AI 基本 BC／RP 回寫
  已接線，AI 殖民地／艦隊／外交與持續 record 等複合效果仍待閉合。證據與規格見
  [`docs/re/random-event-target-audit-20260825.md`](docs/re/random-event-target-audit-20260825.md) 與
  [`docs/spec/random-event-target-selection.md`](docs/spec/random-event-target-selection.md)。
- [ ] **原版 parity 發行閘門**：在上述核心矩陣完成前，三平台包與推廣片只能稱「可玩 remake
  預覽」，不能稱完整原版重製。測試需分開標示 remake 自洽、IDA 靜態證據、原版動態 oracle、
  同狀態視覺與正常玩家路徑；綠色單元測試不得升格為原版一致。

#### 已完成的 remake 工程能力（不代表原版玩法已對齊）

- [x] `ARM/FST` 武器改造與兩條攔截鏈、PD 餘數保存；魚雷 `NR` 射程衰減取消；敵我戰機出擊、PD、最弱護盾面、命中／爆炸／返航模型已接入設計、存檔、快速結算與格子戰術。
- [x] `RACES` 間諜操作（訓練、`STEAL`／`SABOTAGE`／`HIDE`、防守 Agent）、外交餽贈／特殊貿易／條約／協議／納貢，以及艦艇與殖民地 `LEADERS` 的指派、改派、解除、僱用、解雇、存檔與熱座流程。
- [x] 客製種族 22 項選項與心靈感應、幸運、全知、匿蹤艦等消費端；`hover` 外框、`CMBTSFX` 多幀特效、`STREAM`／`STREAMHD` 音樂與音效接線。
- [x] 文字欄寬 polish：共用 `labelRect`／`extraText` 截斷與折行已套用主要長文字面板；Docker + Xvfb 產出 35 張 1280×960 畫廊並抽查外交、艦艇設計、輸入框、種族、戰術、熱座，README 與 `docs/PLAYTEST-2026-08-10.md` 已同步。
- [x] **遊戲內文字版面與發行畫廊回歸（2026-08-12，視覺）**：主選單語言／規則標籤維持 196px 安全欄寬；資料來源缺少 `COMBAT.LBX` 時仍改繪同座標、可點擊的戰術後備控制列。使用者提供的正常路徑截圖推翻了舊的「畫廊未越框」斷言後，新增共用 `textSafeRect`：`NEW GAME` 選擇器、種族按鈕、`RACES` 的 AI 資訊／對談列、殖民地的名稱／產出／建築／職業與按鈕文字，現在各自綁定實際面板的寬度、高度、行距、內縮與整數像素中心；超量內容會折行後省略，不跨入滑桿、下一列或按鈕。Docker + Xvfb 目標測試、35/35 `-gamegallery`，以及實際 `-promo-demo` 正常點擊路徑停在新局、種族、殖民地與 `RACES` 的截圖抽查均通過。此為 remake 視覺驗收，不宣稱原版逐像素對照。
  2026-08-12 再依使用者實拍訂正兩個自洽但錯誤的幾何假設：`NEW GAME` 數值列先前重複加上 `(15,5)` 畫面原點，造成熱區與文字一起偏右下；現改用完整背景量得的可視列矩形。殖民地產出第二列先前從 `y=53` 起畫，16px 字墨跨在 `y=58` 分隔線上；現固定為 `[59,75)`，與第一列、分隔線及 `Buildings` 標題各自分離。回歸測試改讀獨立量測錨點與實際 bitmap 字墨高度，不再拿同一組錯誤常數互證。
- [x] **TCP 多人正式入口與缺資產備援（2026-08-11）**：`NETWORK`／`JOIN GAME` 不再被誤畫為未實作；數據機、序列線、`COMM INFO` 與 TEN 維持灰色且熱區不處理。`NetworkStateHash` 會先正規化每台機器本地的作用中席位，避免主機席位 0 與客戶端席位 1 套用同一快照就誤報分岔。完整資料包缺 `MULTIGM.LBX` 時，改以原版已證實的 482×335 面板與按鈕熱區繪製可讀備援，避免多人頁塌到左上；README 的 `23_multiplayer.png` 已換成此輸出。Docker 抽樣包含實際 `NETWORK`／`START NEW GAME` 名稱彈窗、TCP 加入、共同開局、第一回合 `turn_done` → `turn_ready`，另跑 `internal/netplay` 全包（含斷線、重連、心跳）與 35/35 畫廊，抽查多人頁；不宣稱已做人類跨 NAT 實機驗收。
- [x] **人口成長回寫定點測試（2026-08-11）**：`TestPopulationGrowthWriteback` 已改驗真正不變量：人口 8→12 且不超過上限，農／工／科職務總和必須等於人口，新增者至少進入一種職務。舊「工人必增」斷言與既有 `assignNewColonist` 的缺糧保護衝突；實際固定開局會把 4 位新增人口改派農夫（4/2/2→8/2/2），不是寫回失敗。人口上限測試亦通過。

#### 其他仍會影響 remake 工程交付的工作

##### 發行前的實作／視覺／驗證

- [~] **全畫面動態文字安全框（2026-08-12，進行中）**：檢測與回歸流程已寫入
  [文字版面非視覺檢測流程](docs/tech/text-layout-geometry-gate.md)。目前已把 `INFO` 五個子頁、
  `netNextTurn`、是／否 `confirmbox`、文字輸入框、`HI SCORE`、星圖選取星球面板、`RACES`
  間諜／外交卡片，以及 `NEW GAME` 選擇器值列改為由所屬面板或按鈕推導的 `textSafeRect`。
  繁中、英文、玩家名稱、聊天室與多行訊息均以實際字型量測寬度、行數、內縮與相鄰控制項邊界；
  單行內容會保留游標空間後截斷，多行內容會固定行數後省略。`CONFIRM.LBX` 缺席時現在畫原位
  313×227 的不透明後備面板，讓星圖名稱不會穿過訊息。已為轉換檔案加 AST 直繪護欄；Docker +
  Xvfb 目標測試與 35/35 `-gamegallery` 通過，並以正常玩家路徑抽看 `RACES`、輸入框、確認框、
  新局選擇器與星圖面板。**尚未宣稱全庫安全**：其他靜態候選路徑仍須先找出文字擁有面板、補安全框
  與幾何測試，不能以全域替換直繪呼叫冒充修正。
  2026-08-27 重新逐張檢視 README 實際引用圖後，曾確認 `25_shipdesign.png` 的六艦體空間表
  跨過原版上下內容面板分隔框，`23_multiplayer.png` 的熱座說明則位於 482×335 主面板外；根因是
  舊測試只驗文字框彼此不重疊，沒有驗證其包含於背景美術 panel rectangle。依使用者決策，艦艇
  設計已採現代「選取艦體 → 顯示目前詳情」模式，只顯示目前艦體的已用／總空間及容量條，完整
  武器改造區保留；多人說明與錯誤訊息則共用 TEN／取消鈕之間的面板內兩行區。新增測試會檢查
  `contained by` 所屬美術面板，而不只檢查 `non-overlap`。繁中／英文正式資料畫廊各 38/38 通過，
  兩張問題實圖均目視確認；README 引用的九張 remake 圖片也已同步到同一輪繁中畫廊。這只關閉
  已抽查的具體缺口，不代表全庫安全框已完成。
  同輪發現版控 `30_netwait.png` 仍停在 2026-08-11 的無正式面板舊基準，狀態指紋 `d5e9a2ba`
  與目前繁中 `7024b415` 不同。此差異早於本輪 UI 改動；目前相同資料、語言與參數連跑兩次，
  `30_netwait.png` SHA-256 均為 `0bf9d1eb…43a395a`，多人與艦艇設計圖亦逐位元一致，因此已把
  雙語 netwait 圖更新到目前可重現基準，而不是未查原因就接受指紋漂移。
- [x] **殖民地生產控制（2026-08-11）**：`BUY` 依目前進度扣 BC、於 EndTurn 完成；`AUTO BUILD`
  有明示的固定 remake 優先順序；`REPEAT BUILD` 可循環殖民船、前哨船、運輸艦隊與其他可重複
  Special；`REFIT` 可選停泊同星系的戰鬥艦、檢查 Cruiser 以上的 Star Base 門檻、保存凍結的
  自動最佳模板、完成後回到原星系，取消則依手冊報廢來源艦。所有狀態進 JSON／熱座／TCP 指令重播；
  原版 BUY 中段價格、AUTO 判斷與設計庫選取明確標為強推論或 remake 近似，詳見
  [殖民地生產控制](docs/tech/colony-production-controls.md)。
  Docker 抽樣已通過 BUY 邊界／完成、AUTO 保存、REPEAT Special、REFIT 保存／門檻／完成／取消，
  以及 UI 的 AUTO／REPEAT／REFIT 預覽；命令重播刻意走 `MarshalSnapshot`／`RestoreSnapshot`，
  不以會共用 Go 切片（slice）底層陣列的內部 `snapshot().restore()` 假裝兩台獨立機器。建造與改裝畫面的
  可變長標題、訊息與艦名均已截斷或折行，35 張 Docker 畫廊啟動抽樣通過，代表圖為
  `docs/screenshots/26_buildqueue.png`。
- [x] **目前可玩預覽包的打包工程驗證（2026-08-12）**：Linux AppImage、Windows amd64 ZIP、macOS
  universal tar.gz 已由同一份最新程式碼離線重建，`dist-all/PUBLIC-SHA256SUMS` 已重算並逐檔驗證。
  Linux 以外掛合法資料／字型實際跑過 Docker + Xvfb 的 `-game -promo-demo -noaudio` 導覽；Windows
  ZIP 通過 CRC、根目錄、PE GUI／console 格式檢查；macOS `.app` 通過 tar 結構、啟動器語法與
  `lipo` 的 `x86_64`／`arm64` 檢查。三包均確認不含原版 `.LBX`、音樂、音效或私有字型；真 Windows／
  macOS 主機執行仍不宣稱已驗收。帶使用者私有資料的完整版只保留本機驗收用途。這些產物是
  可玩 remake 預覽，不通過上方「原版 parity 發行閘門」，不得沿用「完整原版重製」稱呼。

##### 需要外部裝置才能完成的驗收

- [ ] **外部音訊驗收**：在有音訊輸出的桌面逐曲聽 `STREAM`／`STREAMHD` 與場景切換，確認音量與曲目；
  Docker 的解碼、長度、峰值、非靜音檢查已完成，不能冒充人耳驗收。
- [x] **Smacker 過場音軌（2026-08-27）**：`internal/smk` 已解 MOO2 真實資產使用的 packed
  8-bit mono／stereo PCM，11025 Hz 會轉為 Mixer 的 22050 Hz；互動路徑會取代背景曲，跳過或
  播完後停止並恢復背景音樂，截圖廊不初始化音訊。IDA、官方格式來源、真 `INTRO.LBX` 與測試證據
  見 `docs/re/cutscene-audio-audit-20260827.md`；DAC／PIT／MSS 逐週期內部依停止線不追回，
  有音訊輸出的桌面抽聽仍由上方「外部音訊驗收」統一追蹤。
- [ ] **最新工作樹重新打包、推廣片與平台真機驗收**：2026-08-12 的 `dist-all` 三平台包與推廣片
  早於目前議會公式、文字版面及後續 source 變更，只能當歷史產物；所有發行 gate 完成後，以同一
  `HEAD` 重新產生 Linux AppImage、Windows amd64 ZIP、macOS universal 包與實際玩家路徑推廣片，
  更新雜湊／metadata，並在 Windows 與 macOS 真機各做啟動、資料路徑、音訊、輸入、存讀檔 smoke。
  在原版 parity 發行閘門完成前，重建產物仍只能標為「可玩 remake 預覽」。

- [x] **captain／common 領袖技能消費端（2026-08-10；招募機率 2026-08-25 補證）**：26 項技能已有至少一個 remake 消費端，包含 Assassin 的逐位行動、Diplomat 的有標註關係代理值、Famous 的雇用費折扣與逐回合招募機率、Megawealth 的回合 BC／維護費、Operations 指揮點數、Spymaster／Telepath 間諜攻防、Galactic Lore 星圖／怪獸／安塔蘭、Ordnance／Security 兩條戰鬥路徑，以及 Fighter Pilot。Tactics 依手冊「This skill is not implemented」保留無效果。細節與測試見 `docs/tech/leader-officer-skills.md`。
- [x] **英文模式安全 fallback（2026-08-10）**：未知自訂名稱、未知族群／艦艇／建造項目、熱座名稱等顯示值改走 ASCII 保留或通用英文 fallback，不改存檔 key；英文 `-gamegallery` 在 Docker + Xvfb 產生 35/35 張 PNG，抽查主選單、星圖、外交、艦艇設計、輸入框，沒有越界／panic／fatal／error。AST 英文棘輪維持 16 條可解釋例外；不把查表 key 全域翻譯。
- [x] **一次 20 回合開局經濟／士氣探針（2026-08-10）**：固定無事件開局 BC 50→264（首回合結算後 58）、人口 8→11、士氣 0% 全程、食物輸出 0→1、工業 6→8、研究 6 維持；沒有負食物或人口死亡螺旋。本輪只記錄體感基線，不擅自改收入公式；測試為 `internal/shell/economy_20_turn_test.go`。
- [x] **本輪 remake gameplay polish（2026-08-11）**：外交 `FoodForCredits`／`ResearchExchange` 特殊貿易已有回合收益與存檔狀態；AI→玩家 `SABOTAGE` 會依 remake 性格政策選任務並讀取玩家建築池；食物複製機補上半食物／半 BC 計算與跨回合 BC 餘數保存（強推論，不冒充原版碎片付款已證實）；原版 `RawStatus=4` 的 `+0x37`／30 門檻已接成領袖清理路徑；IDA 追回活動 Trader 的 raw 經驗分桶、tier 1/2 `×10/×15` 最大加成，並接入 GAM／demo fallback 的貿易目標；本輪再接 `CMBTSHP` 固定 tick 近似、事件／爆炸 strategic consumer、SABOTAGE raw slot helper 對齊／結構化分數／Agent 實際扣除與領袖 ETA callback 近似。這些項目只證明可玩消費端存在；原版 CMBTSHP clock、score table 上游填值與 raw callback 設計／帝國欄位仍須併入核心 parity 矩陣，詳見 [`docs/re/remake-consumer-closure-20260811.md`](docs/re/remake-consumer-closure-20260811.md)。
- [x] **議會／安塔蘭視覺收尾（2026-08-11）**：`COUNCIL.LBX#1` 10 幀與 `ANTAROOM.LBX#1` 55 幀已由原版資產逐幀累積並接入播放；`CMBTSHP` 已接原版 `45*playerColor+rawPicture` 圖片映射，並以移動後固定 tick 播放近似 timer，靜止後不自行旋轉。原版 timer 仍標未知。畫面尺寸、雜湊、外部截圖與未宣稱項目見 [`docs/re/visual-oracle-20260811.md`](docs/re/visual-oracle-20260811.md)。
- [x] **本機完整版與實機推廣片（2026-08-12，最終錄製與剪輯）**：`dist-all/` 集中保留三個僅供本機驗收的完整包（含使用者私有資料子集與 CJK 字型，不可公開散布）；其 SHA-256 核對檔列於 `dist-all/SHA256SUMS`。三平台完整包均由本輪工作樹離線重建，採 55 個經靜態消費端與封裝後畫廊交叉確認的正常玩家路徑 LBX（含 `COMBAT`、`INBOX`、`MULTIGM`、`RACESEL`、`STREAM`／`STREAMHD`），不是把 373 個原版檔全塞入；Linux 完整 AppImage 已在 Docker + Xvfb 從包內資料啟動，Windows ZIP 的 CRC／PE／根目錄、macOS universal `.app` 的結構／`lipo` 與三包的 55 檔集合亦已抽樣驗證。正式預覽片以封裝後 AppImage 在 Docker + Xvfb 走 `-game -promo-demo -promo-hide-cursor -noaudio`，經 21 個正常 UI 點擊實際完成新局、種族、命名旗色、星圖、殖民地人口調配、`RACES` 間諜操作、外交、戰術移動／射擊、撤離與戰果回寫後返回 `RACES`；不以截圖輪播冒充遊玩。依復古遊戲推廣片流程，另用 `scripts/make_live_promo.sh` 加入 4 秒銀河戰略風片頭、五個短暫章節識別與 5 秒片尾；章節框只佔 4:3 畫布外的左右側欄，不遮蓋遊戲 UI。成片為 72.767 秒 H.264/AAC、1280×720、30 fps、48 kHz stereo；完整解碼無警告，音量平均 -18.3 dB、峰值 -2.9 dB，七個代表影格確認片頭、五章實機與片尾文字均在安全框內。錄製若未收到 `LastBattle` 寫回的 `promo-demo: complete` 就拒絕輸出；新局與種族按鈕文字均以按鈕內安全區置中，導覽游標完全隱藏。原版 `STREAM.LBX` 配樂只在錄影後混入，未做逐曲人耳驗收，影片仍不可當作可公開散布素材。錄製、剪輯與權利邊界見 `scripts/capture_promo_gameplay.sh`、`scripts/make_live_promo.sh` 與 `dist-all/promo/README.md`。
- [x] **音訊文件同步（2026-08-11）**：`docs/tech/audio-track-map.md` 已移除「外交音樂是單一 `bgmDiplo` 常數」的過期敘述，改記逐族好曲、壞曲池與原版門檻未知；IDA Pro 靜態證據見 [`docs/re/oracle-static-ida-20260811.md`](docs/re/oracle-static-ida-20260811.md)。
- [x] **多人網路最低可玩鏈與可選可靠性（2026-08-12）**：保留原版決定性 lockstep、以 TCP 取代已失效的 IPX／數據機／序列／TEN。`cmd/moo2` 已從大廳名冊接到主機共同新局（設定／種子／席位快照廣播）、客戶端套用同一快照、玩家指令依席位與玩家編號收集／重播、`turn_done` → `turn_ready` 兩階段回合結算，以及 `NetworkStateHash` 不一致時失敗即關閉；另已加入 resume token 重連、心跳／逾時／重連寬限、challenge-HMAC 身份驗證與可選 TLS 1.3。`MOO2_NET_AUTH` 開啟共享密碼 proof，`MOO2_NET_TLS=1` 開啟加密；NAT 穿透仍需外部 relay 或 UPnP，沒有冒稱內建解法。共同開局的非主機席位現在在建立時即初始化 `AutoBuild`／`RepeatBuild` 平行向量，避免只有客戶端切換該席位後出現 nil／空 slice 形狀差異而誤報分岔。驗證包含 `internal/netplay` loopback TCP／TLS／重連抽樣、`internal/shell` 快照／席位重播與 UI 指紋正規化，以及 Docker + Xvfb `cmd/moo2` 共同開局、首回合 lockstep 與等待階段聊天共存測試。正式 `networkWaitScreen` 已共用原版 `Net_Next_Turn` renderer；`netNextTurnDemo` 只保留無 socket 畫廊資料。
- [x] **可選 AI-to-AI 強化（2026-08-11）**：AI 星選的行星價值／距離模型與議會兩候選人／第三方搖擺票已完成；現在另有可保存、可重播的 AI 彼此戰爭、停戰／互不侵犯／同盟、貿易／研究協議，以及抽象艦隊攻擊最高人口殖民地的戰鬥／佔領解算。這是 remake 模型，不冒充原版逐艦 blueprint；細節見 [`docs/tech/ai-to-ai.md`](docs/tech/ai-to-ai.md)。
- [x] **原版 `.GAM` 匯入（2026-08-11；2026-08-24 多槽勘誤）**：`ImportGAM`／`LoadGAMSession` 已把原版 `.GAM` 轉成可玩的 remake 工作階段；`LoadSession` 依 little-endian `0xE0` magic 自動分流，載入畫面會探測同槽 `SAVE1.GAM`～`SAVE10.GAM`，匯入後另存為 remake JSON，不覆寫原始檔。星系、行星、殖民地／前哨站、玩家／AI、外交旗標、67 筆領袖、艦隊、建築與建造佇列均有對應；八個武器記錄與特殊裝置 bitset 現會完整保存及快照往返，不再靜默丟棄第一槽以外資料。研究完成 byte、特殊裝置 raw ID 的完整語意與原版任命／任期下游維持報告式未知，不猜測。`SAVE10.GAM` 真檔抽樣已通過。
- [x] **敵方戰機下游命中／傷害（2026-08-11）**：ID 31 第二組 `1..4 / 4..16 / 2..7` 與 `sub_3AD57 @ 0x3AD57` 的 1..100 隨機、`roll <= 95` 攻防修正、40 命中門檻、`max-min+1` 插值端點，以及相鄰 `sub_3AC20 @ 0x3AC20` 的直接插值式已分開接入；Bomber profile 走 `ResolveFighterBomb`，兩條結果都逐架進最弱護盾面／裝甲／結構消費。`RawFlags & 4` 的 sub3AD57 表面分支經可達性分析證實不可達；未追回的是攻方加成欄位完整語意、兩份外部索引函式名稱與 raw runtime 輸入，不再把固定 3/5 近似寫成原版值。
- [x] **星際要塞完整已追回火力（2026-08-11）**：`sub_4D18E @ 0x4D18E` 四個槽已保存為 seed／raw／cap `(375,2,99)`、`(187,0,198)`、`(187,4,198)`、`(375,2,99)` 並彙整進安塔蘭終局齊射；class 6 直接 byte stride 讀址 `0x17F69C=900`、`P=750`、`sub_6EE8E @ 0x6EE8E` divisor 中間算式、raw `2/4` 百分比表與 live tech 分支已由 IDA Pro 靜態追回。99/198 是容量上限與拆槽規則，不是固定 runtime 數量；快速戰鬥明示採 full-cap policy，raw flag 正式玩法名稱與 live tech 導出的當下數量仍是 oracle 差異，不阻塞 remake。

#### 核心項目的證據限制（不另開重複待辦）

> 下列內容只補充上方高優先項目的證據邊界，不是第二份工作清單；實際狀態與勾選只維護在上方
> 忠實化項目及 `docs/re/parity-matrix.tsv`，避免同一工作完成後留下另一個未勾項。

- [x] **IDA Pro 靜態 oracle 批次（2026-08-11）**：已用同一份 `Orion2.exe.i64`、IDA Pro 9.4 與非破壞性 IDC 探針確認 `STREAM`／`STREAMHD` 門檻、背景／戰鬥曲池、敵方五級 blueprint writer／非空武器槽、`sub_3AD57` 戰機命中／傷害式、相鄰 `sub_3AC20` 直接插值式、要塞 class 6／seed／raw flag／容量 divisor、外交評分桶與回應分支門檻，以及 `.GAM` `0x3B×0x43` 全局領袖區塊的讀寫對稱。結果與位址衝突、輸入雜湊、證據等級寫入 [`docs/re/oracle-static-ida-20260811.md`](docs/re/oracle-static-ida-20260811.md)；沒有把未知語意或外部符號表別名升格成事實。
- 戰鬥項先以靜態 IDA 補 `sub_3AC20`／`sub_3AD57` runtime 參數、live tech、逐彈逐槽值、
  `ARM` raw bit 與原版艦身 timer caller／consumer；靜態不足才用可啟動 `VESA.COM` 的 DOSBox 做
  邊界實驗。敵方戰機傷害子式與要塞 remake consumer 已存在，但完整戰鬥 parity 尚未閉合。
- 外交／間諜／領袖項仍需特殊 raw 上游／創造力係數、原版 AI 防守策略、一般在職領袖 callback
  的 raw 設計／帝國欄位逐值資料流；SABOTAGE 兩張帝國攻防表的上游已閉合，但 Agent 訓練成本、
  維護與 AI 任務政策仍是獨立的原版忠實度工作。
- 嚴格展示差異包含議會／安塔蘭幀停留、原版 `CMBTSHP` 20 幀 timer、殖民地總覽下排第三格、
  文字描邊／陰影；原版 event record、爆炸 raw 下游與全局 save seed 也仍屬對應核心項目的
  證據限制，不能因 remake consumer 存在就升格為 runtime parity。

#### 可選擴充與明確不做

- AI 星選、`AIRelations`、兩候選人、第三方議會搖擺票、`AIWars`／`AIPolicies`、抽象艦隊與存檔是已完成的 remake 模型；它們不再被用作原版 AI／外交完成證據。原版 `NPC_To_NPC_Treaty_Negotiations_` 與議會投票鏈已列入最高優先 parity 工作。
- TCP 多人最低可玩鏈之外的重連／心跳／加密／身份驗證已完成；NAT 穿透不是本專案內建功能，公網部署仍需外部 relay 或 UPnP，不能把它寫成已解決。
- 不列入工作：`Calc_Tech_Value_` 尚無遊戲消費端的 C–K 候選語意、零消費的內部反組譯探針，以及其他與玩家行為無關的內部功能。除非新 gameplay 證據直接觸發，禁止為它們開新挖掘迴圈。
- 不列入工作：數據機／序列埠／`Comm Info`／TEN 舊硬體或已停止服務；remake 的替代方案是 TCP／熱座。對局開打後中途加入也不列入最低可玩鏈，除非另立產品 scope。

#### 本輪驗收規則

- 每次資料或 UI 修改只跑 Docker 目標測試、`go build`、`-gamegallery` 與代表畫面抽查；使用者已明示不要求完整遊戲測試。
- 任何原版差異必須標成「已證實／強推論／假說／未知」，沒有 `VESA.COM` 就不能把靜態近似升格成 runtime parity。
- 本區完成後才回寫對應文件；`docs/HONEST-STATUS.md` 若出現舊敘述，當輪直接同步，不另建剩餘工作路線圖。
- 本輪實際抽樣：`gamedata`／`engine` 食物複製機、`shell` 特殊貿易／AI SABOTAGE／領袖任期／地面戰、人口成長／上限、`cmd/moo2` Xvfb 單測與 35 張 `-gamegallery` 均通過；使用者已明示不要求把整個 `go test ./...` 包裝成全綠。

### 附錄一、手冊忠實化證據摘要（非待辦）

**目前仍有證據／跨路徑缺口**(2026-08-10；本輪使用者指定的 remake 功能已接完，剩餘項目只列原版 oracle／素材限制)

傳送器已接上 `CombatShip` 四面護盾容量、命中時扣除對應分面、12 格前置與硬化護盾阻擋；
分面索引採固定世界座標的四向近似，原版艦身旋轉與方向命名仍列為證據留白，不把近似寫成已證實。

⚠ 匿蹤力場已從這張表移出:舊理由「AI 艦隊沒有地圖座標」在第 47 項(AI艦隊移動)之後就過期了。
重查結論不變、**理由換了**:AI 的出兵決策讀玩家殖民地,從來不讀玩家艦隊位置,
所以「在星圖上隱形」仍然沒有消費端。缺的是「AI 會不會攔截玩家艦隊」。

### 2026-08-10 飛彈／魚雷／敵方戰機收尾

- [x] `ECCM/EMG/MV` 與魚雷 `ENV/OVR` 已接入設計佔格／成本、既有 `Ship.Mods` 存檔、快速結算、
  格子戰術與依武器類型的設計 UI；MIRV 逐彈頭消耗干擾／匿蹤／位移骰，AMR 只摧毀一枚彈頭。
- [x] `ARM`／`FST` 已接入武器改造選項、佔格／成本、JSON 存檔，以及快速結算／格子戰術的 PD 攔截鏈；ARM 的 raw `0x0800` 對應仍標強推論，FST 的 raw `0x1000` 與速度 +4 已證實。PD quotient/remainder 餘數現在跨同場戰鬥／戰術回合保存，未命中也不清零。
- [x] 魚雷 `NR` 已進入設計 UI、存檔與兩條戰鬥路徑；電漿魚雷按距離每格 −5，`NR` 取消衰減，其他魚雷維持固定傷害。
- [x] 敵方戰機使用強度驅動的 Interceptor／Heavy／Bomber profile，補上 drive、裝甲、戰鬥速度、種族防禦、PD、敵我雙向目標、命中／爆炸特效與返航；這是可重現的 remake 模型，不冒充原版逐艦 blueprint。
- [x] 原版五級敵方設計 writer 與防禦艦非空 weapon slot 已以同一輸入 IDA 追回；`Intruder`／`Interdictor`／`Harbinger` 的 remake raw slot 映射已同步，並修正兩個第 4 槽 `raw flags=0x0004`。要塞已接入已追回直接火力、敵方戰機兩條 raw 傷害式已接入；仍待 runtime／下游 oracle 的是 `sub_3AC20`／`sub_3AD57` 逐彈參數與 ARM raw bit 正式名稱，不阻塞 remake。

### 2026-08-09 艦艇軍官管理切片

- [x] 艦隊畫面選取艦艇 → `LEADERS` → 點艦艇軍官列完成指派、改派與解除；`Ship.OfficerName`
  走 JSON／熱座保存，快速結算與格子戰術的 Weaponry／Helmsman、航行 Navigator、戰後 Engineer
  都讀逐艦指派。
- [x] 軍官畫面的 `HIRE` 進入候選模式，可點指定待僱傭兵；`POOL` 解除指定回人才庫；`DISMISS`
  解雇艦艇軍官並清除逐艦欄位，均有 shell 測試護欄。
- [x] 原版 `.GAM` 的數字 `Officer` ID 已依固定 `_leaders[0..66]` 索引鏈轉成
  `HERODATA`／`shell.Leader.ID`／`Ship.OfficerID`，並保存到 JSON；同一輸入 IDA 另已證實
  `sub_10E2F` 對 `dword_1930DC` 直接讀入 `0x3B×0x43`、`sub_1160B` 對稱寫出。既有
  `OfficerName` 保留給舊 JSON 回退；重製原生 `.GAM` importer 與任命／任期下游規則仍是可選差異，
  證據見 `docs/re/officer-ids.md`。
- [x] 殖民地領袖已加入 `LEADERS` 分頁：可指定、改派、解除與解雇，技能加成會在任職／離職時反向套用，並與艦艇軍官角色互斥；`ColonyLeaderNames` 隨 JSON／熱座席位保存。

### 2026-08-09 正式外交條約與協議切片

- [x] 原版 `+0x627` 的正式外交狀態已以 `gamedata.ForeignPolicy` 對映為和平、互不侵犯、同盟；
  原版 `+0x62F`／`+0x637` 的貿易／研究協議旗標可並存，均可由外交畫面提議、查看、終止並保存。
- [x] 貿易／研究協議依雙方較低人口建立負值投入期，逐回合趨近目標；收益進入玩家與 AI 的 BC／研究
  結算。原版政府倍率與神級商人 +50 個百分點已接；測試：`treaty_test.go`、`engine` 回合結算測試、
  JSON round-trip。
- [x] 原版固定 5%／10% 週期納貢條約已由 `+0x63F`、`sub_52049`、`sub_E1FC7`／`sub_E2710`
  形成可重現切片；`TreatyState`、外交畫面、回合國庫轉移、摘要、存檔與測試均已接。
- [x] 現金／科技／殖民地餽贈、食物換現金／研究交換特殊貿易、正式條約與納貢的提議／終止／回合收益、
  間諜 STEAL／SABOTAGE／HIDE、玩家防守 Agent 訓練／解除均已接入外交／RACES UI、存檔與測試。
  原版 raw 創造力／特殊貿易 byte table、AI 接受門檻、特殊槽位與完整 SABOTAGE 分數仍是 oracle 差異；remake 的 AB／DB／E 與 Agent 訓練／擊殺消費已接，不阻塞 remake。

### 2026-08-09 火線角設計切片

- [x] 原版 `ShipWeapon.Arc` 的 1／2／4／8／15／16 值、Fwd Ext／Back Ext／360 的手冊 +25%／+25%／+50%
  已接入單武器重製模型；設計畫面可循環選擇，成本／佔格、建造判定、JSON 與兩條戰鬥建構路徑均保存／傳遞。
- [x] 已由 `Relative_Bearing @ 0x32AD1`、`Relative_Bearing_XY @ 0x32A20`、`Move_Ship @ 0x3F5F1` 與
  `Ship_Can_Deploy_At @ 0x49043` 證實 16 向朝向與 1／2／4／8 bearing mask；格子戰術的玩家／敵方開火、移動轉向與射界外提示已接，固定測試見 `docs/re/weapon-arcs.md`。
- [x] 原版快速結算的 `QGet_Target @ 0x41F20`、`QGet_Target_SC @ 0x41F80`、
  `Strategic_Combat @ 0x40C2A`、`Missile_Attack @ 0x420C0`、`Strat_Special_Attack @ 0x4221F`
  固定 call graph 未讀取 arc／bearing；快速路徑本來就是抽象傷害，不新增固定 range
  之外的空間模型。格子戰術方向鏈的消費端已接，證據分級見 `docs/re/weapon-arcs.md`。

### 2026-08-10 戰術戰機與視覺收尾

- [x] 玩家戰機中隊的攻擊消費端改讀四面護盾容量，依手冊 p.157 選最弱面，再把護盾剩餘傷害送入裝甲／結構；既有戰機近似武器傷害與命中率假設不變。
- [x] 玩家中隊保存仍有效的主要目標，只有目標失效且尚有射擊時才自動重選；不把原版候選優先分數未知的部分冒充完成。
- [x] 安塔蘭終局防禦艦級與已知戰機艙消費端保存原版 `Intruder ×3`／`Interdictor ×2`／`Harbinger ×7`／星際要塞；Intruder 的 3 個、Harbinger 的 6 個標準 `Fighter Bays` 沿用玩家快速戰鬥戰機貢獻模型，並把 Large／Huge／Titan 對應的目標級數傳入球形武器解算；標準艦非空 weapon ID／數量／raw flags 也已保存；要塞戰力仍是既有代理值。
- [x] 敵方 profile、敵我雙向戰機、戰機接戰前 PD、最弱護盾面、返航補給與 `CMBTSFX` 命中／爆炸序列已接；原版逐艦 blueprint、方向命名與 `sub_3AC20`／`sub_3AD57` 精確 raw runtime 數值仍未證實。
- [x] 共用 hover 外框已接外交、戰術控制列、戰機出擊與可點擊按鈕；`CMBTSFX.LBX` 多幀 delta 畫面改用累積解碼，資產缺失時安全回退。
- [x] `STREAM`／`STREAMHD` 每曲已完成 Docker 技術抽樣（解碼、長度、峰值、非靜音）；真正人耳聽感仍需有音訊輸出的外部驗收，見 `docs/tech/music-listening-qa.md`。

**擋門理由已過期或本來就錯 —— 已於 2026-08-08 全部接完**(第 77 項(元件表真值))

隱形裝置 / 能量吸收器 / 戰鬥艙 / 相位匿蹤 / 測距瞄準器五項上線,見 `internal/shell/cloak.go`、
`energy_absorber.go`、`special_device_map.go`。**這一格清空了** —— 做完一項就回來刪一行,
不留「已完成」的殘影(那正是這份清單過去八次被誤讀的原因)。

同一輪把**原版特殊裝置表**(39 筆 × 空間/成本 × 6 艦級)抽出來,所以:

- 特殊系統的佔格不再是「艦體空間 5%」的估計,成本也不再是 remake 值 —— 兩者都依艦級走真值。
- 巨型通量器的「可用空間 ×125/100」接上了(手冊給比例,執行檔給截斷方式)。
- 抓到**兩條戰鬥路徑的光束分支都沒有填 `Target.HardShield`** —— 解算函式有那個參數,
  而呼叫端漏填拿到零值,不編譯失敗也沒有測試會紅。

### 附錄二、探針證據摘要（非待辦）

| 探針 | 殘量 |
|---|---|
| ① 零覆蓋函式 / ② 零消費常數 | 這組探針只能找 remake 死碼，不能衡量原版 call graph 覆蓋；2026-08-12 起不再用「見底」宣稱 parity |
| ③ 被餵固定值的參數 | **兩側都掃過了**(第 71 項(探針③內部函式))。gamedata 側剩 5 小項;內部函式側剩 4 條**已判定為正確**的 |
| ④ 手冊有而沒抄的資料 | 見上面第一節 |
| **新增:元件表有但沒有消費端** | 隱形裝置已接(第 77 項(元件表真值));剩 1 項已標註的代理(重生程序) |
| **新增:解算函式有參數但呼叫端沒填** | 硬化護盾曾在快速／格子光束呼叫端漏填，已由第 77 項修正；`TestHardShieldReachesBeamCallSites` 走真正 `battleShot` 入口鎖住兩條路徑。目前沒有把已關閉的漏填項重列成待辦。 |

### 附錄三、卡在證據不足的歷史紀錄（非本輪 remake 待辦）

| 項目 | 為什麼停 |
|---|---|
| `Calc_Tech_Value_` 階段 C–K | 三張資料表已解、低風險那半已接;**剩下的擋門是「候選各自代表什麼」**,常數表不能照抄 |
| ~~戰機基地 10 回合整補~~ **已結案,不做** | `Fighter_Garrison_Strength_` 確實從帝國科技與武器表當場算，沒有逐殖民地中隊存量；但 2026-08-27 已訂正：舊 remake 的「10／6／4 × 泛用戰機火力」並不等於原版戰略強度，現已依 `0x5F64C` 精確公式替換。|

### 附錄四、歷史功能盤點（非待辦；目前狀態以上方盤點結論為準）

| 項目 | 現況 |
|---|---|
| **英文模式的引擎層** | 名稱池、開局艦隊名／支援艦艦級、星系一次性發現、隨機事件、AI 突襲、安塔蘭警報、持續事件進度／結束與怪獸入侵報告已接回英文模板；2026-08-09 再補殖民地行星環境／特殊物產、建築／Special action、歷史圖表指標／AI 圖例、NEW GAME 五個值列、殖民地總覽／行星列表摘要、外交對手／選項、戰術敵艦／戰機與熱座席位名稱。2026-08-10 再加未知值安全 fallback，英文畫廊 35/35 張抽查通過。仍保留的中文字面值多為查表 key／dev-only／不可達分支；`internal/` 的查表 key(`special_device_map`、`weapon_damage`、`shipspace`…)不可全域翻譯 |
| ~~**資產分版(1.31 vs 1.5)**~~ **已完成(2026-08-09)** | `cmd/moo2` 支援 `-data13` / `-data15` 各自的多層 LBX 搜尋路徑，主選單切換會重建對應 `assets.Resolver`，`-data` 仍是共用回退；`auto` 讀資料目錄 README 的版本標記。以私有 1.31 基礎資料加 `MOO2-1.50.26.zip` 的 `patch/150/lbx` 實際跑兩版畫廊各 35 張；另修正 1.5 `NEWGAME.LBX` 滿版背景由 #28 順延至 #31 |
| ~~打包路徑~~ **已完成**(2026-08-08 第 83 項(打包路徑)) | 譯表烘進執行檔(`go:embed`),`-i18n <dir>` 為開發覆寫。**從任意目錄跑已實測通過**——過程中被實測打回來兩次(還有七處 `os.Open` 單一 .tsv、以及一處字面改對了卻還走 `os.Open`),兩道防呆測試就是那兩個形狀 |
| ~~淘汰 `-play` 簡約殼~~ **已刪除**(2026-08-08 第 81 項(淘汰簡約殼)) | 448 行 + 兩個旗標 + main.go 的分支全部移除。⚠ 順帶訂正兩件事:`colonyview.go` **不是它的東西**(走 `-colony`,是與 `-lbx`/`-race` 同類的開發用檢視模式,留著);`transition`/`screen` 兩個型別雖然住在 `play.go`,卻是十來個檔案在用的,已拆到 `screen.go` |
| **字型子集 + `go:embed`** | ⚠ **與上一列不是同一件事**(先前寫成同一件)。譯表已烘進去(第 83 項(打包路徑));**字型不能烘**——使用者自備的 TTC 授權不明,不能進 repo 也不能進執行檔(CLAUDE.md `[HARD]`)。真要做只能是「自製/開源字型的子集」,那是另一個題目 |
| 種族關係的 SABOTAGE / HIDE 兩顆鈕 | HIDE 與 SABOTAGE 已接入逐對手任務、存檔與回合結算；SABOTAGE 依原版 `0x1014A4`／`0x10130A`／`0x145EA` 做 70 門檻的建築破壞，remake 的 AB／DB／E 與 Agent 消費已完成；原版三顆鈕左右語意與 raw 完整分數仍未知 |
| ~~**客製種族特殊能力與數值 picks**~~ **remake 已接(2026-08-10)** | `CustomRaceTraits` 已接入官方 22 項選項、存檔／熱座與消費端；心靈感應、全知、匿蹤艦已接到外交、能見度／偵測與戰鬥；Lucky 的壞事件取消與額外好事件累積鏈、一般事件排程及帝國目標純權重已於 2026-08-25 依 IDA 修正。AI／熱座目標效果回寫仍依頂端活表列為未閉合。艦艇防禦、地面戰、諜報三類數值 picks 也已寫入 `Race`。完整逐能力 parity 仍以頂端活表為準。 |
| ~~熱座:指定**哪幾個**帝國是真人~~ **已完成**(2026-08-09) | 新遊戲命名/旗色後加入指定清單，按選取的 `AIPlayers` 索引接管；未選中的 AI 保留，並以測試鎖定選取順序與 UI 互動 |
| ~~熱座:席位補齊玩家側系統~~ **已完成核心轉換**(2026-08-09) | 接管席位保留種族加成、領袖、母星建築、艦隊、殖民地平行陣列與玩家間諜欄位；AI 模型本身沒有的建造佇列／前哨站／傭兵池仍以空值起步，列為誠實模型差異 |
| ~~CMBTSFX 爆炸/光束特效~~ **已接(2026-08-10)** | `CMBTSFX.LBX` 多幀 delta 累積解碼、命中／爆炸序列與安全 fallback 已接戰術畫面；原版逐資產語意仍未知 |
| 繪字描邊/陰影 | 逐字斷行已完成(`uifont/wrap.go`),描邊/陰影仍無;條目自己已把它降為次要 |
| ~~hover highlight 與原版一致~~ **remake 已接(2026-08-10)** | 外交、戰術控制列、戰機出擊與可點擊按鈕已有共用金色外框；只剩人眼觀感微調 |
| ~~音量控制(Music / Sound Fx)~~ **已完成**(2026-08-09) | 遊戲選單使用 `GAME.LBX` 資產 7 的 155×12 音量條；音訊 `Mixer` 提供讀寫/夾限，支援單擊與按住拖曳，中文模式覆蓋背景英文標籤。手冊「靠左關閉」也由 0.0 音量實作；`internal/audio` 與 `cmd/moo2` 各有幾何/範圍測試 |
| ~~多人連線畫面的烘字外洩~~ **已完成**(2026-08-09) | `netinfo.go` 依七個 `MULTIGM.LBX` 狀態資產覆蓋英文標題；四張帶狀態欄的面板也覆蓋 `STATUS`，英文模式保留原圖。既有狀態/尺寸測試外加標題與欄位對照測試 |
| ~~戰術「等待 / 完成」兩顆鈕~~ **已完成**(2026-08-09) | `tacticalScreen` 已有逐艦行動佇列：每艦每回合只行動一次，WAIT 延後到未行動艦之後，DONE 結束目前艦，最後一艦才結算戰機與敵方回擊；手動開火只作用於選中艦。測試：`tacticalturn_test.go` |

**歷史開放問題（不是目前待辦，別當成缺口引用）**

| 問題 | 已知的部分 |
|---|---|
| 殖民地總覽下排第三格,原版在那裡畫什麼? | `colsum.lbx#0` 那一區是「稀疏星點 + 透明」,疊在黑底上就是星空。remake 沒有往那格畫任何東西。**目前沒有證據**顯示原版會在那裡貼縮圖或圖表——要下結論得先去反組譯那個畫面的繪製函式 |
| ~~把反組譯出來的音樂對照表接進 remake~~ **已完成**(2026-08-08 第 78 項(音樂接線) + 第 82 項(音樂兩個缺)，2026-08-11 IDA 靜態校正) | 單一編號空間、兩個 LBX 都載、主選單/星圖每次重擲;科學室播完接隨機 STREAM 1..3;外交好關係曲是逐族記錄欄位 + 1，壞關係曲為 helper(3)+13；remake 使用逐族好曲與壞曲池。原版好／壞切換門檻仍未知，不能再寫成單一外交常數。 |

### 附錄五、網路多人的舊傳輸留白（刻意不做；TCP 多人見上方待辦）

`Modem_Setup` / `NullModem_Setup` / `Comm Info` 與 TEN **不恢復**——數據機、序列線、IPX
與 TEN 服務已不存在；替不存在的硬體做設定畫面不是還原。這個決策不等於 TCP 多人對局已完成，
正式對局接線仍以本頁上方「多人網路對局調整」為準。

### 附錄六、CLAUDE.md 列的交付項（歷史核對：全部在）

`PLAN.md` ✅ / 致謝(README)✅ / `docs/culture/moo2-chinese-cultural-phenomenon.md` ✅ /
`docs/history/moo2-chinese-community.md` + `moo2-history-and-reception.md` ✅ /
`docs/tech/sprite-tile-quality.md` ✅ / `docs/tech/ui-adjustment.md` ✅ /
主選單語言切換 ✅(2026-08-07)。**這一節查過,沒有被追丟的交付項。**

---

### ★ 這一輪的已完成日誌(A–AO,2026-08-08)

> ⚠ 下表**不是待辦**。它記的是這一整輪逐項做掉了什麼、以及每一項是怎麼被找出來的
> ——後者比前者有用,所以留著。

| # | 項目 | 性質 | 誰做 |
|---|---|---|---|
| A | ~~**叛亂系統**(第 46 項)~~ **已完成** — 機率規則從 `Check_Rebellion_` 抄出(每單位 1%、難度、AMC 減半、滅絕加倍)、每回合檢定、地面戰、殖民地還政舊主全部接上;順帶把 `GroundTypeFourth` 定名為叛軍 | 跨模組狀態機 | 主迴圈 ✅ |
| B | **`Calc_Tech_Value_` 抄寫** — 規格 + 三張資料表已解(`docs/re/calc-tech-value.md`、`calc-tech-value-tables.md`),**低風險那半已接**(category 倍率 + 科技應用粒度);高風險的階段 C–K 仍擱置。⚠ **擋門換了**:category enum 語意 2026-08-08 第 52 項已從成員反推出來(41 個乾淨的功能類別),**2026-08-08 第 54 項再降一級**:寫入端都找到了(`[0x89F]` = 政體,四項政府科技各寫一個立即數;`[0x28]`/`[0x205]`/`[0x206]` 是 `Init_NPC_Personalities_Objectives_Themes_` 的三次加權抽選,6/4/7 個候選,難度改權重)。**剩下的擋門是「候選各自代表什麼」**,階段 C/D/E 的常數表仍不能照抄 | RE | sonnet ✅ + 主迴圈接線 |
| C | ~~**文件過期斷言掃描**~~ **已完成** — 約 90 條斷言查出 16 個不符並訂正；歷史報告已由 Git 保存，不再保留錯誤句子的副本 | 機械核實 | sonnet ✅ |
| D | ~~**spy / leader UI**~~ **remake 核心已接** — 領袖畫面座標、捲動鈕、艦艇／殖民地 `LEADERS` 分頁、指派／改派／解除與保存已接；間諜區塊接進**種族關係**畫面(原版沒有獨立間諜畫面),提供訓練、逐對手 STEAL/SABOTAGE/HIDE、餽贈／特殊貿易與外交條約入口。remake 的 SABOTAGE 分數／Agent 消費已接；留白只剩原版 raw 完整分數／特殊槽位、AI 門檻與防守策略 | 挖完 → 接線 | sonnet ✅ → 主迴圈 ✅ |

> **D 項的規劃被一個負面發現改寫了。** 原本的假設是「做一張間諜畫面」——
> 而原版**根本沒有獨立的間諜畫面**(搜過 `Spy_Screen`/`Espionage_Screen` 零命中)。
> 間諜任務指派內嵌在「Races 種族關係」畫面(`Race_Screen_` @ 0x10ACBA)裡。
> 照舊假設去做,方向從一開始就錯。
>
> 領袖畫面(`Add_Officer_Screen_Fields_` @ 0x9264E)挖到 9 個定點按鈕 + 4×2 清單熱區;
> HIRE 鈕 (313,441) 已由主迴圈獨立核過立即數。**留白**:按鈕寬高未查(控制碼指向 LBX 資產)、
> 三顆間諜任務鈕的左右順序未確認、種族頭像/文字座標表未解。
>
> 順帶修正一個我自己派工時寫錯的前提:簡報說原版是 320×200,
> agent 拿三張畫面的立即數(全落在 0..639/0..479)推翻了它——而 remake 本來就是 640×480
> (`interactive.go:39`)。**它沒有照著錯的做**,這正是派工要留把關、也要留反駁空間的理由。
| E | ~~**AI 艦隊移動模型**~~ **已完成**(第 47 項)— AI 有位置與航線,突襲改成「抵達才打」;阿提米絲水雷因此打得到 AI。仍缺:艦員經驗(AI 無逐艦資料)、航線不判星雲/黑洞 | 架構級 | 主迴圈 ✅ |
| F | **戰機基地 10 回合整補** — **2026-08-08 複查:擋門理由仍成立**(見下) | 證據不足,非工時不足 | 擱置 |
| G | **屏障護盾擋生物武器 / 護盾被炸後氣候回退**:前者 2026-08-08 第 52 項**已接**(生物武器名單來自執行檔 category 表的 category 20);後者仍是刻意不做(建築效果一律不可逆) | RE + 手冊 p.99 | 主迴圈 ✅ / 後者不做 |
| H | ~~**安塔蘭母星艦隊強度**~~ **已完成**(第 49 項)— `{0,0,3,2,7,0,…}` 逐位元組解出,3 Large + 2 Huge + 7 Titan + 1 要塞,已換掉「6 艘同級」的保守預設 | RE | sonnet ✅ |
| I | ~~**turn-1 數值 playtest 校準**~~ **已結案** — 複查發現程式碼早就是 農4/工2/科2(與釘死的 oracle 一致),「科3 待定案」那句在 2026-08-06 就過期了 | — | 複查 ✅ |
| J | **聊天列真的走過線**:資料模型(chat.go)與畫面(netnextturn.go)先前都做完了,但**沒有任何 goroutine 在讀連線**——補 `netplay.Session` 訊息幫浦(2026-08-08 第 53 項) | RE + 移植決策 | 主迴圈 ✅ |
| K | **三個「沒查到寫入端」的欄位查到了**(2026-08-08 第 54 項):`[player+0x89F]` = 政體(順帶驗到 remake 的政體編號與原版逐項相同,已釘測試);`[0x28]`/`[0x205]`/`[0x206]` = 三次加權抽選;順帶訂正 GNN 聊天封包號碼記錯 | RE | 主迴圈 ✅ |
| L | **AI 對手自己會研究**(2026-08-08 第 55 項):先前 `ResearchTopic` 200 回合沒換過,科技線完全靠間諜偷玩家的;多選主題的抉擇永遠掛著 = 三個選項全拿。兩個洞都補上,應用項的抉擇用原版 category 倍率 | 探針 + RE | 主迴圈 ✅ |
| M | **領袖永久免費**(2026-08-08 第 56 項):`LeaderMaintenanceCost` 移植好、有測試、零呼叫端。用「一局 300 回合的覆蓋率」找出來的——名字層級的死碼掃描抓不到這類洞 | 覆蓋率探針 + 手冊 | 主迴圈 ✅ |
| N | **征服人口免費全額生產**(2026-08-08 第 57 項):手冊「each alien unit produces only three quarters」沒接,`ProdAlienWorkerOutput` 與 `ProdWorkerOutput` 都是零呼叫端。三項產出都套上,順帶讓「每工人至少 1 產能」的下限第一次真的生效 | 覆蓋率探針 + 手冊 | 主迴圈 ✅ |
| O | **間諜三項手冊加成一支都沒接**(2026-08-08 第 58 項):`spy.go` 的擋門註解列了三個理由,逐條核對後**有一條已經過期**(逐科技模型早就有了)。科技加成接上攻守兩側、政府加成接上玩家側;種族特性強度與 AI 政府欄位仍缺,理由重寫 | 覆蓋率探針 + 手冊 | 主迴圈 ✅ |
| P | **「成就」科技的全帝國效果四條都沒接**(2026-08-08 第 59 項):擋門理由是「remake 無成就追蹤系統」——而成就在 MOO2 就是科技。順帶修掉士氣傳基本政體而非進階政體的同款 bug | 覆蓋率探針 + 手冊 | 主迴圈 ✅ |
| Q | **艦員經驗四條玩家可見加成軌已接**：BA／BD、飛彈閃避與 Bo 均有消費端；2026-08-25 再以 IDA 閉合 Bo、艦隊 Commando／Security 與雙方所屬帝國資料，快速與格子戰術各有測試 | 覆蓋率探針 + 手冊 + IDA | 主迴圈 ✅ |
| R | **兩個為 UI 而寫的函式畫面沒用過**(2026-08-08 第 61 項):`AssimilationProgressNeeded` 與 `CrewXPToNextLevel` 的檔頭都寫著「為了讓玩家看得到」,而畫面從來沒呼叫過。同時完成 36 支零呼叫端函式的逐支分類——其餘都有正當理由 | 覆蓋率探針 | 主迴圈 ✅ |
| S | **護盾減傷五級裡四級是錯的**(2026-08-08 第 62 項):`shieldReduceByName` 回「清單索引 × 2」,手冊值 1/3/5/7/10 被算成 2/4/6/8/10。改用科技當鍵查手冊常數。掃法從「零呼叫端的函式」擴到「零消費的匯出常數」(190 個,其餘多為原版列舉鏡像)| 常數消費掃描 + 手冊 | 主迴圈 ✅ |
| T | **陸戰隊運力不分艦體一律每艘 4**(2026-08-08 第 63 項):手冊 p.121 每個艦體等級都有 Marines 欄(5/8/12/20/30/50),而 remake 一直有艦體等級。同族的 `armorHPByName`(缺裝甲科技倍率)與 `shipStrength`(明說是抽象)查過後不動 | 手冊 p.121(已交叉驗證) | 主迴圈 ✅ |
| U | **武器傷害全是單調估計**(2026-08-08 第 64 項):待辦寫著「需 OCR / 找完整手冊」——而手冊表就在同一份可抽文字的 PDF p.124-127,同一頁的 Size 欄第 16 項就抽過了。9 項換成真值;最嚴重的是排序被弄反(核融合光束原本比中子爆破槍強,手冊上更弱)| 手冊 p.124-127 | 主迴圈 ✅ |
| V | **手冊 18 把武器 remake 只做了 10 把**(2026-08-08 第 64 項):補上離子脈衝砲/引力波束/干擾者/重錘裝置/粒子束/脈衝飛彈/氙素飛彈/質子魚雷八項,研究主題取自執行檔而非猜測,飛彈分類由手冊+執行檔兩個來源背書 | 手冊 p.124-125 + 執行檔主題表 | 主迴圈 ✅ |
| W | **反飛彈火箭沒有元件載體**(2026-08-08 第 64 項):攔截公式(射程/命中率表/解算分支)早就抄完,`hasAMR` 卻恆傳 false 因為沒有船裝得上。補上元件,主題與分類皆取自執行檔。留白:炸彈那一批要先有「只能打行星」的武器種類,是新機制不是新資料 | 手冊 p.127 + 執行檔 | 主迴圈 ✅ |
| X | **炸彈需要「打不到船」的武器種類**(2026-08-08 第 64 項):第 64 項擋下的那個機制。手冊「only useful against planetary targets」——不是低傷害光束,是艦隊戰裡完全沒有這一發(連骰子都不擲)。第 64 項的完整性閘門這輪抓到兩個猜錯的研究主題 | 手冊 p.126 + 執行檔 | 主迴圈 ✅ |
| Y | **球形武器分支零掛載**(2026-08-08 第 64 項):`ResolveSphericalShot` 與 `battleVolley` 的球形 case 從來沒執行過。補上脈衝星/空間壓縮器,順帶給 combatant 補上艦體等級(「per size class of target」要用)。只有壓縮器豁免護盾裝甲——逐武器不是整類 | 手冊 p.126-127 + 執行檔 | 主迴圈 ✅ |
| Z | **p.127 特殊武器盤點 + 活來源表補一列**(2026-08-08 第 64 項):剩下六項全部卡在機制(戰鬥速度/行動禁止/護盾分面)不是卡在資料——硬加會做出「名字對、行為錯」的武器。同時修正活來源表的形狀:它只裝得下子系統級的洞,而第 52–64 項全是「做了但不忠實」 | 盤點 + 文件 | 主迴圈 ✅ |
| AA | **種族特性表:7 個自編數字 → 原版 31 格**(2026-08-08 第 65 項):RACESTUF.LBX asset 7 + 執行檔換算表 + SAVE10 三方核對,挖出全 13 族一手特性表。八族數字是錯的(克拉肯拿錯欄位、阿爾卡里的「戰鬥+15」實為防禦+50、薩克拉成長 30 實為 100…);攻/防拆成兩個獨立特性;順手修掉諾蘭姆低重力扣兩次的 bug;解掉「13 族沒有人具備地底/高重力」這個可證為假的擋門理由 | 逆向 + 忠實化 | 主迴圈 ✅ |
| AB | **布林特性接線**(2026-08-08 第 65 項，2026-08-09 客製種族補線):統帥/惹人厭/寬容/神級商人/魅力五條內建規則改由 31 格特性查詢；客製種族選項另以 `CustomRaceTraits` 保存，低／高重力、穴居、戰爭領主、跨維度與上述四種既有公式會實際生效。未建模的客製深層能力只保存、不宣稱完成。⚠ 中途走錯一次:把內建衍生狀態存進存檔,三個測試同時紅;正解是內建族仍由 `RaceIndex` 推導,只有客製選項本身需要保存 | 忠實化 | 主迴圈 ✅ |
| AC | **高能聚焦裝不上**(2026-08-08 第 66 項):`DamageMountBonusHEF=50` 與公式都寫好了,但 HEF 在手冊裡是**艦載系統**不是武器改造,而 `SpecialOptions` 沒有它——玩家裝不上,於是 hefBonus 恆傳 0。補進元件清單並接進快速結算與格子戰術兩條路;手冊那三句話(加傷害/不加命中/不抵銷衰減)逐句釘住,命中那條用逐骰比對 | 忠實化 | 主迴圈 ✅ |
| AD | **裝甲倍率撤回 + 重裝甲**(2026-08-08 第 67 項):先前斷言「裝甲科技倍率手冊與 openorion2 都沒有」是錯的——手冊 Ship 條目逐級寫著 +100%/+300%/+500%/+700%/10 倍,只是沒讀到那幾頁。裝甲值改用手冊階梯(氙素 120→100,它是 10 倍不是 +1100%);氙素裝甲掛錯主題一併訂正;補上重裝甲系統,`apNegated` 那個恆傳 false 的參數終於有生產端 | 忠實化 + 撤回 | 主迴圈 ✅ |
| AE | **手冊元件完整性盤點 + 飛彈防禦家族**(2026-08-08 第 68 項):撞了兩次「手冊的 System 沒進元件清單」之後改問法,把手冊 88 個元件條目自動對四張表比一次——**47 個裝不上**。分桶記錄(15 個由別的模型承接 / 8 個這輪接上 / 4 個卡機制 / 20 個仍缺可做),讓剩下的那一桶可見。接上的 8 個是飛彈防禦家族(干擾器×3、慣性×2、閃電場、位移裝置、部隊艙),手冊數字全在、生產端全是 0 | 盤點 + 忠實化 | 主迴圈 ✅ |
| AF | **那一桶的第一批 + 修第 68 項的缺失**(2026-08-08 第 68 項):慣性穩定器手冊給三個效果,第 68 項只接了一個——**從單一檔案(missile.go)回推元件效果會漏東西**。接上戰鬥掃描器(+50 命中)、強化船體(結構 ×3)、多相護盾(吸收 +50%)、偵察實驗室(研究 1/2/4/8/16/32)。沒接的十幾個逐項寫明理由,其中結構分析儀/阿基里斯要先重構傷害鏈——那不該夾在資料項裡做 | 忠實化 | 主迴圈 ✅ |
| AG | **傷害鏈收成具名結構**(2026-08-08 第 68 項):`ResolveShotWithMods` 的位置參數排到第 11 個,`false, nil, 0, false` 這種尾巴沒人看得出哪個是什麼。收成 `BeamShot{Attacker, Target}`,舊入口保留為薄包裝(第一條測試就驗兩者逐欄位相同)。卡在後面的結構分析儀(過盾後傷害加倍)與阿基里斯瞄準器(光束無視裝甲)一併接上 | 重構 + 忠實化 | 主迴圈 ✅ |
| AH | **戰鬥速度表 + 引擎階模型**(2026-08-08 第 69 項):從執行檔挖出 6 階 × 6 艦體的戰鬥速度表,三個規律自我驗證(每階 +2 對上手冊註腳、max=min+10 共 36 組成立、大船遞減幅度自己也遞減)——所以只存第一列,其餘用公式算。順帶抓到 `NewFighterSquadron(..., **1**, 0)` 這個硬編值:它在 cmd/ 裡,第 65 項的掃描器看不到——**掃描器的盲區也要記帳**。接上主動權排序(手冊公式,先前是「先造的先打」)、增強引擎、戰機真引擎階 | 逆向 + 忠實化 | 主迴圈 ✅ |
| AI | **戰術移動不再瞬移**(2026-08-08 第 69 項):先前點任何空格都能到。原版棋盤大小是一手事實——`Assign_Combat_Grids_` 的迴圈界限說是 **81×68**(列距 0x88=136B=68 格×2,與內圈上界互相驗證),所以比例尺 1:10 是推出來的不是挑出來的。速度 13..30 → 盤面 1..3 格。移動用曼哈頓距離與射程判定同度量,否則「走得到卻打不到」很難解釋。⚠ 移動力重置必須放在戰損壓縮之後,否則索引錯位 | 逆向 + 忠實化 | 主迴圈 ✅ |
| AJ | **牽引光束 + 停滯力場**(2026-08-08 第 69 項):第 64 項判定它們「卡在機制」,第 136(速度)+137(移動預算)+ 這一項(逐艦狀態)三塊補齊後解掉。手冊括號裡的例子「6 beams would immobilize a Doom Star」把公式釘死了——末日之星第 5 級、6=5+1,另一端巡防艦 1 束,兩端都對上中間就是線性。停滯力場「or be affected by any weapon」那句最容易漏:只做「不能動」會讓它變活靶。每回合重算不累積(產生源被打掉效果就該消失) | 忠實化 | 主迴圈 ✅ |
| AK | **陀螺去穩器**(2026-08-08 第 70 項):第 64 項的擋門理由「光束路徑沒有 per size class 乘數」每個字都對,但它預設了「它是光束」這個不必接受的前提——手冊那兩個特徵(依級數乘、豁免盾甲)都是球形家族的,歸過去一行乘數都不必加。佔格 75 的欄位對齊靠「同一列的脈衝星 = 50 對上第 64 項已記值」確認 | 忠實化 | 主迴圈 ✅ |
| AL | **一回合開幾次火**(2026-08-08 第 70 項):超載電容/快速飛彈架/時間扭曲加速器卡在同一個缺失機制,建一次解三個。冷卻的讀法是關鍵——手冊的 **unused** 是「完全沒開火」不是「沒有連射」,所以連射→單射→連射的無限循環不成立。時間扭曲加速器沒有冷卻(手冊沒那句)。`battleVolley` 的 60 行迴圈體抽成 `battleShot`,沒有這些系統的船 shots==1,RNG 消耗逐位元不變 | 忠實化 + 重構 | 主迴圈 ✅ |
| AM | **轟炸機庫**(2026-08-08 第 70 項):`FighterKind` 的檔頭寫著「手冊 p.127 的**四種**」而底下只有兩個。gamedata 那側的速度與射擊次數**四型都齊、血量只有兩型**——資料層跟著實作層缺,單看 gamedata 會以為本來就只有兩個。轟炸機的三個數字都在正文裡(速度 8/血量 4/出手 1),不必碰那張欄位打散的表。⚠ 轟炸機的炸彈算得進艦隊戰而艦載炸彈算不進去——手冊自己分開寫的,載具不同規則不同 | 忠實化 | 主迴圈 ✅ |
| AN | **探針 ③ 的另外半邊**(2026-08-08 第 71 項):第 65 項的掃描器只看跨套件呼叫,shell/cmd 內部函式從沒掃過。改用 go/ast 掃出 17 條線索,四條是真缺口,同一個主題——**元件/科技的第二個效果沒接**:硬化護盾只在光束路徑生效(手冊 each enemy attack)、攻方掃描器對飛彈干擾的抵銷恆傳 0(迅子 −20/中子 −40)、戰鬥掃描器的 +2 parsec 掃描範圍沒接、三個掃描科技的偵測距離是自編值**而且順序反了**(手冊 空間 2/迅子 4/中子 6,舊版把迅子當最高階 → 研究出迅子反而看得更近)。⚠ 檔頭那句「手冊未公開逐科技數字」**不是過期,是從寫下那天就錯的**。順手修一個真 bug:`VisibleStars` 只把**選中的那一支**艦隊當偵測源,切換選中會改變戰爭迷霧 | 忠實化 | 主迴圈 ✅ |
| AO | **盤點:「元件表有」不等於「效果有接」**(2026-08-08 第 72 項):活表寫「④ 剩 6 個元件」——逐項查完是 3 真擋 + 3 個理由過期/本來就錯 + 2 個根本不在名單上。漏得最徹底的是**隱形裝置**:在 `SpecialOptions` 裡、花得了錢、**整份程式碼沒有任何地方讀它**,而手冊規則是完整的。先前每次盤點問的都是「元件表裡有沒有這一項」,隱形裝置**有**,所以每次都被算成已完成。新問法:掃元件名在表以外出現幾次,零次就是裝飾品。另:`ship_systems.go` 檔尾那份「沒接的與理由」12 項裡有 7 項早就接掉了——**清單自己就是它警告的那種東西**,已整份重寫 | 盤點 | 主迴圈 ✅ |
| AP | **三份文件對齊現況**(2026-08-08 第 72 項):`CONTEXT.md` 先前不存在(rulebook/50 要求),收的詞以**這個專案實際踩過的混淆**為準——快速結算/格子戰術兩條路徑、元件表有≠效果有接、擋門理由會過期、三個驗收工具各自**不能證明什麼**;末尾四個待釐清詞不擅自收斂。WORKLIST 頂端「剩餘工作清冊」底下 40 列有 38 列 ✅——早就是已完成日誌了,重寫成六節剩餘工作表。順帶查核 CLAUDE.md 自己列的七項交付全部都在 | 文件 | 主迴圈 ✅ |
| AQ | **第二次 review:錯誤斷言的副本比原件多**(2026-08-08 第 72 項):第 72 項只動了 WORKLIST,而同一批假斷言在 README / 完成度評估 / oracle 對照 / PLAN.md / ebiten-notes / component-values / newgame-flow / gap-report 裡**都有副本**,讀的人多半只讀副本。`oracle-comparison` 最值得記——**同一份文件自己內部就矛盾**,結論段寫對了、表格沒跟著改。長期記憶有兩份是錯的:一份三句話後來全被推翻,一份**檔名意思與內容相反**(已更名)。`CLAUDE.md` 的過期錨點從兩個變三個,新增最貴的那句:**「已經沒有自驅剩項了」——它讓下一輪不去找** | 文件 + 記憶 | 主迴圈 ✅ |
| AR | **第三次 review:改用掃描器**(2026-08-08 第 72 項):前兩輪都是先想到「哪裡可能有問題」再查,上限是**想不到的就查不到**。改成兩段式掃描——抓同時含否定性斷言關鍵詞與程式碼識別字的 markdown 行(434 條),再數每個識別字在非測試 Go 檔的引用數,「說沒接卻被引用 ≥3 次」= 可疑(103 條)。抓到九條假斷言,**全是前兩輪想不到的**(spherical 武器 / `ShowRelocationLines` UI / `AIOpponent.Leaders` / AI 同星系多殖民地 / 艦員等級 / `RaceTrait` 接線 / `Maintenance` / 間諜 UI / BC 死亡螺旋)。⚠ 也抓到三條**不該動的**(AI 無建築追蹤、無種族欄位、`ModeOriginal` 仍回 Remake)——掃描器只縮小範圍,判定仍要逐條看程式碼 | 文件 | 主迴圈 ✅ |
| AS | **音樂場景表:不需要人耳,執行檔全寫著**(2026-08-08 第 73 項):使用者指出有 IDA Pro 之後這題可解。上一輪判 `Play_Background_Music_`「全檔案零引用 → 死碼」——**位址少加了 `0x10000`(object base)**,真址 `0x2484F`,**15 個呼叫端**。上一輪為了證明零命中做了三種掃描,每一種都很用力,**而全部在錯的位址上找**;正對照本來會當場抓到。改用 `.i64` + 除錯符號表後,20 個呼叫端**每一個都直接對到具名畫面函式**。三個入口全解(≤100 走 STREAM / >100 走 STREAMHD;背景 = STREAM 1/2/3 隨機、戰鬥 = 4/5/6 隨機)。⚠ **主選單是每次隨機三選一,不是固定曲目**。順帶把戰機基地整補從「證據不足」升級成「已查證:原版也是當場算的」 | 逆向 | 主迴圈 ✅ |
| AT | **過期斷言直接刪除,不留刪節線**(2026-08-08 第 74 項):先前用「`~~錯誤~~` + 訂正註記」保留時序,實際效果是**同一句話出現兩次、錯的排在前面**。改成整句刪除只留正確結論——全 repo 清掉 42 條追認註記、十幾處刪節線引述,含 `audio-track-map.md` §7.5 整節錯誤推導。保留兩類:**完成標記**(不是錯誤結論)與**錯誤怎麼發生的教訓**(改寫成正面敘述)。`docs/re/01-gap-report.md` 是唯一保留錯誤推導全文的地方——它是工程日誌,那正是它的內容 | 文件 | 主迴圈 ✅ |
| AU | **活表併進 WORKLIST,gap-report 不再是現況**(2026-08-08 第 74 項):使用者問 gap-report 有沒有保留價值。評估:8,154 行裡有 **196 個位址、131 列表格**,是一次性成本的知識(重挖每項數小時起跳),被 18 個檔案引用——不刪。**但真問題存在**:第 72 項把剩餘工作搬進 WORKLIST 之後兩份活表並存了四輪。處置是拆職責——gap-report 開頭 85 行的活表刪掉換成指標,接手文件與五份 kickoff 的指標全改指這裡;gap-report 的職責寫死成 **RE 硬資料 + 工程日誌**,都不是現況 | 文件 | 主迴圈 ✅ |
| AV | **取消例外:gap-report 也不留錯誤紀錄**(2026-08-08 第 74 項):第 74/74 項把 gap-report 列為例外(「錯誤推導的全文只留在這裡」),使用者決定取消。刪掉 16 處**複述錯誤斷言**的段落——「`X.go` 原本寫著『……』,兩句都不對」這種形狀,只留正確結論。保留 RE 硬資料、完成標記、以及**一律改成正面敘述的教訓**(不寫「那句話是錯的」,寫規則本身)。**現在全 repo 一條規則:任何文件都不留錯誤斷言的內容**,要知道當初怎麼錯的看 git log | 文件 | 主迴圈 ✅ |
| AW | **編號項煉成 12 條規則**(2026-08-08 第 74 項):使用者「條目太多了」。編號項裡的教訓**高度重複**——「擋門理由會過期」出現十次、「兩條戰鬥路徑」出現在第 68–71 項每一項。合併成 12 條放在 gap-report 最前面(判斷還缺什麼 5 條 / 改動完整性 4 條 / 逆向 3 條 + 驗收),每條標出產生它的項次。**編號項降級成參考資料**,要看某條規則的證據再去翻。`CLAUDE.md` / `CONTEXT.md` 的指標同步改成「只有開頭那張規則該讀完」(**不複述條數**——那個數字會長) | 文件 | 主迴圈 ✅ |
| AX | **151 項壓縮:8,163 → 3,582 行**(2026-08-08 第 74 項):第 74 項只加了規則總表,檔案沒變小。這一項動 151 項本身——每項只留**標題 + 結論段 + 帶一手資料的行**(位址/表格/程式碼/識別字/粗體條目),其餘敘述刪除(那正是重複的部分)。**一手資料一筆沒少**:位址 196→197、表格列 778。編號沒動(41 處交叉引用),加八個分組標題。代價如實記錄:機械壓縮切斷了 8 處跨段引述,已清 | 文件 | 主迴圈 ✅ |
| AY | **編號壓縮 152 → 75**(2026-08-08 第 75 項):22 組主題合併(續篇併回父項、同一條線整組併成一項),子項內容保留成 `###` 小節。交叉引用 **802 處、106 個檔案**改寫(268 處在 Go 註解裡),做法是先建含「被併掉項次 → 父項新編號」的對照表再機械替換三種寫法,最後掃一次驗證**沒有引用指向不存在的項次** | 文件 | 主迴圈 ✅ |
| AZ | **交叉引用不必再開文件**(2026-08-08 第 76 項):`見第 68 項` 要開 3,800 行文件才知道是什麼,而程式碼註解裡有 268 處。**717 處裸引用加上短標**——`第 68 項`,一層就懂。多重/範圍引用維持原樣(加了太長),另加一張 75 項速查表。⚠ 順帶修掉合併腳本的 bug:舊編號 1/2/3 出現兩次導致輸出**逐行重複的區塊**、序列缺 2,靠「項數 76 但唯一值 75」的計數對不上才發現——**機械重寫要驗輸出本身的完整性,不只驗引用** | 文件 | 主迴圈 ✅ |
| BA | **複查抓到兩個真的改壞**(2026-08-08 第 76 項):①**跨文件的項次被套用了 gap-report 的對照表**——`rules-implementation-audit.md 第 10 項` 變成 `第 3 項`、`doc-audit 第 3/4 項` 變成 `第 2 項`,五處已還原。**被改成一個存在的錯項次,「有沒有失效引用」那個檢查抓不到**。②短標貼到同編號的另一個小節上(先前手寫的括號停在「分艦隊」這個小節名上,而該項正名是多艦隊模型),含 WORKLIST 49 列日誌。現在全 repo 同一項次只有一個短標 | 文件 | 主迴圈 ✅ |
| BB | **同義不同詞收斂**(2026-08-08):同一個項次在不同檔案被寫成不同短標(例:第 3 項有人寫「殖民地畫面」有人寫「Colony+Event 畫面」),讀的人要多查一次才確定是同一項。以速查表的正名為準,15 個項次、69 處改成一致;另把兩段敘述裡當例子引用的舊短標改寫掉,免得驗證腳本把它當成真的引用。驗證改為三問:字面對不上 0 / 同項次不同短標 無 / 引用短標與速查表不符 無 | 文件 | 主迴圈 ✅ |
| BC | **原版特殊裝置表 39 筆抽出來**(2026-08-08 第 77 項(元件表真值)):`Orion2.exe` 0x17EEE0 起、每筆 47 位元組,含**六個艦級各自的空間與成本**。欄位偏移由 `Special_Devices_Available_`(0x5F9EA)與 `Init_Internal_Space_`(0x36470)兩個讀取端定出。交叉驗證:39 筆科技編號逐筆對上既有 `TECH_*` 列舉零不符;戰鬥艙的負佔格 = 手冊 p.121 艦體空間的一半(連 12=floor(25/2) 的截斷都一致)。順帶挖到巨型通量器 ×125/100 | RE | 主迴圈 ✅ |
| BD | **佔格與成本改讀真值 + 五個元件接完**(2026-08-08):佔格先前是「艦體空間 5%」的估計、成本是 remake 值,兩者都換成原版表(依艦級)。戰鬥艙(負佔格)、隱形裝置(+80 光束防禦/飛彈 50% 未命中/停火一整回合重隱)、相位匿蹤(10 回合不可鎖定後降級)、測距瞄準器(射程 1/3 **只影響命中**)、能量吸收器(存 1/4、自動命中)全部上線 | 戰鬥 | 主迴圈 ✅ |
| BE | **抓到光束路徑漏填硬化護盾**(2026-08-08):`ResolveBeamShot` 從第 68 項(元件盤點+飛彈防禦)起就有這個參數,而**兩條路徑的光束分支都沒填**——而 `combatant.hardShield` 的註解寫著相反的話。既有測試全都直接呼叫解算函式並自己填,驗不到呼叫端。新測試走 `battleShot`,並用正對照確認拿掉修補會紅 | 戰鬥 | 主迴圈 ✅ |
| BF | **音樂場景表接進 remake**(2026-08-08 第 78 項(音樂接線)):第 73 項(音樂場景表)反組譯讀出來的對照表先前只寫在文件裡,程式碼還跑著時長啟發式的猜測值。改成原版的**單一編號空間**(≤100=STREAM、>100=STREAMHD),兩個音樂 LBX 都載;主選單/星圖改成 `Play_Background_Music_` 的 STREAM 1/2/3 **每次重擲**,戰術戰鬥走 STREAM 4/5/6;科學室/事件/議會/安塔蘭廳/艦艇設計/殖民地戰鬥各自接上真值。實測 `stream.lbx` 的 entryIDs=[1 2 3 4 5 6 8 10] **正好就是反組譯點名的那八個**,沒點過的 7/9 正是兩個空槽 | 音訊 | 主迴圈 ✅ |
| BG | **武器表(46 筆)與艦體表抽出來**(2026-08-08 第 79 項(武器表與艦體表)):同一區的另外兩張表。**對撞抓到 remake 抄錯一格**——質子魚雷的傷害 25 / 佔格 20 其實是 A-M 魚雷那一格,執行檔給 40 / 30(手冊 p.125 的表在 PDF 抽取後欄位打散,當初讀成「Proton/A-M Torpedo 25」)。順帶證實電漿砲 1.31 = 30(反組譯的是 1.31 exe)、解出武器類別與彈藥兩欄。**艦體表 +34 那一欄不下結論**(末日之星 400 比泰坦 4000 低一個數量級,解釋不通就不寫進程式碼) | RE | 主迴圈 ✅ |
| BH | **登艦戰建起來,三個擋著的元件全部接完**(2026-08-09 第 80 項(登艦戰)):手冊把解算方式直接指回地面戰(「fight it out in the same way as ground troops do」),而那套解算器早就在了。突擊艇(戰機家族,4 架/隊、每架 1 個陸戰隊單位、速度 6、血量 3,抵達即奪船)、保安站(守方 +20)、傳送器(12 格、面向攻擊方分面失效才可、硬化護盾阻擋)上線,部隊艙補上「可以登艦」那一半。匿蹤力場的擋門理由過期(AI 艦隊現在有座標了),重查後結論不變但換了理由(AI 從不讀玩家艦隊位置) | 戰鬥 | 主迴圈 ✅ |
| BI | **音樂兩個缺補完 + `-play` 殼刪除**(2026-08-08 第 81 項(淘汰簡約殼) + 第 82 項(音樂兩個缺)):科學室播完接隨機 STREAM 1..3(`PlayBGMOnce` + 每幀 `tickBGM`);外交「關係好」= 種族索引 + 1——**上一輪掛在待辦上的那張「逐族靜態表」根本不存在**,`sub_12983` 顯示 `[帝國紀錄+0x25]` 就是種族索引本身。`-play` 簡約殼 448 行整個刪除(先拆出 `screen`/`transition` 兩個型別,它們是十來個檔案在用的) | 音訊 + 清理 | 主迴圈 ✅ |
| BJ | **譯表烘進執行檔**(2026-08-08 第 83 項(打包路徑)):執行檔先前對當前工作目錄有隱性依賴,只有從 repo 根目錄跑才找得到譯表。**改完之後被實測打回來兩次**——第一次還有七處 `os.Open` 直接讀單一 .tsv(不走 `LoadFS`,grep 不到);第二次有一處字面改對了卻還走 `os.Open`。兩次都是「從 /tmp/elsewhere 實際跑一次」抓到的,不是讀程式碼。防呆測試因此有兩道 | 打包 | 主迴圈 ✅ |
| BK | **名稱池改存英文原文**(2026-08-08 第 84 項(名稱池雙語化)):829 條星名 + 672 條艦名。英文池是從中文池**反查**還原的(0 歧義 0 缺漏),中文由注入的翻譯器在**取名當下**翻——所以中文模式輸出逐位元不變(畫廊 34 張 0 張不同,含狀態指紋),英文模式的星圖真的顯示英文星名 | i18n | 主迴圈 ✅ |
| BL | **英文版面撞牆,一層層挖**(2026-08-08 第 85 項(元件名英文)):艦艇設計畫面六層破版——右緣截字 → 壓分隔線 → 豆腐框 `▸` → 元件名沒英文 → 改造標籤沒英文 → 補完英文之後標籤疊在一起。**四層是英文專屬,中文截圖一張都看不到**。順手把改造晶片的熱區與繪製收成同一個座標函式(先前兩份寫死表) | UI | 主迴圈 ✅ |
| BM | **hi-res 畫布 2×(1280×960)上線**(2026-08-08 第 86 項(hi-res 畫布)):**420 個繪製呼叫點一個都沒改**——文字改成「錄下來、最後以 2× 座標與字級重播」(`uifont/record.go`),美術走 nearest 整數倍放大,座標系仍是 640×480。CJK 從 10–13px 變 20–26px 且全部轉向量。代價是 z 序,用 `cmd/moo2/zorder.go` 的自動屏障解掉(所有面板填色/貼圖先沖文字,帶矩形相交判定)。`-uiscale 1` 逐位元回歸驗證 | UI | 主迴圈 ✅ |
| BN | **既有版面缺陷四處 + 截圖 alpha**(同上輪):星圖工具列/殖民地排序列/戰術控制列/艦艇設計的擦底板座標全部改用**英文模式畫廊**量的真值(用蓋著英文的中文截圖量英文位置,本來就量不準)。另修 `saveScreenshot` 存檔前壓不透明黑——先前 34 張基準圖裡那塊「白噪點」其實是 `alpha=0` 被檢視器疊白,玩家螢幕上是原版底圖的星空 | UI | 主迴圈 ✅ |
| BQ | **英文模式引擎層殘量收掉**(2026-08-08 第 88 項(開局艦隊英文名)):開局三艘船改存英文原文 + 支援艦艦級補英文。中文畫廊 34 張逐位元不變(含狀態指紋),英文畫廊真的變 `Pathfinder / Colony Ship` | i18n | 主迴圈 ✅ |
| BP | **戰術控制列七顆鈕接上**(2026-08-08 第 87 項(控制列熱區)):中文化早就做了、熱區一個都沒有——截圖只能證明「按鈕長得對」,證明不了「按鈕能按」。自動/掃描/登船/撤退接真功能(登艦解算放 `shell.ShipBoardingAttack`),等待/完成後續接上逐艦行動佇列；選項仍說明尚未完成。補了熱區、逐艦 WAIT/DONE 與正對照測試 | 戰鬥 | 主迴圈 ✅ |
| BO | **中文折行避頭尾**(同上輪):`uifont.applyKinsoku` —— 開括號不留行尾、收尾標點不在行首(推不下就不推,不撐破面板)。兩條規則都做了正對照 | i18n | 主迴圈 ✅ |

### F 項為什麼擱置(2026-08-08 複查,附證據)

手冊 p.86:「All ground-based squadrons of fighter craft are **totally renewed every 10 turns**.」

要「整補」得先有東西**會少**。remake 的戰機基地戰力是**當場算出來的**,不是存下來的狀態:

```
internal/shell/orbital_bombardment.go
    atk := fighterGarrisonStrengthFor(defender)
```

`fighterGarrisonStrengthFor` 從帝國科技、武器與裝甲當場算戰略強度——**沒有任何地方存著
「這個殖民地現在剩幾隊」**。
所以補一個 10 回合的計時器,它會把一個永遠等於滿額的值重設成滿額,是空轉。

要接得先有中隊耗損。而**手冊沒有描述任何耗損機制**——它只說會整補,沒說什麼時候會少。
憑「整補」反推「一定會耗損」再自訂一套損失規則,那是臆造。

**這一項是證據不足,不是工時不足。** 要往前推,下一步是去反組譯找戰機基地的中隊狀態
(有沒有 per-colony 的中隊計數欄位、誰會扣它),而不是先寫一個計時器。

### 這一輪的分工(使用者 2026-08-08 授權,主迴圈監工)

依 `rulebook/45`:判斷/架構/把關留主迴圈,RE 抄寫與機械核實派 sonnet。
派工邊界一律寫死:**不 commit / 不 push / 不碰 docker 共用資源 / 只動指定檔案**。

## ★ 2026-07-10 session 進展摘要(接手後,對原版/手冊驗證)
> 本段為快速索引,細節散見各 Phase 與 docs/tech/。

**已完成並截圖/資料驗證:**
- [x] **音訊基礎**:PCM WAV 原封播放 + 主選單 BGM + 按鈕音效(`internal/audio`、`audiohook.go`)
- [x] **研究系統完整忠實化**:真 RP 成本 + 每主題抉擇 UI(真中文名)+ 抉擇解鎖元件(對 openorion2 核實,`research-system-status.md`)
- [x] **獨立種族選擇流程**:13 族肖像 + 自訂點數 + 命名/旗色(`newgame-flow.md`)
- [x] **外交畫面破解重建**:DIPLOMAT.LBX 全破解(13 palette+13 房+13 使節,配對律 r)+ 逐族使節疊合 + 13 族對應核實(`diplomat-lbx-layout.md`)
- [x] **戰術戰鬥換原版美術**:STARBG 星空 + COMBAT 控制列 + CMBTSHP 可見艦艇;太空戰接真命中/傷害/過盾/過甲公式(`ResolveShot`);**控制列 7 按鈕中文化**(自動/掃描/登船/撤退/等待/完成/選項)(`tactical-combat-assets.md`)
- [x] **中文化稽核補漏**:galaxy 工具列 ZOOM→縮放、頂部 GAME→遊戲(擦底疊字)
- [x] **`-gamegallery` 端到端截圖廊**:8 畫面互動 app 內全繁中渲染驗證(修無限迴圈 CPU bug:硬性終止+timeout)

**四 directive 收官狀態(對手冊/攻略/一代/EXE 驗證,不再等使用者 oracle):**
- [x] **音樂曲目↔場景定案**(task 13):場景→曲目全部從執行檔的立即數讀出並接進 remake(第 73 項(音樂場景表)解出、第 78 項(音樂接線)接線)。**不需要人耳**——仍未定案的只有「這一條叫什麼名字」,那是命名不影響行為。
- [x] **地面戰係數**(task 14):RE 定案用一代 1oom `game_ground_kill`(d100+force)+ 二代加成表/hits-to-kill;`ResolveGroundBattle` 實作+確定性測試綠。剩 UI 入侵接線(歸 task 16)。
- [x] **真母星初始狀態**(task 15):Average 忠實開局實作(單一母星、Marine Barracks+Star Base、起始科技對 tech.cpp 驗證、建築數公式、1 Colony+2 Scout)+ 測試綠。
- [~] 傘狀項,活表仍有殘量,保留(數量以活表為準,不在這裡複述)

**task 16 分塊進度(2026-07-10,使用者授權自主排序):**
- [x] 殖民地建築 5→40 棟入 `gamedata/buildings.go` + 前置科技 gating(`colony-buildings.md`)
- [x] 行星→產出 yield 表(`planet_yield.go`,climate 食物/mineral 工業/gravity,手冊頁碼有據)
- [x] 維護費由建築算(`BuiltMaintenanceBC`,母星 3 BC,取代無據平坦 5)
- [x] 經濟可持續化(玩家+AI 對稱):饑荒復原 + 食物盈餘收入(手冊 p.25)+ 玩家/AI 母星行星驅動 yields;300 回合自我修復、測試更新到忠實基準
- [x] 修 AI 艦隊投資整數捨去 bug(餘數池,FleetStrength 正確成長)+ AI 接忠實 yield
- [x] 地面戰「模型 + 流程」shell 層接線(task 16 續):陸戰隊生成(Marine Barracks 依手冊公式補充,`advanceMarines`)、載運(`LoadMarines`,運力=艦數×手冊每艘 4 的近似,無獨立運輸艦船體類別,標簡化)、入侵解算(`GameSession.InvadeColony`,組 `gamedata.GroundForce` 接 `ResolveGroundBattle`,rng 依回合+星索引種子化可重現)、勝負後續(星 Owner 轉移 + 殖民地過戶/AI 端移除,`internal/shell/ground_invasion.go` + `ground_invasion_test.go`)。剩 UI 繪製/操作介面未做(不碰 interactive.go,歸後續 task)。
- [x] 地面戰補完:裝甲營房戰車 + 軌道轟炸接線(task 16 續,2026-07-11；2026-08-24 勘誤):`gamedata/ground.go` 的裝甲營房、戰車與入侵鏈已接進活對局。**軌道轟炸**現依 IDA `sub_4257E @ 0x4257E` 採固定三外圈；1.3/1.5 的 5/10 只代表炸彈攻擊當量，累積傷害超過 30000 即停，runtime 結果由呼叫鏈證實除以 40。手冊的 10 次與除以 100 僅保留為顯示估算證據，不再描述 runtime。仍未精確還原的快速戰鬥武器數量與特殊武器鏈已留在頂端活表。
- [x] 武器改造佔格已於 task#36 接線(`ShipDesignSpaceUsedWithMods`),此條的待辦部分已完成
- [x] 戰鬥公式依武器類型分流(**2026-07-11**):飛彈躲避/AMR 攔截/球狀傷害的公式其實先前就已移植自手冊(`gamedata/missile.go`/`gamedata/damage.go`,有測試),只是戰鬥解算(`cmd/moo2/interactive.go` `fireRound`、`internal/shell/session.go` `battleVolley`)全部武器都走 beam 邏輯(`shell.ResolveShot`),飛彈(核飛彈/麥克萊特飛彈)被當 beam 打。這輪修正:新增 `internal/shell/weapon_kind.go` 依武器名分類 beam/missile/spherical(核對手冊「Notes on Spherical Damage」確認死光不是球狀武器,是一般光束武器且是 `DamageForHit` 手冊 worked example 出處,現行武器表也沒有任何真正的球狀武器);新增 `shell.ResolveMissileShot`(AMR 攔截 + Jam Chance 躲避)、`shell.ResolveSphericalShot`(已測試但暫無武器掛載,備妥待未來新增);`fireRound`/`battleVolley` 依 `CombatShip.Kind`/`combatant.kind` 分流,beam 行為不變(回歸測試)。詳見 `docs/tech/tactical-combat-weapon-kinds.md`。
- [x] AI 財政赤字修正:職務保底(MinWorkersForSolvency/DecideColonyJobsSolvent,只 Scientific 挪 1 人)+ 順修 AI 職務回寫 bug;AI BC 從發散(-217)改收斂有界(48),測試綠(見 ai-fiscal-solvency.md)
- [x] TradeGoodsIncome 接線(2026-07-11):貿易品是建造佇列選項——建造選單新增「貿易品」、`engine.ColonyState.TradeGoods` + `syncTradeGoodsFlag`、`RunEmpireTurn` 接上 2:1 換算(`EmpireOutput.TradeGoodsRevenue`)。
- [x] 原版 672 艦名池翻譯並接入(取代硬編 10 名)(2026-07-11:190 組基底詞意譯+羅馬數字流水號保留,`assets/i18n/shipname.json` + `internal/shell/shipnames.go`,見 `docs/tech/proper-noun-strategy.md` 艦名節)
- [x] 原版 829 隨機星名池翻譯並接入(取代二十八宿占位池)(2026-07-11:829 條英文名彼此互不重複——真名/圍棋術語彩蛋/克蘇魯神話等專有名詞優先意譯,虛構短音節規則化音譯,`assets/i18n/starname-random.json` + `internal/shell/starnames.go`,`genGalaxy` 改用 `randomStarNamePool`,二十八宿 `starNamePool` 已移除;見 `docs/tech/proper-noun-strategy.md` 隨機星名節)
- [x] **勝利條件(2026-07-11)**:銀河議會選舉(手冊 GAME_MANUAL.pdf p.183,`gamedata/council.go`
  +`shell/council.go`)——議會成立門檻(半數銀河已殖民 + 存續帝國數)、票數=人口(手冊無精確換算
  公式,近似1:1)、2/3超級多數勝出(沿用 `internal/engine/victory.go` 既有但先前從未接線的
  `CheckHighCouncil`)、AI當選時玩家可 accept/reject(手冊:議會無法強迫接受)、玩家達標立即
  勝利。另接殲滅所有對手勝利(沿用同檔 `CheckExtermination`,`InvadeColony` 攻陷AI唯一殖民地後
  立即偵測)。UI 僅議會畫面文字狀態,無獨立結束畫面/accept-reject 互動介面(見 HONEST-STATUS)。
  Antares母星次元傳送門勝利當時仍全無(**已於 2026-07-11 第二輪接線,見下方新任務**)。飛彈躲避/AMR/球狀傷害已接進戰鬥解算(見 task 16 分塊「戰鬥公式依
  武器類型分流」)。**(舊斷言訂正,2026-07-11 見下一項)**:議會成立門檻最初因本 remake 資料模型
  固定只有 1 個 AI 對手,曾用 `councilMinExtantRacesOverride`(=2)覆寫手冊字面值 3——這個覆寫值
  與相關斷言已隨下一項的多 AI 升級移除/訂正,不再成立。
- [x] **多 AI 對手(N=3)+ 真議會(2026-07-11)**:`NewDemoSession` 由建 1 個 AI 對手擴為 3 個
  (`internal/shell/session.go`)——3 個不同母星星(`genGalaxy` 新增 `aiHomes` 參數,均勻攤開
  母星索引,`aiHomes=1` 時與舊版逐位元相同,`RegenGalaxy` 呼叫端行為不變)、3 種不同種族名+
  `ai.Profile` 性格(席隆人/科學、姆瑞森人/好戰、布拉西人/擴張)、`PlayerSpies` 平行陣列同步。
  議會 generalize:移除 `councilMinExtantRacesOverride`,`councilEligible` 直接用手冊字面值
  `gamedata.CouncilMinExtantRaces`(=3,玩家+3AI=4個帝國,門檻真的可達);`advanceCouncil` 由
  「玩家 vs 單一 AI 二元計票」改為逐帝國(玩家+每個AI各自獨立)算票、2/3門檻用全體總票數判定,
  `PendingCouncilElection.EnemyName` 正確指向實際當選的 AI(非寫死 `AIPlayers[0]`)。~40 回合
  regression 探針驗證:3 個 AI 各自獨立成長(殖民地/軍力隨性格分化)、玩家開局經濟不 regression、
  議會用真門檻正常召開、全程無 panic、spy 對每個 AI 都結算。當時記錄的「AI 選星策略、AI 對 AI
  互動、候選人／第三方搖擺票仍未做」已於 2026-08-11 由行星價值／距離、`AIRelations` 矩陣、
  `EnableAIVsAI` 戰爭／外交模型補上；現行說明見 `docs/tech/ai-to-ai.md`、`docs/tech/victory-conditions.md`。
- [x] **安塔蘭勝利路徑(第三條,2026-07-11 第二輪)**:次元傳送門(手冊 p.106,`gamedata.Buildings`
  早已存在,`BUILDING_DIMENSIONAL_PORTAL`,前置 `TOPIC_MULTIDIMENSIONAL_PHYSICS`,先前建成後無
  任何後續流程)建成後解鎖 `internal/shell/antaran_victory.go` 的 `GameSession.AssaultAntares()`——
  沿用 `ResolveBattle` 同款 `battleVolley` 解算(比一般戰鬥更嚴格:要求防禦方全滅才算勝,呼應手冊
  「defeat the awe-inspiring Antarans」語意)。**母星防禦艦隊戰力手冊/openorion2 均無精確數字**
  (手冊只用「awe-inspiring」定性描述),保守預設 6 艘末日之星等級戰力(合計戰力384),已誠實標注
  待考證。戰勝設 `AntaranHomeworldConquered=true`,`advanceAntaranVictory`(`EndTurn` 呼叫,順序排
  在殲滅之後、議會之前,對齊 `engine.CheckVictory` 文件記載的優先序)偵測並結束遊戲
  (`Reason=engine.VictoryAntaran`)。`CanAssaultAntares()` 前置:遊戲未結束+`!DisableEvents`
  (手冊:關閉安塔蘭攻擊則本路徑不可用)+已建傳送門+艦隊非空。最小 UI:艦隊列表畫面(`fleet()`)
  加一個文字提示熱區,只在前置滿足時顯示,點擊後導向既有戰鬥結果畫面(複用 `LastBattle`)。
  單測:`internal/shell/antaran_victory_test.go`(前置條件各分支擋下、弱艦隊戰敗不誤判、強艦隊
  戰勝後正確偵測勝利、殲滅與安塔蘭同時成立時優先序不亂)。詳見 `docs/tech/victory-conditions.md`
  §4.4。**手冊三條勝利路徑至此全數接線可達成。**
- [x] **間諜 STEAL/SABOTAGE/HIDE 可玩迴圈(2026-08-09)**:`gamedata/spy.go`(手冊 `Notes on Spying` 8 個機率
  函式,先前零呼叫端死碼)接上 `internal/shell/spy.go`——訓練間諜(`TrainSpy`,花 30 BC
  remake 拍板值)→ 每回合結算(`advanceEspionage`,由 `EndTurn` 呼叫)偷科技(STEAL,偷一項
  「對方已知、我方未知」的科技,依 GAME_MANUAL.pdf p.174-175「tries to steal technologies
  you have yet to gain」推出)→ SpyVsSpy 判定(±80 淨值門檻)。玩家 ↔ 每個 AI 對手雙向生效
  (`PlayerSpies`/`AIOpponent.Spies` 皆為平行陣列/逐一結算,`NewDemoSession` 現有 3 個 AI 對手
  時同樣各自獨立算,見上一項多 AI 升級),維護費 opt-in(0 間諜時零影響)。PlayerSpyMissions
  保存逐對手 STEAL/SABOTAGE/HIDE，HIDE 跳過偷科技並套用手冊的 SpyVsSpy +20；SABOTAGE
  依原版 70 門檻與建造成本加權候選移除 AI 殖民地建築。種族／科技／玩家政府攻擊側加成已接；
  remake 的結構化 SABOTAGE 分數與 Agent 消費已接；原版三顆鈕左右順序、raw 完整分數／特殊槽位與 AI 防守 Agent/政體資料仍缺。詳見
  `docs/tech/spy-system.md`。

## ★ 2026-07-11(續 session:字型打磨 + game test 收尾 + 多人考據)
> 本段索引本 session 的完成項;細節見各 commit 與 docs/tech/。
- [x] **點陣中文字型改造(Stage 1/2,使用者指定「一致化 + 點陣字」)**:中文 UI 由 Noto 平滑向量改
  `bitmapfont/v4 FaceTC`(Cubic 11 + Ark Pixel,OFL)。最終為**混合字型**:內文(<18px)點陣、標題
  (≥18px)Noto 向量(避免點陣放大鋸齒);**主選單**整個純 Noto(使用者偏好)。可行性經全語料窮舉
  驗證(2258 字缺字 0)+ 覆蓋回歸測試(墨點判準,守未來新增字)。殖民地擦底改採按鈕面色
  (`plateFace`,修「黑塊蓋浮雕框」)+「已建:…」建築清單依欄寬截斷(`truncateToWidth`,修溢出建造欄)。
  commits `ea8821b`/`5bdb78b`/`4d84146`;設計 review `3589b9e`;決策 `docs/tech/pixel-font-decision.md`、
  `docs/tech/ui-typography-button-review.md`。
- [x] **戰鬥畫面對手改用真實 AI 名**(修硬編「賽隆人」stale 標籤,`PrimaryEnemyName` 取真 AI 種族名),`4a58665`。
- [x] **game test 全面實機驗證 + 修 GUI bug**(task 37/38/39):xvfb GUI 玩家路徑 8 畫面截圖 + 深度回合探針;
  修 3 個 GUI bug(研究盲選/殖民地面板/戰術糊字)、補殖民地總覽 **Empire Summary 面板 + Planetary/
  Production Info 懸停面板**(`c3540ee`,即 /goal 的「game test 回報問題」)。邏輯層 70 回合無 panic、
  三勝利路徑可達。
- [x] **多人對戰通訊考據 + Phase 9 開項**:原版 CD 手冊 OCR 出通訊方式(序列/數據機/IPX 區網/TEN),
  架構=決定性 lockstep;方向定案「保留 lockstep、傳輸換 TCP、先做熱座」。`docs/tech/multiplayer-architecture.md`、
  `d31d182`(見下方 Phase 9)。

## ★ 2026-08-07(反組譯對齊版面 + 英文模式 UI 層)
> 細節見 `docs/re/01-gap-report.md`、`docs/HONEST-STATUS.md`。

- **版面來源改成反組譯**:確立優先序「exe 立即數 > openorion2 `initWidgets` > LBX 資產尺寸交叉
  驗證 > 量圖」。用它重挖了新遊戲設定、殖民地、種族選擇、艦艇設計、地面戰、多人設定、熱座交接。
- **修掉三個看畫面看不出來的還原錯誤**:新遊戲左下框是 PLAYERS(不是 RACE,對手數先前寫死 3)、
  種族選擇左右擺反(原版是左肖像 + 右 2×7 網格)、艦艇設計六個艦體槽**不等距**(15/14/17/14/14/16)。
- **接上先前選了沒作用的設定**:星系年齡(先前被常數寫死成「普通」)、帝國總數 2–8、
  難度五級(補回教學)、起始科技等級的第一個真效果(曲速前開局沒有 FTL,艦隊出不了本星系)。
- **多人**:熱座(同機輪流)可玩 + 主選單 MULTI PLAYER 接上原版設定畫面。
- **英文模式 UI 層**:`cmd/moo2` 自繪畫面全部雙語。優先「讓路」(底下是原版美術就露出烘在圖上的
  英文),其餘用 `b.tr(zh, en)`。護欄是 `lang_gap_test.go`(go/ast 棘輪)。引擎層字串未做。
- **行星表面格點解鎖**(gap report 第 8 項(行星表面調查)):先前記成「獨立工程、卡在幾何表沒抽出來」,
  這輪把 `word_182C9C` 的 7×7 角點表抽出來,並用 `COLONY.LBX#8` 那張**已畫好位置**的高亮
  菱形逐點驗過(表來自程式碼、菱形來自美術檔,兩個獨立來源)。走訪順序來自 `dword_BA784`
  (遠→近,畫家演算法)。另發現建築 sprite 是**已畫好位置**的 640×480 稀疏圖
  (資產 = 型別×36 + 格號,共 49 種),貼 (0,0) 就對位。
  建築編號對照同日補齊:算式來自 `Cache_Load_Bldg_` @ 0xAF6DC,編號由 openorion2 的
  `BUILDING_*` 列舉與原版 `TECHNAME.LBX` 第 295 條起的字串順序互相對上。
  離線合成 9 棟端到端驗證通過(每棟正落在自己的格子裡、遠近遮擋正確)。
  落在 `cmd/moo2/colonysurface.go` + 9 條回歸測試。
- **原版建築表挖出來,建造成本全部換真值**(gap report 第 8 項(行星表面調查)):`Real_Building_Name_`
  兩行組語就指出資料段有 49 筆 × 19 位元組的建築表。+8 = 建造成本 PP、+12 = 維護費、
  +14 = 分類(7 = 衛星)。驗證方式:Armor Barracks 的 +8 = 150 對上先前唯一有實據的那筆,
  +12 與手冊維護費 **40/40 全中**(獨立來源,能全中代表欄位判讀與建築編號對照都對)。
  `EstimatedCost` 欄位整個拿掉——40 項全是真值。舊估計錯得不小:星辰要塞 800→**2500**、
  行星屏障護盾 1200→500、歡樂穹頂 800→250、核心廢料場 550→200。
- **地表擺放規則已讀完**(同第 8 項(行星表面調查)):`Set_Random_Seed(colonyIdx, 0, 144)` → 每個殖民地的
  擺法固定;逐棟蓄水池抽樣選空格;房屋借用衛星編號 3/14/40/41 當四種外觀。
  (原版 PRNG 已於同日實作,見下方第 8 項(行星表面調查)。)
- **半透明標記索引**(gap report 第 8 項(行星表面調查)):原版對來源索引 >= 0xF0 的像素**從不直接寫進畫面**
  ——`Draw_` 走混色查表、`Draw_No_Glass_` 整個跳過。240..255 是十六種半透明標記不是顏色。
  remake 先前一律當顏色上色,建築陰影因此變成洋紅。`internal/lbx` 加了
  `TranslucentIndexMin` / `HasTranslucent` / `ToRGBADropTranslucent`(= 跳過那條路徑),
  既有 `ToRGBA` 行為不變。**不能在解碼層一刀切當透明**:BEAMS 有 69% 是這種索引,
  丟掉光束就消失了。混色表在 BSS(執行期才建),產生它的程式碼還沒找到,不假造係數。
- **殖民地畫面中段還給行星表面,建造佇列搬回原版彈出視窗**(gap report 第 8 項(行星表面調查)):
  `cmd/moo2/buildqueue.go` = `Build_Queue_Popup_` @ 0xB4041(框架 COLBLDG.LBX#0,可建清單
  x 13..184 / y 20+19i、佇列 7 格 x 207..458 / y 329+20i、六顆鈕座標全是反組譯真值);
  入口是框架上那顆 CHANGE(原版它就是「換要蓋什麼」,先前畫成灰的沒接)。
  中段改畫行星表面(當時只有格線 + 建築圖)。
  中文模式 17 張逐像素比對:只有殖民地畫面變了。
- **原版 PRNG + 原版擺放演算法 + 地表底圖**(gap report 第 8 項(行星表面調查),同日):
  - `internal/gamedata/origrand.go` = `Random_` @ 0x1247A0:32-bit LCG(`0x41C64E6D` / `0x3039`)
    + 拒絕取樣,回傳 **1..n**。戰鬥、事件、星系生成日後都能靠它更接近原版。
  - 擺放換成 `Make_Bldg_Array_For_Colony_` @ 0xBC30B:`Set_Random_Seed(colonyIdx)` → 建築
    依編號順序蓄水池抽樣選空格(分類 7 的衛星排除)→ 房屋 人口/3+1 → 依分類氣泡排序。
    只有最後那段隨機微調沒做。
  - 地表底圖:`C_Anims` @ 0xBBA8E 的跳表解出 **COLONY2.LBX#49**(星空)+ **PLANETS.LBX**
    第 `氣候×3 + 變體` 張(地形)。PLANETS 恰好 30 張 640×480 = 10 氣候 × 3 變體,
    與 `gamedata.PlanetClimate` 逐項相同(#27 = Gaia 蔥綠、#3 = Radiated 熔岩)。
    佔位格線移除——原版地表上沒有格線。
  - ⚠ **抓到一個沒有症狀的軸向錯誤**:格陣索引與角點表對調 → 整張佈局鏡射,徵狀只有
    「建築擠在遠端」。定案證據是 `Add_Bldg_Fields_` 同時用 (v1,v2) 當位址與座標參數。
    護欄 `TestColonyGridKeyMatchesOriginalAddressing`。
  - **軌道衛星**(`Draw_Colony_Satellites_` @ 0xBE366):x = 295 ± i×50、y = 162,
    圖是 COLONY.LBX 9/10/11/12/16。⚠ 編號要經過 `sub_BBB9F` 的 `add edx, 9` —— 漏掉會去讀
    資產 0..4,那五格在檔案裡是**零長度**的,畫面上什麼都不會出現而且不報錯。
    抑制規則 `sub_BC21B` 就是原版的星基 → 戰鬥站 → 星際要塞升級鏈。
  - 仍缺:建築集合與原版有差(Colony Base、已完成的一次性改造沒建模)→ 落點尚未對原版實測、
    植被層沒做、地形變體那一欄用 PRNG 代替原版的存檔欄位(不是真值,已標明)。
    擺放微調與母星國會大廈已於第 14 項(地表道路)補上。
- **星圖:艦隊圖示 + 旗色順序**(gap report 第 9 項(星圖艦隊圖示),同日):
  - 艦隊圖示換成原版的(`Get_Ship_Icon_Pict_Seg_`:BUFFER0.LBX `205 + 旗色×4 + 縮放`),
    取代先前那個 8×8 青色方塊。remake 星圖沒有縮放,固定用縮放 0(最小那張)。
  - ⚠ **旗色順序修正**:先前是 紅/黃/綠/**藍/白/紫/橙/棕**,後五個全錯位——而艦隊圖示是
    `205 + 旗色×4`,所以選藍色會開出白色的艦隊,中文模式完全看不出來。兩個獨立來源對上
    才改(BUFFER0 每組實測代表色 + openorion2 `FONT_COLOR_PLAYER_*`):
    正確順序是 **紅/黃/綠/銀/藍/棕/紫/橙**(第 4 色原版叫 SILVER 不是 White)。
  - 拿掉星圖左上那行「研究:<主題名>」——原版星圖沒有這東西,而且它壓在星星與艦隊圖示上。
    右欄第 5 格(綠燒瓶)改放每回合研究點數,跟其他四格一致。
  - 順帶把整張星圖的 9 層繪製順序抄進 gap report,標出 remake 各層的完成度。
- **星圖視差星空背景**(gap report 第 9 項(星圖艦隊圖示),同日):星圖區先前是一片純黑佔位,
  現在鋪的是原版 `Draw_Paralax_` 的三層星空(STARBG.LBX 資產 0/1/2,各 640×480)。
  原版三層各自以不同速率捲動並環繞平鋪;remake 星圖不捲動,位移固定 0 → 三層各貼一次。
  - ⚠ **底色不能省**(踩過):三層全是稀疏的暗點疊在透明上,底色一拿掉透明處就露出底下的
    框架美術,整片星圖變成白底黑點。原版也是先 `Fill` 再貼視差層。
    這條規則**測不到**(ebiten 在 game loop 外不准 `Image.At()`),只能靠截圖驗收。
  - 順帶修掉 `decodeAsset` 對 nil resolver 的 nil 解參考(三個新 helper 都補了守衛)。
  - 仍缺:星雲需要銀河生成先產出星雲表(資料模型缺口,不是繪圖缺口);蟲洞連線、遷移連線、
    星門、外交燈號同理。
- **星圖:星球換成原版 sprite**(gap report 第 9 項(星圖艦隊圖示),同日):先前是自畫色圓。
  資產 = `148 + 光譜×6 +(縮放 + 大小)`,**三個獨立來源互相印證**(反組譯
  `Get_Star_Picture_Seg_`、openorion2 `_starimg[class][zoom+size]`、實測 BUFFER0 那 36 張)。
  - 公式自己證明了自己:光譜 6(黑洞)不加大小 → `148+36+縮放` = **184+縮放**,
    正好等於 openorion2 另外命名的 `ASSET_GALAXY_BHOLE_IMAGES 184`。
  - ⚠ 縮放在原版是**銀河尺寸的函式不是玩家控制**(銀河越大畫越小)。remake 把座標正規化
    攤滿視窗、沒有捲動,那個對應接不過來,**不存在忠實值**;固定用 3(最縮小),
    這是 remake 的選擇不是原版真值。
  - 仍缺:5 幀閃爍動畫、黑洞的獨立繪製路徑與阻擋航線機制。
- **蟲洞**(gap report 第 10 項(蟲洞),同日):第 9/9 項的結論是「剩下的層卡的是資料模型」,
  這一項去補模型。蟲洞是其中最有遊戲價值的——它是**機制**不是裝飾。
  - `Star.Wormhole`(-1 = 無,對應原版 +0x29),**必須雙向**(openorion2 對單向直接丟例外)。
  - 產生照抄原版拿得到的部分:母星/黑洞不可當端點、最多 200 次拒絕取樣、最短距離門檻、
    候選收滿 19 個就停。⚠ **數量與距離門檻的單位沒照抄**——原版數量是銀河產生過程中累加的
    (上限 `galaxySizeParam×4+4`),接不過來;改用同構的 `星數/8` 夾 1..4。非原版真值。
  - 星圖第 1 層畫連線(在星球之前),兩端都沒偵測到就不揭露。
  - `SendFleet`:兩端有蟲洞 → ETA 1(手冊 p.181「a single turn」)。
  - ⚠ **別和隨機事件的蟲洞搞混**——MOO2 兩者都有,remake 也是(`applyWormhole` 早就有)。
  - ⚠ **舊存檔沒有這個欄位,零值是 0**,會讓每顆星都宣稱與星 0 有蟲洞。讀檔走
    `normalizeWormholes`,有專門的回歸測試。
- **把「原版 48 棟 vs remake 40 棟」的差集查清楚**(gap report 第 11 項(48棟建築盤點),同日):
  8 個沒建模的編號逐一認出來並從原版建築表讀真值。結論:2 個是自動給予(Capitol /
  Colony Base,正確地不該在表裡)、3 個已是 `SpecialActions`、**3 個真的缺**
  (Galactic Currency Exchange 250PP/3BC、Stellar Converter 1000PP/6BC、
  Artificial Planet 800PP/0BC)。
  - 順帶驗證:三個 SpecialAction 的成本與重抽逐項相同;且**維護費 0 = 一次性**這條規則
    正好把 8 個編號分成「常駐建築」與「一次性」兩堆。
  - ⚠ **抓到一條死路科技**:`TOPIC_GALACTIC_ECONOMICS`(6000 RP)解鎖的
    `TECH_GALACTIC_CURRENCY_EXCHANGE` 沒有任何東西消費它——研究完什麼都不會發生。
  - 這輪**沒有直接補上**:效果查不到(手冊沒這條、patch 手冊零命中、遊戲資料檔只有名字沒說明、
    建築表沒有效果欄),而不編數字是紀律。下一輪的路線寫在 gap report 第 11 項(48棟建築盤點)。
- **恆星轉換器(行星版)補上 + 一個乾淨的負面結果**(gap report 第 11 項(48棟建築盤點),同日):
  走完第 11 項(48棟建築盤點)留下的路線。三棟裡做成一棟,另兩棟拿到「為什麼做不成」的證據。
  - **先做正對照再下結論**:把 48 個「已建」旗標位移全掃一遍——43 個出現過(技術有效),
    完全沒出現的只有 5 個,而那 5 個正好是 2 衛星 + 3 一次性,與「維護費 0」那條規則
    挑出來的是同一批。**兩個獨立判準指向同一個分組。**
    ⚠ 建築 18 的 `+0x148]` 全檔只出現一次,基底卻是科技陣列不是殖民地——同一個數字出現在
    不同結構裡,這正是不能直接數命中的理由。
  - **恆星轉換器**:手冊 p.106(400 傷 ×2、維護 6)與原版建築表第 42 列(1000 PP、維護 6)
    **維護費逐項相符**才建模。接進 `colonyDefense`(+800)與 `origBuildingID`(42)。
  - ⚠ **推翻一個把記帳慣例當成規則的斷言**:`TestBuildingsCount` 原本釘死 40,理由是
    「手冊全表 35+5」——但那是手冊把恆星轉換器另立一節的記帳方式,原版表裡它就是第 42 棟。
    **40 從一開始就不是原版的數字。** 現在 41。
  - (⚠ 「飛彈基地/地面砲台要等空間模型」這句**當天就被自己推翻了**,見下一項。)
- **飛彈基地/地面砲台接進防禦解算**(gap report 第 11 項(48棟建築盤點),同日,第 11 項(48棟建築盤點)的訂正):
  空間模型早就在(`satellite.go` 的 300/450 都是手冊確認值,`retaliationAttackers` 也早就
  支援這兩棟)——缺的只是 `colonyDefense` 沒接上,它用的是自編的
  `CommandPointsFromBuildings × 10`。
  - ⚠ **挖出一個自相矛盾**:同一座星基在 `colonyDefense` 值 10(比巡洋艦 8 還強),
    在 `retaliationAttackers` 值 3–4(≈ 驅逐艦 tier)——而 `satellite.go` 的校準註解明講
    是後者。**兩邊都測得綠**,因為沒有任何測試同時看這兩條路徑。現在有了。
  - 改用同一套推導後:反擊戰力隨武器科技成長、兩棟防禦建築真的有用、1.3/1.5 的 arc-cost
    差異自動吃到。
  - `TestAIRaidRepelledByFleetAtStar` 的自我守衛**正確地觸發了**(防禦 19 < 願打門檻 21),
    處理方式是把測試裡的母星升級成戰鬥站,不是把模型改回去遷就測試。
  - IDA 的 `-Ohexrays` 這一輪起不來(error code 4,兩次),整項改用手讀 `.asm` 完成。
- **兩個零值陷阱**(都加了回歸測試):`TechLevel` 零值 = 曲速前 → 舊存檔艦隊全凍住;
  `i18n.English` 是 `Lang` 的零值 → 忘了設 lang 的建構路徑會靜默變英文。

## Phase 0 — Kick-off / 可行性(本輪)
- [x] 盤點 openorion2 完成度(`docs/kickoff/01`)
- [x] 中文化策略(`02`)
- [x] 按鈕中文化策略,參考 moo1 避免重蹈覆轍(`03`)
- [x] 字型選擇研究(`04`)
- [x] LBX 資產 + patch 1.3/1.5 處理與版本架構(`05`)
- [x] ebiten 移植策略(`06`)
- [x] 可行性總論(`00`)
- [x] PLAN.md / WORKLIST.md
- [x] .gitignore(擋版權素材)
- [x] README(含致謝)——`README.md` §致謝
- [x] 本機 git commit(push 待使用者確認)

## Phase 1 — 資料層移植(純 Go)
- [x] Go module 初始化 + docker build 環境(`go.mod`、`docker/`、`scripts/build.sh`)
- [x] LBX 容器解析(magic 0xfead、offset 表)— internal/lbx,真實檔驗證
- [x] scan-line RLE 影像解碼 — internal/lbx/image.go
- [x] palette 解析(6-bit → 8-bit)— 解碼與上色解耦(Frame.ToRGBA)
- [x] 影像多幀(frame offset 表)+ 多 palette variant(ToRGBA 套不同 palette)
- [x] Bitmap(8-bit indexed):像素編碼與 Image 相同(image.go 已涵蓋);dirty-block 為 SDL 局部 blit 優化,ebiten 全繪不需,刻意不移植(見 docs/tech/lbx-format.md §2.7)
- [x] 存檔 schema(對照 gamestate.cpp,全部完成並驗證):
  - [x] reader + GameConfig(59B)+ Galaxy/Nebula(32B)
  - [x] Colony×250 / Planet×360 / Star×72 / Leader×67 / Player×8(內嵌 ShipDesign/Weapon/Settler)/ Ship×500
  - [x] 全區段解析驗證:SAVE10.GAM 解出種族 Trilarian/Alkari/Mrrshan/Sakkra/Klackon、首星 Orion、計數全合理、SeqEnd 收斂(203596,合成全零檔同值當回歸護欄)
- [x] 資料枚舉/常數字典(技術/建築/種族特性/氣候/礦產/特殊裝備)— internal/gamedata/enums.go(28 枚舉,自 gamestate.h 生成)+ docs/tech/enums.md + 抽查測試
- [x] 唯讀衍生公式移植(艦艇戰力/HP/戰速、行星產出、雇用費)— internal/gamedata/formulas.go + docs/tech/formulas.md + 測試(researchCost 待 LBX 資料表)
- [x] 檔案覆蓋順序載入(基礎 → 1.31)— internal/assets/resolver.go(有序搜尋路徑、大小寫不敏感、OpenLBX)+ 測試
- [x] 單元測試:lbx/save/gamedata/assets 皆有合成測試;lbx/save 另有 env-gated 真實檔測試(MOO2_LBX_TEST / MOO2_SAVE_TEST)

## Phase 2 — ebiten backend + 最小可跑 ⭐
- [x] ebiten 專案骨架(Update/Draw/Layout)— cmd/moo2 + ebiten v2.9.9
- [x] palette 上色 → `ebiten.Image`(Frame.ToRGBA → NewImageFromImage → DrawImage)
- [x] docker + xvfb 截圖流程打通 — docker/Dockerfile.ebiten(CGO+X11+GL+xvfb)+ scripts/screenshot.sh(ReadPixels 存 PNG,不依賴 WM)
- [x] ★ 顏色視覺驗證:MAINMENU 資產 21 於 ebiten 渲染出完整正確主選單(640×480)
- [x] 確認 MOO2 為 640×480(非 320×200);修正 kickoff 假設
- [x] 滑鼠與鍵盤全面在用(`inpututil`),整個 `-game` 靠它跑
- [x] 已有快取(`overlay.go` / `multiplayer.go`)
- [x] ★ 里程碑 M2:載存檔 → 繪製星系圖(cmd/moo2 -save;SAVE10.GAM 的 36 星依座標/光譜/大小 + 星名 + 星雲,資料驅動)
- [x] `starsprite.go` 用 GALAXY.LBX 0x94(148)+ 星空/星雲已接(`nebula.go`、`starbg_test.go`)

## Phase 3 — UI 框架 + 文字系統 + 主選單(做法見 `08` playbook)
- [x] CJK 渲染:`internal/uifont`(ebiten text/v2,依尺寸快取 face)+ Measure。**2026-07-11 升級為混合字型**:zh 內文用 `bitmapfont/v4 FaceTC` 點陣(<18px)、標題用 Noto 向量(≥18px);en 純 Noto。見 `docs/tech/pixel-font-decision.md`
- [x] 顯示層覆蓋 i18n:`internal/i18n`(TSV 英文即 key + 查無 fallback + TranslateFormat)+ 測試
- [x] [HARD] 只翻顯示層,不動資料層(i18n 設計即如此)
- [x] 字型:NotoSansCJK-Regular.ttc 經 Go opentype.ParseCollection 驗證可解析+量測中文(★ [HARD] 相容檢查通過);galaxy 標題已渲染繁中
- [~] **逐字斷行已完成**(`internal/uifont/wrap.go`);描邊/陰影仍無(`font.go` 只有 Draw/DrawCentered/Measure),條目自己也已把它降為次要
- [ ] 字型子集 pyftsubset(docker)+ go:embed 內嵌(待譯文集齊;目前用完整 .ttc runtime 掛載)
- [x] 主選單中文化 + 截圖校對(cmd/moo2 -menu:擦底疊字六按鈕繼續/載入遊戲/…;before/after 見 docs/reference-screens.md)
- [x] 主選單:語言 中/英 runtime 切換(2026-08-07,`cmd/moo2/interactive.go` 的 `toggleLang`)
      ——先前只有啟動旗標 `-lang`,進了遊戲換不掉,不符 `CLAUDE.md` 那條需求
- [x] **英文模式覆蓋率(UI 層)**:2026-08-07。切換機制早就有,內容沒補齊——overlay 那條路徑
      天生雙語(英文模式跳過擦字疊字、露原版美術),remake **自繪**畫面則寫死中文。
      這輪把 `cmd/moo2` 整層收掉:優先「讓路」(遊戲選單/多人設定/種族選擇/載入儲存/
      戰術控制列 → 直接露原版烘在圖上的英文),其餘用新建的 `b.tr(zh, en)`。
      順帶修掉兩個中文模式看不出來的 bug(RACESEL#33 標題橫幅從沒被畫過、
      `diplomatRaceIndex` 與艦體名各有一份會漂移的重複對照表)。
      護欄:`lang_gap_test.go`(go/ast 棘輪,只能往下調)+ `lang_coverage_test.go`(漏填英文欄)。
- [~] **英文模式覆蓋率(引擎層)**:星名池 / 艦名池 / 開局艦隊名 / 支援艦艦級、殖民地行星環境與建築、
      歷史圖表指標與 AI 圖例、NEW GAME 五個值列都已改成「資料保留識別 key、UI 依語言轉顯示名」
      (第 84、88、89、91、92 項)。
      **剩下的是仍待分類的引擎產生敘述字串**。⚠ `internal/` 的中文字串裡有一大批是**查表 key**
      不是顯示字串,不能全域替換——換成英文會直接查不到。正解是讓引擎回鍵值、UI 端繪字,
      不是把 `tr()` 灑進引擎；目前尚未宣稱引擎層完成。逐檔清單、量法與收尾原則見
      `docs/HONEST-STATUS.md`「英文模式覆蓋率」
- [x] 主選單:版本 1.3/1.5 選擇框架(`toggleVersion`,左下角)
- [x] hover 機制與共用金色外框已接(`interactive.go` 的 `hover` 欄位);垂直置中仍屬觀感微調

## Phase 4 — 畫面重建 + 完整中文化(做法見 `08` playbook)
- [x] 原版畫面對照組(`docs/reference-screens.md`:主選單/行星列表/建造,英文原貌 + 翻譯清單)
- [x] 通用畫面覆蓋渲染器(`cmd/moo2/overlay.go`:資料驅動擦底疊字,選單+行星列表共用)
- [x] 主選單中文化(6 按鈕)+ 行星列表中文化(18 標籤,before/after)
- [x] LBX 字串資源解析 + dumper(`internal/lbx/strings.go` + `cmd/lbxstrings`);TECHNAME 560 條科技名 dump 成功
- [x] **科技/元件名譯表完整(`assets/i18n/tech.json`:419 條唯一全翻)** — 研究主題/領域、武器/裝甲/護盾/引擎/電腦、建築、艦種、武器改造(含縮寫);覆蓋驗證 419/419 無遺漏
- [x] i18n TSV 守護測試(載入所有 assets/i18n/*.json + 佔位符一致性)
- [~] 觀感微調,保留
- [x] 其餘字串源逐一 dump + 翻(2026-07-11 盤點:多數已完成,見 assets/i18n/):科技描述 techdesc.tsv(83)、種族 races/raceinfo.tsv、事件 event.tsv(98)、外交 diplo.tsv(780)、help.tsv(704)、母星名 starname.tsv、技能 skilldesc.tsv、estrings(585)/rstring(178)/antaran、艦名 shipname.tsv(535,同日稍後完成,見下方獨立項)、隨機星名 starname-random.tsv(829,同日稍後完成,見下方獨立項)
- [x] **★ 調色盤鏈解鎖(關鍵)**:對照 openorion2 `gfx.cpp Image::load` 破解「無內嵌調色盤畫面」上色機制(基底提供圖 + 本圖部分內嵌疊加);實作 `cmd/moo2/interactive.go` `resolvePalette`;研究選擇(TECHSEL,借 SCIENCE 調色盤)完整渲染驗證。見 `docs/tech/palette-chain.md`
- [x] 依 `palette-chain.md` 對照表逐畫面上色——機制是 `resolvePalette` + `paletteChain`,各畫面都在用。剩 COLGCBT(地面戰 sprite)的來源未定案,見 `cmd/moo2/groundcombat.go` 檔頭

## Phase 4b — 串接互動(還原原版的骨幹,-game)⭐
> 各原版畫面不再各自獨立 flag,而是串成單一可導覽的互動程式(`cmd/moo2 -game`)。目標:開機進原版主選單,滑鼠點選在原版畫面間跳轉,全繁中。
- [x] 互動骨架:`origScreen`/`origTransition` 介面 + `overlayScreen`(真 LBX 背景 + 中文擦底疊字 + 點擊熱區)+ `sceneBuilder` + `interactiveApp`(ebiten.Game,支援 headless 腳本驗證)
- [x] 導覽:原版主選單(真美術)→「新遊戲/繼續」→ 真原版行星列表 →「返回」→ 主選單(headless 驗證通過)
- [x] 調色盤鏈畫面併入導覽 + 小於全螢幕視窗置中
- [x] 研究選擇畫面**完整中文化**(擦底疊字,PIL 量測校對,完整垂直切片)
- [x] 調色盤鏈擴充多段鏈(`paletteChain []assetRef`;艦隊三段鏈驗證)
- [x] **★ 星系主樞紐(galaxy GUI,BUFFER0.LBX 0)接成遊戲主畫面**:新遊戲→星系主畫面,
  底部工具列(座標取自 galaxy.cpp)導覽到 行星/艦隊(FLEET)/軍官(OFFICER)/科技總覽(INFO);
  各畫面返回樞紐。全部忠實原版美術,headless 驗證導覽鏈通過
- [x] 星系工具列中文化(殖民地/行星/艦隊/領袖/種族/情報/回合)
- [x] 艦隊列表中文化(艦隊作戰/全部/調動/拆解/軍官/支援/戰鬥/返回)
- [x] 軍官列表中文化(殖民地領袖/艦艇軍官/雇用/人才庫/解雇/返回)
- [x] 科技總覽中文化(星曆/歷史圖表/科技總覽/種族統計/回合摘要/參考資料/返回)
- [x] 擦底採樣穩健化:samplePlate 左緣帶+上下橫帶眾數;背景均勻畫面(info)改 overlayScreen.eraseColor 強制底色
- [x] galaxy 工具列 GAME 標題已翻(→遊戲)+ ZOOM 已翻(→縮放)(2026-07-10);行星/艦隊個別按鈕邊緣極微殘(紋理按鈕固有)為長尾
- [x] 各子畫面 RETURN 按鈕精確熱區:`interactive.go` 有逐畫面的精確熱區(`{582,452,52,20}`、`{536,432,82,22}`、`{556,430,84,28}` 等)。
- [x] 科技總覽「科技總覽」列可點進研究選擇畫面(其餘選單項待接)
- [x] 殖民地總覽畫面(COLSUM.LBX 0)接入 COLONIES 按鈕 + 完整中文化
- [x] 種族關係畫面(RACES.LBX 0)接入 RACES 按鈕 + 中文化(種族關係/會晤/報告/宣戰/忽略/加成/返回)
- [x] **★ 真新遊戲流程**:主選單→新遊戲→原版 NEW GAME 設定畫面(1.31 `NEWGAME.LBX#28`／1.5 `#31`,調色盤鏈 RACEOPT#4→NEWGAME#1)→ACCEPT→星系主畫面;中文化(難度/星系大小/星系年齡/玩家數/科技等級/戰術戰鬥/隨機事件/安塔蘭攻擊/取消/接受)
- [x] **★ 獨立種族選擇畫面(2026-07-10,對原版流程還原)**:依 GAME_MANUAL 流程,設定畫面 Accept 改導向獨立種族選擇畫面(`cmd/moo2/raceselect.go`,RACEOPT#0 螢幕框 + 14 族中文名 + 真肖像 RACESEL 15–28 字母序 + 描述 + 取消/接受)。取代原「設定畫面擠一格循環種族」。研究見 `docs/tech/newgame-flow.md`。
  - [~] 版面像素對齊原版 + 用 RACESEL 名稱按鈕圖/描述板;Custom 點數畫面;命名+旗色;依 Starting Civilization 真實母星初始(WORKLIST 續,task 8)
- [x] 回合摘要畫面(TURNSUM.LBX#0)接入 TURN 流程(結束回合→摘要顯示本回合結算:星曆/淨工業/研究/食物/稅收/國庫變化/研究完成)→關閉回星系。中文化(回合摘要/關閉)
- [x] 艦艇設計畫面(DESIGN.LBX#0)接入(艦隊→點艦艇格→艦艇設計)+ 中文化(艦艇設計/巡防艦…末日之星/清除/取消/建造);艦隊 RETURN 改精確熱區
- [x] 議會畫面(COUNCIL.LBX#1)接入 + 投票系統(2026-07-11 大改,見下方「勝利條件」任務):舊版
  `CouncilVote`(無成立門檻、無2/3多數、票數=人口較高者當選)已移除,畫面改讀
  `GameSession.CouncilStatus()` 誠實呈現議會是否已成立/目前票數/是否已分出勝負或待玩家回應
- [x] 已探測定位背景(remain-scan,待接入):讀取存檔 LOADSAVE.LBX#11(空存檔格)、外交 DIPLOMAT.LBX#29(有雜訊待查)
- [x] **存檔/讀檔(remake 自身格式)**:GameSession JSON 序列化(shell/persist.go),AI Decider 以性格重建、含未匯出遊戲狀態;每回合自動存檔(UserConfigDir),主選單「載入遊戲/繼續」讀回續玩。測試 TestSaveLoadRoundTrip(Turn/BC/種族/星系/艦隊/建造/AI 一致且可續跑)
- [x] ESPIONAGE/SABOTAGE/HIDE 三顆任務鈕：remake 已接明確標籤的三任務循環與逐對手存檔；原版左右順序仍未知，SABOTAGE 以已證實的建築破壞切片結算
- [x] **★ 核心遊戲迴圈第一步**:GameSession 接進 -game;TURN 按鈕呼叫 session.EndTurn()
  (結算帝國經濟 + AI 對手決策),星系畫面即時顯示星曆(3500 起,每回合+1年)+ 國庫 BC
  (overlayScreen.extras 動態文字機制)。驗證:TURN×2 → 星曆 3500→3502、國庫 100→106
- [x] 待接入畫面:議會/艦艇設計/回合摘要 已接入並中文化(見上)+ 單一殖民地管理已自建(見下);讀存檔背景已備、入口待接
- [x] 殖民地總覽填即時資料:玩家各殖民地列出「殖民地 N / 農夫 / 工人 / 科學家」(來自 GameSession,對齊原版欄位,extras 動態文字)
- [x] 行星列表填即時資料(每星生成行星:名/氣候/重力/礦產/大小,PIL 量測列對齊)
- [x] 軍官列表填即時資料(領袖名單:名/專長/等級,4 槽位對齊)**(2026-07-11 追加:技能字串已從純裝飾接上真加成,見 Phase 5「領袖/軍官技能」條目)**
- [x] 艦隊畫面填即時資料(艦隊名冊:艦名/艦體等級)
- [x] 造船系統:艦艇設計點艦體等級→BuildShip 加艦到 session→艦隊顯示新艦(第二個互動系統)
- [x] 單一殖民地管理:殖民地總覽點職務欄→ShiftColonyJob 重分配人口(影響下回合經濟);造船改扣國庫 BC(戰力×20)
- [x] 建造佇列:殖民地總覽建造欄可點選建築(住宅/工廠/研究實驗室/星港),結束回合以淨工業累積建造,完成回合摘要通知
- [x] 星圖互動:點星→黃色高亮環+左下角行星資訊面板(名/氣候/大小/重力/礦產)+ 派遣艦隊鈕(星間航行)
- [x] 程序化星系生成:genGalaxy(種子亂數,抖動網格佈 24 星,隨機光譜/大小/洗牌星名,玩家/AI 母星)取代固定佈局
- [x] 星系大小接 NEW GAME 設定:GALAXY SIZE 框可點選(小型12/中型24/大型36/巨型48星),ACCEPT 依選定大小 RegenGalaxy
- [x] 難度設定生效:DIFFICULTY 框可選(簡單/普通/困難/不可能),敵方戰力倍率套用到戰鬥
- [x] **★ DIPLOMAT 解碼修復**:對照 openorion2 發現多幀動畫需「累積各幀(delta)+ 未寫入填 palette[0]」(先前當透明→白噪)。lbx.Image.AccumulatedRGBA;外交議事廳(DIPLOMAT#29,38幀)以真原版圖 + diplomat#0 調色盤渲染,疊外交對談(和平/貿易/威脅)→ 16/16 原版畫面皆真圖
- [x] 種子隨新遊戲變化(newGameSeed 遞增);戰術戰鬥艦艇移動+射程限制開火(格位/選取/移動)
- [x] 數值對齊 MOO2 規格:艦體成本(空殼生產18/60/180/540/1620/4860,設計畫面顯示)、建築成本(自動工廠/海軍營/研究實驗室60、太空港100、星基300)、研究成本(gamedata 權威 cResearchCosts 表)
- [x] 艦艇設計武器元件:選主武器(無武裝/雷射/質量投射器/核飛彈/離子砲,各成本+攻擊加成),建造成本=艦體+武器、戰鬥攻擊=艦體+武器
- [x] 完整艦艇元件系統:武器/裝甲/護盾/特殊四類元件(各含成本+效果係數),設計畫面循環選擇+顯示總價;建造套用(裝甲/護盾→HP、武器/戰鬥電腦→攻擊)
- [x] 元件解鎖綁研究科技:各進階元件標記所需 gamedata 研究主題,未完成研究則鎖住(循環跳過),設計畫面顯示已解鎖數;研究→解鎖元件→造艦系統打通
- [x] 元件品項擴充:29 個 MOO2 真實元件(武器11:雷射→死光/裝甲7:鈦→氙素/護盾6:第一~第十級/特殊5),真譯名(tech.tsv)+ 遞增係數 + 各綁研究科技門檻
- [x] **「需 OCR」是假阻塞**。第 64/64 項用可抽文字的 patch 1.5 PDF p.124-127 拿到全表,18 把武器全部換成真值
- [x] 精確全表:手冊 p.124-127 的完整武器表已抽出(第 64/64 項,可抽文字的 patch 1.5 PDF)。版本專屬 profile 的**規則值**已分版(`gamedata.RuleProfile`);**1.5 實體 LBX 的回歸驗收仍開**,見剩餘工作表第四節。
- [x] **研究自動推進 → 動態解鎖迴圈**:目前主題完成後自動推進到下一個未完成元件主題(researchQueue 依成本遞增),玩數回合便逐步解鎖進階元件。測試 TestResearchUnlockLoopOverTurns 驗證 40 回合解鎖 7→15、完成 6 主題
- [x] 新遊戲種族選擇:NEW GAME 設定畫面加種族選擇框(13 經典種族循環選,顯示名+特性),ACCEPT 套 ApplyRace 起始加成(工業/研究/食物/成長/國庫/戰鬥百分點,對齊各族招牌特性)。測試 TestApplyRaceBonuses/SakkraGrowthFaster/MrrshanCombatBonus
- [x] hover highlight remake 外框已接(外交／戰術／戰機與按鈕共用金色框);原版逐像素觀感仍開放
- [x] 淘汰自製簡約殼(`-play`):2026-08-08 第 81 項已刪除(448 行 + 兩個旗標 + main.go 分支)。可玩迴圈走原版 overlay 畫面 + `internal/engine`
- [x] ~~補齊需全域調色盤鏈的畫面到對照組~~ 這些畫面在遊戲裡早就跑起來了;`docs/reference-screens.md` 的靜態對照組收錄落後於實作,已在該文標明
- [x] dumper 已齊(`cmd/lbxdump`、`cmd/lbxstrings`、`cmd/lbxinfo`),24 份 TSV 共 5,049 條由此而來
- [x] 逐畫面重建:主選單/星系圖/行星清單/殖民地/科技研究/艦隊/軍官/種族資訊/對話框/載存檔皆已建(載存檔在 `cmd/moo2/loadgame.go`,過場截圖廊第 70 拍)。
- [x] `cmd/moo2/overlay.go` 已實作擦底疊字
- [x] LBX 字串譯文表:科技名/描述、種族、事件、外交、星名、help、技能、殖民地、議會、選單等 22 個逐源分檔 TSV 已完成(assets/i18n/*.json);艦名池(2026-07-11 補完,shipname.tsv)、隨機星名池(2026-07-11 補完,starname-random.tsv)均已落地,四個專有名詞池全數定案
- [x] `internal/i18n` 有 `TranslateFormat`(含測試)
- [~] 24 份 TSV 本身就是譯名的單一來源;獨立術語表由 2026-08-08 新建的 `CONTEXT.md` 部分承接(它收的是**專案內部術語**,不是遊戲專有名詞)。「中文(英文)」小字控制碼仍未做
- [x] `scripts/screenshot.sh` + `-gamegallery`(目前 35 張，含 `01b_newgame`)已是常規驗收流程

## Phase 5 — Gameplay 引擎重建
- [x] 回合結算主迴圈(engine.RunEmpireTurn:殖民地經濟聚合+稅收+國庫+研究推進)
- [x] 殖民地經濟:食物/工業/研究/稅收/國庫已實作(engine);人口成長回寫 Population 已補(shell.advancePopulation 累加 PopGrowth 達門檻 +1 人口、新單位依缺糧保護指派、受 PopMax 上限)。本歷史項目的舊 300 點調校門檻已於 2026-08-24 由官方 1.50 尺度 1,000 點取代；目前限制見頂端活表。測試 TestPopulationGrowthWriteback/CappedAtMax
- [x] 建造佇列 + 建築長期效果:advanceBuilds 完工後套用永久產出加成,每殖民地每種只套一次(ColonyBuildings 去重);殖民地總覽顯示已建建築。**(2026-07-11 忠實化訂正)**先前把手冊「殖民地整體固定加成」揉進 per-worker 欄位湊數(自動工廠工業/工人+2、研究實驗室研究/科學家+5 等,小殖民地過度受益、大殖民地不足),現分開建模:per-worker 訂正回手冊值 + 新增 `FlatFood`/`FlatIndustry`/`FlatResearch`(固定加成)、`IncomeBonusPercent`(太空港+50%/證券交易所+100%,逐殖民地精確套用於 `RunEmpireTurn`)、`PopMax` 直接加成(生態圈+2)、`FlatGrowth`(複製中心)。機器人工廠(2026-07-11 已接線,見下)。共 18 棟已忠實建模數值,詳見 `docs/tech/colony-buildings.md` §6。測試 TestBuildingLongTermEffect/TestResearchLabEffect/TestSpaceportIncomeBonusPercent/TestBiospheresRaisesPopMax 等(engine+shell)
- [x] 機器人工廠礦產豐度分級接線(p.82)**(2026-07-11)**:比照重力懲罰的接線手法(`4c2a26a`),`engine.ColonyState` 新增 `MineralRichness gamedata.PlanetMinerals` 欄位,獨立保留建立殖民地當下的原始礦產豐度分類(先前只烘進 `IndustryPerWorker` 靜態費率,事後拿不回原始分類)。零值陷阱處理:`gamedata.ULTRA_POOR` ordinal=0,故全部既有 `ColonyState{...}` 建構點(engine/shell 測試、`cmd/moo2sim`)皆已明確補上本欄位。`applyBuildingEffect` 的機器人工廠 case 依 `gamedata.ProdRoboticFactoryBonus(int(cs.MineralRichness))`(`internal/gamedata/production.go` 既有查表函式,索引與 `mineralProductionTable` 一致)查出手冊固定值(Ultra Poor+5/Poor+8/Abundant+10/Rich+15/Ultra Rich+20)加進 `FlatIndustry`,不動 `IndustryPerWorker`。存檔行星由 `ColonyStateFromSave` 讀 `save.Planet.Minerals`(與 `gamedata.PlanetMinerals` 同源 openorion2 enum ordinal,可直接轉型,同重力)。母星固定 Abundant。測試:`TestRoboticFactoryEffect`(母星 Abundant+10)、`TestRoboticFactoryEffectByMineralRichness`(五級分級逐一驗證,含 UltraPoor+5/Rich+15)(shell)。
- [x] 重力懲罰接進生產管線(**2026-07-11**):`ColonyState` 新增 `PlanetGravity` 欄位,`colonyFood`/`RunColonyTurn` 對食物/工業/研究三種 per-worker 產出套用 `gamedata.GravityPenaltyPercent`(Low-G -25%、Heavy-G -50%;士氣+重力先加總成單一百分點再套一次 `GravityAdjustedProduction`,避免兩次連續整數除法的複合誤差,理由見 `internal/engine/colony.go` 註解)。行星重力產生器 `NormalizeGravity` 旗標由 no-op 變成真的會歸零懲罰。`ColonyStateFromSave`(存檔↔engine 橋接)同步接上 `save.Planet.Gravity`(與 `gamedata.PlanetGravity` 同源 openorion2 enum ordinal,直接轉型)。種族 Low-G/High-G 重力天賦未建模,固定以一般種族為基準;固定加成(Flat*)不吃重力。**已知現實限制**:本專案唯一的殖民地建構點(`NewDemoSession`/`playerHomeworldColony`)固定 Normal-G,尚無「開拓新殖民地」流程會產生 Low-G/Heavy-G 殖民地,故此接線在 demo session 暫不可見,主要對存檔載入模式(`RunGameTurn`)生效。測試 TestRunColonyTurnGravityHeavyPenalty/TestRunColonyTurnGravityNormalizeGravityCancelsPenalty/TestRunColonyTurnGravityNormalGNoPenalty/TestRunColonyTurnGravityAndMoraleCombinedPercent/TestColonyStateFromSaveGravityMapping(engine)
- [x] 士氣(Morale)接進 MoralePercent(**2026-07-11**):`GameSession` 新增 `Government`(`gamedata.MoraleGovernmentType`)欄位,`ApplyGovernment` 記錄政府型態(`Governments` 索引→`moraleGovByIndex`,四選一映射到對應基礎政府,進階政府 Imperium/Confederation/Federation/Galactic Unification 不區分)。新函式 `colonyMoralePercent`(`internal/shell/session.go`)= `gamedata.MoraleGovernmentBase(gov, hasBarracks)`(手冊 -20%/無 Barracks)+ 全息模擬艙(`MoraleHoloSimulatorBonus`+20%)+ 歡樂穹頂(`MoralePleasureDomeBonus`+30%),依 `ColonyBuildings` 讀取已建建築;政府變更(`ApplyGovernment`)與建築完工(`advanceBuilds`→`recalcColonyMorale`)皆會重算。**母星起始 `MoralePercent` 從無據硬編 +10 訂正為忠實值 0**(獨裁 + 已建 Marine Barracks 抵消 -20% 懲罰,無士氣建築加成;見 `playerHomeworldColony` 註解,`TestGameSessionEndTurn` 已同步訂正預期值 33→30)。多種族懲罰、首都失守懲罰與重建解除均已由後續 typed 狀態鏈接入；Virtual Reality Network 仍因手冊定性為「成就」而不在建築表。詳見 `docs/re/capitol-state-audit-20260826.md` 與 `docs/tech/colony-buildings.md`。
- [x] 指揮評等(Command Rating)供需接線(**2026-07-11**):手冊 p.169「size class」公式(Frigate=1..Doom Star=6,`gamedata.ShipCommandCost`,以 Titan=5/Doom Star=6 兩處具體數字交叉驗證)+「每未覆蓋點 -10 BC」超支懲罰,先前 `gamedata.IncomeCommandOverflowCost` 是零呼叫端死碼。供給端:星基+1/戰鬥站+2/星辰要塞+3(三者取代不疊加,`gamedata.CommandPointsFromBuildings`)。`engine.PlayerState` 新增 `CommandPointsSupply`/`UsedCommandPoints` 欄位,`shell.GameSession.EndTurn` 每回合依實際已建成軌道衛星(`totalCommandPointsSupply`)與艦隊(`usedCommandPoints`)重算,`engine.RunEmpireTurn` 算超支併入 `NetBC`(新增 `EmpireOutput.CommandOverflowCost` 曝露懲罰金額)。當時誤判「開局母星 1 座星基(+1)vs 3 艘開局艦艇(需求3),缺口2點恆定-20BC/回合」為手冊忠實結果,實為**regression**(見下方同日修復項)。誠實未做(手冊有數字但架構未跟上,詳見 `docs/tech/moo2-formulas-reference.md`「指揮評等供需」節):通訊科技(Tachyon+1/Hyperspace+3,每軌道衛星)、Imperium 政府 +50%(本專案政府型態全域固定 Dictatorship,無 Imperium 狀態)、Operations 軍官技能(手冊無精確數字)、AI 對手(抽象 FleetStrength 無逐艦清單,供需維持零值無懲罰)。測試 TestShipCommandCost/TestShipCommandCostOutOfRange/TestCommandPointsFromBuildings(gamedata)、TestRunEmpireTurnCommandOverflow/TestRunEmpireTurnCommandSupplyCoversDemand(engine)、TestTotalCommandPointsSupply/TestUsedCommandPoints/TestUsedCommandPointsEmptyFleet/TestEndTurnCommandOverflowPenalty/TestUsedCommandPointsUsesGamedataTable(shell)。
- [x] 指揮評等開局死亡螺旋 regression 修復(**2026-07-11**,同日接線後發現):原版存檔反推每帝國基礎供給 5，另加軌道衛星供給；`gamedata.CommandPointsBase` 與超支測試已接。單一存檔無法區分少數更深層欄位語意，證據邊界見 `docs/tech/moo2-formulas-reference.md`「指揮評等供需」節與頂端活表。
- [x] 科技研究樹推進（`RunResearchPhaseWithRoller` 累積、超額突破率、成功清零；`session.advanceResearch` 自動推進主題）
- [x] 艦隊移動 + 星圖導航:SendFleet 依星距換算 ETA,EndTurn 跨回合推進,抵達標記探索;星圖點星→面板「派遣艦隊至此星」鈕 + 青色艦隊標記 + 航行連線 + ETA 顯示。測試 TestFleetInterstellarMovement
- [x] 艦艇設計畫面在 `interactive.go`,過場截圖廊第 86 拍(`25_shipdesign.png`)
- [x] 戰鬥:格子戰術戰鬥(2026-07-10 換原版美術:STARBG 星空+COMBAT 控制列+可見 CMBTSHP 艦艇+控制列 7 按鈕中文化;逐發用真 ResolveShot 命中/傷害/過盾/過甲);宣戰→戰術戰鬥→戰鬥結果。**(2026-07-11 更新:武器依 beam/missile/spherical 分流,飛彈躲避/AMR/球狀傷害公式接進解算,見 `tactical-combat-weapon-kinds.md`)**。**艦型 sprite 對照已接(task 12,2026-07-11)**:網搜定 CMBTSHP 色塊結構(8 色×45)+ 視覺比對定尺寸,戰鬥依艦級顯示不同大小 sprite、玩家/敵艦不同色塊,取代單一 placeholder(近似對照非原版精確 picture 映射,見 `docs/tech/cmbtshp-ship-sprites.md`)
- [x] 外交對談(2026-07-10 破解 DIPLOMAT.LBX 換原版美術:逐族使節房+使節疊合,13 族對應對 RACESEL 核實);銀河議會選舉勝利條件(2026-07-11,見下方勝利條件任務,取代原本無門檻/無2/3多數的簡化投票)
- **2026-08-11 勘誤**：上一行戰鬥舊摘要的「網搜近似 picture 映射」已由 IDA `sub_30062 @ 0x30062` 取代為已證實的 `45*playerColor+rawPicture`（`rawPicture 0..43`）；原版 20 幀朝向 timer 仍未知，但 remake 已接移動後固定 tick 近似，詳見 `docs/tech/cmbtshp-ship-sprites.md` 與 `docs/re/remake-consumer-closure-20260811.md`。
- [x] 隨機事件 remake 基礎系統：已實作事件具真實效果、種子化可重現並顯示於回合摘要；
  舊 30%／6 種自編事件模型已被原版 0..28 候選與 `sub_2230A` 排程取代。未實作事件仍如實
  消耗候選而不產生空播報；完整效果與全銀河目標以頂端活表為準。
- [x] 安塔蘭人入侵:週期性終局威脅(前20回合寬限,之後每15回合一次),強度隨次數升級,攻母星(人口+BC損失,有界),母星艦隊可部分防禦減損;顯示於回合摘要(紅色警報)。測試 TestAntaresRaidsScheduleAndEscalate/DefenseReducesDamage
- **事件排程現況**：2026-08-25 已以 IDA Pro 閉合並接入原版排程；事件 record、AI／熱座
  全局目標及未實作效果仍分開標示，不以排程測試代替整體 parity。
- [x] 第 47 項(AI艦隊移動)+ 第 55 項(AI 科技先前靠偷)之後,三塊都已完成
  - [ ] 精讀 1oom `game_ai_classic.c`,抽「AI 決策流程」語言無關筆記
  - [ ] 精讀 GameFAQs MOO2 AI FAQ + 策略指南,補 MOO2 特有行為
  - [x] 設計可插拔 AI 介面(ai.Decider)+ 難度加成係數(已用於經濟+態勢)
  - [ ] 標示「必須逆向才能確定」的項目(若有)
- [x] 開新遊戲流程:種族選擇 + 星系大小/難度 → ApplyRace/RegenGalaxy(見 Phase 4b)
- [x] 地形改造(Terraforming)/蓋亞轉化(Gaia Transformation)/土壤改良(Soil Enrichment)接線(**2026-07-11**):`internal/gamedata/terraform.go` 移植好的氣候階梯/人口係數公式先前零呼叫端(死碼),現接進殖民地建造佇列。新增 `engine.ColonyState.Climate` 欄位(比照 `PlanetGravity`/`MineralRichness` 的零值陷阱處理:`gamedata.TOXIC` ordinal=0,`playerHomeworldColony`/`ColonyStateFromSave` 皆已明確補上;此欄位不像 Gravity/MineralRichness 被每回合核心公式讀取,只在地形改造/蓋亞轉化套用瞬間讀寫,故其餘既有測試字面值不受影響、無需逐一補值)。新增 `internal/gamedata/special_actions.go`:`SpecialAction`/`SpecialActions`/`SpecialActionByNameZH`/`AvailableSpecialActions`,把這三項「Special」型別一次性行動(區別於常駐 Building,不計入 `colony-buildings.md` 40 項建築表)排進 `availableBuildOptions`/`allBuildOptions`。前置科技(地形改造 `TOPIC_GENETIC_MUTATIONS`、蓋亞轉化 `TOPIC_TRANS_GENETICS`、土壤改良 `TOPIC_ADVANCED_BIOLOGY`)取自 `openorion2/src/tech.cpp` 的 `research_choices[]`(陣列索引=`ResearchTopic` 列舉值,已與既有 34 項建築前置科技逐一交叉核對 100% 相符,地形改造的 `TOPIC_GENETIC_MUTATIONS` 亦與 `terraform.go` 檔頭「移植自...『Genetic Mutations』章節」的手冊出處吻合)。`shell.advanceBuilds` 新增分流:這三項完工時呼叫 `applySpecialAction`(不記入 `ColonyBuildings` dedup map,因手冊明講地形改造可重複套用,若記入 dedup 會被既有「已建過不再套用」邏輯擋下第二次),推進氣候(`TerraformNextClimateOptions`/`GaiaTransformationCanApply`)並用新增的 `gamedata.TerraformPopMaxAfterClimateChange` 等比例縮放 PopMax、`ClimateFoodPerFarmer` 差值疊加 FoodPerFarmer(保留既有建築加成)。**誠實近似/TODO**:PopMax 縮放非精確重算(remake 無「行星尺寸→基礎人口容量」對映表,詳見該函式註解);建造成本(PP)手冊無數據,比照其餘估計建築的 RP 量級外推(260/900/150),手冊「地形改造每次套用成本遞增」未模擬(固定成本);Barren 地形改造下一級的兩個候選(Desert/Tundra)手冊未給選擇條件,固定選第一個。測試:`TestTerraformPopMaxAfterClimateChange`/`TestSpecialActionByNameZH`/`TestAvailableSpecialActions`(gamedata)、`TestTerraformAdvancesClimateFoodAndPopMax`/`TestTerraformNoOpWhenNoNextClimate`/`TestGaiaTransformationRequiresTerran`/`TestSoilEnrichmentBlockedOnHostileClimate`/`TestSoilEnrichmentWorksOnHospitableClimate`(shell)。詳見 `docs/tech/colony-buildings.md` §6.1 地形改造列、`docs/HONEST-STATUS.md` 2026-07-11 追加段。
- [x] income.go 三個零呼叫端死碼接線(**2026-07-11**,解鎖自本輪稍早的開局經濟平衡修復):
  ①**政府 money 加成**(MANUAL_150.html govt_bonus democracy_money=10→50%/federation_money=15→75%,
  `gamedata.IncomeApplyGovernmentMoneyBonus`)。新增 `gamedata.IncomeGovtMoneyBonusPercent(gov)` 查表
  (Democracy→50、Federation→75、其餘→0)+ `engine.PlayerState.GovtBonusMoneyPercent` 欄位(呼叫端
  算好傳入,同 `Maintenance`/`CommandPointsSupply` 輸入模式)。`shell.GameSession.EndTurn` 依
  `s.Government` 算好傳入,`RunEmpireTurn` 在逐殖民地迴圈**結束後**(帝國層級,非逐殖民地——政府
  是帝國屬性不是殖民地建築)對 `TaxRevenue+FoodSurplusRevenue+TradeGoodsRevenue` 套一次,差額併入
  `TaxRevenue`。demo 預設 Dictatorship→0,no-op;AI 對手無 `Government` 欄位建模,不受影響。
  ②**運輸艦(Freighter)維護費**(每艘使用中 -0.5 BC,`gamedata.IncomeFreighterMaintenanceCost`)。
  新增 `engine.PlayerState.ActiveFreighters` 欄位,`RunEmpireTurn` 算出 `EmpireOutput.FreighterMaintenanceCost`
  併入 `NetBC`。當時(本條寫下時)本專案艦種塑模(`gamedata.ShipType`:`COMBAT_SHIP`/`COLONY_SHIP`/
  `TRANSPORT_SHIP`/`OUTPOST_SHIP`)沒有獨立的「Freighter」艦種,呼叫端恆傳 0,目前 no-op,接線先備妥。
  ★ **此缺口已於同日稍後補上(見上方 Phase 7 §「#4 運輸艦淨現金版本差異」條)**:新增「運輸艦隊」
  建造選項後,玩家側 `ActiveFreighters` 真的變非 0,維護費隨之生效,並補上 1.3/1.5 版本現金加成
  差異。③**士氣對收入的調整**
  (`gamedata.IncomeMoraleAdjustedProduction`,手冊 p.170)**判定為刻意不接**:查證
  `internal/engine/colony.go` `RunColonyTurn` 發現士氣(`MoralePercent`)早就套進食物/工業/研究的
  per-worker 產出(`pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)` 套 `GravityAdjustedProduction`),
  `RunEmpireTurn` 的 `TaxRevenue`(讀 `co.NetIndustry`)/`FoodSurplusRevenue`(讀 `co.FoodSurplus`)/
  `TradeGoodsRevenue`(讀 `co.NetIndustry`)全部是從這個已調整過的產出直接換算,若再套一次士氣就是
  雙重計算(同一筆錢士氣生效兩次)。故不呼叫該函式,判定依據完整記錄在 `engine/empire.go` 註解與
  `docs/tech/moo2-formulas-reference.md`「士氣對收入的影響」節;函式本身與其單元測試保留(驗證公式
  正確,非死碼)。三項在 demo 對局皆 no-op(政府=Dictatorship、無貨運艦種、母星 morale=0),20 回合
  BC 軌跡探針確認接線前後一致(101→130 健康爬升,無 regression)。測試:
  `TestIncomeGovtBonusFormula`/`TestIncomeFreighterMaintenanceCost`/`TestIncomeMoraleAdjustedProduction`/
  `TestIncomeApplyGovernmentMoneyBonus`(gamedata,原有公式測試)、
  `TestRunEmpireTurnGovtBonusMoneyPercent`/`TestRunEmpireTurnGovtBonusMoneyPercentZeroNoOp`/
  `TestRunEmpireTurnFreighterMaintenance`/`TestRunEmpireTurnFreighterMaintenanceZeroNoOp`(engine,新增)、
  `TestEndTurnGovtBonusMoneyWiring`(shell,新增)。詳見 `docs/HONEST-STATUS.md` 2026-07-11 收入死碼段落、
  `docs/tech/moo2-formulas-reference.md`「政府對 BC 收入的加成」/「士氣對收入的影響」節。
- [~] 傘狀項,同 task 16
- [x] 最小拓殖(Colonization)接線(**2026-07-11**):玩家可用已抵達無主星的殖民船建立人口 1 的新殖民地；適居性、人口容量、消耗殖民船、平行陣列與回合經濟均有測試。完整證據與當時模型邊界見 `docs/tech/colonization.md`，目前差異只看頂端活表。
- [x] AI 拓殖建真殖民地(**2026-07-11 追加**):上一條的「仍缺」補上——`aiExpand` 先前只設
  `Star.Owner=2`+`OwnedStars++`,從不建立 `engine.ColonyState`,AI 殖民地數恆為開局母星 1 筆、
  `RunEmpireTurn` 的 `TotalNetIndustry` 永遠停在初始母星產出,AI 版圖擴張與經濟成長脫鉤。抽出
  `internal/shell/colonization.go` 的共用函式 `newColonyFromStar(starIdx, gov, foodBonus,
  indBonus, resBonus) (engine.ColonyState, ok, reason)`,把 `ColonizeStar`(玩家)原本內嵌的
  「氣候/重力/礦產/大小解析 → PopMax 查表 → 起始職務 → 士氣算法」搬進去,兩處呼叫端(玩家
  `ColonizeStar`、AI `aiExpand`)共用同一套建法,不再各算各的。`aiExpand` 佔星時 append 進
  `AIOpponent.Colonies` + `ColonyStars`(AIOpponent 唯二的殖民地平行陣列——不像玩家有
  Builds/ColonyBuildings/PlayerColonyMarines 等逐殖民地建造/駐軍追蹤,因為 EndTurn 對 AI 只呼叫
  `RunEmpireTurn` 結算經濟,從不呼叫那些玩家專屬的 advance* 流程,故無需同步)。**AI 政府型態
  未建模**(`AIOpponent` 無 `Government` 欄位),士氣一律用 `gamedata.MoraleGovDictatorship`
  保守預設;AI 無種族加成模型,`foodBonus`/`indBonus`/`resBonus` 一律傳 0,誠實簡化不臆造。
  維持既有「每 5 回合擴張一次」節奏不變(未改成每回合)。40 回合探針對照:修前 AI 殖民地數恆
  1、FleetStrength 線性成長(3→60);修後 AI 殖民地數隨回合增至 9、FleetStrength 加速成長
  (3→101),玩家開局 BC 軌跡兩版本一致(102→…→96),無 regression。測試:
  `internal/shell/ai_behavior_test.go` 新增 `TestAIExpand_CreatesRealColony`(佔星後建真殖民地、
  平行陣列同步)、`TestAIExpand_ColonyParticipatesInEconomy`(新增殖民地輸出進入帝國結算對應 slot)、
  `TestAIExpand_NoOpWhenNoUnownedStars`(無星可擴張時安全 no-op)。詳見
  `docs/HONEST-STATUS.md` 與頂端活表。
- [x] 第 45~45 項逐條清點，**領袖技能 26 項已有 remake 消費端**；Tactics 依原版自己未實作，Famous 招募機率已補證，Diplomat 接受門檻仍為 oracle 留白
- [x] task#36 已完成(mod 層 + 佔格 + 傷害)

## Phase 6 — 音樂 / 音效
> 第一性原理翻案(2026-07-10):MOO2 **沒有 XMI/MIDI 音樂**,全部是 LBX 內的 22050Hz 8-bit PCM WAV。故無需 SoundFont/OPL 合成——原封播原版 PCM 即 bit-identical。研究定案見 `docs/tech/audio-format.md`。
- [x] ~~逆向 .lbx 音樂(XMI)格式~~ → 實為 PCM WAV,存 STREAM/STREAMHD.LBX(格式研究文件已定案,含 provenance)
- [x] 逆向音效格式 → SOUND.LBX 內 WAV;entry0 為 20-byte 名稱表(BUTTON1…),已解出 68 個具名音效
- [x] ebiten 音訊播放整合 — `internal/audio`(WAV 解碼→16-bit stereo、Mixer BGM 迴圈+SFX;headless 停用避免無音效卡崩潰)+ 單元/真檔測試綠
- [x] 接線:主選單 BGM(STREAMHD)+ 按鈕點擊音效(BUTTON1)— `cmd/moo2/audiohook.go`
- [x] 曲目/UI 事件對應(2026-07-10 定案到靜態溯源極限):外交樂反組譯硬證(track 13/14/15);menu/galaxy/combat 對應 Play 函式在 DOS build 為死碼,維持時長啟發式(誠實標,再定案需聆聽或 Windows build RE)。見 `audio-track-map.md` 第七節
- [x] ~~`CMBTSFX/SPHERSFX` 巢狀音庫格式逆向~~ **(2026-07-11 前提翻案,rulebook 62/63)**:CMBTSFX/SPHERSFX **不是音效庫,是戰鬥視覺特效動畫**(79 資產,爆炸/光束/護盾命中多幀 sprite,標準 LBX 影像,`lbxinfo` 直接解得);戰鬥**音效**全在 SOUND.LBX(68 具名音效已解碼含 NRGBLAST/PHOTON/TORPDO1/EXPL/SHIPHIT1/SHIELD…)。見 `docs/tech/audio-format.md`
- [x] 戰鬥音效接線:SOUND.LBX 的 NRGBLAST/MISLFIRE/SHIPHIT1… 已接進戰術戰鬥(`cmd/moo2/audiohook.go`)
- [x] (選)CMBTSFX 爆炸/光束特效動畫接進戰術戰鬥畫面(視覺增強,2026-08-10)
- [x] ~~SoundFont 處理~~ → 不需要(無 MIDI 音樂)
- [ ] 桌面實測驗收:使用者對原版聆聽比對(主選單 BGM + 點擊音是否為正確曲/音)

## Phase 7 — 版本 1.3 / 1.5(2026-07-11 大幅推進)
- [x] 研究「1.3 → 1.5 規則差異清單」:逐條核對 1730 行 CHANGELOG_150 + MANUAL_150 + PARAMETERS.CFG,`docs/tech/version-1.3-1.5-diff.md`。結論:落在已實作系統的真差異只 3 個(多數 CHANGELOG 是 bug fix 或「新增可調參數但預設=經典值」)
- [x] rule profile 資料結構:`gamedata.RuleProfile` + `GameVersion` + `Profile13/15`(`internal/gamedata/ruleprofile.go`)
- [x] 1.3/1.5 profile 實作 + 驗證(值 + 預設 Profile15=現行 三層回歸斷言)
- [x] 主選單版本切換生效(**2026-07-11 收尾完成；2026-08-24 勘誤**):UI + 開局注入 + 三個 live 消費端全部接線——①軌道轟炸的炸彈攻擊當量(1.3=5/1.5=10；整體仍固定三外圈)②電漿砲傷害③超先進科技研究成本。
- [x] diff 全量表 15 項 2026-07-11 全數盤點完畢
  - [x] **AI 殖民地建築資料模型 + 戰略轟炸傷亡回寫(2026-07-11；2026-08-24 IDA 勘誤)**：`AIOpponent` 的建築、陸戰隊與戰車平行陣列均已持久化。舊版「按建築名字母序全部吸收，再依行星尺寸扣人口」已被 `sub_DCEBD @ 0xDCEBD` 直接反證並移除；現在以 raw building ID 48→1 建候選，排除 `{8,9,26,27,40,41,42,47}`，再混入每名陸戰隊、戰車、建造進度項與人口。抽中成本不足的單位即停止；寫回人口、駐軍、建造進度與建築 map，GAM adapter 同步保存 `BuildProgress` 與最後人口點數。1.3 建築 +1 hit 仍是 CHANGELOG 強推論；八類獨立防禦戰鬥者的精確摧毀鏈仍列在頂端活表。
  - [x] **#14 衛星/軌道防禦基地「space 預算武器平台」+ 版本相依 beam arc-cost(2026-07-11)**:`internal/gamedata/satellite.go` 新增獨立衛星/基地 space 預算(飛彈基地 300、地面砲台 450——手冊 p.78/p.81 確認值;星基/戰鬥站/星辰要塞 250/500/1200——借用 `ShipHullSpace` 同量級近似值)+ arc-cost 佔格公式(比照 `WeaponSpaceWithMods`)+ fit 公式;`RuleProfile` 新增 `SatelliteBeamArcCostPct`(1.3=25/1.5=33)、`GroundBatteryBeamArcCostPct`(1.3=0/1.5=50,CHANGELOG_150.TXT 1.50.7/1.50.10)。`internal/shell/orbital_bombardment.go` `retaliationAttackers` 改簽名讀 defender 科技(`bestUnlockedWeaponValue`,新 helper)+ profile,取代舊 shipStrength 4/8/16 固定 tier,推導出「隨科技變強」+「隨版本 arc-cost 不同而不同」的反擊戰力。校準除數 `SatelliteStrengthScale=20` 使雷射參考點下星基/戰鬥站重現舊 tier 4/8,星辰要塞算出 20(非近似 19,誠實標見常數註解)。平衡 sanity:開局艦隊轟炸開局 AI 母星(僅星基),Profile13/15 各掃 Turn 0..14,最大損艦數皆為 1(不破壞平衡)。測試:`internal/gamedata/satellite_test.go`(fit/arc 公式錨點)+ `internal/shell/satellite_defense_test.go`(版本差異/科技效果/飛彈基地不吃 arc/地面砲台/平衡 sanity)。誠實限制:AI 現行資料模型無研究進度推進機制,`bestUnlockedWeaponValue` 在 `NewDemoSession` 自然對局裡恆落到 fallback 分支(雷射/核飛彈),「科技變強」效果目前只能在單元測試手動建構已解鎖科技的 `PlayerState` 觀察到。
  - [x] **#4 運輸艦淨現金版本差異(2026-07-11 補實作)**:新增「運輸艦隊」(Freighter Fleet)殖民地建造選項(`gamedata.FreighterFleetActionName`,前置科技 `TOPIC_NUCLEAR_FISSION`,估計建造成本 PP60——沿用既有 Special 一次性行動框架,見 `gamedata/special_actions.go`)。完工時 `shell.GameSession.applySpecialAction`:`s.Player.ActiveFreighters += gamedata.FreighterFleetShipsPerBuild`(手冊 p.168:每次建造 +5 艘)+ `s.Player.BC += s.RuleProfile.FreightersCashBonus`(新 `RuleProfile` 欄位,1.3=5/1.5=0,出處 MANUAL_150.html「Buildings & Freighters Free Cash Bug」+ CHANGELOG_150.TXT 1.50.8)。維護費(每艘 0.5 BC/回合)不用另外接——`engine.PlayerState.ActiveFreighters` 先前已接進 `RunEmpireTurn`(恆 0 no-op),本輪讓它真的變非 0,維護費隨之自動生效。**批次 B 的 #10 也已確認非差異**(見 `version-1.3-1.5-diff.md` #10),批次 B 至此結案。**簡化(誠實標)**:只模擬手冊「固定回饋」那一側,不模擬 0-3 BC 建造當下維護費立即扣款那一側;不做完整貨運/補給物流(殖民地間運食物/殖民者)——運輸艦本輪只有「可建造+維護費+版本現金加成」三件事;**AI 未接同一建造流程**,`ActiveFreighters` 對 AI 恆為 0。測試:`TestSpecialActionByNameZH`/`TestAvailableSpecialActions`(gamedata,新增運輸艦隊斷言)、`TestProfile13Values`/`TestProfile15Values`(gamedata,新增 `FreightersCashBonus` 斷言)、`TestFreighterFleetBuild*`(shell,新增:完工增加 ActiveFreighters+國庫、1.3 vs 1.5 現金加成差異、維護費隨後續回合生效、開局不建造回歸不變)。詳見 `docs/tech/version-1.3-1.5-diff.md` #4、`docs/tech/moo2-formulas-reference.md`「運輸艦淨現金版本差異」節。
  - [x] **#13 掃描/偵測距離:輕量戰爭迷霧(2026-07-11)**:新增 `internal/gamedata/detection.go`(`ScannerRangeParsec` 基礎2/Space4/Neutron6/Tachyon8、`OrbitalScannerBonusParsec` 星基+2/戰鬥站+4/星辰要塞+6 擇一取代不疊加、`ParsecToNormalized`=1/10 換算常數、`DetectionRangeNormalized` 加總換算——**全部近似**,手冊無公開 parsec 數字)+ `RuleProfile.SensorRangeVersionBonusParsec`(1.3=0/1.5=1,對應 MANUAL_150.html「Scanners and Communications Discrepancy」修正的整體近似,非逐科技數字)+ `internal/shell/detection.go`(`GameSession.VisibleStars`/`starVisible`,啟用先前無人讀取的 `Star.Explored` 死旗標;可見條件:已探索 ∪ 玩家自己的星 ∪ 落在玩家殖民地/艦隊偵測範圍內)。`cmd/moo2/interactive.go` `drawStarmap` 接上 fog 繪製(未偵測星降噪成暗灰小點、不畫星名/擁有環;可見星維持全繪)。調參依據:量測 `NewDemoSession()` 實際程序化星系(24星,種子42)鄰近星距離,使開局 Profile13 可見 3 顆星、Profile15 可見 7 顆星(母星區可見一小圈、遠星入霧,版本差異可觀察)。**誠實邊界**:fog 純視覺,不 gate 選星/派艦/殖民/轟炸等任何操作;不做敵艦 map blip(AI 艦隊為抽象戰力,無地圖座標,零地基)。測試:`internal/gamedata/detection.go` 無獨立測試檔(純查表函式,經 `ruleprofile_test.go` 的 `SensorRangeVersionBonusParsec` 斷言覆蓋)+ `internal/shell/detection_test.go`(6 個測試:母星可見+範圍外不可見、已探索恆可見、版本差異合成盤面+真實星系、軌道基地加成星辰要塞>星基、艦隊偵測源、`VisibleStars`/`starVisible` 接線+越界安全)。`go build`/`go vet`/`go test` 全過;`moo2sim -turns 20` 經濟軌跡不變(fog 不碰回合邏輯)。
- [x] 資產分版(1.31 vs 1.5 LBX/資料)——`-data13/-data15` 與主選單解析器切換已接；以 1.31 基礎資料 + `MOO2-1.50.26.zip` 的 `patch/150/lbx` 實際跑兩版畫廊各 35 張，並釘住 1.5 `NEWGAME.LBX` 背景資產 #31

## Phase 8 — 文件 / 考究 / 文化 / 研究
- [x] 遊戲歷史與當年評價考究(`docs/history/moo2-history-and-reception.md`,角色:歷史考究專家,14 來源)
- [x] GitHub 致謝(README:openorion2/1oom/mom/字型/社群/Simtex)
- [x] 技術知識庫:LBX 資產格式 / 存檔格式 / 枚舉 / 公式 / ebiten 移植筆記(`docs/tech/`)
- [x] 華人圈中文討論資訊考究章節(`docs/history/moo2-chinese-community.md`,歷史考究專家,31 來源+誠實揭露侷限)
- [x] 華人圈文化現象(`docs/culture/moo2-chinese-cultural-phenomenon.md`,文案作家,事實有本、無 AI 味)
- [x] sprite/tile 畫質優化可行性 markdown(`docs/tech/sprite-tile-quality.md`)
- [x] UI 界面調整可行性 markdown(`docs/tech/ui-adjustment.md`)
- [x] `docs/tech/` 已有 54 篇
- [x] 三平台打包 CI(`docs/tech/packaging.md`):macOS(`.github/workflows/build-macos.yml`,`macos-14` runner 原生編 arm64+amd64 → `lipo` universal → `.app`/`.dmg`/`.tar.gz`)+ Linux/Windows(`.github/workflows/build-desktop.yml`);YAML 經 actionlint + yaml.safe_load 驗證,尚未在真 Mac 上實跑驗證(無 Mac 測試機)
- [x] 本機 Docker 打包腳本（`docs/tech/packaging.md` §5）：`scripts/package-appimage.sh`（Linux AppImage、linuxdeploy+appimagetool）、`scripts/package-windows.sh`（Windows zip）已實際跑過，公開產物統一寫入 `dist-all/MasterOfOrion2-cht-x86_64.AppImage`、`dist-all/MasterOfOrion2-cht-windows-amd64.zip`，並驗證內容（解壓／objdump 確認）。**推翻先前假設**：ebiten v2.9.9 Windows backend 已改純 Go（purego，無 cgo），`CGO_ENABLED=0` 即可跨編，不需 mingw-w64（`build-desktop.yml` 仍裝了 mingw，屬保守多餘，非錯誤，可留後續簡化）。
- [x] `cmd/moo2` 譯表改 `go:embed` 內嵌 + `-i18n <dir>` 開發覆寫(2026-08-08 第 83 項),相對路徑假設已移除,從任意目錄跑實測通過

## Phase 9 — 多人對戰(hotseat / 網路 lockstep→TCP)
> 考據定案見 `docs/tech/multiplayer-architecture.md`(原版通訊 OCR 自 CD 手冊 + 架構佐證自 patch 1.5 手冊)。
> 方向(使用者定案 2026-07-11):**保留原版決定性 lockstep 架構,傳輸換成 TCP**;起步先做熱座。
- [x] 原版多人通訊考據(手冊):序列/數據機/IPX 區網(2–8人)+ TEN 網際網路服務;DirectX 6.1→DirectPlay;決定性 lockstep + host 廣播 config + 同時回合(`docs/tech/multiplayer-architecture.md`)
- [x] **熱座(hotseat)**(2026-08-07):多位真人同機輪流下令。席位交換模型
  `internal/shell/hotseat.go`(原版是 `player[i]` 陣列 + 當前索引,remake 是單數欄位 →
  換人時整組搬進 `seat`);交接畫面 `cmd/moo2/hotseat.go`(座標取自 `Draw_Hotseat_Screen_`
  @ 0x626D6);「結束回合」改成全員下完令才推進世界;席位進存檔(原版也存遊戲模式)。
  ⚠ 非當前席位的帝國在 `EndTurn` 最後才結算(差一個 AI 回合的資訊),勝負判定只對當前
  席位跑——兩點都寫在 `advanceIdleSeats` 檔頭與 `docs/re/01-gap-report.md` 第 3 項(Colony+Event 畫面)。
- [x] **主選單「多人對戰」接上實際流程**(2026-08-07):`cmd/moo2/multiplayer.go`,
  整張版面取自反組譯(`Multi_Player_Screen_` @ 0xF4D99 / `sub_F42CA` / `sub_F009A`),
  含原版自己會隱藏的按鈕(熱座模式下沒有 JOIN GAME)。NETWORK / MODEM / NULL MODEM
  畫成灰的並明示未實作,不假裝可選。
- [x] 第 29 項(決定性化)完成(`internal/shell/determinism.go` 的 `StateHash`/`StateFingerprint`,畫在 `30_netwait.png` 上)
- [x] 第 29–29 項整塊完成(`internal/netplay`,含大廳、區網探索、6 張畫面)
- [x] **指定熱座帝國**(2026-08-27 複核):`Get_Multi_Player_N_Humans_` 會計數控制碼
  `+0x28 == 100` 且狀態 `+0x24 == 0` 的帝國；remake 語意已在新遊戲流程落地——選帝國畫面
  逐一標記要接管的 `AIPlayers` 索引，固定文案由 JSON 供應且受安全框限制，由
  `SetupHotseatWithAIIndices` 建立席位;未選中的 AI 保留。
- [x] **熱座席位核心資料轉換**(2026-08-09):接管席位保留種族加成、領袖、母星建築、艦隊、
  殖民地平行陣列與玩家間諜欄位,AI 關係矩陣同步壓縮。`AIOpponent` 沒有的建造佇列、
  前哨站與傭兵池仍以空值起步,列為模型差異而非假裝已完成。

## ★ 2026-08-07 盤點:gap report 的「最大系統級缺口」四條全部已完成

逐條 grep 之後發現 `docs/re/01-gap-report.md` Part B 的四大缺口清單**整份過期**——
歷史記錄系統(`shell/history.go`)、前哨站(`shell/outpost.go`)、艙損/維修
(`shell/repair.go`)全都已經建好,事件系統早就標記完成。Part A-2 的 Smacker 過場
(`cmd/moo2/cutscene.go` + `internal/smk`)同樣是過期的。

**為什麼重要**:這四條被後續每一輪的摘要當成現況反覆引用,於是「還缺什麼」的判斷
整個偏掉。文件裡的斷言一旦成形就會被當事實傳遞,而程式碼會往前走、文件不會。
細節與訂正後的清單見 gap report 第 12 項(四大系統缺口盤點)。

核實過後真正還缺的:網路多人(整塊)、`Command_Points` 專屬畫面、星圖 4 層
(星雲/遷移連線/星門/外交燈號,卡資料模型)、2 棟建築(真值已抽出,缺效果來源)、
殖民地地表的道路與擺放微調。

> 這幾項之後陸續了結:`Command_Points` 畫面(第 13 項(指揮點數視窗))、地表道路(第 14 項(地表道路))。
> 地表**擺放微調**與**植被層**仍缺,見第 14 項(地表道路)。

## ★ 2026-08-07 指揮點數視窗(gap report 第 13 項(指揮點數視窗))

第 12 項(四大系統缺口盤點)核實後的清單裡最小的一項,做掉了。原版 `Show_Command_Points_Screen_` @ 0x8BAB9
整支只有 30 行:迷你星圖當背景 + 一塊文字視窗、ESC/點擊關閉。欄位組成由執行檔符號表
給出(`_starting_command_points_msg` / `_total_command_points_msg` /
`_total_command_point(s)_used_msg` / `_command_summary_msg`)——**結構是原版真值,
中文用字與視窗座標是 remake 自己的**。入口接在星圖右欄第 2 格(先前只顯示數字、點不開)。

⚠ **順帶抓到一個快取陳舊值**:`Player.CommandPointsSupply` / `UsedCommandPoints` 只在
`EndTurn` 更新,開局時是舊的。視窗第一版畫出「起始 5 + 軌道 0 = 總計 1」自打嘴巴;
星圖右欄那個淨值吃同一組欄位、**同樣是舊的**,只是單獨一個數字沒得對照所以一直沒被發現。
改用 `CommandPointsSupplyNow()` / `CommandPointsUsedNow()` 現算,兩處都修。

**把有關聯的數字放在同一個畫面上,本身就是一種驗證。**

## ★ 2026-08-07 殖民地地表道路(gap report 第 14 項(地表道路))

第 8 項(行星表面調查)留下的「道路沒畫」補完。建築佔 6×6 的**格子**,道路走 7×7 的**格點**;四個方向
(兩條邊 + 兩條對角線)的合法段數是 42+42+36+36 = **156**,與 `COLROADS.LBX` 的資產數
一模一樣——這個等式就是幾何解對了的確認,不必去量圖。產生規則接在建築擺放的同一條亂數流上,
每個有建築的格子抽三次 `Random(2)`,在自己的四條邊上畫框。

**dir 2 / dir 3 的 72 張對角線圖,出貨版從來不會出現。** 全執行檔對那兩個旗標只有寫 0、
沒有一處寫 1,連帶讓產生器裡「空格子」那一整條分支變成死碼。remake 只實作有建築那條——
不是簡化,是認出死碼之後不抄。

⚠ **原版資料裡有兩個位元組級的錯,照抄不修**:繪製順序表少了格點 (5,4)、(3,4) 重複兩次;
包圍判定表的 Δa/Δb 對調了兩筆。修掉會讓 remake 比原版「正確」,而驗收標準是與原版一致。

**方法上的教訓:表不要用手抄。** 第一次照著 IDA 的 `dd` 清單手抄,解出 48/49 又有一處順序
不對,分不清是原版有錯還是自己抄錯。改成直接從 `Orion2.exe` 讀位元組才定案——先用「不重不漏
+ 遞減」當指紋掃全檔(零命中,排除自己解錯位址),再用有把握的前 14 個位元組當錨點掃到唯一位置,
最後連 Go 的表字面量也用腳本產生。

**同輪補上的缺口**:道路之後原版還跑一層**植被**(`COLVEGGI.LBX`),見下方第 14 項(地表道路)。

## ★ 2026-08-07 房屋抖動 + 母星的國會大廈(gap report 第 14 項(地表道路))

把第 14 項(地表道路)留下的「抖動沒補所以道路對不上」補完,順帶抓到一件更基本的事:
**remake 的地表格陣從第一步就少放了一棟建築。**

抖動是排序之後的 8 輪隨機微調。⚠ 原版這段有兩個 bug,照抄不修:第二個座標算完之後被
**無條件覆蓋**(夾到 0 的那半段沒被編出來),於是換位對象**永遠落在主對角線上**;
內圈的偏移變數因此完全沒有作用(迴圈仍要留著,它決定抽幾次亂數)。
這兩點已用 `objdump` 對原始位元組獨立驗過,不是讀錯。

`Get_Bldg_CR_` 還有一個容易漏的語意:**「找一棟建築」會消耗亂數**——命中前每碰到一個空格
就抽一次(它和「找空位」共用同一支函式)。漏掉之後道路整串偏掉,而畫面看起來照樣正常。

**編號 9 = Capitol。** 先前判成「不可建造 → 正確地不在表裡」是錯的：它是實體建築，
佔格、有美術、會被畫出來。開局由非統一政體母星持有；失守後，原版會重新指定首都行星，
並允許在該處以 200 PP 重建。

⚠ **仍不能宣稱與原版逐格相同**:流程結構接完整了,但**建築集合**仍有差
(Colony Base、已完成的一次性改造沒建模),集合差一棟則落點全偏。尚未對原版實測。

## ★ 2026-08-07 殖民地地表植被層(gap report 第 14 項(地表道路))

原版每個殖民地畫面的空地上都長著草木,remake 整層沒有,所以畫面比原版空曠。

植物圖分群組、**每組固定 8 張**,組內編號越大株越大。資產 = `群組×8 + max(Random(8)−1−(a+b)/2, 0)`
—— 後面那項就是**透視**,越遠的格點越容易被壓到最小那株。群組由氣候決定(10 路跳表),
最大群組 12 → **13 組 × 8 = 104**,正好是 `COLVEGGI.LBX` 的資產數(這次直接跑 `lbxinfo` 驗)。
再交叉一項:群組 0 的前四張是 6×15、8×15、9×22、9×22,組內越大越大株,與透視項方向一致。

密度規則反直覺:**0 條路必長草;k 條路(k>0)機率 (k+1)/7**。而且最後那個
`Random(建築數+2)` 的結果永遠不是 0,判斷等於恆真——像是想寫「建築越多越不長草」卻沒生效,
照抄(那次抽樣要消耗亂數)。

繪製是**每格先植被再建築**的交錯,不是獨立一層——差別在遮擋。

⚠ 沒模擬:原版在「有格子被選取」時**一株都不畫**(remake 沒有這個狀態);
每株顏色沒有對原版逐張比對過。

⚠ 順帶踩到一個效能坑:尺寸在**產生**階段就要用(它進位置公式),而地表每幀重算、
`decodeAsset` 又沒快取——不處理就是每幀重解最多 72 張 LBX。

## ★ 2026-08-07 星雲(gap report 第 15 項(星雲))

星圖 4 層做掉第 1 層。**星雲不是裝飾,是有規則的地形**——手冊兩處:艦艇穿越時速度降到
1 秒差距/回合;**戰鬥發生在星雲內時所有護盾失效,裝了硬化護盾的除外**。

判定門檻是「星雲圖那一點的調色盤索引 > 5」,反組譯(`Point_Is_In_Nebula_N_`)與 patch 1.5
手冊逐字互相印證。連手冊都承認這個判定有小破洞(「深處有幾個暗像素會讓該處的星被判成不在
星雲內」)——那是原版行為。數量的四路跳表上限 4,與 `internal/save` 從存檔格式反推的
`maxNebulas = 4` 獨立對上。圖在 STARBG.LBX(和星空層同檔),12 種 × 4 個縮放。

**順帶激活兩段構不到的碼**:`DamageHardShieldBonus` 先前沒有元件載體等於死碼,這輪把
「硬化護盾」加進可選元件(與隱形裝置同主題);戰術戰鬥還擊路徑先前寫死 `hardShield = false`。

⚠ **兩個踩過的錯,都寫進測試**:
- 銀河大小檔位換算 —— 第一版自己編星數門檻,結果「中型」被判成最小檔(星雲數有一半機率是 0)。
  remake 的四個星系大小選項**本身就是那四檔**,直接查表就好。
  徵狀是開局常常一團星雲都沒有,而那看起來完全合理。
- 調色盤鏈 —— 沿用了殖民地的鏈,整團星雲畫成鮮紅色。

兩個都不是讀碼讀出來的,是**加一行 `println` 量出來的**:先量到「畫的時候清單是空的」,
才知道問題在產生不在繪製。

⚠ **移動懲罰沒做**:「降到 1 秒差距/回合」需要一個原本速度的基準,而 remake 的星圖移動沒有
單艦速度模型,硬套就是自己編倍率。Navigator 領袖技能與 Warp Field Interdictor 建築卡在同一個前置。

## ★ 2026-08-07 星圖秒差距模型(gap report 第 16 項(秒差距模型))

第 15 項(星雲)留下的「移動懲罰做不了」,把前置補完。先前星圖移動是 `ETA = ceil(正規化距離 × 8)`,
沒有速度概念,手冊裡四條以「秒差距/回合」表述的規則(星雲、黑洞、Navigator、干擾場)全都無處可掛。

**三個真值把換算釘死**:1 秒差距 = 30 個遊戲單位(`Parsecs_Between_Points_` 裡的 `900 = 30²`,
順帶得知原版的秒差距是整數、無條件進位);四檔銀河尺寸的跳表(506×400 / 759×600 / 1012×800 /
1518×1200);星數門檻 20/36/54/72。三重交叉驗證同時成立——寬恆為 SizeFactor×50.6、高恆為
SizeFactor×40,而**原版存檔 SAVE10.GAM 讀出來就是 759×600 / SizeFactor 15**。

**順帶修掉一個失真**:遊戲提供的星系大小先前是 12/24/36/48(自訂),與原版四檔對不上。
而星雲數、銀河跨距這些表**都是以檔位為索引**的——第 15 項(星雲)那個「開局常常一團星雲都沒有」
就是這麼來的。改成 20/36/54/72。

引擎速度手冊逐條:核融 2 / 融合 3 / 離子 4 / 反物質 5 / 超空間 6 / 相位 7。手冊每條都補了
「引擎完成研究後自動裝到全帝國的船上」——**不是單艦元件**,只看已研究的最高階。

⚠ **又一個畫面上看不出來的坑**:`FleetHasFTL` 對非曲速前開局直接回 true、不看科技表,
於是引擎階查出來是 0 → 航速 0 → **ETA 全被夾成 1**,整個模型形同虛設,而畫面上只是
「每趟都 1 回合到」,看起來像船很快。下界:有 FTL 就至少是核融引擎。

⚠ 未做:黑洞 2 秒差距禁行與干擾場 3 秒差距,常數已入表但**還沒接進派遣判定**——
兩者都卡在同一個前置「路徑經過哪些星」。「穿越星雲」目前近似成起點或終點在星雲內。

## ★ 2026-08-07 星圖航線模型(gap report 第 16 項(秒差距模型))

第 16 項(秒差距模型)留下的三項(黑洞禁行、干擾場、逐段星雲)一次補完——因為**它們其實是同一個問題**:
三條的形狀都是「這條航線離某個東西多近」或「有沒有穿過某塊區域」,一個線段模型全解。

手冊「Ships traveling **through** a nebula」的 through 是重點:**兩端都在雲外、直線穿過去
也算**。第 16 項(秒差距模型)那個「只看起訖點」的近似就是漏在這裡。

幾個刻意的細節:線段不是直線(目的星之外的延長線上有黑洞不該擋路)、起訖點豁免(擋的是路過)、
**干擾場不給 Navigator 豁免**(手冊那句豁免只寫 nebulae and black holes,干擾場是人造的)。

⚠ 星雲判定式改成探針裝進 shell,而它是**未匯出欄位 = 不進存檔**,開新局與讀檔後都要重裝。

**實測可達性**(每種銀河各 12 局,不是斷言是量的):黑洞擋掉 5.7%–13.7% 的目的地,
其餘照走,蟲洞不受限;ETA 在最慢的核融引擎下 1..30 回合,換相位引擎同一趟約 9 回合——
「研究更好的引擎」第一次有了實際意義。

⚠ **第三次踩同一形狀的效能坑**:沿線取樣要問遮罩上百次,而 `decodeAsset` 沒快取。
前兩次是殖民地地表每幀重算、植被尺寸每幀重解。共同成因是 `decodeAsset` 本身無快取、
呼叫端各自為政——下次再遇到就該直接改那一層。

## ★ 2026-08-07 兩種星門(gap report 第 16 項(秒差距模型))

星圖 4 層做掉第 3 層。躍遷門:自己的殖民地之間 +3 秒差距/回合;星際之門:自己的系統之間
一回合到。兩者都是 Achievement 科技——研究到就在自己每個有殖民地的星系各生一個門,
不必逐星建造。

兩個順序上的決定:躍遷門的加成放在懲罰**之前**(星雲與干擾場是「reduced **to** 1」,
是覆寫不是相減,所以它們仍然贏);星際之門放最前面(穩定蟲洞終端、不走實空間,
沿路的懲罰都不適用,與既有蟲洞同語意)。

**這一項能做,是前兩項的收成**:第 16 項(秒差距模型)建秒差距與航速、第 16 項(秒差距模型)建航線——先前沒有
「秒差距/回合」這個量,這兩條規則根本寫不出來。

⚠ 星圖上的標記不是原版畫法:原版是 330 行的逐格動畫,資產來源還沒追出來。先用雙環把
資訊呈現出來——看不出「這顆星有門」,那兩條速度規則等於隱形。

## ★ 2026-08-07 拓殖基地(gap report 第 17 項(拓殖基地))

第 14 項(地表道路)補母星的國會大廈時只治了一半。**編號 11 Colony Base 是相關的另一半**：
指定首都有國會大廈，其餘殖民地有拓殖基地；兩者都是佔一格、有美術、會被畫出的
**實體建築**。國會大廈失守後可在重新指定的首都重建。

資料一直都在(第 11 項(48棟建築盤點)那張差集表對它的註記就是「拓殖時自動」),只是被同一個判斷失誤
連帶漏掉。**「不在建造表」與「不在地表」是兩件事**——同一句話寫第二次,因為同一個誤判
造成了兩個缺口。

護欄除了驗有無,還驗「每個殖民地恰好有國會大廈與拓殖基地其中一棟」,兩者同時出現或
同時缺席都會被抓到。

⚠ task #46 剩下:一次性改造(Gaia/Soil/Terraforming)完成後是否仍佔一格**沒有查證所以沒做**;
對原版實測落點仍未做(需要 archive.org 線上原版逐畫面對照)。

## ★ 2026-08-07 銀河貨幣交易所(gap report 第 18 項(銀河貨幣交易所))

第 11 項(48棟建築盤點)留下的兩個「完全沒有」的建築編號,解掉一個——**因為它根本不是建築**。

手冊寫得很清楚:「Galactic Currency Exchange (**Achievement**) … increases the income
generated by all colonies (from all sources) by 50%」。Achievement 與躍遷門/星際之門同標記,
**研究完成即生效、不必建造**。接在帝國層級的收入乘數上(與政府 money 加成同一層),
因為手冊的字是「all colonies (from all sources)」。

⚠ **為什麼卡了這麼久:自訂的推論規則蓋過了一手來源。** 第 11 項(48棟建築盤點)抽建築表時發現
「維護費 0 = 一次性」這個規律,拿它把編號分堆;18 有成本有維護費 → 判成常駐建築 →
然後「效果是什麼」就查不到了,因為手冊的**建築清單**裡本來就沒有它。它在**科技說明**那一節。
那條啟發式對其餘編號仍成立,錯在把它當成充分條件。

同一個誤判形狀出現三次:Capitol(9)、Colony Base(11)、Currency Exchange(18)——
原版那張 49 筆的表是**通用結構**,不是「可建造建築清單」。

⚠ 剩下 48 Artificial Planet —— **上面這句在下一輪被訂正了,見第 20 項(手冊搜尋假陰性)**:
手冊裡有,搜不到是因為 PDF 用連字排版(`artiﬁcial`)。

## ★ 2026-08-07 AI 主動請求會談(gap report 第 19 項(AI請求會談))

星圖 4 層做掉第 4 層。**卡的不是繪圖,是「誰在請求」這個狀態根本不存在**——remake 的外交
先前只有玩家主動,AI 只會回應。

表示法與版面是真值:原版那支查詢函式整支只有 `mov al, byte_1AB054; retn`(**一個位元遮罩,
每位對手一個 bit**);燈由 x=506 往左排、y=5,兩個都是立即數。

⚠ **觸發條件沒照抄**:原版設 bit 的地方在一支約 30 路跳表的 AI 行動分派函式裡,追出完整
條件成本高收穫有限。改接在既有模型上——**態勢改變時來敲門**,因為 `DecideStance` 的五級裡
有三級本身就是「要跟你講話」的語意。**沒有引入任何新的門檻值。**

**順帶被測試逼出一個設計修正**:第一版把來意寫成中文,被英文模式棘輪測試擋下(缺口 26 條
> 上限 16)。那不只是翻譯問題——**規則層不該吐顯示字串**。改成代碼,顯示文字留 UI 層。

⚠ 燈的圖不是原版的(原版是 per-race 逐格動畫,資產來源沒追);先用「來意色塊 + 一個字」
呈現「誰在敲門、為什麼」。

## ★ 2026-08-07 訂正:「手冊零命中」是假陰性(gap report 第 20 項(手冊搜尋假陰性))

第 18 項(銀河貨幣交易所)結尾寫「48 Artificial Planet 手冊全文搜尋**零命中**(不是漏查)」——**那句是錯的**。
搜不到是因為這本 PDF 用**連字**排版:`artificial` 實際是 `arti` + `ﬁ`(U+FB01)+ `cial`。
改搜小寫 `asteroid` 立刻命中,同一段就把規則講完:

> (Special) … assemble this otherwise useless planetary material into a complete artificial
> planet that can support a colony. This planet is **Barren, Normal G, and mineral Abundant**.
> **Gas giants make Huge worlds, and asteroid belts make Large ones**.

**而且 remake 自己的 `outpost.go` 註解早就寫著同一條規則、同樣的數值**(引手冊 p.50)。
兩個獨立來源一致——又一次「先 grep 自己的 docs」。

真正的阻塞因此從「效果不明」訂正為「**卡在一星一行星模型**」:人造行星按定義是在既有星系裡
**再多**一顆世界,而 remake 的 Stars↔Planets 是一對一,轉換完沒地方放第二個殖民地。
與遷移連線(卡單一艦隊模型)同一類。

**教訓**:「查詢回空」不等於「不存在」。這次的假陰性來源是**排版連字**——而我當時還特地
加註「(不是漏查)」,反而把假陰性寫成了確信。對 PDF 下全文否定判斷前,至少要用小寫、
部分字根再掃一次當正對照。

## ★ 2026-08-07 查證:一次性改造不佔地表格子(gap report 第 21 項(改造不佔格))

第 17 項(拓殖基地)留下「沒有查證所以沒做」的那一項,查證完了——**答案是不需要改**。

**查證方式值得記**:第一版想去找「地形改造完成的那段碼」,但符號表裡 terraform/gaia/soil
一個都沒有。改成**從旗標本身下手**:`grep "136h], 1"` 與 `"136h], 0"`,全檔只有少數幾處,
一眼看到建築完工結算那一處。**找不到名字時,找它必然會碰到的資料。**

結論是定性的:那支函式裡「記旗標」這一步有條件,而條件變數**恰好被清成 0 四次**,
四個分支做的事一看就認得(改氣候、Terran→Gaia、改礦產、寫入整組行星欄位)——
正好對應四個一次性編號 17/37/44/48。旗標沒被設,地表迴圈就不會擺它們。

remake 天然就是對的(SpecialActions 不在建築表裡),加測試把它釘住:哪天有人「順手」把
一次性項目加進建築表,地表會冒出四棟原版沒有的房子,而那看起來完全合理。

**task #46 到此結案**,只剩「對原版實測落點」——那是驗證工作不是實作工作,另計。

## ★ 2026-08-07 黑洞的旋渦動畫(gap report 第 22 項(黑洞動畫))

星圖的黑洞從第 9 項(星圖艦隊圖示)起圖就是對的,但它一直是靜止的。原版會轉。

`Draw_Black_Holes_` 的推進規則整段可讀:計數 %(幀數 × 2),再除以 2 —— **每一幀停留 2 次重畫**,
每個黑洞各有獨立計數器。那個「除以 2」在一般星球的 `Draw_A_Star_` 裡也獨立出現一次
(`sar eax, 1`),所以是兩個來源不是一個。資產面也對得上:`lbxinfo` 給的是黑洞那組 16 幀、
一般星球 5 幀,dump 出來 16 張逐張比對**全不同**,旋渦是真的在轉。

**一般星球刻意不做**:它的閃爍是爆發式的,「何時開始閃」「爆發長度」「全域併發預算」三個常數
都沒追出來。**不編那三個數**。這一項的分界線不是難度,是「規則有沒有解完」——
黑洞的動畫無條件連續,規則完整,所以做;星球的不完整,所以不做,而且把「不做」寫成測試護欄。

**⚠ 只有比例是真值**:remake 把「一次重畫」對應成「一個 ebiten 幀」,而原版的重畫頻率沒解出來,
所以動畫的絕對速度是 remake 的選擇。

順手修掉一個會**靜悄悄壞掉**的東西:星圖 sprite 的快取 key 原本不帶幀號(因為以前只解第 0 幀),
一加動畫就會讓 16 幀全部命中同一張——**畫面完全正常,只是不會動**。這已經是第四個自己長快取的
呼叫端了,根因是 `decodeAsset` 沒有快取;下次再遇到直接修它,不要再加第五層。

## ★ 2026-08-07 訂正:兩個「動畫沒做」其實是原版就不會動(gap report 第 22 項(黑洞動畫))

做完黑洞動畫後去追「艦隊圖示 8 幀為什麼沒動」,查完發現**要改的是文件不是程式**。

先挖到引擎層的規則:原版的通用貼圖器 `sub_12A478` **畫完會自動把幀號 +1**,
所以呼叫端要靜止就得每次先歸零(`sub_12B726`)、要自訂節奏就每次寫死幀號(`sub_12B753`)。
這一條同時解掉兩個問題:

1. **艦隊圖示**:`Draw_Ship_Icons_` 每次繪製前都歸零 → 恆為第 0 幀。而檔頭原本寫的
   「`Cycle_Ship_Icons_` 在跑動畫」也是錯的——那支由鍵盤跳表叫進來、`bx` 是方向,
   是「切換到上/下一支艦隊」。手冊逐字對上:F1 / F2。**兩個獨立來源。**

2. **一般星球閃爍**:第 22 項(黑洞動畫)寫「三個常數沒追出來所以不編」,查證後可以說得更強——
   **出貨版根本不會閃**。啟動它要把 `star[+0x64]` 設成 ≥ 0,而全檔對那欄位的位元組寫入
   只有 reset(0xFF);全域預算 `word_19C164` 更是**只減不加**,不可能是還在運作的閘門。

**正對照做了**(上一輪才因為 PDF 連字把假陰性寫成事實):同樣的搜法去找星球結構 `+0x16`
光譜欄位的寫入端,找得到——**方法會命中,所以這次的零命中是真的零**。
順帶交叉驗證:reset 迴圈跑 `0x48 = 72` 次 = 最大銀河星數,兩邊獨立落在同一個數字。

**副產品:手冊的快捷鍵表**(F1/F2 循環艦隊、F5/F6 切換已殖民星系、F9 測距、F10 快速存檔、
ALT+F9 載入)。其中 **F9 測距最該做**——秒差距模型第 16 項(秒差距模型)就建好了,但玩家在畫面上
看不到任何秒差距數字。手冊另有一組 ALT+F1..F8 設定開關,但那些鍵在 PDF 裡是右側邊欄標籤,
抽出來會排到前一個選項的尾巴,**對應有 off-by-one 風險,所以不寫進表**。

## ★ 2026-08-07 手冊的星圖快捷鍵接上(gap report 第 22 項(黑洞動畫))

第 22 項(黑洞動畫)掃出手冊的快捷鍵表,這一輪把**行文中直接寫死**的那幾個接上:
F1/F2 循環艦隊、F5/F6 切換已殖民星系、**F9 測距**。(邊欄標籤的 ALT+Fn 那組仍不碰。)

**F9 最有價值**:秒差距模型第 16 項(秒差距模型)就建好了(1 秒差距 = 30 遊戲單位、距離取整,引擎速度、
星雲減速、干擾器範圍全掛在上面),**但玩家在畫面上看不到任何秒差距數字**——整套模型是隱形的。
手冊描述的行為是兩段式而且**跟著游標即時更新**:按 F9 → 點第一顆 → 移到哪顆就顯示到哪顆。
截圖驗到 15 秒差距,中型銀河是 25.3 × 20 秒差距、對角約 32,量級對。

**F1/F2 目前只有一個元素**,而那是資料模型的事:remake 的玩家艦隊是單一集合,AI 只有抽象
戰力、在星圖上沒有位置。同一個缺口也卡著遷移連線層。已把限制釘成測試,多艦隊做出來時它會紅。

**同日補完 F10 / ALT+F9**:F10 的「上次的存檔名」就是 `savePath`(開局是自動存檔那一格,
從載入視窗讀過某一格之後改成那一格),**語意天然對上不必另建概念**。覆蓋是原版行為所以不加
確認框,但一定要回報——沒有回報的話按下去成功與失敗看起來完全一樣;既有的 `lastActionMsg`
畫在選中星面板裡、沒選星就看不到,所以另加一個約 3 秒會自己消失的短暫訊息(**會消失是刻意的**,
一直掛著的「已存檔」會被誤讀成「還在存」)。ALT 組合要先於單鍵表判定,否則 ALT+F9 會被當成 F9。

**兩個「看起來完全正常」的坑**:① 提示字釘死在 (30,34),截圖一看正好蓋掉左上角那顆星的名字
→ 改成跟著游標走(星圖每個角落都可能有星,沒有安全的固定位置);② 截圖廊的示範終點寫死索引,
而那顆還沒探索,`starAtScreen` 跳過不可見的星 → 截圖停在提示上什麼都沒畫 → 改成執行時挑一顆
可見的。順手把點擊熱區與懸停判定收斂到同一個 `starHitHalf`,免得出現「點得到卻懸停不到」。

## ★ 2026-08-07 多艦隊模型(gap report 第 23 項(多艦隊模型),第一階段)

把「全帝國只有一支艦隊」這個限制拆掉:`Ships + FleetAtStar/DestStar/ETA/Marines/Tanks`
一組欄位 → `Fleets []Fleet` + `SelectedFleet`。

**難點不是欄位改名,是 `Ships` 有兩種語意**:「這支艦隊的船」(戰鬥、載運、消耗殖民船)與
「全帝國的船」(指揮點數手冊 p.169 明文、國力、艦名編號、外交評估、艦隊列表)。
單一艦隊時兩者剛好相同,所以分錯了也看不出來——**盲改會讓第二類在真的有第二支艦隊時默默算少,
而那時候看起來完全正常,數字只是偏小**。逐處分類(非測試碼約 65 處),並用 `fleet_test.go`
的兩支艦隊測試把分類釘住。

**順帶修正三個行為**:① 修復先前只看選中那一支艦隊有沒有停靠據點,改成逐艦隊各自判定
(原版的迴圈也是逐艦隊走的);② 母星防禦同樣只看選中那一支,於是玩家把視角切到別支艦隊、
母星就「沒有防禦」——**那是操作副作用不該影響世界狀態**;③ 隨機事件的「損失一艘艦」
打的是整個帝國,改用跨艦隊索引。

**存檔遷移順帶發現舊格式一個漏欄**:舊格式有 `fleetMarines` 卻**沒有 `fleetTanks`**,
舊存檔讀回來戰車營一律歸零。那是舊格式本身的洞,新格式序列化整個 Fleet 已補上。
判斷舊檔用「Fleets 欄位在不在」而不是版本號——版本號會被別的改動一起往上帶。

**驗收**:重構不改變行為,所以驗收是「畫面要一模一樣」——重跑截圖廊 29 張,
**28 張逐位元相同**,唯一不同的載入視窗差在存檔時間戳。

**第二階段還沒做**:分/合艦隊的 UI、逐殖民地造艦 + 集結點(那才畫得出遷移連線層)、
AI 艦隊的星圖位置(F1/F2 目前仍只走玩家自己的艦隊)。

## ★ 2026-08-07 遷移連線——星圖 4 層裡的最後一層(gap report 第 23 項(多艦隊模型))

多艦隊(第 23 項(多艦隊模型))做完之後,`Draw_Relocation_Links_` 缺的只剩自己的資料。
兩支函式讀同一個欄位互相印證:`word[星×0x71 + 0x54 + 玩家×2]` = 遷移目標星,−1 = 沒設定。
手冊說新造的艦會 "automatically relocated" —— 那是**一段航程**不是瞬間移動,
所以 remake 建一支往目的地航行的艦隊,星圖上那條線畫的就是它。

**顏色是真值,而且它很暗——那是原版的樣子。** 原版丟給畫線函式的是 8 個調色盤索引
(`dword_81C80` = 6E 6F 70 70 ×2),解出來是 (0,20,0)/(4,56,4)/(0,76,0) 深綠。
手冊自己說「If you'd rather not clutter up the galaxy with them, turn this option off」
——它的定位是**可以關掉的雜訊**。所以不調亮;驗證用截圖加亮去看,
不是為了看得清楚去改一個已經是真值的數字。

**兩個坑**:① 反鋸齒把本來就很暗的線和黑底又混一次,幾乎消失——原版是硬邊像素,關掉才對;
② 截圖廊的示範目標和 F9 測距撞在一起,遷移連線整條藏在測距線底下,看起來像沒做。

**唯一的零值陷阱而且很致命**:`ColonyRelocateTo` 的 Go 零值 0 就是**母星的索引**,
補齊平行陣列時填零值 = 每個新殖民地一建好就把新艦全往母星送,而那看起來完全像遊戲規則。

**順手修掉一個假護欄**:`hotseat.go` 寫著「`TestSeatFieldsCoverPlayerSide` 用反射盯著它」,
而那支測試根本不存在——**指名了不存在的護欄比沒有註解更危險**,它讓人以為這裡有人在看。
改成真的寫一支反射往返測試。

## ★ 2026-08-07 艦隊列表列艦隊 + 清 HONEST-STATUS 過期斷言(gap report 第 23 項(多艦隊模型))

畫面標題是 FLEET OPERATIONS,而 remake 把名冊攤平成一長串船名——那是單艦隊時代的殘留:
**全帝國只有一支艦隊時,「列船」與「列艦隊」看起來一樣**。改成逐艦隊分組(標頭可點擊切換
操作中的艦隊,航行中顯示目的地與回合數)。

⚠ 原版這畫面的美術上就烘著 **RELOCATE**(remake 譯「調動」),而手冊說集結點是在
「Fleet Operations console」設的——**忠實入口就是這裡**。但那顆鈕按下去原版做什麼沒有
反組譯確認,先不接,留星圖那條路能用。

**清掉 HONEST-STATUS 三條過期斷言**:①「行星表面 + 建築擺放子系統仍未做」——同一份文件裡
自己打架,稍晚的段落就寫著全做完了;②「建築集合仍與原版有差」——Colony Base 第 17 項(拓殖基地)補上、
一次性改造第 21 項(改造不佔格)查證後確認不該佔格;③「戰機/航母需先建基礎設施」——說得太重,戰機已接進
快速艦隊戰鬥,真正缺的是戰術格子裡的獨立戰機單位。

**同一份文件前後矛盾這件事本身值得記**:那三條寫在「誠實現況評估」裡,而它正是外部判斷
「還缺什麼」的依據——過期的缺口清單會讓人去做已經做完的事。

## ★ 2026-08-07 RELOCATE 的原版語意(gap report 第 23 項(多艦隊模型))

第 23 項(多艦隊模型)留下「那顆鈕按下去原版做什麼沒有反組譯確認」。查符號表,整組都在,問題直接解掉。

**流程是兩段點選**:先點起點星(必須是自己的殖民地)、再點終點星,**點回自己 = 取消**。
remake 先前的「星圖面板 → 點一顆星」其實是第二段,第一段被略過(用面板選中的那顆當起點)
——合理的捷徑但不是原版入口。現在兩條路都在:艦隊列表的 RELOCATE 走完整兩段(手冊逐字說
集結點是在 Fleet Operations console 設的),星圖面板走捷徑,規則面共用同一支函式。

**四條合法性規則**(反組譯逐條):黑洞起訖都不行;沒探索過的星不行;目的星上有艦隊要跳確認框
(當起點則直接不行);起點必須是自己有殖民地的星。⚠ 確認框那條**沒做**——remake 沒有 modal
對話框的基礎設施,直接允許,是已知的簡化不是漏看。

**順帶記兩個還沒接的**:`Set_All_Star_Relocations_` 與 `Clear_All_Star_Relocations_`
——艦隊列表的 ALL 鈕多半就是它們(一次設定/清除所有殖民地的集結點),尚未確認也尚未接。

## ★ 2026-08-07 分艦隊(gap report 第 23 項(多艦隊模型),task #50 收尾)

原版沒有「艦隊」這個型別,有的是 **ship stack**:`word_192248[stack]` 是頭一艘船 id、
`word_1975D6[船 id×5]` 是「下一艘」的單向串列。`Split_Stack_` 收一組船 id,
把它們摘下來串成新的一個 stack 接在表尾。語意是「選一組船抽出來組成新艦隊,位置不變」。

**擋下三種退化情形**,其中一條是 remake 特有的:「全選」不是拆分(會留下一支空的舊艦隊);
索引越界;**航行中不能拆**——remake 的航行是整段跳的、中途沒有位置,拆出來的那一半沒地方放。
**那是 remake 移動模型的後果不是原版規則**(原版的 stack 隨時有座標)。

**已知簡化**:陸戰隊/戰車營全部留在原艦隊——remake 把它們建模成艦隊層級的數字,
不綁定到特定的船,拆分時沒有「哪些跟著走」的依據。寫成測試釘住。

**⚠ 這個 UI 是 remake 自己加的**:原版艦隊列表的美術上沒有 SPLIT 鈕
(烘著的是 ALL / RELOCATE / SCRAP / LEADERS / Support / Combat / RETURN),
原版是在右側艦艇格選船再下令。remake 的右側格還沒接上選取,先用左側名冊勾選 + 一行文字當入口。

**版面的坑**:第一版把拆分那行放在船清單之後,而名冊往下長——結果和固定在 y=402 的
「攻打安塔蘭母星」疊在一起。改放在艦隊標頭底下,名冊再長也不會撞到底部。

## ★ 2026-08-07 一星多行星:軌道模型資料層(gap report 第 24 項(軌道資料層),第一階段)

「每個星系 5 個軌道」是真值,三個獨立來源:偏移算術(0x54 − 0x4A = 10 bytes = 5 words)、
`System_Planet_Scanned_To_Planet_Id_` 的索引式 `word[星×0x71 + 0x4A + 軌道×2]`、
以及走訪迴圈寫死的上界 `cmp ..., 5; jge`。行星是獨立的一張表(每筆 0x11 = 17 bytes),
`Planet_Orbit_` 讀 `byte[行星 id×0x11 + 3]` = 軌道號——**雙向都有指標**。

**意外發現**:`genPlanets` 早就在骰整個星系了(`RollNumSatellites` + 逐軌道
`RollSatelliteType`),然後挑一顆代表行星,其餘只存成 `SystemBodies` 摘要。
所以缺的不是骰表,是「**其他天體是二等公民**」這個表示法。

**這一階段換的是形狀不是內容**:`Star.Orbits [5]int` + 存取器(`PlanetAt` / `PlanetsAt` /
`PlanetStar` / `PlanetOrbit` / `FreeOrbit`)+ 存檔遷移。行為逐位元不變,兩條測試釘住:
每顆星恰好一個軌道有行星(刻意的限制,下一階段會讓它紅)、`PlanetAt(i)` 必須等於 `i`
(相容性支點,否則舊呼叫端換過來會位移)。

**⚠ 又一次同款零值陷阱,第三次**:軌道表的 Go 零值是 5 個 0,而 0 是行星 0 的索引——
不修的話每顆星都宣稱軌道 0 上有行星 0,**而且不會報錯**,只會讓每顆星的行星資料看起來都一樣。
與 `Star.Wormhole`(零值 0 → 每顆星都連到星 0)、`ColonyRelocateTo`(零值 0 = 母星)
同一個形狀。共同點:**索引型欄位的「沒有」必須是 −1,不能靠零值**。

## ★ 2026-08-07 一星多行星 Step A:呼叫端改走存取器(gap report 第 24 項(軌道資料層))

第 24 項(軌道資料層)建好軌道模型,但所有讀行星的地方仍直接寫 `s.Planets[星]` ——那個式子**假設
Planets 與 Stars 平行**。一旦產生器填滿軌道,`len(Planets)` 就大於 `len(Stars)`,
那些式子會**默默讀到錯的行星**:不會崩、不會報錯,只是資料錯位。

所以先做一步行為不變的遷移(`PlanetAt` / `PlanetOf` 可寫指標 / `PlanetDataAt` 唯讀複本)。
**代表行星的挑法必須與產生器逐字相同**(先找一般行星,整組不宜居才退取第一個天體)
——不一致就會位移,而位移的徵狀是「殖民地總覽說類地、殖民地畫面說凍原」那一類自打嘴巴,
不是崩潰。順手拿掉幾個 `star < len(Planets)` 的邊界檢查:那是平行假設的殘留,
Step B 之後會變成錯的(而且是放行太多)。

**驗收**:重跑截圖廊 **28/29 張逐位元相同**,唯一不同的載入視窗差在存檔時間戳。

**Step B 還沒做**:把 SystemBodies 升格成完整行星、填滿軌道表。⚠ 那會改變同一 seed 的
星系內容(多骰了那些天體的氣候/礦產/重力),要給它們獨立的亂數流——這個專案已經用過這招
(「行星生成用獨立亂數流,不讓抽取次數影響佈局」),否則之後的每一顆星都會漂掉。

## ★ 2026-08-07 一星多行星 Step B:同系天體升格(gap report 第 24 項(軌道資料層),task #51 收尾)

同一顆恆星底下的每一個天體都是完整的 `Planet` 條目,各佔一條軌道。`Planets` 因此
**不再與 Stars 平行**(24 顆星 → 94 顆行星)。非代表天體用**獨立亂數流**——共用一條的話,
第 0 顆星多骰的那幾次會把之後**每一顆**星的代表行星換掉(這個專案已為 genPlanets /
genMonsters / genWormholes 各開一條,同一個理由)。

**測試抓到三個還在假設平行的地方,兩個在產品碼**:`monster.go` 的 `planets[starIdx]`
——怪獸的特殊物產會補到**別的星系**的行星上。這正是第 24 項(軌道資料層)先做「改走存取器」的用意:
**那些式子不會崩,只會讀到錯的行星**。抓到它們的不是編譯器(索引式型別上合法),
是跑起來的不變量測試(「母星一定宜居」「有怪獸的星系一定有特殊物產」)。

**`SystemBodies` 淘汰**——它的原註解寫著「這裡不重複放代表行星本身,避免兩份資料要同步的
老問題」,它知道自己是折衷。現在摘要文字改從軌道表算,只有一份資料。

**三條測試換掉**,因為它們釘的是階段性限制(註解當初就寫「升格之後這條會紅,那時候該改的是
測試」)。換成更強的:軌道條目要指到有效且不重複的行星、每顆行星都掛在某個軌道上、
銀河裡**真的有**多天體星系(沒這條的話「升格」可能只是搬了位置)、代表行星的挑法。

**解鎖**:`FreeOrbit` 現在真的有意義——人造行星可以往空軌道放,同星系多殖民地也有了
資料基礎。兩者的規則接線仍未做。

## ★ 2026-08-07 人造行星:手冊推翻了 remake 自己的假設(gap report 第 25 項(人造行星))

gap report 第 20/24 項寫了兩輪的那句「人造行星按定義是**在既有星系裡再多一顆世界**」——
**錯的**。手冊逐字:「assemble this otherwise useless planetary material into a complete
artificial planet」——它是把**既有的**氣態巨星或小行星帶組裝成行星,那顆天體本來就佔著軌道。
所以前置是「同星系有材料」不是「有空軌道」。測試釘住這個訂正:**五個軌道全滿但有氣態巨星
→ 可以蓋;有空軌道但沒有材料 → 蓋不了**。

那句斷言是從「人造行星」這個**名字**推的,推得很合理,而且它擋了兩輪工作。
又一個「先查一手資料再推論」的實例。

**反組譯逐項吻合**:`sub_13FD9` 走兩趟掃 5 條軌道——第一趟找氣態巨星(型別 2)→ 尺寸 4、
第二趟才找小行星帶(型別 1)→ 尺寸 3。**氣態巨星優先**,而 4/3 正是 Huge/Large,
與手冊「Gas giants make Huge worlds, and asteroid belts make Large ones」逐字對上。
⚠ 兩趟不能合成一趟:合起來的話軌道較內的小行星帶會搶在外側的氣態巨星前面。

**成本用真值不用估值**:第一版順手寫了 900,而第 11 項(48棟建築盤點)抽出來的原版建築表是 **800**。
專案裡已經有真值就不該再估一個。

## ★ 2026-08-07 Set_All 集結點 + 遷移連線顯示開關(gap report 第 23 項(多艦隊模型))

**`Set_All_Star_Relocations_` 有一個猜不到的細節**:它的迴圈裡有一道 `!= −1` 檢查,
**只改已經有集結點的殖民地**,沒設過的不會被順便設上。直覺會做成「全部設成這顆」——
而那會讓玩家按一下 ALL 就把**所有**新殖民地的產出全部抽走,包括他本來想留在原地生產的。
這是從按鈕名字("ALL")推不出來的規則,測試釘住。

`Clear_All` 規則已實作並測試,但**沒有 UI 入口**——原版哪顆鈕對應它沒有確認,不猜。

**遷移連線的顯示開關有地方放了**:遊戲選單上那顆 SETTINGS 鈕本來就是死的(檔頭寫著
「按鈕保留但不接」),現在它展開一列設定,目前只有這一項。⚠ 那一列不是原版版面
(原版有一整個設定畫面),建了那個畫面之後要搬過去——但「點了沒反應的鈕」與
「展開唯一一個真開關的鈕」相比,後者誠實得多。

## ★ 2026-08-07 同星系多殖民地:拓殖的對象是行星不是星(gap report 第 24 項(軌道資料層))

`ColonizeStar` 的閘寫著「該星已有歸屬,不可拓殖」。一星一行星的時代那是對的
(一顆星只有一顆行星,有歸屬就等於沒空位);軌道模型上線之後它變成
「**你自己的星系不准再殖民**」——而手冊 p.61 從頭到尾寫的是 any uncolonized **planet**。

**換掉的東西**:`ColonizePlanet(planet)` 成為入口(`ColonizeStar` 降級成「該星系第一顆
可殖民行星」的捷徑);殖民地多記一個 `PlayerColonyPlanets[i]`;前哨站多記
`Outpost.PlanetIndex`(手冊 p.119:build an outpost **on a single planet**);
殖民地名與地表變體都改用行星索引——否則同星系的兩個殖民地會同名、地表長得一模一樣。

**第四次踩到同一個零值陷阱**:索引型欄位的「未知」必須是 −1。前三次是 `Star.Wormhole`
(全部連到星 0)、`ColonyRelocateTo`(全部指向母星)、`Star.Orbits`(每顆星都宣稱有行星 0)。

**順帶修好一個 bug**:`consumeOutpostForColony` 原本只比對星,所以在一顆行星建殖民地會把
同星系另一顆氣態巨星上的前哨站吃掉、還白送一座海軍陸戰隊營。

**選行星的 UI 不必新造**:原版行星列表(`PLNTSUM.LBX`)右下角本來就烘著
SEND COLONY SHIP / SEND OUTPOST SHIP,那就是原版選行星的地方。先前那個畫面是唯讀展示
(而且列的是 `Planets[0..7]`,與星系無關)。現在列出看得見的星系的所有天體、可點選、
兩顆鈕對選中的行星作用,艦隊不在那個星系就先派過去。

AI 拓殖的候選集是「無主星 + 自己已有殖民地的星系」(`aiExpansionCandidates`,第 24 項(軌道資料層)),同星系可有多個殖民地。
資料模型(`AIOpponent.ColonyPlanets`)已補齊,規則沒動——記在 gap report。

## ★ 2026-08-07 AI 也會在自己的星系裡拓殖(gap report 第 24 項(軌道資料層))

第 24 項(軌道資料層)只改了玩家側,`aiExpand` 的候選集還寫死「只找 `Owner == 0` 的星」——玩家可以在自己的
星系裡塞滿殖民地,AI 一個星系永遠只有一個。那不是原版的規則差異,是 remake 改了一半的不對稱。

**`Star.Owner` 分不出是哪一個 AI**(只有 無主/玩家/AI 三個值),所以「這是不是**我自己**的
星系」要走各自的 `ColonyStars` 清單判斷,不能只看 `Owner`。

**兩個會被灌水的計數器**:`OwnedStars` 只在「本來無主」時才加(同星系多殖民不會讓版圖變大);
`PlanetColonized` 改成全帝國視角,否則 AI 會把殖民地疊到別人已經佔著的行星上。

**順帶修好入侵的一個 bug**:`InvadeColony` 打贏就無條件把星判給玩家。同星系多殖民地之後,
打下其中一個殖民地會讓剩下那個敵方殖民地變成「站在玩家星系裡的敵軍」。現在星的歸屬只在
該 AI 在這顆星上再也沒有殖民地時才翻面,過戶的殖民地也改用 AI 記下的真實行星索引。

## ★ 2026-08-07 是/否確認框 + 一條寫錯的規則(gap report 第 26 項(確認框))

**先訂正**:`relocation.go` 檔頭與 gap report 都寫著「目的星上**有艦隊**時會跳確認框」。
逐指令讀 `Okay_To_Set_Relocate_Star_` 之後,那個條件是 `Star_Guarded_By_Monster_`——
**是怪獸不是艦隊**,而且符號表裡本來就有名字。錯誤斷言的來源大概是「集結點會把新艦送過去,
所以危險的是目的地有敵艦」:一個合理但沒查證的推論。

原版對起點的怪獸是**靜默拒絕**;remake 出一句話——在有滑鼠提示的介面裡,靜默失敗只會讓
玩家以為按鈕壞了。

**確認框的版面全部是立即數**:底框 `CONFIRM.LBX#0` 貼在 (161,117);Y/N 兩顆鈕貼在
(235,302)/(345,302),熱區 51×21(比圖小一圈,左上角重合);文字左緣 204、折行寬 224、
垂直置中於 208。兩個交叉驗證:熱區與圖的左上角重合、文字塊中心 316 ≈ 底框中心 317.5。

**沒有還原的一條**:原版文字放不下時會縮字級(量高度直到 ≤126)。remake 的字型層沒有那組
原版字級,改用固定字級 + 自行折行,寫在檔頭不假裝有。

## ★ 2026-08-07 戰術格子的獨立戰機單位 + 一個讀錯的欄位(gap report 第 27 項(獨立戰機單位))

**先訂正**:`gamedata/combat.go` 的「中隊規模:攔截機 4、重戰機 2(手冊 GM p.127 出擊數欄)」——
那一欄是 **Shots**(每架返航前開幾次火),不是中隊人數。中隊規模在正文,而且寫了兩次,一律 **4 架**
(p.157「squadrons of four」、p.83「squadrons of 4」)。Shots 欄也有正文對照:攔截機「fire 4 times」、
重戰機「一彈一光束 ×2」。**舊值把「一架打幾次」當成了「一隊有幾架」,重戰機庫少算了一半的戰機。**

順帶確認速度/血量欄沒錯:表上攔截機 Speed 8-20 而正文說 10,套 `CombatFighterSpeed` 就通了——
範圍下限是 FTL 0(base−2)、上限是 FTL 6(base+10)。血量欄下限正是正文的「can take 2 / 5 damage」。

**戰機從「一個加成」變成「一個兵種」**:中隊在格子上有自己的位置,出擊 → 飛向目標 →
貼身開火(不像艦砲有射程)→ 彈盡返航 → 回到母艦補血補彈(**但不補人**)→ 可再出擊。
血量是每架的,傷害一架一架吃;只有攔截機能纏鬥;貼身的敵艦會把戰機打下來。

**誠實留白**:戰機自動尋找最弱護盾分面(一般艦艇命中鏈已有四面容量,但戰機的最弱面選擇尚未接入)、
轟炸機/突擊梭(各自依賴另一套系統)、敵方不派戰機(敵艦沒有設計資料)、
FTL 階與裝甲級先傳 1/0、出擊鈕不是原版版面。

## ★ 2026-08-07 ALL 鈕根本不是集結點(gap report 第 23 項(多艦隊模型))

第 23 項(多艦隊模型)把艦隊列表的 **ALL** 鈕接上 `Set_All_Star_Relocations_`,並且自己標了「推測」。
手冊在兩個地方各講了一次它到底是什麼:p.32「To select or deselect all of the ships in the
window, you can use the All button」、p.47「All: Selects all of the ships in the fleet …
(If all the ships are already selected, this deselects them instead)」。
括號那句是 **toggle** 語意。p.47 同時給出那三顆鈕的完整清單:**All / Relocate / Scrap**。

**那 Set_All / Clear_All 從哪裡進來?** 星圖的輸入處理器 `sub_73980`,而且是**鍵盤事件**:
`−1105 → Clear_All`、`−1005 → 切換「下一次點星要 Set_All」模式`。同一支函式裡 −1002/−1001
是與滑鼠 widget id 併列判斷的替代鍵,可見那組負數 id 就是鍵盤來的;兩者差 100,
看起來是「某鍵」與「ALT+同一鍵」。**是哪一顆鍵沒有確認,不猜**(同第 22 項(黑洞動畫)的立場)。

**落地**:ALL → 全選/全不選(選取狀態本來就有,分艦隊用的就是它,所以「全選 → 拆分」
兩下就做得完);Set_All / Clear_All → 名冊下方兩個**明確標示為 remake 自加**的入口
(字前加「＋」),追出鍵碼之後改成星圖快捷鍵。

**一個沒有照手冊改的地方**:p.47 說 Relocate 的終點要「click on another system you've a
colony in」,但 `Okay_To_Set_Relocate_Star_` 對終點沒有這條檢查(那條只在起點分支)。
程式碼是實際行為,手冊那句更像是描述常見用法——不改規則,記下來。

## ★ 2026-08-07 「AI 的遷移設定」不是缺口(gap report 第 28 項(AI遷移非缺口))

表上寫著「AI 沒有逐星的艦隊位置,所以沒有遷移可設」。前半是對的
(`AIOpponent.FleetStrength` 是抽象數字),但後半的結論要先確認**原版的 AI 有沒有在用這個欄位**。

逐一追過那個欄位的**五個寫入者**(`Universe_Generation_` 初始化成 −1、`Set_Relocation_`、
`Clear_Star_Relocation_`、`Set_All_Star_Relocations_`、`Clear_All_Star_Relocations_`),
呼叫端全部是玩家的星圖/遷移 UI,**沒有一個在 AI 的程式碼裡**。讀取端
`Redirect_Newly_Built_Ships_` 確實逐玩家跑,所以 AI 的欄位有人讀——只是永遠是 −1。
欄位逐玩家是因為星球結構替 8 個玩家各留一格:**多人對戰時每個人類玩家用自己那格**。

**結論:原版的 AI 也不設集結點。** remake 什麼都不用做;要替 AI 加會是加一條原版沒有的規則。

**方法上的坑**:第一次用 `grep '*2+54h]'` 找寫入者**漏掉了兩個**——`Set_All`/`Clear_All`
把 `星基 + 玩家×2` 先加好再 `mov [eax+54h], bx`,定址式裡沒有 `*2`。正確做法是先切出每一支
函式,再找「同時碰星表基址、`71h` stride 與 `+54h]`」的那些。那組條件把漏網的撈了回來,
也就是這個「不存在」結論的正對照。

## ★ 2026-08-07 決定性化 + 兩個存檔 bug(gap report 第 29 項(決定性化))

網路多人剩三塊:9 個畫面、傳輸層、**決定性化**。決定性是規則層自己的性質,現在就測得了,
而且要先測——等傳輸層上線才發現規則層本身不決定性,每一次分岔都要先排除是不是網路問題。

**狀態指紋**用存檔快照當正規形式(`SHA-256(json.Marshal(snapshot()))`):新欄位只要進得了存檔
就自動進得了指紋;`encoding/json` 對 map 的鍵保證排序,map 迭代順序不會造成假分岔;
指紋不合時直接 diff 兩邊的 JSON。

**閘門抓到的第一個 bug**:三條長壽命亂數流(事件/發現/間諜)只記種子不記「抽到第幾個數」,
讀檔後整條流從頭開始——存檔洗事件毫無成本。修法有個坑:`math/rand` 的 `Intn` 與 `Float64`
**從底層取走的數量不一樣**,所以「重抽 n 次跳過去」必須連抽的種類都一樣。改成直接騎在
`rand.Source64` 上,每次抽取恰好消耗一個 uint64,快轉就只是丟掉 n 個原始值。

**第二個 bug**:修完亂數流閘門**還是紅的**。diff 存檔 JSON → `RuleProfile` 完全沒進存檔,
讀檔後是零值(既不是 1.3 也不是 1.5,而是「Version=1.3 但數值欄位全 0」的混種)。
**主選單選的版本撐不過一次存讀檔**,而那是 CLAUDE.md 列的專案目標之一。
改成存版本、讀檔時重建完整 profile。

**形狀值得記下來**:兩個 bug 都不是讀程式碼找到的,是閘門紅了之後 diff 出來的。
第二個尤其——修完第一個之後如果沒有那支測試,會直接以為做完了。

## ★ 2026-08-07 傳輸層 + 鎖步協定(gap report 第 29 項(決定性化))

`Net_Next_Turn_` @ 0xFC470 的骨架給了原版的形狀:**鎖步**——各自下完指令 → 廣播「我好了」
(逐玩家旗標陣列 `byte_1AAF7E`)→ `Wait_Until_Net_Opponent_Finished_` 等全部到齊才推進。
remake 照這個形狀做,但傳輸換成 **TCP + JSON**(原版走 IPX / 數據機 / 序列埠,現在的機器
三種都沒有)——**移植決策不是還原**,標在檔頭。

**三個設計決定**:①`internal/netplay` 不相依 `internal/shell`(傳輸層不該知道規則;
端到端測試放外部測試套件,同時 import 兩邊而不讓生產碼耦合);②4 位元組長度前綴
(TCP 沒有訊息邊界),上限 4 MB 是為了**不讓對面一個壞掉的長度欄位要求我們配置 4 GB**;
③指令**依玩家編號**排序——依到達順序套用會讓「誰的封包先到」影響結果,那是鎖步最典型的分岔來源。

**逐回合比對狀態指紋**是原版沒有的一層:分岔一旦發生,不比對的話幾十回合後才會以
「你的畫面跟我不一樣」爆出來,那時候回推不了是哪一步歪的。

**端到端測試**:兩個對等端各跑一份 GameSession,over `net.Pipe()` 交換 25 回合,
每一回合指紋都必須相同(含正對照:整場至少要有一條指令真的生效)。另一支走真的 TCP socket。

**還沒做**:9 個網路畫面(版面座標未抽)+ 指令解譯器(把每一顆按鈕對到一條 Command;
端到端測試只解了三條,夠證明鏈通了但不是完整指令集)。

## ★ 2026-08-07 玩家指令層(gap report 第 29 項(決定性化))

把「玩家按了哪顆鈕」變成一筆可序列化、可重播的資料。三個用途,只有一個是網路:
網路對戰(兩台機器要套用同樣的指令序列)、回放/除錯(指令序列 + 起始種子 = 可重現的 bug 報告)、
熱座(其實已經在做同一件事,只是沒有序列化)。

**兩條刻意的界線**:①指令層**不做前置檢查**——方法自己會回絕不合法的操作,再檢查一次
只會變成兩份會漂開的規則;②**不認得的指令名一律報錯,絕不靜默忽略**——靜默忽略在鎖步裡
是最糟的處理(一邊套用了、另一邊沒有,而且沒有人會知道)。參數不足則走預設值,
因為那是送出端的 bug,而規則層自己會擋掉無效操作。

**型別分開、形狀一樣**:`shell.PlayerCommand` 與 `netplay.Command` 欄位一致但是兩個型別,
傳輸層不 import 規則層;轉換發生在同時認識兩層的組裝端。

**23 條指令 + 三支一致性測試**:表上有的都認得(清單排序、無重複)、走指令層與直接呼叫方法
狀態必須一模一樣(逐條比 StateHash)、每一條都要能過線再被規則層認得(兩份清單漂掉會出現
「單機做得到、連線不同步」的靜默缺口)。端到端測試因此換掉了原本只解三條的迷你解譯器。

**還沒做**:9 個網路畫面(版面座標未抽;`Draw_Net_Next_Turn_Screen_` @ 0xF1075 與
`Add_Net_Next_Turn_Fields_` @ 0xEFCEA 是抽的起點)。

## ★ 2026-08-07 Net_Next_Turn 等待畫面(gap report 第 29 項(決定性化))

**第一張版面是「算」出來的畫面**:remake 移植過的畫面座標都是反組譯裡的立即數,
這一張不是——`Load_Net_Next_Turn_Screen_` 依**資產尺寸**現算:
`x = (0x280 − 資產42.寬)/2`、`y = max(0,(0x1E0 − 三塊總高)/2)`、`[win+0xBF] = 42.高 + 43.高`。
代進 lbxinfo 量到的尺寸 → 標題帶 (5,16)、中段 (5,64)、下段 (5,243)。
`Add_Net_Next_Turn_Fields_` 再給:輸入列 y=430 高 17、玩家列間距 25。

**測試釘的是算式不是數字**:重算一次那兩個式子,並確認三塊是上下相接的
(原版是一塊接一塊堆下去,不是各自定位)。資產換了或算錯了就會紅。

**誠實留白**:玩家列的起始 y 藏在 window 結構欄位裡沒有立即數(間距用真值 25,起點標估計);
原版 y=430 那列是聊天輸入(**2026-08-07 已補上**,見第 36 項(聊天列補完))。
狀態指紋擺在畫面上不是裝飾——分岔時兩邊念一下那八個字元就知道是不是同一個狀態。

**三張畫面不做,而且不是因為做不動**:`Modem_Setup` / `NullModem_Setup` / `Comm Info`
是數據機與序列線的設定,那些硬體現在不存在,remake 走 TCP——**替不存在的硬體做設定畫面
不是還原,是裝飾**。多人設定畫面上那兩顆鈕現在會直說這件事。

**剩 5 張**連線流程的畫面,要等 UI 端的連線流程做出來才有東西可顯示——
先做畫面會做出一堆沒有資料來源的空框。

## ★ 2026-08-07 連線大廳 + Choose_Net_Plyrs 名冊(gap report 第 29 項(決定性化))

上一項結尾就是這一項的前提:先補**大廳**(`internal/netplay/lobby.go`)才有東西可顯示。
主機聽 → 客戶端送 `hello` → 主機指派 id 並**廣播整份名冊**(含種子)。

兩個決定都不是隨手選的:**玩家編號由主機指派**(鎖步的指令排序鍵就是編號,
各自取號會撞號 → 一定分岔)、**種子由主機決定並廣播**(各自產生種子就不是同一局)。

**第一個「尺寸隨資料變」的版面**:`Choose_Network_Plyrs_Screen_` @ 0xF0E17 的
`總高 = 資產28.高 × 列數 − 1 + 資產27.高 + 資產29.高`,中段面板每位玩家重複一次
——1 人 y=163、4 人 y=109、8 人 y=37。`Add_Choose_Net_Plyrs_Fields_` @ 0xEFB50 給每列
`x1=+0x6A / y1=+i×0x24+0x40 / x2=+0x1B3 / y2=y1+0x1D` → 每列 329×29,**列距 36 正好等於
資產 28 的高**(那個相等就是交叉驗證)。

**截圖抓到一個讀程式讀不出來的錯**:狀態那兩行字第一版畫在底框(資產 29)裡面——
38 px 在數字上完全放得下,但那 38 px 的可見內容只有頂端那圈金屬圓角,底下是透明的,
於是第一行壓在圓角上、第二行掉出視窗。**版面的驗收是看圖不是算數字**。已移到底框下方
並補測試釘住(1~8 列都要在畫面內)。

**誠實留白**:沒有文字輸入框 → 「加入」目前只連得上本機;不能點列指派種族
(原版可以,`sub_EFABA`)——要接得把種族選擇整段納入連線流程,不做半套;
沒有重連/心跳/加密,這是區網對戰的最低限度。

**剩 4 張**:`Join_Net` / `Choose_Multi_Net_Game` / `Generic_Net_Info` / `SendGet_Net_Info`。

## ★ 2026-08-07 連線狀態面板(gap report 第 29 項(決定性化))

**反組譯把「還缺 4 張」改寫成「還缺 1 張」**:`Draw_Generic_Net_Info_Screen_` 與
`Draw_Join_Net_Screen_` 是**同一個位址**(0xF19C7)。往上追,`Reload_Generic_Net_Info_`
收一個資產編號當參數,七個 `Reload_*_Info_` 都只是帶不同編號呼叫它
——15 等人加入 / 23 加入中 / 24 等種族資料 / 25 初始化 / 26 傳送 / 30 產生星圖 / 31 接收。
所以那不是七張畫面,是**一個面板 + 一個狀態列舉**。版面照樣是算的:置中於 640×480。

**這一輪修掉兩個自己犯的錯,兩個都是截圖抓到的:**

1. **把 `Add_Waiting_For_Joiners_Field_` 讀成人數欄位。** 截圖上數字壓在 START NET GAME 上,
   才回去查它呼叫的 `sub_1151B0` = **`Add_Button_Field_`**——那個座標是**按鈕**。
   符號名是二手推論,被呼叫的函式是一手事實。
2. **LBX 多幀動畫是 delta 幀。** 第 0 幀完整、之後只帶會變的像素;逐幀獨立上色會讓
   整張面板消失。這個 bug 一直都在,只是截圖廊每張都恰好落在第 0 幀。修在
   `internal/lbx`(`AccumulatedUpToRGBA`)——資產 27、42 也會踩,只是還沒播到。

**誠實留白**:只有「等待加入」有觸發點(主機開大廳 → 這一張 → 點過去進名冊);
「加入對局中」永遠不會停留(`netplay.Join` 是同步的,原版慢是因為 IPX/數據機協商);
人數位置是量的不是真值。

**剩 1 張**:`Choose_Multi_Net_Game`(`Load_Choose_Multi_Net_Game_Screen_` @ 0xF40D3 +
`Add_Choose_Multi_Net_Game_Fields_` @ 0xEFF87 已抽:主面板資產 41,
`y = ((0x1E0 − 高) − 0x51)/2 + 0x25`,10 列對局,列高 22、列距 27、起點 +64)。

## ★ 2026-08-07 區網對局探索 + Choose_Multi_Net_Game(gap report 第 29 項(決定性化))

**這張畫面的資料從哪來,原版沒有回答**:原版走 IPX,而「列出區網上有哪些對局」是
**協定**自帶的服務公告,不是遊戲做的。TCP 沒有這個能力——照抄畫面會得到一張永遠空的清單。
所以先補 `internal/netplay/discovery.go`(UDP 廣播:主機每秒公告名字/位址/人數,
客戶端聽 24502 收集去重),**那一層是移植決策不是還原**。

三個實作決定:來源 IP 覆蓋封包裡寫的(主機常不知道自己對外是哪個位址);
清單依名稱排序不依到達順序(順序決定玩家點到哪一場);`Browser` 不阻塞
(UI 單執行緒,收兩秒再回傳等於凍住兩秒)。測試全走 127.0.0.1,含一支正對照。

**版面**:`x=(0x280−479)/2=80`、`y=((0x1E0−384)−0x51)/2+0x25=44`。那個 −0x51 剛好等於
標題帶的高,但這張**沒有畫標題帶**——是版面上的讓位。照抄數字,不照抄自己對數字的解讀。
每列熱區 362×22、列距 27、最多 10 場;字在列裡垂直置中。

**又一個截圖抓到的錯**:第一版把原版的 `(0x16 − 字高)/2` 加了字高當基線,整欄字掉到下一列
——`uifont.Draw` 底層是 ebiten **text/v2**,y 是**行框上緣**不是基線(v1 才是)。
測試另釘一條:第 i 列的字不得落進第 i+1 列的熱區。

**誠實留白**:UDP 廣播只跨得過同一個區網(原版 IPX 同一個限制,不是退步);沒有簽章加密;
改對局名稱要輸入框(上限 8、需唯一,規則已記在 `netplay.GameNameMax`)。

**9 張網路畫面結案**:6 張做了、3 張明確不做(數據機/序列線硬體已不存在)。
網路多人剩下的是**文字輸入框**——跨網段直連、改對局名、聊天列都等它。

## ★ 2026-08-07 文字輸入彈窗(gap report 第 29 項(決定性化))

remake 先前一路寫著「原版的輸入是內嵌欄位、remake 沒有輸入框」——**那個判斷是錯的**。
`Change_MP_Game_Name_` 呼叫的 `sub_91BB4` 在符號表裡叫 **`Remapped_Input_Box_Popup_`**:
原版有一個獨立的 modal 彈窗,連自己的 LBX 都有(`INBOX.LBX`,底框 288×151 + ACCEPT 鈕 98×28)。

這是連續第三次靠「**符號名是二手推論,被呼叫的函式是一手事實**」修正判斷
(第 29 項(決定性化) `Add_Button_Field_`、第 29 項(決定性化) `Add_Hidden_Field_`)。

**版面**全部是立即數:標題帶 y+3 高 54(字在裡面垂直置中)、輸入欄 (x+34, y+54) 高 26 寬 234、
ACCEPT 鈕 (x+96, y+100)、長度上限夾在 205;彈窗位置 (177, 125)。
⚠ 輸入欄左邊距 34、右邊距 20——**不對稱**,是兩個獨立的立即數。照抄,並寫測試防止有人改成對稱。

**接上去的兩處**:主機開局前先問對局名稱(原版順序,上限 8);清單畫面多一顆
「直接輸入位址」——**原版沒有這顆鈕**(IPX 自己找得到),擺在清單外面的空白帶,
不佔用任何原版座標。

**誠實留白**:輸入走 ebiten `AppendInputChars` 而非原版的逐鍵掃描碼(掃描碼在現代平台拿不到,
而且會擋掉輸入法),代價是插入模式之類的鍵行為沒還原;游標閃爍週期是自己訂的。

**歷史日誌原結論（2026-08-10 勘誤：正式網路對局尚未完成）**:傳輸層 + 鎖步 + 決定性 + 指令層 + 大廳 + 區網探索 + 6 張畫面 + 輸入框已完成；這不等於 `cmd/moo2` 已接共同開局與實際回合。
剩**聊天列**(`Chat_Box_Input_Loop_` @ 0xF55A4 / `Send_Chat_Msg_` @ 0xDD3B8 已定位)——加分項,不是缺口。
(**2026-08-07 補上了**,見下面第 36 項(聊天列補完)。)

## ★ 2026-08-07 TECH LEVEL 的第二個真效果(gap report 第 30 項(TECH LEVEL 第二效果))

`shell.TechLevels` 的註解自己寫著「開局已知科技領域數…**沒有一手表之前不臆造**」
——也就是選單上寫「一般」、實際拿到的是曲速前的科技,而且沒有任何錯誤訊息。

`Init_Player_Tech_` @ 0x5E55F 給了兩樣:**送幾個**(`var_18` = 1 / 6 / 25,由
`byte_199CB5` = NEW GAME 的 TECH LEVEL 決定)、**送哪些**(`word_18111C` =
29, 55, 22, 57, 28, 23 = 工程學 / 核分裂 / 化學 / 物理 / 電子 / 冷聚變)。
前 6 次取固定表,第 7 次起 `sub_FD335` 隨機挑——25 級是「6 固定 + 19 隨機」。

**三方互證**:手冊獨立說「第一個是 field #29」;反組譯 `word_18111C[0] = 0x1D = 29`;
第二個 55 = 核分裂,而 remake 早就把 `FTLTopic` 定成它——手冊說 Average
「已具備星際航行所需的全部科技」,兩條獨立的線指到同一個編號。

**接線時抓到的陷阱**:`applyStartingTech` 只加不減的話,demo 局用預設等級發過的核分裂
會留在曲速前的局裡——正好是「曲速前不該有 FTL」的反例,而且靜默。已改成先清再發 + 正對照測試。
AI 也要一起發(原版是逐玩家跑的)。

**驗收看畫面**:截圖廊 9 張變了而且變得對——科技總覽從 2 項變 7 項(6 + field 0)、
建造清單多出運輸艦隊(前置正是核分裂)。

**誠實留白**:先進級的 19 個隨機主題沒發(缺口大小由 `gamedata.StartingTopicRandomExtras`
回報,不是註解裡的一句話);初始建築數上限 3/5/9 要先有「依人口生成母星建築」的機制。

## ★ 2026-08-07 開局建築的優先清單(gap report 第 31 項(開局建築清單))

上一項留的「初始建築數上限要先有依人口生成的機制」——那個機制缺的其實也是一張表,
而且 `shell.StartingBuildingCount` 的註解自己標了:「實際會生成哪些建築仍取決於
**initial_buildings 優先清單**與已知科技」。

`Init_Homeworld_Colony2_` @ 0x13A3D 給了兩樣:**上限表 `byte_13A3A` = 3, 5, 9**
(與手冊逐字相同)、**優先清單 `word_17D8AC`** = 32 個建築編號,開頭 41 → 8 → 40 是
Star Fortress → Battlestation → Star Base,同一條升級鏈**最強的排最前面**。

**四條獨立的線對上**:手冊說「Pre-warp/Average 只有 Marine Barracks 和 Star Base」;
拿這份清單 × 第 30 項(TECH LEVEL 第二效果)的六個開局主題 × remake 的建築前置表跑一遍,科技條件成立的
**正好只有那兩棟**。清單、主題表、前置表、手冊——四個來源互證。

**驗收是截圖廊零差異**:把寫死的兩棟換成從清單算,一般等級算出來仍是那兩棟,34 張逐位元組相同。
另有正對照測試:新舊兩條路要走到同一個答案。

**缺口被釘在上一層**:先進級仍只有兩棟,不是這一層沒做——名額有 6 個,但科技只夠兩棟,
因為第 30 項(TECH LEVEL 第二效果)留的 19 個隨機主題還沒發。測試附正對照:科技全解時這套機制確實會發滿 6 個。
那 19 個要港 `Choose_Tech_Application_` @ 0xFD335(294 行的 AI 權重選擇器),一次讀就照抄
風險太高,留作獨立一輪。

順帶把 `origBuildingID` 搬進 `internal/gamedata`:畫地表 sprite 與這份清單要靠同一份編號對照。

## ★ 2026-08-07 飛彈速度:那個「手冊自相矛盾」不是矛盾(gap report 第 32 項(飛彈速度))

`missile.go` 的檔頭從移植那天起就寫著「手冊此段自相矛盾…**此落差需日後對實機行為動態驗證**」,
而 `HONEST-STATUS.md` 把它列在「需原版 oracle 對照」。

`Missile_Speed_` @ 0x3CD21 的最後三行解掉了,**不需要 oracle**:

    test [ebp+var_3], 10h     ; 旗標 0x10 = Fast 改造
    jz   short loc_3CE49
    add  edx, 4               ; 只有旗標成立才 +4

所以手冊的**附表 = 沒有 Fast 改造**、**明列公式 = 裝了 Fast 改造**,兩者都對,只是手冊
沒說那個 FastBonus 是有條件的。remake 先前無條件 +4,等於每枚飛彈都預設有改造
——Beam Defense 憑空高 20,飛彈比原版難打下來。

同一支函式還推翻了「基礎速度固定 12」:依武器類型分 6/8/10/12/20/24 六檔,
其中四種 raw kind(0x12/0x13/0x14、0x28)`xor ecx, ecx` **不隨驅動等級變**(很容易漏抄,已釘測試)。

**誠實留白**:`[player+0x8BC]` 那個讓 6→10/8→12/10→14 的玩家旗標還沒追出是什麼,
`MissileBaseSpeed` 留一個 `boosted` 參數、呼叫端傳 false——留誠實的參數比假裝完整好。

**教訓**:「手冊矛盾 → 選一個 → 待實機驗證」在文件裡放了很久,答案一直在執行檔裡。
把它列進「需 oracle」是**分類錯誤**——它需要的是把靜態來源窮盡完。

## ★ 2026-08-07 地面戰:結構不是「未核實」,是抄了一代的(gap report 第 33 項(地面戰結構))

`ground_battle.go` 檔頭寫著「解算結構取自**一代(1oom)**」,`HONEST-STATUS.md` 寫「結構本身未對
MOO2 實機核實」。`Ground_Combat_Round_` @ 0xEC4FE 給出原版的 26 位元組結構(四種部隊,
各有攻擊力/數量/耐受值,+ 當前類型/累積命中/本回合命中與陣亡的類型),欄位剛好排滿沒有空隙。

**三處實質差異**:①平手時原版**雙方都挨打**(兩個獨立的 if),一代是 if/else 只有攻方挨打
——d100 對 d100 平手是 1%,守方原本白拿的優勢沒了;②攻擊力**逐部隊類型**,不是整隊一個;
③累積命中用 `==` 判定,不是扣到 `<=0`。

**順帶消掉一段技術債**:先前有一整段在解釋「為什麼把戰車營排在陣列尾端」(因為只回傳一個總
存活數,要用 min() 推算分兵種)。原版逐類型記數量,戰後直接讀 `Count[類型]` 就是真實存活數。

**誠實留白**:四種部隊類型**沒有對出名字**(remake 只用兩種,不編第三/四種);
每種部隊各自的攻擊力表還沒追,兩種暫填同一個 atkForce(= 維持現行數字,差異留在看得見的地方)。

## ★ 2026-08-07 四種部隊類型是什麼(gap report 第 33 項(地面戰結構))

上一項留的兩個留白同一天追完。`Compute_Ground_Combat_Info_` @ 0xEC3CE 的四個 case:
類型 0 = +10 攻擊 +1 耐受、類型 1 = 基準、類型 2 = −10、類型 3 = −20(基礎取自另一方)。
`Compute_Colony_Ground_Combat_Info_` @ 0xED713 給殖民地填**三格**。

手冊補上名字:「Marine and Armor units … your **militia** are also shown here」——
殖民地正好三種,對照調整量:**類型 0 = 裝甲、1 = 陸戰隊、2 = 民兵**(最弱,合理)。
類型 3 不編名字(殖民地不填它)。

**順帶抓到一個順序錯**:remake 先前把陸戰隊排 0、戰車營排 1,是反的。先前兩種填同一個
攻擊力所以看不出來,接上逐類型的差之後會差 10 點。已訂正 + 測試釘住。

**誠實留白**:只實作立即數的部分(科技加成那兩欄還沒對出意義,回「差不多」的值會讓日後
追出真值時看不出哪裡被污染);守方的民兵沒接(數量公式在 `sub_EC61E`),那格留 0 =
**少算守方兵力**,方向上對玩家有利,是明說的偏差。

## ★ 2026-08-07 民兵接上了(gap report 第 33 項(地面戰結構))

上一項留的 `sub_EC61E` 同一天追完:`Colony_N_Militia_` @ 0xEC61E 逐個人口單位掃
(每個 4 位元組,低 4 bits 是擁有者編號)再**除以 5**。兩個跳過條件都是人口單位上的
資料旗標,而 remake 的人口沒有逐單位模型 → 恆不成立 → **⌊人口 / 5⌋**。

擁有者塞在低 4 bits 這件事與 `Init_Homeworld_Colony2_` 的寫入(`and ebx, 0Fh`)對得上
——**同一個結構在兩支函式裡互證**。

守方現在是陸戰隊 + 民兵兩格(民兵攻擊力低 10),`DefenderStart` 的回報也跟著含民兵。
⚠ 裝甲那格仍留 0:AI 沒有建築追蹤機制,無資料可誠實推導,不臆測。

**改變了平衡但方向是忠實的**:守方憑空多 ⌊人口/5⌋ 個單位(母星 8 人口 → +1),
既有的兩支入侵勝率測試仍綠——偏移沒把勝負翻過去,只是把守方下限抬起來。

## ★ 2026-08-07 地面戰加成塊:難度加成不給玩家(gap report 第 33 項(地面戰結構))

第 33 項(地面戰結構)留的「加成塊欄位還沒對出意義」追完了。`Compute_Player_Ground_Combat_Bonuses_`
@ 0xEC15C 產 19 位元組,大多對應手冊已列出的加成類別(remake 已用手冊的表算過),
但**有兩條手冊完全沒寫**:

**①基礎耐受命中數是 1**(預設一下死一個),某個科技(`[player+0x8AA]`)讓所有部隊變成要兩下。

**②難度加成不給人類玩家**:

    人類玩家    0            ; [player+0x28] == 100 這個標記
    AI 帝國     難度 − 2      ; 普通=0、不可能=+2、教學=−2
    安塔蘭那側  難度×2 − 4    ; 恰好是 AI 的兩倍

兩點值得記:加成是**以「普通」為基準往兩邊偏**,不是一律加成(教學難度下 AI 是負的);
而 `[player+0x28] == 100` 這個人類玩家標記在 `Init_Player_Tech_` 也出現過
——**同一個標記在兩支不相干的函式裡對得上**。

已接進入侵:守方(AI)加「難度 − 2」,攻方(人類玩家)**不加**——不是漏掉,
是原版就沒有,註解寫明以免日後被人「補上」。

**誠實留白**:`[+5]`/`[+7]`/`[+9]` 那三張查表與 `[+0x0B]`/`[+0x10]` 還沒對出意義;
它們對應手冊已列出的加成,remake 已算過,**不重複實作**免得同一個加成被加兩次。

## ★ 2026-08-07 重力種族特性(gap report 第 34 項(重力種族特性))

上一項留的加成塊欄位又追出三個,三個都與手冊互證。

**那個 `else` 就是「互斥」的證據**:原版先看 High-G(`[player+0x8AA]`),不成立才看
Low-G(`[player+0x8A9]`)——而手冊明寫「High-G World and Low-G World are mutually exclusive」。

**High-G 手冊逐字**:「they take **1 hit more** than normal troops before being slain」
= `mov byte ptr [out+0Ch], 1`,而耐受 = `[+0x0C] + 1` → 一般 1 下、High-G 2 下。一字不差。

**Low-G 有落差**:手冊寫「a **10%** penalty」,原版是 `mov byte ptr [ecx+0Dh], 0F6h`
= **定值 −10**。它與其他加成一起加進攻擊力,而那些也都是 +10/+15/+20 的定值——
手冊那個「%」多半是行文的隨手寫法。remake 先前照字面做乘法,註解還寫著「手冊未列出
10% 套用在哪個基準值」——**那個不確定性現在有答案了**。已改成定值。

⚠ 舊測試裡 `100 → 90` 這一列**兩種算法答案剛好相同**——只測那個數驗不出這個改動。
新測試加了 50/10/7/0 與「定值 = 差與基準無關」的性質。

**Subterranean 升級為雙來源**:`mov byte ptr [out+0Eh], 0Ah` + 只有守方才傳那個旗標,
數字與條件都對上 remake 既有的手冊值,沒有改動。

**誠實留白**:`[player+0x8A7]` 看起來是種族地面戰加成,但**沒有直接證據**,不寫進程式碼。

## ★ 2026-08-07 三張查表讀出來了(gap report 第 35 項(三張查表))

索引函式的符號名直接說了:`Player_Best_Armor_` / `Player_Best_Rifle_` /
`Player_Best_Personal_Shield_`,三支都是「從表尾往前找第一個已知的科技」= **取最高階,不是加總**。

先建 VA → 檔案位移的對照(用 `aMultigmLbx` 後面緊接的 `byte_17A061` 反推,delta = 0x7E694,
再用另一個同名字串驗證落在 `;org 178000h` 之後 4 位元組),然後直接從 exe 讀表。

**十二個科技 id 全部對上 remake 的 `Technology` 列舉**,而裝甲的上五項與個人護盾都與手冊
逐字相同——這是「這三張表就是它們」的證明。

**於是抓到兩個實質缺口**:①**鈦裝甲 +5 少了**(手冊沒列,而鈦裝甲開局就有 → 每個帝國、
每一場地面戰都少 5 點);②**整條步槍通道 remake 完全沒有**(Pulse 0 / Laser 5 / Fusion 10 /
Phasor 20 / Plasma 30,上限差 **30 點**)。兩者都已補上並接進玩家與 AI 的 force。

**順帶訂正兩個「給誰」**:加成塊的三個科技旗標也解出來了——Anti-Grav Harness 給**所有類型**、
**Battleoids 只給裝甲**(+10 攻 **+1 耐**,手冊只提 +10)、**動力裝甲只給陸戰隊**。
remake 先前把後兩項都加給整支部隊。常數已記進 gamedata,分兵種的接線留下一輪。

## ★ 2026-08-07 分兵種接線 + 四個 hits 數字被重建出來(gap report 第 33 項(地面戰結構))

上一項留的分兵種接線接完了,而且過程中發現**手冊列的四個 hits 值可以完全由反組譯的加法
結構重建**:陸戰隊 1 = 基礎 1 + 類型 1 的 0;+動力裝甲 2 = 1 + `[out+4]`;戰車營 2 = 1 +
類型 0 的 1;機械戰士 3 = 1 + 1 + `[out+2]`。**四個獨立的手冊數字,由三個獨立的反組譯欄位
加出來**——這種吻合不會來自誤讀。落成 `TestManualHitValuesReconstructFromTheOriginalStructure`。

**於是也發現一個算兩次的坑**:`tankHitsToKillFor` 回的是手冊的**成品值**,而第 33 項(地面戰結構)接上的
`GroundTypeHitsDelta` 是**組成的一部分**,兩個一起用會變成 3 / 4。已改成只用成品值。

**分兵種接線**:Powered Armor 只給陸戰隊、Battleoids 只給裝甲(先前兩項都加給整支部隊);
Anti-Grav Harness 與 Personal Shield 留在共用那份。

**順帶消掉一個「為了繞過錯誤而存在」的守門**:舊的 `tankForceBonusFor` 有個「0 輛戰車不給」
的判斷——那個守門存在的理由正是加成被加進整側的 force,而那本身就是錯的。
加成落到戰車那一格之後,沒有戰車時那格本來就是空的。**修好根因,補丁自己掉下來。**

## ★ 2026-08-07 聊天列(gap report 第 36 項(聊天列補完))

「等待其他玩家」那張畫面的輸入列先前是一條寫著「remake 未實作」的提示帶。做得動的理由是
第 29 項(決定性化)把文字輸入框做出來了,而這一輪把原版那條線整條讀完:`Chat_Box_Input_Loop_` @ 0xF55A4
(進聊天模式 → 非空才送 → 清空重新武裝)、`Send_Chat_Msg_` @ 0xDD3B8(封包型別 `27h`)、
`Receive_Chat_Msg_` @ 0xDD351(環的結構)、`sub_F1075` 的繪製段(兩種前綴 + 版面)。

**四個數字每個都指得出出處**:14 則(`cmp [+47Ch], 0Eh`)、每則 82 byte(`imul …, 52h`)、
內文上限 80(82 − 發話者 1 − NUL 1)、發話者 ≥ 8 是 GNN(`cmp ax, 8 / jge`)。
計數欄位落在 `[+0x47C]` 自己就是第二條線:**14 × 82 = 1148 = 0x47C**,陣列剛好塞滿到那裡。

**版面自己對上**:繪製段給的是相對偏移(x +0x18、首行 y +0x0E、行距 0x0C),套進資產 40(y=243)
→ 首行 257、14 行後底部 424,而 `Add_Net_Next_Turn_Fields_` 給的輸入列在 **430**。
**中間剩 6 px。** 兩個函式互不知情卻嚴絲合縫——第二個獨立來源,落成測試。

**一個必須偏離原版的地方**:80 是緩衝區的 **byte 數**不是字元數。原版單 byte ASCII 切在哪都合法,
UTF-8 切半個字會變亂碼——`ChatTruncate` 守住 80 byte 但截在 rune 邊界(中文約 26 字)。

**誠實留白**:`Send_Chat_Msg_` 只發給 `[player+0x28] == 'd'` 的玩家,但**沒查到那個欄位的寫入端**
——不照抄、不編名字,改成發給所有已連線的對手。送出目前只進本機記錄(鎖步的 `Table` 一回合只收
一則,聊天塞進同一條線會壞掉鎖步),真接上連線時多一個 `WriteFrame` 即可。

**順帶**:比對截圖時發現 `docs/screenshots/` 只有 27 張而 gallery 產 35 張——**八張從沒進過版控,
byte-diff 驗收對它們等於沒跑**。七張決定性的已補進版控;`18_loadgame` 帶存檔時間戳,刻意不收。

## ★ 2026-08-07 整棵研究樹升格成一手驗證(gap report 第 37 項(研究樹一手驗證))

`techtree.go` 的 83 列一直是「逐字轉寫自 openorion2」——**二手**。這一輪從原版執行檔挖出
同一張表對了一次:**74/83 個成本逐字相同、199 條科技歸屬全中、remake 沒有多出任何科技**。

最強的一條是**兩種編碼的交叉驗證**:openorion2 把樹寫成「8 個領域各一串主題」,原版執行檔
寫成「每個主題一個後繼」的鏈——互不知情,**73 條銜接關係逐條吻合**。

**九個不同的全部有解釋**:8 個 Hyper-Advanced 主題是**真版本差異**(1996/1.31 = 15000、
1.5 = 25000,三份執行檔各查一次),remake 的 `RuleProfile` 早就對,現在多了一手來源;
主題 74 XENON 原版寫 15000 但 **next 指向自己**,自環就是「永遠解不開」的編碼,
那 8 個是安塔蘭專屬科技。

**順帶修掉一個真錯誤**:`StarterResearchTopics` 那份手挑的 9 個主題,在開局那一刻有 4 個
不該出現、漏了 3 個該出現的。已改成由樹算(`AvailableResearchTopics`)。
`-game` 主路徑的 `currentAreaTopic` 本來就是對的,這條 next 鏈是它的一手佐證。

**又撞到同一組六個主題**:`sub_FD2F9` 的六個硬編位址減 `0xC4` 正好是第 30 項(TECH LEVEL 第二效果)的六個開局主題
——第三個獨立來源,而且說明了角色:AI 的科技權重要等這六個全完成才啟用一般模式。

**2026-08-25 勘誤**：這是當時以 asm 行數估計的舊結論。現行 IDA 證據為 655 條指令，
先進級十九次應用級單次抽選與人類估值已接線；AI 完整估值仍待 RE。

## ★ 2026-08-07 三面行星護盾 + 自動實驗室 + 再生反應爐(gap report 第 38 項(行星護盾等三棟))

HONEST-STATUS 寫著「部分軍事/防禦建築(~13 棟,需艦隊駐防/軌道防禦系統先落地)」。
照 rulebook 63 對程式碼盤點,實際是 **11 棟**,而且其中三棟**根本不需要新子系統**
——它們接的軌道轟炸早就有了,缺的只是資料。

**三面護盾**:手冊給 −5 / −10 / −20,而且維護費(1/3/5 BC)三棟全部對得上執行檔的建築表
——**減傷與維護費出自同一段文字**,所以那段可信,減傷不是孤證。
「per attack」決定了接在**逐發**傷害而不是總傷害（多次攻擊下差一個數量級），
測試以 `TotalDamage` 釘住逐發減傷；runtime 命中換算現依 IDA 呼叫鏈採除以 40。
手冊寫「取代」不是「疊加」,所以取最強那一面。

**再生反應爐**:接對地方比接上去重要。手冊「不計入污染」那句決定了它**不能**接 `FlatIndustry`
(那個欄位在污染縮減之前併入 gross),改成旗標、在污染切分點之後才加。測試拆成兩個獨立斷言
外加一條正對照(同樣的產能接成 FlatIndustry 污染一定會變)。

**自動實驗室**:手冊 +30 研究點/回合,只動 `FlatResearch`。

**誠實留白**:護盾的「Radiated 轉 Barren」與「生物武器無法進入」沒接
(⚠ 後者 **2026-08-08 第 52 項(生物武器分類)已接**,這一行是當時的快照);剩下 8 棟
(食物複製機/阿提米絲/太空學院/異族管理中心/戰機基地/恆星轉換器…)**是真的需要新子系統**。
`30_netwait.png` 的指紋變了是 `ColonyState` 多欄位的必然結果,畫面其餘像素完全相同。

## ★ 2026-08-07 食物複製機(gap report 第 38 項(行星護盾等三棟))

手冊 p.85 一整句就是完整規格,而且三個限定詞缺一不可:**two-for-one**(2 產能 → 1 食物)、
**1 BC per food**(從國庫)、**as needed**(只補缺口)。

**最後那條是整棟建築的平衡**。漏掉它,殖民地會把全部產能換成食物、再靠既有的餘糧出售
換回 BC——2 產能 → 1 食物 → 0.5 BC,在高產能低稅率下比稅收更好賺。**原版沒有這個東西。**
測試有一條專門釘它:有盈餘時產能與盈餘兩個數字都必須完全不動。

**維護費 10 BC** 手冊與執行檔建築表兩個來源一致,而且是**全表最貴**(第二貴是 5)
——測試連這點都釘住,被改小就失衡。

**接在污染扣完之後、人口成長之前**:換算要用可用產能,而成長同時吃盈餘與淨產能兩個數字。

**誠實留白**:手冊沒說國庫不夠付會怎樣,**不編規則**——換算照做、成本照報。硬加一條
「付不起就不換」會憑空發明規則,而且讓饑荒在破產時突然惡化。

**順帶清掉過期斷言**:`session.go` 那段「其餘 20 項…暫不建模」的清單被第 38/38 項做掉六棟,
已改寫成逐項說明還缺哪個子系統。建築表 41 棟裡未被消費的剩 **7 棟**,都是真的缺子系統。

## ★ 2026-08-07 阿提米絲系統網:水雷子系統(gap report 第 38 項(行星護盾等三棟))

上一項把它列在「真的缺子系統」那一欄。手冊 p.86 其實把整個子系統寫完了,而且
**remake 剛好已經有那兩個輸入**:艦體等級就是 `shipStrength` 的六個類別、護盾等級就寫在
元件名字裡(「第十級護盾」的「十」)。所以缺的不是子系統,是沒人把手冊那段翻成程式碼。

三件事相乘:**觸發**(逐艦,20/30/40/50/80/100% 依艦體)× **水雷數**(8–28)×
**每枚傷害**(20 − 護盾等級)。機率隨體積上升而傷害不隨體積下降,所以水雷網**專打主力艦**
——一群巡防艦大多開得過去,一艘末日之星必中。測試釘住「機率單調上升」這個性質,
以及「第十級護盾把傷害折半」(兩艘同型船跑同一組亂數流比總量)。

**接在「進入」那一刻**(手冊 entering,不是停留每回合),放在探索標記之後、一次性發現之前
——雷區是進門就炸,發現是進去之後才看到的。亂數用回合+星系當種子,有決定性測試。

**誠實留白**:只對玩家艦隊生效(AI 沒有艦隊移動模型,那是 AI 的缺口不是這系統沒接);
手冊沒說水雷會不會消耗,照字面做不消耗;偵察艦/殖民船不是原版艦體等級,原版那些都是
Frigate 艦體上的設計,所以套 20%。建築表未被消費的剩 **6 棟**。

## ★ 2026-08-07 艦員經驗系統(gap report 第 39 項(艦員經驗))

上一項把太空學院列在「缺艦員經驗值子系統」。盤點後發現 remake 的狀況很特別:**三張加成表
已經有了而且都對得上手冊**(BA/BD 在 `formulas.go`、ME 在 `missile.go`),而且**接到船上了**——`Ship.CrewXP` 欄位在,等級由 `shipCrewLevel` 現算,新造艦依殖民地起始經驗;第 60 項(打得準也閃得掉)把它接上攻擊與防禦兩側。

第四條軌 Bo 登艦戰 `{0,5,10,15,20}` 已於 2026-08-25 由
`Crew_Boarding_Combat_Bonus_ @ 0x35CAD` 再證實，並接入快速與格子戰術兩條路徑。

**統帥種族不是「升級快」,是整條階梯平移**:一般種族 Green(0)→Regular(50)→Veteran(150)→
Elite(500),統帥種族 Regular(0)→Veteran(50)→Elite(150)→Ultra-Elite(500)。兩張門檻表各有一個
**−1** 表示「這個種族到不了」,不是寫一個很大的數——那是兩件不同的事。

**等級不存,只存經驗**:`Ship` 只加 `CrewXP`,等級現算。太空學院的「起始等級 +1」也用經驗
表達(起始 XP = 那一級的門檻),不另開欄位。

**戰鬥經驗**：`sub_4B184 @ 0x4B184` 證實被摧毀敵艦 1-based 艦體級總和折半、
捨去、最少 1；即使一艘都沒沉仍為 1。快速結算以多重集合相減還原被擊沉艦，格子戰術
則在 casualty 壓縮時排除 `Captured` 並累加艦體級；兩路徑都只寫回倖存參戰 Ship。

**順帶收掉五個硬寫的 `false`**:`gamedata.Ground*BarracksCap` 的 `warlord` 參數在 shell 有五處
各自硬寫 false,新增 `GameSession.RaceWarlord` 統一。目前沒有種族會設它,行為完全相同。

**現況邊界**：玩家與 AI 都已有持久實艦、每回合 XP 與登艦 Bo；艦艇設計畫面直接造的船
吃不到學院加成，因該路徑沒有「在哪造」的位置資料。建築表未被消費的數量應另依現行程式盤點，
不沿用本節的歷史快照。

## ★ 2026-08-07 征服人口的同化系統(gap report 第 40 項(同化系統))

手冊把整張表逐個政體寫死了:封建 8、邦聯 4、獨裁 8、帝國 4、民主 4、**聯邦 2**、
**統一 20**、銀河統一 15 回合同化一單位人口。

**民主 4 vs 統一 20——差五倍**,那是原版把「征服流」與「和平流」分開的規則手段。
異族管理中心的「1 per 2 turns, **regardless of government**」直接蓋掉政體那一格,
**對統一政體等於十倍速**;測試把那個十倍釘住。

**兩個修正項,一個有數字一個沒有**:排斥種族減半(回合數 ×2,而且手冊說套在 base rate 上,
所以連建築的固定值也吃);魅力種族手冊只說「easily」**沒給數字**,1.5 手冊也沒補
——所以**現在沒有任何效果**,並寫了一支測試 `TestCharismaticHasNoQuantifiedEffectYet`
**把「刻意不做」與「忘了做」分開**:有人塞猜的倍率進去那支會紅。

**累積進度不歸零**:滿 N 回合同化一單位、餘數留著,不是每 N 回合重來——後者會在政體改變或
蓋起管理中心時吃掉已累積的進度。

**順帶抓到一個假斷言**:`session.go` 寫著「異族管理中心:colonyMoralePercent 讀取此建築名」
——**根本沒讀**,那個名字在整個 repo 只出現在資料表與註解裡。已改寫成實際狀況。

**誠實留白**:未同化人口**目前沒有負面效果**(20% 多種族士氣懲罰沒接、叛亂系統不存在)
——機制在、後果還沒接。建築表未被消費的剩 **4 棟**。

## ★ 2026-08-07 戰機基地 + 恆星轉換器,以及一個盤點方法的錯誤(gap report 第 41 項(戰機基地))

兩棟都接進既有的 `retaliationAttackers`(殖民地被轟炸時反擊)。戰機基地手冊 p.79 給
**10 / 6 / 4** 中隊(隨科技**遞減**,每階更強);恆星轉換器手冊 p.111 給「每面 400、四面 1600」。

**第 37 項(研究樹一手驗證)的一手科技表當場抓到一個錯**:我把 `TECH_HEAVY_FIGHTER_BAYS` 寫成
`TOPIC_ADVANCED_ROBOTICS`,查表發現它其實在 `TOPIC_SUPERSCALAR_CONSTRUCTION`(主題 42)
——重戰機那一檔會**永遠進不去**,而且不會讓任何測試變紅。那張表第一次派上用場就是抓錯。

**差一點送出去的雙重計算**:恆星轉換器**早就接過了**(在 `colonyDefense`,用常數而不是
字面字串),既有測試立刻紅了。順著查出兩件事:①同一棟建築擋得住 AI 來襲卻對軌道轟炸不反擊
(已統一到 `retaliationAttackers`);②`StellarConverterDefense = 800` 的來歷寫著
「400 傷 ×2(雙側共 1600)」——**那句自己就矛盾**(400×2=800≠1600),手冊原文是
「每一**面** 400、**四面**合計 1600」。已改成單面 400。

**盤點方法本身有洞**:第 38–40 項報的「剩 N 棟」用的是「字面字串有沒有出現在 buildings.go
以外」,漏判「在 buildings.go 內宣告成常數、由別處引用」的那一棟。補上常數引用與排除註解後
重掃:**41 棟全部有程式碼消費,0 棟未接**。所以第 38 項(行星護盾等三棟)那份 11 棟的清單裡恆星轉換器是誤判,
真正的缺口是 10 棟,第 38–41 項全部接完。

⚠ **「有程式碼消費」不等於「完整還原」**——好幾棟仍有寫明的部分實作,留白各自記在檔頭。

**2026-08-27 勘誤**：本節當時只確認「沒有逐殖民地中隊存量」，卻把它誤推成戰力算法一致。
IDA 重新追查 `0x5F64C` 後已證實原版使用最佳合格 beam／bomb、最佳裝甲與三檔權重，而非
10／6／4 乘泛用戰機近似值；現行程式已替換，完整推翻理由見
`docs/re/fighter-garrison-strength-audit-20260827.md`。

## ★ 2026-08-07 把上一輪自己寫的兩條留白關掉(gap report 第 42 項(關掉兩條留白))

第 41 項(戰機基地)結尾強調「有程式碼消費 ≠ 完整還原」並列出仍有的部分實作。其中兩條**在寫的當下
就已經解得開了**——擋住它們的東西是我自己前幾輪剛加上去的。

**① 多種族 20% 士氣懲罰**:`gamedata.MoraleMultiRacialPenalty` 早就存在而且是死碼,
理由是「remake 無多種族人口追蹤」。第 40 項(同化系統)加上 `UnassimilatedPop` 之後那個理由就不成立了。
接上去之後:攻下來的殖民地真的有代價(第 40 項(同化系統)的「機制在、後果還沒接」關掉)、異族管理中心的
第二條手冊效果生效、同化完的那一刻懲罰消失(`advanceAssimilation` 每輪重算士氣,有測試釘住)。

**② 三面護盾的「Radiated 轉 Barren」**:`ColonyState.Climate` 早就在(地形改造那輪加的),
建成時走既有 `applyClimateChange` 即可(會連帶調整食物與人口上限)。
⚠ **一個刻意的偏離**:手冊寫「as long as the shield remains in place」,remake 接成一次性,
護盾被炸掉不會變回 Radiated——因為 remake 的建築效果**沒有一個是可逆的**,為這一棟另建
「效果可撤銷」機制代價遠大於它修正的失真。**是選擇不是疏忽。**

屏障護盾的「生物武器無法進入大氣層」**已接**(2026-08-08 第 52 項:生物武器名單來自
執行檔 category 表的 category 20 = Bio-Terminator + Death Spores)。

**這一輪的形狀**:沒挖新的一手資料,做的是回頭把自己標的留白逐條檢查——
哪些是真的缺前置、哪些只是當時缺現在有了。**留白清單如果只增不減,
它就會從「誠實記錄」退化成「免責聲明」。**

## ★ 2026-08-07 先進級開局的 19 個隨機主題(gap report 第 43 項(先進級開局主題))

> 2026-08-25 現況：本段是歷史實作紀錄；「985 行」與 weight=1 已被
> `docs/re/starting-tech-application-audit-20260825.md` 推翻，現況只看頂端活表。

這個缺口從第 30 項(TECH LEVEL 第二效果)就開著，當時尚未閉合 `sub_FC845`。

`sub_FD335` 的評分是 `score = weight × horizon ÷ turns`,其中 **turns**(主題成本 ÷ 每回合
研究點)、**horizon**(15 起跳,不夠就 ×3÷2)、**候選集合**(只取現在可研究的主題)、
**加權隨機**(前綴和)**全部讀得出來**——擋住的只有 `weight` 一項。

把 weight 一律當 1 之後 `score = horizon ÷ turns`,選擇仍由成本主導。這比只發六個接近原版
得多:**先進級在原版是開局 25 個主題,remake 先前發 6 個,少的不是精度而是一整級的內容。**

現在發滿 25 個,而且**沿著樹往上走**(只從可研究的挑,測試逐領域檢查「已完成的前面不能有
沒完成的」)、**偏好便宜的**、**決定性**(同種子同開局;玩家與每個 AI 各一條流)。
第 37 項(研究樹一手驗證)的 `OrigTopicCost` 在這裡是燃料——沒那張表就只能用二手成本。

**當時留白（已由 2026-08-25 稽核取代）**：該版 weight 一律 1，且尚未找到
`[player+0x28]` 寫入端；現行程式已接人類估值與應用級單次抽選，AI 完整估值仍待 RE。

**寫測試時抓到一件事**:第一版斷言「先進級 25 個」實際拿到 26,差的正好是
`TOPIC_STARTING_TECH`(第 37 項(研究樹一手驗證)認出來的自環容器主題)。不是程式錯,是斷言少算了一個已知的東西。

## ★ 2026-08-07 上游補完之後,下游要跟著讀真的東西(gap report 第 44 項(下游讀真值))

第 43 項(先進級開局主題)把 19 個隨機主題發出去了,但先進級母星**還是只有兩棟**——`homeworldBuildingsFor`
從**固定表**現算科技集合,**看不到那 19 個**。上游補齊之後,下游如果還在自己算一份,
補齊就傳不下去。改法很小:多一支吃真正 `CompletedTopics` 的版本。

**結果自己對上了**:曲速前 1 主題 → 2 棟、一般 6 主題 → 2 棟(手冊逐字「only start with
Marine Barracks and a Star Base」)、**先進 25 主題 → 6 棟**(名額數 ⌈⅔×8⌉)。
那個 6 **正是第 31 項(開局建築清單)寫測試時留的正對照預測的數字**——缺口補上之後兩邊自己對上,
不是把斷言改成事後諸葛。

**順帶把過期斷言清掉**:第 43 項(先進級開局主題)做完後仍有四處文件/註解寫著「19 個隨機主題還沒接」
(`starting_tech.go` 兩處、gap report 第 30/31/37 項、HONEST-STATUS),全部改成現況。
`TestAdvancedStartIsBlockedByTheMissingRandomTopics` 改名為
`TestAdvancedStartFillsAllBuildingSlots` 並反轉斷言——**它自己當年就寫著「那 19 個若接上了,
這條測試要跟著改」**。

這一輪的形狀與第 42 項(關掉兩條留白)同款:**做完一件事之後,回頭找它讓哪些話變成假的**。
留白與缺口記錄的價值在於反映現況;一旦落後,它就從導航變成誤導。

## ★ 2026-08-07 領袖技能:疊加規則 + 四個技能 + 一個「不是缺口」(gap report 第 45 項(領袖技能))

**手冊給了一條 remake 一直做錯的規則**(p.137 Applicability):只有 **Megawealth 與
Researcher 可累加**,其餘取**最強的那一位**。`applyLeaderColonyBonuses` 先前是無條件 `+=`
——兩個貿易家就加兩份。已改成分組後依規則合成;測試同時釘兩邊(兩個貿易家**不**疊、
兩個科學家**要**疊)。負加成(環保官 −10%)取**絕對值最大**,取數值最大會挑到最弱那位。

**單位是查出來的**:加成值在 `baseSkillValues[2]`、單位在 `skillFormatStrings[2]`,兩張表在
openorion2 是分開的。**教官是固定點數而非百分比**,正好對上手冊「Boosts the number of
experience points earned each turn」——兩個獨立來源指到同一語意。

**接了四個**(標準仍是「有現成的承接欄位」):財務官→`IncomeBonusPercent`、
心靈導師→`MoralePercent`、醫官→`GrowthBonusSum`、**教官→艦員每回合經驗**(第 39 項(艦員經驗)才有的東西)。

**一個「不是缺口」的發現**:手冊 Tactics 那條的最後一句明寫 **This skill is not implemented**
——**原版自己就沒做**,remake 不做它與原版一致。記下來,否則下一個盤點的人會去找它該有什麼效果。

**順帶更正一個誤判**:查這輪時我一度說「`loadHerodataMercs` 沒有呼叫端」——錯的,
它在 `interactive.go:4384` 有呼叫,是 grep 被 `head` 截掉。真英雄池早就接上。

## ★ 2026-08-07 分項百分比:三個 admin 技能缺的只是一個欄位(gap report 第 45 項(領袖技能))

上一項把農業官/勞工官/科學官擋在門外,理由是「`ColonyState` 只有 per-worker 與固定值兩種
欄位」。回頭看那個理由——**缺的只是三個欄位**,而引擎早就有百分比進得去的地方
(士氣與重力就是走那條 `pct`)。

**士氣是三項一起動,這三個不是**,那是它們與士氣的唯一差別,也是測試的主軸:
農業官只動食物、勞工官只動工業、科學官只動研究,另外兩項必須一動也不動。
正對照是「士氣**仍然**三項一起動」——少了它,「分項百分比其實沒接上」也會通過。
第三支釘住「固定加成不吃百分比」。

**科學官 ≠ 科學家**:兩個中文名很像,但在原版是不同技能、不同單位(`%+d%%` vs `%+d`)、
**累加規則也不同**(科學家是手冊明列的累加型,科學官不是)。有一支測試專門釘這件事
——名字像就混用會同時錯兩處。

**領袖技能現況:接了 9 個**。仍未接的與理由:環保官(污染模型是八分之幾的查表,沒有百分比
入口)、**工程師**(艦艇維修速率——remake 照原版 `Repair_Ship_Full_` 做的是**一次修好**,
**那個量在這個模型裡不存在**,不是漏做)、戰術官(原版自己就沒實作)、
其餘 captain/common 技能對應的子系統 remake 沒有。

> ⚠ 2026-08-08 更正:上面工程師那句**只講對了一半**。手冊那條有兩句,這裡只看了第一句。
> 第二句「repairs all structural and internal systems damage **after the battle is won**」
> 對得上既有的 `repairAfterBattle`,已在第 45 項(領袖技能)接上。

## ★ 2026-08-08 技能欄位讀錯了,所以上面接好的技能真英雄一個都拿不到(gap report 第 45 項(領袖技能))

第 45/45 項把 admin 技能一個個接上去,但上游的 `loadHerodataMercs` **把技能位元讀錯了**。
原版 `Leader::hasSkill` 是 `(skills >> (2 * skillnum)) & 0x3`——**每技能 2 bit,值就是技能階**;
remake 讀的是 `1 << 6` / `1 << 9`,一個技能一個 bit。SKILL_RESEARCHER 真正在 bit 12-13,
bit 6 其實是 SKILL_FAMOUS 的低位。**兩個標籤都貼錯人**,而名字與等級都是對的,畫面看不出來。

順著這條線挖出另外四個:`Tier` 寫死 1(進階技能的 +50% 一次都沒發生)、一位英雄只給一項技能
(原版一人可有多項)、艦艇軍官通稱「指揮官」撞到 SKILL_COMMANDO 的譯名(每位艦艇軍官都吃到
地面戰加成)、以及**用翻譯過的中文標籤當識別鍵**——英文模式下三處查表全部落空,
**所有領袖加成同時靜默失效**。

修法:技能 id 才是識別鍵(`shell.Leader.Skills`),標籤只負責顯示;27 個技能的 id ↔ 中英文名
收進 `gamedata/leader_skill_names.go`(英文名逐字取自手冊 p.135-137)。舊資料(demo 領袖、
既有測試、舊存檔)沒有 `Skills` 時退回標籤反查,相容不變。

順帶接上**工程師**:手冊那條的第二句「戰後完全修復」正是 `repairAfterBattle` 在做的事,
加第三個觸發條件即可。`won` 這個新參數**只**影響工程師那一條——前兩條(自動修復元件、
進階損害管制)手冊沒有勝負條件,有一支測試專門釘住「打輸時自動修復元件仍然照修」。

## ★ 2026-08-08 環保官:第三次用同一種方式擋錯(gap report 第 45 項(領袖技能))

第 45 項(領袖技能)擋掉農業官/勞工官/科學官(「沒有分項百分比欄位」)→ 第 45 項(領袖技能)發現缺的只是三個欄位。
第 45 項(領袖技能)擋掉工程師(「維修速率那個量不存在」)→ 第 45 項(領袖技能)發現手冊那條的第二句對得上既有函式。
這一輪輪到環保官(「污染模型是八分之幾的查表,沒有百分比入口」)。

**同一個毛病第三次**:把「這個量是查表算出來的」誤讀成「沒有地方可以再乘一個百分比」。

手冊那句「Reduces **the amount of production that causes pollution**」逐字就是
`PollutionPollutingProduction` 這支函式的名字。接在查表之後、扣容忍值之前;
**不是減產能**(接到 gross 那一行會變成「請一個環保官等於少一個工人」,是手冊那句的反面)。

怎麼疊也有一手依據:手冊 p.90「both are in place, only **one-eighth** of the industry
produces pollution」——1/2 × 1/4 = 1/8,**數字是相乘的**(手冊用的字是 "cumulative",
但相加得不到 1/8,數字贏字面)。有一支測試把這條算術釘在 `PollutionEighths` 上。

**領袖技能現況(2026-08-10):26 項有 remake 消費端。** 仍未接的只剩戰術官(原版自己就沒實作)；
Diplomat 獨立接受門檻仍沒有可證實的 runtime 公式；Famous 招募機率已由 `sub_9781D` 補證，**沒有一項是卡在「找不到入口」了**。

## 工作方式(使用者定案)
- go/ebiten 參考路徑 = `~/master-of-maigc/repo`(魔法大帝繁中化,patch 疊 kazzmir/master-of-magic 引擎)
- **不用多代理 workflow**;翻譯一組一組慢慢做(單代理逐項,使用者可隨時審閱)
- 每輪更新 GitHub(遠端 `main`,已設 upstream)
