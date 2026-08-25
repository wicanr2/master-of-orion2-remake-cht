# AI 艦型藍圖與更新鏈靜態稽核

日期：2026-08-25

## 證據來源

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4，處理器 `metapc`
- 位址基準：IDA linear，DOS/4GW LE image
- 非破壞性探針：`tools/ida/audit_auto_design_ship.py`

正式 `.i64` 唯讀掛載後複製到一次性 `/tmp` 工作目錄。探針只匯出函式邊界、原始名稱、
指令、bytes 與交叉參照，沒有改名、套型別或保存資料庫。

## 已證實

### 初始設計角色序列

raw `sub_54FBF @ 0x54FBF..0x55138` 直接呼叫 `sub_616A5 @ 0x616A5`
（外部修正索引：`Auto_Design_Ship_`）：

- `0x55083 xor ebx, ebx` 後於 `0x55088` 生成 hull 0，因此 role 0。
- AI 分支在 `0x54FEF` 將 `ecx=0`，`0x5506E..0x5507A` 令第一個迴圈 role 為 1。
- `0x550A9..0x550C6` 令 hull 1..4 與 role 1..4 同步遞增。
- `0x550DB` 以玩家記錄 `+0x28 == 100` 區分真人；非真人在 `0x550F0..0x550F9`
  生成 hull 5／role 5，真人則清空該筆。

因此一般 AI 初始設計的可回查序列是 hull 0..5 對 role 0..5。這不是依 hull 猜出的
remake 政策。

### 科技完成後的 AI 更新

raw `sub_57112 @ 0x57112..0x57197`（外部修正索引：
`Update_Player_Ship_Designs_`）：

- 真人分支 `0x57147..0x57155` 只更新自動裝備與建造中艦艇。
- AI 分支先呼叫 raw `sub_57197`，再由 `0x57161 xor ecx,ecx` 開始迴圈。
- `0x5716C movsx ebx,cx` 把目前 hull 索引同時作為 role；`0x57187` 呼叫
  `Auto_Design_Ship_`，`0x5718C..0x57190` 明確迴圈至 hull 4。
- hull 5 不在此科技更新迴圈內。

結論：AI 必須持久保存至少五筆一般艦型藍圖，科技完成時以 role 0..4 重建；抽象
`FleetStrength`、殖民地 `AUTO BUILD` 或按回合生成的 `genEnemyFleet` 都不是這條資料鏈。

### 相鄰 helper

raw `sub_56460 @ 0x56460..0x56484` 在 `0x56465` 固定 `ebx=5`，於 `0x5646B`
只生成 hull 5，隨後以 `off_1800EE` 覆寫名稱。它是 hull 5 的短 helper，不是一般 AI
hull 0..4 更新器。

## 強推論、未知與停止線

- **強推論**：一般 AI 可建造艦艇應引用上述持久藍圖，而非每次戰鬥臨時生成另一套設計。
- **未知**：原版 AI 在多筆藍圖間的精確生產評分、殖民地生產佇列選擇及拆分多支艦隊政策。
- remake 在未知選擇器上可採有標示、可重現且不偽稱 exact 的最小政策；不得倒退為只有
  一個軍力整數。若後續 IDA 證據解出選擇器，再替換該窄政策。

