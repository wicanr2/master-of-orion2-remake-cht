# 殖民地叛亂完整資料流稽核（2026-08-28）

## 問題與證據契約

舊知識只摘出每人口 1% 與兩個倍率，尚未閉合「哪些殖民地／人口會受檢」、舊主抽選、兵力、
地面戰、人口忠誠、易手、訊息與滅國時序；remake 因而自行補了只檢查玩家殖民地、鎮壓會
消滅叛亂人口等行為。本輪只完成 RE 與登記差異，不建立 READY spec，也不修改玩法。

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`；IDA `.i64`
  SHA-256 `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4、`tools/ida/audit_rebellions.py`；位址均為 IDA linear，
  DOS/4GW LE object #1。
- 外部導航符號：`symbols_fixed.tsv`，SHA-256
  `f83116049964c14072547dd087a6316117cf62808c328cc7c087984467030b28`；名稱只供導航。
- 可重生證據：[`evidence/rebellions-ida-20260828.json`](evidence/rebellions-ida-20260828.json)。
- `memset_` 等 C runtime 只作暫存初始化，不納入 RE／remake 分母。

## 回合入口與候選殖民地

`Next_Turn_Calc_ @ 0x136B3` 在 `0x13792` 呼叫
`Check_All_Rebellions_ @ 0xED44A..0xED48B`。後者正向掃描全部殖民地記錄；只跳過
`owner == -1` 的空槽與 `colony+0x06 != 0` 的非活動記錄，再逐筆呼叫
`Check_Rebellion_ @ 0xED260..0xED44A`。因此「只檢查真人玩家殖民地」是錯誤舊斷言：
AI 佔領並保存異族 prisoner 的殖民地也在原版檢查範圍內。這一點為**已證實**。

## prisoner 計數、舊主與機率

`Check_Rebellion_` 先以 `colony+0x12F == 4` 排除未被征服的殖民地，再由 packed population
尾端向前掃描。每格以第二個 byte 的 bit `0x04` 判斷 prisoner，以低 nibble 取 race slot；
只有對應 player record `+0x24 == 0`、仍存續的舊主才列入。函式同時建立各舊主人數表及合計：

1. `Get_Weighted_Choice_Char_ @ 0xFE8DA` 依各舊主 prisoner 人數加權抽出本次還政對象；
2. 機率基數使用**所有仍有存續舊主的 prisoner 合計**，不是抽中舊主自己的數量；
3. `chance = 10 × totalEligiblePrisoners`；目前 owner 為真人（`player+0x28 == 100`）且
   抽中舊主不是真人時，再加 `4×difficulty-8`；
4. Alien Management Center（raw 1，`colony+0x137 != 0`）將正值向零除以 2；
   `colony+0x12F == 0` 再乘 2；
5. `Random_(1000)` 回傳 1..1000，`roll <= chance` 才觸發。

原版此處**不讀 `colony+0x07` 士氣**；士氣不是叛亂機率 consumer。政府只會透過先前同化速度
間接改變剩餘 prisoner，並非在本函式再套一張政府表。以上資料流均為**已證實**；
`colony+0x12F` 的 1／2／3 政策語意仍不在本切片宣稱。

觸發後第二次 `Random_(totalEligiblePrisoners)` 產生 1..合計的叛軍數。值得注意的是，上限不是
抽中舊主自己的 prisoner 數，故原程式可讓叛軍數高於該 race 的 prisoner 數；這是 raw 行為，
不是文件筆誤。

## 地面戰與結果訊息

`Get_Rebellion_Info_ @ 0xEC65A..0xECECA` 建立兩份 55-byte 地面戰記錄：

- 攻方 owner 是抽中的舊主，type 3（record `+0x10`）數量為叛軍骰；
- 守方 type 0／1／2（`+0x0A/+0x0C/+0x0E`）分別取殖民地裝甲、陸戰隊與
  `Colony_N_Militia_`；
- 兩方各自經 `Compute_Player_Ground_Combat_Bonuses_ @ 0xEC15C` 與
  `Compute_Ground_Combat_Info_ @ 0xEC3CE` 建立命中／耐久資料；守方另含殖民地軍官 bonus；
- `Check_Rebellion_` 直接反覆呼叫 `Ground_Combat_Round_ @ 0xEC4FE`，直到任一側目前可戰
  type 索引達終止值 4。

函式建立 raw 訊息 record：type byte 3 為 9，並保存 planet、舊主、初始叛軍數與
`初始叛軍數－剩餘叛軍數`；同一 record 分別送給現任 owner 與舊主。這證明訊息存在及其欄位，
但本切片不把外部導航名稱當成玩家文案精確內容。

## 鎮壓與叛亂成功的回寫

- **守方勝利**：原函式不呼叫人口歸屬、殖民地易手或 `Resolve_Rebellion_Troops_`，也沒有刪除
  packed population。故 remake 目前「鎮壓後消滅全部起事人口」與原版矛盾。
- **叛軍勝利**：以空 conquest-report 指標呼叫
  `Change_Colony_Ownership_ @ 0xECF41..0xED260`，所以不走一般征服的科技擄獲段；接著把攻方
  56-byte 記錄傳給 `Resolve_Rebellion_Troops_ @ 0xECECA..0xECF41`，將剩餘叛軍轉成新殖民地
  陸戰隊 `colony+0x130 = min(remainingRebels, 4, Colony_Infantry_Limit_)`，並寫
  `colony+0x123 = -41`。

`Change_Colony_Ownership_` 先呼叫 `Change_Pop_Ownership_ @ 0xECBF7..0xECE05`，清除地面裝甲／
陸戰隊及三個 ownership-sensitive 建築：Armor Barracks raw 2、Marine Barracks raw 22、
Alien Management Center raw 1。`Change_Pop_Ownership_` 另處理失去殖民地的領袖、舊主人口統計、
首都重指派、星系／殖民地 owner、環境／人口上限、建造佇列與衍生 cache。

packed population 的核心寫回亦已由原始運算元閉合：低 nibble 保留 race，bits 4..6 寫新 owner，
bit `0x0400` 為 prisoner。race 等於新 owner、特殊 slot 8／9，或人口原 owner 已是新 owner者會
清 prisoner；其餘異族設 prisoner。`player+0x8B8` 非零時會繞過後一分支。2026-08-28 的完整
31-byte mapping 與直接消費端普查已把此欄交叉確認為 Telepathic；此處仍保留 raw offset，並把
「繞過該分支」與玩家可見的心靈控制／俘虜處理語意分開，不由 trait 名稱反推更多行為。

以上回寫為**已證實**；建築英文名由受版控 raw ID 對照表交叉驗證。

## 滅國時序

叛亂函式與兩個 ownership helper 都不直接呼叫滅國檢查。唯一 caller 證據顯示
`Check_For_Eliminated_Players_ @ 0xE4EB3..0xE4F49` 是由
`Search_For_Battles_ @ 0xE9D62` 的 `0xEA269` 呼叫；它掃描活動殖民地，對已無殖民地的帝國呼叫
`Eliminate_Player_` 並建立 race-eliminated event。回合主鏈中 `Search_For_Battles_` 位於叛亂
之前，因此本輪叛亂造成的最後殖民地易手不會在叛亂函式內立即滅國；何時由下一次戰鬥搜尋或
其他路徑觀察到，須由滅國排程專題決定，不能寫成「叛亂直接完成滅國」。

## 閉合結論與 remake 邊界

- **已證實**：候選殖民地、packed prisoner／存續舊主 gate、舊主加權抽選、兩次 RNG、精確機率
  順序、AMC／政策分支、地面戰兩側兵力、訊息欄位、勝敗分支、人口 ownership、建築／部隊／
  cache 回寫與獨立滅國 caller。
- **已知 remake 偏差**：只檢查真人殖民地；只有單一 `ConqueredFrom`；鎮壓會刪人口；舊主
  不存在時讓殖民地脫離；叛亂勝利的人口、建築、佇列與最多四名陸戰隊回寫不完整；AI 殖民地
  prisoner 模型不足。
- **仍未知但不阻止本列 RE 閉合**：政策 1／2／3 名稱、原版 RNG 序列逐位元一致，以及滅國被
  下一個 caller 觀察到的完整排程。`player+0x8B8` 已確認為 Telepathic，證據見
  `custom-race-trait-consumer-census-20260828.md`。
- **停止線**：C runtime、編譯器 helper 與平台 API 不納入本列或 remake 分母。
