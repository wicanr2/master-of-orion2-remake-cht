# 間諜嫁禍第三方完整靜態稽核

日期：2026-08-28

## 證據契約

- 輸入 `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- `.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4；DOS/4GW LE image、IDA linear EA；正式資料庫唯讀，分析
  `/tmp` 副本。
- 非破壞性腳本：[`audit_spy_framing.py`](../../tools/ida/audit_spy_framing.py)。
- raw instructions、bytes、函式邊界與 caller windows：
  [`evidence/spy-framing-ida-20260828.json`](evidence/spy-framing-ida-20260828.json)。
- 外部名稱 `Frame_Another_` 只供導覽；下列結論以 raw 暫存器、offset、分支與 caller 為準。

## 觸發時機

**已證實**：`Resolve_Player_Spies_ @ 0x1014A4` 對每個有效 attacker→defender pair 依序擲
attacker d100、defender d100，組成既有攻守差後依任務分支：

- Espionage：淨分數至少 80，或 attacker d100 恰為 100，才嘗試偷科技；淨分數至少 90，
  或 attacker d100 恰為 100，才把 frame 旗標傳給 `Steal_App_ @ 0x10119C`。
- Sabotage：淨分數至少 70，或 attacker d100 恰為 100，才嘗試破壞；淨分數至少 90，
  或 attacker d100 恰為 100，才先呼叫 `Frame_Another_ @ 0x100BC5`。
- Hide／間諜互殺不走此 helper。

嫁禍是成功任務的自動結果，不是外交畫面中的玩家選擇，也不是成功後額外擲一顆真假骰。
若偷竊沒有合法科技或破壞沒有合法建築，成功 helper 會在寫外交事件前返回，不會只留下
一筆空的 reason 2／4。

## `Frame_Another_ @ 0x100BC5` 呼叫契約

兩個 caller 共同證實：

- `EAX`：真實 attacker player index。
- `EDX`：defender player index。
- `ECX`：被偷 application ID；Sabotage 傳 `-1`。
- `EBX`：in/out attribution player index 指標；caller 先填入真實 attacker。

函式回傳是否找到至少一個候選；兩個 caller實際依輸出 player 是否仍等於 attacker 決定 reason，
因此無候選時即自然退回未嫁禍，不需要另一條 fallback writer。

## 候選資格與抽樣

**已證實**：函式以 playerCount−1 到 0 的降冪順序掃描全部帝國，候選必須：

1. 不是 attacker，也不是 defender。
2. 候選→attacker 的接觸 `+0x584` 非零。
3. 候選→defender 的接觸 `+0x584` 非零。
4. 若 `ECX` 是被偷 application ID，候選該 application 的 raw status `+0x117[application]`
   不能等於 3；Sabotage 傳 `-1`，不檢查科技狀態。

函式按 attacker→candidate 正式政策 `+0x627` 分成七桶。每個合格候選先增加該桶計數，再以
`Random(bucketCount)==1` 做標準 reservoir sampling，所以每桶內候選均勻；掃描降冪不會造成
低／高帝國索引偏差。

掃描完成後依 raw policy 桶固定優先序選第一個非空桶：

```text
6 → 5 → 4 → 0 → 1 → 3 → 2
```

本頁保留 raw policy 數字；既有外交證據已知 4／5／6 是不同戰爭 policy、1／2 是正式條約，
但不為這個排序自創「最可信」或「最可疑」等未證實設計名稱。

## reason、關係與訊息 payload

**已證實**：

- `Steal_App_` 成功授予科技後才執行嫁禍候選；attribution 等於 attacker 時 reason 1，
  不等時 reason 2。
- `Destroy_Random_Building_` 成功移除建築後，attribution 等於 attacker 時 reason 3，
  不等時 reason 4。
- 兩者都呼叫 `Change_Relations_`，基礎 delta 為
  `-(Random(15)+Random(5)) = -2..-20`。偷竊 payload 是 application ID；破壞 payload 是
  colony／building slot。
- 成功 helper 另建立兩筆結果訊息 record；其中保存真 attacker、attribution、defender 與
  科技或殖民地／建築 payload。這些 record 會讓攻擊方與受害方看到各自結果，不應只保存一條
  無 attribution 的中文摘要。

`Change_Relations_` 的 pending reason／magnitude、下一回合記憶與訊息 consumer 已由
[`change-relations-callers-audit-20260828.md`](change-relations-callers-audit-20260828.md) 及
[`ai-human-spy-incident-audit-20260828.md`](ai-human-spy-incident-audit-20260828.md) 閉合。

## 收斂判定

原版 1.31 的嫁禍觸發、候選資格、科技排除、七桶 reservoir、政策優先序、無候選 fallback、
reason 1／2／3／4、關係變化與結果 record 均已閉合。remake 尚未保存 attributed third party，
屬資料模型與實作差異；依 RE-first gate，本輪不寫新 spec、不修改 Go 玩法。
