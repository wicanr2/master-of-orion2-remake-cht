# Advanced Civilization 行星選取與平衡稽核（2026-08-30）

## 證據契約

- 輸入 `Orion2.exe` SHA-256：
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫 SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`。
- 工具：IDA Pro 9.4；位址均為 `Orion2.exe.i64` 的 IDA linear address。
- 可重生匯出：`evidence/advanced-civilization-ida-20260830.json`；腳本為
  `tools/ida/audit_advanced_civilization.py`。

外部符號只供導覽。以下已證實公式均保留 raw 函式、bytes、operand、caller 與 distant-tail
邊界；`qsort_`、`memset_` 與記憶體配置器只記錄輸入輸出，不納入玩法 RE。

## 外層時序與額度

`Init_New_Game_ @ 0x124A6` 呼叫 `Advanced_Civilization_Colonies_ @ 0x62C70`，其順序為：

1. 配置並初始化逐玩家、逐星系、逐行星及 planet-owner scratch tables。
2. 把每名玩家既有母星寫入選取集合。
3. `Build_Adv_Civ_Star_List_ @ 0x62E98` 建距離表，接著建立合格行星清單與 worth。
4. `Choose_Adv_Civ_Planets_ @ 0x63577` 依隨機化玩家順序輪流挑選。
5. `Twiddle_Selected_Adv_Civ_Planets_ @ 0x638A9` 平衡已選行星。
6. 對 planet-owner table 中的全部已選行星呼叫 `Init_Colony_ @ 0x13FB8`。

`Num_Adv_Civ_Planets_ @ 0x62BB7` 的 raw 算式為：

```text
quota = trunc((trunc(starCount / 2) * 10) / playerCount)
quota = trunc((quota + 9) / 10) - 1
```

全部運算使用 signed 整數除法。最後的 `-1` 排除已經存在的母星，所以這是每名玩家的額外
行星額度，不是總殖民地數。`starCount`／`playerCount` 的正式全域名稱由相鄰生成鏈交叉驗證；
不把公式改寫成可能改變負數或奇數邊界的浮點 `ceil`。

## 候選建立與價值

### 距離與合格行星

`Build_Adv_Civ_Star_List_` 為每名玩家保存各星系距離，排除 raw star type 9，且要求星系內沒有
該玩家殖民地。只接受距離 `<=9` 的星系；`Build_Adv_Civ_Planet_List_ @ 0x63035` 再掃每星最多
五顆行星，候選必須是有效 planet ID、raw planet type 3、尚無 owner，且尚未被其他玩家候選表
占用。每名玩家最多保存 360 筆候選。

### worth 與排序

`Get_Adv_Civ_Planet_Worthiness_ @ 0x63259` 先呼叫
`Uncolonized_Planet_Worth_To_Player_ @ 0xD27A7` 取得 base worth，再乘
`Proximity_Bonus_ @ 0x63312` 並除 100。若候選位於玩家母星星系，base worth 另乘 67%（raw
`×0x43/100`），最後加回 proximity result。函式同時輸出母星星系旗標與 proximity 值。

`Proximity_Bonus_` 掃描所有玩家母星至候選所在星系的距離：只計該玩家自己的已占星系，
距離 1–2 加 20、3–4 加 10、距離 5 加 5，其他不加。`Get_Worth_For_All_Planets_ @ 0x63156`
把候選按 final worth 由高到低排序；比較器 `0x62BE1` 是 raw `rhsWorth-lhsWorth`。

## 輪流選取

`Choose_Adv_Civ_Planets_` 先隨機化最多八名玩家的順序，再逐輪呼叫
`Get_Next_Adv_Civ_Planet_ @ 0x6341C`。候選還必須：

- 尚未被其他玩家選取；
- 不與已選 planet 的 owner table 衝突；
- 距離門檻不超過 `max(difficulty+9, trunc(maxDistance/10))`；
- 玩家尚未達到前述額外行星 quota。

成功後同時寫 planet owner、star owner 與逐玩家選取記錄；候選不足的玩家會退出本輪。若所有
玩家都有候選，選取結果會移入每玩家最多 20 筆的後續平衡表，原候選標成已消費並維持排序。
這是輪流分配，不是一次把全圖最高分 planets 平均切片。

## 90% 平衡與 special 再分配

`Get_Players_Planet_Worthiness_Average_ @ 0x63CA0` 對玩家所有已選 planets 重新算 worth，回傳：

```text
average10 = trunc(sumWorth * 10 / selectedCount)
```

`Twiddle_Selected_Adv_Civ_Planets_` 先找最高 `average10`，跳過該玩家；其餘玩家逐顆處理，直到
自身平均達到 `trunc(bestAverage10 * 90 / 100)`。每顆最多六次呼叫
`Twiddle_Planet_ @ 0x63B8F`；後者由 `Random(3)` 選一種升級：

- planet `+0x05` 加 1，最大 4；
- planet `+0x08` 加 1，最大 9；
- planet `+0x0A` 加 1，最大 4。

三種升級都會消耗對應的局部剩餘次數；欄位的正式玩家名稱仍沿用 planet schema，不以本切片
單獨猜測。每次升級後立即重算玩家平均，達 90% 就停止。

另一路先計每名玩家擁有 raw special 4／5／10 的數量。落後玩家只有在少於最高平均玩家時才會
嘗試補 special；候選 planet 必須目前 special 為 0、所屬星系 special 也為 0，且不是具
Artifacts Homeworld trait 的母星。最多嘗試 100 次抽取 special；只接受 4／5／10，成功後同步
planet `+0x0F` 與 star `+0x28`。這說明 Artifacts trait 只是防止其母星被此再分配覆寫，不是把
所有 Advanced Civilization planets 固定改成 raw 10。

## 閉合與 remake 邊界

- **已證實**：額度公式、距離／owner 候選 gate、base worth＋proximity、排序、隨機玩家輪替、
  90% 平衡門檻、每顆六次三類升級、special 4／5／10 再分配與最後殖民地初始化。
- **未知但不阻塞本切片**：`+0x05／+0x08／+0x0A` 的完整 enum 名稱已由其他 planet schema
  文件管理；`Random` 內部演算法、記憶體配置與 qsort runtime 不納入 remake。
- remake 目前沒有等價的 Advanced Civilization 全圖選取／平衡器；依全域 RE-first gate，本輪
  只補證據，不建立猜測性 Go 實作或 READY spec。
