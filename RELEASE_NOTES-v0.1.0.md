# v0.1.0 — 繁體中文 remake alpha

日期：2026-08-11

這是第一個可公開下載的 alpha 發行候選。核心 remake 已串接成可玩的多帝國 4X 迴圈，並包含
繁體中文介面、殖民／研究／造艦／戰鬥／外交、AI 選星與議會票、多人 TCP／熱座最低可玩鏈，以及
Linux、Windows、macOS 的公開包。

## 下載與驗證

Release assets：

- `MasterOfOrion2-cht-x86_64.AppImage`（Linux amd64，9,378,296 bytes）
- `MasterOfOrion2-cht-windows-amd64.zip`（Windows amd64，18,133,442 bytes）
- `MasterOfOrion2-cht-macos-universal.tar.gz`（macOS `arm64`／`x86_64`，17,140,001 bytes）
- `PUBLIC-SHA256SUMS`

下載後可執行：

```text
sha256sum -c PUBLIC-SHA256SUMS
```

公開包不含原版遊戲資料。請以 `-data <合法的 MOO2 .LBX 資料目錄>` 啟動，並視平台／環境以
`-font <可用 CJK 字型>` 指定繁中字型。macOS 包未做 Apple 正式公證，首次執行可能需要使用者
手動允許未公證程式。

## 測試摘要

本輪採抽樣測試，remake 可玩度暫定 **70／100**。最新 Docker + Xvfb 畫廊抽樣為 35/35 張、
退出碼 0；`LBX asset id 31` 警告、panic、fatal、error 均為 0。完整限制與尚未完成的玩家路徑
見 [`docs/GAME-TEST-REPORT-2026-08-11.md`](docs/GAME-TEST-REPORT-2026-08-11.md)。

仍待補證的項目包括 20 回合經濟／士氣體感、客製種族完整開局、`RACES` 間諜面板、`LEADERS`
管理、魚雷 NR 距離數值、`.GAM` fixture 匯入、原版 DOSBox 動態 oracle，以及真 Windows／macOS
主機驗收。NAT 穿透仍需外部 relay 或 UPnP。

## 素材與影片權利

公開 Release 不包含任何原版 `.LBX`、`STREAM`／`STREAMHD`、音效或私有字型。使用者授權資料
產出的 72 秒、有原版 `STREAM.LBX` 音樂的推廣片只保留於本機預覽，未放入公開 Release；若要公開
散布影片，請先替換為已取得發布權的配樂。
