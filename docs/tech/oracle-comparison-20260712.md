# 原版 oracle 對照記錄(2026-07-12)

> 來源:archive.org 線上 DOS 原版 MOO2 v1.40b23(`msdos_Master_of_Orion_II_-_Battle_at_Antares_1996`),
> 透過 Chrome 實機開啟逐畫面截圖,對照 remake 現況與使用者實測 issue(`issues/20260712-issues.md`)。
> 用途:把「跟原版差在哪」落成實據,決定修哪些、怎麼修。邊測邊補。

## 對照方法
- 原版:archive.org DOSBox 模擬器(音效 SoundBlaster;音訊在此環境靜音,音樂只驗機制不驗聽感)。
- remake:`-gamegallery` 產出的 8 張當前畫面(scratchpad/remake-gallery)。

---

## 逐畫面對照

### 主選單(MAIN MENU)
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 六按鈕 | CONTINUE / LOAD GAME / NEW GAME / MULTI PLAYER / HALL OF FAME / QUIT GAME | 同六項(座標 0px 對齊 mainmenu.cpp) | ✅ 相符 |
| **無存檔時 CONTINUE/LOAD GAME** | **灰階停用**(反白只有 NEW GAME 起) | 目前點 Continue 靜默無反應(issue #2) | ❌ **應對齊**:無存檔時 Continue/Load 要 disable |
| 版本切換 1.3/1.5 | 原版無(單版) | remake 自建鈕 | remake 加值,保留 |

**→ issue #2 結論**:原版 Continue 無存檔時**本來就不能按**(灰階)。remake 應:①無存檔時 Continue/Load 顯示為停用、不可點;②Load Game 點下去開「存檔選擇畫面」(原版 Load Game Window)。使用者「點繼續沒出現選存檔」是因為 remake 沒 disable、也沒存檔選單。

---

### 新遊戲設定(NEW GAME)
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 上排 | DIFFICULTY / GALAXY SIZE / GALAXY AGE | 同 | ✅ |
| 下排左框 | **PLAYERS**(對手數,顯示「5 / 5 Players」) | 標成 **RACE**(點它進種族選擇) | ❌ **忠實度差異**:原版此格是 PLAYERS 數量,RACE 是 Accept 後的獨立畫面 |
| 下排中框 | TECH LEVEL(Average) | 同 | ✅ |
| 三勾選 | TACTICAL COMBAT / RANDOM EVENTS / ANTARANS ATTACK | 同 | ✅ |
| 底部 | CANCEL / ACCEPT | 同 | ✅ |

**→ 新發現(非使用者列的 issue)**:remake 把原版的「PLAYERS(對手數量選擇)」格挪用成「RACE」入口,少了「選對手數量」這個原版設定。建議:把該格正名為 PLAYERS + 加對手數量選擇,種族選擇維持 Accept 後獨立畫面(remake 已有 raceSelect)。

---

### 種族選擇(SELECT RACE)
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 版面 | **肖像在左**大圖 + 右側 **13 族 2 欄按鈕** + Custom | **族名 1 欄清單在左** + 肖像在右(自繪畫面) | ◐ 版面左右相反、按鈕改清單 |
| 族數 | 13 族 + Custom | 13 族 + 自訂種族 | ✅ 數量相符 |
| 肖像/描述 | 原版逐族肖像(hover 換圖) | 用原版肖像 + 中文描述 | ✅ 素材原版 |
| 標題 | SELECT RACE(右上) | 選擇你的種族(置中) | ◐ 位置不同 |

**→ 發現**:remake 種族選擇是**自繪滿版畫面**(非原版 RACESEL.LBX 版面),肖像素材對但排版左右相反、按鈕格改成文字清單。與使用者「用原版 LBX、不自創」原則有落差。優先度中(功能可用、肖像正確),要更忠實需照原版版面重排(肖像左/2欄按鈕右)。原版 hover 族名即換肖像,remake 是點選清單換。

### 命名 + 旗幟(NAME / SELECT BANNER COLOR)
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 畫面數 | **兩個獨立畫面**:先命名(文字輸入)→ 再 SELECT BANNER COLOR | **合併成一頁**(命名 + 旗色同屏) | ◐ 流程合併 |
| 旗幟呈現 | **真旗幟圖形**(帶徽記,4×2 共 8 面) | **純色方塊**(8 色小方格) | ❌ 用色塊非原版旗幟圖 |
| 顏色 | 紅/金/綠/銀/藍/棕/紫/橘 | 紅/黃/綠/藍/白/紫/橘/棕(白 vs 原版銀,順序略異) | ◐ 色盤近似 |

**→ 發現**:remake 把原版「命名」+「SELECT BANNER COLOR」兩畫面合併,且旗幟改用純色方塊(原版是帶徽記的旗幟圖形,資產應在 LBX)。要更忠實需拆回兩畫面 + 用原版旗幟圖。優先度中(功能可用)。

## 新遊戲流程小結
四張都能玩、素材多為原版,但 remake 有系統性「自繪簡化」傾向:PLAYERS 格挪作 RACE、種族選擇左右相反、命名旗色合併 + 色塊。**這批是「像不像原版版面」的中優先項,非可玩性阻塞。**

### 星系主畫面(GALAXY / GAME)
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 工具列 | COLONIES PLANETS FLEETS ZOOM LEADERS RACES INFO **TURN** | 殖民地 行星 艦隊 縮放 領袖 種族 情報 **結束回合** | ✅ 順序/翻譯相符 |
| **回合鈕字面**(issue #3-1) | **TURN** | 原為「回合」→ 已改「結束回合」(使用者偏好,比直譯清楚) | ✅ 已修 |
| 右側 5 格狀態欄 | 國庫50/+12 · 指揮 **+3(5)** · 食物+0 · 貨運+0(0) · 研究none | 稅15%/50 · 指揮 **−2(紅)** · 食物0 · 貨運0 · 研究(領袖臉) | ⚠ **指揮開局原版正值、remake 負值**,待查 |
| 星圖密度 | 密(Medium 滿是亮星) | 稀(4~5 顆明星 + 少數暗點) | ⚠ 生成密度/星系大小疑似有差 |
| 研究狀態位置 | 右欄「none」格 | 疊左上角地圖 | ◐ 位置不同 |

**→ 發現(非使用者列的,較重要)**:①remake 開局**指揮點數 −2 紅字**,原版是 +3(5) 正值——remake 開局艦隊/供給與原版對不上,可能開局就赤字(接續先前 SAVE10 指揮校準,仍未完全對齊)。②remake **星圖密度明顯比原版稀**——同 Medium 設定下原版星星多很多,remake 生成的星數/亮度偏少。兩者都需進一步對照數值(原版 Medium 星系幾顆星?)。

### 星系點擊行星視窗(issue #6)✅ 找到修法
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 點星後 | 開「**Star System <名>**」**彈窗**(行星軌道圖) | 側邊資訊面板(postDraw 繪) | ◐ 呈現不同 |
| **關閉方式** | 彈窗右下有明確 **`CLOSE` 按鈕** | **無 CLOSE / 無關閉路徑**(issue #6:關不掉) | ❌ **remake 缺關閉鈕** |
| 附加 | 左下有系統改名文字欄 | 無 | ◐ |

**→ issue #6 修法明確**:原版星系彈窗有 **CLOSE 鈕**。remake 點星顯示資訊面板卻沒有任何關閉方式,所以「無法消失」。修:①加 CLOSE 鈕(對齊原版彈窗右下),或 ②點視窗外 / 再點同星 → 取消選取(`SelectedStar = -1`)。建議兩者都做(CLOSE 鈕忠實 + 點外關閉方便)。

### 領袖/軍官(LEADERS,issue #4)
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 分頁 | Colony Leaders / Ship Officers | 同 | ✅ |
| **開局領袖槽** | **4 格全空**(開局無領袖,須雇用) | 空(使用者以為 bug) | ✅ **remake 忠實**——原版開局也空 |
| 底部鈕 | HIRE / POOL / DISMISS / RETURN | 同 | ✅ |
| 右側 | 星系小圖 + 左右箭頭 + minimap | remake 現況待補 | ◐ |

**→ issue #4 結論(重大翻案)**:
- ①「領袖空白無圖」開局**忠實**(原版也空,須雇用),非 bug。
- ②**原版開局 POOL 按鈕是「暗/停用」的**(使用者實測),只能按 RETURN——代表**原版開局本來就沒有可雇領袖,領袖隨回合陸續出現在池中**。所以「無人可以僱用」在開局**也是忠實的**。
- ③**反而要注意 remake 別過度**:remake 補了 HERODATA 後若**開局就塞滿可雇傭兵**,會比原版**不忠實**(原版開局 POOL 空/停用)。remake 應對齊:開局 POOL 停用 / 空,領袖依 `advanceMercOffers` 隨回合出現(需查 remake 現行是否開局即滿池)。
- ④HERODATA.LBX 打包仍需修(否則池**永遠**空,連後續回合也沒領袖)——已修 LBX_LIST。
- ⑤殘留 #4b:已雇領袖**肖像渲染**仍是 TODO。

### 情報 INFO(issue #5)✅ 完全解開
| 項目 | 原版 | remake | 判定 |
|---|---|---|---|
| 結構 | **單一畫面 + 5 分頁**(History Graph / Tech Review / Race Statistics / Turn Summary / Reference),點分頁**切右側面板內容** | 只印 5 個靜態標籤,僅「Tech Review」有熱區 | ❌ remake 未做分頁切換 |
| **Tech Review** | 切到**科技總覽清單分頁**(右側面板內) | **錯接到 `research()` 選研究畫面**,點一下即返回星系 | ❌ **issue #5-2 根因** |
| **History Graph** | 可按,顯示國力/分數折線圖 | **沒接熱區** | ❌ **issue #5-1 根因** |
| Race Statistics / Turn Summary | 可按,各自分頁 | 未接 | ❌ |
| 左下 | **淨收入分解**(INCOME 15BC / BUILDINGS 20% / FREIGHTERS / SHIPS / SPIES / TRIBUTE / LEADERS 條狀圖) | 無 | ◐ remake 缺收入分解 |

**→ issue #5 修法**:INFO 應是「單畫面 5 分頁切右側內容」的檢視器,不是跳板。
- **#5-2**:Tech Review **不該跳 `research()`**——應在 INFO 內顯示「科技清單」分頁(語意錯置)。最小修:別再連到 research-select 讓人以為「進去就退」(改顯示科技清單面板,或暫停用並標未實作)。
- **#5-1**:History Graph 接上(國力折線圖)或明標未實作、停用不可按(別讓玩家以為壞了)。
- 完整版:INFO 做成 5 分頁檢視器 + 淨收入分解(較大,屬 UI 增強)。

### 事件黃框(issue #3)— remake 渲染問題,直接查 code
原版事件是對話窗;remake 的「黃色框」是事件文字擦底/邊框 artifact,屬 remake 渲染面,不需原版截圖即可查修(未觸發事件截原版,列 code 查)。

---

## 整合修復計畫(依 oracle 對照結果排序)

### A. 打包修復(script 已改,待重打包 AppImage)
- **#1 沒音樂 / #4a 無資料**:`LBX_LIST` 已補 stream/streamhd/sound/herodata ✅(待重打包才生效)。

### B. 明確 bug(code 修,高優先)
- **#3-1 回合鈕**:「回合」→「結束回合」✅ 已改譯文(原版字面 TURN,使用者偏好清楚版)。
- **#6 星系視窗關不掉**:加 CLOSE 鈕(對齊原版彈窗)+ 點窗外/再點同星取消選取。
- **#5-2 Tech Review 進去就退**:info 的「tech」**別再跳 `research()`**;改顯示科技清單或暫停用標未實作。
- **#5-1 History Graph 不能按**:接上或停用標未實作(別讓玩家以為壞了)。
- **#3 事件黃框**:查 remake 事件顯示渲染,去掉多餘黃框。
- **#2 Continue 無反應**:無存檔時 Continue/Load **停用**(對齊原版灰階);Load Game 做存檔選擇畫面。

### C. 忠實度對齊(中優先,非可玩性阻塞)
- 新遊戲 **PLAYERS 格正名 + 對手數量選擇**(remake 誤挪作 RACE 入口)。
- 領袖**開局 POOL 停用**(對齊原版;防 remake 補 herodata 後開局塞滿池)。
- #4b 已雇領袖**肖像渲染**。
- 種族選擇版面(肖像左/2欄按鈕右)、命名旗色拆兩畫面+原版旗幟圖、INFO 5 分頁檢視器+收入分解——**較大 UI 工程**,屬「更像原版」。

### D. 需再查證的數值疑點(oracle 抓到但需深查)
- **指揮點數**:remake 開局 **−2 紅字** vs 原版 **+3(5)**——開局艦隊/供給對不上,查 remake 開局艦隊組成。
- **星圖密度**:remake 明顯比原版稀——查星系生成星數 vs 原版 Medium。

### 修復順序建議
1. 重打包(解 #1/#4a)+ B 組明確 bug(#6/#5-2/#5-1/#3/#2)——一輪修完再打包給使用者複測。
2. D 組數值疑點深查(指揮赤字、星圖密度)。
3. C 組忠實度對齊(依使用者優先序,較大工程分批)。
- 情報 INFO(issue #5-1 歷史圖表、#5-2 科技總覽)
- 星系主畫面(issue #3 事件黃框、#3-1 回合鈕、#6 星系狀態視窗)
- 領袖/軍官(issue #4 肖像 + 雇用)
- 情報 INFO(issue #5-1 歷史圖表、#5-2 科技總覽)
