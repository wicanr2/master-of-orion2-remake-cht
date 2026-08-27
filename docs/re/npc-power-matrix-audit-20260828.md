# NPC 國力矩陣 `+0x5EC` 靜態稽核（2026-08-28）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`。
- IDA 資料庫 SHA-256：
  `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；IDA Pro 9.4
  只分析一次性副本，位址均為 IDA linear EA、DOS/4GW LE object #1。
- 原始函式、bytes、operand 與交叉參照匯出：
  [`evidence/npc-treaty-negotiations-ida-20260827.json`](evidence/npc-treaty-negotiations-ida-20260827.json)。

## 已證實：資料形狀與唯一直接 writer

`+0x5EC` 的直接站點共九處：四個讀取函式 `sub_27A3D／sub_4F59B／sub_500CF／sub_51DCE`，
以及唯一直接 writer `sub_D3D34 @ 0xD3D34..0xD574D`。後者函式 bytes SHA-256
`1574ebdb562d0547e3d83e9e44dc966e6ebfb3ff0b6186399424b812d7343aab`。

`0xD42EA..0xD4309` 對每個帝國記錄從 `+0x5EC` 起清零 32 bytes，因此不是單一國力值，而是
八個方向性 `dword`。`0xD4334..0xD43BB` 再掃描所有 stride `0x81` 艦艇記錄：owner byte `+0x63`
小於 8 時，對每個正常帝國槽把該艦在 `dword_1AA1DC` 的對應 `word` 累加到
`ownerRecord + 0x5EC + 4*observer`。

## 已證實：逐艦 producer

在同一個 `sub_D3D34` 前段，每艘存活／有效艦對各帝國槽呼叫 `sub_5EF17`，結果寫入每艦 16 bytes
的八個 `word`。相關函式身分：

- `sub_5EF17 @ 0x5EF17..0x5EF33`，bytes SHA-256
  `c3753e02ce04bc80748ca694758bc4b518457d0579fa72bae2c23aa9b8786185`；
- `sub_5EF4B @ 0x5EF4B..0x5F2C3`，bytes SHA-256
  `d624036078d35f84ae6a44408661db3a33754d0378278c1d2e3d860a32af7b8c`；
- `sub_5F871 @ 0x5F871..0x5F8BF`，bytes SHA-256
  `6d862c7137f34b8127a265537ae6bab90299e68553a9b6e28e653bf5d88e7da9`。

`sub_5EF4B` 掃每艦八個武器槽，讀武器 ID、數量、改造 flags、武器類別與基礎值；處理重／點防
改造、戰機類特殊換算、命中率下限／上限、船員／傷害修正及觀察者相關防禦值，最後形成方向性的
逐艦武器效能。`sub_5F871` 提供由帝國科技／特殊 owner 決定的扣減輸入。這些資料流證明
`+0x5EC` 不是艦體數量、建造成本或全銀河無條件可見的 `FleetStrength`。

## 已證實：消費公式

`sub_500CF @ 0x500CF` 對指定方向使用 `+0x5EC`：

```text
ratio = min(800, 100 * (sourcePower + 1) / (targetPower + 1))
ratio /= 2 for each source third-party war
```

該比例直接供 NPC 納貢與宣戰決策。公式已於 remake 接入；尚未對齊的是 producer 的逐艦方向值。

## 結論與停止條件

- **已證實**：資料形狀、唯一直接 writer、逐艦逐觀察者累加、主要武器槽輸入及 ratio consumer。
- **尚待實作**：把原版八槽武器記錄、改造 flags、船員／損傷與觀察者防禦輸入映射到 remake 的
  `Ship／WeaponMounts／PlayerState`，產生 `AIPowerRaw[owner][observer]` 並保存重播所需狀態。
- **目前近似**：宣戰／納貢仍以 hull-only `FleetStrength` 代入；必須繼續標為強推論資料投影，
  不得因 consumer 公式已對齊而宣稱完整 parity。
