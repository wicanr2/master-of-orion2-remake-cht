# 格子戰術控制列文案與入口稽核（2026-08-26）

## 輸入與工具

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫：`Orion2.exe.i64`，SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4；位址為 DOS/4GW image 的 IDA linear address。
- 非破壞性匯出器：`tools/ida/audit_tactical_weapon_controls.py`。輸出保留 raw 函式名、
  邊界、caller、指令 bytes 與資料參照，不在 IDA 資料庫做推測性改名。

## 已證實

- 實際 `COMBAT.LBX` inventory 證實 asset 0 是 640×129 控制列，asset 11 是其 256 色
  palette provider；按鈕英文烘在 asset 0。資產形狀與調色盤驗證詳見
  `docs/tech/tactical-combat-assets.md`。
- `sub_2F4EE @ 0x2F4EE..0x30052` 有至少 17 個直接 caller，包括 `0x2C5E0`、
  `0x2C6AC`、`0x33085`、`0x33C6D` 與多個戰鬥更新端；它是主戰鬥繪製鏈的高扇入根。
- `sub_34921 @ 0x34921..0x34AB6` 有 caller `0x30005`、`0x30048`、`0x33959`、
  `0x34C7E`、`0x34CA7`，並呼叫既有 bitmap／字型 helper。它與外部導覽名
  `Draw_Fake_Combat_Buttons_` 相符，但名稱本身不作證據。
- 外部符號把 `Combat_Screen_` 放在 `0x478A2`；IDA 在該位址沒有函式。相鄰
  `sub_478A3 @ 0x478A3..0x478FC` 由 `sub_47939` 內的 `0x47DE8` 呼叫，只讀 stride
  `0x169` 記錄的 `+0x13E`、`+0x150`、`+0x151`、`+0x15E`、`+0x15F`、`+0x160`、
  `+0x165` 旗標並回傳布林值；不可把它當完整戰鬥畫面或控制列入口。
- 戰術外層仍是 `sub_47939 @ 0x47939..0x48EDF`；WAIT／DONE 與 Ship Initiative 的
  回合消費端證據見 `docs/re/game-menu-popup-ui-text-audit-20260826.md`。

## 強推論與未知

- **強推論**：`sub_2F4EE` 是主要戰鬥畫面繪製，`sub_34921` 是其中的假按鈕／失效按鈕
  視覺 helper。函式拓樸、外部導覽名與 COMBAT asset 三者吻合，但本輪不以名稱覆蓋 raw 定位。
- **未知**：七顆按鈕各自的原版 widget ID、原版 SCAN／BOARD 狀態提示與戰報逐字字串。
  `dword_1A1244`（外部名 `_combat_button_strings`）只有 `sub_CC81C @ 0xCC96F` 的直接 writer；
  指標間接 consumer 未由 xref 自動恢復，不能把該名稱當作完整字串表證據。
- remake 的掃描摘要、登艦結果、AUTO 停止原因及模式提示是玩家可操作的現代轉接文案。
  它們須外部化並維持雙語、格式佔位符與安全框，但不宣稱是原版逐字翻譯。
- OPTIONS 原版會進入設定流程；remake 已有 13 列 SETTINGS 畫面，但目前轉場會返回星圖，
  尚未保存並返回同一場戰術戰鬥。因此戰鬥內 OPTIONS 只顯示準確限制，不再錯稱設定畫面不存在。

## Remake 對映

- 按鈕幾何與熱區共用 `barButtonsCHT` 的七個中心；資料結構只保存 action ID 與文案鍵。
- 正版英文模式露出 `COMBAT.LBX#0` 烘字；繁中模式及缺資產 fallback 由
  `assets/i18n/ui.json` 提供標籤。
- 所有控制列狀態、掃描、登艦與撤退訊息由 `tactical.*` 語意鍵取得；Go 不保存玩家句子。

