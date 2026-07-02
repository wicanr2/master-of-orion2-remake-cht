# 遊戲字串源總清單與翻譯進度

窮舉 openorion2 `TextManager::load` 載入的所有字串源(mom playbook [HARD]:漏一源 = 整類畫面英文)。目標:完成所有訊息翻譯。譯文放 `assets/i18n/*.tsv`(英文原文即 key)。

## 進度總表

| 源 (.lbx) | 格式 | 資產 | 條數 | 內容 | TSV | 狀態 |
|---|---|---|---|---|---|---|
| TECHNAME | loadStrings | 0 | 419(唯一) | 科技/元件名 | tech.tsv | ✅ 完成 |
| SKILDESC | loadAsset | 0 | 27 | 軍官技能名 | skills.tsv | ✅ 完成 |
| RACESTUF | loadStrings | 0 | 64 | 種族特性/自訂 pick | races.tsv | ✅ 完成 |
| STARNAME | loadAsset | 0 | 13 | 母星名 | — | ⏳ 專有名詞,保留原文 |
| STARNAME | loadAsset | 1 | 829 | 星名池(隨機) | — | ⏳ 專有名詞池,低優先 |
| SHIPNAME | loadAsset | 0 | 672 | 艦名池(隨機) | — | ⏳ 專有名詞池,低優先 |
| RACESTUF | loadStrings | 8 | ~ | 種族背景資訊 | — | ⏳ 待(敘述) |
| SKILDESC | loadAsset | 1 | 27 | 技能描述 | skilldesc.tsv | ✅ 完成 |
| TECHDESC | loadAsset | 0..7 | ~849 | 科技描述 | — | ⏳ 待(大宗敘述) |
| MAINTEXT | loadFile | 群組 | 主 UI 文字 | — | ⏳ 待(UI 大宗) |
| EVENTMSE | loadFile | 群組 | ~152 | 隨機事件訊息 | — | ⏳ 待(敘述) |
| DIPLOMSE | loadFile | 群組 | ~180 | 外交對白 | — | ⏳ 待(敘述) |
| COUNCMSG | loadFile | 群組 | ~84 | 議會訊息 | — | ⏳ 待 |
| ANTARMSG | loadFile | 群組 | ~48 | 安塔蘭訊息 | — | ⏳ 待 |
| RSTRING0 | loadStrings | 0 (off 4) | ~178 | 種族相關字串 | — | ⏳ 待 |
| ESTRINGS | loadStrings | 0 (off 6) | ~813 | 事件/介面字串 | — | ⏳ 待(大宗) |
| HESTRNGS | loadStrings | 0 (off 6) | ~398 | help 短字串 | — | ⏳ 待 |
| HELP | 特殊 | 多 | 大量 | 百科全文 | — | ⏳ 待(最大宗敘述) |
| CREDITS | loadAsset | 0 | ~ | 製作名單 | — | ⏳ 待 |
| HERODATA | — | — | ~ | 軍官資料 | — | ⏳ 待 |
| billtext/jimtext/kentext… | loadFile | 群組 | misc 敘述 | — | ⏳ 待 |

## 翻譯策略

- **名稱/短語組**(科技/技能/特性/UI 標籤):優先,已多數完成。英文原文即 key,顯示層覆蓋。
- **敘述組**(描述/事件/外交/help):逐源 dump → 逐條翻,量大,分批推進。
- **專有名詞池**(星名/艦名/母星,隨機產生):MOO2 星名(Orion/Altair…)多為經典,傾向保留原文或「中文(原文)」;暫列低優先。
- 每組完成後 TSV 守護測試(載入 + 佔位符一致)把關。

> 「完成所有訊息翻譯」= 上表所有敘述/UI 源翻完;專有名詞池另定策略。
