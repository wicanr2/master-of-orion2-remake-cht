# 殖民地總覽文案與版面稽核（2026-08-27）

## 證據分級

- **已證實**：`COLSUM.LBX#0` 是殖民地總覽背景；表格有名稱、農夫、工人、科學家、建造欄與九個可視列。底部排序列位置已由英文資產字墨量測修正。
- **已證實（手冊 p.38–40＋LBX 幾何）**：下方四格依序為 Planetary Info、Production Info、Mini Map、Empire Summary；第三格的透明底星點是原版縮圖底圖，不是解碼錯誤。
- **強推論／remake adapter**：目前懸停列選擇、Planetary／Production 動態文字與 Empire Summary 五行採現有 session 投影；原版完整 draw consumer、逐字與數值格式尚未由 IDA 閉合。
- **已證實（remake source）**：總覽仍有 21 個 `tr`；氣候、重力、礦產、大小共 23 組雙語名稱及 unknown fallback 也內嵌在 Go。表格與 `postDraw` 只限寬，沒有垂直安全框。

## 本輪邊界

固定玩家文案與環境顯示名移至 `ui.json`，Go 只保留 enum 到穩定 key 的映射。維持九列、既有熱區、懸停及數值來源，不把手冊／adapter 證據升格成原版逐字 parity。

runtime 字型量測另證實 7px 繁中字墨仍高 16px，原本建造主列與 `y+13` 已建副列必然重疊；兩者改為同一列格式，主要建造資訊置前並以省略號收束。
同理，Planetary Info 舊六列×14px 也無法落入 89px 面板；移除與上表重複的殖民地序號後，三個文字面板統一為五列×17px。
