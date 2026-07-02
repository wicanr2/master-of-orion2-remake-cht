# i18n Catalog 架構與同形詞稽核

## 顯示層覆蓋(display-layer override)

譯文以**英文原文為 key**存在 `assets/i18n/*.tsv`,渲染時 `Catalog.Translate(englishText)` 查表換中文。
好處:不改動遊戲資料檔、不依賴字串在 LBX 中的 index,任何畫面只要拿到英文原文即可翻譯。

## 風險:跨來源同形詞

MOO2 引擎的字串查詢是 **`(表, id)` 位置式**(`misctext(file,id)`、`techname(id)`、`racename(id)`、`_estrings[id]`…),
所以同一個英文字在不同表可對應不同語意——引擎不會混淆,因為它從不用「英文文字」當 key。

我們的顯示層覆蓋改用「英文文字」當 key,若把**所有 TSV 合併成單一 catalog**,同形詞會碰撞(後載入者靜默覆蓋)。

## 架構決策:per-source catalog

**忠於 MOO2 的 per-table 設計:每個字串來源(TSV)載入為獨立 catalog,渲染某來源的字串時用該來源的 catalog 查。**
如此同形詞各表獨立、各自正確,不需強制全域統一。`cmd/moo2` 目前各 demo 模式即各載一個 TSV,符合此模型;
未來完整遊戲迴圈應維護 `map[source]*Catalog`,而非單一大表。

## 同形詞稽核(2026-07-03)

跨 TSV 同英文 key 但譯文不同者共 32 組。處理原則:

- **直接可比較的畫面 → 強制統一**:`races.tsv`(種族選擇)與 `raceinfo.tsv`(種族資訊面板)是同一批 RACESTUF 特性,
  玩家在兩畫面直接對照,已統一 11 組特性用字(地底棲/富礦母星/貧礦母星/半機械化/食岩/令人厭惡/魅力非凡/富創造力/適應力強/心靈感應/軍閥)。
- **跨不同來源的同形詞 → 保留各自譯文**(per-source 查詢下正確,非 bug)。已逐一 review,清單如下。

### 已 review、接受的 per-source 同形詞(21 組)

| 英文 key | 各來源譯文 |
|---|---|
| `Armor` | misc=裝甲(艦艇裝備分組) / tech=裝甲車(**地面戰鬥單位**,與 Marines/Spy 同列,正確) |
| `Colony` | hestrings=殖民地 / misc=殖民(科技分組) |
| `Confederation` | estrings=邦聯制 / tech=邦聯 |
| `Democracy` | estrings=民主制 / races=民主 |
| `Description` | estrings=描述 / hestrings=說明 |
| `Dictatorship` | estrings=獨裁制 / races=獨裁 |
| `Farming Leader` | estrings=農業專家 / officer=農業領袖 |
| `Federation` | estrings=聯邦制 / tech=聯邦 |
| `Feudal` | estrings=封建制 / races=封建 |
| `Financial Leader` | estrings=財政專家 / officer=財政領袖 |
| `Imperium` | estrings=帝制 / tech=帝國 |
| `Jump Gate` | hestrings=跳躍之門 / tech=躍遷門 |
| `Labor Leader` | estrings=勞工專家 / officer=勞工領袖 |
| `No Special` | estrings/maintext/tech=無特殊 / hestrings=無特殊系統 |
| `Outpost Ship` | estrings=前哨站船 / tech=前哨船 |
| `Ruthless` | estrings=冷酷無情 / officer=無情者(軍官頭銜) |
| `Science Leader` | estrings=科學專家 / officer=科學領袖 |
| `Spy Master` | estrings=間諜頭子 / officer=諜報大師(軍官頭銜) |
| `Trade Goods` | estrings=貿易商品 / rstring=貿易品 |
| `Unification` | estrings=統一制 / races=統一 |
| TEN 安裝說明長句 | help / hestrings 兩版措辭略異(同義) |

> 政體類(Feudal/Democracy/…)estrings 帶「制」尾、races/tech 為裸詞:前者為「政體狀態顯示」、
> 後者為「可選政體/科技名」,語境不同,保留。若日後改單一 catalog,須改為 per-source 或先統一本表。
