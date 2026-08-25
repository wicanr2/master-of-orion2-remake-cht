# 研究應用選擇與授予時序稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、IDAPython
- 位址基準：IDA linear，DOS/4GW LE object #1
- 匯出腳本：`tools/ida/audit_research_breakthrough.py`
- 方法：唯讀開啟 `.i64`，保留原函式名、位址、bytes、運算元與 caller／callee；未修改資料庫。

## 問題與舊模型

remake 原先在研究突破後把主題設成 `PendingChoice`，先暫時解鎖第一個科技，再讓玩家於
回合摘要前改選。這個模型符合「每主題擇一」的結果，卻沒有原版選擇時序證據。

## 已證實

### 玩家在研究開始前便選定 application

`sub_10DC12 @ 0x10DC12` 是 `TECHSEL.LBX`／`SCIENCE.LBX` 研究選擇畫面：

- `0x10DCC9` 先讀玩家目前 field `player+0x321`。
- 點擊一個可研究項目後，`0x10E389..0x10E39D` 將該 row 的 field 寫回 `player+0x321`。
- 緊接的 `0x10E39D..0x10E3AA` 由 row application table 取 application ID，寫回
  `player+0x322`。
- 因此 `+0x322` 是「目前正在研究、突破後將取得的 application」，不是突破後才產生的選擇。

入口 `sub_10DB69 @ 0x10DB69` 以 mode 0 呼叫此畫面；`sub_10DBCE @ 0x10DBCE` 以 mode 1
呼叫同一畫面。兩種模式共用上述 field/application 寫回。

### 突破消費已選定的 application

`sub_E4410 @ 0xE4410` 由 `Check_For_Research_Breakthrough_ @ 0xE44E0` 成功路徑呼叫：

- field `<75` 且非全授予分支時，`0xE44CD..0xE44D6` 讀 `player+0x322`，再呼叫
  `sub_E4204` 授予該 application。
- field `>=75` 是 Hyper-Advanced，增加對應 level byte，不走一般 application 選擇。

### Creative 與固定全授予 field

`0xE443C..0xE4451` 在 `player+0x8B5 != 0` 時啟用全授予；`0xE4481..0xE44A8` 掃描該
field 的 application table，對每個 status 1 application 呼叫 `sub_E4204`。手冊的 Creative
語意與此 consumer 一致，因此 `+0x8B5 = Creative` 為已證實。

field raw 22、23、28、29、55、57 亦被硬編為全授予，不受 Creative 限制。

### Uncreative 在可選集合形成時隨機限縮

`sub_E4204 @ 0xE4204` 在取得 application 後，若 `player+0x8B4 != 0`，於
`0xE43D4..0xE43EE` 對該 application 所屬 field 呼叫 `sub_E408F @ 0xE408F`。
`sub_E408F`：

- 掃描 field application table；status 1 代表已有可選項，status 0 代表候選。
- 若尚無 status 1，使用 `Random(candidateCount)` 的 reservoir-style 選擇，最後只把一個候選
  application 設為 status 1（`0xE4111..0xE411F`）。

結合手冊的 Uncreative 語意，`+0x8B4 = Uncreative` 為已證實。這表示亂數限縮發生在
下一個 field 的可選集合形成階段，不是突破後才從完整清單任選。

### AI 也先決定 application

`sub_DC288 @ 0xDC288` 由 `sub_DCA69` 呼叫；需要新研究時先呼叫 `sub_FD335`，再於
`0xDC2CC` 寫 `player+0x322` application，`0xDC2D2..0xDC2DE` 由 application table 反查 field
並寫 `player+0x321`。AI 與真人都在投入研究前具有 field/application 配對。

## `sub_E4204` 已證實的副作用邊界

`sub_E4204` 至少執行：

1. `application status`（`player+0x117+application`）寫 3，並設 `player+0x323=1`。
2. 少數具名 application 更新政府／runtime 欄位或觸發全星系資料更新。
3. 呼叫 `sub_E2D72` 重算玩家衍生科技狀態。
4. 若取得的是目前 application，尋找同 field 下一個可用 application；field 已清空時呼叫
   `sub_E401D` 推進 field 並把研究進度 `player+0x1EB` 清零。
5. Uncreative 為後續 field 形成單一可用 application。
6. 部分 application 觸發 `sub_57112`、`sub_10038C`、`sub_E5296` 等下游更新。

remake 不複製原版 runtime record；等價消費由 `CompletedTopics`／`ChosenTech`、科技旗標同步、
艦艇設計更新及各規則查詢完成。尚未逐一對齊的特殊 application callback 必須保留為剩餘工作，
不能以資料模型不同宣稱完整 parity。

## 勘誤結論

舊文件與程式的「突破後才擇一」時序被 `sub_10DC12` 寫端及 `sub_E4410` 讀端直接否定。
remake 應在選定研究 field 時一併選定 application；突破只授予已選定項，Creative 例外為全授予。
