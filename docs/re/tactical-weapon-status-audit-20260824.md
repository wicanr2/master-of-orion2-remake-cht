# 戰術武器槽狀態稽核（2026-08-24）

## 問題

remake 已在格子戰術中自動派送八個武器槽，但玩家看不到各槽狀態，也不能讓單一槽略過下一次
齊射或持續停火。本輪只追玩家可見的武器狀態契約，不延伸到 Win95／滑鼠 driver。

## 輸入與工具

- `Orion2.exe` SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- `Orion2.exe.i64` SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- IDA Pro 9.4，映像 `ida-pro-9.4-ver3:py312-v1`
- 位址基準：IDA linear，DOS/4GW LE image
- 原版說明抽取：`cmd/moo2/embedded/i18n/help.tsv` SHA-256
  `e55e686d0a17e930f7f7d96f693aef24f0047c9c9d049f16ae6b513f1516623b`
- 非破壞性探針：`tools/ida/audit_tactical_weapon_controls.py`；原 `.i64` 唯讀掛載，容器內複本分析。

## 已證實

1. 原版說明明寫目前艦艇的武器以三色顯示：綠色可用、橘色略過下一次艦艇開火後恢復正常、
   紅色持續關閉直到玩家重新開啟；近迫防禦即使紅色仍可攔截來襲飛彈。右鍵點單一武器顯示資訊。
2. 外部符號索引的 `Draw_Weapon_Status_Display_` 對應原始函式 `sub_2E2CD @ 0x2E2CD`，函式邊界
   `0x2E2CD..0x2E98D`。它以艦艇 record stride `0x139` 取目前艦，武器槽自 `+0x52` 起、
   stride `0x0B`，迴圈上限八槽；`0x2E5B7`／`0x2E5BD` 與 `0x2E622`／`0x2E628`
   共同讀槽尾 `+0x5B`／`+0x5C` 決定顯示狀態。
3. 外部索引的 `Do_Combat_Turn_` 對應 `sub_42F7F @ 0x42F7F`。`0x44D65` 比較槽尾
   `+0x5C == 2`，`0x44D6B` 寫回 `+0x5C = 1`。這是原版 runtime 中明確存在的單次狀態恢復。
4. `Draw_Weapon_Status_Display_` 直接 caller 分布於 `0x2C57C`、`0x2C63E`、
   `0x2F7E3`、`0x2F8AE`、`0x2F91C`、`0x2FA38`、`0x338F4`、`0x33C61`、
   `0x3EFF8`；因此它不是孤立 helper。

## 強推論與未知

- **強推論**：runtime 狀態 `2 → 1` 對應手冊的橘色待命 → 綠色可用。手冊語意與消費端寫回
  相互支持，但本輪沒有推測性改名 IDA 欄位。
- **未知**：原版左鍵 widget 的確切按鈕 ID、三態每次點擊的循環方向，以及右鍵資訊彈窗的外觀與完整
  呼叫鏈。remake 採可重播的「可用 → 待命 → 關閉 → 可用」單鍵循環，右鍵则以現有有界戰術訊息列顯示逐槽資訊；兩者皆標為操作 adapter，
  不宣稱該循環順序已由反組譯證實。
- **後續已閉合**：2026-08-25 已由 tactical-point-defense-slots-audit-20260825.md
  補齊逐槽防禦消費端；紅色／橘色／綠色只控制主動齊射，不阻擋飛彈與戰機自動迎擊。

## Remake 對映與驗證

- `CombatShip` 保存每槽 `TacticalWeaponMode`，只存在本場戰鬥，不污染持久設計。
- UI 用綠／橘／紅顯示八槽；點槽循環模式，文字與熱區共用固定安全矩形。
- 右鍵槽位會顯示名稱、工作數、傷害上限、射界、彈藥與改造，不改變開火模式。
- 待命槽在同艦至少一個其他槽成功開火後恢復可用；關閉槽維持關閉。
- deterministic 測試固定 RNG，驗證傷害、彈藥、模式恢復、全關閉 fail-closed 與 640×480 幾何。

## 剩餘不確定性

原版右鍵彈窗的精確外觀／呼叫鏈仍未知；玩家可用的右鍵明細轉接已完成。PD 紅色例外已由
tactical-point-defense-slots-audit-20260825.md 閉合，不再列為本切片剩項。
