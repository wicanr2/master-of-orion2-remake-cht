# v0.1.0 — 繁體中文 remake alpha

日期：2026-08-12

這是準備上傳的 alpha 發行候選。核心 remake 已串接成可玩的多帝國 4X 迴圈，並包含
繁體中文介面、殖民／研究／造艦／戰鬥／外交、AI 選星與議會票、多人 TCP／熱座最低可玩鏈，以及
Linux、Windows、macOS 的公開包。

## 下載與驗證

預定 Release assets：

- `MasterOfOrion2-cht-x86_64.AppImage`（Linux amd64，9,202,168 bytes）
- `MasterOfOrion2-cht-windows-amd64.zip`（Windows amd64，9,045,616 bytes）
- `MasterOfOrion2-cht-macos-universal.tar.gz`（macOS `arm64`／`x86_64`，17,206,727 bytes）
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

本輪已抽樣覆核新局／客製種族、`RACES` 間諜、領袖、魚雷改造與 `.GAM` 匯入的已接流程；仍待補證的
重點是外部音訊逐曲人耳確認、真 Windows／macOS 主機驗收，以及僅在追求逐值等價時才需要的原版
DOSBox 動態 oracle。NAT 穿透仍需外部 relay 或 UPnP。

## 素材與影片權利

公開 Release 不包含任何原版 `.LBX`、`STREAM`／`STREAMHD`、音效或私有字型。使用者授權資料
產出的 61 秒實機遊玩推廣片只保留於本機預覽，未放入公開 Release；其畫面由封裝後 AppImage 在
Docker + Xvfb 走新局、種族、星圖、殖民地人口調配、`RACES` 間諜、外交與戰術流程即時錄得，但配樂
仍是原版 `STREAM.LBX`，若要公開散布影片，請先替換為已取得發布權的配樂。
