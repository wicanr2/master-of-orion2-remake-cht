# 銀河霸主 II remake 誠實現況

> 更新：2026-08-29。本文只描述目前狀態；剩餘工作的唯一活表是
> [`WORKLIST.md`](../WORKLIST.md)，原版證據邊界以
> [`docs/re/parity-matrix.tsv`](re/parity-matrix.tsv) 為準。

## 有日期的三項工程快照

2026-08-29 的 README 儀表板為：remake 功能 74%、原版玩法對齊 51%、發行驗證 33%。功能
比例由 40 個實作／交付母項的 26 完成、7 進行中、7 待辦計算；另有 22 個已完成 RE-only
證據切片，不混入功能分母。對齊比例由 41 列矩陣逐列重分的 9 已對齊、24 部分對齊、8 明確
未對齊計算；發行比例為三個獨立閘門通過一個。這些數字不可合成「總還原度」，詳細公式只在
README 維護，狀態改變時四份來源必須同輪更新。RE 閉合本身不等於 remake 已接線。

## 目前執行策略

先補齊 `docs/re/parity-matrix.tsv` 所列玩家可見玩法的 RE 知識庫，再討論規格與
Go／Ebitengine 實作。RE 閉合必須保留原始位址、輸入狀態、規則或資料表、玩家可見
消費端、證據等級與未知邊界；在此閘門關閉前，新發現只登記證據與 remake 差異，
不直接修改玩法程式或把推論寫成測試真值。

編譯器生成輔助函式、C runtime／標準函式庫與 Windows API／平台內部行為不屬於
RE 完成分母或 remake 範圍，例如 stack probe、stack overflow check、SEH、`fopen`、
`fclose`、`fork` 及一般檔案／程序包裝函式。分析時只辨識並跨過這些 pattern；若其
呼叫邊界會改變玩家可見結果，才保存最小輸入、輸出、錯誤與時序契約。

## 目前可證實的成果

- Go／Ebitengine 重製版已有可遊玩的多帝國 4X 迴圈，包含新遊戲、星圖、殖民、
  研究、外交、間諜、領袖、艦隊、兩條戰鬥路徑、議會與勝利條件。
- 主要畫面已接入原版 LBX 美術、繁體中文疊字、存檔往返、畫廊抓圖與
  STREAM／STREAMHD／SOUND 解碼。技術測試可驗證音訊非靜音；人耳逐曲聽感仍需
  有實際音訊輸出的外部環境。
- 議會人口票數與召開排程已根據靜態反組譯證據修正；候選人、棄權與外交投票精確公式仍待閉合。
- 客製種族選項已寫入 `CustomRaceTraits` 並參與存檔與玩法消費；AI 也已
  保存 `RaceIndex` 並在研究、地面戰、殖民與外交部分路徑消費種族特性。
- Stealthy Ships 的星圖、AI、自動設計、profile 與科技估值 direct consumer 已閉合；IDA 證據
  顯示三種匿蹤裝置不向快速結算提供通用數值，但其 bitfield 會進入格子戰術記錄。現有資料流
  強推論否定 trait 與裝置的戰鬥等價；raw 6／23 的格子狀態機、10 回合倒數、目標合法性、
  +80 防禦與飛彈 50% miss 已追回，raw 31 不進該戰術狀態機。
- 飛彈 ECCM／EMG／MV、魚雷 ENV／OVR／NR、ARM／FST、ECM／慣性防禦、掃描科技、
  艦員與 Helmsman 效果已有戰鬥消費端；快速結算與格子戰術仍必須分別驗證。

## 尚未能宣稱的事

- 「可遊玩」不等於「原版逐值對齊」。難度分支、完整回合結算、
  外交／間諜／領袖、原版 AI、事件／安塔蘭、戰鬥、艦艇設計與其餘客製種族
  仍有待追回的原版規則或只有 remake 近似。
- 歷史圖四項 350 格、動態除數與最終分數倍率已依 IDA 靜態證據接線；科技歷史 raw
  `player+0x224` 的唯一 writer、83 筆 topic 表與研究後記錄時序已閉合，完成主題成本重建為
  本版已證實。逐玩家殲滅星曆 `+0x1F2[target]` 的兩條戰鬥 producer 與唯一 writer 亦已閉合，
  但 remake 尚未保存 8×8 歸屬矩陣。Charismatic／Repulsive
  同化則已閉合為 240 點 raw 進度及精確倍增／減半分支。
- 原版 `Next_Turn_Calc_ @ 0x136B3..0x13822` 的 52 個直接呼叫、兩次殖民地 derived-state
  重算、主要條件閘門與回合外層時序已閉合；各子系統內部資料流仍須依 parity matrix
  分列追查。綠色單元測試只能證明 remake 內部自洽，不能代替原版 oracle。
- 官方五級 AI Growth／Food／Prod／Res／BC、Command Deficit 與 Spy 難度表已於 2026-08-26
  接入 AI 回合，且不作用於玩家；quarter 捨入、士氣／重力先後及 Spy 攻守共同注入尚無完整
  IDA 指令級證據，維持強推論。
- AI 常態研究已依 `All_AI_Tech_Select_ → sub_DC288 → sub_FD335` 改為一次 application 級
  估值抽選，raw profile 與研究亂數位置可存檔。AI 生產也已從全帝國單一造艦池改成逐殖民地
  建築產品與進度，接上 raw 1..48 可建 gate、難度濾門及加權抽選控制流；其中 raw
  2／4／5／6／7／8／10／12／13／15／16／17／19／20／21／22／23／24／25／26／27／28／29／30／31／32／33／34／35／36／37／38／39／40／41／42／43／44／46／47 已依直接解碼跳表與 typed priority gate、
  人口成長／主要人口種族、帝國食物差額、國庫／淨 BC 平方根因子及兵營星系壓力 context 使用完整精確分數；
  其餘分數區域、帝國配額、支援／戰鬥艦
  產品、艦隊、外交與其餘 state machine 尚未閉合，不能以這些切片宣稱完整原版 AI。
- AI↔AI 宣戰已接一般理由 23，以及 `sub_25DF1` 的政府理由 20、輪值敵意理由 68、
  食物赤字理由 113；後者由 AI 帝國食物結餘產生並保存 `player+0x7EC` 對應 streak。
  超空間亂流外層 gate 與跨維度免疫已接；`+0x60E` 已保留原版
  `.GAM` raw 並接上已證實 consumer；1.31 沒有直接 runtime writer。理由 22 的殖民破壞怨值
  producer／consumer 與 AI↔AI 實艦戰鬥傷亡鏈亦已閉合；`player+0x60F..+0x88F` 外交方向區
  也已建立逐欄資料形狀、初始化、producer／consumer 與未知邊界契約。remake 尚未接回完整
  方向模型與實艦戰鬥，因此仍不能宣稱完整原版 AI 已完成。
- 畫面已有大量原版座標證據，但動態文字測量、中文分頁、按鈕置中、hover 及
  英文模式回歸仍是活躍 polish。README 的艦艇空間資訊與多人說明列已於 2026-08-27 移回各自
  美術面板，並新增 panel containment 測試；這只證明兩個具體畫面，不代表其餘動態欄位已全庫
  掃描完成，也不使用「逐畫面像素對齊已全部完成」的總括斷言。
- 多人連線的現代化可玩核心與安全／NAT 擴充要分開驗收；後者不因本機 smoke test
  通過就視為公開網路已完整驗收。

## 證據紀律

- Windows API／Win95 平台內部行為不屬玩法 RE 範圍；只保留玩家可見契約，remake 採有標註的
  現代近似，不把作業系統內部實作列成 parity 缺口。
- DOS 主程式已證實使用 Watcom C/C++ 32-bit runtime＋DOS/4GW；Win95 主程式已證實屬
  Microsoft Visual C++ runtime 家族（PE linker 4.20），但兩者精確編譯器小版本仍未知。
  stack probe、x87 初始化、C++／SEH 例外處理等 compiler-generated helper 已建立 bytes／xref
  指紋並排除於玩法完整度之外；詳見
  [`compiler-runtime-fingerprint-20260827.md`](re/compiler-runtime-fingerprint-20260827.md)。

- 優先順序是原版執行檔立即數／交叉參照、官方手冊與 patch 資料、openorion2、
  LBX 形狀與有界原版實驗。
- 每項規則以「已證實、強推論、假說、未知」標示；原始位址、bytes、工具版本與
  位址空間不可被語意別名取代。
- 歷史翻案只保留在 [`docs/re/01-gap-report.md`](re/01-gap-report.md) 與 Git 紀錄；現況文件不保留
  過期斷言、刪節線或「先前寫錯」的長篇時序。
