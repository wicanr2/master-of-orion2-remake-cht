# 拓殖／前哨站外部結果文案規格

1. `internal/shell` 只回傳穩定拒絕碼及必要 typed 參數，不回傳完整玩家句子。
2. 拓殖結果攜帶 `ColonizationRefusalCode`、`gamedata.PlanetType` 與 `gamedata.MonsterKind`；前哨結果採對應的 `OutpostRefusalCode` 與怪獸種類。
3. `cmd/moo2` 是唯一顯示轉接層，使用 `assets/i18n/ui.json` 的 `galaxy.colonization.*`、`galaxy.outpost.*` 模板。
4. 未知代碼必須使用外部 fallback；不可在 Go 內補一段玩家可見文字。
5. 測試至少覆蓋雙語鍵存在、格式參數數量、一般拒絕與動態怪獸／天體類別。
