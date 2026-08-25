# 隨機事件 9「超空間亂流」靜態稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、`tools/ida/audit_event_hyperspace_flux.py`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 方法：唯讀 `.i64` 交叉參照、指令與 operand reference 匯出；未改名、未改資料庫。

## 已證實：事件記錄

1. `sub_2230A` 的事件 9 分支位於 `0x22A6E..0x22A7A`，只把固定事件記錄的
   `+0x06` 年齡寫成 0；它不選帝國、殖民地或星系，是全銀河事件。
2. `sub_233FA @ 0x233FA` 是唯一直接讀事件 9 固定 record 的查詢器。record 狀態為
   2、4 或 6 時回 1，其餘回 0；直接 callers 分布於事件選擇、主畫面、航線操作、AI 移動、
   AI 殖民、艦隊計算及航行設定。
3. `sub_206A2` 的事件 9 分支位於 `0x20B66..0x20BAC`：只在狀態 2 消費；raw age
   大於 4 時擲 `Random(20)==1`，成功把狀態改成 5。共用尾端 `0x2121D..0x2123E`
   又在 raw age 大於 20 時強制改成 5，最後將 age 加一。
4. 因建立當回合 age=0 也會走一次 consumer，前五次 consumer 不擲解除骰；第六次
   （進入時 age=5）起每回合 5%，進入時 age=21 則不論骰值強制結束。
5. `sub_21371` 的事件 9 展示分支位於 `0x215AB..0x215D6`；狀態 5／6 使用共用
   GNN 結束展示鏈，規則端只需保存 active／ending 的玩家可見狀態，不需翻譯 UI driver。

## 已證實：玩家可見消費端

- `sub_9A07B @ 0x9A07B` 在建立航線操作狀態前查 `sub_233FA`；亂流 active 時直接返回。
- `sub_9D252 @ 0x9D252` 在兩個航線端點互動分支查 active；成立時播放訊息資產 319，
  重設航線操作狀態，不建立航行命令。
- `Move_All_AI_ @ 0xDBB29`、`sub_DBC5C` 與 `sub_DBCC8` 都在 AI 移動／目標鏈先查 active；
  但玩家記錄 `+0x8BC != 0` 時可繞過阻擋。
- `sub_E6CAA @ 0xE6CAA`（AI 殖民航行鏈）及 `sub_FF799 @ 0xFF799`（艦隊航行設定鏈）
  具有相同 `player+0x8BC` 例外。`sub_FF799` 在非例外帝國遇到亂流時寫入停止旗標後立即返回。
- `sub_EF827 @ 0xEF827` 在每回合艦隊計算入口查 active，證實亂流不只阻止新命令，也介入
  已存在航程的每回合計算。

## 強推論與 remake 投影

- `player+0x8BC` 是跨維度（Trans-Dimensional）能力：多個亂流 caller 只把它當免疫旗標，
  官方手冊又明載跨維度種族免受超空間亂流影響。此結論為「獨立手冊語意＋多 caller 資料流」
  的強推論；本切片不以推測改寫 IDA 名稱。
- remake 以 `PersistentHyperspaceFlux` 全局 record 表達 active 狀態；玩家與熱座的在途
  `Fleet.ETA`、AI 的 `FleetETA` 在 active 回合保持不變，且非跨維度帝國不得下達新航行命令。
  原版有逐艦停止旗標，remake 只有逐艦隊 ETA，故這是 typed projection，不宣稱 raw 逐艦相同。

## 版本差異與停止線

- 本輸入在 `Determine_Event_` 對彗星／怪獸的 `sub_233FA` 分支方向，與 patch 1.50
  `CHANGELOG_150.TXT`「Fixed hyperspace flux check for comet and monsters so they can't happen」
  對應到已知修正前行為。此切片只閉合亂流本身；彗星／怪獸候選的 1.31／1.50 差異另列
  profile 規格，不把單一 EXE 分支冒稱兩版本共同規則。
- 航線畫面訊息資產 319 的逐字內容、狀態 4／6 的網路展示時序不影響規則結果，不深挖。

