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
- [x] 外交「壞關係」音樂已取得呼叫點硬證(`Get_Random(3)+13` → track 13/14/15 三選一,見第七節);「好關係」音樂證實為逐族資料表驅動,尚未取得資料表本身數值。
- [ ] 若能取得 khinsider/VGMdb 的可讀頁面(目前 403),或找到 STREAMHD 逐條播放對照表,回填第 5.4 節的「哪個配對=哪一族」。
- [ ] 追出寫入 `_diplomacy_good_music`/`_diplomacy_bad_music` 的無名函式(obj1+0x9082 一帶)本身的呼叫點,確認觸發時機(目前只知其被跳到,未知誰呼叫它)。
- [ ] 追出 empire 記錄(stride 0xea9=3753 bytes,base 於全域指標 `ds:0x1ff98`)offset+0x25 欄位的「逐族預設值」靜態表,才能把 `_diplomacy_good_music` 從公式(`該族記錄.byte[0x25]+1`)落到每族的實際 track index。
- [ ] `Play_Background_Music_`(obj1+0x1484f)/`Play_Combat_Music_`(obj1+0x1496c)在全檔案找不到任何呼叫點/絕對位址參照(見第七節 7.4)——若後續要接著查,建議先排除「這兩個函式在這個 build 是死碼」的假設(如:對照 `ORION95.EXE`/其他 patch 版本的同名符號是否同樣零引用)。

## 七、第三輪(2026-07-10):反組譯呼叫點硬證

> 任務:把 5.2 節「靜態字串存在」升級為「呼叫點引數值」。方法:先解出 `Orion2.exe` 的 LE(Linear Executable)可執行檔格式與 Watcom 除錯符號表格式(兩者均無公開 sample 可直接套用現成工具,故本輪手刻解析後用 `objdump -b binary -m i386` 驗證),再用「反向溯源 SOP」(`rulebook/62`)從函式位址往回找呼叫點與引數常數。全程 docker 跑 `objdump`(bounded、named container、跑完即清)。

### 7.1 檔案格式:LE 可執行檔解析結果

`Orion2.exe`(2,644,842 bytes)是 MZ stub + LE(Linear Executable,DOS4GW 版,非 OS/2 的 LX)。關鍵欄位(LE header 在檔案偏移 `0x292e4`):

- 2 個 object:object #1(code)base=`0x10000`,size=`0x160695`,353 頁;object #2(data)base=`0x178000`,size=`0x5dcd0`,20 個實際儲存頁(其餘為 BSS 零填)。
- Object Page Table(每筆 4 bytes:3-byte 大端頁號 + 1-byte flags)經逐筆核對(全部 373 筆),**全部 identity-mapped、未壓縮**(`pagenum == 全域頁序`,flags=0)。
- **踩雷記錄(scope 提醒)**:LE header 內大部分表格指標欄位(`objtab`/`objmap`/`fpagetab`/`frectab`/`impmod`/`datapage`…)是**相對 header 起點**的偏移,不是相對檔案開頭——這與部分公開文件對 `e32_datapage`「絕對於檔案開頭」的描述不同(至少對本檔不成立;可能是 Watcom wlink 產出的變體慣例,或該說法本來就只適用 LX/OS2)。最初用「絕對於檔案開頭」代入算出的位置全部落在 fixup record table 裡(看起來像程式碼但其實是重複的 fixup 記錄樣式,byte pattern `07 10 xx xx 02/01 xx xx xx 00` 反覆出現),靠這個「不像函式、像規律結構」的異常訊號抓出算錯,改用「相對 header」重算後,對應位置立刻變成合法、可完整解碼的 x86 指令——這是本輪唯一一次真的卡住,靠 rulebook/62 SOP(換一種换算再驗證,而非直接判定「格式解不開」)排除。
- 資料頁起點(datapage,相對 header)= `0x6f040` → 絕對檔案位置 `0x98324`。Object #1(code)因 identity mapping,`檔案位置 = datapage絕對值 + object內偏移` 直接成立(已用 373 筆全數核對)。

### 7.2 Watcom 除錯符號表格式(手刻反推,無公開 spec 可套)

除錯符號字串(`Play_Background_Music_` 等)未被 LE header 的 `debuginfo`/`debuglen` 欄位登記(兩者皆為 0),需靠字串掃描定位。逐筆比對後反推出的記錄格式(可變長度,無 terminator,緊接下一筆):

```
[4-byte 位址,object 內偏移,小端]
[2-byte object 編號:0001=code, 0002=data]
[2-byte type(未解出精確語意,proc 觀察到 0x0010/0x000e 都出現過,不影響位址解讀)]
[1-byte class:0x04=procedure, 0x02=data variable]
[1-byte 名稱長度]
[名稱,ASCII,無 NUL]
```

用「namelen 是否等於後面 ASCII 名稱長度」交叉驗證,20+ 筆記錄零例外,判定格式正確。

解出的位址(object 內偏移):

| 符號 | object | 偏移 | 備註 |
|---|---|---|---|
| `Play_Background_Music_` | 1(code) | `0x1484f` | 全檔案唯一出現處=此除錯記錄本身,見 7.4 |
| `Play_Combat_Music_` | 1(code) | `0x1496c` | 同上 |
| `Start_Diplomacy_Music_` | 1(code) | `0xd0d5`(除錯表記載值) | **非真正函式進入點**,見 7.3 |
| `_diplomacy_good_music` | 2(data) | `0x22a3c` | |
| `_diplomacy_bad_music` | 2(data) | `0x22a46` | |
| `_diplomacy_current_music` | 2(data) | `0x22a44` | 順帶解出,非任務原點名單 |
| `_diplomacy_new_music` | 2(data) | `0x22a32` | 順帶解出,用途未追 |
| `_diplomacy_fade_music_flag` | 2(data) | `0x22a38` | 順帶解出,用途未追 |
| `_response_message` | 2(data) | `0x22a40` | **易混淆記錄**:夾在 good_music 和 current_music 之間,本輪一度誤猜是音樂變數,查完整符號表後證實是對話文字 ID,與音樂無關 |

### 7.3 `Start_Diplomacy_Music_`:呼叫點與函式本體(高信心)

除錯表記載位址 `0xd0d5` 反組譯後開頭是 `adc eax,...`、`mov edx,1`……不像函式起點;往下 27 bytes 在 `0xd0f0` 才出現乾淨的 `push ebx;push ecx;push edx;push esi;push edi;enter 0x18,0`,且其正前方剛好是另一個函式的 `ret`——判斷 `0xd0f0` 才是真正進入點,`0xd0d5` 可能對應除錯表裡的其他標記(未進一步查明原因,不影響下面的呼叫點結論)。

用「掃全部 353 頁 code object 的每一個 byte、抓 `E8`(CALL rel32)、算目的位址」的暴力法(涵蓋所有對齊與非對齊情形),在整個 code object 裡**唯一一處**呼叫目的地等於 `0xd0f0`:

- 呼叫點:object1 偏移 `0xa67`(`call 0xd0f0`),呼叫前**沒有任何引數設置**(前面是連續 4 個無參數呼叫:`call 0xc9f20; call 0xc10a4; call 0xd35d; call 0xd0f0; call 0x3dc7c`)——`Start_Diplomacy_Music_` 是 **void 函式**,不接收場景/種族參數。

`Start_Diplomacy_Music_` 本體(`0xd0f0`–`0xd4d1`,以 `ret` 為界確認)是雙層迴圈掃過所有 empire 兩兩配對(迴圈上界 = `word ds:0x21998`,可能是「目前玩家數」),讀每個 empire 3753-byte(`0xea9`)記錄(base 指標在全域 `ds:0x1ff98`)裡偏移 `0x24`/`0x28` 的關係狀態欄位,偵測「關係翻轉」並更新配對記錄;函式本體內**沒有**直接寫入 `_diplomacy_good_music`/`_diplomacy_bad_music` 或呼叫 `Play_*` 系列函式。

### 7.4 `_diplomacy_good_music` / `_diplomacy_bad_music` 的實際賦值(高信心,硬證)

在 code object 裡搜尋所有直接寫入這兩個變數位址(`0x22a3c`/`0x22a46`)的指令,各找到**唯一一處**(皆在 object1 偏移 `0x9082`–`0x90c6` 一段無名函式內,與 `Start_Diplomacy_Music_` 相鄰但非同一函式;此函式本身在除錯表裡對應的符號名未查出):

```asm
; object1+0x908c .. +0x90a8   ── _diplomacy_good_music
908c: movsx eax, di                    ; eax = 種族索引(來源:外層迴圈變數)
908f: imul  eax, eax, 0xea9            ; eax = 種族索引 * 3753(該族記錄的 stride)
9095: mov   edx, [ds:0x1ff98]          ; edx = empire 記錄陣列 base(全域指標)
909b: movzx ax, byte [edx+eax+0x25]    ; ax  = 該族記錄.byte[0x25]   ← 逐族資料
90a1: inc   eax                        ; ax += 1
90a2: mov   ds:0x22a3c, ax             ; _diplomacy_good_music = 該族記錄.byte[0x25] + 1

; object1+0x90ad .. +0x90c0   ── _diplomacy_bad_music
90a8: mov   eax, 0x3                   ; eax = 3(Get_Random 的上界參數)
90ad: mov   word ds:0x22a44, 0xffff    ; _diplomacy_current_music = 0xFFFF(重置為「無播放」哨兵值)
90b6: call  0x111b10                   ; Get_Random(3) → eax = 0,1,2(均勻亂數,見下方驗證)
90bb: add   eax, 0xd                   ; eax += 13
90c0: mov   ds:0x22a46, ax             ; _diplomacy_bad_music = Get_Random(3) + 13
```

**`0x111b10` 驗證為標準亂數函式**:反組譯其本體看到 `0xFFFFFFFF / N` 拒絕取樣門檻計算,接著用乘數 `0x41C64E6D`(=1,103,515,245)、加數 `0x3039`(=12,345)——這正是經典 C 函式庫 `rand()` 的 LCG 常數(POSIX/minstd 慣用值),確認 `Get_Random(N)` 語意 = 回傳 `[0, N-1]` 均勻亂數。

**結論(STREAMHD track index,0-based,對應 `musicClips[i]`)**:

- **外交「壞關係」音樂 = `Get_Random(3) + 13` → track 13、14 或 15 三選一(均勻亂數)。** 高信心,無歧義,可直接寫入常數。
- **外交「好關係」音樂 = 該族 empire 記錄 offset `0x25` 欄位 + 1。** 這是**逐族資料驅動**,不是單一常數;本輪未能追出該欄位的靜態預設值表(欄位在**執行期配置**的 empire 記錄裡,其初始值理論上來自一張「各族預設資料」的靜態表,但本輪未定位到該表——列入待辦)。**確定的結論是**:原版外交音樂本來就不是單一曲目,而是依「目前跟該族關係好/壞」動態切換,且「好關係」進一步依種族不同而不同。

### 7.5 `Play_Background_Music_` / `Play_Combat_Music_`:撞牆,誠實記錄(未解)

依 `rulebook/62` SOP,對這兩個位址跑了三種不同角度的反向溯源,全部零命中(非「查一次沒有就放棄」):

1. **直接呼叫掃描**:353 頁 code object(1,445,888 bytes)逐 byte 掃 `E8`(CALL rel32),計算目的位址——**沒有任何一處**目的地等於 `0x1484f` 或 `0x1496c`。
2. **絕對位址參照掃描**:全檔案(2,644,842 bytes)搜尋這兩個值以 4-byte 小端出現(涵蓋「物件內偏移」與「object base + 偏移」兩種可能編碼)——**除了除錯符號表本身那一筆,沒有任何其他出現**(不論在程式碼或資料區)。
3. **Fixup(重定位)表掃描**:LE 內部 fixup record table(相對 header 偏移 `0xca9` 到 `0x6c328`,約 610KB)搜尋「目標 object=1、偏移=這兩個值」的重定位項——同樣零命中(用已知確定成立的 `Start_Diplomacy_Music_` 真正位址 `0xd0f0` 反測,亦零命中,證實 near-call 不需要 fixup,這條路本來就測不到 near-call 目標,方法論本身沒有問題,只是再次確認「這兩個函式沒有被任何靜態可見的方式引用」)。

**判讀**:這兩個函式在 `Orion2.exe`(DOS 版)這個 build 裡,找不到任何靜態可達的呼叫點或位址參照——最貼近的解讀是**這個 build 裡是死碼/未接線的舊函式**(可能被其他不透過這兩個具名函式的機制取代),而不是「格式解不開」或「要動態才知道」。**不排除**其他可能性(如 self-modifying code、`ORION95.EXE`/其他 patch 版本裡仍有引用),但本輪證據不支持任何一種能繼續往下查的靜態路徑,故如實記錄撞牆於此,不強行歸因。`bgmMenu`/`bgmGalaxy`/`bgmCombat` 三個場景常數**維持第二輪的時長啟發式**,未升級。
