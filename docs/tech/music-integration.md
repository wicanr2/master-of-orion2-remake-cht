# 音樂/音效整合可行性(go/ebiten)

CLAUDE.md「其他需求」提到音樂整合。本文盤點原版 MOO2 的音訊格式、ebiten 的音訊能力,與整合路線的取捨。

## 1. 原版 MOO2 音訊格式(依遊戲檔實證)

遊戲 zip(`original_game/…1996.zip`)內的音訊相關檔:

| 檔 | 內容 |
|---|---|
| `SCORE.LBX` | **音樂曲目**(XMIDI / .XMI 格式,Miles Sound System「AIL」的擴充 MIDI)|
| `SAMPLE.AD` | AdLib FM 音色庫(instrument bank，給 OPL2 合成用)|
| `SAMPLE.OPL` | OPL FM 音色定義(AdLib/OPL2/3 音色)|
| `SOUND.LBX` | 一般音效(UI 等,數位取樣 PCM)|
| `CMBTSFX.LBX` | 戰鬥音效 |
| `SPHERSFX.LBX` | 球形武器音效 |
| `SETSOUND.EXE` | 音效卡設定工具(AdLib/SB/MT-32 選擇)|

**關鍵**:MOO2 音樂**不是**數位音檔,而是 **XMIDI 樂譜(SCORE.LBX)+ FM 音色庫(SAMPLE.OPL)**,執行期由 AdLib/OPL FM 合成器即時演奏(或 MT-32/GM 硬體)。這決定了整合難度——不能直接播放,需先合成成波形。

## 2. ebiten 音訊能力與落差

`ebiten/v2/audio` 播放 **PCM / WAV / Ogg Vorbis / MP3** 的取樣波形串流,**不內建 MIDI/XMI 合成或 OPL FM 模擬**。因此:
- **音效(SFX)**:LBX 內是數位取樣 → 抽出轉 PCM/WAV 即可餵 ebiten,直接可行。
- **音樂**:XMIDI + OPL 音色 → 中間必須經過「合成成波形」這一步,ebiten 才能播。

## 3. 整合路線(逐一列可行性與取捨)

### 音樂
| 路線 | 作法 | 優 | 缺 |
|---|---|---|---|
| **A. 離線 OPL 合成預渲染(最保真)** | 抽 SCORE.LBX 的 XMI → 用 OPL2 模擬器(如 DOSBox 的 dbopl、或 libADLMIDI 帶 SAMPLE.OPL 音色)離線渲染成 Ogg → ebiten 播 | 音色**與原版 AdLib 完全一致**;離線一次做完、執行期零負擔 | 需建離線工具鏈(XMI 解析 + OPL 模擬)|
| **B. 離線 SoundFont 預渲染** | XMI→標準 MIDI→用 GM SoundFont(fluidsynth)渲染成 Ogg | 工具成熟(呼應 game-promo-video 的 MIDI+SoundFont 經驗)| **音色偏離原版**(SoundFont≠原版 AdLib FM),失去 90 年代 FM 質感 |
| **C. 執行期即時 OPL 合成** | Go 內嵌 OPL 模擬器即時演奏 XMI | 可動態(隨遊戲狀態換曲/淡入淡出)| 需 Go OPL 合成庫,工作量與風險最高 |

### 音效
- 從 `SOUND.LBX`/`CMBTSFX.LBX`/`SPHERSFX.LBX` 抽數位取樣 → 轉 ebiten 可播格式(PCM/WAV)→ 執行期即時播。直接可行,類似既有 LBX 資產抽取。

## 4. 素材來源鐵則(呼應專案原則)

[HARD] **配樂用原版、不自產**(同 `rulebook/93` promo-video 素材鐵則)。原版音樂是遊戲體驗的一部分,
應忠實還原(路線 A 的 OPL 保真最符合「保全歷史」宗旨),不得用自製或替代配樂冒充。
玩家需自備正版遊戲資料(SCORE.LBX/SAMPLE.OPL 等),與圖形資產同一政策。

## 5. 建議

1. **音效先做**(直接可行、投報高):抽 SFX LBX → PCM → ebiten。
2. **音樂用路線 A(離線 OPL 保真)**:建 XMI→OPL 離線渲染工具鏈(libADLMIDI + SAMPLE.OPL 音色 → Ogg),
   保留原版 AdLib 音色;執行期只播 Ogg,零合成負擔。路線 B(SoundFont)僅作為工具鏈未就緒前的暫代。
3. 音樂/音效整合列 WORKLIST 後期項(可玩畫面與輸入優先);先確認 SCORE.LBX 的 XMI 解析可行性(小 spike)。

> 不確定處:SCORE.LBX 內 XMI 的實際封裝(是否標準 XMIDI、有無 MOO2 專屬包裝)需實際抽一首解析驗證後才能定案;
> 本文依「XMIDI + SAMPLE.OPL AdLib 音色」的檔案證據推論,細節待 spike 確認。
