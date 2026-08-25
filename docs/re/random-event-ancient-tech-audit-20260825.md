# 古代遺骸科技事件逆向稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4／IDAPython
- 位址基準：DOS/4GW object #1 的 IDA 線性位址
- 非破壞性匯出：`tools/ida/audit_event_ancient_tech.py`
- 推論等級：下列選擇、事件紀錄與授予鏈為「已證實」；函式語意名稱只作導覽。

## `sub_58853 @ 0x58853` 的輸入與輸出

唯一 caller 是 `sub_2230A @ 0x2258E`。caller 以 `eax=目標 player slot`、
`ebx=&count`、`edx=&兩格 word 輸出` 呼叫；`0x22593..0x2259A` 要求至少產生一個候選，
否則事件不適用。

helper 先把 `count` 清零，最後最多依序寫入兩個 application ID：

1. 未知的光束武器 application。
2. 未知的銀河最高護盾 application。

### 全銀河光束基準

- `sub_568EB @ 0x568EB` 對正常 player slot 掃武器 ID 0..39，只接受類別 0 光束且
  application 狀態為 3 的項目，以 `word_17F819[weapon*0x1C]` 最大傷害嚴格較大者為最佳；
  初值是武器 3「Laser Cannon」。
- `sub_58853 @ 0x58882..0x588A6` 再掃所有 player slot，選出最大傷害最高的帝國光束；
  相同傷害保留先出現者。
- `0x588F7..0x58977` 掃武器 1..39，只接受類別 0、非 Xenon topic、目標尚未知的 application，
  且其最大傷害與銀河基準差介於 0..50。候選以最小非負差更新；同差因 `jg` 判斷由後出現者
  覆蓋。若候選正是基準武器，內部最佳差設為 1，這個 raw tie 行為也須保留。

### 全銀河護盾基準

- `sub_5679E @ 0x5679E` 對正常 player slot 掃 0x3B-byte 表的索引 1..5，回傳最高已知索引。
- raw `word_17F6A7[index*0x3B]` 依序是 application `33,34,35,36,37`，與 enum 逐項對應
  Class I／III／V／VII／X Shield，因此表語意由資料與已知 application 雙重證實。
- `sub_58853 @ 0x58871..0x5887F` 取所有帝國中的最高索引；`0x589A1..0x589CE` 只在目標
  尚未知該 application 時把它附加到輸出。

## 事件紀錄與消費端

`sub_2230A` case 0 在 `0x2283E..0x22871` 把第一個 application 寫入 `record.word+3`，
第二個（若有）寫入 `record.word+5`。沒有任何 RP 金額或隨機選取。

`sub_206A2 @ 0x206A2` case 0 在 `0x2073A..0x20762` 對第一項呼叫
`sub_E4204 @ 0xE4204`；第二項非零時於 `0x2076D..0x207A2` 再呼叫一次。
`sub_E4204` 的逐 application 擁有權與特殊 callback 已由
[`research-application-callback-audit-20260825.md`](research-application-callback-audit-20260825.md)
獨立閉合。

## Remake 對映與邊界

- 使用 `OrigWeaponTable` 的原版 ID、類別、application 與最大傷害，不能用 remake 戰力估值代替。
- 全銀河基準涵蓋活躍玩家／熱座席位與 AI 的科技擁有權；不把死亡帝國的舊科技加入基準。
- 最多兩個 application 依原版順序授予，使用共用 callback，並更新玩家／AI 艦艇設計。
- 原版特殊 player slot 8..14 的固定裝備回退不屬正常八帝國對局；remake 沒有這類額外槽，
  不冒稱已重製該內部資料槽。
