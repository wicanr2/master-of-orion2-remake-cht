# MOO2 音訊格式:音樂與音效(第一性原理逆向)

> 目的:定案《銀河霸主 II》音樂/音效的真實儲存格式,推翻「MOO2 音樂是 XMIDI」的初始假設,並記錄每個結論的 byte 級證據與社群來源(provenance)。
> 日期:2026-07-10。方法:靜態溯源(`rulebook/62`)+ 社群交叉比對(`rulebook/64/65`,reference-before-reverse)。
> 驗收原則:因為我們**原封不動播放原版 PCM bytes**,音質與原版 bit-identical——「跟原版一樣」由建構方式保證,非近似。

## 一句話結論

**MOO2 沒有 MIDI 音樂。** 全部音樂與音效都是 **PCM WAV(22050 Hz、8-bit、未壓縮)**,直接存在 `.LBX` 封存檔內。Miles Sound System 的 `.MDI/.OPL/.MT` 檔是各音效卡的驅動/音色庫,遊戲**實際上不用它們放音樂**。

## 推翻的初始假設(記取:別憑格式簽名想當然)

資料目錄有 `ADLIB.MDI`(magic `AIL3MDI`)、`SAMPLE.OPL`、`ADRV688.DIG`、`DIG.INI`、`AILDRVR.LST` → 這些是 **Miles Design Audio Interface Library v3.02** 的簽名,直覺會推論「音樂 = XMIDI」。但:

1. **全 LBX 掃不到 XMIDI 的 `FORM/XDIR/XMID` 資料** — 只有 `BILLTEXT.LBX` 有 1 個巧合的 `FORM`。
2. `Orion2.exe` 內的 `FORM/CAT/XMID/XDIR` 字串是 **Miles 函式庫的錯誤訊息**(如 `XMID sound hardware not found`、`Could not allocate SEQUENCE structures`),**不是音樂資料**。
3. `MOX.SET` 不是曲目集,是 **multiplayer 數據機設定**(內含 AT 指令、電話號碼 `5551212`)。

社群權威證實(VOGONS《Improving music Quality in MOO2》討論串):MOO2 音樂是 **22kHz 8-bit stereo PCM**,「seemingly no midi music whatsoever in the game」;Miles 的 MIDI 音樂設定形同虛設。

## 檔案佈局(byte 級證據)

| LBX | entries | WAV(RIFF)數 | 內容 | 聲道 |
|---|---|---|---|---|
| `STREAM.LBX`(29 MB)| 11 | 8 | **背景音樂**(標準版)| stereo |
| `STREAMHD.LBX`(21 MB)| 21 | 21 | **背景音樂(HD 版)** | stereo |
| `SOUND.LBX`(4.25 MB)| 69 | 69 | **UI/介面/武器/爆炸音效** | 多為 mono |
| `CMBTSFX.LBX` / `SPHERSFX.LBX` | — | 0 | 戰鬥音效庫(**非 RIFF**,巢狀索引原始音庫,待逆向)| — |

所有 WAV 的 `fmt ` chunk 一致:`audioFormat=0x0001`(PCM)、`sampleRate=22050`、`bitsPerSample=8`。→ **播放端統一用 22050 Hz audio context,零重採樣**。

### LBX 內 WAV 的切幀細節

- LBX 標準容器(magic `0xfead`,見 `internal/lbx`)。每個 entry = `[offset[i], offset[i+1])`。
- `STREAM/STREAMHD` 的 **entry 0 是 22-byte 彩蛋字串** `cats rule dogs drool\r\n`(開發者塞的簽名槽,非音訊)。真正的 WAV 從 entry 1 起,乾淨地以 `RIFF` 開頭。
- 抽取策略:對每個 entry 掃 `RIFF....WAVE`,讀 RIFF chunk size 切出完整 WAV;無 `RIFF` 的 entry(彩蛋槽/空槽)略過。

### SOUND.LBX 音效名稱表(entry 0)

`SOUND.LBX` 的 entry 0 是 **20-byte 定長記錄表**(名稱順序 = 音效 entry 順序):前 8 byte 為 NUL 補齊的名稱,後 12 byte 為保留欄(對真實檔案逐 byte 核對,樣本中恆為 0,用途未知)。可在程式端解析成「名稱→index」查表。**注:早期版本本文件誤記為「8-byte 定長」,實作曾因此誤把保留欄當名稱、只解出第一筆——2026-07-10 對照真實 SOUND.LBX 逐 byte 核實後更正為 20-byte 跨距。** 關鍵 UI 音效名:

```
BUTTON1 BUTTON2 BUTTON4 BUTTON9 BUTTONA BUTTONB BUTTONE   ← 介面按鈕音
IONPULSE POP_UP_OR_MISSL POP_DOWN WARP_CORE_WARNING       ← 別名段的可讀描述
KABOOM RAWEXPL EXPL-1..EXPL-5 PHOTON TORPDO1 MISLFIRE     ← 武器/爆炸
XPORTIN XPORTOUT PODLAND SHIPMOVE WARP2D                   ← 移動/傳送
MONSTR2/5/6/7/9 ATAKSHIP VORTEX1                           ← 怪物/太空生物
```

名稱表尾端有 `NEW...` 分隔 + 一段別名/描述(如 `WAV_1_TOTAL`、`POP_UP_OR_MISSL`)。程式只解析前段定長名。

## 對播放架構的意涵

1. **不需任何合成器**(無 XMIDI→MIDI、無 SoundFont、無 OPL3 模擬)。repo 既有的 `internal/lbx` 解碼器已能取出 entry;只需把 entry 內的 WAV 交給 ebiten audio。
2. **忠實度天然滿足**:播放原版 PCM bytes = 與原版數位音訊驅動輸出 bit-identical。原版 DOS 音質本就是 22kHz/8-bit,不「美化」才是忠實。
3. **音樂用 `STREAMHD.LBX`**(21 條,Win95 版採用的較完整音樂);`STREAM.LBX` 為標準版備援。
4. **版權**:音訊只在玩家自備的 gamedata 內,**絕不入 repo**(同 LBX 美術政策)。remake 執行期從玩家的 LBX 即時抽取。

## 待辦(對原版 oracle 才能定案)

- [ ] STREAM/STREAMHD 各 entry → 曲目語意(主選單主題/戰鬥/勝利/失敗/外交…):需對原版實測聆聽 + 社群曲目表交叉比對。
- [ ] SOUND 各 entry → UI 事件對應(哪個 BUTTONx 用在哪類按鈕):同上。
- [ ] `CMBTSFX/SPHERSFX` 巢狀音庫格式逆向(戰鬥期音效)。

## 來源

- VOGONS,《Improving music Quality in Master Of Orion 2》— 確認 22kHz/8-bit PCM、無 MIDI 音樂。<https://www.vogons.org/viewtopic.php?t=38812>
- The Orion Nebula 論壇,《Moo 2 .lbx and .mex image/sound extraction》— stream/streamhd 為 WAV。<https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=210>
- VGMPF,Miles Sound System / XMI 背景。<https://www.vgmpf.com/Wiki/index.php/Miles_Sound_System>
- byte 級證據為本機 gamedata(`/home/anr2/moo2-private-build/gamedata/mastori2`)直接 `xxd`/`grep` 實測,非引用。
