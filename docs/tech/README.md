# 技術文件知識庫

逆向 + 移植過程中確認的格式、數值資料與工程心得。每輪把新看到/翻譯的數值資料整理進來。

## 資料格式(逆向紀錄)

- [lbx-format.md](lbx-format.md) — `.lbx` 資產封存檔:容器(magic `0xfead`)、影像 header/flags、frame offset 表、內嵌調色盤(6-bit→8-bit)、scan-line RLE 解碼。
- [savegame-format.md](savegame-format.md) — `save?.gam` 存檔:版本 `0xe0`、關鍵 offset(colonyCount `0x25b`、galaxy `0x31be4`)、各實體結構大小與欄位佈局、真實檔驗證數據。
- [enums.md](enums.md) — 28 個資料枚舉對照(技術 212/研究主題 83/建築 49/種族特性 32/特殊裝備…),英文名即 gameplay 邏輯 key,中文欄待填(中文化術語表基礎)。**自動生成**(`scripts/gen-enums.py` 讀 openorion2 gamestate.h),對應 Go:`internal/gamedata/enums.go`。

## 待補(後續輪次)

- 枚舉 enums.md 的中文譯名欄(中文化階段填)。
- 唯讀衍生公式(艦艇戰力/研究成本/行星產出/雇用費)。
- ebiten porting 心得、音樂/音效格式、patch 1.3/1.5 差異、選單擴展。

> kick-off 階段的策略/可行性文件在 `../kickoff/`。
