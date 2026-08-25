# 研究突破與完成回寫靜態審查（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4，DOS/4GW LE object #1 的 IDA 線性位址
- 非破壞性匯出：`tools/ida/audit_research_breakthrough.py`
- 匯出保留原始名稱、位址、bytes、運算元、caller 與 callee；本文語意不取代原始定位。

## 已證實的每回合突破公式

`Check_For_Research_Breakthrough_` 的原始函式名為 `sub_E44E0 @ 0xE44E0`，唯一直接 caller 是
`sub_E4F49 @ 0xE4FA1`。玩家結構欄位與流程如下：

1. `player+0x321` 為目前研究 field；值 0 時於 `0xE44EC..0xE44EE` 直接返回。
2. `sub_E1EF4 @ 0xE1EF4` 先由 `sub_E1E96 @ 0xE1E96` 取得 field 成本，再取
   `player+0x1EB` 累積進度與 `player+0xAC` 本回合研究產出之和。
3. `sub_E1EC6 @ 0xE1EC6` 只在 `total > cost > 0` 時計算
   `floor((total-cost)*100/cost)`，上限 100；整數除法成 0 時改為 1。
4. `sub_E44E0 @ 0xE4502` 把 `player+0xAC` 加到 `player+0x1EB`。只有本回合產出
   `>0` 且 chance `>0` 才於 `0xE450C..0xE4518` 呼叫 `sub_1247A0(100)`；該 helper
   已由既有獨立證據證實回傳 1..100，判定為 `roll <= chance`。
5. 成功時 `player+0x30=1`、`player+0x1EB=0`，再呼叫 `sub_E4410 @ 0xE4410`。
   因此原版沒有「剛好達成本立即完成」，也不把超額研究結轉下一 field。

## 完成應用與種族分支

- `sub_E4410` 對 field `<75` 走應用授予；field `>=75` 增加對應 Hyper level byte。
- `player+0x8B5 !=0` 會讓 `sub_E4410` 掃描該 field 的全部有效應用並逐一呼叫
  `sub_E4204 @ 0xE4204`；其寫端與 UI consumer 已於後續稽核閉合，確認為 Creative。
- 原始 switch 讓 raw field `{22,23,28,29,55,57}` 進入應用表掃描，其中 `{22,55,57}`
  另設 `bl=1` 並於掃描後呼叫 `sub_54FBF`。這些集合與分支已證實，但本切片尚未把它們
  命名成 ResearchAll 或其他玩家語意，避免只靠現行 Go 表反向證明原版分類。
- 一般 field 最終以 `player+0x322` 的應用呼叫 `sub_E4204`。其寫入端、Uncreative 選項形成與
  真人／AI 選擇時序已於
  [`research-application-selection-audit-20260825.md`](research-application-selection-audit-20260825.md)
  閉合並取代舊近似。

## Remake 對映與剩餘邊界

- `gamedata.ResearchBreakthroughChance`／`ResearchBreakthroughSucceeded` 保存已證實公式與端點。
- `engine.RunResearchPhaseWithRoller` 只在原版會擲骰的條件下注入 1..100 roller；突破失敗保留
  全部累積進度，成功清零。
- `shell.GameSession.researchBreakthroughRoll` 使用現有可存檔研究亂數流，正常玩家、熱座與 AI
  帝國回合皆接線。這保證 remake 可重播；原版使用全域 LCG，兩者不宣稱逐 seed 同步。
- `sub_E4204` 少數特殊 application callback 仍由既有待辦逐項追查；本切片與後續選擇時序
  稽核合併後，已閉合「預選 application → 產出累積 → 突破擲骰 → 授予 → 進度回寫」。
