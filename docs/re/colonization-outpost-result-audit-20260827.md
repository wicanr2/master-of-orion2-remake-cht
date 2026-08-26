# 拓殖／前哨站結果文案稽核（2026-08-27）

## 證據與結論

- **已證實（手冊）**：`GAME_MANUAL.pdf` p.55、p.61–62 規定殖民船只能在清除敵艦與太空怪獸後，於未殖民的實體行星建立人口 1 的殖民地。
- **已證實（手冊）**：p.56、p.85、p.119、p.133 規定前哨船可在氣態巨星與小行星帶建立無產出的掃描／補給據點，之後同一天體建立殖民地會留下 Marine Barracks。
- **已證實（remake source）**：`ColonizePlanet`、`BuildOutpost` 與 `BuildOutpostOnPlanet` 已實作上述主要 gate，但失敗結果以中文 `string` 穿過 `internal/shell` 到 UI。
- **強推論（介面轉接）**：現有中文／英文完整句子是 remake 提示，不是從原版字串表逐句追回的 oracle；應保留規則 gate，改以 typed 結果碼接外部 JSON，不把 remake 句子冒稱原版精確文字。

本輪不改變拓殖或前哨玩法，只移除規則層玩家文案。怪獸種類與不可直接殖民的天體類別以 typed 參數傳遞，避免把動態名稱預先拼成中文。
