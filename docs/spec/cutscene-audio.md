# Smacker 過場音訊規格

1. 讀取 header 的七軌格式與每幀 audio chunk；MOO2 使用的 packed 8-bit PCM 必須
   支援單／雙聲道與 11025／22050 Hz。
2. 音訊在正常互動路徑第一次更新時開始，取代背景音樂；截圖廊不得初始化音訊裝置。
3. 玩家跳過或影片播完時停止過場音訊，再恢復一般背景音樂；`tickBGM` 不得在過場中
   把已播完的單次音軌誤接成隨機背景曲。
4. 資產缺失、沒有 track 0 或 codec 未支援時，影片仍須可播放與轉場。
5. 玩家可見固定文案一律由 JSON／YAML 載入；本功能不得增加 Go 內嵌提示文字。

## 驗收

- 合成 mono／stereo packet 驗 predictor、delta 與 wraparound；真 `INTRO.LBX` 驗完整
  音軌解碼、非空、合法取樣率及全片影像不回歸。
- 純音訊測試驗 11025→22050 的 frame 數與聲道轉換。
- Docker＋Xvfb 編譯／測試 `cmd/moo2`；有音訊輸出的桌面抽聽仍列外部驗收。

