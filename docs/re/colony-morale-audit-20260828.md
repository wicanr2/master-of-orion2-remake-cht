# 殖民地士氣 producer 與產出消費鏈稽核（2026-08-28）

## 問題與證據契約

既有 remake 主要依手冊建立士氣常數；`capitol-state-audit-20260826.md` 只閉合失都分支，
尚未證明原版完整 producer、raw 刻度、混合種族判定、領袖公式與報表 consumer。本輪只追
玩家可見玩法；`sprintf_`、字串複製及文字框 helper 依 runtime／平台停止線排除。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4、`tools/ida/audit_colony_turn_chain.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱不作證據。
- 可重生證據：
  [`evidence/colony-turn-chain-ida-20260828.json`](evidence/colony-turn-chain-ida-20260828.json)。

## raw 刻度與 writer／consumer

`sub_DDB25 @ 0xDDB25..0xDDEFB` 以 colony record 為輸入，將士氣 raw 值寫入
`colony+0x07`。所有分項先以 **5 個百分點為一單位**相加；報表段
`0xDDD7F..0xDDEF6` 顯示時逐項乘 5。`Pre_Import_Computing_ @ 0xE1DB0` 以空報表指標
呼叫 producer，`sub_DDEFB @ 0xDDEFB..0xDDF2C` 則以 2000-byte 暫存文字呼叫同一 producer，
再交給文字框 consumer。

`Colony_Job_Production_ @ 0xDE280` 將 `colony+0x07 × P` 直接加入以 `20P` 為基準的
逐人口 accumulator，所以一個 raw 單位恰為 `P/20 = 5%`。統一系政府與特定 Android
職務會跳過這個一般士氣項；這條 producer→cache→食物／工業／研究消費鏈為**已證實**。

## 政府、軍營、失都與科技表

原版政府 raw 值為 0..7：封建、邦聯、獨裁、帝國、民主、聯邦、統一、銀河統一。

| 原始資料／條件 | raw 單位 | 玩家百分點 |
| --- | ---: | ---: |
| `byte_DD4C4 @ 0xDD4C4`, bytes `FC FC FC 00 00 00 00 00` | `[-4,-4,-4,0,0,0,0,0]` | `[-20,-20,-20,0,0,0,0,0]` |
| Marine Barracks `colony+0x14C` 或 Armor Barracks `+0x138`; `byte_DD4CF` | 前四政府 `+4` | `+20` |
| 失都且指定 Capitol 尚未重建；`byte_DD4CC`, bytes `F6 F9 FC` | 政府族 `[-10,-7,-4]` | `[-50,-35,-20]` |
| Holo Simulator raw 20，`colony+0x14A` | `+4` | `+20` |
| Pleasure Dome raw 31，`colony+0x155` | `+6` | `+30` |
| Virtual Reality Network，`player+0x1DA == 3` | `+4` | `+20` |
| Psionics，`player+0x1AA == 3`，且政府族為封建／獨裁 | `+2` | `+10` |

科技狀態基址是 `player+0x117`；`0x1DA-0x117=195`、`0x1AA-0x117=147`，與受版控
`TECH_VIRTUAL_REALITY_NETWORK=195`、`TECH_PSIONICS=147` 交叉閉合。帝國政府的 `+20%`
不是另一個隱藏常數：raw 3 的基礎表為 0，而軍營表仍加 4，因此無軍營為 0、有軍營為 +20%。
統一兩級在 `rawGovernment/2 == 3` 時於 `0xDDBA9` 跳過其餘一般士氣來源。

## 混合種族：`sub_DDAD4 @ 0xDDAD4..0xDDB25`

函式只在總人口至少 2 且 Alien Management Center raw 1（`colony+0x137`）不存在時掃描
packed population。它取每格低 nibble race slot，忽略 Android／Natives 等 slot `>=8`；
只要找到兩個不同的正常 race slot 就回傳 `-4`，即 `-20%`，否則回 0。

因此原版判定是「殖民地是否同時存在至少兩個正常種族」，不是「是否仍有未同化人口」。
同化狀態、prisoner flag 與 owner race 都不是此 helper 的比較條件。這推翻 remake 目前以
`UnassimilatedPop > 0` 代替 multi-racial 的模型；該差異只登記，待 RE gate 關閉後進 READY spec。

## 心靈導師領袖

`sub_DD9F2 @ 0xDD9F2` 由 colony→planet→star 取得指派的行政領袖；若領袖 raw `+0x37`
非零則視為不可用。`sub_DDB25` 讀領袖技能 byte `+0x2B`：

- 一般 Spiritual Leader bit `0x40`：raw bonus `L+1`；
- 進階 Spiritual Leader bit `0x80`：raw bonus `(3L+4)/2`；
- `L = sub_94951(leader)` 的有效經驗等級，該 helper 將結果上限夾到 5。

乘回 5 個百分點後，一般階為 `5×(L+1)%`；進階階為
`5×floor((3L+4)/2)%`。後者等價於一般值乘 1.5 後採原版整數取整。若兩個 bit 同時存在，
程式先檢查 `0x80`，只採進階公式。這補上舊文件所稱「手冊沒有精確 Spiritual Leader
公式」的留白。remake 已使用技能 tier／經驗公式，但其顯示等級 adapter 目前只夾到
expLevel 4；原版 helper 可回 5，須在 READY spec 階段核對真實英雄資料是否可到此邊界。

## 閉合結論與 remake 邊界

- **已證實**：raw 5% 刻度、政府八格表、兩種軍營、三類失都表、兩棟建築、兩項科技、
  統一系 bypass、混合種族精確判定、心靈導師兩階公式、pre-import writer、報表 consumer，
  以及 `DE280` 的食物／工業／研究消費端。
- **已知 remake 偏差**：多種族目前以 `UnassimilatedPop > 0` 近似；領袖最高經驗邊界須重驗。
- **尚未在本切片宣稱**：BC／income 是否另讀士氣、叛亂如何消費士氣，以及 AI 對人口 race
  group 的完整保存。它們分別屬 BC 產出、叛亂及 AI 人口模型切片，不以手冊敘述冒充閉合。
- `sprintf_`、文字複製與文字框內部是 C runtime／UI 平台服務；只保留玩家可見報表 callsite，
  不納入 RE 知識庫完成分母或 remake 範圍。
