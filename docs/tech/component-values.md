# 艦艇元件數值來源(provenance)

記錄 `-game` 艦艇設計元件(`internal/shell/session.go` 的 `WeaponOptions`/`ArmorOptions`/
`ShieldOptions`/`SpecialOptions`)數值的**來源狀態**,避免把「憑印象填的數字」誤當權威。
遵循 provenance 原則:每個數值註明「已確認 / 單調估計 / 版本相依」。

## 為什麼不能一次填齊「精確手冊數字」(2026-07-11 更新:結論已部分推翻,見下方修正)

第一性原理盤點 on-disk 可用來源(**原始盤點,已知有誤,保留供對照**):

| 來源 | 狀態(原始盤點) | 能否取得完整武器表 |
|---|---|---|
| `openorion2/src/*` | 只有渲染,不含戰鬥/武器數值 | 否(它是渲染殼非引擎) |
| `original_game/…CD Manual.pdf` | 9 頁**掃描圖**,文字抽取 0 字元 | 否(需 OCR;且未必含完整附錄) |
| `moo2_patch1.5/MANUAL_150.html` | patch 1.5 **變更日誌**,只列被改動的項目 | 否(非完整表,但含確認錨點) |
| 私有 gamedata LBX | 武器數值烘在執行檔/資料表,無現成 dumper | 否(需另行逆向) |

原結論:「線上無機讀的完整權威武器表」——**這個結論是誤判**。原始盤點漏看了
`moo2_patch1.5/GAME_MANUAL.pdf`(188 頁,patch 1.5 隨附的**完整**遊戲手冊,由 Google Docs
匯出、可正常 `pdftotext` 抽字,不是掃描圖),誤以為它跟 `original_game/…CD Manual.pdf` 一樣是
掃描版。**Ship Design 章節(p.119-132)其實有完整的武器 Damage/Size/Cost 表**(束射 13 項、
飛彈/魚雷 7 項、炸彈 6 項、戰機/特殊武器 13 項,逐項含傷害值),詳見 `ship-design-space.md`
§2(該文件聚焦 Size/佔格欄,Damage 欄同一張表可交叉取用)。

**因此「線上無機讀的完整權威武器表」不成立**——完整表就在 `GAME_MANUAL.pdf`,只是先前沒讀對檔案。
本檔下方「已確認錨點」小節建立於 `MANUAL_150.html`(1.50 patch 摘要)找到的零星錨點,**尚未**
用 `GAME_MANUAL.pdf` 的完整表回頭核對/更新 `internal/shell/session.go` 的 `WeaponOptions.Value`
(那會動到既有戰鬥平衡數字,超出本輪「艦艇設計空間格」任務範圍,留待下一輪「武器數值忠實化」
任務逐項核對接線,避免同一輪任務範圍發散)。

## 一個重要發現:係數本身依版本而異

`MANUAL_150.html` 明確記載電漿砲傷害 **1.31 為 6/30,1.50 改為 4/20**。這代表「精確係數」
不是單一定值,而是**版本相依**——正好對應本專案要支援 1.3/1.5 雙版本的需求(Phase 7)。
未來精確對齊時應建立**版本專屬 rule profile**,而非硬填一組數字。

## 已確認錨點(取自 MANUAL_150.html)

| 元件 | 確認值 | 原文 |
|---|---|---|
| 中子爆破槍 Neutron Blaster | 最大傷害 **12** | "Neutron Blaster (12 max damage)" |
| 高斯砲 Gauss Cannon | 戰鬥最大 **18**(戰略欄 10) | "Gauss max damage is higher than Neutron Blaster (18 vs 12)" |
| 電漿砲 Plasma Cannon | **20**(1.50);**30**(1.31) | "Plasma Cannon min/max damage from 6/30 to 4/20" |

其餘武器 Value 為依科技階遞增的**單調估計**(雷射 4→死光 25),保持排序合理與遊戲可玩,
但未經手冊逐條核對。裝甲/護盾/特殊同理(尚無 on-disk 確認錨點)。

## 待辦(精確對齊的可行路徑)

- [ ] OCR 掃描版手冊武器/裝甲/護盾附錄(若附錄存在於該 PDF;9 頁本可能不含,需找完整手冊)
- [ ] 或逆向私有 gamedata 武器資料表,寫 dumper 抽精確值(對齊 provenance 原則)
- [ ] 建立版本專屬 profile(1.3 vs 1.5),數值分版存放,而非單一硬編碼
