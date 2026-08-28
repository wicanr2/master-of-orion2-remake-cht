# 戰略轟炸快速戰鬥鏈稽核（2026-08-24）

## 結論

- **已證實**：`sub_4257E @ IDA linear 0x4257E` 是既有符號索引所稱
  `Strategic_Bombardment_`／`Resolve_Strat_Colony_Damage` 的同一原始定位。
- **已證實**：它建立兩側快速戰鬥記錄後，以 `cx=0` 開始，尾端 `inc ecx`、`cmp cx,3`、
  `jl 0x4261F`，固定執行三個外層回合。
- **已證實**：每個外層回合會遍歷戰鬥者 1..`word_1998C0-1`，依序呼叫
  `sub_41F80`、`sub_4221F`、`sub_420C0`；這三支分別消費光束／一般武器、飛彈與特殊武器欄。
- **已證實**：行星累積傷害位於快速戰鬥記錄 `+0x1F`。若值大於 `0x7530`（30,000），
  將外層計數設為 6 以提前離開；函式最後以 signed integer 除以 `0x28`（40）回傳。
- **已證實**：唯一直接 caller `sub_4267B @ 0x4267B` 在 `0x42713` 呼叫後，於 `0x42726`
  將 AX 直接寫入轟炸結果 record `+3`。`sub_E7678` 讀該欄並與 record `+6` 相加；
  `sub_E87D2` 的多條玩家／AI 路徑亦直接消費 `+3/+6`。因此 `/40` 是 runtime 結果尺度，
  不能再套手冊 UI 估算的 `/100`。
- **已證實**：現行 Go 把 patch 1.31／1.50 的炸彈「5／10 次攻擊當量」誤當成所有武器的
  外層齊射數，導致光束、飛彈與特殊武器也跑 5／10 次，與原始控制流衝突。

## 證據身分

| 項目 | 值 |
|---|---|
| 輸入檔 | ZIP member `mastori2/Orion2.exe` |
| SHA-256 | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| IDA 資料庫 SHA-256 | `6562313be340a6bb80d43f25478446ba0bae24285ac86f0419b4f7de02a14fd0` |
| 工具 | IDA Pro 9.4.0.260610，SDK 940 |
| 位址空間 | IDA linear；不可直接當檔案偏移 |
| 可重生匯出 | `tools/ida/audit_strategic_bombardment.py` |
| 本輪匯出 | `/tmp/moo2-strategic-bombardment-20260824.json`（私有研究輸出，不進 Git） |

資料庫未帶原始 executable，因此匯出 JSON 的即時計算欄為
`database-input-not-present`；上表 executable 雜湊沿用同一資料庫建立時已記錄的輸入 inventory，
不可把資料庫雜湊誤寫成 executable 雜湊。

## 原始指令錨點

| 位址 | 原始指令／資料流 | 證據等級 |
|---|---|---|
| `0x4261B` | `xor ecx,ecx` | 已證實：外層計數歸零 |
| `0x4262D` | `call sub_41F80` | 已證實：第一攻擊鏈 |
| `0x42636` | `call sub_4221F` | 已證實：飛彈攻擊鏈 |
| `0x4263D` | `call sub_420C0` | 已證實：特殊攻擊鏈 |
| `0x42647` | `cmp word ptr [record+1Fh],7530h` | 已證實：30,000 提前停止門檻 |
| `0x42660..0x42665` | `inc ecx; cmp cx,3; jl 0x4261F` | 已證實：固定三外圈 |
| `0x4266C..0x42676` | 讀 `+0x1F`、`idiv 0x28` | 已證實：最終除以 40 |
| `0x42713..0x42726` | 呼叫 `sub_4257E`、`mov [record+3],ax` | 已證實：`/40` 回傳值直接成為結果欄 |

`sub_416CF @ 0x416CF` 的 33-byte combatant、type 固定 attack-count 表、最佳武器選取、行星與
七種防禦接線、三層軌道基地 record 及持久回寫，已於 2026-08-28 全部閉合。原版戰略模式不按
設計槽逐門計數；詳見 [`strategic-bombardment-full-audit-20260828.md`](strategic-bombardment-full-audit-20260828.md)。

## 版本參數停止線

patch 1.50 文件的「Bomb weapons now get bomb hits equivalent to 10 instead of 5 attacks」只支持
炸彈類武器的 5／10 攻擊當量差異，不支持把整個 `Strategic_Bombardment_` 外圈改成 5／10。
remake 因此把版本欄位限縮為炸彈專用；光束、飛彈與特殊攻擊仍依原始三外圈。1.31 的完整
attack-count 表已追回；1.50 bomb 當量如何修改三外圈內部仍是版本相依未知，不阻塞 1.31 RE。
