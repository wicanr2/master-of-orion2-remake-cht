# 外交條約評分與玩家需求回應垂直證據鏈（2026-08-28）

## 範圍、輸入與命名勘誤

本輪閉合兩條玩家可見外交鏈：AI 評估並產生條約提案，以及玩家提出九類要求後 AI 的
接受／拒絕與成功 consumer。標準文字載入、網路封包、畫面 helper、C runtime 與 Windows API
不進玩法分母。

- `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- `Orion2.exe.i64` SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4／IDAPython；位址為 IDA linear、DOS/4GW LE object #1。
- 非破壞性證據：[`diplomacy-response-ida-20260828.json`](evidence/diplomacy-response-ida-20260828.json)
  與 [`audit_diplomacy_response.py`](../../tools/ida/audit_diplomacy_response.py)。

外部名稱曾把 `sub_539D9` 稱為 `Get_Gift_Response_`；目前資料庫的直接 caller 證實它只由
`Diplomacy_Demand_ @ 0x18B79` 呼叫九次，正確邊界是「需求回應」，不是餽贈接受器。
`sub_53146` 則只有 `Check_Treaty_Proposal_ @ 0x5272D` 一個 caller、五個 callsite；不能再把
兩者合併成一個泛稱的 `DiplomacyResponse` 公式。

## AI 條約提案評分 `sub_53146`（已證實）

五個 callsite 的 proposal kind 與基準門檻如下；kind 依相鄰 `Start_Treaty_`／trade／research
consumer 交叉驗證：

| kind | 玩家可見提案 | 傳入基準 |
|---:|---|---|
| 1 | 和平 | 75 |
| 2 | 互不侵犯 | `signed(source→target +0x6D7)+125` |
| 3 | 同盟 | 由雙方艦隊、人口、科技、關係與持續時間組合 |
| 7 | 貿易 | 60 |
| 8 | 研究 | 40 |

每個 caller 都在 `call sub_53146` 前執行 `xor ecx,ecx`。依原始 Watcom register calling
convention，隱藏 modifier 類別固定為 0；反編譯器顯示的未初始化 `v9/v10` 不可當未知參數。

令 source 為評分 AI、target 為提案對象。原始分數依序為：

```text
score = signed(source→target +0x6D7)
      + signed(source→target +0x617)
      + signed(source→target +0x7EE)
if source government == 4:
    score += signed(source→target +0x7EE)

threshold = proposalBase - governmentBias[source government]
if source government == 4 and source→target +0x727 == 1:
    threshold = proposalBase - specialGovernmentBias

score += Random(100) - threshold
score += int16(source→target +0x68F)
```

接著 `Adjust_Diplomat_Modifiers_ @ 0x524C3` 令方向 `+0x68F/+0x69F/+0x6AF/+0x6BF`
各減 10；類別 0 再使 `+0x68F -= Random(30)+20`。這是持久化副作用，不只是局部評分。

最後加入 target 的特性與領袖：

- target `+0x8B3` 非零：`+50`；否則 target `+0x8B2` 非零：`-50`。
- target `+0x8B8` 非零：`+25`。
- target `+0x1DF == 3`：`+30`。
- `Best_Leader_Diplomat_Bonus_ @ 0xE5E09`：掃 67 筆 59-byte 領袖記錄，只取仍可用且
  owner 相符者；進階／普通 Diplomat 分別為 `15*(level+1)`／`10*(level+1)`，取最大值。

最終四桶為 `<-75→0`、`[-75,-50)→2`、`[-50,0)→1`、`>=0→3`。四桶不是單純
接受／拒絕；`Check_Treaty_Proposal_` 會再依 kind、四桶、1/2/3/4 隨機支線與現有條約狀態
選訊息、提出附帶 BC／科技／殖民地交換，或不提出。完整 caller bytes 已包含於證據 JSON。

## 玩家九類需求 `sub_539D9`（已證實 raw 契約）

函式參數為 requester、recipient、mode 0..8、payload、success word pointer、message word
pointer。基礎分數先組合：

- `Fleet_Comparison_ @ 0x51DCE`：方向 `+0x5EC` 比率，夾 `[-300,300]`。
- `Population_Comparison_ @ 0x51E98`：人口差百分比後再以其絕對值作隨機上界。
- `Tech_Comparison_ @ 0x51E3B`：雙方科技總量百分比，夾 `[-300,300]`。
- `Find_Worst_Modifier_ @ 0x50FDF`：四個方向 modifier 的最小值；議會狀態 3 時為 0。
- `Random(200)`、recipient 政府表偏差的兩倍，以及模式專屬成本。

模式專屬成本與接受後 consumer：

| mode | payload／評分成本 | `Diplomacy_Give_In_To_Demand_ @ 0x1B487` consumer |
|---:|---|---|
| 0 | 第三帝國；依其正式狀態、關係、艦隊比與真人 gate 扣分 | recipient 對 payload 的 `+0x897` 寫 0 |
| 1 | 第三帝國；扣其政府偏差與 80，真人 payload 強制 score 0 | `+0x897` 寫 1 |
| 2 | 第三帝國；扣 `2*signed(recipient→payload +0x617)` | `+0x897` 寫 2 |
| 3 | 無 payload；扣 50 | 建立 recipient→requester tier 1 tribute |
| 4 | 無 payload；扣 150 | 建立 recipient→requester tier 2 tribute |
| 5 | 無 payload；扣 50 | 無資源移轉；成功時寫方向 `+0x806=Random(50)+Random(50)+10` |
| 6 | technology id；扣 `100*TechValue/recipientTechTotal` | requester 取得該科技 |
| 7 | star id；扣 `min(1000*starValue/recipientPopulation,300)` | 通過 `May_Surrender_Star_` 後移交星系 |
| 8 | 無 payload；扣 50 | 寫方向 `+0x80E=Random(50)+Random(50)+10` |

mode 0..2 的玩家文案名稱仍需由 JIMTEXT 資產或正常原版畫面確認；raw payload、評分與狀態
writer 已閉合，不以猜測標籤取代 `+0x897`。mode 5／8 同理保留 raw cooldown 欄位，不冒稱其
正式 UI 名稱。這些名稱留白不妨礙重建已證實的行為，但 UI spec 不得自行命名為原版術語。

不論結果，評分後都呼叫 `Adjust_Diplomat_Modifiers_` 五次，所以四個方向 modifier 各減 50。
接著依目前正式狀態分兩套門檻：

| 正式狀態 | 分數 | 訊息／副作用 |
|---|---:|---|
| 1／2 | `<-100` 或 recipient personality helper == 6 | 165；宣戰；失敗 |
| 1／2 | `[-100,-50)` | 166；三類條約全部解除；失敗 |
| 1／2 | `[-50,-25)` | 167；關係再減 10、鏡射；失敗 |
| 1／2 | `[-25,0)` | 168；失敗 |
| 1／2 | `>=0` | 169；成功 |
| 其他 | `<-150` 或 personality helper == 6 | 165；宣戰；失敗 |
| 其他 | `[-150,-75)` | 167；關係再減 10、鏡射；失敗 |
| 其他 | `[-75,0]` | 168；失敗 |
| 其他 | `>0` | 169；成功 |

`Get_Player_Diplomacy_Personality_ @ 0x53E96` 在 government 4 且方向 `+0x727==1` 時回 6，
否則回 recipient government raw。成功 pointer 只有 1 才由 caller 呼叫 `sub_1B487`；因此
需求失敗不得提前轉移 BC、科技、星系或建立納貢。

## RE 結論與 remake 差異

條約提案的輸入欄位、政府／特性／領袖修正、亂數、modifier 消耗、四桶與 proposal caller；
玩家九類需求的模式、payload、成本、雙門檻、訊息、宣戰／解約／關係副作用及成功 consumer
均已形成垂直證據鏈。RE 列可關閉。

remake 現行 `GameSession.DiplomacyResponse` 直接依 action 建立條約並給固定關係增益，沒有上述
原版評分、拒絕、附帶交換與 modifier 持久狀態，屬**明確不對齊**。依 RE-first gate，本輪只
登記差異，不撰寫 spec 或修改 Go。
