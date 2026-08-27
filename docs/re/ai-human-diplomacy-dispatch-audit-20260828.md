# AI 對真人外交 dispatcher 與正式戰爭 policy 稽核

日期：2026-08-28

## 證據契約

- `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- IDA Pro 9.4；DOS/4GW LE image、IDA linear EA；正式資料庫未修改。
- 證據：
  [`evidence/ai-human-diplomacy-dispatch-ida-20260828.json`](evidence/ai-human-diplomacy-dispatch-ida-20260828.json)。

## 已證實

1. `sub_F5A9F @ 0xF5A9F..0xF67C1` 是 67-case 訊息接收 dispatcher，不是 AI 決策門檻函式。
   case 28 收到 player record 更新後，依 `player+0xE60` bit 把來源 bit 寫入
   `byte_1AB054`；case 29／31／66 在同步或處理外交畫面後清除該 bit。
2. `sub_51078 @ 0x51078..0x5138E` 是正式宣戰 writer。雙方皆非 human marker
   `player+0x28==100` 時 policy 雙向寫 4；任一方為 human 時，難度 0–2 寫 5、難度 3–4
   寫 6。它另清協議、把關係寫 `-75..-99`，並重設戰爭計時與冷卻。
3. 因此 `Treaty.FormalPolicy` 的 raw 4／5／6 是攻擊目標估值的權威輸入。先前
   `aiForeignPolicyFor` 忽略正式 4／5／6，再由中文 `StanceName` 與 normalized relation
   猜回 4／5，會把原版 policy 5 與 6 的非單調權重算錯。

## Remake 對映

- 已接：正式 4→`DiploLimitedWar`、5→`DiploWar`、6→`DiploTotalWar`，直接供
  `AIEnemyColonyValue` 消費；顯示態勢不得覆蓋正式 raw policy。
- 未接：AI↔真人宣戰候選所需的方向 `+0x68F/+0x69F/+0x717/+0x72F` typed 狀態與
  `sub_25DF1 → sub_51078` producer。舊 `DecideStance` 仍只作這個缺口的明示相容 fallback，
  不宣稱原版忠實。
- Win95／多人訊息傳輸內部不重製；remake 只保存 action、policy 與玩家可見請求狀態。
