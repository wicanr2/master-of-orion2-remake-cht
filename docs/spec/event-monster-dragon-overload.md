# Dragon Breath OVR 規格

## Typed raw flag

- `gamedata.MonsterWeaponModOverloadedTorpedo = 0x4000`。
- `monsterTypedMods` 將該 bit 轉為 `gamedata.ModOverloadedTorpedo`；只因 Dragon Breath 是
  原版 category 2 魚雷才可被 OVR 消費端接受。

## 格子戰術

- Dragon Breath 的基礎 300 傷害先套 OVR +50%，得到近距 450。
- 現有魚雷射擊路徑繼續套距離衰減；ID 40 的原版專屬每格 15 尚須由 weapon profile 明確
  區分，不能誤用 Plasma Torpedo 的每格 5。
- 彈藥、命中與護盾／裝甲下游維持既有 category 2 路徑。

## 快速怪物戰鬥

- 快速結算沒有格位，使用近距 OVR 傷害 450；不自行捏造平均距離。
- 這是原版 raw mod 的已證實效果加上快速結算既有無距離限制，距離部分仍明示為 adapter。

## 驗收

- ID 40 raw `0x4000` 轉成 OVR；一般未帶該 bit 的魚雷不受影響。
- `MonsterWeaponDamageRange` 保留原始 mount 基礎值 300..300；只供沒有 typed mod
  消費端的快速結算使用之 `MonsterWeaponQuickDamageRange` 回傳 450..450，避免格子戰術
  又套一次 OVR 而變成 675。
- 格子戰術建構出的 Dragon 第二槽保留 raw `0x4000`，typed mods 含 OVR。
- 快速與格子戰術均通過既有怪物戰鬥回歸。
