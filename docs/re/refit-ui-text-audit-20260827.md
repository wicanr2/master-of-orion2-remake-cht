# 殖民地 REFIT 畫面與文案稽核（2026-08-27）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4、IDAPython
- 位址基準：IDA linear，DOS/4GW LE object #1
- 唯讀匯出：`tools/ida/audit_refit_ui.py`
- 本輪輸出：`/tmp/refit-ui-20260826.json`（跨午夜工作的本輪一次性大型證據，不提交）

官方手冊第 70–71、131–132 頁仍是玩家規則的文字來源；本輪 IDA 用來核對函式邊界、
caller／callee 與舊符號衝突，不以函式名稱單獨證明公式。

## 已證實

IDA 資料庫中的原始函式邊界如下：

| 原始定位 | 邊界 | 本輪可證用途 |
|---|---|---|
| `sub_B2190` | `0xB2190..0xB230D` | 外部符號所稱加入改裝艦至佇列的候選函式；由 `sub_B2542` 呼叫 |
| `sub_C0ED4` | `0xC0ED4..0xC0F3E` | 第一張 REFIT 動畫 loader 候選；由 `sub_C187B` 呼叫 |
| `sub_C14E8` | `0xC14E8..0xC1576` | 第一張 REFIT popup draw 候選；由 `sub_C187B` 兩處呼叫 |
| `sub_C187B` | `0xC187B..0xC19BA` | 第一張殖民地 REFIT popup 外層候選 |
| `sub_C19BA` | `0xC19BA..0xC19D1` | 第二張 REFIT 動畫 loader 候選；由 `sub_C20AF` 呼叫 |
| `sub_C1D6B` | `0xC1D6B..0xC1E74` | 第二張 REFIT popup draw 候選；由 `sub_C20AF` 兩處呼叫 |
| `sub_C20AF` | `0xC20AF..0xC221F` | 第二張殖民地 REFIT popup 外層候選 |
| `sub_CDB2A` | `0xCDB2A..0xCDB3C` | REFIT help list 候選；由 `sub_C187B` 呼叫 |
| `sub_D108B` | `0xD108B..0xD10D2` | `names.txt` 誤標為 `Refit_Cost_` 的相鄰短函式 |
| `sub_D10EE` | `0xD10EE..0xD2754` | `func_names.txt` 標為 `Refit_Cost_` 的大型函式；會呼叫相鄰 `sub_D108B` |
| `sub_D33D1` | `0xD33D1..0xD356F` | AI refit desirability 候選 |

上述「用途」只承接外部符號作導覽；已證實的是函式邊界、原始位址及 caller／callee。尤其
`sub_D108B` 與 `sub_D10EE` 是兩個不同函式，文件不得再把 `names.txt` 與 `func_names.txt`
當成同一定位。

官方手冊已證實玩家可見契約：

- 改裝必須選同一艦體設計；
- Cruiser 以上需要 Star Base 或更高級軌道基地；
- 改裝中的艦艇不可使用，從佇列移除會毀掉來源艦；
- 成本為 `max(2 × (新設計成本 − 舊設計成本), floor(標準艦體成本 / 4))`。

## 強推論與近似

- 兩組 loader／draw／popup 的形狀支持原版至少有兩階段 REFIT 選擇介面，但本輪未追回每個
  widget 的精確字串 ID、座標與選擇資料流，因此不把 remake 的單頁清單稱為像素對齊。
- remake 尚無完整持久化設計庫，故選定現役艦後凍結「同艦體自動最佳模板」。這是已揭露的
  替代 UX；不是從 `sub_C187B`／`sub_C20AF` 證出的原版選擇規則。

## Remake 映射

- 原版規則與 remake 近似邊界維持在 `docs/tech/colony-production-controls.md`。
- `internal/shell` 只回傳 typed REFIT 錯誤碼與必要參數，不保存玩家語言句子。
- `cmd/moo2/refit.go` 以 `refit.*` 鍵從 `assets/i18n/ui.json` 取得標題、列表模板、預覽、
  按鈕、警告與錯誤訊息。
- 列表、標題、說明、預覽三行、按鈕與底部訊息都有明確雙軸安全框；超長值採單行省略或
  固定行數收束，不直接把文字畫出框外。

## 未知與停止線

- 原版兩階段 popup 的逐 widget 座標、字串表索引、animation timing 與設計庫篩選次序仍未知。
- `sub_D10EE` 很大；本輪不因外部名稱就宣稱其中所有控制流都是成本公式。玩家可見成本契約已有
  手冊與 remake 共用消費端，若未來公式實測矛盾，再以窄切片追查其輸入與回傳。
- 這些未知不阻塞可玩的 REFIT 垂直鏈與文案／版面收尾，但必須留在 parity 矩陣中。
