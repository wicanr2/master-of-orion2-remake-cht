# 事件怪物特殊武器 runtime 稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`
- SHA-256：`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4；IDA linear、DOS/4GW LE object #1
- 資料庫：`Orion2-blueprint.i64`
- 匯出腳本：`tools/ida/audit_monster_weapon_runtime.py`
- 方法：唯讀 IDAPython 匯出；未修改原始函式名、位址、運算元或資料庫註記。

## 已證實

### ID 44：Plasma Flux

- `sub_39F1D @ 0x39F1D` 將 weapon ID `0x2C` 限制在距離值 `<= 1`。
- `sub_2A545 @ 0x2A545` 把 `0x2C` 與球形武器放在同一消費群組。
- 勘誤：`sub_289C4 @ 0x289C4` 是戰術 AI 武器估值，不是 runtime 命中端。真正消費端是
  `sub_3A82F @ 0x3A82F` → `sub_ADE18 @ 0xADE18` 的 effect type 2；它以 CMBTSFX asset 6
  半寬 96px 枚舉雙方鄰艦、套距離平方衰減，再依尺寸級數加一逐段擲值。完整資料流見
  [`event-monster-plasma-flux-spread-audit-20260825.md`](event-monster-plasma-flux-spread-audit-20260825.md)。

### ID 45：Caustic Slime

- `sub_39F1D` 將 weapon ID `0x2D` 限制在距離值 `<= 5`。
- `sub_3A82F @ 0x3A82F` 的 `loc_3AA50` 呼叫 `sub_ACF83 @ 0xACF83`。
- `sub_ACF83` 依武器表 min/max、射手加成與該槽數量逐次擲值，最後把總強度累加到
  目標 combat record `+0x43`；重複命中因此會堆疊，而非覆蓋。
- `sub_4A5CE @ 0x4A5CE` 每回合找出 `+0x43 > 0` 的存活艦，將同一強度依序送入
  `sub_39985 @ 0x39985` 四次（四個護盾朝向），之後把 `+0x43` 減 5，最低夾 0。
- `sub_39985` 先扣指定朝向護盾，剩餘傷害繼續進入艦體傷害流程；因此不能把黏液
  簡化成只扣單一護盾值。
- `sub_4C9F6 @ 0x4C9F6` 初始化 combat record 時把 `+0x43` 清零。

## 強推論與停止線

- 舊稱固定六格已被 executable＋實際 asset 推翻：半徑為 96px，約 4.8 個 20px 格位。
  remake 只保存格中心，因此缺少原版艦艇 sprite 內部像素中心，是明示資料模型近似。
- 後續勘誤：Dragon raw `0x4000` 已由 +50 ordnance 消費端、category 2 改造表及 ID 40
  每格 -15 分支閉合為 OVR；見
  [`event-monster-dragon-raw4000-audit-20260825.md`](event-monster-dragon-raw4000-audit-20260825.md)。
