# 銀河議會投票逆向稽核（2026-08-24）

## 結論

- **已證實**：`sub_156D4 @ IDA linear 0x156D4` 以各帝國人口的
  `ceil(population/10)` 選出基礎票最高兩名候選人；嚴格大於比較使同票時保留較低帝國索引。
- **已證實**：`Council_Votes_ @ 0x15EBC` 對每個非候選 AI 分別呼叫兩次
  `Vote_Check_ @ 0x16021`。只通過 A 時投 A、只通過 B 時投 B、兩邊都通過或都未通過時棄權。
- **已證實**：候選 AI 投自己；真人帝國不在這段自動計票，而由
  `sub_1633C @ 0x1633C` 顯示「候選 A／候選 B／棄權」三選一。
- **已證實**：`Vote_Check_` 最後呼叫 `Random_(200)`，回傳 1..200；擲骰小於等於分數才支持候選人。
  因此 remake 舊有的固定 `Relation >= 8` 與「比較兩邊關係後投較高者」不是原版規則。
- **已證實**：`sub_15DF8 @ 0x15DF8` 的當選門檻是
  `(2*totalVotes+2)/3`，即 `ceil(2*totalVotes/3)`；棄權票仍留在總票數分母。
- **已證實**：`Vote_Check_` 的完整 raw score 是目前關係、reputation、違約積怨、NPC proposal
  response memory、另一候選戰爭狀態、star record #0 控制者、兩項種族特性、Imperium、
  上屆選票及三種協議的有號整數和；高難度真人候選再以 `6/(difficulty+6)` 向零截斷。
  舊版列為未知的四欄與 `sub_78398` 已閉合，不再是 RE 留白。

## 證據身分

| 項目 | 值 |
|---|---|
| 輸入檔 | 私有 RE 工作區的 `Orion2.exe` |
| SHA-256 | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| IDA 資料庫 SHA-256 | `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e` |
| 外部符號表 SHA-256 | `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28` |
| 工具 | IDA Pro 9.4.0.260610，SDK 940 |
| 位址空間 | IDA linear；不可直接當檔案偏移 |
| 可重生匯出 | `tools/ida/audit_council_voting.py` |
| 版控證據 | [`evidence/council-voting-ida-20260828.json`](evidence/council-voting-ida-20260828.json) |

## `Vote_Check_` 分數資料流

參數依 caller 的暫存器資料流為：目標候選人、投票者、另一名候選人。

| 原始定位 | 運算 | 語意與證據等級 |
|---|---:|---|
| voter `+0x627[target]` | `>=4 → -1`、`==2 → 200` | **已證實控制流**；正式外交政策 raw 值已由條約切片映射 |
| voter `+0x617[target]` | signed byte | **已證實**：目前方向關係分數；writer／鏡射鏈見 `change-relations-callers-audit-20260828.md` |
| voter `+0x6D7[target]` | signed byte | **已證實**：方向 reputation；初始化為 0，由宣戰／解約與外交回應依方向及事件更新，亦直接參與 NPC 談判 |
| voter `+0x7EE[target]` | `2 * signed byte` | **已證實**：方向違約積怨；`Break_Treaties_` 對受害者與知情第三方寫入 |
| voter `+0x827[target*2]` | `-20 * int16` | **已證實**：NPC proposal response memory；`NPC_Proposal_Rejection_Accept_` 依接受／拒絕及 personality 更新並雙向鏡射 |
| voter `+0x627[other] >=4` | `+100` | **已證實控制流**；與另一候選人敵對時加分 |
| `sub_78398() == target` | `+40` | **已證實 raw**：回傳 star record #0 五顆行星中第一座有效殖民地的 owner；外部名稱 `Orion_Owner_` 支撐「控制 Orion」的強推論 |
| target `+0x8B3 ==1` | `+40` | **已證實**：Charismatic |
| target `+0x8B2 ==1` | `-100` | **已證實**：Repulsive |
| target `+0x1DF ==3` | `+30` | **已證實運算與編號**：政府 raw 3 = Imperium |
| voter `+0xE71 == target` | `+50` | **已證實**：上一屆議會投票目標；`sub_161E4` 寫入 |
| voter `+0x627[target] ==1` | `+50` | **已證實控制流** |
| voter `+0x62F[target] !=0` | `+25` | **已證實**：貿易協定 |
| voter `+0x637[target] !=0` | `+25` | **已證實**：研究協定 |
| target `+0x28 ==100` 且 difficulty>2 | `score*6/(difficulty+6)` | **已證實**：`+0x28==100` 是真人 marker；difficulty 為 raw 0..4，`idiv` 對有號值向零截斷 |

### `sub_78398` 的停止邊界

`sub_78398 @ 0x78398..0x783ED` 固定掃描 `_star` 基址第一筆 record 的五個 planet index
（`+0x4A..+0x52`）。每個非 `-1` index 取 17-byte `_planet` record；raw type `+0x04==3`
且 `planet+0x00` colony index 非 `-1` 時，回傳 361-byte `_colony` record 的 owner byte。
沒有合格殖民地回傳 `-1`。這條資料鏈足以證實「第一顆特殊星系的殖民 owner」是 `+40` 的
玩家可見輸入；`Orion_Owner_` 名稱只作 Orion 語意的強推論，不取代 raw 位址證據。

原版 1.31 的投票 score producer／兩次獨立檢定／棄權／回寫／勝負 consumer 至此閉合。
remake 目前只消費政策、協定、種族特性、政府與上一屆選票；方向 reputation、違約積怨、
proposal response memory、Orion owner 與高難度真人縮放尚未完整接入，所以仍是**明確不對齊**。
依 RE-first gate，本輪只更新證據，不改玩法程式或測試期望。

## 投票後關係寫回

`sub_161E4 @ 0x161E4` 已證實：投 A／B 會把該投票者的全部基礎票加到候選人累計，並記錄
`+0xE71`；棄權則寫 `0xFF`。原版同時調整候選人關係（支持者 `+24/-12`、棄權雙方 `-6`）。
remake 本輪先保存上一屆選票供下一屆分數使用；關係調整另受現有外交尺度與熱座資料模型影響，
未在沒有規格對照時直接移植 raw 數值。
