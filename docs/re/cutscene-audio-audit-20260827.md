# Smacker 過場音軌稽核（2026-08-27）

## 證據

- **已證實（原版程式）：**IDA Pro 9.4 的 `sub_14DF7 @ 0x14DF7..0x15085`
  是 `Play_Cinematic_` 播放外層；其下游 `SMACKSOUND*` 與 `_SmackDoPCM` 證明
  原版過場會播放音軌。原始執行檔與位址基準沿用
  [玩家路徑稽核](cutscene-player-path-audit-20260827.md)，非破壞性匯出沿用
  `docs/re/evidence/cutscene-player-path-ida.json`。
- **已證實（格式）：**FFmpeg 官方 `libavformat/smacker.c` 定義標頭 audio rate／flag、
  每幀 audio chunk 與 packed flag；官方 `libavcodec/smacker.c` 的 Smacker audio decoder
  定義 LSB-first Huffman tree、predictor 與 8-bit wraparound delta。
  來源：[demuxer](https://github.com/FFmpeg/FFmpeg/blob/master/libavformat/smacker.c)、
  [decoder](https://github.com/FFmpeg/FFmpeg/blob/master/libavcodec/smacker.c)。
- **已證實（真實資產）：**`INTRO.LBX` SHA-256
  `0e9d21dba93937f441b7e5c04df917c8c6c16ae90ec0c11cec2b279b31e3a45c`，
  track 0 為 11025 Hz、packed、stereo、8-bit；1407 幀中 1394 幀帶 audio chunk，
  最大解壓 chunk 23748 bytes。抽查結局檔另涵蓋 22050 Hz packed mono 8-bit。

## 實作邊界

- `internal/smk` 解碼 MOO2 實際使用的 packed 8-bit mono／stereo PCM，保留原生
  11025／22050 Hz；`internal/audio` 再統一成 22050 Hz stereo16LE 交給 Mixer。
- Bink audio、DCT 與 16-bit Smacker 音軌未在 MOO2 抽查資產出現，採失敗即關閉，
  不以靜音或雜訊冒充支援。
- DAC、PIT、DMA、MSS 與 DOS wall-clock 依硬體時序停止線不逆向；remake 以 sample
  duration、影片 frame rate、跳過時停止音訊構成 hardware-spec approximation，
  不宣稱逐週期一致。

