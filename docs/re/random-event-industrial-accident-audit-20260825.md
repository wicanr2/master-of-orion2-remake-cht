# 隨機事件 10：工業事故靜態稽核（2026-08-25）

## 證據契約

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- 工具：IDA Pro 9.4；位址為 DOS/4GW LE object #1 的 IDA 線性位址。
- 唯讀匯出器：`tools/ida/audit_event_industrial_accident.py`。
- 下列位址與控制流標為**已證實**；typed 人口群組不保存原版 packed colonist
  順序所造成的差異另標為**近似**。

## 目標資格

1. `sub_2230A @ 0x22636..0x22669` 先拒絕 `player+0x8B6 == 1`，再呼叫
   `sub_231B4`。`+0x8B6 - +0x89F = 23`，且受版控的 31 格特性索引表已證實
   23 是 `TRAIT_TOLERANT`，所以環境耐受帝國不會成為事件 10 目標。
2. `sub_231B4 @ 0x231B4..0x2325E` 最多重試 200 次：先以
   `sub_23DA0` 均勻選殖民地，再接受 `planet+8 > 1`。`planet+8` 已由
   `random-event-climate-audit-20260825.md` 證實是氣候，因此只接受
   Barren..Gaia，排除 Toxic／Radiated。
3. 200 次全失敗後，線性掃描仍持續覆寫結果，故 fallback 是最高索引的合格
   殖民地。`sub_23DA0` 另拒絕 raw `colony+0x13F != 0`；此欄尚無 typed
   玩家語意，remake 不猜測，標為**未知且非阻塞**。

## 效果與亂數

1. `sub_23833 @ 0x23833..0x238A8` 設
   `H = floor(population * (Random(3)+Random(3)) / 10)`；沒有至少 1 的下限。
2. `sub_DD2F2 @ 0xDD2F2..0xDD351` 先呼叫 `sub_DD13E` 恰好 H 次，再呼叫
   `sub_DCEBD` 結算恰好 1 點一般戰略殖民地傷害。
3. `sub_DD13E @ 0xDD13E..0xDD2F2` 每次先對所有非 Android packed colonist
   做 reservoir sampling。全殖民地只有 Android 時該次特殊命中直接浪費，之後的
   一般傷害仍會發生。
4. 人口至少 2 時，以 `Random(population+marines+tanks)` 在人口／陸戰隊／戰車間
   分配；命中人口即刪除預先抽中的非 Android 殖民者。人口少於 2 時先消耗部隊，
   無部隊才扣最後一名非 Android 的百分之一人口點數。
5. 最後一名殖民者歸零時，只有 49 格建築旗標也全空才標記殖民地毀滅；因此
   「人口 0、仍有建築」是原版可存在的中間狀態。
6. `sub_206A2 @ 0x20BB1..0x20C07` 只讀事件保存的 planet index、套用上述傷害並
   呼叫 `sub_E2A70` 重算殖民地；不改氣候。舊 remake 的「輻射使氣候惡化並固定
   扣 2 人口」斷言不成立。

## Remake 邊界

- **近似**：`PopulationGroups` 保存 race slot、職務與 prisoner，卻不保存原版 packed
  陣列順序。實作保留相同候選集合與 reservoir 分布，以固定 typed 順序重播，不能宣稱
  同 seed 下逐人索引完全相同。
- **已證實且必須保留**：Android 不進特殊人口候選、環境耐受免疫事件、氣候門檻、
  200 次重試與最高索引 fallback、H 可為 0，以及最後固定 1 點一般傷害。

