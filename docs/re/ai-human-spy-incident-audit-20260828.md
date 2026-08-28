# AI 對真人間諜事件記憶稽核

日期：2026-08-28

## 證據契約

- `Orion2.exe` SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `.i64` SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- IDA Pro 9.4；DOS/4GW LE image、IDA linear EA；正式資料庫唯讀，分析容器只操作 `/tmp` 副本。
- 原始函式名、位址、bytes、xref、callsite 視窗與 operand 掃描見
  [`evidence/ai-human-incident-producer-ida-20260828.json`](evidence/ai-human-incident-producer-ida-20260828.json)。
- Hex-Rays 輸出只供導覽；變數名與型別不是證據。

## 已證實

1. `Change_Relations_ @ 0x4E3B5` 在一般 policy、有效變化非零時，只保留絕對值最大的
   pending 事件：reason 寫方向 `+0x64F`，套政府、Charismatic 與難度後的有效變化寫
   `+0x65F`。reason 只有 1..9 且負變化才會成為後續真人外交事件。
2. `sub_10119C @ 0x10119C..0x10130A` 的科技竊取成功路徑，在授予科技後呼叫
   `Change_Relations_(-(Random(15)+Random(5)), attacker, defender, reason, 0, tech)`。
   未嫁禍時 reason=1；嫁禍給第三方時 reason=2。
3. `Steal_App @ 0x10130A..0x1014A4` 只有實際選到並移除建築後才呼叫同一關係函式；
   幅度相同，未嫁禍時 reason=3、嫁禍時 reason=4。沒有合法建築候選不產生事件。
4. `sub_252D5` 既有純規則會在下一回合把 `+0x64F/+0x65F` 依政府、正式條約、貿易、研究
   與永久違約記憶轉成 `+0x71F`；`sub_4F0DC` 保存 reason 1..9 到 `+0x6CF`，再由
   `sub_544A1` 消費。pending 幅度不能直接冒充 memory。

## Remake 對映與限制

- 玩家成功且實際取得科技／拆掉建築時，現在分別記 reason 1／3，套原版關係變化並保存
  pending reason／magnitude；下一回合重用 `OriginalNPCIncidentMemoryStep` 形成 memory。
- 新欄位隨 JSON snapshot 往返；AI 對真人 target composer 因而可從正常間諜玩家路徑取得
  `+0x71F/+0x6CF`，不再只能依新局預設值。
- 原版嫁禍不是玩家選擇：attacker d100=100 或淨分數至少 90 時自動嘗試，第三方依雙方接觸、
  偷竊科技狀態與七個正式政策桶抽選；無候選時退回真 attacker。完整鏈見
  [`spy-framing-audit-20260828.md`](spy-framing-audit-20260828.md)。remake 尚無 attributed player
  資料，因此 reason 2／4 仍是**原版 RE 已閉合、remake 未實作**。
- reason 5／7／8／9 的其他正常 producer、動態宣戰 reason 與完整訊息 consumer 已由
  [`change-relations-callers-audit-20260828.md`](change-relations-callers-audit-20260828.md)
  逐 caller 閉合；reason 6 沒有獨立 literal producer，仍只可能來自動態／網路輸入。
  嫁禍 reason 2／4 的 remake 資料模型缺口不因此消失。
