# 主選單動態文案與版面稽核（2026-08-27）

## 證據分級

- **已證實**：原版 `MAINMENU.LBX#21` 與既有 overlay catalog 提供 CONTINUE、LOAD GAME、
  NEW GAME、MULTI PLAYER、HALL OF FAME、QUIT GAME 六顆按鈕；既有原版／remake 對照見
  `docs/tech/oracle-comparison-20260712.md`，原版入口與停用狀態索引見 `docs/re/01-gap-report.md`。
- **已證實（remake source）**：`menu()` 左下語言與規則版本列分別使用 `b.tr(...)` 與
  `fmt.Sprintf(b.tr(...))`；固定文案仍內嵌於 Go，繪製只限制 `maxW`，沒有高度契約。
- **已證實（remake requirement）**：語言切換與 1.3／1.5 規則版本切換由 `CLAUDE.md` 明定，
  點擊後會重建畫面；版本切換另由 `selectGameVersion` 保證規則與資產根同步。
- **未知／不適用**：沒有證據顯示原版主選單存在這兩條左下控制；它們是 remake 擴充，不能以
  原版六顆按鈕或背景美術推導精確座標、字句或字級。

## 本輪結論

保留原版六顆按鈕與既有 action key，不更動玩法。只將兩條 remake 擴充文案及主選單相關轉場
名稱移到 `ui.json`，並讓文字安全框由各自 220×22 點擊列推導。驗證只能證明 remake 文案分層
與版面自洽，不升格為原版 parity。

## 實機畫廊驗證

- **已證實（remake runtime）**：繁中與英文各 35 張正版資料畫廊均完成；首次英文畫廊顯示舊長句
  雖未越界卻被省略，故外部文案縮為語意等價的 `(switch)`／`（切換）`。第二次英文
  `01_menu.png` 已完整顯示兩列，無省略、越界或上下侵入。
