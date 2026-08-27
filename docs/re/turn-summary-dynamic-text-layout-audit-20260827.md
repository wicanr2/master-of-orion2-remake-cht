# 回合摘要動態文案與版面稽核（2026-08-27）

## 問題與證據來源

本輪確認 `turnSummary()` 的經濟、破產、研究與建造通知是否為原版逐欄復刻，並檢查動態列
能否安全落在 `TURNSUM.LBX#0` 的內容框內。

- 原版輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具與位址基準：IDA Pro 9.4；DOS/4GW image 的 IDA linear address。
- 既有審查：`info-subscreen-text-audit-20260826.md`、
  `serious-turn-summary-setting-audit-20260827.md` 與實際 `TURNSUM.LBX` 畫廊。

## 已證實

- 原版 INFO 子頁的回合摘要可見字串至少含 `MESSAGE SUMMARY AS OF SD:`、`Net Income:`、
  間諜、同盟、戰爭、條約與進貢欄位；現行獨立 `TURNSUM.LBX#0` 畫面的經濟拆分不是原版
  逐欄復刻。
- `byte_199BE8 @ 0x199BE8` 是 Serious Turn Summary 設定；`sub_FE0EA @
  0xFE0EA..0xFE250` 掃描 18-byte 報告記錄並決定是否開啟整張摘要。它不是逐行過濾器。
- 官方 help 將飢荒、叛亂及無力支付維護造成的資產處分列為 serious report 例子。

## 強推論、remake 適配與未知

- **強推論：**經濟總額、破產處分、研究與完工通知是現行 remake 回合結算的必要玩家回饋。
- **remake 適配：**淨工業／研究、食物／稅收與國庫拆分，及建造完成句式均是現行 UI 組裝；
  只可標成可操作摘要，不可稱原版逐欄或逐字 parity。
- **未知：**原版獨立 `TURNSUM.LBX#0` 動態列表的完整欄位來源、排序、分頁／捲動與逐字模板。
  本輪不以現行 Go 行為反向證明原版。

## 找到的實作缺陷

- 先前的 `LastBuilt []string` 由規則層直接拼中文句子，英文只能沿用中文；且地震刪殖民地時錯把它當成
  殖民地平行陣列按索引裁切。它應改為 typed `BuildNotice`，UI 再依 JSON 模板與建造名稱翻譯。
- 基礎四列後的 `yy` 可被破產、飢荒、叛亂、研究、多項完工、事件、安塔蘭與突襲無界累加，
  最終會侵入 y=324 的 `CLOSE` 按鈕。動態內容必須受固定安全框、最大行數與末行省略號約束。

## Remake 對映與驗證

- 固定句子與格式模板使用 `turnsummary.*` JSON 鍵；Go 只帶年度、產出、數量與 typed notice。
- 建造項目名稱由既有 `buildItemLabel` 翻譯，艦名與行星名維持動態資料。
- 內容列限制於背景局部座標 x=40..360、y=62..306；動態區自 y=168 起，任何輸入都不得進入
  y=324 的按鈕列。
- 來源掃描、格式參數、typed notice、最壞情境高度、全套測試及雙語 `06_turnsummary.png`
  抽樣共同驗收；畫廊不升格原版逐字證據。
