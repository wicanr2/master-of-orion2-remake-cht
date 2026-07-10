# MOO2 音訊曲目/音效對應表

> 目的:把 `STREAM.LBX` / `STREAMHD.LBX`(音樂)與 `SOUND.LBX`(音效)的 entry,對應到「這是哪首曲/哪個 UI 事件」,供遊戲各場景選對背景音樂與音效。
> 日期:2026-07-10(第二輪,不需人耳聆聽,改用資料+文獻交叉推定)。搭配讀:[`audio-format.md`](audio-format.md)(格式)。
> **驗收原則(鐵律)**:曲名↔entry 的最終對應**必須對原版聆聽確認**。本檔嚴格區分「已驗證(指紋比對)」與「推定(待聆聽)」,不把推理當已對齊(記取專案 `rulebook/65` 教訓)。第二輪把「推定」的證據等級拉高(EXE 架構事實 + 官方曲名慣例 + 本機時長實測),但**仍非曲名級確證**——見第六節「定案表」逐格信心標註。

## 一、已驗證:STREAM.LBX 的 8 條 = 官方原聲帶(byte-size 指紋)

archive.org 的《MOO2: Battle at Antares OST》(Laura Barratt,自遊戲檔抽 22kHz WAV)其 8 個檔名格式為 `cut<LBX偏移>-<byte大小>`。**其 byte 大小逐一等於我方 dump 的 STREAM 條目**,連 LBX offset 遞增順序都吻合 → 確認 STREAM.LBX 就是官方原聲帶本體。

| 我方 dump | bytes | 時長 | archive.org OST 檔(`cut<off>-<size>`) | 對上 |
|---|---|---|---|---|
| stream_01 | 5012208 | 113.65s | cut**2070**-**5012208** | ✓ |
| stream_02 | 5527010 | 125.33s | cut5014278-**5527010** | ✓ |
| stream_03 | 1887988 | 42.81s | cut10541288-**1887988** | ✓ |
| stream_04 | 4910256 | 111.34s | cut12429276-**4910256** | ✓ |
| stream_05 | 5994270 | 135.92s | cut17339532-**5994270** | ✓ |
| stream_06 | 2466700 | 55.93s | cut23333802-**2466700** | ✓ |
| stream_08 | 2583936 | 58.59s | cut25800524-**2583936** | ✓ |
| stream_10 | 846862 | 19.20s | cut28384482-**846862** | ✓ |

(entry 0 = 22-byte 彩蛋槽 `cats rule dogs drool`;07/09 為非 WAV 空槽,故序號跳號。)

**已驗證**:STREAM 這 8 條=玩家熟知的 MOO2 原聲帶長版。**尚未驗證**:每條的曲名(archive.org metadata 無標題)。

## 二、曲名池(last.fm,順序未證)

last.fm 的曲目清單(依播放數排序,非必為 LBX 順序):

```
Theme 1 · Theme 2 · Theme 3            ← 一般銀河/選單主題
Psilon · Meklar · Gnolam · Darlok · Trilarian · …  ← 各種族外交主題
Battle 1 · Battle 2 · Battle 3          ← 戰鬥
```

3 主題 + ~14 種族 + 3 戰鬥 ≈ **20 條 → 正好等於 `STREAMHD.LBX` 的 20 條**。故推定:
- **`STREAMHD.LBX`(20 條,較短 13–42s)= 完整具名音樂集**(含每個種族的外交主題)。
- **`STREAM.LBX`(8 條,較長 42–136s)= 精選/長版**(官方 OST 收錄的就是這組)。

## 三、推定對應(第一輪,★待聆聽確認,勿當定案)

假設 `STREAMHD.LBX` 的 entry 順序 = last.fm 曲目順序(**未證**):

| STREAMHD entry | 時長 | 推定曲名 | 用於場景(推定) |
|---|---|---|---|
| streamhd_01 | 38.54s | Theme 1 | **主選單/標題** |
| streamhd_02 | 19.17s | Theme 2 | 星系圖/一般 |
| streamhd_03 | 42.66s | Theme 3 | 一般 |
| streamhd_04..17 | 19–24s | 各種族主題(Psilon/Meklar/…) | 對應種族外交畫面 |
| streamhd_18..20 | 15–21s | Battle 1/2/3 | 戰鬥 |

> ⚠ 上表**每一列都是待驗證假設**(2026-07-10 第一輪產物)。streamhd_01 是否真為「Theme 1 主選單曲」、各種族順序是否吻合,**都要對原版聆聽或逐條試聽 dump 檔才能定案**。第二輪(見第五節)用資料+文獻交叉推定把部分項目的證據等級拉高,但曲名級身分仍未證實。

## 四、SOUND.LBX 音效(名稱已知,語意用途待確認)

entry0 名稱表已解出 68 個具名音效(見 dump 清單)。UI 相關候選:

| 名稱 | entry | 時長 | 推定用途 |
|---|---|---|---|
| BUTTON1 | 34 | 0.39s | 一般按鈕點擊(**目前接線用此**) |
| BUTTON2 | 35 | 0.88s | 按鈕變體 |
| BUTTON4 | 37 | 0.19s | 短促點擊 |
| BUTTON9 | 51 | 0.39s | 按鈕變體 |
| BUTTONA/B/E | 39/40/41 | 0.10–0.27s | 極短 UI 回饋 |
| SCRENMEC/SCRNMEC3 | 48/49 | — | 畫面切換機械音 |
| XPORTIN/XPORTOUT | 55/56 | ~3s | 傳送進/出 |

其餘(NRG*/EXPL-*/PHOTON/TORP*/MONSTR*/KABOOM…)為武器/爆炸/怪物音,屬戰鬥期,待戰鬥系統接線時對應。

## 五、第二輪定案(2026-07-10):資料+文獻交叉推定(不需人耳聆聽)

任務:不聆聽,改用「openorion2 原始碼 + Orion2.exe 除錯字串 + 官方曲名文獻 + 本機時長實測」四路交叉,把 `cmd/moo2/audiohook.go` 的推定常數再校一輪。逐路證據與結論如下。

### 5.1 openorion2:零 provenance(死路,已排除)

`grep -rn "music\|streamhd\|bgm\|Play.*Music"` 遍搜 `openorion2/src` 只命中 LBX **檔名**字串(如 `mainmenu.lbx`),`gui.h:493` 留有 `// TODO: Add support for transition audio`。**openorion2 完全沒有音樂播放邏輯**——是純渲染殼,不是引擎(見專案 memory `openorion2-is-renderer-not-engine`)。此路無法提供任何場景↔曲目常數,確認排除,不必再回頭查。

### 5.2 Orion2.exe(DOS 版):除錯字串反映真實程式架構(強證據,架構層級)

`Orion2.exe` 是 Watcom 編譯、**未 strip** 除錯/追蹤字串的版本,`strings -n4` 直接讀出完整函式名清單(非反組譯,純靜態字串擷取):

```
Play_Background_Music_   Play_Combat_Music_        Start_Diplomacy_Music_
_diplomacy_good_music    _diplomacy_bad_music      _diplomacy_current_music
Fade_Music_Up / Down     Register_Music_Callback_  Play_Streaming_Music_1H
```

三個可直接下結論的架構事實:

1. **`Play_Combat_Music_` 與 `Play_Background_Music_` 是兩個獨立函式**——戰鬥音樂是專屬派發,不是背景樂的延伸或子集。支持「combat 應該選一條與一般場景曲截然不同的曲目」。
2. **`Start_Diplomacy_Music_` 搭配 `_diplomacy_good_music` / `_diplomacy_bad_music` 兩個獨立變數**——外交畫面的音樂**依當下與該族關係好壞切換**,不是「每族固定一首」的單曲模型。目前 remake 用單一 `bgmDiplo` 常數是**簡化實作**,不是曲目選錯,而是好/壞分支尚未做(見「待辦」)。
3. `DIPLOMSE/DIPLOMSF/DIPLOMSG/DIPLOMSI/DIPLOMSP/DIPLOMSS.LBX` 經確認是外交**文字**的語言別(英/法/德/義/波/西),與種族無關,不要誤讀成「6 個外交場景」。

（`ORION95.EXE` 為 PE32,無 COFF symbol table,對應不到函式位置,故用 DOS 版字串;兩者都在同一 gamedata 目錄,SETSOUND.EXE 內只有 MIDI 驅動設定字串,與曲目對應無關,亦排除。）

### 5.3 官方曲名文獻:「Race-Peace / Race-War」命名慣例(佐證,非本作品直接證據)

Steam《Master of Orion: Soundtrack & Score》(App 468020,2016 重製版,同一作曲家 Dave Govett 掛名)曲名清單含 `PSILON Race-Peace`(#32)、`PSILON Race-War`(#33)、`MEKLAR/DARLOK/HUMAN Race-Peace`/`Race-War` 等——**證實「每族一對和平/戰爭曲」是這個系列跨代的設計慣例**,與 5.2 的 `_diplomacy_good_music`/`_diplomacy_bad_music` 架構完全吻合。

> ⚠ **重要限制**:這是 2016 重製版的曲目與編號,**不是** 1996 年 MOO2 原版 `STREAMHD.LBX` 的內部順序或內容——不可把該編號直接套到本檔案的 entry index。只作為「和平/戰爭配對」這個結構性事實的獨立佐證,不作為身分/順序證據。VOGONS、Orion Nebula 論壇、ModdingWiki LBX Format 頁面經查**均未記載** `STREAMHD.LBX` 的 entry↔曲名對應(已查、無收穫,非漏查)。khinsider / VGMdb 曲目頁面回傳 403,無法直接讀取,退而用 WebSearch 摘要驗證,但同樣未取得 LBX 內部序號。

### 5.4 本機時長實測:20 條 entry 的精確秒數與配對訊號(自產訊號,弱證據)

用 `internal/audio.LoadMusic` 實際解碼 `STREAMHD.LBX`(21 entries,20 條可解出 WAV,entry 0 為非 WAV 彩蛋槽,與 `STREAM.LBX` 同構),量出每條精確秒數(`musicClips[i]` 即 entry `i+1`):

| clip idx | entry | 秒數 | clip idx | entry | 秒數 |
|---|---|---|---|---|---|
| 0 | 1 | 38.54 | 10 | 11 | 24.02 |
| 1 | 2 | 19.17 | 11 | 12 | 13.28 |
| 2 | 3 | 42.66 | 12 | 13 | 21.32 |
| 3 | 4 | 24.05 | 13 | 14 | 38.41 |
| 4 | 5 | 19.21 | 14 | 15 | 22.88 |
| 5 | 6 | 19.20 | 15 | 16 | 21.34 |
| 6 | 7 | 28.35 | 16 | 17 | 14.61 |
| 7 | 8 | 42.66 | 17 | 18 | 21.32 |
| 8 | 9 | 19.19 | 18 | 19 | 15.99 |
| 9 | 10 | 24.01 | 19 | 20 | 20.58 |

觀察到的近乎相同時長配對(誤差 ≤0.13s):`(0,13)` 38.54/38.41、`(2,7)` 42.66/42.66(精確相同)、`(4,5)` 19.21/19.20、`(1,8)` 19.17/19.19、`(9,10)` 24.01/24.02、`(12,17)` 21.32/21.32(精確相同)——**6 組候選配對**,與 5.2/5.3 的「和平/戰爭配對」結構吻合,是支持性訊號。其餘 8 條(idx 3,6,11,14,15,16,18,19)無明顯配對夥伴,較可能是不需要好壞分支的曲目(主題/戰鬥/其他)。

> ⚠ **這是本檔自產的推論訊號,不是外部 oracle**(對應 `rulebook/65` 的告誡)。時長相近不等於「同一首歌的兩個版本」,也不能反推是哪一族、哪個場景——只能當「STREAMHD 內部確有配對結構」的弱佐證,不可當「某 entry = 某族/某場景」的證據使用。

### 5.5 定案表(2026-07-10 第二輪)

| 場景 | 目前值(clip idx) | 依據 | 信心 |
|---|---|---|---|
| 主選單/標題(`bgmMenu`) | 0(entry1,38.54s)**不變** | entry 1 為首條、屬「長曲」群(38–43s),契合原聲帶慣例「Theme 1 開場」;無新證據推翻 | 中 |
| 星系圖/一般經營(`bgmGalaxy`) | 2(entry3,42.66s)**改自 1** | 原值 entry2(19.17s)落在「中長曲群」(疑似種族主題時長區間),不合長時間迴圈播放的一般場景樂;entry3 屬獨立長曲(與 entry8 同 42.66s 成對,但本身可單獨當一般場景樂用) | 低(時長分群推論,非曲名確證) |
| 外交(`bgmDiplo`) | 3(entry4,24.05s)**不變** | 落在「中長曲群」(19–24s,疑似種族主題區間);5.2/5.3 已證實外交音樂本應依關係好壞切換,目前仍是單一常數,是**簡化實作**而非曲目選錯——好壞分支未做,列入待辦 | 中(場景分類有架構佐證;曲目身分未證) |
| 戰鬥(`bgmCombat`) | 16(entry17,14.61s)**改自 17** | 原值 entry18(21.32s)與 entry13(21.32s)精確同長,較像落入「配對曲池」而非獨立戰鬥曲;entry17(14.61s)無配對夥伴、短促,較符合 `Play_Combat_Music_` 獨立分派、短迴圈的直覺 | 低(時長分群推論,非曲名確證) |
| 勝利/失敗、安塔蘭 | 未指定 | 查無 openorion2/EXE 字串/文獻對應這兩個場景的專屬曲目函式或曲名;`Play_Streaming_Music_1H`、`Set_Music_For_Game_Popup_` 等字串顯示可能走「彈出視窗音樂」而非獨立 BGM 軌,但無法定位到具體 entry | 低(維持未指定,不硬猜) |
| 各族專屬外交曲(Psilon/Meklar/…) | 未細分(仍共用 `bgmDiplo`) | 5.4 的 6 組配對訊號支持「STREAMHD 內有多族配對曲」存在,但無法反推「entry X = 哪一族」;沒有可靠索引可查 | 低(結構存在,身分未知) |

**結論**:menu/diplo 維持原推定值(有時長輪廓/架構佐證,無新證據推翻);galaxy/combat 各改了一個索引(時長分群推論支持,仍非曲名確證)。四項皆非「聽過確認」等級——**驗收仍需第六節的人耳試聽**,本節只把「先猜哪個」的依據從純 last.fm 播放量排序,提升到「EXE 架構事實 + 官方曲名慣例 + 本機時長分群」三路交叉,誠實地說,confidence 仍多為中/低。

## 六、如何驗收(給使用者/後續)

1. 抽出全部音檔到可試聽目錄:
   ```bash
   docker run --rm -v "$PWD:/src" -v "$PWD/.docker-cache/go:/go" \
     -v /home/anr2/moo2-private-build/gamedata/mastori2:/data:ro -w /src moo2-ebiten \
     bash -c 'go build -buildvcs=false -o /tmp/m ./cmd/moo2 && xvfb-run -a /tmp/m -audiodump /out -data /data'
   # 或桌面直接跑 AppImage 後用 -audiodump 到家目錄
   ```
2. 用播放器聽 `streamhd_01..20` 與 `stream_01..10`,把「哪條是主選單曲、哪條是戰鬥曲、各種族主題」回填第五節的定案表,把推定改成已驗證。
3. 聽 `sound_034_BUTTON1` 等,確認 UI 點擊音選得對。

## 待辦

- [ ] 對原版/dump 聆聽,定案 STREAMHD 20 條的曲名與場景(第五節仍為中/低信心推定)。
- [ ] 定案 STREAM 8 長版各自曲名(可與 STREAMHD 同名長版對照)。
- [ ] 決定遊戲各場景用 STREAM(長版)或 STREAMHD(完整集):主選單、星系圖、各種族外交、戰鬥。
- [ ] SOUND 各 BUTTONx 的實際 UI 用途區分。
- [ ] 外交音樂補做「好/壞關係切換」(對應 Orion2.exe 的 `_diplomacy_good_music`/`_diplomacy_bad_music`),目前 `bgmDiplo` 是單一固定曲的簡化實作。
- [ ] 若能取得 khinsider/VGMdb 的可讀頁面(目前 403),或找到 STREAMHD 逐條播放對照表,回填第 5.4 節的「哪個配對=哪一族」。
