# `Diplomacy_Growth_` 關係成長靜態稽核（2026-08-27）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4／IDAPython，映像 `ida-pro-9.4-idapython:locked-v1`。
- 位址空間：IDA linear EA，DOS/4GW LE object #1。
- 資料庫：正式 `Orion2.exe.i64` 唯讀掛載，容器內 `/tmp/Orion2.exe.i64`
  一次性副本；未改名、未寫回註解。
- 匯出：[`evidence/diplomacy-growth-ida-20260827.json`](evidence/diplomacy-growth-ida-20260827.json)，
  `sub_4DD6B @ 0x4DD6B..0x4E3B5`，428 條指令，bytes SHA-256
  `f7dbf0a3a2320ec4d2b35f6fbf4337e288a938ddf83618744d73d7b8d5ccc7b3`。
- 直接 caller：`Next_Turn_Calc_` 內 `0x13706`（原名 `sub_136B3`）；因此本函式
  屬每回合正常玩家路徑，不是選單或除錯 helper。

## 已證實：配對與亂數邊界

1. `0x4DD74..0x4E015` 以 `outer=0..playerCount-1`、`inner=outer+1..playerCount-1`
   掃描每對帝國一次。只有兩者都是 raw human marker `player+0x28 == 100`
   才跳過整對條約成長。
2. 亂數 helper `sub_1247A0 @ 0x1247A0` 已由專案共用證據釘死為
   `Random(n) = 1..n`（`docs/re/01-gap-report.md` 開頭規則）。
3. 條件 `Random(100) <= 100-currentRelation` 直接來自
   `0x4DE10..0x4DE2A`、`0x4DE5B..0x4DE77`、`0x4DEAB..0x4DEC7`、
   `0x4DF2D..0x4DF48` 與 `0x4DF80..0x4DF9B`。`currentRelation` 是 raw signed byte
   `observer player + target slot + 0x617`，尺度 `-100..100`。

## 已證實：條約分支與順序

以 `outer` 的方向 record 檢查條約，呼叫 `Change_Relations_` 時固定
`EBX=inner`、`EDX=outer`；即由高槽 `inner` 當 actor，變更對 `outer` 的關係，
再由 `Change_Relations_` 鏡射關係分數。同一對帝國依下列順序消費亂數與目前分數：

| 原始欄位／條件 | 位址 | 守門 | 基礎 delta | reason |
|---|---:|---|---:|---:|
| `+0x627 == 1` 互不侵犯 | `0x4DE07..0x4DE43` | 百分比守門 | `Random(3)` | 0 |
| `+0x62F == 1` 貿易協議 | `0x4DE52..0x4DE93` | 百分比守門 | `Random(3)` | 12 |
| `+0x637 == 1` 研究協議 | `0x4DEA2..0x4DEE3` | 百分比守門 | `Random(3)` | 13 |
| `+0x627 == 2` 同盟 | `0x4DEF0..0x4DF11` | 無百分比守門 | `Random(5)` | 0 |
| `+0x63F == 1` 納貢 mode 1 | `0x4DF22..0x4DF64` | 百分比守門 | `Random(3)` | 14 |
| `+0x63F == 2` 納貢 mode 2 | `0x4DF75..0x4DFB7` | 百分比守門 | `Random(8)` | 14 |

每次呼叫都會立即讀取前一次寫回後的 relation；不得先計算全部 delta
再一次相加。`Change_Relations_` 的可達共用公式已由
[`random-event-diplomatic-incidents-audit-20260825.md`](random-event-diplomatic-incidents-audit-20260825.md)
的 2026-08-27 勘誤證據重新匯出；本切片的條約狀態只取 1／2，不會命中
戰爭早退。

## 已證實：目標初始化與逐回合漂移

- `sub_4D78E @ 0x4D78E..0x4DAB2` 由 `Next_Turn_Calc_` 的初始化路徑呼叫；
  bytes SHA-256 為
  `b96299e71aad5ecf5d8b3f31aab1feb4607ba0af277ebb5b72c1b8f07ebb4df3`。
- `player+0x25` 是原版種族索引。初始化器以
  `byte_180ED4 + 28*observerRace + 2*targetRace` 讀取 14×14 signed-word
  表，將相同 low byte 寫入 current `+0x617` 與 target `+0x61F`。
- `byte_180ED4` 原始 392 bytes 的 SHA-256 為
  `05e582491a6173319e8d57d0751d24dd5629f73f00dc3d7c2a5b979e39efa831`；
  raw hex 與逐列 signed low byte 已收在本頁的 JSON 證據。
- `0x4E11D..0x4E276` 每條 observer→target 邊先擲 `Random(105)`；只有結果
  `> abs(current)` 才擲 `Random(4)`，等於 1 時再以 `Random(2)-1` 得到 0／1
  step。鎖定 word `observer+2*target+0x737 == 0` 時，current 才向目標靠近。
  正式狀態 `>=4` 時，不論鎖定與否都把高於 `-90` 的 current 壓到 `-90`。
- **資料模型投影**：remake 只有 AI→玩家單邊 `AIOpponent.Relation`，因此以
  AI 種族為 observer、玩家種族為 target 保存 `+0x61F`。自訂種族不在原表，
  失敗即關閉為 0；反向與 AI↔AI 全矩陣仍是資料模型缺口。

## 已證實但本切片不實作

- `0x4DDB9..0x4DDF8`：`Random(2)==1`、relation `<-50` 且 raw `+0x72F>0`
  時 `+1`。`+0x72F` 的 writer／玩家語意未閉合，等級為「未知欄位的已證實消費」。
- `0x4E01B..0x4E117`：human observer 對 non-human target 的領袖／實力複合負面
  reason 5；`sub_DBCC8` 輸出欄位及領袖 record 尚未逐項釘死。
- `0x4E282..0x4E339` 鏡射分數並遞減 raw `+0x806/+0x80E` 計時器；
  `0x4E33B..0x4E3B0` 在 raw game mode 3 將所有帝國關係設為狀態 6／分數 -100。
  這些另立規格，不與條約增益混寫。

## Remake 投影與證據等級

- **已證實**：上表六個條約分支、掃描順序、亂數邊界、reason 及呼叫方向。
- **已證實**：`Change_Relations_` 中 current-dependent 縮放、raw 政體 0／4、
  target Charismatic、`-100..100` 夾值及非同盟最高 65。
- **remake 資料模型投影**：現有 UI／AI 消費 `Relation=-40..40`。為避免 raw
  `+1..3` 在每回合轉換時永久捨去，必須另存 `-100..100` raw 餘數，顯示尺度
  只作投影。
- **強推論**：remake 未有 AI 當局政府欄位，以 AI 種族的原版開局政體
  作 actor government fallback；政府升級後的差異繼續明列為資料模型缺口。
- 1.50 二進位未取得；本頁只證實 1.31 executable。
