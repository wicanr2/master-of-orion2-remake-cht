# Update_Player_Ship_Designs_ 靜態稽核

日期：2026-08-24

## 問題

`Auto_Design_Ship_` 已證實會產生六筆 blueprint，但舊文件尚未回答科技完成時是完整覆寫玩家
設計、只換自動設備，或會連現役艦一起免費改裝。本輪追查 `Update_Player_Ship_Designs_` 的
caller、模式分支、真人／AI 分支與直接 helper，避免把 template 更新、建造中艦艇和殖民地
`REFIT` 混為一談。

## 輸入與工具

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫 SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4，處理器 `metapc`
- 位址基準：IDA linear，DOS/4GW LE image
- 探針：`tools/ida/audit_update_player_ship_designs.py`
- 操作：正式 `.i64` 唯讀掛載後複製到一次性 `/tmp`；不改名、不套型別、不保存正式資料庫。

外部導覽名稱採 `symbols_fixed.tsv`；`symbols_full.tsv` 在此區仍相鄰錯位。所有結論保留 raw
函式名、位址與指令，不把外部名稱本身當成證據。

## 已證實

### root 與 caller

- raw `sub_57112 @ 0x57112..0x57197`；修正索引名稱為 `Update_Player_Ship_Designs_`。
- 三個 caller：
  - `0xE4307`，位於 raw `sub_E4204 @ 0xE4204`，修正索引為 `Player_Gets_Tech_App_`；
  - `0xE5ABB`，位於 raw `sub_E5832 @ 0xE5832`，修正索引為 `Twiddle_Initial_Homeworlds_`；
  - `0xFEC45`，位於 raw `sub_FEC20 @ 0xFEC20`，修正索引為 `Make_Player_Into_AI_`。
- root 的 `eax/ax` 是 player 索引；`player*0xEA9` 與 `dword_197F98` 再次證實同一 player record。

### 模式與真人／AI 分支

1. `0x5711D` 比較 `byte_199CB4 == 1`。成立時呼叫 raw `sub_574A1 @ 0x574A1` 後返回；修正
   索引名稱為 `Update_Player_Strategic_Ship_Designs_`。
2. 非戰略分支讀 `player+0x28`：
   - `==100`（真人標記已有其他獨立 consumer 證實）時，依序呼叫 raw
     `sub_572AB @ 0x572AB`、raw `sub_57197 @ 0x57197` 後返回；
   - 非 100（AI）時先呼叫 `sub_57197`，再以 `ecx=0..4`、stride `0x63` 呼叫
     `Auto_Design_Ship_ @ 0x616A5`。因此 AI 完整重建 hull 0..4，這條迴圈不含 hull 5。

### 真人設計更新保留玩家武器

raw `sub_572AB @ 0x572AB..0x573C7`（修正索引 `Update_Automatic_Design_Equipment_`）逐 hull
0..5：

- `Best_Armor_ @ 0x5685F` 的結果寫 design `+0x16`；
- `Best_Warp_Drive_ @ 0x56726` 若不同於 design `+0x13`，先依 `byte_17FE90` 調整 design
  `+0x60`，再寫 `+0x13`；
- `Get_FTL_Speed_ @ 0x575D6` 結果寫 design `+0x14`；
- `Cost_Of_Design_ @ 0x6B577` 結果寫 design `+0x5E`；
- 沒有清空 99-byte record，也沒有呼叫武器／特殊裝置 helper。

所以真人取得科技時是更新自動裝備與衍生成本，不會把玩家手動武器／特殊裝置整筆洗掉。

### 建造中艦艇

raw `sub_57197 @ 0x57197..0x572AB`（修正索引 `Update_Equipment_On_Ships_Being_Built_`）只有在
player `+0x24 == 0` 時掃描 ship record；篩選 owner 等於該 player 且 raw status 為 4 或 6，
更新 record `+0x13/+0x14/+0x16`，並以 `byte_17FE90` 調整 `+0x60`。它不是掃描所有現役艦，
也不是殖民地 `REFIT`。

### 戰略模式六筆重建

raw `sub_574A1 @ 0x574A1..0x57566` 逐 hull 0..5：保存 design `+0x5C`、以 99 bytes 清除舊
record、呼叫 raw `sub_57871 @ 0x57871`（修正索引 `Auto_Strategic_Design_Ship_`），再依 hull
更新 `+0x5C`。hull 0..4 以每 hull 八個值循環，hull 5 固定寫 `0x2B`。`+0x5C` 的正式 UI
語意本輪未升格。

## 強推論與未知

- **強推論**：design／ship `+0x13/+0x14/+0x16` 分別是 warp drive、FTL speed 與 armor；
  helper 名稱、值比較與同欄讀寫互相支持，但本輪沒有重新追全部 UI label consumer。
- **未知**：ship raw status 4／6 的完整 enum 名稱，以及 design `+0x5C` 的正式欄位名。
- **資料模型缺口**：remake 尚無戰略／戰術開局模式、AI 六筆設計庫或「建造中實體 ship record」。
  殖民地建造佇列只保存建造項，完成時才建立 `Ship`。

## Remake 對應

- 真人六筆 `ShipBlueprint` 在取得新科技應用後，只把 `Armor` 更新為目前最佳已解鎖裝甲；
  武器、特殊裝置、mods、arc、ammo、尾端 raw mounts 與 `SpecialIDs` 均保留。
- remake 的引擎階由玩家已知科技即時計算；建造中艦艇尚未實體化，因此無需另寫 pending ship，
  完工時自然讀當下科技／設計。這是資料模型對應，不宣稱已重現 raw status 4／6。
- 戰略模式與 AI 完整重建留在 parity 矩陣，不以真人 tactical branch 代替。

