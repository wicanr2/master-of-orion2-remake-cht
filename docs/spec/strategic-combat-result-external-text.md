# 戰略戰鬥動態結果外部文案規格

1. `internal/shell` 的轟炸、入侵、心靈控制及怪獸戰術前置 gate 只回傳穩定 typed code。
2. `cmd/moo2` 依 action 對應 `galaxy.bombard.refusal.*`、`galaxy.invasion.refusal.*`、`galaxy.mind_control.refusal.*` 與 `galaxy.monster_combat.refusal.*`。
3. 固定玩家句子只存在 `assets/i18n/ui.json`；未知 code 使用同檔 fallback。
4. 規則測試驗證 code，UI 測試驗證中英文鍵及 fallback 已解析；不得以翻譯句子作規則斷言。
5. 本規格不改成功戰報、傷害解算、資產動畫或戰術戰果回寫。
