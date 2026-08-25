# 隨機領袖招募靜態稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256 `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4／IDAPython；位址均為 IDA linear、DOS/4GW LE object #1。
- 可重生匯出：`tools/ida/audit_random_officer_recruitment.py`。腳本保留原始函式名、位址、bytes、caller/callee 與資料值；本文件的語意是附加層。
- 範圍：純靜態證據，未宣稱原版 PRNG 位元序列或 UI 逐幀一致。

## 已證實

1. `sub_97A66 @ 0x97A66` 才是每回合招募檢查。`Next_Turn_Calc_ @ 0x136B3` 在 `0x137D8` 直接呼叫它。它先呼叫 `sub_9781D` 算百分比，再把 `Random(100)-1` 與百分比比較；兩類在職領袖合計達八名時不擲骰。
2. `sub_97AD4 @ 0x97AD4` 是通過擲骰後的候選寫回：呼叫 `sub_97B2D`，把候選索引寫入玩家記錄 `+0x5A1`，把當前星曆寫入 `+0xE6A`，並把玩家索引寫入領袖記錄 `+0x3A`。舊外部符號把它標成 `Random_Officer_Check`，已被 caller 與資料流否定。
3. `sub_9781D @ 0x9781D` 的招募百分比：
   - 相對星曆 `sub_64395 = current_stardate - 35000` 小於 5 時為 0。
   - 基礎值為「距上次 offer 的回合數 + 1」；從未 offer 則以相對星曆代入。
   - Charismatic（玩家 `+0x8B3`）加 5；Repulsive（`+0x8B2`）減 10。
   - 每位目前屬於該玩家且狀態 `+0x39` 為 0 或 1 的 Famous 領袖加入加成。`dword_17D2CD=0x40` 與 `dword_17D2DF=0x80` 是 Famous 的一般／進階兩個 bit：一般加 `level+1`，進階加 `15*(level+1)/10`。
   - 最後以 `max(1, 艦長數 + 殖民地領袖數 + 1)` 作整數除法。
4. `sub_97B2D @ 0x97B2D` 從候選前綴隨機抽取：`relative_stardate/5 + 10 + player_slot_count`；Charismatic 再加 10，否則 Repulsive 將前綴有號除以 2；上限 67。最多嘗試 100 次。`word_199998` 的語意由 `Next_Turn_Calc_ @ 0x137D1..0x137E7` 獨立證實：它是呼叫 `Random_Officer_Check_` 的玩家槽迴圈上限，不是目前玩家索引。
5. 候選類型位於領袖記錄 `+0x23`：0 與 1 各自最多四名。`sub_97C2D @ 0x97C2D` 要求 owner `+0x3A == -1`、排除索引 65/66，且狀態 `+0x39` 也為 -1。
6. `sub_9776C` 與 `sub_977AF` 分別計算 type 0／1 的在職數；都遍歷 67 筆、核對 owner，並經 `sub_9773F` 狀態 predicate。
7. `sub_93D4B @ 0x93D4B` 的經驗等級門檻依序為 60、150、300、500、1000；招募加成把結果限制在 5。

## 勘誤與限制

- `sub_979A0 @ 0x979A0` 是領袖經驗調整鏈的一部分，不是 `Chance_To_Hire_Hero`。
- `sub_97C64 @ 0x97C64` 增加經驗並標記升級，不是 `Leader_Available_For_Hire`。
- remake 的 `Turn` 以開局後回合數表達，規格以 `Turn` 對應相對星曆；候選池不足 67 筆時安全夾到實際長度。
- remake 使用可存檔的獨立亂數流，不宣稱複製原版 PRNG；已證實的是抽取範圍、順序與消費條件。
