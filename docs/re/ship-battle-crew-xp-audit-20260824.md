# 戰後艦員經驗靜態稽核（2026-08-24）

## 問題與舊歧義

remake 原先依手冊實作「被擊沉敵艦艦體級總和除二、至少 1」，但自行把「沒有擊沉」
解釋成 0，且只有快速結算接線。這份稽核回答：原版何時累加、誰取得 XP、零擊沉
如何處理，以及戰鬥 XP 是否共用每回合 500 上限。

## 輸入與工具

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- IDA 資料庫來源：`Orion2.exe.i64`
- `.i64` SHA-256：`4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`
- 工具：IDA Pro 9.4，SDK 940；本文位址均為 **IDA linear address**。
- 可重生探針：`tools/ida/audit_battle_crew_xp.py`
- 探針契約：不改名、不套型別、不寫回資料庫；同列保留 raw name、位址、bytes、
  operand 與交叉參照。攤平 `.asm` 只用來導航，結論以 `.i64` 函式邊界與資料流為準。

## 已證實

### 1. 戰後入口與唯一 caller

`sub_4B184 @ 0x4B184` 的唯一直接 caller 是 `sub_47939 @ 0x47939` 的
`0x48D52`。caller 在 `0x48D43..0x48D52` 將戰後結果及兩側資料放入 `EBX／EDX／EAX`
後呼叫它。`sub_4B184` 依結果在 `0x4B19B..0x4B1CA` 選出後續 recipient side 的
raw id。

證據等級：**已證實**。結果值的完整 enum 名稱尚未恢復，因此本文不替 raw 值改名。

### 2. 被摧毀艦艇的艦體級累加器

`sub_4B184` 處理 battle record 時使用固定 stride `0x139`：

- record `+0x20`：side／owner raw id；
- record `+0x24`：戰後狀態；
- record `+0x25`：零起算艦體級；
- record `+0x1E`：對應持久 Ship record 索引，`-1` 表示沒有可回寫艦。

在 `0x4B2D8..0x4B32C`，記錄被切到 raw status 5 時，程式把 `byte +0x25 + 1`
依 `+0x20` 所屬側加入兩個不同累加器。這證實 XP 採一到六級艦體級，而非 HP、造價、
戰力或擊沉艦數。俘獲／換旗分支沒有進入這個摧毀累加段。

證據等級：stride、偏移、`+1` 與分側累加為 **已證實**；`+0x24 == 5` 的完整狀態
enum 名稱仍為 **未知**，但該路徑的摧毀語意另由後續清除 Ship／leader 的 consumer
交叉支持。

### 3. recipient 與精確公式

`0x4B810..0x4B953` 逐筆走 battle record：

1. `0x4B8E7..0x4B8EF` 要求 record `+0x20 == SI`，即戰後選出的 side；
2. `0x4B8F1..0x4B8F8` 要求 record `+0x1E != -1`，必須連到持久 Ship；
3. `0x4B8FA..0x4B908` 選取對手側的艦體級累加器；
4. `0x4B90C..0x4B911` 以帶號除二（非負輸入等同 `floor(sum/2)`）；
5. `0x4B913..0x4B919` 小於 1 時強制為 1；
6. `0x4B921..0x4B947` 由 battle record `+0x1E` 找到 `0x81`-byte Ship record，直接
   `add [Ship+0x72], dx`。

因此原版規則是：

```text
battle_xp_per_recipient = max(1, floor(sum(destroyed_enemy_hull_class_1_to_6) / 2))
```

尤其當摧毀總和為 0 時仍得到 **1**，推翻舊文件「零擊沉回 0」的斷言。寫入端是直接
`add`，沒有呼叫 `sub_149D5`，也沒有比較 `word_17D186 = 500`；戰鬥 XP **不套用**
每回合 XP consumer 的 500 上限。

證據等級：公式、最少 1、直接寫入、無 500 cap、recipient side 與持久 Ship link
門檻皆為 **已證實**。

### 4. raw survivor predicate 的保留

在 XP 分支前，`0x4B857..0x4B8D1` 對 `+0x24 == 5`、`+0x1E` 與 `+0x4B`
另有清除／轉移分支；只有未被該分支提前帶走、且最後符合 winner side 與持久 Ship link
者才寫 XP。`+0x4B == 2` 的完整 enum 名稱尚未證實，不以推測名稱覆蓋。

remake 沒有原版完整 battle-record 狀態機，故以「戰後留在玩家艦隊的持久 Ship」表示
recipient。這是由原版 raw predicate 映射到現有資料模型的 **強推論**，不是逐欄 exact。

## Remake 映射與修正

- `gamedata.CrewBattleXPFromDestroyedHullClassSum` 實作 `max(1, sum/2)`；只在沒有
  recipient／沒有勝方時由 caller 不發放，不再用空擊沉清單回 0。
- 快速結算依 combatant 持久索引辨識真正倖存參戰艦，再發 XP；戰後另依手冊 p.119
  移除遭敵軍交戰的支援艦，支援艦不是 XP recipient。
- 格子戰術畫面在每次壓縮 `HP <= 0` 敵艦時累加 `SizeClass+1`；`Captured` 艦不累加。
- `ApplyCombatOutcome` 將該總和寫入 deterministic command args，重播時使用同一值；舊
  command 沒有此欄時以當回合初始敵艦級總和作向後相容近似。
- `BattleResult.CrewXPGained` 記錄每艘 recipient 實得值。

## 尚未證實／非阻塞限制

- battle record `+0x24`、`+0x4B` 的完整 enum 名稱與所有捕獲／撤退狀態轉移仍未知。
- remake 的 AI 艦隊沒有持久逐艦資料，無法表示原版 AI winner-side Ship XP 寫回。
- remake 快速結算仍以抽象敵艦強度生成艦體級；本輪只對齊 XP consumer，不冒稱敵艦
  blueprint exact。
