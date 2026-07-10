# 自訂種族(Custom Race)Picks 點數表

## 來源與方法(先講清楚,避免誤用)

1. **`GAME_MANUAL.pdf`(= `moo2_patch1.5/GAME_MANUAL.pdf`,全文已轉存於
   `manual.txt`)的「Custom」章節(印刷頁 18–27,對應 `manual.txt` 第 655–1067 行)
   只有\*\*文字敘述\*\*,**完全沒有列出每個選項實際花費幾點 Picks 的數字**。手冊原文明講數字是畫面上的東西,
   不是印在紙本手冊裡:
   > "To the right of every option is that option's Pick Modifier."(`manual.txt` 677–678 行)

   換句話說:**紙本手冊本身抄不到點數**,這點如實記錄,不臆造。

2. 點數數字改從 **patch 1.5 安裝包內建的官方 modding 參考文件**取得,這是隨 1.5 一起發布、
   給玩家/mod 作者用來改規則的權威資料檔,不是我方推測:
   - `moo2_patch1.5/MOO2-1.50.26.zip → patch/150/docs/config.json`,
     `parameters[].name == "race_pick"` 與 `"number_of_race_picks"` 兩筆的 `"default"` 陣列
     (共 80 個數字,對應 80 個具名欄位,一一對應無歧義)。
   - 交叉核對:`moo2_patch1.5/MOO2-1.50.26.zip → patch/150/docs/MANUAL_150.html`
     「Custom Race Options Picks」小節(修改設定的文字說明,附範例)。範例原文:
     > "changing negative growth -50% for -4 picks to a positive +25% for 2 picks:
     > `race_pick growth1_cost = 2 25;`"
     > "Food +1 for 4 picks is written as: `race_pick farming2_cost = 4; race_pick farming2_value = 2;`"

     這兩句「範例」直接印證了 `growth1`(預設 -4 picks / -50%)與 `farming2`(預設 4 picks / +1 食物)
     兩筆預設值,與 `config.json` 的 `default` 陣列完全吻合 → **兩份獨立檔案互相驗證,不是單一來源臆測**。

3. **版本消歧(重要,呼應 `docs/tech/patch15-cfg-data-source.md` 既有警告)**:
   `MOO2-1.50.26.zip` 裡另外還有 `patch/150/mods/150i/cfg/PICKS_SP.CFG` 與 `PICKS_MP.CFG`,
   是**選用(opt-in)的「150 improved」(150i)mod** 對 Picks 點數的重新平衡,兩者互相之間也不同(SP/MP
   兩套平衡),**且皆與 `config.json` 的預設值不同**(已逐欄位比對,見「附錄」)。
   - `config.json` 的 `default` 位於 `patch/150/docs/`(非 `mods/150i/`),且 `patch/ORION2.CFG`、
     `patch/150/USER.CFG`(未啟用任何 mod 的基本設定檔)都**沒有**覆寫 `race_pick`,
     代表**未啟用 150i mod 時,遊戲就是用 `config.json` 這組預設值**。
   - 據此判斷:`config.json` 的預設值 = 未修改的 classic/1.5 基準點數表;`PICKS_SP/MP.CFG` = 另一個
     選用 mod 的數字,**不可混用**。本檔主表只採 `config.json` 預設值,150i 的另一組數字列在文末附錄
     供你參考取捨。
   - 仍有一點無法 100% 排除:changelog 提過「Gently adjusted racial picks valuation」等字樣,
     但相關條目都明確標註在 `150 improved + picks (150ip)` 底下(150i mod 專屬),目前找不到任何證據
     顯示 base 1.5(不開 150i)的預設值與原版 1.3 不同。**若要對齊 1.3(而非 1.5),建議另外用 1.3
     執行檔或社群 1.3 wiki 資料再核一次**,本檔目前只能保證「這是 1.5 預設(未開 150i)」。

## 起始 Picks / Score

- 起始:**10 Picks、200% Score**(`manual.txt` 673 行;點清空「Galactic Normal」狀態下的畫面顯示值)。
- Picks 規則:選項右側的 Pick Modifier 為正 = 花費 picks,為負 = 退回 picks;**不可靠負值選項換到超過
  10 點的額外 picks,也不可用負的 picks 總額開局**(`manual.txt` 676–681 行)。
- `config.json` 的 `number_of_race_picks` 預設:`maximum_positive_picks = 10`,
  `maximum_negative_picks = -10`,`evolutionary_mutation_bonus = 4`(演化突變科技可再加 4 picks,
  對應手冊 4100 行:「may choose 4 Picks worth of racial specials」)。
- Score 公式,手冊只給例子、沒給封閉公式(如實照抄,不外推):
  - 剩餘 0 picks 未用 → Score 100%(`manual.txt` 688–689 行)
  - 剩餘 5 picks 未用 → Score 150%(`manual.txt` 690–692 行)
  - 完全不選(剩滿 10 picks)→ Score 200%(起始畫面數字,673 行)
  - 三點連線是 100% + 未用picks×10%,但手冊沒有明文寫成這條公式,這裡只列已證實的三個資料點。

## 主表:Picks 選項與點數

點數欄「(退 N)」= 負值,代表選了會**退回** N 點 picks(倒扣的disadvantage);其餘為正值 = 花費。
「效果值」欄的數字是 `config.json` 的 `_value`,**Farming/Money 兩類效果值依 `MANUAL_150.html` 註記需除以 2**
(如實轉述:"the Food and BC bonus values are doubled: -1, 2, 4 translate in game to -½, +1, +2")。

| 主題 | 選項(英文原名 / 建議中譯) | 點數 | 效果值(遊戲內實際數字) | 互斥/備註 | 出處 |
|---|---|---|---|---|---|
| Colonial Production | Population Growth,差(Poor Growth / 人口成長‧差) | 退4 | 成長率 -50% | growth1;與 growth2/3 三選一或不選 | 描述:manual.txt 701–704 行(p.19);點數:config.json `race_pick.growth1_cost/-value` |
| Colonial Production | Population Growth,佳(Good Growth / 人口成長‧佳) | 3 | 成長率 +50% | growth2 | 同上 |
| Colonial Production | Population Growth,優(Great Growth / 人口成長‧優) | 6 | 成長率 +100% | growth3 | 同上 |
| Colonial Production | Farming,差(Poor Farmers / 農業‧差) | 退3 | 每農夫 -0.5 食物(原始值-1,除以2) | farming1 | 描述:707–711 行;點數:`farming1_cost/-value` |
| Colonial Production | Farming,佳(Good Farmers / 農業‧佳) | 4 | 每農夫 +1 食物(原始值2) | farming2 | 同上;範例見 MANUAL_150.html「Food +1 for 4 picks」 |
| Colonial Production | Farming,優(Great Farmers / 農業‧優) | 7 | 每農夫 +2 食物(原始值4) | farming3 | 同上 |
| Colonial Production | Industry,差(Poor Industry / 工業‧差) | 退3 | 每工人 -1 產能 | industry1 | 描述:714–718 行;點數:`industry1_cost/-value` |
| Colonial Production | Industry,佳(Good Industry / 工業‧佳) | 3 | 每工人 +1 產能 | industry2 | 同上 |
| Colonial Production | Industry,優(Great Industry / 工業‧優) | 6 | 每工人 +2 產能 | industry3 | 同上 |
| Colonial Production | Science,差(Poor Science / 研究‧差) | 退3 | 每科學家 -1 研究 | science1 | 描述:721–725 行(p.19–20);點數:`science1_cost/-value` |
| Colonial Production | Science,佳(Good Science / 研究‧佳) | 3 | 每科學家 +1 研究 | science2 | 同上 |
| Colonial Production | Science,優(Great Science / 研究‧優) | 6 | 每科學家 +2 研究 | science3 | 同上 |
| Colonial Production | Money,差(Poor Money / 商業‧差) | 退4 | 每人稅賦承受力 -0.5 BC(原始值-1,除以2) | money1 | 描述:728–731 行;點數:`money1_cost/-value` |
| Colonial Production | Money,佳(Good Money / 商業‧佳) | 5 | 每人 +0.5 BC(原始值1) | money2 | 同上 |
| Colonial Production | Money,優(Great Money / 商業‧優) | 8 | 每人 +1 BC(原始值2) | money3 | 同上 |
| Combat & Espionage | Ship Defense,差(SD-20 / 艦防‧差) | 退2 | -20% | defense1 | 描述:738–742 行(p.20);點數:`defense1_cost/-value` |
| Combat & Espionage | Ship Defense,佳(SD+25 / 艦防‧佳) | 3 | +25% | defense2 | 同上 |
| Combat & Espionage | Ship Defense,優(SD+50 / 艦防‧優) | 7 | +50% | defense3 | 同上 |
| Combat & Espionage | Ship Attack,差(SA-20 / 艦攻‧差) | 退2 | -20% | attack1 | 描述:745–748 行;點數:`attack1_cost/-value` |
| Combat & Espionage | Ship Attack,佳(SA+20 / 艦攻‧佳) | 2 | +20% | attack2 | 同上 |
| Combat & Espionage | Ship Attack,優(SA+50 / 艦攻‧優) | 4 | +50% | attack3 | 同上 |
| Combat & Espionage | Ground Combat,差(GC-10 / 地面戰‧差) | 退2 | -10 | ground1 | 描述:751–753 行;點數:`ground1_cost/-value` |
| Combat & Espionage | Ground Combat,佳(GC+10 / 地面戰‧佳) | 2 | +10 | ground2 | 同上 |
| Combat & Espionage | Ground Combat,優(GC+20 / 地面戰‧優) | 4 | +20 | ground3 | 同上 |
| Combat & Espionage | Spying,差(Spy-10 / 諜報‧差) | 退3 | -10 | spying1 | 描述:756–757 行;點數:`spying1_cost/-value` |
| Combat & Espionage | Spying,佳(Spy+10 / 諜報‧佳) | 3 | +10 | spying2 | 同上 |
| Combat & Espionage | Spying,優(Spy+20 / 諜報‧優) | 6 | +20 | spying3 | 同上 |
| Government | Feudal(封建) | 退4 | 見下方「政府型態效果」表 | 四種政府互斥(四選一) | 描述:767–799 行(p.20–21);點數:`race_pick.feudal` |
| Government | Dictatorship(獨裁) | 0 | 同上 | 互斥 | 描述:801–825 行(p.21–22);點數:`race_pick.dictatorship` |
| Government | Democracy(民主) | 7 | 同上 | 互斥 | 描述:827–855 行(p.21–22);點數:`race_pick.democracy` |
| Government | Unification(統一) | 6 | 同上 | 互斥 | 描述:858–889 行(p.22–23);點數:`race_pick.unification` |
| Special Abilities | Low-G World(低重力世界) | 退5 | 見描述 | 與 High-G 互斥(903 行) | 描述:898–903 行(p.23);點數:`race_pick.lowg_world` |
| Special Abilities | High-G World(高重力世界) | 6 | 見描述 | 與 Low-G 互斥(912 行) | 描述:906–912 行;點數:`race_pick.highg_world` |
| Special Abilities | Normal-G World(標準重力世界) | 0(不可購買,基準狀態) | — | 未選 Low-G/High-G 時的預設 | 手冊未把它列為可購買選項,推論自 Low-G/High-G 互斥描述 |
| Special Abilities | Aquatic(水生) | 5 | 環境相容性見描述 | — | 描述:915–918 行;點數:`race_pick.aquatic` |
| Special Abilities | Subterranean(穴居) | 6 | 最大人口 +2×星球等級;防守 GC+10 | — | 描述:921–926 行;點數:`race_pick.subterranean` |
| Special Abilities | Large Home World(大型母星) | 1 | 更高人口/食物/研究/產能起手 | — | 描述:932–934 行(p.24);點數:`race_pick.large_hw` |
| Special Abilities | Rich Home World(富礦母星) | 2 | 產能加速 | 與 Poor Home World 互斥(939 行) | 描述:937–939 行;點數:`race_pick.rich_hw` |
| Special Abilities | Poor Home World(貧礦母星) | 退1 | 產能變慢 | 與 Rich Home World 互斥(944 行) | 描述:942–944 行;點數:`race_pick.poor_hw` |
| Special Abilities | Artifacts World(遠古遺跡世界) | 3 | 科學家 5 研究/人(取代原 3) | — | 描述:947–949 行;點數:`race_pick.arti_world` |
| Special Abilities | Cybernetic(機械化) | 4 | 消耗食物+礦各半;戰後全修 | 與 Lithovore 互斥(958 行) | 描述:952–958 行;點數:`race_pick.cybernetic` |
| Special Abilities | Lithovore(食岩) | 10 | 免農業,吃礦維生 | 與 Cybernetic 互斥(965 行) | 描述:961–965 行;點數:`race_pick.lithovore` |
| Special Abilities | Repulsive(惹人厭) | 退6 | 外交受限;同化減半 | 與 Charismatic 互斥(973 行) | 描述:968–973 行(p.24–25);點數:`race_pick.repulsive` |
| Special Abilities | Charismatic(魅力非凡) | 3 | 外交加成;同化加快 | 與 Repulsive 互斥(986 行) | 描述:976–986 行;點數:`race_pick.charismatic` |
| Special Abilities | Uncreative(缺乏創造力) | 退4 | 每科技領域僅解鎖1種應用 | 與 Creative 互斥(992 行) | 描述:989–992 行;點數:`race_pick.uncreative` |
| Special Abilities | Creative(富創造力) | 8 | 每科技領域解鎖全部應用 | 與 Uncreative 互斥(998 行) | 描述:995–998 行;點數:`race_pick.creative` |
| Special Abilities | Tolerant(環境耐受) | 10 | 可用星球面積 +25% | — | 描述:1001–1007 行;點數:`race_pick.tolerant` |
| Special Abilities | Fantastic Traders(貿易奇才) | 4 | 貿易條約利潤+50%;餘糧換錢加倍;貿易品收入+50% | — | 描述:1010–1013 行(p.25);點數:`race_pick.fantastic_traders` |
| Special Abilities | Telepathic(心靈感應) | 6 | 外交+25;間諜+10;可心靈征服;瞬間同化;可用擄獲船艦 | — | 描述:1016–1026 行;點數:`race_pick.telepathic` |
| Special Abilities | Lucky(幸運) | 3 | 免疫災難;好事件加成 | — | 描述:1032–1034 行(p.26);點數:`race_pick.lucky` |
| Special Abilities | Omniscient(全知) | 3 | 開局全圖已知;敵艦動向全見 | — | 描述:1037–1041 行;點數:`race_pick.omniscient` |
| Special Abilities | Stealthy Ships(隱形艦隊) | 4 | 長程偵測隱形(不影響戰鬥) | — | 描述:1044–1046 行;點數:`race_pick.stealthy_ships` |
| Special Abilities | Trans-Dimensional(跨維度) | 5 | 航速+2/戰鬥速度+4;免超空間亂流 | — | 描述:1049–1053 行;點數:`race_pick.trans_dimensional` |
| Special Abilities | Warlord(戰爭領主) | 4 | 船員經驗+1級;軍營容量雙倍;每殖民地 Command Rating+2 | — | 描述:1056–1062 行(p.26–27);點數:`race_pick.warlord` |

## 附錄:政府型態的機制效果(非 picks 點數,是選了之後的規則效果)

以下數字來自手冊正文(敘述性文字,已核對頁碼/行號)與 `config.json` 的 `govt_bonus` 參數表交叉驗證
(`govt_bonus` 的值 × 5 = 百分比加成,兩份來源完全吻合,不是臆測):

| 政府 | 造艦成本 | 研究 | 稅收 | 防禦間諜加成 | 同化征服人口 | 首都被攻陷 |
|---|---|---|---|---|---|---|
| Feudal(封建) | 2/3 標準 | 1/2 標準 | — | 0 | 8 turns | 全境陷入無政府,Morale -50%,直到重建首都 |
| Confederation(封建進階) | 1/3 標準 | 3/4 標準 | — | 0 | 4 turns | 同 Feudal |
| Dictatorship(獨裁) | — | — | — | +10 | 8 turns | 全境 Morale -35%,直到重建首都 |
| Imperium(獨裁進階) | — | — | — | +15 | 4 turns | 同 Dictatorship;另外 Command Rating +50%,Morale 額外 +20% |
| Democracy(民主) | — | +50% | +50% | -10 | 4 turns | Morale -20%,直到重建首都;禁止清除征服人口 |
| Federation(民主進階) | — | +75% | +75% | -10 | 2 turns | 同 Democracy |
| Unification(統一) | — | — | — | +15 | 20 turns | 無首都概念,不受影響;食物/產能 +50%(取代 morale) |
| Galactic Unification(統一進階) | — | — | — | +15 | 15 turns | 同 Unification;食物/產能 +100% |

出處:Feudal/Confederation 767–799 行;Dictatorship/Imperium 801–825 行;Democracy/Federation
827–855 行;Unification/Galactic Unification 858–889 行(皆 `manual.txt`,約印刷頁 20–23)。

## 附錄:選用 mod「150 improved」(150i)的另一組點數(僅供參考,非預設值,勿直接採用)

`moo2_patch1.5/MOO2-1.50.26.zip → patch/150/mods/150i/cfg/PICKS_SP.CFG`(單機)與
`PICKS_MP.CFG`(多人)是**選用(opt-in)** mod 對點數的重新平衡,與本檔主表的預設值不同,
且 SP/MP 兩者彼此也不同。摘錄幾個差異較大的例子(格式:主表預設 → 150i-SP → 150i-MP):

- `maximum_negative_picks`:-10 → -12 → -10
- Money 差(money1_cost):-4 → -6 → -8
- Government Feudal:退4 → 退6 → 退10
- Government Unification:6 → 8 → 6
- Telepathic:6 → 8 → 5
- Fantastic Traders:4 → 2 → 1
- Uncreative:退4 → 退6 → 退10

完整清單見上述兩個 CFG 檔(已用 `unzip` 取出核對,非猜測)。**本專案若要重現「原版/1.5 預設」玩法,
應採用主表數字;只有明確要做 150i mod 支援時才用這組。**

## 手冊完全沒列具體數字的項目

- **Normal-G World**:手冊沒有把它列成一個「可購買」選項,只在 Low-G/High-G 的描述裡提到「Normal-G」
  作為比較基準。判斷它是未選 Low-G/High-G 時的預設狀態,cost=0,但這是推論,非手冊明文列出的選項列。
- **各級距的畫面顯示名稱**(例如遊戲畫面上 Farming 三級可能顯示成"Poor / Good / Great"或其他字眼):
  手冊正文與 `config.json` 都只用 `growth1/2/3`、`farming1/2/3`…這種欄位序號,**沒有列出畫面上實際顯示的
  文字標籤**。上表「差/佳/優」是我依數值高低給的**建議中譯分級名**,不是手冊或設定檔裡的原文字串。
  若要跟原版畫面像素/文字對齊,需要另外找遊戲截圖或字串資源檔核對实際 UI 標籤文字。

## 交給你裁決的不確定點

1. **1.3 vs 1.5 base 是否完全相同**:目前只能證明「1.5 base(未開 150i mod)」= 本檔主表數字;
   找不到反例證明 1.3 原版數字不同,但也沒有直接證據證明兩者完全一致。若你手邊有原版 1.3 執行檔或
   社群已核實的 1.3 wiki 點數表,建議再對一次。
2. **中譯分級名稱**(差/佳/優)是我建議的譯法,非官方字串,你可能想換成更貼近原版 UI 的字眼
   (例如「劣等/普通/優良」或直接沿用英文級距數字)。
3. **150i mod 的點數是否要支援**:附錄列出的 150i 數字只是抄錄存查,专案要不要做「選用平衡模式」
   需要你決定,目前 WORKLIST 沒有這項。
4. **Score Multiplier 公式**:手冊只給 3 個數據點(0/5/10 剩餘 picks → 100%/150%/200%),隱含
   「每剩 1 pick +10%」的線性關係,但手冊沒有明文寫出這條公式,也沒說負值 picks(倒扣後總花費超過
   10)時 Score 會怎麼算,需要你確認是否要以此線性假設實作,或者只在有剩餘 picks 時套用。
