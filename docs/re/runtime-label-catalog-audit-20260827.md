# Runtime 顯示標籤 catalog 稽核（2026-08-27）

## 範圍與既有證據

本輪稽核 `cmd/moo2/englishlabels.go`。它原本是英文安全 fallback 的集中層，但仍直接保存
玩家會看到的英文值，且舊中文 literal 棘輪無法偵測純英文硬編字串。

- 新遊戲五個選擇器、原版值圖與座標證據沿用
  `docs/re/newgame-setup-text-audit-20260827.md`：`Newgame_Screen_ @ 0xCD435`、
  `sub_CCE2E`、`sub_CCC3D` 與 `NEWGAME.LBX#1..22`。選項索引與規則資料不改。
- INFO 五子畫面與 `BILLTEXT.LBX` 字串證據沿用
  `docs/re/info-subscreen-text-audit-20260826.md`。歷史圖表四指標是 remake adapter 顯示名，
  不冒稱原版逐字。
- 戰機 runtime 公式錨點沿用
  `docs/re/tactical-fighter-text-layout-audit-20260827.md` 的 `sub_3AC20`、`sub_3AD57`、
  `sub_3D2DF`。本輪只移動 FighterKind 的顯示標籤，不改傷害、命中或動畫。
- 熱座真人計數證據沿用
  `docs/re/hotseat-empire-selection-audit-20260827.md` 的 `sub_121F0 @ 0x121F0..0x12227`；
  remake 的席位名稱格式仍是資料模型轉接，不是原版逐字 UI。

## 已證實的 source 缺陷

- `englishlabels.go` 直接保存四個歷史指標、十五個新遊戲值、`You`、四個戰機種類、
  `%d more`、`%s Ship %s`、兩種 `Player %s` 模板及 Unknown Empire／Ship／Player。
- 這些字串都由正常玩家畫面消費：新遊戲、INFO 歷史、行星列表、格子戰術與熱座交接。
- 舊 `TestEnglishModeGapDoesNotGrow` 只掃未包裝的漢字 literal；純英文硬編文案不會被計數，
  因此其 13 條結果不能證明外部化完成。

## 分級與停止線

- **已證實：**索引、typed enum、種族 ID、戰機 kind 與動態數值仍是規則資料，不可改成翻譯字串識別。
- **強推論：**既有中英文詞義足以維持目前玩家介面；外部化不改操作或規則。
- **remake adapter：**歷史圖例的「你」、舊存檔未知值、英文敵艦組字與熱座 fallback 格式。
- **未知：**原版上述 adapter 的逐字 fallback、大小寫與標點；本輪不以現有英文詞升格 parity。

## Remake 對映

- 固定值改由 `ui.json` 穩定鍵提供；Go 只保留 enum／索引到鍵的映射及動態參數。
- 未知值依目前語言使用 `common.unknown*`，不再在繁中無效索引時回傳英文 `Unknown`。
- 舊存檔的原始中文字串仍只用來辨識 race／type，不修改存檔；顯示 fallback 由 catalog 決定。
- 來源契約須禁止本輪移出的英文句子重新進入 `englishlabels.go`，並以所有索引、雙語格式與
  正常畫廊抽樣驗收。
