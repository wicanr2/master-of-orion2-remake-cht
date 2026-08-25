# Dragon weapon raw `0x4000` 稽核（2026-08-25）

## 證據身分

- 輸入：`Orion2.exe`，SHA-256
  `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- 工具：IDA Pro 9.4；IDA linear、DOS/4GW LE object #1
- 資料庫：`Orion2-blueprint.i64`
- 唯讀匯出器：`tools/ida/audit_monster_weapon_runtime.py`

## 已證實

1. Dragon loader `sub_57A02 @ 0x57A02` 對 weapon ID 40 寫入 raw mods `0x4000`。
   原版武器表把 ID 40 定義為 category 2 魚雷、彈藥 1、基礎傷害 300。
2. `sub_6A406 @ 0x6A406` 以 mask `0x4000` 查 15-byte weapon-mod record 的最後一筆；
   `0x17FD15 + 14×15` 的 cost／space 都是 `+50`。這與手冊 OVR「魚雷過載、成本與空間
   +50%、彈頭強度 +50%」一致，而不是 NR 的 +25% 或 ENV 的 +100%。
3. `sub_3CEB7 @ 0x3D1A9` 與 `sub_3D2DF @ 0x3D5CA` 都讀 missile runtime record
   `+0x12` high-byte bit `0x40`；成立時把 `word_199224[owner/type]` 的 ordnance 百分比加 50，
   再乘 weapon ID 的 `word_17F819` 最大傷害。這是直接的 +50 百分點傷害消費端。
4. `sub_3D2DF @ 0x3D5F2..0x3D68B` 對 ID 40 另走每格 15 點的距離衰減；低位 raw `0x20`
   才跳過這段，故 `0x4000` 不是 NR。ID 20 的一般 Plasma Torpedo 在相鄰分支每格減 5。
5. `0x4000`、+50 cost/space、category 2 限定與 +50 ordnance 四項獨立證據一致，因此可將
   Dragon ID 40 的 raw mask 命名為 OVR（Overloaded Torpedo），證據等級為已證實。

## 手冊差異

手冊怪物條目稱 Dragon Breath 最大 300，這是武器表基礎值；1.31 executable 的 Dragon
設計另帶 OVR，runtime 會先加 50% 再依距離每格減 15。remake 以 executable 的實際設計與
消費端為 oracle，文件必須同時保留這項手冊／runtime 差異，不把 300 誤稱為最終近距傷害。

