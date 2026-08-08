# MOO2 音訊曲目/音效對應表

> 目的:把 `STREAM.LBX` / `STREAMHD.LBX`(音樂)與 `SOUND.LBX`(音效)的 entry,對應到「這是哪首曲/哪個 UI 事件」,供遊戲各場景選對背景音樂與音效。
> 日期:2026-08-08(第四輪,場景對應改用反組譯立即數並接進 remake)。搭配讀:[`audio-format.md`](audio-format.md)(格式)。
> **證據等級**:「場景播哪一首」是**執行檔的立即數**(第三節),一手證據,remake 已照它接線。
> 「這一條叫什麼名字」仍未定案,需要人耳——但那是命名,不影響行為(第六節)。
> ⚠ 本檔第五節保留 2026-07-10 那一輪的**結構性**結論(戰鬥音樂獨立分派、外交依關係切換、
> STREAMHD 有配對曲),它們與第三節一致;那一輪的**逐場景猜測值**已刪除,因為第三節推翻了它們。

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

## 三、場景 → 曲目:反組譯直接讀出來的(remake 已接線)

**場景播哪一首不是推定,是執行檔裡的立即數**(第 73 項(音樂場景表))。曲**名**仍未定案,但那是命名,不影響行為。

三個入口:

| 函式 | 位址 | 行為 |
|---|---|---|
| `Play_Streaming_Music_` | `0x24677` | 指定曲目。**編號 ≤100 → `STREAM.LBX` 的 entry;>100 → `STREAMHD.LBX` 的 entry(索引 = 編號 − 100)** |
| `Play_Background_Music_` | `0x2484F` | `clock() % 3 + 1` → **STREAM 1/2/3 隨機**(15 個呼叫端) |
| `Play_Combat_Music_` | `0x2496C` | `clock() % 3 + 4` → **STREAM 4/5/6 隨機**(只有 `Tactical_Combat_` 呼叫) |

`Play_Streaming_Music_` 的「下一首」參數有兩個哨兵值:`-1` = 沒有下一首;`-2` = 播完接隨機 STREAM 1..3。

| 場景(除錯符號真名) | 曲目編號 | remake 落點 |
|---|---|---|
| `main__0` / 主選單、星圖、多數畫面 | `Play_Background_Music_` → STREAM 1/2/3 隨機 | `playBackgroundMusic()` |
| `Tactical_Combat_` | `Play_Combat_Music_` → STREAM 4/5/6 隨機 | `playCombatMusic()` |
| `Science_Room_` / `_Tech_Select_` | STREAMHD **#17**(播完接隨機 STREAM 1..3) | `research()` |
| `Start_Main_Event_` / `Draw_Event_Screen_` | STREAMHD **#18** | `eventScreen()` |
| `Main_Council_Screen_` | STREAMHD **#19** | `council()` |
| `Main_Antaran_Room_Screen_` | STREAMHD **#20** | `antaranRoom()` |
| `Design_Screen_` / `Draw_Design_Screen_` | STREAM **#8** | `shipDesign()` |
| `Colony_Combat_Screen_` | STREAM **#10** | `newGroundCombatScreen()` |
| `Draw_Diplomacy_Synch_Mode_` | STREAMHD **#`word_19AA44`**(逐族/依關係動態) | `playDiplomacyMusic()`,見第七節 |

### 一道事後才發現的交叉驗證

實測玩家資料夾(`LoadMusic` 的回傳值):

```
stream.lbx   count=11 可用=8  entryIDs=[1 2 3 4 5 6 8 10]
streamhd.lbx count=21 可用=20 entryIDs=[1..20]
```

反組譯點名的 STREAM 曲目是 **1/2/3**(背景)、**4/5/6**(戰鬥)、**8**(艦艇設計)、**10**(殖民地戰鬥)
——**正好就是這八個存在的槽**,而它從來沒點過的 **7 與 9**,正是那兩個非 WAV 的空槽。
若編號指的是「第幾條 WAV」而不是 entry id,這個吻合不會成立。

### 兩個先前的缺,2026-08-08 都補上了(第 82 項(音樂兩個缺))

- **科學室播完接下一首**:`Mixer.PlayBGMOnce` + `tickBGM` 輪詢,對應 `-2` 哨兵。
- **外交「關係好」的逐族曲**:`該族種族索引 + 1`(STREAMHD 1..13)。
  ⚠ 上一輪記成「逐族資料驅動,那張逐族靜態表還沒追出來」——**那張表不存在**。
  `sub_12983`(帝國建立)顯示 `[帝國紀錄+0x25]` 就是**種族索引**本身:
  `Random_(13,1)` 從 13 族裡挑到不重複再寫進去,接著用同一個索引查 `dword_192630[idx*4]`
  取種族名字串。公式就是字面上的「索引 + 1」,沒有中間表。

  仍未追出來的是**切換條件**(原版依什麼判定「關係好/壞」)——那在
  `Start_Diplomacy_Music_` 的呼叫端,不是這兩個變數的賦值處。remake 用關係分數 >= 0
  當分界,那是 remake 的讀法。

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

### 5.5 這一路推到哪裡為止

5.1–5.4 建立的是**結構**:戰鬥音樂是獨立分派、外交音樂依關係好壞切換、STREAMHD 內確有配對曲。
這些結論到今天仍然成立,而且與第三節反組譯讀出來的東西一致。

**但「哪個場景播哪一條」不必再從時長推**——第三節的立即數直接給了答案,
而那條路推出來的結果與時長分群的猜測**並不相同**(例如主選單根本不是固定曲,是 STREAM 1/2/3 隨機)。
留這一節是因為它的結構性結論仍在用;逐場景的猜測值已刪。

## 六、還需要人耳的是什麼

**場景對應不需要**(第三節,反組譯的立即數)。仍需要聆聽的只有**曲名**——「streamhd_04 是不是 Psilon 主題」這一類,而那不影響任何行為。

要試聽的話,抽檔指令:

```bash
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD:/src" -v "$PWD/.docker-cache/go:/go" \
  -v /home/anr2/moo2-private-build/gamedata/mastori2:/data:ro -w /src moo2-ebiten \
  bash -c 'go build -buildvcs=false -o /tmp/m ./cmd/moo2 && xvfb-run -a /tmp/m -audiodump /out -data /data'
```

## 待辦

- [x] **STREAMHD 20 條的場景已定案**——由 `Play_Streaming_Music_` 呼叫端的立即數直接給出(第 73 項(音樂場景表))。仍未定的只有「曲名」(那是命名,不影響行為)。
- [ ] 定案 STREAM 8 長版各自曲名(可與 STREAMHD 同名長版對照)。
- [x] **各場景用 STREAM 還是 STREAMHD:原版自己決定好了**——單一編號空間,≤100 走 STREAM、>100 走 STREAMHD;主選單/星圖走 `Play_Background_Music_`(STREAM 1/2/3 **隨機**,不是固定一首),戰術戰鬥走 `Play_Combat_Music_`(STREAM 4/5/6 隨機)。
- [ ] SOUND 各 BUTTONx 的實際 UI 用途區分。
- [x] 外交「壞關係」音樂已取得呼叫點硬證(`Get_Random(3)+13` → track 13/14/15 三選一,見第七節);「好關係」音樂證實為逐族資料表驅動,尚未取得資料表本身數值。
- [ ] 若能取得 khinsider/VGMdb 的可讀頁面(目前 403),或找到 STREAMHD 逐條播放對照表,回填第 5.4 節的「哪個配對=哪一族」。
- [x] **`_diplomacy_good_music` 的公式已解**:`該族種族索引 + 1`。`[帝國紀錄+0x25]` 是種族索引本身(`sub_12983` 帝國建立時 `Random_(13,1)` 挑到不重複再寫入,並用同一個索引查 `dword_192630[idx*4]` 取名字)。**先前記著的「逐族靜態表」不存在。**
- [ ] 追出 `Start_Diplomacy_Music_` 的**呼叫端**,確認原版依什麼條件判定「關係好/壞」(remake 目前用關係分數 >= 0,是 remake 的讀法)。
      ⚠ **位址要用 IDA 資料庫的線性位址,不是除錯表的 `obj1+` 偏移**——兩者差 `0x10000`(object base),
      這個坑讓第二輪把 `Play_Background_Music_` 誤判成死碼(見 §7.5)。`obj1+0x9082` → 線性 `0x19082`。
- [x] ~~追 offset+0x25 的「逐族預設值」靜態表~~ **那張表不存在**:`+0x25` 就是種族索引(見上)。
- [x] **`Play_Background_Music_` / `Play_Combat_Music_` 的呼叫端已找到**:先前判「零引用」是位址少加 `0x10000`(object base),真址 `0x2484F` / `0x2496C`,共 15 + 1 個呼叫端。

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

### 7.5 三個音樂入口(2026-08-08 用 `.i64` + 除錯符號表解出)

| 函式 | 行為 |
|---|---|
| `Play_Streaming_Music_` @ `0x24677` | 指定曲目。**編號 ≤ 100 → `STREAM.LBX` 索引;> 100 → `STREAMHD.LBX` 索引 = 編號 − 100**(單一編號空間) |
| `Play_Background_Music_` @ `0x2484F` | `clock() % 3 + 1` → **STREAM 1 / 2 / 3 隨機**,15 個呼叫端(主選單、外交、議會、安塔蘭廳、艦艇設計、殖民地戰鬥…) |
| `Play_Combat_Music_` @ `0x2496C` | `clock() % 3 + 4` → **STREAM 4 / 5 / 6 隨機**,只有 `Tactical_Combat_` 叫它 |

`Play_Streaming_Music_` 的「下一首」參數:`-1` = 無;`-2` = 播完接隨機 STREAM 1..3。

完整的場景→曲目表見 `docs/re/01-gap-report.md` 第 73 項(音樂場景表)。

> **方法教訓**(這一段比曲目表更值得留):第二輪曾對 `0x1484f` / `0x1496c` 跑了三種掃描
> (逐 byte 掃 `E8` CALL、掃 4-byte 絕對位址、掃 LE fixup 表)全部零命中,據此判定
> 「這個 build 裡是死碼」。**位址少加了 `0x10000`(object base)**,三種掃描都在錯的地方找。
> 零命中時先做**正對照**——拿一個已知一定被呼叫的函式跑同一套掃描,零命中就知道查法壞了。
> 另一個放大器是掃 `.asm` 文字而不是查 IDA 資料庫的 xref 圖。
