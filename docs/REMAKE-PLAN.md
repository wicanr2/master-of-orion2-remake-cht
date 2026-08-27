# 銀河霸主 2 重製計畫

## 法律與完整性邊界

- 原始遊戲檔、正版資料與原版存檔由使用者提供，只讀掛載，不進 Git 儲存庫。
- 重製使用乾淨的 Go／Ebiten 程式碼；`internal/save` 的原版 `.GAM` 解析器與 remake JSON 存檔分開。
- 可寫測試、截圖與轉檔產物只寫工作樹指定輸出或容器 `/tmp`；一次性工作使用 `docker run --rm`。
- 繁體中文翻譯與開源／使用者授權字型分開處理；授權不明字型不嵌入執行檔。
- Windows API／Win95 平台內部呼叫只視為外部契約；逆向到玩家可見輸入、輸出、錯誤與必要
  時序即停止，Go／Ebitengine 以現代平台近似實作，不追作業系統內部等價。
- 進度對外採 README 三項百分比儀表板，不合成單一還原度；即時工作仍只看 `WORKLIST.md`。

## 資產與證據盤點

| 類別 | 位置／格式 | 證據 | 狀態 |
|---|---|---|---|
| 執行檔 | 私有 `ORION95.EXE`／1.31、patch 1.5 | 反組譯位址與資料表索引 | 部分規則已接，候選估值仍有未知 |
| 存檔 | 私有 `SAVE*.GAM`、remake JSON | `internal/save`、`internal/shell/persist.go` | 兩條路徑分開且有測試 |
| 圖像 | LBX（背景、精靈、字串／資產） | `internal/lbx`、畫廊與資產測試 | 已解碼並接多數畫面 |
| 音訊 | `STREAM*.LBX`／`SOUND.LBX` PCM | `internal/audio`、音訊文件 | 已接場景與音效；Docker 技術抽樣通過，桌面聽感仍需驗 |
| 手冊 | `GAME_MANUAL.pdf`、`MANUAL_150.html` | `docs/knowledge-base/manual-cht` | 逐條引用，未證實項標級別 |
| 參考碼 | `openorion2/src` | 僅作次級渲染／資料交叉驗證 | 不視為遊戲引擎真值 |

## 架構

```text
平台／UI → shell 遊戲流程／規則 → gamedata 真值與公式
        → assets／lbx／save／uifont／audio 等可重用基礎層
```

## 執行與驗收入口

本文件只定義長期架構與法律邊界，不維護會漂移的系統狀態、下一步清單或完成比例：

- 目前實作工作：`WORKLIST.md` 頂端唯一活表。
- 逐系統原版對齊：`docs/re/parity-matrix.tsv`。
- 可證實成果、近似與未知：`docs/HONEST-STATUS.md`。
- 驗收證據：`docs/VERIFICATION-MATRIX.md`。
- 深層位址、資料流與公式：`docs/re/01-gap-report.md` 及其主題稽核文件。

每個玩法切片固定依「原版資料／函式 → typed data → 規則 → 玩家與 AI 消費端 → UI →
存檔 → 抽樣測試」閉合。Windows API／Win95 內部行為與 DOS 音訊硬體時序只驗玩家可見契約，
不列入玩法逆向完成閘門。
