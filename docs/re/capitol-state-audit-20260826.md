# Capitol 指定行星、攻陷與重建狀態鏈

## 問題

既有證據只確認 `player+0x29` 是 Capitol 行星索引，尚未回答 Capitol 被攻陷後如何
選定重建行星、何時套用士氣懲罰，以及 raw building 9 如何重新進入建造鏈。本輪只追
玩家可見規則，不追 Win95／DOS 平台內部 helper。

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA Pro 9.4；正式資料庫 `Orion2.exe.i64` 的一次性副本，正式資料庫 SHA-256
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 位址基準：IDA 線性位址（DOS/4GW image）。
- 非破壞性匯出器：`tools/ida/audit_capitol_state.py` 與
  `tools/ida/audit_capital_morale.py`。匯出保留 raw `sub_` 名稱、原始位址、bytes 與
  caller；外部 `symbols_fixed.tsv` 名稱只作導覽，不作證據。

## 已證實規則

1. `sub_13FD9 @ 0x13FD9` 的 raw 9 完工分支 `0x14126..0x1413B` 把殖民地
   `+0x02` 行星索引寫入所有者 `player+0x29`。`Colony_Building_Score_`
   `sub_D0036 @ 0xD0036` 的 `0xD06E1..0xD0712` 只在非統一政體且目前殖民地行星
   等於 `player+0x29` 時回傳 100。
2. 原版建築狀態陣列從 `colony+0x137` 起；raw 9 的零基槽是
   `colony+0x13F`。`Colony_Morale_`（外部符號導覽；raw `sub_DDB25 @ 0xDDB25`）
   在 `0xDDC59..0xDDCB2` 沿 `player+0x29 → planet record → colony record` 讀
   `colony+0x13F`。指定行星仍存在但 raw 9 為零時，依政府表加入「無首都」懲罰；
   Unification／Galactic Unification 直接跳過。既有把 `colony+0x13F` 一概列為
   未知事件 filter 的文件敘述因此不再成立。
3. 殖民地過戶 helper `sub_ECBF7 @ 0xECBF7` 在 `0xECC36..0xECC56` 偵測 raw 9，
   呼叫 `sub_145EA` 移除 Capitol；若失去的行星等於舊擁有者 `player+0x29`，則在
   `0xECC69..0xECC78` 呼叫 `sub_ECB65`。
4. `sub_ECB65 @ 0xECB65` 排除失去的殖民地，掃描同一舊擁有者的其餘殖民地，以
   `colony+0x0A` 人口最高者為新的指定行星。比較使用 `<` 才跳過，因此同人口時，
   由倒序掃描最後留下較低殖民地索引。沒有殖民地時把 `player+0x29` 寫成 `-1`。
   此步只指定重建地點，不會直接補 raw 9。
5. 同一過戶 helper 在更換 owner 後檢查新擁有者 `player+0x29`；若為 `-1`，把被
   攻陷的行星設成其指定行星。這仍不補 raw 9，故非統一政體會承受懲罰並由 raw 9
   建造鏈重建。
6. `Player_Capitol_World_`（外部符號導覽；raw `sub_C5FB0 @ 0xC5FB0`）先尋找
   所有者相符且 `colony+0x13F != 0` 的殖民地；找不到時呼叫相鄰
   `sub_C5F5C @ 0xC5F5C`，在該玩家殖民地中作 reservoir sampling fallback。這是
   畫面／導覽 fallback，不會覆寫上述失都後的指定行星規則。

## 證據等級與 remake 邊界

- `player+0x29`、`colony+0x13F`、攻陷移除、人口最高重指派、無殖民地 `-1`、
  新擁有者空指標初始化、士氣懲罰與 raw 9 分數：**已證實**。
- 同人口 tie-break：**已證實**，來自 `sub_ECB65` 的倒序與 `<` 分支。
- 原版 UI 如何向真人解釋重建，以及 raw 建造名稱的本地化字串：本輪未追；不影響
  typed 規則與既有建造 UI，列為 **未知但非阻塞**。
- remake 必須保存每帝國的指定 Capitol 行星，並以殖民地建築集合保存 raw 9 是否完成；
  不得再以 `Colonies[0]`、母星或星系索引冒充。

