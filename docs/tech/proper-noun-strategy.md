# 專有名詞池在地化策略

《銀河霸主 II》的敘述性文字(科技、事件、外交、百科)已全數翻譯完成,唯獨四個「名詞池」性質特殊,
翻或不翻各有代價,需要獨立策略而非套用一般文字的翻譯流程。本文件盤點四池:母星名(13)、
艦名(672)、隨機星名(829)三池均已翻譯落地,RACENAME(104)依既有先例定案保留原文。

## 名詞池總覽

四池皆由 `openorion2/src/lbx.cpp` 於啟動時整批載入,彼此互不相關:

| 池 | 來源 (`.lbx`) | asset | 條數 | 內容 | 取用時機 |
|---|---|---|---|---|---|
| 母星名 | STARNAME | 0 | 13(固定) | 13 種族各自的母星名稱 | 開場種族選定後立即顯示 |
| 隨機星名 | STARNAME | 1 | 829 | 銀河生成時隨機分配給非母星恆星 | 每局隨機抽取,每次遊玩只會看到其中一小部分 |
| 艦名 | SHIPNAME | 0 | 672 | 新艦下水時隨機命名 | 每局隨機抽取,同上 |
| AI 統治者姓名 | RACENAME | 0 | 104 | 電腦對手的個人姓名 | 開場依種族/難度隨機指派給該局 AI 玩家 |

## 1. 策略決策

| 名詞池 | 建議策略 | 理由 |
|---|---|---|
| 母星名(13) | **意譯優先,無對應天文名則音譯** | 固定 13 個,每局都會看到、記憶度高,是種族身份的一部分(如「人類來自太陽」)。多數名稱借用真實星座/亮星名,中文天文界已有既定譯名,直接沿用比自創音譯更準確、更有「原來如此」的閱讀樂趣。純虛構的名稱(如 Sssla、Kholdan)才退回音譯。 |
| 隨機星名(829) | **已落地(2026-07-11)** | 完整 dump 829 條後逐條核對,發現這池並非泛用「短音節隨機生成」,而是大量借用真實恆星/星座學名、克蘇魯神話(Lovecraft/Dunsany/Clark Ashton Smith)專有名詞、圍棋術語彩蛋、歷史神話人名地名,只有一部分是真正的 2–4 字母虛構音節。查得到真名/既定譯名的一律優先意譯(天文/圍棋/神話各依領域慣用中文譯名),查無實據的虛構音節才走規則化音譯。829 條英文名彼此互不重複。 |
| 艦名(672) | **已落地(2026-07-11 本輪)** | 樣本(Sparrow I、Starling、Hawk I、Falcon II…)顯示這池不是隨機音節,而是「真實英文詞(鳥類/軍武詞彙)+ 羅馬數字流水號」的固定表。實際採用方案:基底詞(190 組,去重後 535 條唯一英文名)查中文既定慣用譯名意譯(麻雀、獵鷹、鷹、蒼鷹…),純虛構複合詞按字面意譯保持簡短威武感(Swiftwing→疾翼、Sharpbeak→利喙、Bloodfoot→血足、Dreadwing→懼翼);羅馬數字流水號原樣保留、與基底詞空格分隔(Falcon II→獵鷹 II),不轉阿拉伯數字、不加「號」;重複英文名一律譯為相同中文。取代原本硬編的 10 個中文艦名循環。 |
| AI 統治者姓名(104) | **保留原文** | 這是本專案既有、一致的判斷:`officer.tsv`(軍官姓名)與 `credits.tsv`(製作名單人名)都選擇保留原文,只翻譯職稱/頭銜,不翻「人名」本身。RACENAME 屬同一類別——它是電腦對手的個人身份標籤,玩家透過反覆遊玩記住「Ariel 很好戰」這類跨局印象,音譯反而打斷這層辨識度,且翻了 104 個人名也不會讓任何玩法或劇情文字更清楚。維持既有先例,不另立新規則。 |

## 2. 母星名譯案(13 個)

先確認每個母星對應的種族(依 `openorion2/src/lbx.cpp` 的 `_homeworlds`/`STARNAME asset0` 與遊戲設定考證,
種族百科文字已在 `assets/i18n/help.tsv` 譯出,可交叉核對種族特徵):

| 英文 | 對應種族 | 建議中文 | 理由 |
|---|---|---|---|
| Sol | Human(人類) | 太陽 | 「Sol」即拉丁文「太陽」,人類母星就是地球所在的太陽系,直接借用中文既有天文名,玩家一眼認出「這是我們自己的家」。 |
| Altair | Alkari(阿爾卡里人,鳥類種族) | 牛郎星 | Altair 是牛郎星的國際通用星名,中文天文界譯名已固定;牛郎織女傳說裡「喜鵲搭橋」又恰好與 Alkari 的鳥類設定隱隱呼應,借用真實星名比自創音譯更有味道。 |
| Ursa | Bulrathi(布拉希人,熊型種族) | 大熊座 | Ursa(拉丁文「熊」)是大熊座的學名詞根,Bulrathi 本身就是熊人,譯名與種族設定完全對上,沿用既有中文星座譯名最自然。 |
| Draconis | Elerian(伊勒利安人,人形神秘者) | 天龍座 | Draconis 是天龍座(Draco)的星表命名形式(如 Alpha Draconis),中文譯名固定;雖然 Elerian 外觀不是龍形,但這是「借真實星名」而非「按種族外貌意譯」的案例,保留原始星圖脈絡。 |
| Nazin | Darlok(達洛克人,變形間諜) | 納辛 | 純虛構名稱,無對應真實星名。音譯取中性偏冷的字面(納、辛),呼應 Darlok 母星「終年陰雲、恆星昏暗」的設定基調,不用過度陰森的字避免顯得刻意。 |
| Kholdan | Klackon(克拉肯人,蜂巢型昆蟲種族) | 科爾丹 | 純虛構名稱。音譯維持中性、略帶古樸感,貼合克拉肯人地底蜂巢文明的異星感。 |
| Meklon | Meklar(麥克拉人,機械化種族) | 麥克隆 | 與種族名 Meklar 同詞根(Mek-),沿用既有種族譯名的「麥克」音節,「隆」字帶機械運轉的聽感,音義兼顧且與角色譯名一致。 |
| Fieras | Mrrshan(姆爾沙人,貓型戰士種族) | 菲拉斯 | 純虛構名稱,標準音譯。英文原文與「fierce(兇猛)」諧音,呼應 Mrrshan 好戰的武士文化,但中文譯名維持音譯不強行意譯,避免過度發揮。 |
| Mentar | Psilon(賽隆人,心靈研究種族) | 門塔爾 | 音譯貼合英文發音,「門塔」音近「mental(心智)」,與賽隆人專精心靈/研究科技的種族設定隱約呼應。 |
| Sssla | Sakkra(薩克拉人,爬蟲型種族) | 斯斯拉 | 原文的三重 S 明顯是刻意做出的嘶聲擬聲效果(爬蟲類特徵),音譯保留「斯斯」的氣音感,盡量還原這個設計巧思而非簡化成單一「斯」。 |
| Cryslon | Silicoid(矽晶人,結晶生命體) | 克萊斯隆 | 純虛構名稱,標準音譯。種族名「矽晶」本身已承載「水晶/礦物」的意象,母星名不必重複意譯,維持音譯與其他虛構母星名一致。 |
| Trilar | Trilarian(特里拉人,水棲跨維度種族) | 特里拉 | Trilar 本身就是種族名 Trilarian 的縮寫詞根,母星名與既有種族譯名(特里拉)直接沿用同一組字,不另造新詞。 |
| Gnol | Gnolam(諾嵐人,矮人型商業種族) | 諾爾 | Gnol 是種族名 Gnolam 的詞根,沿用既有種族譯名共享的「諾」字,搭配簡短音譯「爾」,與種族名一脈相承又保留母星獨立辨識度。 |

可直接寫入 TSV 的英文/中文兩欄清單:

```
Altair	牛郎星
Ursa	大熊座
Nazin	納辛
Draconis	天龍座
Gnol	諾爾
Sol	太陽
Kholdan	科爾丹
Meklon	麥克隆
Fieras	菲拉斯
Mentar	門塔爾
Sssla	斯斯拉
Cryslon	克萊斯隆
Trilar	特里拉
```

> 落地時建議另立 `assets/i18n/starname.tsv`(或併入既有種族相關 TSV),並在 `docs/tech/string-sources.md`
> 的 STARNAME asset0 列補上 TSV 檔名與「✅ 完成」狀態——本文件只定案譯名,不動 TSV/程式碼。

## 3. 隨機池的可行性評估

### 隨機星名(829,STARNAME asset1)——已落地(2026-07-11)

完整 dump 全部 829 筆並逐條核對後,發現這份表遠不只是「2–4 字母短促虛構音節」的泛用假設(舊版本文件
的樣本 Uz、Pax、Ecu、Lir、Yed、Kif、Zin、Nir、Goi、Tao、Vij、Ras 只是全表最短的一小段),實際組成是
`openorion2/src/lbx.cpp` 讀成的一份固定字串表(依英文字母長度 3→10 排列,非即時組字),內容包含:

1. **真實恆星/星座學名**:Vega(織女星)、Rigel(參宿七)、Sadr(天津一)、Deneb(天津四)、
   Aldebaran(畢宿五)、Fomalhaut(北落師門)、Polaris(北極星)、Arcturus(大角)、Regulus(軒轅十四)、
   Procyon(南河三)、Algol(大陵五)、Mizar(開陽)等一線亮星,以及 Draco/Lyra/Crux/Cygni/Taurus/
   Virgo/Cetus/Pegasus/Andromeda 等星座屬格/主格形式。
2. **克蘇魯神話 + Dunsany/Clark Ashton Smith 專有名詞**:Arkham(阿卡姆)、Carcosa(卡爾克薩)、
   Hastur(哈斯塔)、Yaddith(亞底斯)、Kadath(卡達斯)、Ulthar(烏撒)、R'lyeh(拉萊耶)、
   Hyperborea(極北之地)、Zothique(佐希克)、Poseidonis(波塞多尼斯)、Dunwich(敦威治)等,數量
   遠超預期,顯示這份表是刻意致敬 Lovecraft 系文學,不是隨機拼音。
3. **圍棋術語彩蛋**:Joseki(定石)、Fuseki(佈局)、Sente(先手)、Gote(後手)、Atari(打吃)、
   Komi(貼目)、Tesuji(手筋)、Shicho(徵子)、Semeai(對殺)、Kosumi(小尖)、Komoku(小目)、
   Hasami(夾)、Kakari(掛)、Kyusho(急所)、Dame(單官)等十餘個日文圍棋術語混在星名池裡,
   應是開發團隊留給圍棋玩家的隱藏彩蛋。
4. **歷史/神話/科學人名地名**:Galileo(伽利略)、Hawking(霍金)、Ptolemy(托勒密)、
   Herschel(赫歇爾)、Babylon(巴比倫)、Osiris(奧西里斯)、Thoth(托特)、Ishtar(伊絲塔)等。
5. 扣除以上四類,剩下才是真正 2–4 字母的短促虛構音節(Uz、Pax、Ecu、Lir、Yed、Kif、Zin…)。

**採用方案**:查得到真名/既定譯名的一律優先意譯——恆星/星座沿用天文界慣用漢名,圍棋術語沿用中文
圍棋界慣用術語(保留彩蛋樂趣),神話/人名沿用常見譯名或既定音譯;查無實據的虛構音節走規則化音譯
(短促、避生僻字、盡量避免與其他條目撞字)。829 條英文名逐一核對後彼此互不重複。落地產出:
`assets/i18n/starname-random.tsv`(829 條英文/中文對照 + 分類說明)、`internal/shell/starnames.go`
(`randomStarNamePool`,依原始索引順序的 829 條循環池)。`genGalaxy` 已改用此池取代舊有的
二十八宿占位池(`starNamePool` 已移除)。

**把握度較低、建議覆核的條目**(音同字異或真假難辨,已盡力給出最相近譯名,非硬掰冷僻典故):
Nath/Wesat/Wasat/Zuban(疑似阿拉伯星名前綴殘片,但單獨出現時對應哪顆星把握不足)、
Mirak(拼法與真實星名 Merak 有出入)、Scheader(拼法與真實星名 Schedar 有出入)、
Browntes(拼法與神話巨人 Brontes 有出入)、Vhoorl/Oriab/Cathuria/Inganok/Sarkomand/Olathoe/
Lomar/Knyan/Ubboth(Dunsany/Lovecraft 夢境週期地名,中文譯名尚無高度統一版本)、
Mane Go(疑似 mango 諧音雙關,非把握確定)、Chay Foo(組合來源不明)。

### 艦名(672,SHIPNAME asset0)

樣本(Sparrow I、Starling、Pin Feather、Striker、Hawk I、Falcon II…)明顯不是隨機音節,而是
「真實英文詞(以鳥類、軍武風格詞彙為主)+ 羅馬數字流水號」的固定表。去除羅馬數字後綴、重複詞根後,
基底詞彙量估計遠小於 672(常見 MOO2 艦名清單裡鳥類/猛禽詞反覆出現,如 Hawk/Falcon/Sparrow 各自有
I~III 多個變體)。

**可行性優於隨機星名**:
- 基底詞多為英文常用詞,中文已有穩定慣用譯名(Sparrow=麻雀/雀鷹、Falcon=獵鷹、Hawk=鷹、
  Striker=打擊者/突擊艦 等),翻譯品質有把握,不必臨時造詞。
- 羅馬數字後綴(I/II/III)可原樣保留或轉全形羅馬數字,不需要額外翻譯決策。

**落地結果(2026-07-11)**:完整 dump 672 筆後去重統計,基底詞彙量為 190 組(去重唯一英文名 535 條),
落在可一次性處理的規模,已排入本輪翻譯。譯表存於 `assets/i18n/shipname.tsv`(535 條英文/中文對照,
含 provenance 註解),實際採用的 672 條循環池(依原始索引順序、含重複)寫入
`internal/shell/shipnames.go` 的 `shipNamePool`,取代 `internal/shell/session.go` 原本硬編的
10 個中文艦名循環。少數拼字明顯訛誤或高度罕見的西洋兵器/器物名(如 Falchard、Malvosin、
Skein Dbuh、Seibe Bow)已盡力比對最相近的真實詞源(Falchion、historical Malvoisin trebuchet、
Sgian-dubh)給出翻譯,但把握度低於其餘條目,留待之後有更多考據再校正。

## 4. 小結

| 名詞池 | 本輪產出 |
|---|---|
| 母星名(13) | 全部 13 個譯案已定案,可直接落地(見第 2 節 TSV) |
| AI 統治者姓名(104) | 定案保留原文,不另翻譯 |
| 艦名(672) | **已落地**:190 組基底詞(535 條唯一英文名)意譯 + 羅馬數字流水號保留,見 `assets/i18n/shipname.tsv` + `internal/shell/shipnames.go` |
| 隨機星名(829) | **已落地**(2026-07-11):829 條英文名彼此互不重複,真名/圍棋術語/神話專有名詞優先意譯,虛構短音節規則化音譯,見 `assets/i18n/starname-random.tsv` + `internal/shell/starnames.go` |
