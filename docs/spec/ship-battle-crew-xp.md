# 戰後艦員經驗規格

## 範圍

本規格只處理太空戰鬥結束後、持久玩家艦艇的 `CrewXP` 寫回。每回合自然成長、太空
學院、Instructor 與等級門檻另見 `ship-crew-xp-turn.md`。

## 規則

1. 只有玩家贏得戰鬥且至少有一艘持久玩家艦存活時發放。
2. 對每艘被摧毀、未俘獲的敵艦，取一到六的艦體級並加總。
3. 每艘 recipient 得到 `max(1, floor(sum/2))`。
4. 摧毀總和為 0 仍為每艘 recipient 加 1。
5. 戰鬥 XP 直接加到既有 `CrewXP`，不套每回合 consumer 的 500 上限。
6. 快速結算與格子戰術必須呼叫同一個規則函式；畫面層只負責回報被摧毀艦體級總和。
7. 俘獲艦不算入被摧毀總和。

## 持久化與重播

- `CmdCombatOutcome.Args[0:3]` 保留既有 `playerStart／enemyStart／won`。
- 新 command 在 `Args[3]` 附加 `destroyedEnemyHullClassSum`。
- 重播新 command 必須使用記錄值，不能依目前畫面重算。
- 舊 command 缺少 `Args[3]` 時，若勝利則以當回合 `genEnemyFleet` 的完整艦體級總和作
  向後相容近似；敗北不發 XP。

## 驗收

- 純規則邊界：sum 0→1、1→1、2→1、3→1、4→2、12→6。
- 500 XP 的艦艇打贏後可成為 501，證明未誤套每回合 cap。
- 快速結算只讓戰後倖存艦取得 XP。
- 戰術壓縮同時遇到擊沉與俘獲時，只累加擊沉艦。
- command round-trip／重播保留 `Args[3]` 並產生相同 XP。

## 證據

原版 raw 位址、bytes、predicate 與信賴等級見
`docs/re/ship-battle-crew-xp-audit-20260824.md`。

