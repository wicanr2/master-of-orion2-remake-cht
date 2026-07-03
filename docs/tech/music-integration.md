# 音樂 / 音效整合可行性(go/ebiten remake)

> 討論 MOO2(Master of Orion II: Battle at Antares, 1996)原版的音樂與音效如何整合進本專案的 go/ebiten 重製。
> 來源:原版遊戲檔逆向(`original_game/…zip` 內 `mastori2/`)+ 執行檔字串 + LBX 容器解析 + 社群考據 + ebiten 官方文件。
> 逆向數值(LBX 音訊資產大小、WAV `fmt` 規格)皆本專案實測,標「實測」。判準:**完整性優先、配樂用原版真實素材、不自產逼近**(對齊 rulebook 83 / 93)。

---

## 0. 結論先講(關鍵發現)

任務初始假設是「原版音樂是 XMI/MIDI,要 XMI→MIDI→SoundFont 或 XMI→OPL 離線渲染」。**逐位元組驗證後推翻此假設**:

- **MOO2 的遊戲配樂本身就是預先錄好的數位 PCM 音訊(RIFF/WAVE)**,存在 `STREAM.LBX` / `STREAMHD.LBX`,規格 **22050 Hz / 8-bit / 2 聲道(立體聲)**(實測)。
- 因此**不需要**跑 XMIDI 合成、OPL FM 模擬或 SoundFont 渲染這條複雜路線 —— 直接把 LBX 裡的 WAV 抽出來,交給 ebiten 播即可。
- 執行檔確實連結了 Miles AIL 的 XMIDI 子系統、隨附一堆 `.MDI` MIDI 驅動與音色庫(`SAMPLE.*`),代表引擎**有能力**做 MIDI 合成;但「實際遊戲配樂走 PCM 串流」由 LBX 內實體資產 + 社群考據支持。MIDI 那條路對重製而言**不是必要路徑**(見 §1 不確定標註)。

> **更正前一版斷言(rulebook 62/63/83)**:本文前一版曾寫「音樂在 `SCORE.LBX`,是 XMIDI + `SAMPLE.OPL` FM 音色,需 OPL 離線渲染」。**這是憑檔名(SCORE=樂譜)臆測、未驗證的錯誤斷言。** 實測 `SCORE.LBX`(17 資產)內 **XMI 標記(FORM/XDIR/XMID)= 0、RIFF = 0**,其 asset 0 開頭 `80 02 e0 01` = **640×480**,即 LBX **影像**格式(計分/記分板畫面圖),**與音樂無關**。真正的音樂在 `STREAM/STREAMHD`(PCM),見 §2。舊斷言已作廢。

一句話:音樂與音效整合對本專案而言是**「抽 PCM → 餵 ebiten」**的直線工程,不是「合成器移植」的難題。

---

## 1. 原版音訊架構(逆向確認)

### 1.1 音效引擎:Miles AIL(Audio Interface Library)

`Orion2.exe` / `SETSOUND.EXE` 字串實證(實測):

```
;Miles Design Audio Interface Library V3.02 of 18-Jan-95
%F Copyright (C) 1995 Miles Design, Incorporated
$R:\NET\LIBS\AIL\DEV3\FLAT\ailxmidi.c
_XMI_construct_MDI_driver@   _XMI_send_channel_voice_message   Invalid XMIDI sequence
DOS/4G Copyright (C) Rational Systems  (WATCOM C/C++32 Run-Time)
```

→ 原版是 Watcom C + DOS/4G extender + **Miles Design AIL v3.02**(即後來的 Miles Sound System, MSS)。AIL 分兩個子系統:

- **MDI(MIDI/XMIDI)**:序列化 MIDI 音樂,依硬體用不同 `.MDI` driver 合成。
- **DIG(Digital)**:數位 PCM 音訊播放(串流音樂 + 取樣音效)。

### 1.2 隨附的驅動與設定檔(實測,`SETSOUND.EXE` 選擇)

| 類型 | 副檔名 | 檔案(節選) | 用途 |
|---|---|---|---|
| MIDI driver | `.MDI` | `ADLIB.MDI` `OPL3.MDI` `SBLASTER.MDI` `SBPRO1/2.MDI` `SBAWE32.MDI` `MT32MPU.MDI` `MPU401.MDI` `ULTRA.MDI` `TANDY.MDI` `PCSPKR.MDI` `PAS.MDI` `NULL.MDI` | 各音效卡的 XMIDI 合成 |
| Digital driver | `.DIG` | `SB16.DIG` `SBLASTER.DIG` `SBPRO.DIG` `PROAUDIO.DIG` `RAP10.DIG` `SNDSCAPE.DIG` `ULTRA.DIG` `ADRV688.DIG` | 各音效卡的 PCM 播放 |
| 音色庫 | `.AD/.OPL/.BNK/.MT/.CAT` | `SAMPLE.AD` `SAMPLE.OPL` `SAMPLE.BNK`(OPL/AdLib FM 音色)`SAMPLE.MT`(MT-32)`SAMPLE.CAT` | XMIDI 合成用的樂器音色定義 |
| 設定 | `.INI/.SET` | `MDI.INI` `DIG.INI` `MOX.SET` `MT32/MDI.INI` `SB16/MDI.INI` `SC55/MDI.INI` | 玩家在 `SETSOUND.EXE` 選的硬體 |

> **不確定標註(rulebook 62/93)**:上表 MDI + 音色庫證明引擎**具備** MIDI 合成能力,但**沒有**在任何 LBX 裡找到 XMI 音樂序列資產(`STREAM/STREAMHD` 是 PCM;`SCORE.LBX` 是影像)。可能情形:① 遊戲配樂全走 PCM 串流,MDI/音色庫是 AIL SDK 的預設隨附(Miles 一律附這些)或僅用於少數 jingle;② 某些短音樂走 XMIDI 但序列內嵌在別的 LBX/EXE。**本專案未逐一驗證每首曲子來源**;但主配樂資產(§2)已是 PCM,此不確定性**不影響重製決策**。若日後要完整保全「MIDI 版音樂」再回頭追 XMI 序列位置(靜態溯源 SOP),別再憑檔名假設。

---

## 2. 音樂資產:預渲染 PCM(核心,實測)

社群考據:「MOO2 全部音樂是低品質 PCM,22 kHz / 8-bit / 立體聲」,並指 `STREAM` / `STREAMHD` 資源實為 WAV(見來源)。本專案 LBX 解析證實。

兩個音樂容器都是標準 LBX(magic `0xfead`,格式見 `lbx-format.md`),每個資產是一段內嵌 **RIFF/WAVE**:

- WAV `fmt` chunk 實測:`audioFormat=1`(PCM)、`channels=2`、`sampleRate=22050`、`bitsPerSample=8`。
- 8-bit PCM 為**無號**(unsigned,靜音值 = `0x80`),這是 WAV 8-bit 慣例;16-bit 才是有號。時長換算 = `size ÷ (22050×2×1)`(8-bit 每樣本 1 byte)。

### 2.1 `STREAM.LBX`(~29 MB,11 資產,實測)

| asset | 大小(byte) | 概略時長 | 備註 |
|---|---|---|---|
| 0 | 22 | ~0 s | 空樁(stub) |
| 1 | 5,012,208 | ~113.7 s | 音樂軌 |
| 2 | 5,527,010 | ~125.3 s | 音樂軌 |
| 3 | 1,887,988 | ~42.8 s | 音樂軌 |
| 4 | 4,910,256 | ~111.3 s | 音樂軌 |
| 5 | 5,994,270 | ~135.9 s | 音樂軌 |
| 6 | 2,466,700 | ~55.9 s | 音樂軌 |
| 7 | 22 | ~0 s | 空樁 |
| 8 | 2,583,936 | ~58.6 s | 音樂軌 |
| 9 | 22 | ~0 s | 空樁 |
| 10 | 846,862 | ~19.2 s | 音樂軌 |

→ 7 首實體樂曲。第一段 RIFF 起於檔案 offset `0x816`(實測),`fmt` 為 PCM/2ch/22050/8-bit。

### 2.2 `STREAMHD.LBX`(~21 MB,21 資產,實測)

20 首實體樂曲(asset 0 為 22-byte 空樁),單曲 ~13–43 s。`HD` 推測為「high-detail / 另一組情境音樂」(名稱臆測,**未驗證** HD 與非-HD 的實際差異;別憑檔名下定論)。

> 兩檔合計 ~50 MB 原始 PCM。8-bit/22 kHz 音質先天有限,這是原版事實,**不要**試圖「AI 升頻 / 重新合成」冒充原曲(rulebook 93 鐵則:配樂用原版真實素材)。

---

## 3. 音效資產(實測)

| 檔案 | 大小 | 資產數 | 內容 | 格式 |
|---|---|---|---|---|
| `SOUND.LBX` | ~4.25 MB | 69 | UI / 泛用音效 | **內嵌 RIFF/WAVE**(69 個 RIFF、68 個 WAVEfmt,實測);規格同 §2 |
| `CMBTSFX.LBX` | ~2.29 MB | 79 | 戰鬥音效 | **無 RIFF header**(0 個 RIFF,實測)→ 疑為裸 PCM 或自訂小 header |
| `SPHERSFX.LBX` | ~0.53 MB | 2 | Antaran / 球體音效 | **無 RIFF header** |

- `SOUND.LBX` 每個資產直接是完整 WAV,抽出即可用。
- `CMBTSFX.LBX` / `SPHERSFX.LBX` 資產**沒有 RIFF 包裝**:資產開頭是一段疑似自訂 header(如 `CMBTSFX` asset 0 開頭 `49 00 49 00 …` 不是 `RIFF`),需比對 openorion2 是否有對應解碼;**格式尚未逐位元組驗證**(待辦)。合理假設是同為 22 kHz/8-bit 裸 PCM 樣本 + 少量 metadata,但**別在驗證前當定論**。

> **待辦(靜態溯源)**:查 openorion2 `src/` 是否有 SFX/sound 資產解碼路徑(初步 grep 未見音訊解碼;openorion2 對音訊著墨少,`gui.h` 僅 `// TODO: Add support for transition audio`)。若 openorion2 無現成解碼,`CMBTSFX/SPHERSFX` 的 header 要自行反追(參考 shikadi ModdingWiki 的 MOO2 頁 / LBX 音訊條目)。

---

## 4. ebiten 音訊能力與限制

`github.com/hajimehoshi/ebiten/v2/audio`(來源:官方 pkg.go.dev):

- **播放後端**:底層 `ebitengine/oto`,跨平台(含 Linux/Windows/Web)。
- **內建 decoder**:`audio/wav`(WAV)、`audio/mp3`(MP3)、`audio/vorbis`(Ogg Vorbis)。**沒有** MIDI/XMI decoder。
- **內部 PCM 格式**:串流須為 **signed 16-bit LE 立體聲** 或 **32-bit float LE 立體聲**。
- **WAV decoder**:接受 1/2 聲道、**8-bit 或 16-bit** LE PCM,會轉成 2 聲道 16-bit → **原版 8-bit WAV 可直接吃**。
- **取樣率**:程序內只能有**一個** `audio.Context`,固定一個 sample rate(如 44100 或 48000)。`wav.Decode` 會**自動重取樣**把 22050 Hz 對齊到 context 取樣率(亦可用 `audio.Resample*` 手動重取樣)。
- **裸 PCM(無 header)**:若餵無 RIFF 的資料,ebiten 要求 **signed 16-bit LE 立體聲**(或 32-bit float)。原版是 **unsigned 8-bit**,故裸 PCM 音效**要先轉成 signed 16-bit** 再包成 reader。

**含意**:ebiten 不吃 MIDI —— 這正是為何「原版音樂已是 PCM」是好消息;完全繞開 MIDI 合成。

---

## 5. 整合路線(逐一列可行性與取捨)

### 路線 A —— 直接抽 PCM WAV → ebiten 播(音樂主路線,**推薦**)

- 從 `STREAM.LBX` / `STREAMHD.LBX` / `SOUND.LBX` 用既有 `internal/lbx` 抽出每個資產(已是 RIFF WAV)。
- 交給 `wav.DecodeF32` / `wav.Decode`;8-bit + 22 kHz 由 decoder 自動處理與重取樣。
- **優點**:零轉檔、保真(就是原版位元)、程式最單純。**缺點**:~50 MB 原始 WAV 體積大;不宜打包進 binary/repo(且有版權,見路線 F)。

### 路線 B —— 離線預轉 Ogg Vorbis 壓縮(**推薦搭配 A 的優化**)

- 建置期把抽出的 WAV 用 **ffmpeg(docker)** 轉 Ogg Vorbis:`ffmpeg -i in.wav -c:a libvorbis -q:a 3 out.ogg`。
- ~50 MB PCM → 數 MB;執行期用 `audio/vorbis` 播,取樣率仍由 context 對齊。
- **優點**:體積小、ebiten 原生支援、串流解碼省記憶體。**缺點**:有損(但來源已是 8-bit,聽感差異小);多一個建置步驟。
- **注意**:轉檔在 docker 內(對齊本專案 [HARD] 編譯/工具走 docker),別污染系統。

### 路線 C —— 無 header 音效的處理(`CMBTSFX` / `SPHERSFX`)

- 先反追其資產 header(§3 待辦)確認 sample rate / bit / 聲道與 data 起點。
- 取出裸樣本 → 轉成 ebiten 要的 **signed 16-bit LE 立體聲**(8-bit unsigned → 減 128 再 `<<8`;單聲道 → 複製成雙聲道),包成 `io.Reader` 播;或乾脆在建置期補上標準 WAV header 後走路線 A/B。
- **取捨**:一次性寫個轉換工具即可;屬「格式對齊」工程,不難但需先驗證 header。

### 路線 D —— XMI→MIDI→SoundFont / XMI→OPL 離線渲染(**本專案不需要,僅存為理論路線**)

- 若某音樂真的只有 XMIDI 序列(§1 不確定項),流程為:抽 XMI → 轉標準 MIDI → 用 SoundFont/`fluidsynth`(docker)或 libADLMIDI + 原版 `SAMPLE.OPL` 音色離線渲染成 Ogg → 走路線 B。OPL 音色路線最能還原 90 年代 AdLib 質感。
- 這正是 `game-promo-video-ffmpeg` skill 記錄的 MIDI+SoundFont 抽樂經驗可複用之處。
- **現況**:因主配樂已是 PCM(§2),**此路線目前無對象**。列出以備:① 日後發現 XMIDI-only 的音樂;② 想保全「原版 MT-32/OPL 音色版」音樂做歷史保存(rulebook 83 完整性)時再啟用。Go 無現成 XMI decoder(gomidi/midi 等只做標準 SMF),XMI→SMF 需自寫或借用外部工具(如 UXX/Munt/libADLMIDI)。

### 路線 E —— 即時 MIDI / OPL 合成(**不推薦**)

- 在 Go 內即時跑 MIDI/OPL 合成器。複雜度高、CPU 成本、音色還原難。
- 本專案配樂是固定曲目 → 用 A/B 的預渲染即可,無理由即時合成。

### 路線 F —— 版權 / 素材來源鐵則(貫穿全部,[HARD])

- 配樂/音效**一律用原版真實素材**,**不得自產逼近**(自寫合成器、AI 生成曲頂替)—— rulebook 93 鐵則;retro 玩家對原曲有記憶,假的一聽就穿。
- **但**原版音樂是他人著作權:抽出的 WAV/Ogg 音樂與音效**一律 gitignore、不入 github patch repo**(本專案 repo 本就只放 patch、不放原始素材)。玩家用自己合法持有的原版遊戲檔在本機抽取、本機播放。
- 若日後要做**對外公開**的推廣影片並配上這些原版音樂,先向使用者提示 IP 風險再決定(rulebook 93 但書)。

---

## 6. 建議(落地順序)

1. **音樂**:路線 A 先跑通(抽 `STREAM.LBX` WAV → ebiten `wav` 播,驗證出聲、循環正確),再加路線 B(建置期 docker ffmpeg 轉 Ogg)縮體積。音樂資產 gitignore。
2. **UI/泛用音效**:`SOUND.LBX` 走路線 A(已是 WAV,直接播);短音效可全載入記憶體(`audio.NewPlayerFromBytes`)。
3. **戰鬥音效**:先做路線 C 的 header 反追(§3 待辦),確認 `CMBTSFX/SPHERSFX` 格式後補 WAV header 或轉 16-bit。
4. **MIDI(D/E)**:目前不做;僅在確認有 XMIDI-only 音樂、或要保全 MT-32/OPL 音色版時再議。
5. 音樂/音效列 WORKLIST 後期項(可玩畫面與輸入優先);全程遵守版權鐵則(路線 F)。

## 7. 待驗證 / 不確定(誠實標註)

- [ ] `CMBTSFX.LBX` / `SPHERSFX.LBX` 資產的實際 header 與 PCM 規格(§3)—— 靜態反追 + 對照 openorion2 / shikadi ModdingWiki。
- [ ] 是否有任何音樂/jingle 走 XMIDI 而非 PCM(§1 不確定項);`SAMPLE.*` 音色庫是否真被遊戲用到。
- [ ] `STREAMHD` 與 `STREAM` 的關係(HD 是高音質?不同情境音樂?)—— 名稱臆測,待實聽/比對確認。
- [ ] 各曲目對應的**遊戲情境**(主選單/星圖/戰鬥/勝負畫面)—— 播放觸發邏輯要另查 openorion2 或原版行為,別憑檔名/順序假設。
- [ ] 每首曲子 8-bit unsigned → 16-bit 的實際位元轉換與循環點(loop point)是否需特別處理。

---

## 來源

- 原版遊戲檔逆向:`original_game/Master_of_Orion_II_-_Battle_at_Antares_1996.zip`(執行檔字串、`.MDI`/`.DIG`/`SAMPLE.*` 清單、LBX 音訊資產大小、WAV `fmt` 規格、`SCORE.LBX` 為 640×480 影像 均本專案實測)。
- LBX 容器格式:本專案 `docs/tech/lbx-format.md`(magic `0xfead`、offset 表、影像 header)。
- ebiten 音訊:[audio](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/audio) / [audio/wav](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/audio/wav) / [ebitengine/oto](https://github.com/ebitengine/oto)(pkg.go.dev,官方)。
- 社群考據(MOO2 音樂為 22 kHz/8-bit PCM、STREAM/STREAMHD 為 WAV):[VOGONS: Improving music Quality in MOO2](https://www.vogons.org/viewtopic.php?t=38812)、[ModdingWiki: Master of Orion II](https://moddingwiki.shikadi.net/wiki/Master_of_Orion_II)、[ModdingWiki: LBX Format](https://moddingwiki.shikadi.net/wiki/LBX_Format)、[The Orion Nebula 論壇:LBX 音訊抽取](https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=210)。
- LBX 抽取工具參考:[LbxExtractor](https://github.com/MOO2Extractor/LbxExtractor)、[lbxtract-go](https://codeberg.org/bazub/lbxtract-go)。
- 相關鐵則:rulebook 83(retro 完整性)、rulebook 93(配樂用原版真實素材、IP 但書)、rulebook 62(靜態溯源)、rulebook 63(code/資產是唯一真相,清作廢斷言)。
