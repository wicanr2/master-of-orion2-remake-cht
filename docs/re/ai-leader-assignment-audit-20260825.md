# AI 領袖招募與任命靜態稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4／IDAPython；位址為 IDA linear、DOS/4GW LE object #1。
- 可重生匯出：`tools/ida/audit_ai_leader_assignment.py`；保留原始名稱、位址、bytes 與 caller/callee。
- 範圍：靜態資料流；不宣稱原版 PRNG 位元序列一致。

## 已證實

1. `sub_D7439 @ 0xD7439..0xD7662` 才是完整 AI 領袖回合處理器；`Next_Turn_Calc_` 在 `0x1370B` 呼叫它。舊腳本錨定的 `0xD7662` 是下一個函式邊界，不是 `Do_AI_Leaders_`。
2. 處理器先從 `word_199998` 的高槽往 1 號槽迭代，跳過 `player +0x28 == 100` 的已滅亡帝國；0 號槽是人類玩家，不走 AI 接受決策。
3. AI 的待聘候選位於玩家記錄 `+0x5A1`。處理完無論接受或拒絕都在 `0xD75B8..0xD75C4` 清回 `-1`。`Do_AI_Leaders_` 位於 `Random_Officer_Check_` 的全玩家迴圈（`0x137D1..0x137E7`）之前，因此本回合生成的 offer 到下一回合才由 AI 決策。
4. `0xD7490..0xD7577` 是技能接受 gate。可直接由 2-bit 遮罩定位的偏好為：Assassin、Famous、Megawealth、Operations、Spymaster、Telepath，以及艦長的 Helmsman、Weaponry。Researcher 另受玩家 `+0x59D` gate；Trader 只在 Repulsive `+0x8B2` 非零時通過。沒有任一偏好技能就不進可負擔判斷。
5. `0xD7581..0xD75B3` 取得雇用費後採嚴格條件 `hireCost + 50 < treasury`；通過才扣款並以接受旗標呼叫 `sub_9718F`，否則以拒絕旗標呼叫。`sub_9718F @ 0x9718F` 把接受者設為 owner、ETA 0、status 0；拒絕者設為 status 4，之後走 30 回合 limbo 清理鏈。
6. `sub_D73D4 @ 0xD73D4` 每回合反向掃描 67 位領袖：status 0 的 type 0／非 0 分別交給 `sub_D6FDA`／`sub_D7171`；status 1 只有 type 0 交給 `sub_D7078`。`sub_D7078` 是同位置艦艇間的艦長重派，沒有 status 1 的殖民地重派路徑。
7. `sub_D6FDA @ 0xD6FDA` 只考慮同 owner、狀態／任命欄位合格且尚無領袖的船，並以 raw ship 欄位 `+0x10/+0x12/+0x16/+0x15` 全部不減、`+0x72` signed word 嚴格增加的支配比較，最後呼叫 `sub_98C23` 寫入任命。`sub_D7078` 另要求候選與目前指派船同位置，排序相同。
8. `word_199998` 是玩家槽數：`Next_Turn_Calc_ @ 0x137D1..0x137E7` 用它作全玩家 `Random_Officer_Check_` 迴圈上限。這同時勘誤先前把它解釋成目前玩家索引的說法。

## 強推論與停止線

- `+0x59D` 在 Researcher gate 的玩家可見語意尚未由本切片完整命名；remake 以「目前有有效研究主題」作可測代理，標為強推論。
- 艦艇的五欄支配排序已證實，但 remake 的 `Ship` 沒有五個 raw 欄位的一對一表示；以可用艦艇的結構／武器規模決定優先順序，不冒稱逐值一致。
- `sub_D7171` 的完整管理領袖評分、AI cache producer、tie-break 與 ETA callback 已於 2026-08-28 閉合；本頁不再保留「殖民地 raw 評分未知」的過期結論。完整證據見 [`leader-turn-chain-audit-20260828.md`](leader-turn-chain-audit-20260828.md)。remake 目前仍採人口與產出近似，須等 RE-first gate 關閉後另立 READY spec。
