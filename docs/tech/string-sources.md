# 遊戲字串源總清單與翻譯進度

窮舉 openorion2 `TextManager::load` 載入的所有字串源(mom playbook [HARD]:漏一源 = 整類畫面英文)。目標:完成所有訊息翻譯。譯文放 `assets/i18n/*.tsv`(英文原文即 key)。

**完整性狀態(2026-07-02):`TextManager::load` 全部 22 個來源已逐一 account。**
所有敘述/UI 類字串源均已翻譯完成;**專有名詞池**共 4 個(STARNAME 母星/隨機星名、SHIPNAME 艦名、RACENAME AI 統治者名),依策略逐一定案——母星名(13)、艦名(672)、隨機星名(829)均已翻譯落地,RACENAME(104)依既有先例定案保留原文。
即「完成所有訊息翻譯」的顯示文字部分已達成,四個專有名詞池全數定案(僅 RACENAME 定案為保留原文,非未完成)。

## 進度總表

| 源 (.lbx) | 格式 | 資產 | 條數 | 內容 | TSV | 狀態 |
|---|---|---|---|---|---|---|
| TECHNAME | loadStrings | 0 | 419(唯一) | 科技/元件名 | tech.tsv | ✅ 完成 |
| SKILDESC | loadAsset | 0 | 27 | 軍官技能名 | skills.tsv | ✅ 完成 |
| RACESTUF | loadStrings | 0 | 64 | 種族特性/自訂 pick | races.tsv | ✅ 完成 |
| STARNAME | loadAsset | 0 | 13 | 母星名 | starname.tsv | ✅ 完成(實星名用中文天文名,虛構音譯) |
| STARNAME | loadAsset | 1 | 829 | 星名池(隨機) | starname-random.tsv | ✅ 完成(2026-07-11,829 條英文名彼此互不重複;真名/圍棋術語/神話專有名詞優先意譯,虛構短音節規則化音譯,見 proper-noun-strategy.md) |
| SHIPNAME | loadAsset | 0 | 672 | 艦名池(隨機) | shipname.tsv | ✅ 完成(2026-07-11,190 組基底詞意譯+羅馬數字流水號保留,見 proper-noun-strategy.md) |
| RACENAME | loadAsset | 0 | 104 | AI 統治者姓名 | — | ⏳ 專有名詞池,保留原文 |
| RACESTUF | loadStrings | 8 | 32 | 種族資訊面板 | raceinfo.tsv | ✅ 完成 |
| SKILDESC | loadAsset | 1 | 27 | 技能描述 | skilldesc.tsv | ✅ 完成 |
| TECHDESC | loadAsset | 0..3 | 151(唯一) | 元件名+短描述 | techdesc.tsv | ✅ 完成(名稱由 tech.tsv 覆蓋,80 描述另翻;實測全庫 0 缺口) |
| MAINTEXT | loadFile | 群組 | 14 | 探索特殊事件 | maintext.tsv | ✅ 完成 |
| EVENTMSE | loadFile | 群組 | 99 | GNN 事件新聞 | event.tsv | ✅ 完成 |
| DIPLOMSE | 特殊 | 180 | 770 | 外交對白 | diplo.tsv | ✅ 完成 |
| COUNCMSG | loadFile | 群組 | 14 | 議會勝選訊息 | council.tsv | ✅ 完成 |
| ANTARMSG | loadFile | 群組 | 8 | 安塔蘭威嚇 | antaran.tsv | ✅ 完成 |
| RSTRING0 | loadStrings | 0 (off 4) | 177 | 殖民摘要+回合報告 | rstring.tsv | ✅ 完成 |
| ESTRINGS | loadStrings | 0 (off 6) | 811 | 介面/事件(除錯碎片 fallback 英文) | estrings.tsv | ✅ 玩家訊息完成(576) |
| HESTRNGS | loadStrings | 0 (off 6) | 395 | help/提示訊息 | hestrings.tsv | ✅ 完成 |
| HELP | 特殊 | 0 | 707→704 | 百科全文 | help.tsv | ✅ 完成 |
| CREDITS | loadAsset | 0 | 57→13 | 製作名單 | credits.tsv | ✅ 職務標籤完成(人名保留) |
| HERODATA | loadAsset | 0 | 65 | 軍官職稱 | officer.tsv | ✅ 完成(姓名為專有名詞保留) |
| BILLTEXT/BILLTEX2 | loadFile(step6) | 0 | 62 | 資訊/研究/科技分組 UI | misc.tsv | ✅ 引擎引用索引完成 |
| KENTEXT/JIMTEXT2 | loadFile(step6) | 0 | (含上) | 艦員等級/冠詞 | misc.tsv | ✅ 引擎引用索引完成 |
| JIMTEXT/KENTEXT1 等其餘 misc | loadFile(step6) | 0 | 未接線 | 未實作畫面文字 | — | ⏳ 待對應畫面移植(openorion2 未引用) |
| BILLTEXT str_id 10-13,26-60 | loadFile(step6) | 0 | 38 | 參考資料/回合摘要/**17 級外交關係形容詞**/間諜/條約 | misc.tsv | ✅ 完成(預先翻好待畫面接線) |

## 翻譯策略

- **名稱/短語組**(科技/技能/特性/UI 標籤):優先,已多數完成。英文原文即 key,顯示層覆蓋。
- **敘述組**(描述/事件/外交/help):逐源 dump → 逐條翻,量大,分批推進。
- **專有名詞池**(星名/艦名/母星,隨機產生):MOO2 星名(Orion/Altair…)多為經典,傾向保留原文或「中文(原文)」;暫列低優先。
- 每組完成後 TSV 守護測試(載入 + 佔位符一致)把關。

> 「完成所有訊息翻譯」= 上表所有敘述/UI 源翻完;專有名詞池另定策略。

## misc 檔的語言交錯(重要技術細節)

`billtext/jimtext/kentext` 系列用 `loadFile(archive, offset=lang_id, step=LANG_GROUPS=6, group_id=0, groups=1)` 載入:
**同一邏輯字串的 6 種語言(En/De/Fr/Es/It/En)存在 6 個連續 asset**,英文(lang_id=0)取 `asset = str_id × 6`。
逆向 dump 時若直接用 raw asset index 會混到德/法文 → 錯 key。務必用 `str_id × 6` 還原英文。
(對比:`antarmsg/councmsg` 用 `offset=0, step=1, group_id=lang_id, groups=6`,是「每語言一整塊」的另一種交錯,English 取前 1/6 連續段。)

openorion2 僅接線少數 misc str_id(資訊畫面標題、收入分類、研究領域、科技分組、艦員等級);
其餘為未實作畫面的文字,列「待對應畫面移植」,不投機翻譯(靜態溯源:不翻引擎不顯示的死字串)。
