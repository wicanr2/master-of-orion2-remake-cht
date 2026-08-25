# 科技應用授予 callback 稽核（2026-08-25）

## 問題

原版 `sub_E4204` 授予科技應用後有哪些玩家可見副作用；remake 是否把原版的衍生快取重算誤當成必須持久化的新狀態。

## 證據基線

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4，既有 `.i64` 資料庫，`tools/ida/audit_research_breakthrough.py`
- 位址基準：IDA linear；DOS/4GW LE object #1
- 匯出契約：保留 `sub_*` 原名、線性位址、bytes、運算元與 caller／callee；未修改 IDB。

## 已證實

1. `sub_E4204 @ 0xE4204` 先把 application 狀態寫成 3，並設玩家科技更新旗標 `+0x323 = 1`。
2. application 24（`TECH_BATTLEOIDS`）會以 `sub_E412B(player, 14)` 檢查 application 14（`TECH_ARMOR_BARRACKS`）；允許時遞迴呼叫 `sub_E4204(player, 14)`。因此 Battleoids 會額外授予 Armor Barracks，且不應覆蓋 Astro Engineering 原先選到的另一項科技。
3. application 42／65／77／92 直接更新原版政府 raw 值。remake 的政府升級已由科技擁有權在 `internal/shell/assimilation.go` 動態消費，毋須保存原版 raw byte。
4. application 99／170／176／180 逐星呼叫 `sub_E5296 @ 0xE5296`。該函式重建星系 owner／殖民地防禦／科技 bitmask；它不是航速公式。remake 沒有這組 raw cache，而是由星系、殖民地、建築與科技狀態即時計算。
5. `sub_E2D72 @ 0xE2D72` 對玩家每座非前哨殖民地呼叫 `sub_E1D59 @ 0xE1D59`，再呼叫 `sub_E2D09 @ 0xE2D09`。前者重算殖民地食物、工業、研究、污染、士氣等衍生欄位；後者重算玩家與殖民地摘要。remake 的 `prepPlayerDerived`／`RunEmpireTurn` 已採來源狀態重算。
6. `sub_10038C @ 0x10038C` 對所有玩家呼叫 `sub_10034D @ 0x10034D`、`sub_56726` 與 `sub_57597 @ 0x57597`，回寫最佳引擎與 FTL 衍生值。remake 的航速與設計更新由科技擁有權即時計算，玩家／AI 設計另有既有更新入口。

## Remake 對映

- 新增可序列化的額外 application 集合，表示不屬於「該主題唯一選擇」的連帶授予；不能用改寫 `ChosenTech[TOPIC_ASTRO_ENGINEERING]` 表示 Armor Barracks。
- 所有科技擁有權 helper 先查額外集合，再沿用 `CompletedTopics`／`ExplicitChoice`／`ChosenTech` 的舊存檔語意。
- 研究完成、科技偷竊／餽贈、遺跡與開局授予共用 Battleoids callback；callback 必須冪等。
- 不複製原版星系、殖民地、玩家與引擎 raw cache，只測試其現有動態消費端。

## 剩餘不確定性

- 本切片沒有把 `sub_E1D59` 每個下游 helper 逐一重新命名；它們是已證實的衍生重算鏈，不影響 Battleoids 額外授予的玩家語意。
- 原版 application 狀態陣列可同時持有同主題多項科技；remake 既有 `ChosenTech` 是相容模型。新增集合只補足額外授予／多來源取得，不改寫既有存檔。
